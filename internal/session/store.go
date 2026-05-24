package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"
)

// SessionDir returns the directory where session files are stored.
func SessionDir() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, "Footage", "sessions")
	return dir, os.MkdirAll(dir, 0755)
}

// New creates a new session with a fresh ULID.
func New(name string) *Session {
	now := time.Now()
	return &Session{
		ID:        "session_" + ulid.Make().String(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Files:     []string{},
		ActiveIdx: 0,
	}
}

// Save writes a session to its JSON file. Uses temp+rename for atomicity.
func Save(s *Session) error {
	dir, err := SessionDir()
	if err != nil {
		return err
	}
	s.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a session by ID.
func Load(id string) (*Session, error) {
	dir, err := SessionDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// List returns all sessions sorted by UpdatedAt descending.
func List() ([]*Session, error) {
	dir, err := SessionDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()
		id := name[:len(name)-len(".json")]
		s, err := Load(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})
	return sessions, nil
}

// Delete removes a session file. Does not delete the manifest.
func Delete(id string) error {
	dir, err := SessionDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
