// Package messages provides the message bucket for async delivery.
// Adapted from tsdb's Client pattern:
// - Channel-based non-blocking submission (buffered channel like tsdb's Client.Channel)
// - Background goroutine flushes to disk periodically (tsdb's batchLogs pattern)
// - File-based ingestion with scan+process+delete on pull (tsdb's Ingest pattern)
package messages

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/gophergolang/public-charity/internal/storage"
)

// Message is a single pending offer/notification for a user.
type Message struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	From      string `json:"from"`
	Category  string `json:"category"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	RuleID    string `json:"rule_id,omitempty"`
}

// Bucket is a channel-based message writer adapted from tsdb's Client.
// Messages are submitted non-blocking to a buffered channel, then flushed
// to disk by a background goroutine on a timer (tsdb's batchLogs pattern).
type Bucket struct {
	ch       chan pendingMsg
	flushInterval time.Duration
	wg       sync.WaitGroup
	stop     chan struct{}
}

type pendingMsg struct {
	username string
	msg      *Message
}

const channelBuffer = 500 // same buffer size as tsdb Client.Channel

// NewBucket creates a bucket with a background flush goroutine.
// Mirrors tsdb's NewClient + batchLogs pattern.
func NewBucket(flushInterval time.Duration) *Bucket {
	b := &Bucket{
		ch:            make(chan pendingMsg, channelBuffer),
		flushInterval: flushInterval,
		stop:          make(chan struct{}),
	}
	b.wg.Add(1)
	go b.flushLoop()
	return b
}

// Submit queues a message for async write. Non-blocking if buffer isn't full.
// Mirrors tsdb's Client.Publish pattern.
func (b *Bucket) Submit(username string, msg *Message) {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixNano()
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg_%d", msg.Timestamp)
	}
	select {
	case b.ch <- pendingMsg{username: username, msg: msg}:
	default:
		// Channel full — write directly to avoid data loss
		writeMsg(username, msg)
	}
}

// flushLoop is the background goroutine that periodically drains the channel
// and writes messages to disk. Adapted from tsdb's batchLogs timer pattern.
func (b *Bucket) flushLoop() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.drain()
		case <-b.stop:
			b.drain() // flush remaining on shutdown
			return
		}
	}
}

func (b *Bucket) drain() {
	for {
		select {
		case pm := <-b.ch:
			if err := writeMsg(pm.username, pm.msg); err != nil {
				log.Printf("bucket write %s: %v", pm.username, err)
			}
		default:
			return
		}
	}
}

// Stop gracefully shuts down the flush loop.
func (b *Bucket) Stop() {
	close(b.stop)
	b.wg.Wait()
}

func msgDir(username string) string {
	return filepath.Join("messages", username)
}

func writeMsg(username string, msg *Message) error {
	path := filepath.Join(msgDir(username), msg.ID+".json")
	return storage.WriteJSON(path, msg)
}

// Pull reads all pending messages and deletes them from the bucket.
// This is the batch-pull-on-boot pattern: app calls once, gets everything, files are cleaned up.
// Mirrors tsdb's Ingest scan+process+delete loop via storage.ScanAndProcess.
func Pull(username string) ([]Message, error) {
	dir := msgDir(username)
	var msgs []Message
	_, err := storage.ScanAndProcess(dir, ".json", func(path string, data []byte) error {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		msgs = append(msgs, msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// Count returns the number of pending messages.
func Count(username string) int {
	files, err := storage.ListFiles(msgDir(username), ".json")
	if err != nil {
		return 0
	}
	return len(files)
}

// WriteSync writes a message directly (bypassing the channel). Used by the gateway API.
func WriteSync(username string, msg *Message) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixNano()
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg_%d", msg.Timestamp)
	}
	return writeMsg(username, msg)
}
