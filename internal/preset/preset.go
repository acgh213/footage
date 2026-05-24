package preset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Tag is a hotkey-label-color binding within a preset.
type Tag struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Color string `json:"color"`
}

// Preset groups a named set of tags with an optional auto-notes flag.
type Preset struct {
	Name      string `json:"name"`
	Tags      []Tag  `json:"tags"`
	AutoNotes bool   `json:"auto_notes"`
}

// PresetDir returns the global preset storage directory.
func PresetDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Footage", "presets"), nil
}

// Load reads a preset by name from the presets directory.
func Load(name string) (*Preset, error) {
	dir, err := PresetDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return nil, err
	}
	var p Preset
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Save writes a preset to disk atomically.
func Save(p *Preset) error {
	dir, err := PresetDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, p.Name+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// List returns the names of all presets in the presets directory.
func List() ([]string, error) {
	dir, err := PresetDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// ListAll loads and returns all presets.
func ListAll() ([]Preset, error) {
	names, err := List()
	if err != nil {
		return nil, err
	}
	presets := make([]Preset, 0, len(names))
	for _, name := range names {
		p, err := Load(name)
		if err != nil {
			continue
		}
		presets = append(presets, *p)
	}
	return presets, nil
}
