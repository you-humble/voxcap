package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/you-humble/voxcap/internal/config"
	"github.com/you-humble/voxcap/internal/recorder"
	"github.com/you-humble/voxcap/internal/wav"
)

func main() {
	configPath := flag.String("config", "", "Path to config file (default: $VOXCAP_CONFIG or config/config.json)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Using config: %s\n", cfg.OutputFile)

	var deviceType malgo.DeviceType
	switch cfg.DeviceType {
	case "loopback":
		deviceType = malgo.Loopback
	default:
		log.Fatalf("Unsupported device type: %s", cfg.DeviceType)
	}

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

	// Create a WAV-file for loop
	wavWriter, err := wav.New(
		cfg.OutputFile,
		cfg.SampleRate,
		cfg.Channels,
		cfg.BitsPerSample,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a recorder for loopback (system sound)
	rec, err := recorder.NewRecorder(ctx, deviceType, wavWriter)
	if err != nil {
		log.Fatal(err)
	}

	// Start capture
	if err := rec.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("🔴 Recording... Press Enter to stop")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// Stop capture and filalize the WAV
	if err := rec.Stop(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Saved: %s\n", cfg.OutputFile)
}
