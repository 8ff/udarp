package main

import (
	"fmt"
)

func encodeGridLocator(grid string) uint16 {
	firstLetter := grid[0] - 'A'
	secondLetter := grid[1] - 'A'
	firstNumber := grid[2] - '0'
	secondNumber := grid[3] - '0'

	encoded := (uint16(firstLetter) * 10 * 24 * 24) + (uint16(secondLetter) * 10 * 24) + (uint16(firstNumber) * 24) + uint16(secondNumber)
	return encoded
}

func decodeGridLocator(encoded uint16) string {
	secondNumber := encoded % 24
	encoded /= 24

	firstNumber := encoded % 10
	encoded /= 10

	secondLetter := encoded % 24
	encoded /= 24

	firstLetter := encoded

	return string(firstLetter+'A') + string(secondLetter+'A') + string(firstNumber+'0') + string(secondNumber+'0')
}

func main() {
	gridLocator := "KP12"
	encoded := encodeGridLocator(gridLocator)
	fmt.Printf("Encoded grid locator for %s: %x %016b\n", gridLocator, encoded, encoded)

	decoded := decodeGridLocator(encoded)
	fmt.Printf("Decoded grid locator: %s\n", decoded)
}
