package handler

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
	FilteredNodes map[int64]gosmparse.Node
	Ways          map[int64]gosmparse.Way
	FullWays      map[int64]gosmparse.Way

	WayNames     map[string]string
	Areas        map[int64]gosmparse.Relation
	Districts    map[int64]gosmparse.Way
	Countries    map[int64]gosmparse.Relation
	highWayTags  map[string]bool
	areaTags     map[string]bool
	districtTags map[string]bool
	addressTags  map[string]string
}

// New creates new instance of Handler
func New() *Handler {
	h := &Handler{
		mu:            &sync.Mutex{},
		Nodes:         make(map[int64]gosmparse.Node),
		FilteredNodes: make(map[int64]gosmparse.Node),
		Ways:          make(map[int64]gosmparse.Way),
		FullWays:      make(map[int64]gosmparse.Way),
		WayNames:      make(map[string]string),
		Areas:         make(map[int64]gosmparse.Relation),
		Districts:     make(map[int64]gosmparse.Way),
		Countries:     make(map[int64]gosmparse.Relation),
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
	h.areaTags = map[string]bool{
		"town":    false,
		"city":    false,
		"village": false,
		"hamlet":  false,
	}
	h.districtTags = map[string]bool{
		"neighbourhood": false,
		"suburb":        false,
	}
	h.addressTags = map[string]string{
		"addr:street":      "addr:housenumber",
		"amenity":          "name",
		"building":         "name",
		"addr:housenumber": "",
		"shop":             "name",
		"office":           "name",
		"public_transport": "name",
		"cuisine":          "name",
		"railway":          "name",
		"sport":            "name",
		"natural":          "name",
		"tourism":          "name",
		"leisure":          "name",
		"historic":         "name",
		"man_made":         "name",
		"landuse":          "name",
		"waterway":         "name",
		"aerialway":        "name",
		"aeroway":          "name",
		"craft":            "name",
		"military":         "name",
	}

	return h
}

// ReadNode - called once per node
func (h *Handler) ReadNode(item gosmparse.Node) {
	h.mu.Lock()
	h.Nodes[item.ID] = item
	for k, v := range h.addressTags {
		if item.Tags[k] != "" {
			if v == "" {
				h.FilteredNodes[item.ID] = item
			}
			if v != "" && item.Tags[v] != "" {
				h.FilteredNodes[item.ID] = item
			}
		}
	}
	h.mu.Unlock()
}

// ReadWay - called once per way
func (h *Handler) ReadWay(item gosmparse.Way) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.districtTags[item.Tags["place"]]; ok {
		h.Districts[item.ID] = item
	}
	h.FullWays[item.ID] = item
	for k, v := range h.addressTags {
		if item.Tags[k] != "" {
			if v == "" {
				h.Ways[item.ID] = item
			}
			if v != "" && item.Tags[v] != "" {
				h.Ways[item.ID] = item
			}
		}
	}

	if _, ok := h.highWayTags[item.Tags["highway"]]; !ok {
		return
	}
	if item.Tags["addr:street"] != "" && item.Tags["addr:housenumber"] != "" {

		h.Ways[item.ID] = item
	}

	// convert int64 to string
	var wayIDString = strconv.FormatInt(item.ID, 10)

	if val, ok := item.Tags["addr:street"]; ok {
		h.WayNames[wayIDString] = val
	} else if val, ok := item.Tags["name"]; ok {
		h.WayNames[wayIDString] = val
	} else {
		return
	}
	for _, nodeid := range item.NodeIDs {
		var nodeIDString = strconv.FormatInt(nodeid, 10)
		h.InvertedIndex[nodeIDString] = append(h.InvertedIndex[nodeIDString], wayIDString)
	}
}

// ReadRelation - called once per relation
func (h *Handler) ReadRelation(item gosmparse.Relation) {
	h.mu.Lock()
	if item.Tags["admin_level"] == "2" {
		h.Countries[item.ID] = item
	}
	if _, ok := h.areaTags[item.Tags["place"]]; ok {
		h.Areas[item.ID] = item
	}
	h.mu.Unlock()
}
