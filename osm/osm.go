package osm

import (
	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/elastic"
	"github.com/maddevsio/ariadna/osm/handler"
	"github.com/maddevsio/ariadna/osm/parser"
	"golang.org/x/sync/errgroup"
)

// Importer struct represents needed values to import data to elasticsearch
type Importer struct {
	handler *handler.Handler
	parser  *parser.Parser
	config  *config.Ariadna
	e       *elastic.Client
	eg      errgroup.Group
}

// NewImporter creates new instance of importer
func NewImporter(c *config.Ariadna) (*Importer, error) {
	i := &Importer{config: c}
	p, err := parser.NewParser(c.OSMFilename)
	if err != nil {
		return nil, err
	}
	i.parser = p
	e, err := elastic.New(c)
	if err != nil {
		return nil, err
	}
	i.e = e
	i.handler = handler.New()
	return i, nil
}
func (i *Importer) parse() error {
	return i.parser.Parse(i.handler)
}
func (i *Importer) updateIndices() error {
	return i.e.UpdateIndex()
}

// Start starts parsing
func (i *Importer) Start() error {
	if err := i.parse(); err != nil {
		return err
	}
	if err := i.updateIndices(); err != nil {
		return err
	}
	i.eg.Go(i.crossRoadsToElastic)
	i.eg.Go(i.nodesToElastic)
	i.eg.Go(i.waysToElastic)
	return nil
}

// WaitStop is wrapper around waitgroup
func (i *Importer) WaitStop() {
	i.eg.Wait()
}
func uniqString(list []string) []string {
	uniqueSet := make(map[string]bool)
	for _, x := range list {
		uniqueSet[x] = true
	}
	result := make([]string, 0, len(uniqueSet))
	for x := range uniqueSet {
		result = append(result, x)
	}
	return result
}
