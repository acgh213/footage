package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type pointer struct {
	LastSessionID string `json:"last_session_id"`
}

func pointerPath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "Footage", "last_session.json"), nil
}

// SavePointer records the last active session ID.
func SavePointer(id string) error {
	path, err := pointerPath()
	if err != nil {
		return err
	}
	data, _ := json.Marshal(pointer{LastSessionID: id})
	return os.WriteFile(path, data, 0644)
}

// LoadLast loads and returns the last-used session, or nil if none.
func LoadLast() (*Session, error) {
	path, err := pointerPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var p pointer
	if err := json.Unmarshal(data, &p); err != nil || p.LastSessionID == "" {
		return nil, nil
	}
	return Load(p.LastSessionID)
}
