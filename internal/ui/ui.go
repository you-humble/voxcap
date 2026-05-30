package ui

// Event represents a user action.
type Event int

const (
	EventToggle  Event = iota // Space
	EventSave                 // S
	EventDiscard              // R
	EventMix                  // M
	EventQuit                 // Q
)

// UI abstracts the user interface.
type UI interface {
	// Init initializes the UI. Called once at startup.
	Init() error
	// Close cleans up the UI. Called once at shutdown.
	Close()
	// WaitEvent blocks until the next user event.
	WaitEvent() (Event, error)
	// ShowStatus updates the status line.
	ShowStatus(status Status)
	// ShowResults displays final results.
	ShowResults(results []FileResult)
}

// Status represents the current recording status.
type Status int

const (
	StatusReady Status = iota
	StatusRecording
	StatusPaused
	StatusSaved
	StatusDiscarded
	StatusMixed
)

// FileResult holds info about a saved file.
type FileResult struct {
	Name string
	Size int64
}
