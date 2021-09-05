package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGetTimezone(t *testing.T) {
	assert.NotNil(t, mustGetTimezone("America/Chicago"))
}
