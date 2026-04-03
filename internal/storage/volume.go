package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var BasePath = "/data"

func fullPath(parts ...string) string {
	all := append([]string{BasePath}, parts...)
	return filepath.Join(all...)
}

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

func ReadJSON(path string, v any) error {
	data, err := os.ReadFile(fullPath(path))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func Exists(path string) bool {
	_, err := os.Stat(fullPath(path))
	return err == nil
}

func Delete(path string) error {
	return os.RemoveAll(fullPath(path))
}

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

func WriteRaw(path string, data []byte) error {
	full := fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(full, data, 0644)
}

func ReadRaw(path string) ([]byte, error) {
	return os.ReadFile(fullPath(path))
}
