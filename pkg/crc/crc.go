package crc

import "fmt"

// Check if the byte and privided crc8 checksum is valid
func Match8(input []byte, checksum byte) bool {
	return Encode8(input) == checksum
}

// Check if the byte and privided crc16 checksum is valid
func Match16(input []byte, checksum uint16) bool {
	return Encode16(input) == checksum
}

// Check if the byte and privided crc32 checksum is valid
func Match32(input []byte, checksum uint32) bool {
	return Encode32(input) == checksum
}

// Function that strips prepended crc8 checksum from []byte and verifies it, if valid returns the byte array without the checksum, if invalid returns nil
func Decode8Array(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input is empty")
	}
	if Match8(input[1:], input[0]) {
		return input[1:], nil
	}
	return nil, fmt.Errorf("checksum does not match")
}

// Function that goes over [][]byte and prepends crc8 checksum to each byte array
func Encode8Chunks(input [][]byte) [][]byte {
	for i, b := range input {
		input[i] = append([]byte{Encode8(b)}, b...)
	}
	return input
}

// Function that goes over [][]byte and takes prepended crc8 from each byte array and verifies it, if invalid it nullifies the byte array, if valid it strips the checksum and returns the byte array
func Decode8Chunks(input [][]byte) ([][]byte, int) {
	corruptChunks := 0
	for i, b := range input {
		if len(b) == 0 {
			continue
		}
		if Match8(b[1:], b[0]) { // Verify the checksum
			input[i] = b[1:] // Strip the checksum
		} else {
			input[i] = nil // Nullify the byte array
			corruptChunks++
		}
	}
	if corruptChunks == len(input) {
		return nil, corruptChunks
	} else {
		return input, corruptChunks
	}
}

// Function which does crc8 checksum on input byte array
func Encode8(input []byte) byte {
	var crc byte = 0
	for _, b := range input {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x07
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// Function which does crc16 checksum on input byte array
func Encode16(input []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range input {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// Function which does crc32 checksum on input byte array
func Encode32(input []byte) uint32 {
	var crc uint32 = 0xFFFFFFFF
	for _, b := range input {
		crc ^= uint32(b)
		for i := 0; i < 8; i++ {
			if crc&0x00000001 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
