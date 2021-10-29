package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type WorkoutDBRecord struct {
	ID        uint64
	Data      []byte
	SentAt    time.Time
	CreatedAt time.Time
	UserUUID  string
}

type User struct {
	UUID    string
	Token   string
	Refresh string
}

func (wr *WorkoutDBRecord) Decode() *RawWorkoutData {
	return ReadWorkoutData(wr.Data)
}

func (wr *WorkoutDBRecord) AsGob() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(wr); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (wr *WorkoutDBRecord) FromGob(data []byte) (*WorkoutDBRecord, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var rec WorkoutDBRecord
	if err := dec.Decode(&rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (u *User) Validate() error {
	if u.UUID == "" {
		return fmt.Errorf("user is missing uuid")
	} else if u.Token == "" {
		return fmt.Errorf("user is missing auth token")
	} else if u.Refresh == "" {
		return fmt.Errorf("user is missing refresh token")
	}

	return nil
}

func (u *User) AsGob() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(u); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (u *User) FromGob(data []byte) (*User, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var user User
	if err := dec.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
