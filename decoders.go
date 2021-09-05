package main

import (
	"encoding/binary"
	"time"
)

func DecodeByteNumber(data byte) uint8 {
	return uint8(data)
}

func DecodeTwoByteNumber(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data[0:2])
}

func DecodeThreeByteNumber(data []byte) uint32 {
	b := make([]byte, 4)
	copy(b[0:], data[0:3])
	return binary.LittleEndian.Uint32(b)
}

func DecodeDuration(seconds, factor float32) time.Duration {
	millis := seconds * 1000.0 * factor
	return time.Duration(millis) * time.Millisecond
}
