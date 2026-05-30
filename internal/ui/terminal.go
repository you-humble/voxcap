package ui

import (
	"fmt"

	"github.com/you-humble/voxcap/internal/input"
)

// Terminal implements UI for console.
type Terminal struct {
	kb input.KeyReader
}

// NewTerminal creates a terminal UI.
func NewTerminal(kb input.KeyReader) *Terminal {
	return &Terminal{kb: kb}
}

func (t *Terminal) Init() error {
	fmt.Println("🎤 VoxCap")
	fmt.Println("   [Space] Start/Pause/Resume")
	fmt.Println("   [S]     Save")
	fmt.Println("   [R]     Discard")
	fmt.Println("   [M]     Mix latest")
	fmt.Println("   [Q]     Quit")
	fmt.Println()
	return nil
}

func (t *Terminal) Close() {
	// keyboard is closed by main
}

func (t *Terminal) WaitEvent() (Event, error) {
	for {
		key, err := t.kb.ReadKey()
		if err != nil {
			return 0, err
		}

		switch key {
		case ' ':
			return EventToggle, nil
		case 's', 'S':
			return EventSave, nil
		case 'r', 'R':
			return EventDiscard, nil
		case 'q', 'Q':
			return EventQuit, nil
		case 'm', 'M':
			return EventMix, nil
		}
	}
}

func (t *Terminal) ShowStatus(status Status) {
	switch status {
	case StatusReady:
		fmt.Print("\r                                    \r")
	case StatusRecording:
		fmt.Print("\r🔴 Recording... (Space=pause, S=save, R=reset)  ")
	case StatusPaused:
		fmt.Print("\r⏸️  Paused... (Space=resume, S=save, R=reset)    ")
	case StatusSaved:
		fmt.Print("\r✅ Saved                          \n\n")
	case StatusMixed:
		fmt.Print("\r✅ Mixed                          \n\n")
	case StatusDiscarded:
		fmt.Print("\r🗑️  Discarded                      \n\n")
	}
}

func (t *Terminal) ShowResults(results []FileResult) {
	for _, r := range results {
		fmt.Printf("   %s: %d bytes\n", r.Name, r.Size)
	}
	fmt.Println()
}
