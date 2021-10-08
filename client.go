package main

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Post(uri, body string) error {
	return nil
}
