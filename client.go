package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 7,
		},
	}
}

func (c *Client) PostForm(uri string, data url.Values) (map[string]interface{}, error) {
	req, _ := http.NewRequest(http.MethodPost, uri, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			log.WithError(err).WithField("status", res.Status).Error("error posting form")
		} else {
			log.WithFields(log.Fields{
				"uri":    uri,
				"status": res.Status,
				"body":   string(body),
			}).Error("error posting form")
		}
		return nil, fmt.Errorf(res.Status)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Client) Post(uri, body string, headers map[string]string) error {
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(body))
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			log.WithError(err).WithField("status", res.Status).Error("non-200 posting")
		} else {
			log.WithFields(log.Fields{
				"uri":    uri,
				"status": res.Status,
				"body":   string(body),
			}).Error("non-200 posting")
		}
		return fmt.Errorf(res.Status)
	}

	return nil
}

func (c *Client) GetJSON(uri string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		log.WithFields(log.Fields{
			"uri":    uri,
			"status": res.Status,
			"body":   string(body),
		}).Error("error getting json")

		return nil, fmt.Errorf(res.Status)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}

	return m, nil
}
