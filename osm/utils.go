package osm

import (
	"encoding/json"
	"strings"

	geo "github.com/kellydunn/golang-geo"
	"github.com/maddevsio/ariadna/model"
	"github.com/missinglink/gosmparse"
	geojson "github.com/paulmach/go.geojson"
)

func (i *Importer) wayToJSON(way gosmparse.Way) ([]byte, error) {
	var coords [][]float64
	for _, nodeID := range way.NodeIDs {
		node := i.handler.Nodes[nodeID]
		coords = append(coords, []float64{node.Lon, node.Lat})
	}
	geom := geojson.NewLineStringGeometry(coords)
	return i.marshalJSON(way.Tags, geom)
}

func (i *Importer) nodeToJSON(node gosmparse.Node) ([]byte, error) {
	geom := geojson.NewPointGeometry([]float64{node.Lon, node.Lat})
	return i.marshalJSON(node.Tags, geom)
}

func (i *Importer) marshalJSON(tags map[string]string, geom *geojson.Geometry) ([]byte, error) {
	var street = tags["addr:street"]
	var name = tags["name"]
	var houseNumber = tags["addr:housenumber"]
	shape, err := geom.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var address = model.Address{
		Street:      street,
		Name:        name,
		Shape:       shape,
		HouseNumber: houseNumber,
	}
	if address.Street != "" {
		if strings.Contains(address.Street, "улица") {
			address.Prefix = "улица"
			address.Street = strings.TrimSpace(strings.Replace(address.Street, "улица", "", -1))
		}
		if strings.Contains(address.Street, "проспект") {
			address.Prefix = "проспект"
			address.Street = strings.TrimSpace(strings.Replace(address.Street, "проспект", "", -1))
		}
		if strings.Contains(address.Street, "бульвар") {
			address.Prefix = "бульвар"
			address.Street = strings.TrimSpace(strings.Replace(address.Street, "бульвар", "", -1))
		}
		if strings.Contains(address.Street, "переулок") {
			address.Prefix = "переулок"
			address.Street = strings.TrimSpace(strings.Replace(address.Street, "переулок", "", -1))
		}
	}
	for name, country := range i.countries {
		var lat, lon float64
		switch geom.Type {
		case geojson.GeometryLineString:
			lon = geom.LineString[0][0]
			lat = geom.LineString[0][1]
		case geojson.GeometryPoint:
			lon = geom.Point[0]
			lat = geom.Point[1]
		default:
			continue
		}
		if country.Contains(geo.NewPoint(lat, lon)) {
			address.Country = name
		}
	}
	for key, area := range i.areas {
		info := strings.Split(key, "+")
		name := info[0]
		place := info[1]
		var lat, lon float64
		switch geom.Type {
		case geojson.GeometryLineString:
			lon = geom.LineString[0][0]
			lat = geom.LineString[0][1]
		case geojson.GeometryPoint:
			lon = geom.Point[0]
			lat = geom.Point[1]
		default:
			continue
		}
		if area.Contains(geo.NewPoint(lat, lon)) {
			switch place {
			case "city":
				address.City = name
			case "town":
				address.Town = name
			case "suburb":
				address.District = name
			case "village":
				address.Village = name
			case "neighbourhood":
				address.District = name
			}
		}
	}

	return json.Marshal(address)
}
