package corrupt

import (
	"math/rand"
)

// Function that flips a bit at position pos
func FlipBit(a byte, pos uint) byte {
	a ^= (1 << pos)
	return a
}

// Function that takes in a string of 1/0s and num to flip randomly, it flips bits without repeating
func FlipStringBits(s string, num int) string {
	flips := make(map[int]bool)
	for i := 0; i < num; i++ {
		flip := rand.Intn(len(s))
		_, ok := flips[flip]
		if ok {
			i--
			continue
		}
		flips[flip] = true
	}
	for i := 0; i < len(s); i++ {
		_, ok := flips[i]
		if ok {
			if s[i] == '0' {
				s = s[:i] + "1" + s[i+1:]
			} else {
				s = s[:i] + "0" + s[i+1:]
			}
		}
	}
	return s
}

// Function that takes in []int of 1/0s and num to flip randomly, it flips bits without repeating
func FlipIntBits(a []int, num int) []int {
	flips := make(map[int]bool)
	b := make([]int, len(a))
	copy(b, a)
	for i := 0; i < num; i++ {
		flip := rand.Intn(len(a))
		_, ok := flips[flip]
		if ok {
			i--
			continue
		}
		flips[flip] = true
	}
	for i := 0; i < len(a); i++ {
		_, ok := flips[i]
		if ok {
			if a[i] == 0 {
				b[i] = 1
			} else {
				b[i] = 0
			}
		}
	}
	return b
}

// Function that takes in []byte and num to flip randomly, it flips bits without repeating
func FlipByteBits(a []byte, num int) []byte {
	flips := make(map[int]bool)
	b := make([]byte, len(a))
	for i := 0; i < num; i++ {
		flip := rand.Intn(len(a))
		_, ok := flips[flip]
		if ok {
			i--
			continue
		}
		flips[flip] = true
	}
	for i := 0; i < len(a); i++ {
		_, ok := flips[i]
		if ok {
			b[i] = FlipBit(a[i], uint(rand.Intn(8)))
		}
	}
	return b
}

// Function that compares two []byte and returns the number of different bits
func CompareByteBits(a []byte, b []byte) int {
	var diff int
	for i := 0; i < len(a); i++ {
		for j := 0; j < 8; j++ {
			if (a[i]>>j)&1 != (b[i]>>j)&1 {
				diff++
			}
		}
	}
	return diff
}

// Function that compares two []int and returns the number of different bits
func CompareIntBits(a []int, b []int) int {
	var diff int
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			diff++
		}
	}
	return diff
}

// Function that compares two strings of 1/0s and returns the number of different bits
func CompareStringBits(a string, b string) int {
	var diff int
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			diff++
		}
	}
	return diff
}
