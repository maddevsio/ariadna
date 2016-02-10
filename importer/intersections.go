package importer

import (
	ggeo "github.com/paulmach/go.geo"

	"fmt"
)

func GetRoadIntersections(Roads []JsonWay) []JsonNode {
	// TODO: optimize it
	// Use https://github.com/pierrre/geohash
	// Or https://en.wikipedia.org/wiki/K-d_tree
	var Intersections []JsonNode
	for _, way := range Roads {
		path := ggeo.NewPath()

		for _, point := range way.Nodes {
			path.Push(ggeo.NewPoint(point.Lng(), point.Lat()))
		}
		for _, way2 := range Roads {
			line := ggeo.NewPath()
			if way.ID == way2.ID {
				continue
			}
			for _, point := range way2.Nodes {
				line.Push(ggeo.NewPoint(point.Lng(), point.Lat()))
			}
			if path.Intersects(line) {
				fmt.Println("Intersects")
				points, segments := path.Intersection(line)
				for i, _ := range points {
					var FirstName string
					var SecondName string
					if way.Tags["name"] != "" {
						FirstName = way.Tags["name"]
					} else {
						FirstName = way.Tags["addr:street"]
					}
					if way2.Tags["name"] != "" {
						SecondName = way2.Tags["name"]
					} else {
						SecondName = way2.Tags["addr:street"]
					}
					tags := make(map[string]string)
					tags["name"] = FirstName + " " + SecondName
					InterSection := JsonNode{way.ID + way2.ID, "node", points[i].Lng(), points[i].Lat(), tags}
					Intersections = append(Intersections, InterSection)
					Logger.Info(fmt.Sprintf("Intersection %d at %v with path segment %d on %s and %s", i, points[i], segments[i][0], FirstName, SecondName))
				}
			}
		}
		fmt.Println("Processed" + way.Tags["name"])
	}
	return Intersections
}
