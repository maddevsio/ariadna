package osm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/maddevsio/ariadna/model"
	geojson "github.com/paulmach/go.geojson"
)

func (i *Importer) crossRoadsToElastic() error {

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
