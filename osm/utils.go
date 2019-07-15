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
	var address = model.Address{
		Street:      street,
		Name:        name,
		Shape:       geom,
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
	for countryID := range i.countries {
		country := i.countries[countryID]
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
		point := geo.NewPoint(lat, lon)
		if country.geom.Contains(point) {
			address.Country = country.name
		}
		for townID := range country.towns {
			town := country.towns[townID]
			if town.geom.Contains(point) {
				switch town.placeType {
				case "city":
					address.City = town.name
				case "town":
					address.Town = town.name
				case "hamlet":
					address.Village = town.name
				case "village":
					address.Village = town.name
				}
			}
			for districtID := range town.districts {
				district := town.districts[districtID]
				if district.geom.Contains(point) {
					address.District = district.name
				}
			}
		}

	}

	return json.Marshal(address)
}
