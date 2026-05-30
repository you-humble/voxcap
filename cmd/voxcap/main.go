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
	configPath := flag.String("config", "", "Path to config file (default: $VOXCAP_CONFIG or configs/config.json)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	recorders, err := createRecorders(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	startAll(recorders)

	fmt.Println("🔴 Recording... Press Enter to stop")
	bufio.NewReader(os.Stdin).ReadString('\n')

	stopAll(recorders)

	printResults(cfg)
}

// createRecorders builds a Recorder for each device in config.
func createRecorders(ctx *malgo.AllocatedContext, cfg *config.Config) ([]*recorder.Recorder, error) {
	var recorders []*recorder.Recorder

	for _, devCfg := range cfg.Devices {
		wavWriter, err := wav.New(
			devCfg.OutputFile,
			cfg.SampleRate,
			cfg.Channels,
			cfg.BitsPerSample,
		)
		if err != nil {
			return nil, err
		}

		deviceType := toDeviceType(devCfg.Type)

		rec, err := recorder.NewRecorder(ctx, deviceType, wavWriter)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", devCfg.Type, err)
		}
		recorders = append(recorders, rec)
	}

	return recorders, nil
}

// toDeviceType maps config string to malgo.DeviceType.
func toDeviceType(t string) malgo.DeviceType {
	switch t {
	case "loopback":
		return malgo.Loopback
	default:
		return malgo.Capture // microphone and anything else
	}
}

// startAll starts all recorders, fatals on first error.
func startAll(recorders []*recorder.Recorder) {
	for _, rec := range recorders {
		if err := rec.Start(); err != nil {
			log.Fatal(err)
		}
	}
}

// stopAll stops all recorders, logs errors but continues.
func stopAll(recorders []*recorder.Recorder) {
	for _, rec := range recorders {
		if err := rec.Stop(); err != nil {
			log.Printf("Error stopping recorder: %v", err)
		}
	}
}

// printResults shows saved file sizes.
func printResults(cfg *config.Config) {
	fmt.Println("✅ Recording saved")
	for _, devCfg := range cfg.Devices {
		info, _ := os.Stat(devCfg.OutputFile)
		if info != nil {
			fmt.Printf("   %s: %d bytes\n", devCfg.OutputFile, info.Size())
		}
	}
}
