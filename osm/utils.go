package osm

import (
	"encoding/json"

	"github.com/maddevsio/ariadna/model"
	"github.com/missinglink/gosmparse"
	geojson "github.com/paulmach/go.geojson"
)

func (i *Importer) wayToJSON(way gosmparse.Way) ([]byte, error) {
	var coords [][]float64
	for nodeID := range way.NodeIDs {
		node := i.handler.Nodes[int64(nodeID)]
		coords = append(coords, []float64{node.Lon, node.Lat})
	}
	geom := geojson.NewLineStringGeometry(coords)
	shape, err := geom.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return i.marshalJSON(way.Tags, shape)
}

func (i *Importer) nodeToJSON(node gosmparse.Node) ([]byte, error) {
	geom := geojson.NewPointGeometry([]float64{node.Lon, node.Lat})
	shape, err := geom.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return i.marshalJSON(node.Tags, shape)
}

func (i *Importer) marshalJSON(tags map[string]string, shape []byte) ([]byte, error) {
	street := tags["addr:street"]
	name := tags["name"]
	houseNumber := tags["addr:housenumber"]
	return json.Marshal(model.Address{
		Country:     "KG",
		Street:      street,
		Name:        name,
		Shape:       shape,
		HouseNumber: houseNumber,
	})
}
