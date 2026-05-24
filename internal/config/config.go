package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user preferences and app settings.
type Config struct {
	LastSession   string    `json:"last_session,omitempty"`
	DefaultPreset string    `json:"default_preset,omitempty"`
	SeekShort     float64   `json:"seek_short"`
	SeekLong      float64   `json:"seek_long"`
	MPVPath       string    `json:"mpv_path,omitempty"`
	FFmpegPath    string    `json:"ffmpeg_path,omitempty"`
	LLM           LLMConfig `json:"llm"`
}

// LLMConfig holds settings for the LLM query engine.
type LLMConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
	Model    string `json:"model,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Footage", "config.json"), nil
}

// Load reads the config file, returning defaults if it doesn't exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes the config to disk atomically.
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
