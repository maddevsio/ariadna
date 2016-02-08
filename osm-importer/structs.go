package importer

import "github.com/kellydunn/golang-geo"

type JsonWay struct {
	ID       int64             `json:"id"`
	Type     string            `json:"type"`
	Tags     map[string]string `json:"tags"`
	Centroid map[string]string `json:"centroid"`
	Nodes    []*geo.Point      `json:"nodes"`
}

type Tags struct {
	housenumber string
	street      string
}

type JsonNode struct {
	ID   int64             `json:"id"`
	Type string            `json:"type"`
	Lat  float64           `json:"lat"`
	Lon  float64           `json:"lon"`
	Tags map[string]string `json:"tags"`
}

type JsonRelation struct {
	ID       int64               `json:"id"`
	Type     string              `json:"type"`
	Tags     map[string]string   `json:"tags"`
	Centroid map[string]string   `json:"centroid"`
	Nodes    []map[string]string `json:"nodes"`
}

type JsonEsIndex struct {
	Country     string             `json:"country"`
	City        string             `json:"city"`
	Village     string             `json:"village"`
	Town        string             `json:"town"`
	District    string             `json:"district"`
	Street      string             `json:"street"`
	HouseNumber string             `json:"housenumber"`
	Name        string             `json:"name"`
	OldName     string             `json:"old_name"`
	Centroid    map[string]float64 `json:"centroid"`
	Geom        interface{}        `json:"geom"`
}

type PGNode struct {
	ID      int64
	Name    string
	OldName string
	Lng     float64
	Lat     float64
}

type Settings struct {
	PbfPath    string
	BatchSize  int
	ConfigPath string
	IndexPath string
}

type Translate struct {
	Original  string
	Translate string
}

type Config struct {
	IndexName    string `json:"index_name"`
	PGConnString string `json:"pg_conn_string"`
	IndexType    string `json:"index_type"`
}

var C Config
