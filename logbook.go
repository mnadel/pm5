package main

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
	return l.client.Post("", wo.AsJSON())
}
