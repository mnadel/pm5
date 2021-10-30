package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	db *bolt.DB
}

func NewDatabase(c *Configuration) *Database {
	(&FileManager{}).Mkdirs(c.DBFile)

	db, err := bolt.Open(c.DBFile, 0644, nil)
	if err != nil {
		log.WithError(err).WithField("db", c.DBFile).Fatal("cannot open db")
	}

	d := &Database{db}
	if err := d.init(); err != nil {
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
	if err := user.Validate(); err != nil {
		return err
	}

	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		encoded, err := user.AsGob()
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

		for k, v := c.First(); k != nil; k, _ = c.Next() {
			if user, err := (&User{}).FromGob(v); err != nil {
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

		wo, err := (&Workout{}).FromGob(v)
		if err != nil {
			return err
		}

		wo.SentAt = time.Now()

		encoded, err := wo.AsGob()
		if err != nil {
			return err
		}

		return b.Put(Itob(wo.ID), encoded)
	})
}

func (d *Database) GetWorkout(id uint64) (*Workout, error) {
	var rec *Workout

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return nil
		}

		if r, err := (&Workout{}).FromGob(b.Get(Itob(id))); err != nil {
			return err
		} else {
			rec = r
		}

		return nil
	})

	return rec, err
}

func (d *Database) UpdateWorkout(w *Workout) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("workouts"))
		if err != nil {
			return err
		}

		if w.ID == 0 {
			return fmt.Errorf("record missing id")
		}

		encoded, err := w.AsGob()
		if err != nil {
			return err
		}

		return b.Put(Itob(w.ID), encoded)
	})
}

func (d *Database) CreateWorkout(w *Workout) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("workouts"))
		if err != nil {
			return err
		}
		id, _ := b.NextSequence()

		// update record
		w.ID = id
		w.CreatedAt = time.Now()

		encoded, err := w.AsGob()
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

func (d *Database) GetPendingWorkouts() ([]*Workout, error) {
	data := make([]*Workout, 0)

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return fmt.Errorf("cannot get workout bucket")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			wo, err := (&Workout{}).FromGob(v)
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

func (d *Database) GetWorkouts() ([]*Workout, error) {
	data := make([]*Workout, 0)

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("workouts"))
		if b == nil {
			return fmt.Errorf("cannot get workout bucket")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			wo, err := (&Workout{}).FromGob(v)
			if err != nil {
				return err
			}

			data = append(data, wo)
		}

		return nil
	})

	return data, err
}

func (d *Database) init() error {
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
