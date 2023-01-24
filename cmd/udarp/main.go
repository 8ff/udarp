package main

/*
math.Round(freq) has been added as a test, maybe remove it if not needed
Figure out a way to detect preamble and based on that preamble figure out the length of the tones
*/

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
	"sort"
	"strconv"

	"github.com/8ff/udarp/pkg/audio"
	"github.com/8ff/udarp/pkg/buffer"
	"github.com/8ff/udarp/pkg/misc"

	"github.com/gen2brain/malgo"
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

type Config struct {
	HTTP_Listen_Addr string
	StdinDebug       bool
	PlaybackDevice   *malgo.DeviceInfo
	CaptureDevice    *malgo.DeviceInfo
	WindowSize       int
	WindowsInFrame   int
	SampleRate       uint32
	Freq             struct {
		Lo float64 // Low frequency filter
		Hi float64 // High frequency filter
	}
	TimeSlotChannel chan map[float64][]float64
}

func (conf *Config) toneDecoder() {
	// 	//	// Get PCM data on stdin, processes it and pushes it on the channel
	// 	// We expect audio to be S16_LE

	// spectrum := make(map[int64]map[float64][]float64)
	sampleRate := 44100
	// The smaller the window size, the less accurate frequency is
	windowSize := 1000 // in ms
	windowSamples := int(float32(sampleRate) * float32(windowSize) / 1000.0)

	/*
		If spectral width is equal to 1, then angle is either 0, 180, -180
		It should be possible to use an fft size of sampleRate

	*/

	//	fftSize := 16384
	//	fftSize := 32768
	//	fftSize := 65536
	// fftSize := 40960
	fftSize := 131072
	//	fftSize := int(math.Pow(2, math.Ceil(math.Log2(float64(windowSamples)))))
	w := make([]float64, fftSize)

	spectralWidth := float64(sampleRate) / float64(fftSize)
	fmt.Fprintf(os.Stderr, "SAMPLE_RATE: %d\n", sampleRate)
	fmt.Fprintf(os.Stderr, "WINDOW_SIZE: %v\n", windowSize)
	fmt.Fprintf(os.Stderr, "WINDOW_SAMPLES: %d[samples]\n", windowSamples)
	fmt.Fprintf(os.Stderr, "FFT_SIZE: %d\n", fftSize)
	fmt.Fprintf(os.Stderr, "SPECTRAL_WIDTH: %v[hertz]\n", spectralWidth)

	// Initialize the context.
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// buffer := new(bytes.Buffer)
	var buffer buffer.Buffer
	buffer.Grow(10000) // Set the buffer size to 10000 bytes
	fftBlockCounter := 0
	t := int64(0)
	// timeSlot := make([]map[float64][]float64, 0)

	if !conf.StdinDebug { // If STDIN debug is enabled, we don't want to start the device
		// Configure the device.
		deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
		deviceConfig.Capture.DeviceID = conf.CaptureDevice.ID.Pointer()
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = conf.SampleRate
		deviceConfig.Alsa.NoMMap = 1

		// Callback which is called when the device receives frames.
		onRecvFrames := func(audioSample2, audioSample []byte, framecount uint32) {
			buffer.Write(audioSample)
		}

		misc.Log("info", ">> [Recording...]")
		captureCallbacks := malgo.DeviceCallbacks{
			Data: onRecvFrames,
		}

		// Initialize the device.
		device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = device.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		// If we are in debug mode, we read from STDIN
		conf.WindowsInFrame = 1000 // Set this to something high, we will then subtract the number of windows we read from STDIN

		r := bufio.NewReader(os.Stdin)
		stdinBuffer := make([]byte, 100)
		for {
			bytesRead, err := r.Read(stdinBuffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s.", err)
				os.Exit(2)
			}
			buffer.Write(stdinBuffer[:bytesRead])
		}
	}

	for {
		///////////////////////////////////////////////////////////
		//	fmt.Printf("%d\n", buffer.Bytes())

		for {
			if buffer.Len() >= 2 {
				actual_len := buffer.Len()
				// Below code fills "w" with samples, once its filled (fftBlockCounter == windowSamples) it runs the fft, then rewrites "w" from 0
				// This means that we are doing fft on blocks, rather replacing a part of samples
				// Maybe replace 3% of samples and do the fft, take a look at graphs to see if it makes a difference
				// GET AROUND THIS
				if fftBlockCounter < windowSamples { // Fill the window
					data := buffer.Next(2)
					if len(data) < 2 || len(w) < fftBlockCounter {
						// TODO: Remove Debug
						fmt.Printf("fftBlockCounter: %d | windowSamples: %d | actual_len: %d | data: %v | len(w): %d | len(data): %d | buffer.Len(): %d\n", fftBlockCounter, windowSamples, actual_len, data, len(w), len(data), buffer.Len())
					}
					w[fftBlockCounter] = float64(int16(binary.LittleEndian.Uint16(data))) / 32768.0
					fftBlockCounter++
				} else {
					// Window is filled, process FFT for the window
					t += int64(windowSize)
					window.Apply(w, window.Hamming)
					c := fft.FFTReal(w)

					// spectrum[t] = make(map[float64][]float64)
					windowAngle := 0.0
					freqSlice := make(map[float64][]float64, 0)

					for i := 0; i < len(c)/2; i++ {
						freq := math.Round(float64(i) * float64(sampleRate) / float64(fftSize))
						//						freq := float64(i) * float64(sampleRate) / float64(fftSize)
						if freq < conf.Freq.Lo || freq > conf.Freq.Hi { // We only care about the frequencies we are interested in
							continue
						}

						r, angle := cmplx.Polar(c[i])
						angle *= 360.0 / (2 * math.Pi)
						if dsputils.Float64Equal(r, 0) {
							angle = 0 // (When the magnitude is close to 0, the angle is meaningless)
						}
						r = r / float64(fftSize)

						if freq == 0.0 {
							windowAngle = angle
						}

						if freq > 1512.00 && freq < 1522.00 { // TEMPORARY
							freqSlice[freq] = append(freqSlice[freq], r)
							fmt.Sprintf("%d,%f,%f,%.1f\n", t, freq, r, windowAngle) // FREQ/R/ANGLE // This is placeholder to avoid warnings about unused variables
						}
					}

					// Automatically adjust WindowsInFrame if stdinDebug is enabled
					if conf.StdinDebug {
						if buffer.Len() < int(conf.SampleRate)*2 {
							conf.WindowsInFrame = 1000 - conf.WindowsInFrame + 1
							misc.Log("debug", fmt.Sprintf("Auto adjusted WindowsInFrame to %d", conf.WindowsInFrame))
						} else {
							conf.WindowsInFrame--
						}
					}
					// Push push parsed FFT of the window to channel
					conf.TimeSlotChannel <- freqSlice
					fftBlockCounter = 0
				}
			} else {
				break
			}
		}
	}
	// device.Uninit()
}

func (conf *Config) fetchWindow() {
	frame := make([]map[float64][]float64, 0)

	// List on conf.TimeSlotChannel
	for {
		freqSlice := <-conf.TimeSlotChannel
		frame = append(frame, freqSlice)

		if len(frame) == conf.WindowsInFrame {
			conf.parseFrame(frame)
			// Shift by config.WindowsInFrame size
			frame = frame[conf.WindowsInFrame:]
		}
	}
}

func (conf *Config) parseFrame(frame []map[float64][]float64) {
	// Here we get a timeSlot size of config.WindowsInFrame, each element is a map of frequencies and 3 of their values
	var binarySlice []float64
	var chartFrames = make([][]string, 0)

	// Analyze here
	for index, freqMap := range frame {
		freqKeys := make([]float64, 0, len(freqMap))
		for k := range freqMap {
			freqKeys = append(freqKeys, k)
		}
		sort.Float64s(freqKeys)

		var channelAvg float64

		for _, freq := range freqKeys {
			rMap := freqMap[freq]
			var rSum float64
			for _, rv := range rMap {
				rSum += rv
			}
			//		fmt.Fprintf(os.Stderr, "T: %d F: %f R: %f\n", index, freq, rSum)
			//			fmt.Printf("%d,%f,%f\n", index, freq, rSum)
			channelAvg += rSum
		}
		channelAvg = channelAvg / float64(len(freqKeys))
		fmt.Printf("Q: %d,%f\n", index, channelAvg)

		// Add frame to chartData
		chartFrames = append(chartFrames, []string{strconv.Itoa(index), fmt.Sprintf("%.3f", channelAvg*1000)})
		binarySlice = append(binarySlice, channelAvg)
	}

	// Marshall chartData to json
	chartDataJson, err := json.Marshal(chartFrames)
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error marshalling chartData to json: %s", err))
	}

	// Store chartDataJson to lastFrame
	lastFrame = chartDataJson

	// Broadcast chart data using bcastWs
	bcastWs(chartDataJson)

	// TODO: This may not work in all cases, if we have a very high bit it will throw off the min/max/mid values
	// Find min/max/mid values
	max := 0.0
	min := 99999999.0
	mid := 0.0

	for _, v := range binarySlice {
		if max < v {
			max = v
		}

		if v < min {
			min = v
		}
	}
	mid = ((max - min) / 2) + min
	totalBits := 16
	bitSpacing := 2
	ambleLen := 2

	for i, _ := range binarySlice {
		bitSlice := make([]int, 0)
		floatSlice := make([]float64, 0)

		for bitIndex := 0; bitIndex < totalBits*bitSpacing; bitIndex += bitSpacing {
			if (i + bitIndex) < len(binarySlice) {
				floatSlice = append(floatSlice, binarySlice[i+bitIndex])
				if binarySlice[i+bitIndex] > mid {
					bitSlice = append(bitSlice, 1)
				} else {
					bitSlice = append(bitSlice, 0)
				}
			}
		}

		if len(bitSlice) >= totalBits {
			preamble := ""
			postamble := ""
			for i := 0; i < ambleLen; i++ {
				preamble += fmt.Sprintf("%d", bitSlice[i])
			}

			for i := len(bitSlice) - ambleLen; i < len(bitSlice); i++ {
				postamble += fmt.Sprintf("%d", bitSlice[i])
			}

			if preamble == "01" && postamble == "01" {
				fmt.Fprintf(os.Stderr, ">>>> SCORE T:%d %d\n", i, bitSlice)
			}

			fmt.Fprintf(os.Stderr, "T:%d %d\n", i, bitSlice)
		}

		//				fmt.Sprintf("T:%d [%d]\n", i, bitSlice)
		//				fmt.Printf("T:%d [%f]\n", i, floatSlice)
	}
}

// Function that checks if there is a -l flag
func (conf *Config) parseFlags() bool {
	for _, arg := range os.Args {
		switch arg {
		case "-l", "--list-devices", "--list":
			listDevices()
			os.Exit(0)
		case "--stdin":
			conf.StdinDebug = true
		}
	}
	return false
}

// Function that lists all audio devices by callign getDevices and prints an md5 hash of the deviceID
func listDevices() {
	playbackDevices, captureDevices, err := audio.GetAudioDevices()
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error getting audio devices: %s", err))
		os.Exit(1)
	}

	// Print playback devices
	fmt.Println("\x1b[32m**** Playback devices ****\x1b[0m")
	for _, device := range playbackDevices {
		fmt.Printf("ID: \x1b[32m%s\x1b[0m - Name: \x1b[32m%s\x1b[0m\n", misc.Md5HashString(device.Name()), device.Name())
	}

	// Print capture devices
	fmt.Println("\n\x1b[31m**** Capture devices ****\x1b[0m")
	for _, device := range captureDevices {
		fmt.Printf("ID: \x1b[31m%s\x1b[0m - Name: \x1b[31m%s\x1b[0m\n", misc.Md5HashString(device.Name()), device.Name())
	}
}

// Function that gets AudioDevices from environment variables
func (conf *Config) configAudioDevices() {
	playbackHash := os.Getenv("PLAYBACK_DEVICE")
	captureHash := os.Getenv("CAPTURE_DEVICE")

	// Check if audio devices are set
	if playbackHash == "" || captureHash == "" {
		misc.Log("error", "Audio devices not set")
		os.Exit(1)
	}

	var err error
	conf.PlaybackDevice, err = audio.DeviceFromHash(playbackHash)
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error getting playback device: %s", err))
		os.Exit(1)
	}

	conf.CaptureDevice, err = audio.DeviceFromHash(captureHash)
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error getting capture device: %s", err))
		os.Exit(1)
	}

	// Print device names
	misc.Log("info", fmt.Sprintf("Using [%s - %s] as playback device", playbackHash, conf.PlaybackDevice.Name()))
	misc.Log("info", fmt.Sprintf("Using [%s - %s] as capture device", captureHash, conf.CaptureDevice.Name()))

}

func main() {
	config := Config{}
	config.HTTP_Listen_Addr = "0.0.0.0:3000"
	config.Freq.Lo = 0
	config.Freq.Hi = 2500
	config.WindowSize = 1000
	config.WindowsInFrame = 10
	config.SampleRate = 44100
	config.TimeSlotChannel = make(chan map[float64][]float64)
	config.parseFlags()
	go config.fetchWindow()
	if !config.StdinDebug {
		config.configAudioDevices()
	}
	go config.serveHTTP()
	config.toneDecoder()
}
