package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/maddevsio/ariadna/osm"
	"github.com/maddevsio/ariadna/parser"
)

func main() {
	p, err := parser.NewParser("kyrgyzstan-latest.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	h := osm.New()
	err = p.Parse(h)
	if err != nil {
		log.Fatal(err)
	}
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
				fmt.Printf("http://openstreetmap.org/node/%-15v %v\n", nodeid, strings.Join(uniqueNames, " / "))
			}
		}
	}
	for nodeID := range h.Nodes {
		fmt.Println(nodeID)
		//for _, member := range node.Members {
		//	spew.Dump(h.Ways[member.ID])
		//}
	}
	for nodeID := range h.Ways {
		fmt.Println(nodeID)
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
