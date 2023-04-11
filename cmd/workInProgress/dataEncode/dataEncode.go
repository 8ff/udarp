package main

import (
	"fmt"
)

func encodeGridLocator(grid string) uint16 {
	firstLetter := grid[0] - 'A'
	secondLetter := grid[1] - 'A'
	firstNumber := grid[2] - '0'
	secondNumber := grid[3] - '0'

	encoded := (uint16(firstLetter) << 10) | (uint16(secondLetter) << 5) | (uint16(firstNumber) << 2) | uint16(secondNumber)
	return encoded
}

func decodeGridLocator(encoded uint16) string {
	firstLetter := (encoded >> 10) & 0x1F
	secondLetter := (encoded >> 5) & 0x1F
	firstNumber := (encoded >> 2) & 0x7
	secondNumber := encoded & 0x3

	return string(firstLetter+'A') + string(secondLetter+'A') + string(firstNumber+'0') + string(secondNumber+'0')
}

func main() {
	gridLocator := "FN03"
	encoded := encodeGridLocator(gridLocator)
	fmt.Printf("Encoded grid locator for %s: %x %016b\n", gridLocator, encoded, encoded)

	decoded := decodeGridLocator(encoded)
	fmt.Printf("Decoded grid locator: %s\n", decoded)
}
