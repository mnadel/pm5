package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
)

type Logbook struct {
	db     *Database
	client *Client
	config *Configuration
	auth   *AuthRecord
}

func NewLogbook(config *Configuration, db *Database, client *Client) *Logbook {
	return &Logbook{
		db:     db,
		client: client,
		config: config,
	}
}

func (l *Logbook) PostWorkout(wo *WorkoutData) error {
	if l.auth == nil {
		auth, err := l.db.GetAuth()
		if err != nil {
			return err
		}

		l.auth = auth

		if newAuth, err := RefreshAuth(l.config, l.client, auth); err != nil {
			if err := l.db.SetAuth(newAuth.Token, newAuth.Refresh); err != nil {
				log.WithError(err).Info("unable to save new tokens")
			}
		} else {
			log.WithError(err).Info("unable to get new tokens")
		}
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", l.auth.Token),
		"Content-Type":  "application/json",
	}

	m := map[string]interface{}{
		"type":         "rower",
		"date":         wo.LogEntry.Format(ISO8601),
		"distance":     uint64(wo.Distance),
		"time":         uint64(wo.ElapsedTime.Seconds() * 10),
		"weight_class": "H",
		"workout_type": wo.WorkoutType.AsString(),
		"stroke_rate":  wo.AverageStrokeRate,
		"drag_factor":  wo.AvgDragFactor,
		"pace":         uint64(wo.AvgPace.Seconds() * 10),
	}

	payload, _ := json.Marshal(m)

	uri := fmt.Sprintf("https://%s/api/users/me/results", l.config.LogbookHost)

	return l.client.Post(uri, string(payload), headers)
}

func RefreshAuth(config *Configuration, client *Client, currentAuth *AuthRecord) (*AuthRecord, error) {
	if os.Getenv("PM5_CLIENT_SECRET") == "" {
		return nil, fmt.Errorf("missing: PM5_CLIENT_SECRET")
	}

	uri := fmt.Sprintf("https://%s/oauth/access_token", config.LogbookHost)

	data := url.Values{}
	data.Set("client_id", "ymMRExBCsS6HqDm9ShMEPRvpR3Hh2DPb3FTtiazX")
	data.Set("client_secret", os.Getenv("PM5_CLIENT_SECRET"))
	data.Set("grant_type", "refresh_token")
	data.Set("scope", "results:write")
	data.Set("refresh_token", currentAuth.Refresh)

	resp, err := client.PostForm(uri, data)
	if err != nil {
		return nil, err
	}

	if val, ok := resp["access_token"]; !ok {
		return nil, fmt.Errorf("didn't find access_token")
	} else if val == "" {
		return nil, fmt.Errorf("cannot get access_token")
	}

	if val, ok := resp["refresh_token"]; !ok {
		return nil, fmt.Errorf("didn't find refresh_token")
	} else if val == "" {
		return nil, fmt.Errorf("cannot get refresh_token")
	}

	return &AuthRecord{
		Token:   resp["access_token"].(string),
		Refresh: resp["refresh_token"].(string),
	}, nil
}
