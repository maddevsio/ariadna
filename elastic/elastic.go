package elastic

import (
	es "github.com/elastic/go-elasticsearch/v7"
)

type Client struct {
	Conn *es.Client
}

func New(addresses []string) (*Client, error) {
	c, err := es.NewClient(es.Config{
		Addresses: addresses,
	})
	if err != nil {
		return nil, err
	}
	return &Client{Conn: c}, nil
}
