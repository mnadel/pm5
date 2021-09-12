package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCharacteristic(t *testing.T) {
	device := NewPM5Device(&Configuration{})

	workoutChar := device.FindRowingCharacteristic(MustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"))
	assert.NotNil(t, workoutChar)
	assert.Equal(t, "workout", workoutChar.Name)
}

func TestIsPM5(t *testing.T) {
	assert.True(t, IsPM5("PM5 431409475 Row"))
	assert.True(t, IsPM5("PM5 5882300 Row"))
	assert.False(t, IsPM5("PM5"))
	assert.False(t, IsPM5("Row"))
}
