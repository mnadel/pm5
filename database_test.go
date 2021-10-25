package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseWriteRecord(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	db.CreateWorkout(&WorkoutDBRecord{
		Data: []byte{0xc, 0xa, 0xf, 0xe, 0xb, 0xa, 0xb, 0xe},
	})

	c = MustInt(db.Count())
	assert.Equal(t, 1, c)
}

func TestDatabaseGetWorkout(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	wo := &WorkoutDBRecord{
		Data: []byte{0xc, 0xa, 0xf, 0xe, 0xb, 0xa, 0xb, 0xe},
	}

	db.CreateWorkout(wo)
	assert.NotEqual(t, 0, wo.ID)

	c = MustInt(db.Count())
	assert.Equal(t, 1, c)

	fetched, err := db.GetWorkout(wo.ID)
	assert.NoError(t, err)
	assert.Equal(t, wo.ID, fetched.ID)
	assert.Equal(t, []byte{0xc, 0xa, 0xf, 0xe, 0xb, 0xa, 0xb, 0xe}, fetched.Data)
}

func TestDatabaseGetPendingRecords(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	for i := 0; i < 5; i++ {
		db.CreateWorkout(&WorkoutDBRecord{
			Data: []byte{byte(i)},
		})
	}

	c = MustInt(db.Count())
	assert.Equal(t, 5, c)

	db.MarkSent(1)
	db.MarkSent(3)
	db.MarkSent(5)

	pending, err := db.GetPendingWorkouts()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pending))
	assert.Equal(t, uint64(2), pending[0].ID)
	assert.Equal(t, uint64(4), pending[1].ID)
}

func TestDatabaseMarkSent(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	for i := 0; i < 5; i++ {
		wo := &WorkoutDBRecord{
			Data: []byte{byte(i)},
		}
		db.CreateWorkout(wo)
		db.MarkSent(wo.ID)
	}

	c = MustInt(db.Count())
	assert.Equal(t, 5, c)

	pending, err := db.GetPendingWorkouts()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pending))
}
