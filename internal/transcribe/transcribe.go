package transcribe

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Whisper transcribes a WAV file using whisper.cpp.
// whisperPath is the path to the whisper.cpp executable.
// modelPath is the path to the ggml model file.
func Whisper(wavFile, whisperPath, modelPath string) (string, error) {
	if !filepath.IsAbs(whisperPath) {
		exe, err := os.Executable()
		if err == nil {
			whisperPath = filepath.Join(filepath.Dir(exe), whisperPath)
		}
	}
	if !filepath.IsAbs(modelPath) {
		exe, err := os.Executable()
		if err == nil {
			modelPath = filepath.Join(filepath.Dir(exe), modelPath)
		}
	}

	base := wavFile[:len(wavFile)-len(filepath.Ext(wavFile))]

	cmd := exec.Command(whisperPath,
		"-m", modelPath,
		"-f", wavFile,
		"-l", "ru",
		"-otxt",
		"-of", base,
	)
	cmd.Dir = filepath.Dir(whisperPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper: %w\n%s", err, out)
	}

	txtData, err := os.ReadFile(base + ".txt")
	if err != nil {
		return "", fmt.Errorf("read transcript: %w", err)
	}

	return string(txtData), nil
}
