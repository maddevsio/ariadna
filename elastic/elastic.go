package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/maddevsio/ariadna/config"
)

type Client struct {
	conn         *es.Client
	config       *config.Ariadna
	createdIndex string
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
	indexName := fmt.Sprintf("%s-%d", c.config.ElasticIndex, time.Now().Unix())
	res, err := c.conn.Indices.Create(indexName)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not create index: %v", res)
	}
	res, err = c.conn.Indices.PutAlias([]string{indexName}, c.config.ElasticIndex)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not create alias: %v", res)
	}
	c.createdIndex = indexName
	return nil
}
func (c *Client) DeleteIndices() error {
	var indicesToDelete []string
	r := esapi.IndicesGetAliasRequest{Name: []string{c.config.ElasticIndex}}
	res, err := r.Do(context.TODO(), c.conn.Transport)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not create alias: %v", res)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var schema map[string]interface{}
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return err
	}
	for key := range schema {
		if key != c.createdIndex && strings.Contains(key, c.config.ElasticIndex) {
			indicesToDelete = append(indicesToDelete, key)
		}
	}
	res, err = c.conn.Indices.Delete(indicesToDelete)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("could not delete indices: %v", res)
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
