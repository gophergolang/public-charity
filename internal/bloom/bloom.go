package bloom

import (
	"hash"
	"hash/fnv"
	"math"
	"sync"

	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/storage"
)

const (
	filterSize   = 1024 // bits
	filterBytes  = filterSize / 8
	numHashes    = 3
)

type Filter struct {
	bits [filterBytes]byte
}

func (f *Filter) Add(item string) {
	for _, idx := range hashPositions(item) {
		f.bits[idx/8] |= 1 << (idx % 8)
	}
}

func (f *Filter) MayContain(item string) bool {
	for _, idx := range hashPositions(item) {
		if f.bits[idx/8]&(1<<(idx%8)) == 0 {
			return false
		}
	}
	return true
}

func (f *Filter) HasAny() bool {
	for _, b := range f.bits {
		if b != 0 {
			return true
		}
	}
	return false
}

func (f *Filter) Bytes() []byte {
	out := make([]byte, filterBytes)
	copy(out, f.bits[:])
	return out
}

func FromBytes(data []byte) *Filter {
	f := &Filter{}
	copy(f.bits[:], data)
	return f
}

func hashPositions(item string) []uint {
	positions := make([]uint, numHashes)
	var h hash.Hash64
	for i := 0; i < numHashes; i++ {
		h = fnv.New64a()
		h.Write([]byte(item))
		h.Write([]byte{byte(i)})
		positions[i] = uint(h.Sum64() % filterSize)
	}
	return positions
}

// Grid manages all bloom filters: one per cell per category.
type Grid struct {
	mu      sync.RWMutex
	filters map[string]*Filter // key: "{cell_id}/{category}"
}

func NewGrid() *Grid {
	return &Grid{
		filters: make(map[string]*Filter),
	}
}

func key(cellID, category string) string {
	return cellID + "/" + category
}

func (g *Grid) Get(cellID, category string) *Filter {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if f, ok := g.filters[key(cellID, category)]; ok {
		return f
	}
	return &Filter{}
}

func (g *Grid) HasUsers(cellID, category string) bool {
	return g.Get(cellID, category).HasAny()
}

func (g *Grid) AddUser(cellID, category, username string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	k := key(cellID, category)
	f, ok := g.filters[k]
	if !ok {
		f = &Filter{}
		g.filters[k] = f
	}
	f.Add(username)
}

func (g *Grid) UpdateUser(u *manifest.User) {
	for _, cat := range manifest.Categories {
		if u.NeedScores.AboveThreshold(cat) {
			g.AddUser(u.CellID, cat, u.Username)
		}
	}
}

// RebuildCell rebuilds all bloom filters for a cell from the users in it.
func (g *Grid) RebuildCell(cellID string, users []*manifest.User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, cat := range manifest.Categories {
		k := key(cellID, cat)
		f := &Filter{}
		for _, u := range users {
			if u.NeedScores.AboveThreshold(cat) {
				f.Add(u.Username)
			}
		}
		g.filters[k] = f
	}
}

func (g *Grid) Persist() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for k, f := range g.filters {
		path := "bloom/" + k + ".bloom"
		if err := storage.WriteRaw(path, f.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (g *Grid) Load() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	cellDirs, err := storage.ListDir("bloom")
	if err != nil {
		return nil // no bloom data yet
	}
	for _, cellID := range cellDirs {
		for _, cat := range manifest.Categories {
			path := "bloom/" + cellID + "/" + cat + ".bloom"
			data, err := storage.ReadRaw(path)
			if err != nil {
				continue
			}
			g.filters[key(cellID, cat)] = FromBytes(data)
		}
	}
	return nil
}

// EstimateCount gives a rough count of items in a filter (for monitoring).
func (f *Filter) EstimateCount() int {
	setBits := 0
	for _, b := range f.bits {
		setBits += popcount(b)
	}
	if setBits == 0 {
		return 0
	}
	m := float64(filterSize)
	k := float64(numHashes)
	n := -(m / k) * math.Log(1-float64(setBits)/m)
	return int(math.Round(n))
}

func popcount(b byte) int {
	count := 0
	for b != 0 {
		count += int(b & 1)
		b >>= 1
	}
	return count
}
