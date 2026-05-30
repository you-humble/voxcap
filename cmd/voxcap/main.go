//go:build windows

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gen2brain/malgo"

	"github.com/you-humble/voxcap/internal/config"
	"github.com/you-humble/voxcap/internal/input"
	"github.com/you-humble/voxcap/internal/session"
	"github.com/you-humble/voxcap/internal/ui"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
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

	kb, err := input.NewKeyboard()
	if err != nil {
		log.Fatal(err)
	}
	defer kb.Close()

	terminal := ui.NewTerminal(kb)
	terminal.Init()

	sess := session.New(ctx, cfg, terminal)

	for {
		event, err := terminal.WaitEvent()
		if err != nil {
			continue
		}

		if quit := sess.HandleEvent(event); quit {
			fmt.Println("\n👋 Bye")
			return
		}
	}
}
