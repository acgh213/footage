package config

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		SeekShort: 5.0,
		SeekLong:  30.0,
	}
}
