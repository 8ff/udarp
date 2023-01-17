package fskGenerator

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func float642uint16(f float64) uint16 {
	f = f * 32768
	if f > 32767 {
		f = 32767
	}
	if f < -32768 {
		f = -32768
	}
	return uint16(f)
}

func FlexFsk(sampleRate, bitDurationMS int, toneFreq float64, bits []int) []byte {
	var freq float64
	var outputBuffer []byte

	/* NOTES
	   b - shapes what the top of the curve looks like
	   t := bellWidth*(float64(i)-shiftLRValue*numberOfSamples)/numberOfSamples
	*/

	b := 1.0
	pi := math.Pi
	numberOfSamples := float64((sampleRate / 1000) * (bitDurationMS))

	for _, bit := range bits {
		switch bit {
		case 0:
			freq = 0
		case 1:
			freq = toneFreq
		}

		for i := 0; i < (sampleRate/1000)*(bitDurationMS); i++ {
			t := 2 * (float64(i) - 0.5*numberOfSamples) / numberOfSamples
			c := pi * math.Sqrt(2.0/math.Log(2.0))
			gfsk_pulse := 0.5 * (math.Erf(c*b*(t+0.5)) - math.Erf(c*b*(t-0.5)))
			sample := gfsk_pulse * math.Sin((2*math.Pi)/float64(sampleRate)*freq*float64(i)*1.0)
			// fmt.Fprintf(os.Stderr, "%d,%f\n", i, sample)

			var buf [2]byte
			binary.LittleEndian.PutUint16(buf[:], float642uint16(sample))
			for i := 0; i < len(buf); i++ {
				outputBuffer = append(outputBuffer, buf[i])
			}

		}
	}

	return outputBuffer
}

func Fsk(bits []int) []byte {
	var freq float64
	sampleRate := 44100
	bitDurationMS := 2000
	var outputBuffer []byte

	/* NOTES
	   b - shapes what the top of the curve looks like
	   t := bellWidth*(float64(i)-shiftLRValue*numberOfSamples)/numberOfSamples
	*/

	b := 1.0
	pi := math.Pi
	numberOfSamples := float64((sampleRate / 1000) * (bitDurationMS))

	for _, bit := range bits {
		switch bit {
		case 0:
			freq = 0
		case 1:
			freq = 1520
		}

		// 50 51 52 53 54 |55| 56 57 58 59 60
		// 61 62 63 64 65 |66| 67 68 69 70 71

		// 00 01 02 03 04 |05| 06 07 08 09 10
		// 11 12 13 14 15 |16| 17 18 19 20 21

		for i := 0; i < (sampleRate/1000)*(bitDurationMS); i++ {
			t := 2 * (float64(i) - 0.5*numberOfSamples) / numberOfSamples
			c := pi * math.Sqrt(2.0/math.Log(2.0))
			gfsk_pulse := 0.5 * (math.Erf(c*b*(t+0.5)) - math.Erf(c*b*(t-0.5)))
			sample := gfsk_pulse * math.Sin((2*math.Pi)/float64(sampleRate)*freq*float64(i)*1.0)
			//					sample := math.Sin((2*math.Pi)/float64(sampleRate)*freq*float64(i)*1.0)
			//		fmt.Printf("%d,%f\n", i, sample)
			fmt.Fprintf(os.Stderr, "%d,%f\n", i, sample)
			//fmt.Fprintf(os.Stderr, "%d,%f\n", i, gfsk_pulse)

			var buf [2]byte
			binary.LittleEndian.PutUint16(buf[:], float642uint16(sample))
			for i := 0; i < len(buf); i++ {
				outputBuffer = append(outputBuffer, buf[i])
			}

		}
	}

	return outputBuffer
}
