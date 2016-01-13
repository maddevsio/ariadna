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
	ggeo "github.com/paulmach/go.geo"
	_ "github.com/lib/pq"
	"database/sql"
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
func getDecoder(file *os.File) *osmpbf.Decoder {
	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}
	return decoder
}

var CitiesAndTowns, Roads []JsonWay

func main() {
	config := getSettings()

	db := openLevelDB("db")
	defer db.Close()

	file := openFile(config.PbfPath)
	defer file.Close()
	decoder := getDecoder(file)


	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}

	_, err = client.CreateIndex("addresses").Do()
	if err != nil {
		// Handle error
		fmt.Println(err)
	}

	fmt.Println("Searching cities, villages, towns and districts")
	tags := buildTags("place~city,place~village,place~suburb,place~town")
	CitiesAndTowns, _ = run(decoder, db, tags)

	fmt.Println("Cities, villages, towns and districts found")

	file = openFile(config.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	fmt.Println("Searching addresses")
	tags = buildTags("addr:street+addr:housenumber,amenity,shop")
	AddressWays, AddressNodes := run(decoder, db, tags)
	fmt.Println("Addresses found")
	JsonWaysToES(AddressWays, client)
	JsonNodesToEs(AddressNodes, client)

	file = openFile(config.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	tags = buildTags("highway")
	Roads, _ = run(decoder, db, tags)
	RoadsToPg()
	fmt.Println("Searching all roads intersecitons")
	Intersections := GetRoadIntersectionsFromPG()
	JsonNodesToEs(Intersections, client)
}

func RoadsToPg() {
	pg_db, err := sql.Open("postgres", "host=localhost user=geo password=geo dbname=geo sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pg_db.Close()

	_, err = pg_db.Query(`DROP TABLE IF EXISTS road;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pg_db.Query(`DROP TABLE IF EXISTS road_intersection;`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pg_db.Query(`CREATE TABLE road (
			id serial not null primary key,
			node_id bigint not null,
			name varchar(255) null,
			coords geometry
		);`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pg_db.Query(`create table road_intersection (
			id serial not null primary key,
			node_id bigint not null,
			name varchar(200) null,
			coords geometry
		);`)

	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created tables")
	fmt.Println("Populating...")
	const insQuery = `INSERT INTO road (node_id, name, coords) values($1, $2, ST_GeomFromText($3));`
	for _, road := range Roads {
		linestring := "LINESTRING("

		for _, point := range road.Nodes {
			linestring += fmt.Sprintf("%s %s,", strconv.FormatFloat(point.Lng(), 'f', 16, 64), strconv.FormatFloat(point.Lat(), 'f', 16, 64))
		}
		linestring = linestring[:len(linestring) - 1]
		linestring += ")"
		insert_query, err := pg_db.Prepare(insQuery)

		if err != nil {
			panic(err)
		}
		defer insert_query.Close()

		name := ""
		if road.Tags["name"] != "" {
			name = road.Tags["name"]
		} else {
			name = road.Tags["addr:name"]
		}

		_, err = insert_query.Exec(road.ID, cleanAddress(name), linestring)
		if err != nil {
			log.Fatal(err)
		}
	}
	searchQuery := `
		INSERT INTO road_intersection( coords, name, node_id)
			(SELECT DISTINCT (ST_DUMP(ST_INTERSECTION(a.coords, b.coords))).geom AS ix,
			concat(a.name, ' ', b.name) as InterName,
			a.node_id + b.node_id
			FROM road a
			INNER JOIN road b
			ON ST_INTERSECTS(a.coords,b.coords)
			WHERE geometrytype(st_intersection(a.coords,b.coords)) = 'POINT'
		);
	`
	fmt.Println("Started searching intersections ...")
	_, err = pg_db.Query(searchQuery)

	if err != nil {
		log.Fatal(err)
	}

}

type PGNode struct {
	ID   int64
	Name string
	Lng  float64
	Lat  float64

}

func GetRoadIntersectionsFromPG() []JsonNode {
	var Nodes []JsonNode
	pg_db, err := sql.Open("postgres", "host=localhost user=geo password=geo dbname=geo sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pg_db.Close()
	rows, err := pg_db.Query("SELECT node_id, name, st_x((st_dump(coords)).geom) as lng, st_y((st_dump(coords)).geom) as lat from road_intersection")

	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var node PGNode
		rows.Scan(&node.ID, &node.Name, &node.Lng, &node.Lat)
		tags := make(map[string]string)
		tags["name"] = node.Name
		jNode := JsonNode{node.ID, "node", node.Lat, node.Lng, tags}
		Nodes = append(Nodes, jNode)
	}
	return Nodes
}

func GetRoadIntersections(done chan bool) []JsonNode {
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
					log.Printf("Intersection %d at %v with path segment %d on %s and %s", i, points[i], segments[i][0], FirstName, SecondName)
				}
			}
		}
		fmt.Println("Processed" + way.Tags["name"])
	}
	done <- true
	return Intersections
}
func normalizeAddress(address string) string {
	if strings.Contains(address, "улица") {
		return fmt.Sprintf("улица %s", strings.Replace(address, "улица", "", -1))

	}
	if strings.Contains(address, "проспект") {
		return fmt.Sprintf("проспект %s", strings.Replace(address, "проспект", "", -1))

	}
	if strings.Contains(address, "переулок") {
		return fmt.Sprintf("переулок %s", strings.Replace(address, "переулок", "", -1))

	}
	return address
}
func cleanAddress(address string) string {
	if strings.Contains(address, "улица") {
		return strings.Replace(address, "улица", "", -1)

	}
	if strings.Contains(address, "проспект") {
		return strings.Replace(address, "проспект", "", -1)

	}
	if strings.Contains(address, "переулок") {
		return strings.Replace(address, "переулок", "", -1)

	}
	if strings.Contains(address, "микрорайон") {
		return strings.Replace(address, "микрорайон", "", -1)
	}

	return address
}
func JsonNodesToEs(Addresses []JsonNode, client *elastic.Client) {
	fmt.Println("Populating elastic search index")
	bulkClient := client.Bulk()
	for _, address := range Addresses {
		cityName, villageName, suburbName := "", "", ""
		for _, city := range CitiesAndTowns {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				switch city.Tags["place"] {
				case "city":
					cityName = city.Tags["name"]
				case "village":
					villageName = city.Tags["name"]
				case "suburb":
					suburbName = city.Tags["name"]
				}
			}
		}

		centroid := make(map[string]float64)
		centroid["lat"] = address.Lat
		centroid["lon"] = address.Lon
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), address.Tags["addr:housenumber"], cleanAddress(address.Tags["name"]), centroid, nil}
		index := elastic.NewBulkIndexRequest().Index("addresses").Type("address").Id(strconv.FormatInt(address.ID, 10)).Doc(marshall)
		bulkClient = bulkClient.Add(index)

	}
	_, err := bulkClient.Do()
	if err != nil {
		fmt.Println(err)
	}

}
func JsonWaysToES(Addresses []JsonWay, client *elastic.Client) {
	fmt.Println("Populating elastic search index")
	bulkClient := client.Bulk()
	for _, address := range Addresses {
		cityName, villageName, suburbName := "", "", ""
		var lat, _ = strconv.ParseFloat(address.Centroid["lat"], 64)
		var lng, _ = strconv.ParseFloat(address.Centroid["lon"], 64)
		for _, city := range CitiesAndTowns {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(lat, lng)) {
				switch city.Tags["place"] {
				case "city":
					cityName = city.Tags["name"]
				case "village":
					villageName = city.Tags["name"]
				case "suburb":
					suburbName = city.Tags["name"]
				}
			}
		}
		var points [][][]float64
		for _, point := range address.Nodes {
			points = append(points, [][]float64{[]float64{point.Lat(), point.Lng()}})
		}

		pg := gj.NewPolygonFeature(points)
		centroid := make(map[string]float64)
		centroid["lat"] = lat
		centroid["lon"] = lng
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), address.Tags["addr:housenumber"], cleanAddress(address.Tags["name"]), centroid, pg}
		index := elastic.NewBulkIndexRequest().Index("addresses").Type("address").Id(strconv.FormatInt(address.ID, 10)).Doc(marshall)

		bulkClient = bulkClient.Add(index)

	}
	_, err := bulkClient.Do()
	if err != nil {
		fmt.Println(err)
	}
}

func buildTags(tagList string) map[string][]string {
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}
	return conditions
}

func run(d *osmpbf.Decoder, db *leveldb.DB, tags map[string][]string) ([]JsonWay, []JsonNode) {
	var Ways []JsonWay
	var Nodes []JsonNode
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
					node := onNode(v)
					Nodes = append(Nodes, node)
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
	return Ways, Nodes
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
	Centroid    map[string]float64 `json:"centroid"`
	Geom        interface{} `json:"geom"`
}

func onNode(node *osmpbf.Node) JsonNode {
	marshall := JsonNode{node.ID, "node", node.Lat, node.Lon, node.Tags}
	return marshall
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
	bufval.WriteString(strconv.FormatFloat(node.Lat, 'f', 16, 64))
	bufval.WriteString(":")
	bufval.WriteString(strconv.FormatFloat(node.Lon, 'f', 16, 64))
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
	centroid["lat"] = strconv.FormatFloat(compute.Lat(), 'f', 16, 64)
	centroid["lon"] = strconv.FormatFloat(compute.Lng(), 'f', 16, 64)

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