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

	if _, err := f.Write(make([]byte, 44)); err != nil {
		f.Close()
		return nil, err
	}

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

	write := func(v interface{}) error {
		return binary.Write(w.file, binary.LittleEndian, v)
	}

	// RIFF header
	if _, err := w.file.Write([]byte("RIFF")); err != nil {
		return fmt.Errorf("write 'RIFF': %w", err)
	}
	if err := write(uint32(36) + w.dataSize); err != nil {
		return fmt.Errorf("write riff size: %w", err)
	}
	if _, err := w.file.Write([]byte("WAVE")); err != nil {
		return fmt.Errorf("write 'WAVE': %w", err)
	}

	// fmt subchunk
	if _, err := w.file.Write([]byte("fmt ")); err != nil {
		return fmt.Errorf("write 'fmt ': %w", err)
	}
	if err := write(uint32(16)); err != nil {
		return fmt.Errorf("write fmt chunk size: %w", err)
	}
	if err := write(uint16(1)); err != nil {
		return fmt.Errorf("write audio format: %w", err)
	}
	if err := write(w.channels); err != nil {
		return fmt.Errorf("write channels: %w", err)
	}
	if err := write(w.sampleRate); err != nil {
		return fmt.Errorf("write sample rate: %w", err)
	}
	if err := write(byteRate); err != nil {
		return fmt.Errorf("write byte rate: %w", err)
	}
	if err := write(blockAlign); err != nil {
		return fmt.Errorf("write block align: %w", err)
	}
	if err := write(w.bitsPerSample); err != nil {
		return fmt.Errorf("write bits per sample: %w", err)
	}

	// data subchunk
	if _, err := w.file.Write([]byte("data")); err != nil {
		return fmt.Errorf("write 'data': %w", err)
	}
	if err := write(w.dataSize); err != nil {
		return fmt.Errorf("write data size: %w", err)
	}

	return nil
}
