package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

type LogbookAuthTokens struct {
	AccessToken  string
	RefreshToken string
}

type Logbook struct {
	config *Configuration
	db     *bolt.DB
}

func NewLogbook(config *Configuration) (*Logbook, error) {
	dir := filepath.Dir(config.ConfigFile)
	dbfile := filepath.Join(dir, "pm5.db")

	db, err := bolt.Open(dbfile, 0600, nil)
	if err != nil {
		log.WithError(err).WithField("file", dbfile).Error("cannot open db")
		return nil, err
	}

	return &Logbook{
		config: config,
		db:     db,
	}, nil
}

func (lb *Logbook) Close() {
	lb.db.Close()
}

func (lb *Logbook) Authenticate() {
	fmt.Println("Please visit:")
	fmt.Println("https://log.concept2.com/oauth/authorize?client_id=&scope=user:read,results:write&response_type=code&redirect_uri=https://auth.pm5-book.worker.dev/c2")
}

func (lb *Logbook) GetTokens() (*LogbookAuthTokens, error) {
	toks := LogbookAuthTokens{}

	if err := lb.db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("c2_auth"))
		if err != nil {
			return err
		}
		toks.AccessToken = string(b.Get([]byte("access")))
		toks.RefreshToken = string(b.Get([]byte("refresh")))
		return nil
	}); err != nil {
		log.WithError(err).Error("error retrieving tokens")
		return nil, err
	}

	return &toks, nil
}

func (lb *Logbook) SetTokens(access, refresh string) error {
	err := lb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("c2_auth"))
		if err != nil {
			return err
		}
		if err := b.Put([]byte("access"), []byte(access)); err != nil {
			return err
		}
		if err := b.Put([]byte("refresh"), []byte(refresh)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("error storing tokens")
		return err
	}

	log.Info("saved tokens")

	return nil
}

func (lb *Logbook) PublishWorkout(w WorkoutData) error {
	return nil
}

func (lb *Logbook) accessToken() string {
	toks, err := lb.GetTokens()
	if err != nil {
		log.WithError(err).Fatal("error reading tokens")
	}

	endpoint := lb.config.LogbookEndpoint + "/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", "")
	data.Set("client_secret", toks.AccessToken)
	data.Set("refresh_token", toks.RefreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("scope", "user:read,results:write")

	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.WithError(err).Fatal("error building refresh token request")
	}

	r.Header.Add("content-type", "application/x-www-form-urlencoded")
	r.Header.Add("content-length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		log.WithError(err).Fatal("error fetching refresh token")
	}
	defer res.Body.Close()

	log.WithField("code", res.Status).Info("got response")

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Fatal("error reading response")
	}

	var payload map[string]interface{}
	json.Unmarshal(body, &payload)

	return payload["access_token"].(string)
}
