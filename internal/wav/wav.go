package wav

import (
	"encoding/binary"
	"fmt"
	"os"
)

type Writer struct {
	file          *os.File
	dataSize      uint32
	sampleRate    uint32
	channels      uint16
	bitsPerSample uint16
}

func New(filename string, sampleRate uint32, channels, bitsPerSample uint16) (*Writer, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	f.Write(make([]byte, 44))
	return &Writer{
		file:          f,
		sampleRate:    sampleRate,
		channels:      channels,
		bitsPerSample: bitsPerSample,
	}, nil
}

// WritePCM write raw PCM-samples
func (w *Writer) WritePCM(data []byte) (int, error) {
	n, err := w.file.Write(data)
	w.dataSize += uint32(n)
	return n, err
}

// Close completes the header and close the file
func (w *Writer) Close() error {
	if err := w.writeWAVHeader(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

func (w *Writer) CloseWithoutHeader() error {
	return w.file.Close()
}

func (w *Writer) writeWAVHeader() error {
	if _, err := w.file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek to start: %w", err)
	}

	byteRate := w.sampleRate * uint32(w.channels) * uint32(w.bitsPerSample) / 8
	blockAlign := w.channels * w.bitsPerSample / 8

	writeBytes := func(data []byte) error {
		_, err := w.file.Write(data)
		return err
	}

	writeBinary := func(v interface{}) error {
		return binary.Write(w.file, binary.LittleEndian, v)
	}

	if err := writeBytes([]byte("RIFF")); err != nil {
		return fmt.Errorf("write 'RIFF': %w", err)
	}
	if err := writeBinary(uint32(36 + w.dataSize)); err != nil {
		return fmt.Errorf("write riff size: %w", err)
	}
	if err := writeBytes([]byte("WAVE")); err != nil {
		return fmt.Errorf("write 'WAVE': %w", err)
	}
	if err := writeBytes([]byte("fmt ")); err != nil {
		return fmt.Errorf("write 'fmt ': %w", err)
	}
	if err := writeBinary(uint32(16)); err != nil {
		return fmt.Errorf("write fmt chunk size: %w", err)
	}
	if err := writeBinary(uint16(1)); err != nil {
		return fmt.Errorf("write audio format (PCM): %w", err)
	}
	if err := writeBinary(w.channels); err != nil {
		return fmt.Errorf("write channels: %w", err)
	}
	if err := writeBinary(w.sampleRate); err != nil {
		return fmt.Errorf("write sample rate: %w", err)
	}
	if err := writeBinary(byteRate); err != nil {
		return fmt.Errorf("write byte rate: %w", err)
	}
	if err := writeBinary(blockAlign); err != nil {
		return fmt.Errorf("write block align: %w", err)
	}
	if err := writeBinary(w.bitsPerSample); err != nil {
		return fmt.Errorf("write bits per sample: %w", err)
	}
	if err := writeBytes([]byte("data")); err != nil {
		return fmt.Errorf("write 'data': %w", err)
	}
	if err := writeBinary(w.dataSize); err != nil {
		return fmt.Errorf("write data size: %w", err)
	}

	return nil
}
