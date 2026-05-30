package mix

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/you-humble/voxcap/internal/wav"
)

// Mono merges two mono WAV files into one by summing samples with clipping.
// If files have different lengths, the longer one is truncated.
func Mono(file1, file2, output string) error {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("read %s: %w", file1, err)
	}
	data2, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("read %s: %w", file2, err)
	}

	samples1 := bytesToSamples(data1[44:])
	samples2 := bytesToSamples(data2[44:])

	// If one file is empty, just copy the other
	if len(samples1) == 0 && len(samples2) == 0 {
		return fmt.Errorf("both files are empty")
	}
	if len(samples1) == 0 {
		return copyWAV(output, data2)
	}
	if len(samples2) == 0 {
		return copyWAV(output, data1)
	}

	// Use shorter length
	n := len(samples1)
	if len(samples2) < n {
		n = len(samples2)
	}

	// Mix with clipping
	mixed := make([]int16, n)
	for i := 0; i < n; i++ {
		sum := int32(samples1[i]) + int32(samples2[i])
		if sum > math.MaxInt16 {
			sum = math.MaxInt16
		} else if sum < math.MinInt16 {
			sum = math.MinInt16
		}
		mixed[i] = int16(sum)
	}

	return writeWAV(output, mixed, 16000, 1, 16)
}

// copyWAV copies raw WAV bytes to output file.
func copyWAV(output string, data []byte) error {
	return os.WriteFile(output, data, 0644)
}

// writeWAV creates a WAV file from samples.
func writeWAV(output string, samples []int16, sampleRate uint32, channels, bits uint16) error {
	w, err := wav.New(output, sampleRate, channels, bits)
	if err != nil {
		return err
	}
	defer w.Close()

	for _, s := range samples {
		b := [2]byte{
			byte(s & 0xFF),
			byte((s >> 8) & 0xFF),
		}
		if _, err := w.WritePCM(b[:]); err != nil {
			return err
		}
	}
	return nil
}

// bytesToSamples converts raw PCM bytes to int16 samples.
func bytesToSamples(data []byte) []int16 {
	n := len(data) / 2
	samples := make([]int16, n)
	for i := 0; i < n; i++ {
		samples[i] = int16(data[i*2]) | int16(data[i*2+1])<<8
	}
	return samples
}

// MixLatest finds the latest loopback and mic files, mixes them, returns output path.
func MixLatest(loopPattern, micPattern string) (string, error) {
	loopFiles, _ := filepath.Glob(loopPattern)
	micFiles, _ := filepath.Glob(micPattern)

	if len(loopFiles) == 0 || len(micFiles) == 0 {
		return "", fmt.Errorf("no files to mix")
	}

	// Use last ones
	loopFile := loopFiles[len(loopFiles)-1]
	micFile := micFiles[len(micFiles)-1]

	output := fmt.Sprintf("output/mixed_%s.wav", time.Now().Format("20060102_150405"))
	if err := Mono(loopFile, micFile, output); err != nil {
		return "", err
	}

	return output, nil
}
