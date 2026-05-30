package config

import (
	"os"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	json := `{
		"sample_rate": 16000,
		"channels": 1,
		"bits_per_sample": 16,
		"devices": [
			{"type": "loopback", "output_file": "output/loop.wav"},
			{"type": "microphone", "output_file": "output/mic.wav"}
		]
	}`
	tmp := writeTemp(t, json)
	defer os.Remove(tmp)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", cfg.SampleRate)
	}
	if len(cfg.Devices) != 2 {
		t.Errorf("len(Devices) = %d, want 2", len(cfg.Devices))
	}
	if cfg.Devices[0].Type != "loopback" {
		t.Errorf("Devices[0].Type = %s, want loopback", cfg.Devices[0].Type)
	}
}

func TestLoadDefaults(t *testing.T) {
	json := `{
		"devices": [
			{"type": "loopback"}
		]
	}`
	tmp := writeTemp(t, json)
	defer os.Remove(tmp)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", cfg.SampleRate)
	}
	if cfg.Channels != 1 {
		t.Errorf("Channels = %d, want 1", cfg.Channels)
	}
	if cfg.BitsPerSample != 16 {
		t.Errorf("BitsPerSample = %d, want 16", cfg.BitsPerSample)
	}
	if cfg.Devices[0].OutputFile != "output/loopback.wav" {
		t.Errorf("OutputFile = %s, want output/loopback.wav", cfg.Devices[0].OutputFile)
	}
}

func TestLoadInvalidDevice(t *testing.T) {
	json := `{
		"devices": [
			{"type": "invalid"}
		]
	}`
	tmp := writeTemp(t, json)
	defer os.Remove(tmp)

	_, err := Load(tmp)
	if err == nil {
		t.Error("Load() should return error for invalid device type")
	}
}

func TestLoadNoDevices(t *testing.T) {
	json := `{
		"devices": []
	}`
	tmp := writeTemp(t, json)
	defer os.Remove(tmp)

	_, err := Load(tmp)
	if err == nil {
		t.Error("Load() should return error for empty devices")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("nonexistent.json")
	if err == nil {
		t.Error("Load() should return error for missing file")
	}
}

func TestResolvePathEnv(t *testing.T) {
	os.Setenv("VOXCAP_CONFIG", "/custom/path.json")
	defer os.Unsetenv("VOXCAP_CONFIG")

	path := resolvePath("")
	if path != "/custom/path.json" {
		t.Errorf("resolvePath() = %s, want /custom/path.json", path)
	}
}

func TestResolvePathExplicit(t *testing.T) {
	path := resolvePath("/explicit.json")
	if path != "/explicit.json" {
		t.Errorf("resolvePath() = %s, want /explicit.json", path)
	}
}

func TestResolvePathDefault(t *testing.T) {
	path := resolvePath("")
	if path != "configs/config.json" {
		t.Errorf("resolvePath() = %s, want configs/config.json", path)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("CreateTemp() error: %v", err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}
