package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Logbook struct {
	db     *Database
	client *Client
	config *Configuration
}

func NewLogbook(config *Configuration, db *Database, client *Client) *Logbook {
	return &Logbook{
		db:     db,
		client: client,
		config: config,
	}
}

// see: https://log.concept2.com/developers/documentation/#authentication-access-token-post
func (l *Logbook) PostWorkout(user *User, wo *WorkoutData) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", user.Token),
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

// see https://log.concept2.com/developers/documentation/#authentication-access-token-post
func (l *Logbook) RefreshAuth(user *User) error {
	if l.config.OAuthSecret == "" {
		panic("missing: PM5_OAUTH_SECRET")
	}

	uri := fmt.Sprintf("https://%s/oauth/access_token", l.config.LogbookHost)

	data := url.Values{}
	data.Set("client_id", PM5_OAUTH_APPID)
	data.Set("client_secret", l.config.OAuthSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("scope", "results:write")
	data.Set("refresh_token", user.Refresh)

	resp, err := l.client.PostForm(uri, data)
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
