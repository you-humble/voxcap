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
	w, err := New("test_header.wav", 16000, 1, 16)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	data := []byte{0x01, 0x02, 0x03, 0x04}
	if _, err := w.WritePCM(data); err != nil {
		t.Fatalf("WritePCM() error: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	defer os.Remove("test_header.wav")

	// Read file back
	raw, err := os.ReadFile("test_header.wav")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if len(raw) < 44+4 {
		t.Fatalf("file too short: %d bytes", len(raw))
	}

	// RIFF
	if string(raw[0:4]) != "RIFF" {
		t.Errorf("chunkID = %s, want RIFF", string(raw[0:4]))
	}

	// ChunkSize (36 + dataSize)
	chunkSize := binary.LittleEndian.Uint32(raw[4:8])
	expectedChunkSize := uint32(36 + 4)
	if chunkSize != expectedChunkSize {
		t.Errorf("chunkSize = %d, want %d", chunkSize, expectedChunkSize)
	}

	// Format
	if string(raw[8:12]) != "WAVE" {
		t.Errorf("format = %s, want WAVE", string(raw[8:12]))
	}

	// Subchunk1ID
	if string(raw[12:16]) != "fmt " {
		t.Errorf("subchunk1ID = %s, want 'fmt '", string(raw[12:16]))
	}

	// Subchunk1Size
	subchunk1Size := binary.LittleEndian.Uint32(raw[16:20])
	if subchunk1Size != 16 {
		t.Errorf("subchunk1Size = %d, want 16", subchunk1Size)
	}

	// AudioFormat
	audioFormat := binary.LittleEndian.Uint16(raw[20:22])
	if audioFormat != 1 {
		t.Errorf("audioFormat = %d, want 1 (PCM)", audioFormat)
	}

	// NumChannels
	numChannels := binary.LittleEndian.Uint16(raw[22:24])
	if numChannels != 1 {
		t.Errorf("numChannels = %d, want 1", numChannels)
	}

	// SampleRate
	sampleRate := binary.LittleEndian.Uint32(raw[24:28])
	if sampleRate != 16000 {
		t.Errorf("sampleRate = %d, want 16000", sampleRate)
	}

	// ByteRate
	_ = binary.LittleEndian.Uint32(raw[28:32]) // skip

	// BlockAlign
	_ = binary.LittleEndian.Uint16(raw[32:34]) // skip

	// BitsPerSample
	bitsPerSample := binary.LittleEndian.Uint16(raw[34:36])
	if bitsPerSample != 16 {
		t.Errorf("bitsPerSample = %d, want 16", bitsPerSample)
	}

	// Subchunk2ID
	if string(raw[36:40]) != "data" {
		t.Errorf("subchunk2ID = %s, want 'data'", string(raw[36:40]))
	}

	// Subchunk2Size
	subchunk2Size := binary.LittleEndian.Uint32(raw[40:44])
	if subchunk2Size != 4 {
		t.Errorf("subchunk2Size = %d, want 4", subchunk2Size)
	}

	// Verify data bytes
	if raw[44] != 0x01 || raw[45] != 0x02 || raw[46] != 0x03 || raw[47] != 0x04 {
		t.Errorf("data = %x, want 01020304", raw[44:48])
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
