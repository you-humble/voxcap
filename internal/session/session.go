package session

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gen2brain/malgo"

	"github.com/you-humble/voxcap/internal/config"
	"github.com/you-humble/voxcap/internal/mix"
	"github.com/you-humble/voxcap/internal/recorder"
	"github.com/you-humble/voxcap/internal/transcribe"
	"github.com/you-humble/voxcap/internal/ui"
	"github.com/you-humble/voxcap/internal/wav"
)

type State int

const (
	StateIdle State = iota
	StateRecording
	StatePaused
)

type Session struct {
	ctx           *malgo.AllocatedContext
	cfg           *config.Config
	state         State
	recs          []*recorder.Recorder
	ui            ui.UI
	mixing        bool
	lastMixInput  string
	lastMixOutput string
}

func New(ctx *malgo.AllocatedContext, cfg *config.Config, u ui.UI) *Session {
	return &Session{ctx: ctx, cfg: cfg, ui: u, state: StateIdle}
}

func (s *Session) HandleEvent(event ui.Event) (quit bool) {
	if s.mixing {
		return false
	}

	switch event {
	case ui.EventToggle:
		s.toggle()
	case ui.EventSave:
		s.save()
	case ui.EventDiscard:
		s.discard()
	case ui.EventMix:
		s.mixLatest()
	case ui.EventTranscribe:
		s.transcribeLatest()
	case ui.EventQuit:
		s.quit()
		return true
	}
	return false
}

func (s *Session) Results() []ui.FileResult {
	var results []ui.FileResult
	for _, d := range s.cfg.Devices {
		info, _ := os.Stat(d.OutputFile)
		if info != nil {
			results = append(results, ui.FileResult{
				Name: d.OutputFile,
				Size: info.Size(),
			})
		}
	}
	return results
}

// MixLatest mixes the most recent loopback and mic recordings.
func (s *Session) MixLatest() (string, error) {
	return mix.MixLatest("output/loopback_*.wav", "output/mic_*.wav")
}

func (s *Session) toggle() {
	switch s.state {
	case StateIdle:
		if err := s.start(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		s.state = StateRecording
		s.ui.ShowStatus(ui.StatusRecording)

	case StateRecording:
		s.pause()
		s.state = StatePaused
		s.ui.ShowStatus(ui.StatusPaused)

	case StatePaused:
		s.resume()
		s.state = StateRecording
		s.ui.ShowStatus(ui.StatusRecording)
	}
}

func (s *Session) TranscribeLatest() (string, error) {
	mixedFiles, _ := filepath.Glob("output/mixed_*.wav")
	if len(mixedFiles) == 0 {
		return "", fmt.Errorf("no mixed files to transcribe")
	}

	latest := mixedFiles[len(mixedFiles)-1]

	s.mixing = true
	defer func() { s.mixing = false }()

	fmt.Print("\r⏳ Transcribing...                ")

	text, err := transcribe.Whisper(latest, s.cfg.WhisperPath, s.cfg.ModelPath)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (s *Session) save() {
	if s.state == StateRecording {
		s.pause()
	}
	for _, r := range s.recs {
		if err := r.Save(); err != nil {
			log.Printf("Save error: %v", err)
		}
	}
	s.recs = nil
	s.state = StateIdle
	s.ui.ShowStatus(ui.StatusSaved)
	s.ui.ShowResults(s.Results())
	s.ui.ShowStatus(ui.StatusReady)
}

func (s *Session) discard() {
	if s.state == StateRecording {
		s.pause()
	}
	for _, r := range s.recs {
		if err := r.Discard(); err != nil {
			log.Printf("Discard error: %v", err)
		}
	}
	for _, d := range s.cfg.Devices {
		if err := os.Remove(d.OutputFile); err != nil {
			log.Printf("Remove error: %v", err)
		}
	}
	s.recs = nil
	s.state = StateIdle
	s.ui.ShowStatus(ui.StatusDiscarded)
	s.ui.ShowStatus(ui.StatusReady)
}

func (s *Session) mixLatest() {
	// Don't mix if already mixed these files
	loopFiles, _ := filepath.Glob("output/loopback_*.wav")
	micFiles, _ := filepath.Glob("output/mic_*.wav")

	if len(loopFiles) == 0 || len(micFiles) == 0 {
		fmt.Printf("No files to mix\n")
		return
	}

	loopFile := loopFiles[len(loopFiles)-1]
	micFile := micFiles[len(micFiles)-1]

	// Deduplicate: don't mix the same pair twice
	pairKey := loopFile + "|" + micFile
	if pairKey == s.lastMixInput {
		// fmt.Printf("Already mixed: %s\n", s.lastMixOutput)
		return
	}

	fmt.Print("\r⏳ Mixing...                      ")
	s.mixing = true
	defer func() { s.mixing = false }()

	output := fmt.Sprintf("output/mixed_%s.wav", time.Now().Format("20060102_150405"))
	if err := mix.Mono(loopFile, micFile, output); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	s.lastMixInput = pairKey
	s.lastMixOutput = output

	s.ui.ShowStatus(ui.StatusMixed)
	fmt.Printf("   %s\n", output)
	s.ui.ShowStatus(ui.StatusReady)
}

func (s *Session) transcribeLatest() {
	s.ui.ShowStatus(ui.StatusTranscribing)

	text, err := s.TranscribeLatest()
	if err != nil {
		fmt.Printf("\rError: %v\n", err)
		s.ui.ShowStatus(ui.StatusReady)
		return
	}

	fmt.Print("\r✅ Transcript                    \n\n")
	fmt.Println(text)
	fmt.Println()
	s.ui.ShowStatus(ui.StatusReady)
}

func (s *Session) quit() {
	if s.state == StateRecording || s.state == StatePaused {
		s.discard()
	}
}

func (s *Session) start() error {
	for _, devCfg := range s.cfg.Devices {
		ext := filepath.Ext(devCfg.OutputFile)
		base := devCfg.OutputFile[:len(devCfg.OutputFile)-len(ext)]
		filename := fmt.Sprintf("%s_%s%s", base, time.Now().Format("20060102_150405"), ext)

		w, err := wav.New(filename, s.cfg.SampleRate, s.cfg.Channels, s.cfg.BitsPerSample)
		if err != nil {
			return fmt.Errorf("wav: %w", err)
		}

		devCfg.OutputFile = filename

		dt := malgo.Capture
		if devCfg.Type == "loopback" {
			dt = malgo.Loopback
		}

		r, err := recorder.NewRecorder(s.ctx, dt, w)
		if err != nil {
			return fmt.Errorf("%s: %w", devCfg.Type, err)
		}

		if err := r.Start(); err != nil {
			return fmt.Errorf("start %s: %w", devCfg.Type, err)
		}

		s.recs = append(s.recs, r)
	}
	return nil
}

func (s *Session) pause() {
	for _, r := range s.recs {
		if err := r.Pause(); err != nil {
			log.Printf("Pause error: %v", err)
		}
	}
}

func (s *Session) resume() {
	for _, r := range s.recs {
		if err := r.Resume(); err != nil {
			log.Printf("Resume error: %v", err)
		}
	}
}
