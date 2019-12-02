package osm

import (
	"encoding/json"
	"strings"

	geo "github.com/kellydunn/golang-geo"
	"github.com/maddevsio/ariadna/model"
	"github.com/missinglink/gosmparse"
)

func (i *Importer) wayToJSON(way gosmparse.Way) ([]byte, error) {
	var coords [][]float64
	for _, nodeID := range way.NodeIDs {
		node := i.handler.Nodes[nodeID]
		coords = append(coords, []float64{node.Lon, node.Lat})
	}
	x := 0.0
	y := 0.0
	numPoints := float64(len(coords))
	for _, point := range coords {
		x += point[0]
		y += point[1]
	}
	return i.marshalJSON(way.Tags, model.Location{Lat: y / numPoints, Lon: x / numPoints})
}

func (i *Importer) nodeToJSON(node gosmparse.Node) ([]byte, error) {
	return i.marshalJSON(node.Tags, model.Location{Lat: node.Lat, Lon: node.Lon})
}

func (i *Importer) marshalJSON(tags map[string]string, location model.Location) ([]byte, error) {
	var street = tags["addr:street"]
	var name = tags["name"]
	var houseNumber = tags["addr:housenumber"]
	var address = model.Address{
		Street:      street,
		Name:        name,
		Location:    location,
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
		point := geo.NewPoint(address.Location.Lat, address.Location.Lon)
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
