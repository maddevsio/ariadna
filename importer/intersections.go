package importer

import (
	"fmt"

	"github.com/dhconnelly/rtreego"
	ggeo "github.com/paulmach/go.geo"
)

var Index *rtreego.Rtree

func BuildIndex(Roads []JsonWay) {
	Index = rtreego.NewTree(2, 25, 50)
	var err error
	for _, way := range Roads {
		maxlat, minlat, maxlon, minlon := way.GetXY()

		p := rtreego.Point{minlon, minlat}
		if way.Tags["name"] == "Полярная" {
			fmt.Println(minlat, minlon)
			fmt.Println(maxlat - minlat)
			fmt.Println(maxlon - minlon)
		}
		way.Rect, err = rtreego.NewRect(p, []float64{(maxlon - minlon), (maxlat - minlat)})
		if err != nil {
			continue
		}
		Index.Insert(&way)
	}
	Logger.Info("Index built")
	fmt.Println(Index.Size())
}

func SearchIntersections(Roads []JsonWay) []JsonNode {
	var Intersections []JsonNode
	for _, way := range Roads {
		maxlat, minlat, maxlon, minlon := way.GetXY()
		path := ggeo.NewPath()
		for _, point := range way.Nodes {
			path.Push(ggeo.NewPoint(point.Lng(), point.Lat()))
		}
		p := rtreego.Point{minlon, minlat}
		way.Rect, _ = rtreego.NewRect(p, []float64{maxlon - minlon, maxlat - minlat})
		results := Index.SearchIntersect(way.Rect)

		for _, result := range results {
			way2 := result.(*JsonWay)
			fmt.Printf("%s intersectss %s \n", way.Tags["name"], way2.Tags["name"])
			line := ggeo.NewPath()
			if way.ID == way2.ID {
				continue
			}
			for _, point := range way2.Nodes {
				line.Push(ggeo.NewPoint(point.Lng(), point.Lat()))
			}
			if path.Intersects(line) {
				points, _ := path.Intersection(line)
				for i, _ := range points {
					tags := make(map[string]string)
					tags["name"] = way.Tags["name"] + " " + way2.Tags["name"]
					InterSection := JsonNode{way.ID + way2.ID, "node", points[i].Lat(), points[i].Lng(), tags, true}
					Intersections = append(Intersections, InterSection)
				}
			}
		}
	}
	return Intersections

}
