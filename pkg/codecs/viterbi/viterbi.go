package viterbi_codec

import (
	"encoding/binary"
	"fmt"

	"github.com/8ff/udarp/pkg/crc"
	"github.com/8ff/viterbi"
)

type Params struct {
	Constraint         int
	Polynomials        []int
	ReversePolynomials bool
}

// Function that does Init and returns viterbi_codec
func Init(params Params) (*viterbi.ViterbiCodec, error) {
	// Initialize a codec.
	codec, err := viterbi.Init(viterbi.Input{Constraint: params.Constraint, Polynomials: params.Polynomials, ReversePolynomials: params.ReversePolynomials})
	if err != nil {
		return nil, err
	}

	return codec, nil
}

func Encode(codec *viterbi.ViterbiCodec, data []byte) ([]int, error) {
	// Add CRC.
	crc16 := crc.Encode16(data)
	crcBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(crcBytes, crc16)
	crcData := append(data, crcBytes...)

	// Encode data.
	encodedBits := codec.EncodeBytes(crcData)

	return viterbi.BitsToInts(encodedBits), nil
}

func Decode(codec *viterbi.ViterbiCodec, data []int) ([]byte, error) {
	// Decode data.
	decodedBits := codec.Decode(viterbi.IntsToBits(data))

	// Strip and verify CRC.
	decodedBytes := viterbi.BitsToBytes(decodedBits)
	decodedData := decodedBytes[:len(decodedBytes)-2]
	decodedCrc := decodedBytes[len(decodedBytes)-2:]

	decodedCrc16 := binary.BigEndian.Uint16(decodedCrc)
	if !crc.Match16(decodedData, decodedCrc16) {
		return nil, fmt.Errorf("CRC mismatch")
	} else {
		return decodedData, nil
	}
}
