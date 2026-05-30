package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	OutputFile    string `json:"output_file"`
	SampleRate    uint32 `json:"sample_rate"`
	Channels      uint16 `json:"channels"`
	BitsPerSample uint16 `json:"bits_per_sample"`
	DeviceType    string `json:"device_type"`
}

func Load(explicitPath string) (*Config, error) {
	path := resolvePath(explicitPath)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.SampleRate == 0 {
		cfg.SampleRate = 16000
	}
	if cfg.Channels == 0 {
		cfg.Channels = 1
	}
	if cfg.BitsPerSample == 0 {
		cfg.BitsPerSample = 16
	}
	if cfg.OutputFile == "" {
		cfg.OutputFile = "output.wav"
	}

	return &cfg, nil
}

// resolvePath determines the config file path.
// Priority: explicit > VOXCAP_CONFIG env > default "config.json".
func resolvePath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv("VOXCAP_CONFIG"); env != "" {
		return env
	}
	return "config.json"
}
