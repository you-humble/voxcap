package wav

import (
	"encoding/binary"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	w, err := New("test.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer w.CloseWithoutHeader()
	defer os.Remove("test.wav")

	if w.sampleRate != 16000 {
		t.Errorf("sampleRate = %d, want 16000", w.sampleRate)
	}
	if w.channels != 1 {
		t.Errorf("channels = %d, want 1", w.channels)
	}
	if w.bitsPerSample != 16 {
		t.Errorf("bitsPerSample = %d, want 16", w.bitsPerSample)
	}
	if w.dataSize != 0 {
		t.Errorf("dataSize = %d, want 0", w.dataSize)
	}
}

func TestWritePCM(t *testing.T) {
	w, err := New("test.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer w.CloseWithoutHeader()
	defer os.Remove("test.wav")

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := w.WritePCM(data)
	if err != nil {
		t.Fatalf("WritePCM() error: %v", err)
	}
	if n != 1024 {
		t.Errorf("WritePCM() wrote %d bytes, want 1024", n)
	}
	if w.dataSize != 1024 {
		t.Errorf("dataSize = %d, want 1024", w.dataSize)
	}
}

func TestWriteWAVHeader(t *testing.T) {
	w, err := New("test.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer os.Remove("test.wav")

	data := []byte{0, 1, 2, 3}
	w.WritePCM(data)
	w.Close()

	f, err := os.Open("test.wav")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer f.Close()

	// RIFF
	riff := make([]byte, 4)
	f.Read(riff)
	if string(riff) != "RIFF" {
		t.Errorf("header = %s, want RIFF", string(riff))
	}

	// File size
	var fileSize uint32
	binary.Read(f, binary.LittleEndian, &fileSize)
	expectedSize := uint32(36 + 4)
	if fileSize != expectedSize {
		t.Errorf("fileSize = %d, want %d", fileSize, expectedSize)
	}

	// WAVE
	wave := make([]byte, 4)
	f.Read(wave)
	if string(wave) != "WAVE" {
		t.Errorf("wave = %s, want WAVE", string(wave))
	}

	// fmt
	fmtChunk := make([]byte, 4)
	f.Read(fmtChunk)
	if string(fmtChunk) != "fmt " {
		t.Errorf("fmt = %s, want 'fmt '", string(fmtChunk))
	}

	// Subchunk1Size
	var subchunk1Size uint32
	binary.Read(f, binary.LittleEndian, &subchunk1Size)
	if subchunk1Size != 16 {
		t.Errorf("subchunk1Size = %d, want 16", subchunk1Size)
	}

	// Audio format
	var audioFormat uint16
	binary.Read(f, binary.LittleEndian, &audioFormat)
	if audioFormat != 1 {
		t.Errorf("audioFormat = %d, want 1", audioFormat)
	}

	// Channels
	var channels uint16
	binary.Read(f, binary.LittleEndian, &channels)
	if channels != 1 {
		t.Errorf("channels = %d, want 1", channels)
	}

	// Sample rate
	var sampleRate uint32
	binary.Read(f, binary.LittleEndian, &sampleRate)
	if sampleRate != 16000 {
		t.Errorf("sampleRate = %d, want 16000", sampleRate)
	}

	// Byte rate
	var byteRate uint32
	binary.Read(f, binary.LittleEndian, &byteRate)

	// Block align
	var blockAlign uint16
	binary.Read(f, binary.LittleEndian, &blockAlign)

	// Bits per sample
	var bitsPerSample uint16
	binary.Read(f, binary.LittleEndian, &bitsPerSample)

	// data
	dataChunk := make([]byte, 4)
	f.Read(dataChunk)
	if string(dataChunk) != "data" {
		t.Errorf("data chunk = %s, want 'data'", string(dataChunk))
	}

	// Data size
	var dataSize uint32
	binary.Read(f, binary.LittleEndian, &dataSize)
	if dataSize != 4 {
		t.Errorf("dataSize = %d, want 4", dataSize)
	}
}

func TestMultipleWrites(t *testing.T) {
	w, err := New("test.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer w.CloseWithoutHeader()
	defer os.Remove("test.wav")

	for i := 0; i < 10; i++ {
		data := make([]byte, 100)
		_, err := w.WritePCM(data)
		if err != nil {
			t.Fatalf("WritePCM() error at %d: %v", i, err)
		}
	}

	if w.dataSize != 1000 {
		t.Errorf("dataSize = %d, want 1000", w.dataSize)
	}
}

func TestCloseWithoutHeader(t *testing.T) {
	w, err := New("test.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer os.Remove("test.wav")

	w.WritePCM([]byte{1, 2, 3})
	w.CloseWithoutHeader()

	// File should exist but header not finalized
	info, err := os.Stat("test.wav")
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if info.Size() != 44+3 {
		t.Errorf("file size = %d, want 47", info.Size())
	}
}
