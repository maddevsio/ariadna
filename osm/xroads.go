package osm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	geo "github.com/kellydunn/golang-geo"
	"github.com/maddevsio/ariadna/model"
)

func (i *Importer) crossRoadsToElastic() error {
	i.logger.Info("started to search crossroads")
	buf, err := i.searchCrossRoads()
	if err != nil {
		return err
	}
	i.logger.Info("crossroads found")
	return i.e.BulkWrite(buf)
}

func (i *Importer) searchCrossRoads() (bytes.Buffer, error) {
	var buf bytes.Buffer
	replacer := strings.NewReplacer(
		"улица", "",
		"переулок", "",
		"бульвар", "",
		"проспект", "",
	)
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
				address := model.Address{
					Country:      "KG",
					Name:         replacer.Replace(strings.Join(uniqueNames, " ")),
					Location:     model.Location{Lat: node.Lat, Lon: node.Lon},
					Intersection: true,
				}
				for countryID := range i.countries {
					country := i.countries[countryID]
					lat := address.Location.Lat
					lon := address.Location.Lon
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

				data, err := json.Marshal(address)
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
