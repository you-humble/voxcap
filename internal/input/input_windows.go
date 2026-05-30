//go:build windows

package input

import (
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle     = kernel32.NewProc("GetStdHandle")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")
)

const (
	stdInputHandle     = ^uint32(10) + 1 // STD_INPUT_HANDLE
	enableLineInput    = 0x0002
	enableEchoInput    = 0x0004
	enableProcessInput = 0x0001
	keyEvent           = 0x0001
)

type inputRecord struct {
	EventType uint16
	_         [2]byte
	KeyEvent  keyEventRecord
}

type keyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

// Keyboard reads keys in raw mode on Windows.
type Keyboard struct {
	handle  syscall.Handle
	oldMode uint32
}

// NewKeyboard creates a raw-mode keyboard reader.
func NewKeyboard() (*Keyboard, error) {
	handle, _, err := procGetStdHandle.Call(uintptr(stdInputHandle))
	if err.(syscall.Errno) != 0 {
		return nil, err.(syscall.Errno)
	}

	k := &Keyboard{handle: syscall.Handle(handle)}

	procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&k.oldMode)))

	newMode := k.oldMode &^ (enableLineInput | enableEchoInput | enableProcessInput)
	procSetConsoleMode.Call(handle, uintptr(newMode))

	return k, nil
}

// Close restores original console mode.
func (k *Keyboard) Close() {
	procSetConsoleMode.Call(uintptr(k.handle), uintptr(k.oldMode))
}

// ReadKey blocks until keypress, returns rune.
func (k *Keyboard) ReadKey() (rune, error) {
	for {
		var record inputRecord
		var read uint32

		ret, _, err := procReadConsoleInput.Call(
			uintptr(k.handle),
			uintptr(unsafe.Pointer(&record)),
			1,
			uintptr(unsafe.Pointer(&read)),
		)
		if ret == 0 {
			return 0, err
		}

		if record.EventType == keyEvent && record.KeyEvent.KeyDown != 0 {
			return rune(record.KeyEvent.UnicodeChar), nil
		}
	}
}
