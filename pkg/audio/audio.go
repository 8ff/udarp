package audio

import (
	"bytes"
	"fmt"
	"os"

	"github.com/8ff/udarp/pkg/misc"
	"github.com/gen2brain/malgo"
)

func GetDefaultPlaybackDevice() (*malgo.DeviceInfo, error) {
	// Get default playback device.
	playbackDevices, _, err := GetAudioDevices()
	if err != nil {
		return nil, err
	}

	if len(playbackDevices) == 0 {
		return nil, fmt.Errorf("no playback devices found")
	}

	return &playbackDevices[0], nil
}

func GetDefaultCaptureDevice() (*malgo.DeviceInfo, error) {
	// Get default capture device.
	_, captureDevices, err := GetAudioDevices()
	if err != nil {
		return nil, err
	}

	if len(captureDevices) == 0 {
		return nil, fmt.Errorf("no capture devices found")
	}

	return &captureDevices[0], nil
}

// GetAudioDevices returns a list of all playback and capture devices.
func GetAudioDevices() ([]malgo.DeviceInfo, []malgo.DeviceInfo, error) {
	var playbackDevices []malgo.DeviceInfo
	var captureDevices []malgo.DeviceInfo

	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return playbackDevices, captureDevices, err
	}
	defer func() {
		_ = context.Uninit()
		context.Free()
	}()

	// Playback devices.
	infos, err := context.Devices(malgo.Playback)
	if err != nil {
		return playbackDevices, captureDevices, err
	}
	playbackDevices = infos

	// Capture devices.
	infos, err = context.Devices(malgo.Capture)
	if err != nil {
		return playbackDevices, captureDevices, err
	}
	captureDevices = infos

	return playbackDevices, captureDevices, nil
}

func DeviceFromHash(deviceHash string) (*malgo.DeviceInfo, error) {
	if deviceHash == "" {
		return nil, fmt.Errorf("device hash is empty")
	}

	// Lookup device IDs in listDevices
	playbackDevices, captureDevices, err := GetAudioDevices()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Go over all playback devices and find the one with the given ID
	for _, device := range playbackDevices {
		if misc.Md5HashString(device.ID.String()) == deviceHash {
			return &device, nil
		}
	}

	// Go over all capture devices and find the one with the given ID
	for _, device := range captureDevices {
		if misc.Md5HashString(device.ID.String()) == deviceHash {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device with hash %s not found", deviceHash)
}

func PrettyPrintDevices() {
	playbackDevices, captureDevices, err := GetAudioDevices()
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error getting audio devices: %s", err))
		os.Exit(1)
	}

	// Print playback devices
	fmt.Println("\x1b[32m**** Playback devices ****\x1b[0m")
	for _, device := range playbackDevices {
		fmt.Printf("ID: \x1b[32m%s\x1b[0m - Name: \x1b[32m%s\x1b[0m\n", misc.Md5HashString(device.ID.String()), device.Name())
	}

	// Print capture devices
	fmt.Println("\n\x1b[31m**** Capture devices ****\x1b[0m")
	for _, device := range captureDevices {
		fmt.Printf("ID: \x1b[31m%s\x1b[0m - Name: \x1b[31m%s\x1b[0m\n", misc.Md5HashString(device.ID.String()), device.Name())
	}
}

func PlayWave(deviceConfig malgo.DeviceConfig, buffer []byte) error {
	toneBuffer := new(bytes.Buffer)
	toneBuffer.Write(buffer)

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		return err
	}

	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	samplesConsumed := int64(0)
	playbackDone := make(chan bool)

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		copy(pOutputSample, toneBuffer.Next(len(pOutputSample)))
		samplesConsumed += int64(len(pOutputSample))
		if samplesConsumed > int64(len(buffer)) {
			playbackDone <- true
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		return err
	}

	done := <-playbackDone
	if done {
		fmt.Fprintf(os.Stderr, "SAMPLES_CONSUMED: %d PLAYBACK_DONE\n", samplesConsumed)
	}
	return nil
}
