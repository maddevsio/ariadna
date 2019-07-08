package osm

import (
	"fmt"

	geo "github.com/kellydunn/golang-geo"
	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/elastic"
	"github.com/maddevsio/ariadna/osm/handler"
	"github.com/maddevsio/ariadna/osm/parser"
	"github.com/missinglink/gosmparse"
	"golang.org/x/sync/errgroup"
)

// Importer struct represents needed values to import data to elasticsearch
type Importer struct {
	handler   *handler.Handler
	parser    *parser.Parser
	config    *config.Ariadna
	e         *elastic.Client
	eg        errgroup.Group
	countries map[string]*geo.Polygon
	areas     map[string]*geo.Polygon
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
	i.countries = make(map[string]*geo.Polygon)
	i.areas = make(map[string]*geo.Polygon)
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
	i.areasToPolygons()
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
func (i *Importer) areasToPolygons() {
	for _, area := range i.handler.Areas {
		i.areas[fmt.Sprintf("%s+%s", area.Tags["name"], area.Tags["place"])] = i.relationToPolygon(area)
	}
	for _, country := range i.handler.Countries {
		i.countries[country.Tags["name"]] = i.relationToPolygon(country)
	}
}
func (i *Importer) relationToPolygon(area gosmparse.Relation) *geo.Polygon {
	var points []*geo.Point
	for _, member := range area.Members {
		node, ok := i.handler.Nodes[member.ID]
		if ok {
			points = append(points, geo.NewPoint(node.Lat, node.Lon))
		}
		if !ok {
			way := i.handler.FullWays[member.ID]
			for _, nodeID := range way.NodeIDs {
				node := i.handler.Nodes[nodeID]
				points = append(points, geo.NewPoint(node.Lat, node.Lon))
			}
		}

	}
	return geo.NewPolygon(points)
}
