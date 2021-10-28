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

func NewDatabase(c *Configuration) *Database {
	(&FileManager{}).Mkdirs(c.DBFile)

	db, err := bolt.Open(c.DBFile, 0644, nil)
	if err != nil {
		log.WithError(err).WithField("db", c.DBFile).Fatal("cannot open db")
	}

	d := &Database{db}
	if err := d.initDB(); err != nil {
		panic(err)
	}

	return d
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) Stats() bolt.Stats {
	return d.db.Stats()
}

func (d *Database) UpsertUser(user *User) error {
	if user.UUID == "" {
		return fmt.Errorf("user is missing uuid")
	}

	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		encoded, err := EncodeUserRecord(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(user.UUID), encoded)
	})
}

func (d *Database) GetUsers() ([]*User, error) {
	users := make([]*User, 0)

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf("users bucket not found")
		}

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if user, err := DecodeUserRecord(b.Get(k)); err != nil {
				return err
			} else {
				users = append(users, user)
			}
		}

		return nil
	})

	return users, err
}

func (d *Database) GetUser(uuid string) (*User, error) {
	users, err := d.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.UUID == uuid {
			return user, nil
		}
	}

	return nil, nil
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

func (d *Database) UpdateWorkout(w *WorkoutDBRecord) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("workouts"))
		if err != nil {
			return err
		}

		if w.ID == 0 {
			return fmt.Errorf("record missing id")
		}

		encoded, err := EncodeWorkoutRecord(w)
		if err != nil {
			return err
		}

		return b.Put(Itob(w.ID), encoded)
	})
}

func (d *Database) CreateWorkout(w *WorkoutDBRecord) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("workouts"))
		if err != nil {
			return err
		}
		id, _ := b.NextSequence()

		// update record
		w.ID = id
		w.CreatedAt = time.Now()

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
	data := make([]*WorkoutDBRecord, 0)

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
	data := make([]*WorkoutDBRecord, 0)

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

func (d *Database) initDB() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("workouts")); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists([]byte("users")); err != nil {
			return err
		}

		return nil
	})
}

func (d *Database) PrintDB() error {
	users, err := d.GetUsers()
	if err != nil {
		return err
	}

	fmt.Println("######## users ########")

	for _, user := range users {
		fmt.Println("***** uuid", user.UUID)
		fmt.Println("     token", user.Token)
		fmt.Println("   refresh", user.Refresh)
	}

	workouts, err := d.GetWorkouts()
	if err != nil {
		return err
	}

	fmt.Println("######## workouts ########")

	for _, workout := range workouts {
		raw := ReadWorkoutData(workout.Data)

		fmt.Println("******* id", workout.ID)
		fmt.Println("      user", workout.UserUUID)
		fmt.Println("   created", workout.CreatedAt)
		fmt.Println("      sent", workout.SentAt)
		fmt.Println("      data", workout.Data)
		fmt.Println("   decoded", raw.Decode().AsJSON())
	}

	return nil
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

func EncodeUserRecord(user *User) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(user); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeUserRecord(data []byte) (*User, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var user User
	if err := dec.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
