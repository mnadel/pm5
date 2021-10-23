package main

import "fmt"

type Logbook struct {
	db     *Database
	client *Client
}

func NewLogbook(db *Database, client *Client) *Logbook {
	return &Logbook{
		db:     db,
		client: client,
	}
}

func (l *Logbook) PostWorkout(wo *WorkoutData) error {
	_, err := l.db.GetAuth()
	if err != nil {
		return err
	}

	headers, err := l.newHeaders()
	if err != nil {
		return err
	}

	return l.client.Post("", wo.AsJSON(), headers)
}

func (l *Logbook) newHeaders() (map[string]string, error) {
	headers := make(map[string]string)

	tok, err := l.bearerToken()
	if err != nil {
		return nil, err
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %s", tok)

	return headers, nil
}

func (l *Logbook) bearerToken() (string, error) {
	headers, err := l.newHeaders()
	if err != nil {
		return "", err
	}

	resp, err := l.client.GetJSON("", headers)
	if err != nil {
		return "", err
	}

	if v, ok := resp["access_token"]; ok {
		return v.(string), nil
	} else {
		return "", fmt.Errorf("cannot find access_token")
	}
}
