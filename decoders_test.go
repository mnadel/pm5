package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeByteNumber(t *testing.T) {
	data := byte(3)
	actual := DecodeByteNumber(data)
	assert.Equal(t, uint8(3), actual)
}

func TestDecodeTwoByteNumber(t *testing.T) {
	data := []byte{1, 1}
	actual := DecodeTwoByteNumber(data)
	assert.Equal(t, uint16(257), actual)
}

func TestDecodeThreeByteNumber(t *testing.T) {
	data := []byte{0, 0, 1}
	actual := DecodeThreeByteNumber(data)
	assert.Equal(t, uint32(65536), actual)
}

func TestDecodeThreeByteDuration(t *testing.T) {
	data := []byte{1, 1, 1} // 65793
	elapsed := float32(DecodeThreeByteNumber(data))
	dur := DecodeDuration(elapsed, 0.01)
	assert.Equal(t, MustParseDuration("657.93s"), dur) // 65793/100
}

func TestDecodeTwoByteDuration(t *testing.T) {
	data := []byte{1, 1} // 257
	elapsed := float32(DecodeTwoByteNumber(data))
	dur := DecodeDuration(elapsed, 0.1)
	assert.Equal(t, MustParseDuration("25.7s"), dur) // 257/10
}
