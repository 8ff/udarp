package viterbi_codec_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	viterbi_codec "github.com/8ff/udarp/pkg/codecs/viterbi"
)

func TestRandomData(t *testing.T) {

	// Generate random input data.
	inputData := make([]byte, 100)
	rand.Read(inputData)

	// Set encoder and decoder parameters.
	constraint := 7
	polynomials := []int{79, 109}
	// If this is set to true, test will FAIL as we are using separate encoder and decoder.
	reversePolynomials := false

	// Initialize a codec.
	codec, err := viterbi_codec.Init(viterbi_codec.Params{Constraint: constraint, Polynomials: polynomials, ReversePolynomials: reversePolynomials})
	if err != nil {
		t.Fatalf("Init failed with error: %v", err)
	}

	// Encode input data.
	encodedData, err := viterbi_codec.Encode(codec, inputData)
	if err != nil {
		t.Fatalf("Encode failed with error: %v", err)
	}

	// Decode encoded data.
	decodedData, err := viterbi_codec.Decode(codec, encodedData)
	if err != nil {
		t.Fatalf("Decode failed with error: %v", err)
	}

	// Check that decoded data matches input data.
	if !bytes.Equal(inputData, decodedData) {
		t.Fatalf("Decoded data doesn't match input data")
	}
}
