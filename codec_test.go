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

func TestDecodeDateTime(t *testing.T) {
	d := DecodeDateTime([]byte{8, 43, 22, 18})
	assert.Equal(t, "2021-08-16 18:22", d.Format("2006-01-02 15:04"))

	d2 := DecodeDateTime([]byte{89, 42, 38, 17})
	assert.Equal(t, "2021-09-05 17:38", d2.Format("2006-01-02 15:04"))
}
