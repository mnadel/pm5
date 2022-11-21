package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Syncer struct {
	db           *Database
	logbook      *Logbook
	cancelCh     chan struct{}
	logLimiter   *RateLimiter
	alertLimiter *RateLimiter
}

func NewSyncer(logbook *Logbook, db *Database) *Syncer {
	return &Syncer{
		logbook:      logbook,
		db:           db,
		cancelCh:     make(chan struct{}, 1),
		logLimiter:   NewRateLimiter(time.Minute * 5),
		alertLimiter: NewRateLimiter(time.Hour * 1),
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
	}

	s.logLimiter.MaybePerform(func() {
		log.WithField("count", len(pendings)).Info("found records to sync")
	})

	for _, pending := range pendings {
		raw := ReadWorkoutData(pending.Data)
		parsed := raw.Decode()

		user, err := s.db.GetUser(pending.UserUUID)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": pending.UserUUID,
				"id":   pending.ID,
			}).Error("cannot find user")

			continue
		}

		if err := s.logbook.RefreshAuth(user); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": user,
			}).Error("cannot refresh auth")

			continue
		} else if err = s.db.UpsertUser(user); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": user,
			}).Warn("cannot upsert auth")
		}

		log.WithFields(log.Fields{
			"user":  user.UUID,
			"token": user.Token,
			"id":    pending.ID,
			"dt":    parsed.LogEntry.Format(ISO8601),
		}).Info("syncing record")

		err = s.logbook.PostWorkout(user, parsed)
		if err == nil || err.Error() == "409 Conflict" { // if no error or already sent, mark sent
			if err := s.db.MarkSent(pending.ID); err != nil {
				log.WithError(err).WithField("id", pending.ID).Error("error marking workout sent")
			}
		} else {
			log.WithError(err).WithField("id", pending.ID).Error("error posting workout")

			if s.logbook.config.SlackNotificationURL != "" {
				s.alertLimiter.MaybePerform(func() {
					body, err := json.Marshal(map[string]string{
						"text": "PM5 alert: " + err.Error(),
					})

					if err != nil {
						log.WithError(err).WithField("id", pending.ID).Error("error encoding notification")
						return
					}

					buf := bytes.NewBuffer(body)
					_, err = http.Post(s.logbook.config.SlackNotificationURL, "application/json", buf)
					if err != nil {
						log.WithError(err).WithField("id", pending.ID).Error("error notifying")
					}
				})
			}
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
