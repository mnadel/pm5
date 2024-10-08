package main

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGobEncodingMigrations(t *testing.T) {
	type A struct {
		ID  int
		Foo string
	}

	type B struct {
		ID   int
		Foo  string
		Time time.Time
	}

	a := A{3, "three"}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(a); err != nil {
		assert.NoError(t, err)
	}

	bufRead := bytes.NewBuffer(buf.Bytes())
	dec := gob.NewDecoder(bufRead)

	var b B
	if err := dec.Decode(&b); err != nil {
		assert.NoError(t, err)
	}

	assert.Equal(t, b.ID, 3)
	assert.Equal(t, b.Foo, "three")
	assert.True(t, b.Time.IsZero())
}

func TestGobEncodingUsers(t *testing.T) {
	user := &User{
		UUID:    "a",
		Token:   "b",
		Refresh: "c",
	}

	encoded, err := user.AsGob()
	assert.NoError(t, err)

	decoded, err := (&User{}).FromGob(encoded)
	assert.NoError(t, err)

	assert.Equal(t, "a", decoded.UUID)
	assert.Equal(t, "b", decoded.Token)
	assert.Equal(t, "c", decoded.Refresh)
}
