package osm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/elastic"
	"github.com/maddevsio/ariadna/model"
	"github.com/maddevsio/ariadna/osm/handler"
	"github.com/maddevsio/ariadna/osm/parser"
	geojson "github.com/paulmach/go.geojson"
)

// Importer struct represents needed values to import data to elasticsearch
type Importer struct {
	handler *handler.Handler
	parser  *parser.Parser
	config  *config.Ariadna
	e       *elastic.Client
	wg      sync.WaitGroup
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
	i.wg.Add(1)
	go i.crossRoadsToElastic()
	return nil
}

// WaitStop is wrapper around waitgroup
func (i *Importer) WaitStop() {
	i.wg.Wait()
}
func (i *Importer) crossRoadsToElastic() error {
	defer i.wg.Done()
	buf, err := i.searchCrossRoads()
	if err != nil {
		return err
	}
	return i.e.BulkWrite(buf)
}
func (i *Importer) searchCrossRoads() (bytes.Buffer, error) {
	var buf bytes.Buffer
	for nodeid, wayids := range i.handler.InvertedIndex {
		uniqueWayIds := uniqString(wayids)
		if len(uniqueWayIds) > 1 {
			var names []string
			sort.Strings(uniqueWayIds)
			for _, wayid := range uniqueWayIds {
				names = append(names, i.handler.WayNames[wayid])
			}
			var uniqueNames = uniqString(names)
			sort.Strings(uniqueNames)
			if len(uniqueNames) > 1 {
				id, err := strconv.Atoi(nodeid)
				if err != nil {
					return buf, err
				}
				node := i.handler.Nodes[int64(id)]
				// Point coordinates are in x, y order
				// (easting, northing for projected coordinates, longitude, latitude for geographic coordinates)
				geom := geojson.NewPointGeometry([]float64{node.Lon, node.Lat}) // https://geojson.org/geojson-spec.html#id9
				raw, err := geom.MarshalJSON()
				if err != nil {
					return buf, err
				}
				data, err := json.Marshal(model.Address{
					Country:  "KG",
					City:     "",
					Village:  "",
					Town:     "",
					District: "",
					Street:   strings.Join(uniqueNames, " "),
					Shape:    raw,
				})
				if err != nil {
					return buf, err
				}
				meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, nodeid, "\n"))
				data = append(data, "\n"...)
				buf.Grow(len(meta) + len(data))
				buf.Write(meta)
				buf.Write(data)
			}
		}
	}
	return buf, nil
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
