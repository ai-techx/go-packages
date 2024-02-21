package input

import (
	"fmt"
	"github.com/gen2brain/malgo"
)

type WhisperInput struct {
	device *malgo.Device
}

// NewWhisperInput returns a new instance of whisperInput.
func NewWhisperInput() Input {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})

	if err != nil {
		fmt.Printf("Failed to initialize context: %v\n", err)
		return nil
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)

	var capturedSampleCount uint32
	pCapturedSamples := make([]byte, 0)
	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {

		sampleCount := framecount * deviceConfig.Capture.Channels * sizeInBytes

		newCapturedSampleCount := capturedSampleCount + sampleCount

		pCapturedSamples = append(pCapturedSamples, pSample...)

		capturedSampleCount = newCapturedSampleCount

	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)

	return &WhisperInput{
		device: device,
	}
}

// Run runs the whisper input.
func (w *WhisperInput) Run(yield func(input string, err error) bool) {

}

func (w *WhisperInput) record() {

}
