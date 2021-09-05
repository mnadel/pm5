package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCharacteristic(t *testing.T) {
	device := NewPM5Device(&Configuration{})

	workoutChar := device.FindCharacteristic(mustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"))
	assert.NotNil(t, workoutChar)
	assert.Equal(t, "workout", workoutChar.Name)
}
