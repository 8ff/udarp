package rs

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

type Params struct {
	DataShards   int
	ParityShards int
	ChunkSize    int
}

/*
TODO
- [ ] Add function to add crc bytes to [][]byte output from RS encoder, it should take parameters on which crc to use and return all of the combined data
- [ ] Change hardcoded data shards from 1 to DataShards
*/

// Function that chunks the data into 1 shard of [][]byte and pads uneven shards with 0x00, shardSize is automatically calculated
func Chunk(params Params, data []byte) ([][]byte, int, error) {
	bytesPadded := (params.DataShards * params.ChunkSize) - len(data)

	if len(data) == 0 {
		return nil, bytesPadded, fmt.Errorf("data is empty")
	} else if len(data) > params.DataShards*params.ChunkSize {
		return nil, bytesPadded, fmt.Errorf("data is too large for the given parameters")
	}

	shard := make([][]byte, params.DataShards)
	// Initialize all shards
	for shardIndex := 0; shardIndex < params.DataShards; shardIndex++ {
		shard[shardIndex] = make([]byte, params.ChunkSize)
		if len(data) > params.ChunkSize {
			copy(shard[shardIndex], data[:params.ChunkSize])
			data = data[params.ChunkSize:]
		} else {
			copy(shard[shardIndex], data)
			data = data[len(data):]
		}
	}

	// shard[0] = make([]byte, params.ChunkSize)
	// shard[0] = data
	return shard, bytesPadded, nil
}

// Function which encodes the data using reedsolomon
func Encode(params Params, chunks [][]byte) ([][]byte, error) {
	enc, err := reedsolomon.New(len(chunks), params.ParityShards)
	if err != nil {
		return nil, err
	}

	// Add parity shards to chunks
	for i := 0; i < params.ParityShards; i++ {
		chunks = append(chunks, make([]byte, len(chunks[0])))
	}

	err = enc.Encode(chunks)
	if err != nil {
		return nil, err
	}

	ok, err := enc.Verify(chunks)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("verification failed")
	}

	return chunks, nil
}

// Decode the data using reedsolomon
func Decode(params Params, data [][]byte) ([]byte, error) {
	enc, err := reedsolomon.New(params.DataShards, params.ParityShards)
	if err != nil {
		return nil, err
	}

	// Reconstruct the missing data
	err = enc.Reconstruct(data)
	if err != nil {
		return nil, err
	}

	// Verify the data
	var ok bool
	ok, err = enc.Verify(data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("verification failed")
	}

	rdata := make([]byte, 0)
	for i := 0; i < params.DataShards; i++ {
		rdata = append(rdata, data[i]...)
	}

	return rdata, nil
}

func DecodeOnlyData(params Params, data [][]byte) ([]byte, error) {
	enc, err := reedsolomon.New(params.DataShards, params.ParityShards)
	if err != nil {
		return nil, err
	}

	// Reconstruct the missing data
	err = enc.ReconstructData(data)
	if err != nil {
		return nil, err
	}

	rdata := make([]byte, 0)
	for i := 0; i < params.DataShards; i++ {
		rdata = append(rdata, data[i]...)
	}

	return rdata, nil
}

// Function that goes over [][]byte and returns []byte
func DeflateBlocks(data [][]byte) []byte {
	var result []byte
	for _, b := range data {
		result = append(result, b...)
	}
	return result
}

// Function that goes over []byte and returns [][]byte splitting by chunkSize
func InflateBlocks(data []byte, chunkSize int) [][]byte {
	var result [][]byte
	for i := 0; i < len(data); i += chunkSize {
		result = append(result, data[i:i+chunkSize])
	}
	return result
}

/********** TEMPLATES **********/
/*
// Function which encodes the data using reedsolomon
func Encode(dataShards, parityShards, shardSize int) ([][]byte, error) {
	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, err
	}

	data := make([][]byte, shardSize)
	for i := range data {
		data[i] = make([]byte, 10)
	}

	// Fill some data into the data shards with FF
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			data[i][j] = 0xFF
		}
	}

	err = enc.Encode(data)
	if err != nil {
		return nil, err
	}

	var ok bool
	ok, err = enc.Verify(data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("verification failed")
	}

	return data, nil
}

// Decode the data using reedsolomon
func Decode(data [][]byte) ([]byte, error) {
	enc, err := reedsolomon.New(5, 5)
	if err != nil {
		return nil, err
	}

	// Delete some shards
	// data[0] = nil
	// data[1] = nil
	// data[2] = nil
	// data[3] = nil
	// data[4] = nil
	// data[5] = nil
	// data[6] = nil
	// data[7] = nil
	// data[8] = nil
	// data[9] = nil

	// Reconstruct the missing data
	err = enc.Reconstruct(data)
	if err != nil {
		return nil, err
	}

	// Verify the data
	var ok bool
	ok, err = enc.Verify(data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("verification failed")
	}

	rdata := make([]byte, 0)
	for i := 0; i < 5; i++ {
		rdata = append(rdata, data[i]...)
	}

	return rdata, nil
}
*/

// func test1() {
// 	// Encode the data
// 	data, err := Encode()
// 	if err != nil {
// 		fmt.Println("Error while encoding the data", err)
// 		return
// 	}

// 	edata := make([]byte, 0)
// 	for i := 0; i < 5; i++ {
// 		edata = append(edata, data[i]...)
// 	}

// 	// Print the data
// 	fmt.Println("ENCODED:", data)

// 	// Flip 50 random bits in the data
// 	for shard := 0; shard < 3; shard++ {
// 		for i := 0; i < 10; i++ {
// 			data[shard][i] ^= 1
// 		}
// 	}

// 	// Flip random bits in data[8]
// 	for i := 0; i < 10; i++ {
// 		data[8][i] ^= 1
// 	}

// 	// Convert edata to [][]byte

// 	// Decode the data
// 	decoded, err := Decode(data)
// 	if err != nil {
// 		fmt.Println("Error while decoding the data", err)
// 		return
// 	}

// 	// Print the decoded data
// 	fmt.Println("DECODED:", decoded)

// 	// Compare edata and decoded data
// 	if string(edata) != string(decoded) {
// 		fmt.Println("!!!! DATA_MISMATCH !!!!")
// 	}
// }
