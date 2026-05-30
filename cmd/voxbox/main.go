package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/gen2brain/malgo"
)

func main() {
	f, err := os.Create("output.wav")
	if err != nil {
		log.Fatal(err)
	}
	f.Write(make([]byte, 44))

	var dataSize uint32

	ctx, err := malgo.InitContext(
		[]malgo.Backend{malgo.BackendWasapi},
		malgo.ContextConfig{},
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Free()
	defer ctx.Uninit()

	cfg := malgo.DefaultDeviceConfig(malgo.Loopback)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = 16000

	onData := func(_, input []byte, _ uint32) {
		n, _ := f.Write(input)
		dataSize += uint32(n)
	}

	device, err := malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{Data: onData})
	if err != nil {
		log.Fatal(err)
	}

	if err := device.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("🔴 Запись... Enter для остановки")
	bufio.NewReader(os.Stdin).ReadString('\n')

	device.Uninit()
	writeWAVHeader(f, dataSize, 16000, 1, 16)
	f.Close()

	fmt.Println("✅ Сохранено: output.wav")
}

func writeWAVHeader(f *os.File, dataSize, sampleRate uint32, channels, bits uint16) {
	f.Seek(0, 0)
	byteRate := sampleRate * uint32(channels) * uint32(bits) / 8
	blockAlign := channels * bits / 8

	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(f, binary.LittleEndian, channels)
	binary.Write(f, binary.LittleEndian, sampleRate)
	binary.Write(f, binary.LittleEndian, byteRate)
	binary.Write(f, binary.LittleEndian, blockAlign)
	binary.Write(f, binary.LittleEndian, bits)
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, dataSize)
}
