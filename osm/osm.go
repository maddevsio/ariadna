package osm

import (
	"strconv"
	"sync"

	"github.com/missinglink/gosmparse"
)

// Handler - Load all elements in to memory
type Handler struct {
	mu            *sync.Mutex
	InvertedIndex map[string][]string
	Nodes         map[int64]gosmparse.Node
	Ways          map[int64]gosmparse.Way
	WayNames      map[string]string
	Areas         map[int64]gosmparse.Relation
	Relations     map[int64]gosmparse.Relation
	highWayTags   map[string]bool
	placeTags     map[string]bool
}

// New creates new instance of Handler
func New() *Handler {
	h := &Handler{
		mu:            &sync.Mutex{},
		Nodes:         make(map[int64]gosmparse.Node),
		Ways:          make(map[int64]gosmparse.Way),
		WayNames:      make(map[string]string),
		Relations:     make(map[int64]gosmparse.Relation),
		Areas:         make(map[int64]gosmparse.Relation),
		InvertedIndex: make(map[string][]string),
	}
	h.highWayTags = map[string]bool{
		"motorway":    false,
		"trunk":       false,
		"primary":     false,
		"secondary":   false,
		"residential": false,
		"service":     false,
		"tertiary":    false,
		"road":        false,
	}
	h.placeTags = map[string]bool{
		"neighbourhood": false,
		"town":          false,
		"suburb":        false,
		"city":          false,
		"village":       false,
	}
	return h
}

// ReadNode - called once per node
func (h *Handler) ReadNode(item gosmparse.Node) {
	h.mu.Lock()
	if item.Tags["addr:street"] != "" && item.Tags["addr:housenumber"] != "" {

		h.Nodes[item.ID] = item
	}
	if item.Tags["amenity"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["building"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["addr:housenumber"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["shop"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["office"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["public_transport"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["cuisine"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["railway"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["sport"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["natural"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["tourism"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["leisure"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["historic"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["man_made"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["landuse"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["waterway"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["aerialway"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["aeroway"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["craft"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	if item.Tags["military"] != "" && item.Tags["name"] != "" {
		h.Nodes[item.ID] = item
	}
	h.mu.Unlock()
}

// ReadWay - called once per way
func (h *Handler) ReadWay(item gosmparse.Way) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if item.Tags["addr:street"] != "" && item.Tags["addr:housenumber"] != "" {

		h.Ways[item.ID] = item
	}
	if item.Tags["amenity"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["building"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["addr:housenumber"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["shop"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["office"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["public_transport"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["cuisine"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["railway"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["sport"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["natural"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["tourism"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["leisure"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["historic"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["man_made"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["landuse"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["waterway"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["aerialway"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["aeroway"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["craft"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if item.Tags["military"] != "" && item.Tags["name"] != "" {
		h.Ways[item.ID] = item
	}
	if _, ok := h.highWayTags[item.Tags["highway"]]; !ok {
		return
	}
	if item.Tags["addr:street"] != "" && item.Tags["addr:housenumber"] != "" {

		h.Ways[item.ID] = item
	}

	// convert int64 to string
	var wayIDString = strconv.FormatInt(item.ID, 10)

	// get the best name from the tags
	if val, ok := item.Tags["addr:street"]; ok {
		h.WayNames[wayIDString] = val
	} else if val, ok := item.Tags["name"]; ok {
		h.WayNames[wayIDString] = val
	} else {
		return
	} // store the way ids in an array with the nodeid as key
	for _, nodeid := range item.NodeIDs {
		var nodeIDString = strconv.FormatInt(nodeid, 10)
		h.InvertedIndex[nodeIDString] = append(h.InvertedIndex[nodeIDString], wayIDString)
	}
}

// ReadRelation - called once per relation
func (h *Handler) ReadRelation(item gosmparse.Relation) {
	h.mu.Lock()
	h.Relations[item.ID] = item
	if _, ok := h.placeTags[item.Tags["place"]]; ok {
		h.Areas[item.ID] = item
	}
	h.mu.Unlock()
}
