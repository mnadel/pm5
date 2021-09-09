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

func DecodeDateTime(data []byte) time.Time {
	n := DecodeTwoByteNumber(data)

	month := n & 0x0f
	day := (n >> 4) & 0x1f
	year := 2000 + ((n >> 9) & 0x7f)

	min := DecodeByteNumber(data[2])
	hr := DecodeByteNumber(data[3])

	return time.Date(int(year), time.Month(month), int(day), int(hr), int(min), 0, 0, time.Local)
}
