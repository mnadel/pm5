package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	db *bolt.DB
}

type WorkoutDBRecord struct {
	ID     uint64
	Data   []byte
	SentAt time.Time
}

func NewDatabase(c *Configuration) *Database {
	db, err := bolt.Open(c.DBFile, 0644, nil)
	if err != nil {
		log.WithError(err).WithField("db", c.DBFile).Fatal("cannot open db")
	}

	return &Database{db}
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) Stats() bolt.Stats {
	return d.db.Stats()
}

func (d *Database) MarkSent(id uint64) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return fmt.Errorf("cannot get workouts bucket")
		}

		v := b.Get(Itob(id))
		if v == nil {
			return fmt.Errorf("record not found: %d", id)
		}

		wo, err := DecodeWorkoutRecord(v)
		if err != nil {
			return err
		}

		wo.SentAt = time.Now()

		encoded, err := EncodeWorkoutRecord(wo)
		if err != nil {
			return err
		}

		return b.Put(Itob(wo.ID), encoded)
	})
}

func (d *Database) GetWorkout(id uint64) (*WorkoutDBRecord, error) {
	var rec *WorkoutDBRecord

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return nil
		}

		if r, err := DecodeWorkoutRecord(b.Get(Itob(id))); err != nil {
			return err
		} else {
			rec = r
		}

		return nil
	})

	return rec, err
}

func (d *Database) SaveWorkout(w *WorkoutDBRecord) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("workouts"))
		if err != nil {
			return err
		}
		id, _ := b.NextSequence()

		// update record
		w.ID = id

		encoded, err := EncodeWorkoutRecord(w)
		if err != nil {
			return err
		}

		return b.Put(Itob(id), encoded)
	})
}

func (d *Database) Count() (int, error) {
	var count int

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}

		return nil
	})

	return count, err
}

func (d *Database) GetPendingWorkouts() ([]*WorkoutDBRecord, error) {
	var data []*WorkoutDBRecord

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return fmt.Errorf("cannot get workout bucket")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			wo, err := DecodeWorkoutRecord(v)
			if err != nil {
				return err
			}

			if wo.SentAt.IsZero() {
				data = append(data, wo)
			}
		}

		return nil
	})

	return data, err
}

func (d *Database) GetWorkouts() ([]*WorkoutDBRecord, error) {
	var data []*WorkoutDBRecord

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return fmt.Errorf("cannot get workout bucket")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			wo, err := DecodeWorkoutRecord(v)
			if err != nil {
				return err
			}

			data = append(data, wo)
		}

		return nil
	})

	return data, err
}

func (wr *WorkoutDBRecord) Decode() *RawWorkoutData {
	return ReadWorkoutData(wr.Data)
}

func EncodeWorkoutRecord(r *WorkoutDBRecord) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeWorkoutRecord(data []byte) (*WorkoutDBRecord, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var wr WorkoutDBRecord
	if err := dec.Decode(&wr); err != nil {
		return nil, err
	}

	return &wr, nil
}
