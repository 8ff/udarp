package main

import (
	"fmt"
	"os"

	"github.com/8ff/udarp/pkg/audio"
	"github.com/8ff/udarp/pkg/fskGenerator"
	"github.com/gen2brain/malgo"
)

func main() {
	wave := fskGenerator.FlexFsk(44000, 200, 1520.00, []int{0, 1, 0, 1, 0, 1, 0, 1})

	defaultPlaybackDevice, err := audio.GetDefaultPlaybackDevice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.DeviceID = defaultPlaybackDevice.ID.Pointer()
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1

	err = audio.PlayWave(deviceConfig, wave)
	if err != nil {
		fmt.Println(err)
	}
}
