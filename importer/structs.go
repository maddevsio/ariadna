package importer

import (
	"math"

	"github.com/dhconnelly/rtreego"
	"github.com/kellydunn/golang-geo"
)

type JsonWay struct {
	ID       int64             `json:"id"`
	Type     string            `json:"type"`
	Tags     map[string]string `json:"tags"`
	Centroid map[string]string `json:"centroid"`
	Nodes    []*geo.Point      `json:"nodes"`
	Rect     *rtreego.Rect     `json:"-"`
}

func (way *JsonWay) Bounds() *rtreego.Rect {
	return way.Rect
}

func (way *JsonWay) GetXY() (x, y, z, j float64) {
	var maxlat, minlat, maxlon, minlon float64
	minlat = float64(99999999999)
	minlon = float64(99999999999)
	for _, point := range way.Nodes {
		x, y := getXY(point.Lat(), point.Lng())
		maxlat = math.Max(maxlat, x)
		minlat = math.Min(minlat, x)
		maxlon = math.Max(maxlon, y)
		minlon = math.Min(minlon, y)
	}
	return maxlat, minlat, maxlon, minlon
}
func getXY(lat, lng float64) (float64, float64) {
	LAT := (lat * math.Pi) / 180
	LON := (lng * math.Pi) / 180
	X := 6371 * math.Sin(LAT) * math.Sin(LON)
	Y := 6371 * math.Cos(LAT)
	return X, Y
}

type Tags struct {
	housenumber string
	street      string
}

type JsonNode struct {
	ID           int64             `json:"id"`
	Type         string            `json:"type"`
	Lat          float64           `json:"lat"`
	Lon          float64           `json:"lon"`
	Tags         map[string]string `json:"tags"`
	Intersection bool              `json:"-"`
}

type JsonRelation struct {
	ID       int64               `json:"id"`
	Type     string              `json:"type"`
	Tags     map[string]string   `json:"tags"`
	Centroid map[string]string   `json:"centroid"`
	Nodes    []map[string]string `json:"nodes"`
}

type JsonEsIndex struct {
	Country           string             `json:"country"`
	City              string             `json:"city"`
	Village           string             `json:"village"`
	Town              string             `json:"town"`
	District          string             `json:"district"`
	Street            string             `json:"street"`
	HouseNumber       string             `json:"housenumber"`
	Name              string             `json:"name"`
	OldName           string             `json:"old_name"`
	HouseName         string             `json:"housename"`
	PostCode          string             `json:"postcode"`
	LocalName         string             `json:"local_name"`
	AlternativeName   string             `json:"alternative_name"`
	InternationalName string             `json:"international"`
	NationalName      string             `json:"national"`
	OfficialName      string             `json:"official"`
	RegionalName      string             `json:"regional"`
	ShortName         string             `json:"short_name"`
	SortingName       string             `json:"sorting"`
	TranslatedName    string             `json:"translated"`
	Custom            bool               `json:"custom"`
	Intersection      bool               `json:"intersection"`
	Centroid          map[string]float64 `json:"centroid"`
	Geom              interface{}        `json:"geom"`
}

type PGNode struct {
	ID      int64
	Name    string
	OldName string
	Lng     float64
	Lat     float64
}

type Translate struct {
	Original  string
	Translate string
}
