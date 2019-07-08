package osm

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/maddevsio/ariadna/model"
	geojson "github.com/paulmach/go.geojson"
)

func (i *Importer) waysToElastic() error {
	buf, err := i.getWays()
	if err != nil {
		return err
	}
	return i.e.BulkWrite(buf)
}
func (i *Importer) getWays() (bytes.Buffer, error) {
	var buf bytes.Buffer
	for wayID, node := range i.handler.Ways {
		var coords [][]float64
		for nodeID := range node.NodeIDs {
			node := i.handler.Nodes[int64(nodeID)]
			coords = append(coords, []float64{node.Lon, node.Lat})
		}
		geom := geojson.NewLineStringGeometry(coords)
		shape, err := geom.MarshalJSON()
		if err != nil {
			return buf, err
		}
		data, err := json.Marshal(model.Address{
			Country:     "KG",
			Street:      node.Tags["addr:street"],
			Name:        node.Tags["name"],
			Shape:       shape,
			HouseNumber: node.Tags["addr:housenumber"],
		})
		if err != nil {
			return buf, err
		}
		meta := []byte(fmt.Sprintf(`{ "index": { "_id": "%d" } }%s`, wayID, "\n"))
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}
	return buf, nil
}
func (i *Importer) nodesToElastic() error {
	buf, err := i.getNodes()
	if err != nil {
		return err
	}
	return i.e.BulkWrite(buf)
}
func (i *Importer) getNodes() (bytes.Buffer, error) {
	var buf bytes.Buffer
	for nodeID, node := range i.handler.Nodes {
		geom := geojson.NewPointGeometry([]float64{node.Lon, node.Lat})
		shape, err := geom.MarshalJSON()
		if err != nil {
			return buf, err
		}
		data, err := json.Marshal(model.Address{
			Country:     "KG",
			Street:      node.Tags["addr:street"],
			Name:        node.Tags["name"],
			Shape:       shape,
			HouseNumber: node.Tags["addr:housenumber"],
		})
		if err != nil {
			return buf, err
		}
		meta := []byte(fmt.Sprintf(`{ "index": { "_id": "%d" } }%s`, nodeID, "\n"))
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}
	return buf, nil
}
