// Useful tags for Bishkek
// addr:street+addr:housenumber - Get all known addresses
// place~city - Get all cities
// place~suburb - Get districts
// place~village - Get villages
// building,shop - get all buildings and shops
// highway - Get all roads

package main

import (
	"log"
	"fmt"
	"os"
	"bytes"
	"flag"
	"io"
	"math"
	"strconv"
	"strings"
	"runtime"
	"gopkg.in/olivere/elastic.v3"
	"github.com/kellydunn/golang-geo"
	gj "github.com/paulmach/go.geojson"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/qedus/osmpbf"
)


type Settings struct {
	PbfPath   string
	BatchSize int
}

func getSettings() Settings {

	// command line flags
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")

	flag.Parse()
	args := flag.Args();

	if len(args) < 1 {
		log.Fatal("invalid args, you must specify a PBF file")
	}

	return Settings{args[0], *batchSize }
}

func main() {

	// configuration
	config := getSettings()


	// open pbf file
	file := openFile(config.PbfPath)
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}
	_, err = client.CreateIndex("addresses").Do()
	if err != nil {
		// Handle error
		fmt.Println(err)
	}

	db := openLevelDB("db")
	defer db.Close()

	tags := buildTags("place~city")
	Cities := run(decoder, db, tags)
	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))

	tags = buildTags("place~village")
	Villages := run(decoder, db, tags)

	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))

	tags = buildTags("place~suburb")
	SubUrbs := run(decoder, db, tags)

	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))

	tags = buildTags("addr:street+addr:housenumber")
	Addresses := run(decoder, db, tags)

	for _, address := range Addresses {
		var cityName, villageName, suburbName string
		var lat, _ = strconv.ParseFloat(address.Centroid["lat"], 64)
		var lng, _ = strconv.ParseFloat(address.Centroid["lon"], 64)
		for _, city := range Cities {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(lat, lng)) {
				cityName = city.Tags["name"]
			}
		}
		for _, village := range Villages {
			polygon := geo.NewPolygon(village.Nodes)
			if polygon.Contains(geo.NewPoint(lat, lng)) {
				villageName = village.Tags["name"]
			}
		}
		for _, suburb := range SubUrbs {
			polygon := geo.NewPolygon(suburb.Nodes)
			if polygon.Contains(geo.NewPoint(lat, lng)) {
				suburbName = suburb.Tags["name"]
			}
		}
		p := gj.NewPointGeometry([]float64{lat, lng})
		var points [][][]float64
		for _, point := range address.Nodes {
			points = append(points, [][]float64{[]float64{point.Lat(), point.Lng()}})
		}

		pg := gj.NewPolygonFeature(points)
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, address.Tags["addr:street"], address.Tags["addr:housenumber"], address.Tags["name"], p, pg}
		row, err := client.Index().
		Index("addresses").
		Type("address").
		Id(strconv.FormatInt(address.ID, 10)).
		BodyJson(marshall).
		Do()

		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(row.Created, row.Id)
	}
	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))
	tags = buildTags("addr:street+addr:housenumber")
	AddrNodes := processNodes(decoder, db, tags)

	for _, address := range AddrNodes {
		var cityName, villageName, suburbName string
		for _, city := range Cities {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				cityName = city.Tags["name"]
			}
		}
		for _, village := range Villages {
			polygon := geo.NewPolygon(village.Nodes)
			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				villageName = village.Tags["name"]
			}
		}
		for _, suburb := range SubUrbs {
			polygon := geo.NewPolygon(suburb.Nodes)
			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				suburbName = suburb.Tags["name"]
			}
		}
		p := gj.NewPointGeometry([]float64{address.Lat, address.Lon})
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, address.Tags["addr:street"], address.Tags["addr:housenumber"], address.Tags["name"], p, nil}
		row, err := client.Index().
		Index("addresses").
		Type("address").
		Id(strconv.FormatInt(address.ID, 10)).
		BodyJson(marshall).
		Do()

		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(row.Created, row.Id)
	}

	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))
	tags = buildTags("building,shop")
	BNodes := processNodes(decoder, db, tags)

	for _, address := range BNodes {
		var cityName, villageName, suburbName string
		for _, city := range Cities {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				cityName = city.Tags["name"]
			}
		}
		for _, village := range Villages {
			polygon := geo.NewPolygon(village.Nodes)
			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				villageName = village.Tags["name"]
			}
		}
		for _, suburb := range SubUrbs {
			polygon := geo.NewPolygon(suburb.Nodes)
			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				suburbName = suburb.Tags["name"]
			}
		}
		p := gj.NewPointGeometry([]float64{address.Lat, address.Lon})
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, address.Tags["addr:street"], address.Tags["addr:housenumber"], address.Tags["name"], p, nil}
		row, err := client.Index().
		Index("addresses").
		Type("address").
		Id(strconv.FormatInt(address.ID, 10)).
		BodyJson(marshall).
		Do()

		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(row.Created, row.Id)
	}

	file = openFile(config.PbfPath)
	defer file.Close()

	decoder = osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(-1))

	tags = buildTags("building,shop")
	Buildings := run(decoder, db, tags)

	for _, address := range Buildings {
		var cityName, villageName, suburbName string
		var lat, _ = strconv.ParseFloat(address.Centroid["lat"], 64)
		var lng, _ = strconv.ParseFloat(address.Centroid["lon"], 64)
		for _, city := range Cities {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(lat, lng)) {
				cityName = city.Tags["name"]
			}
		}
		for _, village := range Villages {
			polygon := geo.NewPolygon(village.Nodes)
			if polygon.Contains(geo.NewPoint(lat, lng)) {
				villageName = village.Tags["name"]
			}
		}
		for _, suburb := range SubUrbs {
			polygon := geo.NewPolygon(suburb.Nodes)
			if polygon.Contains(geo.NewPoint(lat, lng)) {
				suburbName = suburb.Tags["name"]
			}
		}
		p := gj.NewPointGeometry([]float64{lat, lng})
		var points [][][]float64
		for _, point := range address.Nodes {
			points = append(points, [][]float64{[]float64{point.Lat(), point.Lng()}})
		}

		pg := gj.NewPolygonFeature(points)
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, address.Tags["addr:street"], address.Tags["addr:housenumber"], address.Tags["name"], p, pg}
		row, err := client.Index().
		Index("addresses").
		Type("address").
		Id(strconv.FormatInt(address.ID, 10)).
		BodyJson(marshall).
		Do()

		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(row.Created, row.Id)
	}


}


func buildTags(tagList string) map[string][]string {
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}
	return conditions
}

func processNodes(d *osmpbf.Decoder, db *leveldb.DB, tags map[string][]string) []JsonNode {
	var Nodes []JsonNode
	batch := new(leveldb.Batch)

	var nc uint64
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:
				nc++
				cacheQueue(batch, v)
				if batch.Len() > 50000 {
					cacheFlush(db, batch)
				}
				if !hasTags(v.Tags) { break }
				v.Tags = trimTags(v.Tags)

				if containsValidTags(v.Tags, tags) {
					node := onNode(v)
					Nodes = append(Nodes, node)
				}


			case *osmpbf.Way:
			case *osmpbf.Relation:
				continue
			default:

				log.Fatalf("unknown type %T\n", v)

			}
		}
	}
	return Nodes
}

func run(d *osmpbf.Decoder, db *leveldb.DB, tags map[string][]string) []JsonWay {
	var Ways []JsonWay
	batch := new(leveldb.Batch)

	var nc, wc, rc uint64
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {

			case *osmpbf.Node:
				nc++
				cacheQueue(batch, v)
				if batch.Len() > 50000 {
					cacheFlush(db, batch)
				}
				if !hasTags(v.Tags) { break }
				v.Tags = trimTags(v.Tags)

				if containsValidTags(v.Tags, tags) {
					//					onNode(v)
				}

			case *osmpbf.Way:
				if batch.Len() > 1 {
					cacheFlush(db, batch)
				}
				wc++

				if !hasTags(v.Tags) { break }

				v.Tags = trimTags(v.Tags)
				if containsValidTags(v.Tags, tags) {
					latlons, err := cacheLookup(db, v)
					if err != nil { break }
					var centroid = computeCentroid(latlons);
					way := onWay(v, latlons, centroid)
					Ways = append(Ways, way)
				}

			case *osmpbf.Relation:
				if !hasTags(v.Tags) { break }
				v.Tags = trimTags(v.Tags)
				rc++

			default:

				log.Fatalf("unknown type %T\n", v)

			}
		}
	}
	return Ways
}

type JsonNode struct {
	ID   int64               `json:"id"`
	Type string              `json:"type"`
	Lat  float64             `json:"lat"`
	Lon  float64             `json:"lon"`
	Tags map[string]string   `json:"tags"`
}

type JsonRelation struct {
	ID       int64               `json:"id"`
	Type     string              `json:"type"`
	Tags     map[string]string   `json:"tags"`
	Centroid map[string]string   `json:"centroid"`
	Nodes    []map[string]string `json:"nodes"`
}

type JsonEsIndex struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	Town        string `json:"town"`
	District    string `json:"district"`
	Street      string `json:"street"`
	HouseNumber string `json:"housenumber"`
	Name        string `json:"name"`
	Centroid    interface{} `json:"centroid"`
	Geom        interface{} `json:"geom"`
}

func onNode(node *osmpbf.Node) JsonNode {
	marshall := JsonNode{node.ID, "node", node.Lat, node.Lon, node.Tags}
	return marshall
}

type JsonWay struct {
	ID       int64               `json:"id"`
	Type     string              `json:"type"`
	Tags     map[string]string   `json:"tags"`
	Centroid map[string]string   `json:"centroid"`
	Nodes    [] *geo.Point             `json:"nodes"`
}

type Tags struct {
	housenumber string
	street      string
}

func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string) JsonWay {
	var points [] *geo.Point
	for _, latlon := range latlons {
		var lat, _ = strconv.ParseFloat(latlon["lat"], 64)
		var lng, _ = strconv.ParseFloat(latlon["lon"], 64)
		points = append(points, geo.NewPoint(lat, lng))
	}
	marshall := JsonWay{way.ID, "way", way.Tags, centroid, points, }
	return marshall

}
// queue a leveldb write in a batch
func cacheQueue(batch *leveldb.Batch, node *osmpbf.Node) {
	id, val := formatLevelDB(node)
	batch.Put([]byte(id), []byte(val))
}

// flush a leveldb batch to database and reset batch to 0
func cacheFlush(db *leveldb.DB, batch *leveldb.Batch) {
	err := db.Write(batch, nil)
	if err != nil {
		log.Fatal(err)
	}
	batch.Reset()
}

func cacheLookup(db *leveldb.DB, way *osmpbf.Way) ([]map[string]string, error) {

	var container []map[string]string

	for _, each := range way.NodeIDs {
		stringid := strconv.FormatInt(each, 10)

		data, err := db.Get([]byte(stringid), nil)
		if err != nil {
			log.Println("denormalize failed for way:", way.ID, "node not found:", stringid)
			return container, err
		}

		s := string(data)
		spl := strings.Split(s, ":");

		latlon := make(map[string]string)
		lat, lon := spl[0], spl[1]
		latlon["lat"] = lat
		latlon["lon"] = lon

		container = append(container, latlon)

	}

	return container, nil
}



func formatLevelDB(node *osmpbf.Node) (id string, val []byte) {

	stringid := strconv.FormatInt(node.ID, 10)

	var bufval bytes.Buffer
	bufval.WriteString(strconv.FormatFloat(node.Lat, 'f', 6, 64))
	bufval.WriteString(":")
	bufval.WriteString(strconv.FormatFloat(node.Lon, 'f', 6, 64))
	byteval := []byte(bufval.String())

	return stringid, byteval
}

func openFile(filename string) *os.File {
	// no file specified
	if len(filename) < 1 {
		log.Fatal("invalid file: you must specify a pbf path as arg[1]")
	}
	// try to open the file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func openLevelDB(path string) *leveldb.DB {
	// try to open the db
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// check tags contain features from a whitelist
func matchTagsAgainstCompulsoryTagList(tags map[string]string, tagList []string) bool {
	for _, name := range tagList {

		feature := strings.Split(name, "~")
		foundVal, foundKey := tags[feature[0]]

		// key check
		if !foundKey {
			return false
		}

		// value check
		if len(feature) > 1 {
			if foundVal != feature[1] {
				return false
			}
		}
	}

	return true
}

// check tags contain features from a groups of whitelists
func containsValidTags(tags map[string]string, group map[string][]string) bool {
	for _, list := range group {
		if matchTagsAgainstCompulsoryTagList(tags, list) {
			return true
		}
	}
	return false
}

// trim leading/trailing spaces from keys and values
func trimTags(tags map[string]string) map[string]string {
	trimmed := make(map[string]string)
	for k, v := range tags {
		trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v);
	}
	return trimmed
}

// check if a tag list is empty or not
func hasTags(tags map[string]string) bool {
	n := len(tags)
	if n == 0 {
		return false
	}
	return true
}

// compute the centroid of a way
func computeCentroid(latlons []map[string]string) map[string]string {
	var points []geo.Point
	for _, each := range latlons {
		var lon, _ = strconv.ParseFloat(each["lon"], 64);
		var lat, _ = strconv.ParseFloat(each["lat"], 64);
		point := geo.NewPoint(lat, lon)
		points = append(points, *point)
	}

	var compute = getCentroid(points);

	var centroid = make(map[string]string)
	centroid["lat"] = strconv.FormatFloat(compute.Lat(), 'f', 6, 64)
	centroid["lon"] = strconv.FormatFloat(compute.Lng(), 'f', 6, 64)

	return centroid
}

// compute the centroid of a polygon set
// using a spherical co-ordinate system
func getCentroid(ps []geo.Point) *geo.Point {

	X := 0.0
	Y := 0.0
	Z := 0.0

	var toRad = math.Pi / 180
	var fromRad = 180 / math.Pi

	for _, point := range ps {

		var lon = point.Lng() * toRad
		var lat = point.Lat() * toRad

		X += math.Cos(lat) * math.Cos(lon)
		Y += math.Cos(lat) * math.Sin(lon)
		Z += math.Sin(lat)
	}

	numPoints := float64(len(ps))
	X = X / numPoints
	Y = Y / numPoints
	Z = Z / numPoints

	var lon = math.Atan2(Y, X)
	var hyp = math.Sqrt(X * X + Y * Y)
	var lat = math.Atan2(Z, hyp)

	var centroid = geo.NewPoint(lat * fromRad, lon * fromRad)

	return centroid;
}