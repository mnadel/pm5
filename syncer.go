package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Syncer struct {
	db       *Database
	logbook  *Logbook
	cancelCh chan struct{}
}

func NewSyncer(logbook *Logbook, db *Database) *Syncer {
	return &Syncer{
		logbook:  logbook,
		db:       db,
		cancelCh: make(chan struct{}, 1),
	}
}

func (s *Syncer) Close() {
	s.cancelCh <- struct{}{}
}

func (s *Syncer) Sync() {
	pendings, err := s.db.GetPendingWorkouts()
	if err != nil {
		log.WithError(err).Error("cannot get workouts to sync")
		return
	} else {
		log.WithField("count", len(pendings)).Info("found records to sync")
	}

	for _, pending := range pendings {
		raw := ReadWorkoutData(pending.Data)
		parsed := raw.Decode()

		log.WithFields(log.Fields{
			"id": pending.ID,
			"dt": parsed.LogEntry.Format(ISO8601),
		}).Info("syncing record")

		err := s.logbook.PostWorkout(parsed)
		if err == nil || err.Error() == "409 Conflict" { // if no error or already sent, mark sent
			if err := s.db.MarkSent(pending.ID); err != nil {
				log.WithError(err).WithField("id", pending.ID).Error("error marking workout sent")
			}
		} else {
			log.WithError(err).WithField("id", pending.ID).Error("error posting workout")
		}
	}
}

func (s *Syncer) Start() {
	go func() {
		s.Sync()

	loop:
		for {
			timer := time.NewTimer(time.Second * 30)
			<-timer.C

			select {
			case <-s.cancelCh:
				break loop
			default:
			}

			s.Sync()
		}

		log.Info("syncer shut down")
	}()
}
