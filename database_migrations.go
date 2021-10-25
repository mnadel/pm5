package main

import log "github.com/sirupsen/logrus"

type DBMigrator struct {
	db         *Database
	migrations []Migration
}

type Migration func(dbm DBMigrator, rec *WorkoutDBRecord, wo *WorkoutData) (bool, error)

func NewDBMigrator(db *Database) *DBMigrator {
	return &DBMigrator{
		db:         db,
		migrations: []Migration{DBMigrator.migration_20211025},
	}
}

func (dbm *DBMigrator) Migrate() error {
	records, err := dbm.db.GetWorkouts()
	if err != nil {
		return err
	}

	for _, record := range records {
		if migrated, err := dbm.MigrateRecord(record); err != nil {
			return err
		} else if migrated {
			log.WithField("id", record.ID).Info("saving migrations")

			if err := dbm.db.UpdateWorkout(record); err != nil {
				log.WithError(err).WithField("id", record.ID).Error("cannot migrating")
				return err
			}
		} else {
			log.WithField("id", record.ID).Info("no migrations applied")
		}
	}

	return nil
}

func (dbm *DBMigrator) MigrateRecord(rec *WorkoutDBRecord) (bool, error) {
	raw := rec.Decode()
	decoded := raw.Decode()

	var updatedRecord bool

	for i, migration := range dbm.migrations {
		log.WithFields(log.Fields{
			"id":        rec.ID,
			"bytes":     rec.Data,
			"raw":       raw,
			"decoded":   decoded,
			"migration": i,
		}).Info("checking record")

		if migrated, err := migration(*dbm, rec, decoded); err != nil {
			return false, err
		} else if migrated {
			log.WithFields(log.Fields{
				"id":        rec.ID,
				"migration": i,
			}).Info("applied migration")

			updatedRecord = true
		}
	}

	return updatedRecord, nil
}

func (dbm DBMigrator) migration_20211025(rec *WorkoutDBRecord, wo *WorkoutData) (bool, error) {
	if rec.CreatedAt.IsZero() {
		log.WithFields(log.Fields{
			"id":         rec.ID,
			"created_at": wo.LogEntry,
		}).Info("setting CreatedAt")

		rec.CreatedAt = wo.LogEntry

		return true, nil
	}

	return false, nil
}
