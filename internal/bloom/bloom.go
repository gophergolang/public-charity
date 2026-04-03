// Package bloom provides per-cell per-category bloom filters for the geographic grid.
// Adapted from golangdaddy/tsdb's Period+BloomFilter pattern:
// each grid cell embeds a bloom filter per need category for O(1) membership triage.
package bloom

import (
	"bytes"
	"encoding/gob"
	"hash/fnv"
	"sync"

	bloomfilter "github.com/bits-and-blooms/bloom/v3"

	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/storage"
)

const (
	// Tuned for ~10,000 items per cell per category with 0.01% false positive rate.
	// Same parameters used in golangdaddy/tsdb for high-throughput triage.
	expectedItems    = 10000
	falsePositiveRate = 0.0001
)

// Cell mirrors tsdb's Period struct: a bloom filter + RWMutex protecting a single
// grid cell + category combination. The bloom filter enables sub-microsecond triage
// so the agent service can skip entire cells without reading any manifests.
type Cell struct {
	filter *bloomfilter.BloomFilter
	sync.RWMutex
	n int64 // occurrence counter (like tsdb's Period.n)
}

func newCell() *Cell {
	return &Cell{
		filter: bloomfilter.NewWithEstimates(expectedItems, falsePositiveRate),
	}
}

// Add hashes the username with FNV-1a (same hash family as tsdb labels) and adds it.
func (c *Cell) Add(username string) {
	c.Lock()
	defer c.Unlock()
	c.filter.Add(fnvHash(username))
	c.n++
}

// Test checks if the username might be in this cell (probabilistic).
func (c *Cell) Test(username string) bool {
	c.RLock()
	defer c.RUnlock()
	return c.filter.Test(fnvHash(username))
}

// HasAny returns true if any items have been added.
func (c *Cell) HasAny() bool {
	c.RLock()
	defer c.RUnlock()
	return c.n > 0
}

// Count returns the number of items added.
func (c *Cell) Count() int64 {
	c.RLock()
	defer c.RUnlock()
	return c.n
}

// Encode serializes the bloom filter using GOB (tsdb pattern: GOB for all persistence).
func (c *Cell) Encode() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(c.n); err != nil {
		return nil, err
	}
	if err := enc.Encode(c.filter); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode restores a cell from GOB bytes.
func (c *Cell) Decode(data []byte) error {
	c.Lock()
	defer c.Unlock()
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&c.n); err != nil {
		return err
	}
	c.filter = &bloomfilter.BloomFilter{}
	return dec.Decode(c.filter)
}

// fnvHash produces FNV-1a hash bytes for bloom filter insertion.
// FNV-1a is the same hash family used throughout tsdb for labels and configs.
func fnvHash(s string) []byte {
	h := fnv.New64a()
	h.Write([]byte(s))
	b := h.Sum(nil)
	return b
}

// Grid manages all bloom filters: one Cell per grid cell per category.
// This is the top-level structure analogous to tsdb's Tree, holding the
// full spatial index in memory.
type Grid struct {
	mu    sync.RWMutex
	cells map[string]*Cell // key: "{cell_id}/{category}"
}

func NewGrid() *Grid {
	return &Grid{
		cells: make(map[string]*Cell),
	}
}

func cellKey(cellID, category string) string {
	return cellID + "/" + category
}

// getOrCreate lazily initializes cells on demand (tsdb pattern: lazy period creation).
func (g *Grid) getOrCreate(cellID, category string) *Cell {
	k := cellKey(cellID, category)
	g.mu.RLock()
	c, ok := g.cells[k]
	g.mu.RUnlock()
	if ok {
		return c
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	// Double-check after write lock
	if c, ok = g.cells[k]; ok {
		return c
	}
	c = newCell()
	g.cells[k] = c
	return c
}

func (g *Grid) Get(cellID, category string) *Cell {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if c, ok := g.cells[cellKey(cellID, category)]; ok {
		return c
	}
	return nil
}

// HasUsers checks if a cell has any users in a given category.
func (g *Grid) HasUsers(cellID, category string) bool {
	c := g.Get(cellID, category)
	if c == nil {
		return false
	}
	return c.HasAny()
}

// AddUser adds a username to the appropriate cell+category bloom filter.
func (g *Grid) AddUser(cellID, category, username string) {
	c := g.getOrCreate(cellID, category)
	c.Add(username)
}

// UpdateUser adds the user to all category filters where their score exceeds threshold.
func (g *Grid) UpdateUser(u *manifest.User) {
	for _, cat := range manifest.Categories {
		if u.NeedScores.AboveThreshold(cat) {
			g.AddUser(u.CellID, cat, u.Username)
		}
	}
}

// RebuildCell rebuilds all bloom filters for a cell from scratch.
// Required because bloom filters don't support removal (tsdb rebuilds similarly on cleanup).
func (g *Grid) RebuildCell(cellID string, users []*manifest.User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, cat := range manifest.Categories {
		k := cellKey(cellID, cat)
		c := newCell()
		for _, u := range users {
			if u.NeedScores.AboveThreshold(cat) {
				c.filter.Add(fnvHash(u.Username))
				c.n++
			}
		}
		g.cells[k] = c
	}
}

// Persist writes all bloom filters to disk using GOB encoding (tsdb's BackupToDisk pattern).
func (g *Grid) Persist() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for k, c := range g.cells {
		data, err := c.Encode()
		if err != nil {
			return err
		}
		path := "bloom/" + k + ".gob"
		if err := storage.WriteRaw(path, data); err != nil {
			return err
		}
	}
	return nil
}

// Load restores all bloom filters from disk (tsdb's LoadBackupFile pattern).
func (g *Grid) Load() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	cellDirs, err := storage.ListDir("bloom")
	if err != nil {
		return nil // no bloom data yet
	}
	for _, cellID := range cellDirs {
		for _, cat := range manifest.Categories {
			path := "bloom/" + cellID + "/" + cat + ".gob"
			data, err := storage.ReadRaw(path)
			if err != nil {
				continue
			}
			c := newCell()
			if err := c.Decode(data); err != nil {
				continue
			}
			g.cells[cellKey(cellID, cat)] = c
		}
	}
	return nil
}

// Stats returns total cells and total items across the grid (for monitoring).
func (g *Grid) Stats() (cellCount int, itemCount int64) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, c := range g.cells {
		cellCount++
		itemCount += c.Count()
	}
	return
}
