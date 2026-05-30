package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	SampleRate    uint32         `json:"sample_rate"`
	Channels      uint16         `json:"channels"`
	BitsPerSample uint16         `json:"bits_per_sample"`
	Devices       []DeviceConfig `json:"devices"`
	WhisperPath   string         `json:"whisper_path"`
	ModelPath     string         `json:"model_path"`
}

// DeviceConfig holds configuration for a single recording device.
type DeviceConfig struct {
	Type       string `json:"type"`        // "loopback" or "microphone"
	OutputFile string `json:"output_file"` // path to output WAV file
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

	// Apply defaults
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 16000
	}
	if cfg.Channels == 0 {
		cfg.Channels = 1
	}
	if cfg.BitsPerSample == 0 {
		cfg.BitsPerSample = 16
	}

	if cfg.WhisperPath == "" {
		cfg.WhisperPath = "whisper.exe"
	}
	if cfg.ModelPath == "" {
		cfg.ModelPath = "models/ggml-base.bin"
	}

	// Validate devices
	for i, dev := range cfg.Devices {
		if dev.Type != "loopback" && dev.Type != "microphone" {
			return nil, fmt.Errorf("unsupported device type: %s", dev.Type)
		}
		if dev.OutputFile == "" {
			cfg.Devices[i].OutputFile = fmt.Sprintf("output/%s.wav", dev.Type)
		}
	}
	if len(cfg.Devices) == 0 {
		return nil, fmt.Errorf("no devices configured")
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
	return "configs/config.json"
}
