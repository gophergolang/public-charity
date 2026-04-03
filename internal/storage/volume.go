// Package storage provides flat-file I/O on Fly.io volumes.
// Adapted from tsdb's file-based persistence patterns:
// - GOB encoding for internal data (bloom filters, backups)
// - JSON encoding for API-facing manifests
// - File-based ingestion with scan+process+delete loops
package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var BasePath = "/data"

func init() {
	if v := os.Getenv("DATA_PATH"); v != "" {
		BasePath = v
	}
}

func fullPath(parts ...string) string {
	all := append([]string{BasePath}, parts...)
	return filepath.Join(all...)
}

// WriteJSON writes a value as indented JSON (used for manifests — human-readable).
func WriteJSON(path string, v any) error {
	full := fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(full, data, 0644)
}

// ReadJSON reads a JSON file into v.
func ReadJSON(path string, v any) error {
	data, err := os.ReadFile(fullPath(path))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteGOB writes a value as GOB-encoded bytes (tsdb pattern: efficient binary persistence).
func WriteGOB(path string, v any) error {
	full := fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return fmt.Errorf("gob encode: %w", err)
	}
	return os.WriteFile(full, buf.Bytes(), 0644)
}

// ReadGOB reads a GOB-encoded file into v.
func ReadGOB(path string, v any) error {
	data, err := os.ReadFile(fullPath(path))
	if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

// Exists checks if a path exists on the volume.
func Exists(path string) bool {
	_, err := os.Stat(fullPath(path))
	return err == nil
}

// Delete removes a file or directory tree.
func Delete(path string) error {
	return os.RemoveAll(fullPath(path))
}

// ListDir returns sorted directory entry names.
func ListDir(path string) ([]string, error) {
	entries, err := os.ReadDir(fullPath(path))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// ListFiles returns sorted filenames with a given suffix.
func ListFiles(path, suffix string) ([]string, error) {
	entries, err := os.ReadDir(fullPath(path))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// WriteRaw writes raw bytes to a path.
func WriteRaw(path string, data []byte) error {
	full := fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(full, data, 0644)
}

// ReadRaw reads raw bytes from a path.
func ReadRaw(path string) ([]byte, error) {
	return os.ReadFile(fullPath(path))
}

// ScanAndProcess implements tsdb's file-based ingestion pattern:
// list files in a directory, call a handler for each, delete on success.
// Used by the message bucket and any async file-based processing.
func ScanAndProcess(dir, suffix string, handler func(path string, data []byte) error) (int, error) {
	files, err := ListFiles(dir, suffix)
	if err != nil {
		return 0, nil // directory doesn't exist yet = nothing to process
	}
	processed := 0
	for _, f := range files {
		path := filepath.Join(dir, f)
		data, err := ReadRaw(path)
		if err != nil {
			continue
		}
		if err := handler(path, data); err != nil {
			continue
		}
		Delete(path)
		processed++
	}
	return processed, nil
}
