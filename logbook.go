package main

import (
	"encoding/json"
	"fmt"
	"net/url"

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

// see: https://log.concept2.com/developers/documentation/#authentication-access-token-post
func (l *Logbook) PostWorkout(wo *WorkoutData) error {
	if l.auth == nil {
		if auth, err := l.db.GetAuth(); err != nil {
			return err
		} else {
			l.auth = l.tryGetNewRefreshToken(auth)
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

func (l *Logbook) tryGetNewRefreshToken(currentAuth *AuthRecord) *AuthRecord {
	log.Info("attempting to get new refresh token")

	newAuth, err := RefreshAuth(l.config, l.client, currentAuth)
	if err != nil {
		log.WithError(err).Info("unable to get new tokens")
		return currentAuth
	}

	err = l.db.SetAuth(newAuth.Token, newAuth.Refresh)
	if err != nil {
		log.WithFields(log.Fields{
			"new_token":   newAuth.Token,
			"new_refresh": newAuth.Refresh,
		}).Info("unable to save these tokens")
	} else {
		log.Info("saved new refresh token")
	}

	return newAuth
}

// see https://log.concept2.com/developers/documentation/#authentication-access-token-post
func RefreshAuth(config *Configuration, client *Client, currentAuth *AuthRecord) (*AuthRecord, error) {
	if PM5_OAUTH_SECRET == "" {
		panic("missing: PM5_OAUTH_SECRET")
	}

	uri := fmt.Sprintf("https://%s/oauth/access_token", config.LogbookHost)

	data := url.Values{}
	data.Set("client_id", PM5_OAUTH_APPID)
	data.Set("client_secret", PM5_OAUTH_SECRET)
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
