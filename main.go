package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/elastic"
	"github.com/maddevsio/ariadna/model"
	"github.com/maddevsio/ariadna/osm"
	"github.com/maddevsio/ariadna/parser"
	geojson "github.com/paulmach/go.geojson"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	p, err := parser.NewParser(c.OSMFilename)
	if err != nil {
		log.Fatal(err)
	}
	h := osm.New()
	err = p.Parse(h)
	if err != nil {
		log.Fatal(err)
	}
	e, err := elastic.New(c)
	if err != nil {
		log.Fatal(err)
	}
	err = e.UpdateIndex()
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	for nodeid, wayids := range h.InvertedIndex {

		// uniqify wayids
		uniqueWayIds := uniqString(wayids)
		if len(uniqueWayIds) > 1 {

			// generate way names
			var names []string
			sort.Strings(uniqueWayIds)
			for _, wayid := range uniqueWayIds {
				names = append(names, h.WayNames[wayid])
			}

			// only unique ones
			var uniqueNames = uniqString(names)
			sort.Strings(uniqueNames)
			if len(uniqueNames) > 1 {
				id, err := strconv.Atoi(nodeid)
				if err != nil {
					log.Fatal(err)
				}
				node := h.Nodes[int64(id)]
				// Point coordinates are in x, y order
				// (easting, northing for projected coordinates, longitude, latitude for geographic coordinates)
				geom := geojson.NewPointGeometry([]float64{node.Lon, node.Lat}) // https://geojson.org/geojson-spec.html#id9
				raw, err := geom.MarshalJSON()
				if err != nil {
					log.Fatal(err)
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
					log.Fatal(err)
				}
				meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, nodeid, "\n"))
				data = append(data, "\n"...)
				buf.Grow(len(meta) + len(data))
				buf.Write(meta)
				buf.Write(data)
			}
		}
	}
	err = e.BulkWrite(buf)
	if err != nil {
		log.Fatal(err)
	}
} // convenience func to uniq a set
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
