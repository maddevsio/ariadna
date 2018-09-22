package importer

import (
	"fmt"
	"github.com/kellydunn/golang-geo"
	"github.com/qedus/osmpbf"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"strconv"
)

func Run(d *osmpbf.Decoder, db *leveldb.DB, tags map[string][]string) ([]JsonWay, []JsonNode) {
	var Ways []JsonWay
	var Nodes []JsonNode
	batch := new(leveldb.Batch)

	var nc, wc, rc uint64
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			logger.Fatal(err.Error())
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:
				nc++
				cacheQueue(batch, v)
				if batch.Len() > 50000 {
					cacheFlush(db, batch)
				}
				if !hasTags(v.Tags) {
					break
				}
				v.Tags = trimTags(v.Tags)

				if containsValidTags(v.Tags, tags) {
					node := onNode(v)
					Nodes = append(Nodes, node)
				}

			case *osmpbf.Way:
				if batch.Len() > 1 {
					cacheFlush(db, batch)
				}
				wc++

				if !hasTags(v.Tags) {
					break
				}

				v.Tags = trimTags(v.Tags)
				if containsValidTags(v.Tags, tags) {
					latlons, err := cacheLookup(db, v)
					if err != nil {
						break
					}
					var centroid = computeCentroid(latlons)
					way := onWay(v, latlons, centroid)
					Ways = append(Ways, way)
				}

			case *osmpbf.Relation:
				if !hasTags(v.Tags) {
					break
				}
				v.Tags = trimTags(v.Tags)
				rc++

			default:

				logger.Fatal(fmt.Sprintf("unknown type %T\n", v))

			}
		}
	}
	return Ways, Nodes
}

func onNode(node *osmpbf.Node) JsonNode {
	marshall := JsonNode{node.ID, "node", node.Lat, node.Lon, node.Tags, false}
	return marshall
}

func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string) JsonWay {
	var points []*geo.Point
	for _, latlon := range latlons {
		var lat, _ = strconv.ParseFloat(latlon["lat"], 64)
		var lng, _ = strconv.ParseFloat(latlon["lon"], 64)
		points = append(points, geo.NewPoint(lat, lng))
	}
	marshall := JsonWay{
		ID:       way.ID,
		Type:     "way",
		Tags:     way.Tags,
		Centroid: centroid,
		Nodes:    points}
	return marshall

}
