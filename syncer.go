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

func (s *Syncer) Start() {
	go func() {
	loop:
		for {
			timer := time.NewTimer(time.Second * 30)
			<-timer.C

			select {
			case <-s.cancelCh:
				break loop
			default:
			}

			pendings, err := s.db.GetPendingWorkouts()
			if err != nil {
				log.WithError(err).Error("cannot get workouts to sync")
				continue
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

				if err := s.logbook.PostWorkout(parsed); err != nil {
					log.WithError(err).WithField("id", pending.ID).Error("error posting workout")
				} else if err := s.db.MarkSent(pending.ID); err != nil {
					log.WithError(err).WithField("id", pending.ID).Error("error marking workout sent")
				}
			}
		}

		log.Info("syncer shut down")
	}()
}
