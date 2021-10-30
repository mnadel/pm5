package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseWriteRecord(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	db.CreateWorkout(&Workout{
		Data: []byte{0xc, 0xa, 0xf, 0xe, 0xb, 0xa, 0xb, 0xe},
	})

	c = MustInt(db.Count())
	assert.Equal(t, 1, c)
}

func TestDatabaseGetWorkout(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	c := MustInt(db.Count())
	assert.Equal(t, 0, c)

	wo := &Workout{
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
		db.CreateWorkout(&Workout{
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
		wo := &Workout{
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

func TestDatabaseGetUsers(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())

	users, err := db.GetUsers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(users))

	for i := 0; i < 5; i++ {
		db.UpsertUser(&User{
			UUID:    fmt.Sprintf("user-%d", i),
			Token:   "b",
			Refresh: "c",
		})
	}

	users, err = db.GetUsers()
	assert.NoError(t, err)
	assert.Equal(t, 5, len(users))
}

func TestDatabaseGetUser(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	db.UpsertUser(&User{
		UUID:    "a",
		Token:   "b",
		Refresh: "c",
	})

	user, err := db.GetUser("a")
	assert.NoError(t, err)
	assert.Equal(t, "a", user.UUID)
	assert.Equal(t, "b", user.Token)
	assert.Equal(t, "c", user.Refresh)
}

func TestDatabaseUpsertExistingUser(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())
	db.UpsertUser(&User{
		UUID:    "a",
		Token:   "b",
		Refresh: "c",
	})

	// one user
	users, err := db.GetUsers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))

	// whose id is a and token is b
	user, err := db.GetUser("a")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "a", user.UUID)
	assert.Equal(t, "b", user.Token)

	// update token
	user.Token = "foo"
	db.UpsertUser(user)

	// still only one user
	users, err = db.GetUsers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))

	// re-fetch, check token changed
	user, err = db.GetUser("a")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "a", user.UUID)
	assert.Equal(t, "foo", user.Token)
}

func TestDatabaseUpsertNewUser(t *testing.T) {
	db := NewDatabase(NewTestConfiguration())

	db.UpsertUser(&User{
		UUID:    "a",
		Token:   "b",
		Refresh: "c",
	})

	users, err := db.GetUsers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "a", users[0].UUID)
}
