package recorder

import (
	"log"

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
		if _, err := r.writer.WritePCM(input); err != nil {
			log.Printf("WritePCM error: %v", err)
		}
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

func (r *Recorder) Pause() error {
	return r.device.Stop()
}

func (r *Recorder) Resume() error {
	return r.device.Start()
}

func (r *Recorder) Save() error {
	if err := r.device.Stop(); err != nil {
		return err
	}
	return r.writer.Close()
}

func (r *Recorder) Discard() error {
	if err := r.device.Stop(); err != nil {
		return err
	}
	return r.writer.CloseWithoutHeader()
}
