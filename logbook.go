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
	user   *User
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
	if l.user == nil {
		if user, err := l.db.GetUser(PM5_USER_UUID); err != nil {
			return err
		} else {
			l.user = user
			if err := l.tryGetNewRefreshToken(); err != nil {
				return err
			}
		}
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", l.user.Token),
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

func (l *Logbook) tryGetNewRefreshToken() error {
	log.Info("attempting to get new refresh token")

	if err := RefreshAuth(l.config, l.client, l.user); err != nil {
		log.WithError(err).Info("unable to get new tokens")
	}

	if err := l.db.UpsertUser(l.user); err != nil {
		log.WithFields(log.Fields{
			"new_token":   l.user.Token,
			"new_refresh": l.user.Refresh,
		}).Info("unable to save these tokens")

		return err
	} else {
		log.Info("saved new refresh token")
	}

	return nil
}

// see https://log.concept2.com/developers/documentation/#authentication-access-token-post
func RefreshAuth(config *Configuration, client *Client, user *User) error {
	if PM5_OAUTH_SECRET == "" {
		panic("missing: PM5_OAUTH_SECRET")
	}

	uri := fmt.Sprintf("https://%s/oauth/access_token", config.LogbookHost)

	data := url.Values{}
	data.Set("client_id", PM5_OAUTH_APPID)
	data.Set("client_secret", PM5_OAUTH_SECRET)
	data.Set("grant_type", "refresh_token")
	data.Set("scope", "results:write")
	data.Set("refresh_token", user.Refresh)

	resp, err := client.PostForm(uri, data)
	if err != nil {
		return err
	}

	if val, ok := resp["access_token"]; !ok {
		return fmt.Errorf("didn't find access_token")
	} else if val == "" {
		return fmt.Errorf("cannot get access_token")
	}

	if val, ok := resp["refresh_token"]; !ok {
		return fmt.Errorf("didn't find refresh_token")
	} else if val == "" {
		return fmt.Errorf("cannot get refresh_token")
	}

	user.Token = resp["access_token"].(string)
	user.Refresh = resp["refresh_token"].(string)

	return nil
}
