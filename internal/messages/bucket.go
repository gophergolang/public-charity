package messages

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gophergolang/public-charity/internal/storage"
)

type Message struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	From      string `json:"from"`
	Category  string `json:"category"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	RuleID    string `json:"rule_id,omitempty"`
}

func msgDir(username string) string {
	return filepath.Join("messages", username)
}

func Write(username string, msg *Message) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg_%d", msg.Timestamp)
	}
	path := filepath.Join(msgDir(username), msg.ID+".json")
	return storage.WriteJSON(path, msg)
}

// Pull reads all pending messages and deletes them from the bucket.
func Pull(username string) ([]Message, error) {
	dir := msgDir(username)
	files, err := storage.ListFiles(dir, ".json")
	if err != nil {
		return nil, nil // no messages
	}
	msgs := make([]Message, 0, len(files))
	for _, f := range files {
		var msg Message
		path := filepath.Join(dir, f)
		if err := storage.ReadJSON(path, &msg); err != nil {
			continue
		}
		msgs = append(msgs, msg)
		storage.Delete(path)
	}
	return msgs, nil
}

func Count(username string) int {
	files, err := storage.ListFiles(msgDir(username), ".json")
	if err != nil {
		return 0
	}
	return len(files)
}
