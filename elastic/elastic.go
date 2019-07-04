package elastic

import (
	"bytes"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/maddevsio/ariadna/config"
)

type Client struct {
	conn   *es.Client
	config *config.Ariadna
}

func New(conf *config.Ariadna) (*Client, error) {
	c, err := es.NewClient(es.Config{
		Addresses: conf.ElasticURLs,
	})
	if err != nil {
		return nil, err
	}
	return &Client{conn: c, config: conf}, nil
}

func (c *Client) UpdateIndex() error {
	res, err := c.conn.Indices.Delete([]string{c.config.ElasticIndex})
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not delete index: %v", res)
	}
	res, err = c.conn.Indices.Create(c.config.ElasticIndex)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not create index: %v", res)
	}
	return nil
}

func (c *Client) BulkWrite(buf bytes.Buffer) error {
	res, err := c.conn.Bulk(bytes.NewReader(buf.Bytes()), c.conn.Bulk.WithIndex(c.config.ElasticIndex))
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not perform bulk insert: %v", res)
	}
	return nil
}
