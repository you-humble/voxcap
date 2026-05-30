package recorder

import (
	"github.com/gen2brain/malgo"
	"github.com/you-humble/voxcap/internal/wav"
)

type Recorder struct {
	device *malgo.Device
	writer *wav.Writer
}

func NewRecorder(
	ctx *malgo.AllocatedContext,
	deviceType malgo.DeviceType,
	writer *wav.Writer,
) (*Recorder, error) {
	cfg := malgo.DefaultDeviceConfig(deviceType)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = 16000

	r := &Recorder{writer: writer}

	onData := func(_, input []byte, _ uint32) {
		r.writer.WritePCM(input)
	}

	device, err := malgo.InitDevice(
		ctx.Context,
		cfg,
		malgo.DeviceCallbacks{Data: onData},
	)
	if err != nil {
		return nil, err
	}
	r.device = device
	return r, nil
}

func (r *Recorder) Start() error {
	return r.device.Start()
}

// Pause stops the device but keeps the file open for appending.
func (r *Recorder) Pause() error {
	return r.device.Stop()
}

// Resume restarts the device, appending to the same file.
func (r *Recorder) Resume() error {
	return r.device.Start()
}

// Save stops the device and finalizes the WAV file.
func (r *Recorder) Save() error {
	r.device.Stop()
	return r.writer.Close()
}

// Discard stops the device and closes the file without fixing header.
func (r *Recorder) Discard() error {
	r.device.Stop()
	return r.writer.CloseWithoutHeader()
}
