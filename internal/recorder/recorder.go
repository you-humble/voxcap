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
	cfg.Capture.Format = malgo.FormatS16 // 16-bit samples
	cfg.Capture.Channels = 1             // mono
	cfg.SampleRate = 16000               // 16 kHz (formaat Whisper)

	r := &Recorder{writer: writer}

	// Callback, whitch malgo runs in a separate flow,
	// when the new sample is ready.
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

func (r *Recorder) Stop() error {
	r.device.Uninit()
	return r.writer.Close()
}
