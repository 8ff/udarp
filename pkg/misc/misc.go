package misc

import (
	"crypto/md5"
	"fmt"
	"time"
)

// Function that calculates all possible combinations of bits given a number of bits
func GenerateBitCombinations(bits int) [][]int {
	var combinations [][]int
	for i := 0; i < (1 << uint(bits)); i++ {
		var combination []int
		for j := 0; j < bits; j++ {
			if i&(1<<uint(j)) != 0 {
				combination = append(combination, 1)
			} else {
				combination = append(combination, 0)
			}
		}
		combinations = append(combinations, combination)
	}
	return combinations
}

// Function that generates all possible codewords given a number of bits and xors them to produce an added parity bit
func GenerateCodewords(bits int) [][]int {
	var codewords [][]int
	combinations := GenerateBitCombinations(bits)
	for i := 0; i < len(combinations); i++ {
		var codeword []int
		for j := 0; j < len(combinations[i]); j++ {
			codeword = append(codeword, combinations[i][j])
		}
		codeword = append(codeword, 0)
		for j := 0; j < len(codeword); j++ {
			if j == len(codeword)-1 {
				break
			}
			codeword[len(codeword)-1] = codeword[len(codeword)-1] ^ codeword[j]
		}
		codewords = append(codewords, codeword)
	}
	return codewords
}

func Float642uint16(f float64) uint16 {
	f = f * 32768
	if f > 32767 {
		f = 32767
	}
	if f < -32768 {
		f = -32768
	}
	return uint16(f)
}

// Function which runs md5 hash on input string and returns the string
func Md5HashString(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

func Log(level, msg string) {
	switch level {
	case "info":
		fmt.Printf("\x1b[32m%s [INFO] %s\x1b[0m\n", time.Now().Format("15:04:05"), msg)
	case "error":
		fmt.Printf("\x1b[31m%s [ERROR] %s\x1b[0m\n", time.Now().Format("15:04:05"), msg)
	case "warning":
		fmt.Printf("\x1b[33m%s [WARNING] %s\x1b[0m\n", time.Now().Format("15:04:05"), msg)
	case "debug":
		fmt.Printf("\x1b[36m%s [DEBUG] %s\x1b[0m\n", time.Now().Format("15:04:05"), msg)
	default:
		fmt.Printf("%s [UNKNOWN] %s\n", time.Now().Format("15:04:05"), msg)
	}
}
