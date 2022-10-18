package voice

import (
	"sync"

	"github.com/gen2brain/malgo"
)

var AudioDevice = &audioDevice{}

// `audioDevice` uses Golang binding of 'miniaudio' to get voice input from microphone
// and output through callback
type audioDevice struct {
	mu sync.Mutex

	ctx *malgo.AllocatedContext
	dev *malgo.Device

	callback func([]byte)
}

// The `malgo.InitDevice()` costs 300+ milliseconds,
// some early words may lost if speak right after entering speech mode,
// so instead of start/stop repeatedly, just leave it on.
func (a *audioDevice) StartCapture() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.dev != nil { // already started
		return nil
	}

	var e error

	a.ctx, e = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		// fmt.Printf("LOG <%v>\n", message)
	})
	if e != nil {
		return e
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture) // only capture
	deviceConfig.Capture.Format = malgo.FormatS16            // s16
	deviceConfig.Capture.Channels = 1                        // 1: mono
	deviceConfig.SampleRate = 16000                          // 8k is too low

	a.dev, e = malgo.InitDevice(a.ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: func(_, pSample []byte, _ uint32) {
			a.callback(pSample) // send audio bytes to Callback
		},
	})
	if e != nil {
		a.ctx.Uninit()
		a.ctx.Free()

		return e
	}

	return a.dev.Start()
}
func (a *audioDevice) SetCallback(fn func([]byte)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callback = fn
}
