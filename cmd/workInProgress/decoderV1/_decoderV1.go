package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/8ff/udarp/pkg/crc"
	"github.com/8ff/udarp/pkg/misc"
	"github.com/8ff/udarp/pkg/rs"
)

/*
TODO
- [ ] Add interleaving to minimize the effect of burst errors
- [ ] Sort combinations so they start with all 1s and end with all 0s

*/

type runStats struct {
	Pass              bool
	CorruptBits       int
	TotalBits         int
	AttemptsToSuccess int
	TotalBlocks       int
	CorruptBlocks     int
}

var ComputedBitCombinations [][]int

// Logging function that accepts string and log level, it prints time, log level and the string, info is printed in green color, error is printed in red color, warning is printed in yellow color
func ll(level, msg string) {
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Function which prints diff between 2 input []uint8
func printDiff(a []uint8, corrupt []uint8) {
	var origString string
	var corruptString string
	var diff string
	corruptBits := 0

	for i := 0; i < len(a); i++ {
		if a[i] != corrupt[i] {
			diff += "X"
			origString = origString + fmt.Sprintf("%d", a[i])
			corruptString = corruptString + fmt.Sprintf("%d", corrupt[i])
			corruptBits++
		} else {
			diff = diff + "."
			origString = origString + fmt.Sprintf("%d", a[i])
			corruptString = corruptString + fmt.Sprintf("%d", corrupt[i])
		}
	}

	// Print all strings
	fmt.Printf("\n%s\n%s\n%s\nDiff Bits: %d\n", origString, corruptString, diff, corruptBits)
}

// Function that sets bit at position pos
func setBit(a byte, pos uint) byte {
	a |= (1 << pos)
	return a
}

// Function that clears bit at position pos
func clearBit(a byte, pos uint) byte {
	mask := byte(1 << pos)
	return a &^ mask
}

// Function that flips a bit at position pos
func flipBit(a byte, pos uint) byte {
	a ^= (1 << pos)
	return a
}

// Function that takes in [][]byte, converts it to []byte and then to a []int of 1 and 0s
func convertToBits(input [][]byte) []int {
	var bits []int
	for _, b := range input {
		for _, bit := range b {
			for i := 0; i < 8; i++ {
				if bit&(1<<uint(i)) != 0 {
					bits = append(bits, 1)
				} else {
					bits = append(bits, 0)
				}
			}
		}
	}
	return bits
}

// Function that takes in []int of 1 and 0s and converts it to []byte then splits it by chunkSize into [][]byte
func convertToBytes(input []int, chunkSize int) [][]byte {
	var bytes []byte
	for i := 0; i < len(input); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			if input[i+j] == 1 {
				b = setBit(b, uint(j))
			}
		}
		bytes = append(bytes, b)
	}

	var chunks [][]byte
	for i := 0; i < len(bytes); i += chunkSize {
		chunks = append(chunks, bytes[i:i+chunkSize])
	}
	return chunks
}

// Function that converts []int to []uint8
func convertToUint8(input []int) []uint8 {
	var output []uint8
	for _, i := range input {
		output = append(output, uint8(i))
	}
	return output
}

// Function that sorts [][]int by number of 1s in each slice (descending)
func sortCombinations(combinations [][]int) [][]int {
	for i := 0; i < len(combinations); i++ {
		for j := i + 1; j < len(combinations); j++ {
			if countOnes(combinations[i]) < countOnes(combinations[j]) {
				combinations[i], combinations[j] = combinations[j], combinations[i]
			}
		}
	}
	return combinations
}

// Function that counts number of 1s in a slice
func countOnes(input []int) int {
	var ones int
	for _, bit := range input {
		if bit == 1 {
			ones++
		}
	}
	return ones
}

// Function that compares 2 []int and returns the number of different bits
func compareBits(a []int, b []int) int {
	var diff int
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			diff++
		}
	}
	return diff
}

// Function that goes over 2 [][]byte and returns the number of different bits
func compareChunks(a [][]byte, b [][]byte) int {
	var diff int
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			if a[i][j] != b[i][j] {
				diff++
			}
		}
	}
	return diff
}

// Function that takes in [][]byte and interleaves it into a single []byte taking 1 byte from each slice at a time
func interleave(input [][]byte) [][]byte {
	var interleaved []byte
	for i := 0; i < len(input[0]); i++ {
		for _, slice := range input {
			interleaved = append(interleaved, slice[i])
		}
	}

	var chunks [][]byte
	chunkSize := len(input[0])
	for i := 0; i < len(interleaved); i += chunkSize {
		chunks = append(chunks, interleaved[i:i+chunkSize])
	}

	return chunks
}

// Function that takes in []byte and chunkSize and deinterleaves it into [][]byte taking 1 byte from each slice at a time
func deinterleave(input [][]byte) [][]byte {
	// Put input into a single slice
	var singleSlice []byte
	var correctOrder []byte
	for _, slice := range input {
		singleSlice = append(singleSlice, slice...)
	}

	for i := 0; i < len(input); i++ {
		for j := i; j < len(singleSlice); j += len(input) {
			correctOrder = append(correctOrder, singleSlice[j])
		}
		// deinterleaved = append(deinterleaved, slice)
	}

	var deinterleaved [][]byte
	// Split correctOrder into chunks of len(input) into deinterleaved
	for i := 0; i < len(correctOrder); i += len(input[0]) {
		deinterleaved = append(deinterleaved, correctOrder[i:i+len(input[0])])
	}

	return deinterleaved
}

func testNoCrcV3() runStats {
	ditLengthMs := 1000
	numOfBitsToCorrupt := 50

	// *** To avoid padding, make sure data is equal to DataShards * ChunkSize ***
	/*
		At 50 error setting



		4 11 2 17% 240bits 2/100 decodes GOOD
		3 12 2 16% 240bits 6/100 decodes GOOD
	*/

	params := rs.Params{DataShards: 3, ParityShards: 12, ChunkSize: 2}

	if ComputedBitCombinations == nil {
		ComputedBitCombinations = misc.GenerateBitCombinations(params.DataShards + params.ParityShards)
	}

	crcBytes := 2
	dataBytes := (params.DataShards * params.ChunkSize) - crcBytes // Not 10 because we need to leave 2 byte for crc16

	// *** 1. Generate original data ***
	originalData := make([]byte, dataBytes)
	for i := 0; i < dataBytes; i++ {
		originalData[i] = byte(0xFF)
	}
	fmt.Printf("ORIGINAL_DATA: %x\n", originalData)

	crcData := make([]byte, 2)
	// Run crc.Encode16 on originalData and convert uint16 to []byte crcData
	binary.LittleEndian.PutUint16(crcData, crc.Encode16(originalData))
	crcData = append(crcData, originalData...)
	originalData = crcData

	// fmt.Printf("CHUNKY_DATA: %x\n", originalData)

	// *** 2. Chunk data ***
	chunks, bytesPadded, err := rs.Chunk(params, originalData)
	if err != nil {
		fmt.Printf("Chunking failed: %s\n", err)
		return runStats{}
	}
	if bytesPadded != 0 {
		fmt.Printf("*** Bytes padded: %d\n", bytesPadded)
	}

	// fmt.Printf("CHUNKED_DATA: %x\n", chunks)

	// *** 3. Encode data using reedsolomon ***
	encodedChunks, err := rs.Encode(params, chunks)
	if err != nil {
		fmt.Printf("Encoding failed: %s\n", err)
		return runStats{}
	}
	fmt.Printf("ENCODED_DATA: %x\n", encodedChunks)

	// *** 4. Interleave data ***
	interleaved := interleave(encodedChunks)
	fmt.Printf("INTERLEAVED_DATA: %x\n", interleaved)

	/************************* RADIO *************************/
	bits := convertToBits(interleaved)
	// fmt.Printf("TOTAL_DATA_BITS: %d\n", len(bits))
	// fmt.Printf("BITS: %x\n", bits)

	// Corrupt data
	// corruptBlockIndex := make([]int, len(encodedChunks))
	corruptBits := make([]int, len(bits))
	copy(corruptBits, bits)
	for i := 0; i < numOfBitsToCorrupt; i++ {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(bits))

		// Figure out which block the bit is in given the index and params.ChunkSize
		for j := 0; j < len(encodedChunks); j++ {
			if randIndex < (j+1)*params.ChunkSize*8 {
				// corruptBlockIndex[j]++
				break
			}
		}

		if corruptBits[randIndex] == 1 {
			corruptBits[randIndex] = 0
		} else {
			corruptBits[randIndex] = 1
		}
	}

	// Print corrupt block index
	// fmt.Printf("CORRUPT_BLOCK_INDEX: %x\n", corruptBlockIndex)

	// Count and print how many corruptBlocks there are
	// var corruptBlocks int
	// for _, block := range corruptBlockIndex {
	// 	if block > 0 {
	// 		corruptBlocks++
	// 	}
	// }

	// ll("warning", fmt.Sprintf("CORRUPT_BLOCKS: %d/%d", corruptBlocks, len(encodedChunks)))

	importedBytes := convertToBytes(corruptBits, params.ChunkSize)
	fmt.Printf("IMPORTED_BYTES: %x\n", importedBytes)

	// Deinterleave data
	deinterleaved := deinterleave(importedBytes)
	fmt.Printf("DEINTERLEAVED_DATA: %x\n", deinterleaved)

	// Compare deinterleaved with encodedChunks and print number of errors in each block
	visualBlocksErrors := make([]int, len(deinterleaved))
	var corruptBlocks int
	var totalErrors int
	for i, block := range deinterleaved {
		errors := 0
		for j, b := range block {
			if b != encodedChunks[i][j] {
				errors++
			}
		}
		totalErrors += errors
		visualBlocksErrors[i] = errors
		//		fmt.Printf("BLOCK_%d_ERRORS: %d/%d\n", i, errors, len(block))
	}

	// Count how many blocks are corrupt and store in corruptBlocks
	for _, block := range visualBlocksErrors {
		if block > 0 {
			corruptBlocks++
		}
	}

	// Print visualBlocksErrors
	ll("warning", fmt.Sprintf("VISUAL_BLOCKS_ERRORS: %x", visualBlocksErrors))

	// *********************** END_RADIO *************************

	start := time.Now()
	// combos := combinations(len(deinterleaved))
	combos := ComputedBitCombinations
	log.Printf("Generated combos for %d bytes, total: %d Took: %dms\n", len(importedBytes), len(combos), time.Since(start).Milliseconds())

	decodeFailCount := 0
	invalidCrcCount := 0
	successfullDecode := -1
	successfullCombo := []int{}

	// timerStart := time.Now()

	for i, combo := range combos {

		// Copy importedBytes to tempSlice
		tempSlice := make([][]byte, len(deinterleaved))
		copy(tempSlice, deinterleaved)
		// Go over combos bits, if bit is 0 then nil that slice in importedBytes, if bit is 1 then leave it
		for j, bit := range combo {
			if bit == 0 {
				tempSlice[j] = nil
			}
		}

		// Decode the data using reedsolomon
		decoded, err := rs.DecodeOnlyData(params, tempSlice)
		if err == nil {
			// Decode successful
			// ll("info", "*** RS_DECODE_PASS ***")
			// ll("debug", fmt.Sprintf("DECODED: %x", decoded))

			// Take 2 bytes from decoded and convert them to uint16
			checksum := binary.LittleEndian.Uint16(decoded[0:2])

			// Check crc8 of decoded data
			if crc.Match16(decoded[2:], checksum) {
				successfullDecode = i
				successfullCombo = combo
				ll("info", "*** CRC_PASS ***")
				ll("info", fmt.Sprintf("CORRECT_DATA: %x", decoded[2:]))

				res := bytes.Compare(originalData, decoded)
				if res != 0 {
					ll("error", "*** DATA_MISMATCH")
				}

				break
			} else {
				invalidCrcCount++
				// ll("error", "*** CRC_FAIL ***")
			}
		} else {
			decodeFailCount++
		}
	}

	// ll("error", fmt.Sprintf("TOOK: %d", time.Since(timerStart).Milliseconds()))

	corruptBitCount := compareBits(bits, corruptBits)
	level := "info"
	if successfullDecode == -1 {
		level = "error"
	}

	// Print stats
	ll(level, fmt.Sprintf("CHUNKS: %d CHUNK_SIZE: %d MINIMUM_CHUNKS: %d TOTAL_BITS: %d DATA_BYTES: %d", len(encodedChunks), params.ChunkSize, params.DataShards, len(bits), params.ChunkSize*params.DataShards-2))
	ll(level, fmt.Sprintf("DIT_LEN: %dms TRANSMIT_TIME: %.0fs or %.0fm", ditLengthMs, float64(ditLengthMs*len(bits))/1000, float64(ditLengthMs*len(bits))/1000/60))
	ll(level, fmt.Sprintf("ATTEMPT: %d DECODE_FAILS: %d INVALID_CRCS: %d CORRUPT_BITS: %d/%d CORRUPT_BLOCKS: %d/%d", successfullDecode, decodeFailCount, invalidCrcCount, corruptBitCount, len(bits), corruptBlocks, len(encodedChunks)))
	ll(level, fmt.Sprintf("COMBO: %v", successfullCombo))

	pass := successfullDecode != -1

	return runStats{Pass: pass, CorruptBits: corruptBitCount, TotalBits: len(bits), AttemptsToSuccess: successfullDecode, TotalBlocks: len(encodedChunks), CorruptBlocks: corruptBlocks}

}

func testNoCrcV3NoInterleave() runStats {
	ditLengthMs := 1000
	numOfBitsToCorrupt := 50

	// *** To avoid padding, make sure data is equal to DataShards * ChunkSize ***
	//	params := rs.Params{DataShards: 3, ParityShards: 12, ChunkSize: 3} // LAST GOOD
	params := rs.Params{DataShards: 3, ParityShards: 13, ChunkSize: 3}

	if ComputedBitCombinations == nil {
		ComputedBitCombinations = misc.GenerateBitCombinations(params.DataShards + params.ParityShards)
	}

	crcBytes := 2
	dataBytes := params.DataShards*params.ChunkSize - crcBytes // Not 10 because we need to leave 2 byte for crc16

	// *** 1. Generate original data ***
	originalData := make([]byte, dataBytes)
	for i := 0; i < dataBytes; i++ {
		originalData[i] = byte(0xFF)
	}
	fmt.Printf("ORIGINAL_DATA: %x\n", originalData)

	crcData := make([]byte, 2)
	// Run crc.Encode16 on originalData and convert uint16 to []byte crcData
	binary.LittleEndian.PutUint16(crcData, crc.Encode16(originalData))
	crcData = append(crcData, originalData...)
	originalData = crcData

	// fmt.Printf("CHUNKY_DATA: %x\n", originalData)

	// *** 2. Chunk data ***
	chunks, bytesPadded, err := rs.Chunk(params, originalData)
	if err != nil {
		fmt.Printf("Chunking failed: %s\n", err)
		return runStats{}
	}
	if bytesPadded != 0 {
		fmt.Printf("*** Bytes padded: %d\n", bytesPadded)
	}

	// fmt.Printf("CHUNKED_DATA: %x\n", chunks)

	// *** 3. Encode data using reedsolomon ***
	encodedChunks, err := rs.Encode(params, chunks)
	if err != nil {
		fmt.Printf("Encoding failed: %s\n", err)
		return runStats{}
	}
	fmt.Printf("ENCODED_DATA: %x\n", encodedChunks)

	/************************* RADIO *************************/
	bits := convertToBits(encodedChunks)
	// fmt.Printf("TOTAL_DATA_BITS: %d\n", len(bits))
	// fmt.Printf("BITS: %x\n", bits)

	// Corrupt data
	corruptBlockIndex := make([]int, len(encodedChunks))
	corruptBits := make([]int, len(bits))
	copy(corruptBits, bits)
	for i := 0; i < numOfBitsToCorrupt; i++ {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(bits))

		// Figure out which block the bit is in given the index and params.ChunkSize
		for j := 0; j < len(encodedChunks); j++ {
			if randIndex < (j+1)*params.ChunkSize*8 {
				corruptBlockIndex[j]++
				break
			}
		}

		if corruptBits[randIndex] == 1 {
			corruptBits[randIndex] = 0
		} else {
			corruptBits[randIndex] = 1
		}
	}

	// Print corrupt block index
	fmt.Printf("CORRUPT_BLOCK_INDEX: %x\n", corruptBlockIndex)

	// Count and print how many corruptBlocks there are
	var corruptBlocks int
	for _, block := range corruptBlockIndex {
		if block > 0 {
			corruptBlocks++
		}
	}

	ll("warning", fmt.Sprintf("CORRUPT_BLOCKS: %d/%d", corruptBlocks, len(encodedChunks)))

	// Convert bits to []uint8

	// printDiff(convertToUint8(bits), convertToUint8(corruptBits))

	importedBytes := convertToBytes(corruptBits, params.ChunkSize)
	fmt.Printf("IMPORTED_BYTES: %x\n", importedBytes)

	// *********************** END_RADIO *************************

	start := time.Now()
	// combos := combinations(len(deinterleaved))
	combos := ComputedBitCombinations
	log.Printf("Generated combos for %d bytes, total: %d Took: %dms\n", len(importedBytes), len(combos), time.Since(start).Milliseconds())

	decodeFailCount := 0
	invalidCrcCount := 0
	successfullDecode := -1
	successfullCombo := []int{}
	for i, combo := range combos {
		// Copy importedBytes to tempSlice
		tempSlice := make([][]byte, len(importedBytes))
		copy(tempSlice, importedBytes)
		// Go over combos bits, if bit is 0 then nil that slice in importedBytes, if bit is 1 then leave it
		for j, bit := range combo {
			if bit == 0 {
				tempSlice[j] = nil
			}
		}

		// Decode the data using reedsolomon
		decoded, err := rs.DecodeOnlyData(params, tempSlice)
		if err == nil {
			// Decode successful
			// ll("info", "*** RS_DECODE_PASS ***")
			// ll("debug", fmt.Sprintf("DECODED: %x", decoded))

			// Take 2 bytes from decoded and convert them to uint16
			checksum := binary.LittleEndian.Uint16(decoded[0:2])

			// Check crc8 of decoded data
			if crc.Match16(decoded[2:], checksum) {
				successfullDecode = i
				successfullCombo = combo
				ll("info", "*** CRC_PASS ***")
				ll("info", fmt.Sprintf("CORRECT_DATA: %x", decoded[2:]))

				res := bytes.Compare(originalData, decoded)
				if res != 0 {
					ll("error", "*** DATA_MISMATCH")
				}

				break
			} else {
				invalidCrcCount++
				// ll("error", "*** CRC_FAIL ***")
			}
		} else {
			decodeFailCount++
		}
	}

	corruptBitCount := compareBits(bits, corruptBits)
	level := "info"
	if successfullDecode == -1 {
		level = "error"
	}

	// Print stats
	ll(level, fmt.Sprintf("CHUNKS: %d CHUNK_SIZE: %d MINIMUM_CHUNKS: %d TOTAL_BITS: %d DATA_BYTES: %d", len(encodedChunks), params.ChunkSize, params.DataShards, len(bits), params.ChunkSize*params.DataShards-2))
	ll(level, fmt.Sprintf("DIT_LEN: %dms TRANSMIT_TIME: %.0fs or %.0fm", ditLengthMs, float64(ditLengthMs*len(bits))/1000, float64(ditLengthMs*len(bits))/1000/60))
	ll(level, fmt.Sprintf("ATTEMPT: %d DECODE_FAILS: %d INVALID_CRCS: %d CORRUPT_BITS: %d/%d", successfullDecode, decodeFailCount, invalidCrcCount, corruptBitCount, len(bits)))
	ll(level, fmt.Sprintf("COMBO: %v", successfullCombo))

	pass := successfullDecode != -1

	return runStats{Pass: pass, CorruptBits: corruptBitCount, TotalBits: len(bits), AttemptsToSuccess: successfullDecode, TotalBlocks: len(encodedChunks), CorruptBlocks: corruptBlocks}

}

func testNoCrcV2() runStats {
	ditLengthMs := 1000
	numOfBitsToCorrupt := 50

	// *** To avoid padding, make sure data is equal to DataShards * ChunkSize ***
	//	params := rs.Params{DataShards: 3, ParityShards: 12, ChunkSize: 3} // LAST GOOD
	params := rs.Params{DataShards: 3, ParityShards: 13, ChunkSize: 3}
	crcBytes := 2
	dataBytes := params.DataShards*params.ChunkSize - crcBytes // Not 10 because we need to leave 2 byte for crc16

	// *** 1. Generate original data ***
	originalData := make([]byte, dataBytes)
	for i := 0; i < dataBytes; i++ {
		originalData[i] = byte(0xFF)
	}
	fmt.Printf("ORIGINAL_DATA: %x\n", originalData)

	crcData := make([]byte, 2)
	// Run crc.Encode16 on originalData and convert uint16 to []byte crcData
	binary.LittleEndian.PutUint16(crcData, crc.Encode16(originalData))
	crcData = append(crcData, originalData...)
	originalData = crcData

	// fmt.Printf("CHUNKY_DATA: %x\n", originalData)

	// *** 2. Chunk data ***
	chunks, bytesPadded, err := rs.Chunk(params, originalData)
	if err != nil {
		fmt.Printf("Chunking failed: %s\n", err)
		return runStats{}
	}
	if bytesPadded != 0 {
		fmt.Printf("*** Bytes padded: %d\n", bytesPadded)
	}

	// fmt.Printf("CHUNKED_DATA: %x\n", chunks)

	// *** 3. Encode data using reedsolomon ***
	encodedChunks, err := rs.Encode(params, chunks)
	if err != nil {
		fmt.Printf("Encoding failed: %s\n", err)
		return runStats{}
	}
	fmt.Printf("ENCODED_DATA: %x\n", encodedChunks)

	// *** 4. Interleave data ***
	interleaved := interleave(encodedChunks)
	fmt.Printf("INTERLEAVED_DATA: %x\n", interleaved)

	/************************* RADIO *************************/
	bits := convertToBits(interleaved)
	// fmt.Printf("TOTAL_DATA_BITS: %d\n", len(bits))
	// fmt.Printf("BITS: %x\n", bits)

	// Corrupt data
	corruptBlockIndex := make([]int, len(encodedChunks))
	corruptBits := make([]int, len(bits))
	copy(corruptBits, bits)
	for i := 0; i < numOfBitsToCorrupt; i++ {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(bits))

		// Figure out which block the bit is in given the index and params.ChunkSize
		for j := 0; j < len(encodedChunks); j++ {
			if randIndex < (j+1)*params.ChunkSize*8 {
				corruptBlockIndex[j]++
				break
			}
		}

		if corruptBits[randIndex] == 1 {
			corruptBits[randIndex] = 0
		} else {
			corruptBits[randIndex] = 1
		}
	}

	// Print corrupt block index
	fmt.Printf("CORRUPT_BLOCK_INDEX: %x\n", corruptBlockIndex)

	// Count and print how many corruptBlocks there are
	var corruptBlocks int
	for _, block := range corruptBlockIndex {
		if block > 0 {
			corruptBlocks++
		}
	}

	ll("warning", fmt.Sprintf("CORRUPT_BLOCKS: %d/%d", corruptBlocks, len(encodedChunks)))

	// Convert bits to []uint8

	// printDiff(convertToUint8(bits), convertToUint8(corruptBits))

	importedBytes := convertToBytes(corruptBits, params.ChunkSize)
	fmt.Printf("IMPORTED_BYTES: %x\n", importedBytes)

	// Deinterleave data
	deinterleaved := deinterleave(importedBytes)
	fmt.Printf("DEINTERLEAVED_DATA: %x\n", deinterleaved)

	// *********************** END_RADIO *************************

	start := time.Now()
	combos := misc.GenerateBitCombinations(len(deinterleaved))
	log.Printf("Generated combos for %d bytes, total: %d Took: %dms\n", len(importedBytes), len(combos), time.Since(start).Milliseconds())

	decodeFailCount := 0
	invalidCrcCount := 0
	successfullDecode := -1
	successfullCombo := []int{}
	for i, combo := range combos {
		// Copy importedBytes to tempSlice
		tempSlice := make([][]byte, len(deinterleaved))
		copy(tempSlice, deinterleaved)
		// Go over combos bits, if bit is 0 then nil that slice in importedBytes, if bit is 1 then leave it
		for j, bit := range combo {
			if bit == 0 {
				tempSlice[j] = nil
			}
		}

		// Decode the data using reedsolomon
		decoded, err := rs.DecodeOnlyData(params, tempSlice)
		if err == nil {
			// Decode successful
			// ll("info", "*** RS_DECODE_PASS ***")
			// ll("debug", fmt.Sprintf("DECODED: %x", decoded))

			// Take 2 bytes from decoded and convert them to uint16
			checksum := binary.LittleEndian.Uint16(decoded[0:2])

			// Check crc8 of decoded data
			if crc.Match16(decoded[2:], checksum) {
				successfullDecode = i
				successfullCombo = combo
				ll("info", "*** CRC_PASS ***")
				ll("info", fmt.Sprintf("CORRECT_DATA: %x", decoded[2:]))

				res := bytes.Compare(originalData, decoded)
				if res != 0 {
					ll("error", "*** DATA_MISMATCH")
				}

				break
			} else {
				invalidCrcCount++
				// ll("error", "*** CRC_FAIL ***")
			}
		} else {
			decodeFailCount++
		}
	}

	corruptBitCount := compareBits(bits, corruptBits)
	level := "info"
	if successfullDecode == -1 {
		level = "error"
	}

	// Print stats
	ll(level, fmt.Sprintf("CHUNKS: %d CHUNK_SIZE: %d MINIMUM_CHUNKS: %d TOTAL_BITS: %d DATA_BYTES: %d", len(encodedChunks), params.ChunkSize, params.DataShards, len(bits), params.ChunkSize*params.DataShards-2))
	ll(level, fmt.Sprintf("DIT_LEN: %dms TRANSMIT_TIME: %.0fs or %.0fm", ditLengthMs, float64(ditLengthMs*len(bits))/1000, float64(ditLengthMs*len(bits))/1000/60))
	ll(level, fmt.Sprintf("ATTEMPT: %d DECODE_FAILS: %d INVALID_CRCS: %d CORRUPT_BITS: %d/%d", successfullDecode, decodeFailCount, invalidCrcCount, corruptBitCount, len(bits)))
	ll(level, fmt.Sprintf("COMBO: %v", successfullCombo))

	pass := successfullDecode != -1

	return runStats{Pass: pass, CorruptBits: corruptBitCount, TotalBits: len(bits), AttemptsToSuccess: successfullDecode, TotalBlocks: len(encodedChunks), CorruptBlocks: corruptBlocks}

}

func testNoCrc() {
	// *** To avoid padding, make sure data is equal to DataShards * ChunkSize ***
	params := rs.Params{DataShards: 5, ParityShards: 10, ChunkSize: 2}
	dataBytes := 8 // Not 10 because we need to leave 2 byte for crc16

	// *** 1. Generate original data ***
	originalData := make([]byte, dataBytes)
	for i := 0; i < dataBytes; i++ {
		originalData[i] = byte(0xFF)
	}
	fmt.Printf("ORIGINAL_DATA: %x\n", originalData)

	crcData := make([]byte, 2)
	// Run crc.Encode16 on originalData and convert uint16 to []byte crcData
	binary.LittleEndian.PutUint16(crcData, crc.Encode16(originalData))
	crcData = append(crcData, originalData...)
	originalData = crcData

	fmt.Printf("CHUNKY_DATA: %x\n", originalData)

	// *** 2. Chunk data ***
	chunks, bytesPadded, err := rs.Chunk(params, originalData)
	if err != nil {
		fmt.Printf("Chunking failed: %s\n", err)
		return
	}
	if bytesPadded != 0 {
		fmt.Printf("*** Bytes padded: %d\n", bytesPadded)
	}

	fmt.Printf("CHUNKED_DATA: %x\n", chunks)

	// *** 3. Encode data using reedsolomon ***
	encodedChunks, err := rs.Encode(params, chunks)
	if err != nil {
		fmt.Printf("Encoding failed: %s\n", err)
		return
	}
	fmt.Printf("ENCODED_DATA: %x\n", encodedChunks)

	// ******** TRANSMIT ********
	// fmt.Printf("TOTAL_DATA_BITS: %d\n", len(crcEncodedData)*len(crcEncodedData[0])*8)

	// Corrupt data
	// crcEncodedData[0][0] = clearBit(crcEncodedData[0][0], 0)
	// crcEncodedData[1][0] = clearBit(crcEncodedData[1][0], 0)
	// fmt.Printf("CORRUPT_DATA: %x\n", crcEncodedData)

	/************************* RADIO *************************/
	bits := convertToBits(encodedChunks)
	fmt.Printf("TOTAL_DATA_BITS: %d\n", len(bits))
	fmt.Printf("BITS: %x\n", bits)

	// Corrupt data
	corruptBlockIndex := make([]int, len(encodedChunks))
	numOfBitsToCorrupt := 20
	corruptBits := make([]int, len(bits))
	copy(corruptBits, bits)
	for i := 0; i < numOfBitsToCorrupt; i++ {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(bits))

		// Figure out which block the bit is in given the index and params.ChunkSize
		for j := 0; j < len(encodedChunks); j++ {
			if randIndex < (j+1)*params.ChunkSize*8 {
				corruptBlockIndex[j]++
				break
			}
		}

		if corruptBits[randIndex] == 1 {
			corruptBits[randIndex] = 0
		} else {
			corruptBits[randIndex] = 1
		}
	}

	// Print corrupt block index
	fmt.Printf("CORRUPT_BLOCK_INDEX: %x\n", corruptBlockIndex)

	// Count and print how many corruptBlocks there are
	var corruptBlocks int
	for _, block := range corruptBlockIndex {
		if block > 0 {
			corruptBlocks++
		}
	}

	ll("warning", fmt.Sprintf("CORRUPT_BLOCKS: %d", corruptBlocks))

	// Convert bits to []uint8

	// printDiff(convertToUint8(bits), convertToUint8(corruptBits))

	importedBytes := convertToBytes(corruptBits, params.ChunkSize)
	fmt.Printf("IMPORTED_BYTES: %x\n", importedBytes)
	// *********************** END_RADIO *************************

	// Start a timer to time how long it takes to decode
	// start := time.Now()
	// fmt.Printf("START_TIME: %s\n", start)

	// Stop the timer
	// elapsed := time.Since(start)
	// fmt.Printf("DECODE_TIME: %s\n", elapsed)

	combos := misc.GenerateBitCombinations(len(importedBytes))
	log.Printf("Generated combos for %d bytes, total: %d\n", len(importedBytes), len(combos))

	decodeFailCount := 0
	invalidCrcCount := 0
	successfullDecode := 0
	successfullCombo := []int{}
	for i, combo := range combos {
		// Copy importedBytes to tempSlice
		tempSlice := make([][]byte, len(importedBytes))
		copy(tempSlice, importedBytes)
		// Go over combos bits, if bit is 0 then nil that slice in importedBytes, if bit is 1 then leave it
		for j, bit := range combo {
			if bit == 0 {
				tempSlice[j] = nil
			}
		}

		// Decode the data using reedsolomon
		decoded, err := rs.DecodeOnlyData(params, tempSlice)
		if err == nil {
			// Decode successful
			// ll("info", "*** RS_DECODE_PASS ***")
			// ll("debug", fmt.Sprintf("DECODED: %x", decoded))

			// Take 2 bytes from decoded and convert them to uint16
			checksum := binary.LittleEndian.Uint16(decoded[0:2])

			// Check crc8 of decoded data
			if crc.Match16(decoded[2:], checksum) {
				successfullDecode = i
				successfullCombo = combo
				ll("info", "*** CRC_PASS ***")
				ll("info", fmt.Sprintf("CORRECT_DATA: %x", decoded[2:]))
				break
			} else {
				invalidCrcCount++
				// ll("error", "*** CRC_FAIL ***")
			}
		} else {
			decodeFailCount++
		}
	}

	corruptBitCount := compareBits(bits, corruptBits)
	level := "info"
	if successfullDecode == 0 {
		level = "error"
	}
	ll(level, fmt.Sprintf("ATTEMPT: %d DECODE_FAILS: %d INVALID_CRCS: %d CORRUPT_BITS: %d/%d", successfullDecode, decodeFailCount, invalidCrcCount, corruptBitCount, len(bits)))
	ll(level, fmt.Sprintf("COMBO: %v", successfullCombo))

	// res := bytes.Compare(originalData, decoded)
	// if res == 0 {
	// 	fmt.Println("*** PASS")
	// } else {
	// 	fmt.Println("*** DATA_MISMATCH")
	// }
}

func testRs() {
	// *** To avoid padding, make sure data is equal to DataShards * ChunkSize ***
	params := rs.Params{DataShards: 5, ParityShards: 10, ChunkSize: 2}
	dataBytes := 10

	// *** 1. Generate original data ***
	originalData := make([]byte, dataBytes)
	for i := 0; i < dataBytes; i++ {
		originalData[i] = byte(0xFF)
	}
	fmt.Printf("ORIGINAL_DATA: %x\n", originalData)

	// *** 2. Chunk data ***
	chunks, bytesPadded, err := rs.Chunk(params, originalData)
	if err != nil {
		fmt.Printf("Chunking failed: %s\n", err)
		return
	}
	if bytesPadded != 0 {
		fmt.Printf("*** Bytes padded: %d\n", bytesPadded)
	}

	fmt.Printf("CHUNKED_DATA: %x\n", chunks)

	// *** 3. Encode data using reedsolomon ***
	encodedChunks, err := rs.Encode(params, chunks)
	if err != nil {
		fmt.Printf("Encoding failed: %s\n", err)
		return
	}
	fmt.Printf("ENCODED_DATA: %x\n", encodedChunks)

	// *** 4. Append crc8 to chunks ***
	crcEncodedData := crc.Encode8Chunks(encodedChunks)
	fmt.Printf("CRC_ENCODED_DATA: %x\n", crcEncodedData)

	// ******** TRANSMIT ********
	fmt.Printf("TOTAL_DATA_BITS: %d\n", len(crcEncodedData)*len(crcEncodedData[0])*8)

	// Corrupt data
	// crcEncodedData[0][0] = clearBit(crcEncodedData[0][0], 0)
	// crcEncodedData[1][0] = clearBit(crcEncodedData[1][0], 0)
	// fmt.Printf("CORRUPT_DATA: %x\n", crcEncodedData)

	/************************* RADIO *************************/
	bits := convertToBits(crcEncodedData)
	fmt.Printf("TOTAL_DATA_BITS: %d\n", len(bits))
	fmt.Printf("BITS: %x\n", bits)

	// Corrupt data
	// Flip 10% of bits from bits
	for i := 0; i < len(bits); i++ {
		if rand.Intn(100) < 2 {
			if bits[i] == 1 {
				bits[i] = 0
			} else {
				bits[i] = 1
			}
		}
	}

	importedBytes := convertToBytes(bits, params.ChunkSize+1) // +1 for crc8
	fmt.Printf("IMPORTED_BYTES: %x\n", importedBytes)
	// *********************** END_RADIO *************************

	// *** 5. Decode crc8 ***
	crcDecoded, corruptChunks := crc.Decode8Chunks(importedBytes)
	fmt.Printf("CRC_DECODED_CHUNKS: %x\n", crcDecoded)
	if corruptChunks != 0 {
		fmt.Printf("*** CORRUPT_CHUNKS: %d\n", corruptChunks)
	}

	// // Decode the data using reedsolomon
	decoded, err := rs.Decode(params, crcDecoded)
	if err != nil {
		fmt.Printf("*** DECODING_FAILED: %s\n", err)
		return
	}

	// // Print decoded data
	fmt.Printf("DECODED: %x\n", decoded)

	res := bytes.Compare(originalData, decoded)
	if res == 0 {
		fmt.Println("*** PASS")
	} else {
		fmt.Println("*** DATA_MISMATCH")
	}
}

func testInterleave() {
	slices := 15
	chunkSize := 3
	data := make([][]byte, slices)
	// counter := 1
	for i := 0; i < slices; i++ {
		data[i] = make([]byte, chunkSize)
	}

	// // Fill data with random data
	for i := 0; i < slices; i++ {
		for j := 0; j < chunkSize; j++ {
			// Fill it with sequential data
			//			data[i][j] = byte(counter)
			// Random data
			data[i][j] = byte(rand.Intn(256))
			// counter++
		}
	}

	fmt.Printf("DATA: %x\n", data)
	interleaved := interleave(data)
	fmt.Printf("INTERLEAVED: %x\n", interleaved)
	deinterleaved := deinterleave(interleaved)
	fmt.Printf("DEINTERLEAVED: %x\n", deinterleaved)

	if compareChunks(data, deinterleaved) != 0 {
		ll("error", "INTERLEAVING_FAILED, chunks are different")
	}
}

func main() {
	// testInterleave()
	// testRs()
	// testNoCrc()
	// testNoCrcV2()

	// Run 	testNoCrcV2() 100 times and capture the results, print how many times it passed
	// and how many times it failed
	runs := make([]runStats, 0)
	numRuns := 100
	var pass int
	for i := 0; i < numRuns; i++ {
		stats := testNoCrcV3()
		// stats := testNoCrcV3NoInterleave()
		if stats.Pass {
			runs = append(runs, stats)
			pass++
		}
	}
	fmt.Printf("PASS: %d, FAIL: %d\n", pass, numRuns-pass)

	var corruptBits int
	var totalBits int
	for i := 0; i < len(runs); i++ {
		corruptBits += runs[i].CorruptBits
		totalBits += runs[i].TotalBits
	}

	// Print all stats from runs
	for i := 0; i < len(runs); i++ {
		ll("info", fmt.Sprintf("RUN: %d RATIO: %f%% CORRUPT_BITS: %d TOTAL_BITS: %d ATTEMPTS_TO_SUCCESS: %d TOTAL_BLOCKS: %d CORRUPT_BLOCKS: %d", i, (float64(corruptBits)/float64(totalBits))*100, runs[i].CorruptBits, runs[i].TotalBits, runs[i].AttemptsToSuccess, runs[i].TotalBlocks, runs[i].CorruptBlocks))
	}

	// Make a [][]byte with 3 slices and 2 bytes in each slice

	// combos := combinations(15)
	// // Print first 10 combinations
	// for i := 0; i < 10; i++ {
	// 	fmt.Printf("%d: %x\n", i, combos[i])
	// }

	// Sort using sortCombinations and print first 10 combinations
	// combo1 := sortCombinations(combos)
	// for i := 0; i < 10; i++ {
	// 	fmt.Printf("%d: %x\n", i, combo1[i])
	// }

	// fmt.Printf("COMBINATIONS: %d\n", len(combos))

	/*
		      INTERLEAVED_DATA: [15ffff 1550ba ba5095 7f7f95 d03a3a 80ffff 80403f 3f40e5 9a9ae5 255a5a ffffff ffffff ffffff ffffff ffffff]
			DEINTERLEAVED_DATA: [1580ff ffffff ffffff 1580ff 5040ff ba3fff ba3fff 5040ff 95e5ff 7f9aff 7f9aff 95e5ff d025ff 3a5aff 3a5aff]
	*/
}
