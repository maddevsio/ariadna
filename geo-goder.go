
// Useful tags for Bishkek
// addr:street+addr:housenumber - Get all known addresses
// place~city - Get all cities
package main

import (
	"log"
	"encoding/json"
	"fmt"
	"os"
	"bytes"
	"flag"
	"io"
	"math"
	"strconv"
	"strings"
	"runtime"

	"database/sql"
	"github.com/paulmach/go.geo"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/qedus/osmpbf"
	_ "github.com/lib/pq"
)


type Settings struct {
	PbfPath           string
	Tags              map[string][]string
	BatchSize         int
}

func getSettings() Settings {

	// command line flags
	tagList := flag.String("tags", "", "comma-separated list of valid tags, group AND conditions with a +")
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")

	flag.Parse()
	args := flag.Args();

	if len( args ) < 1 {
		log.Fatal("invalid args, you must specify a PBF file")
	}

	// invalid tags
	if( len(*tagList) < 1 ){
		log.Fatal("Nothing to do, you must specify tags to match against")
	}

	// parse tag conditions
	conditions := make(map[string][]string)
	for _, group := range strings.Split(*tagList,",") {
		conditions[group] = strings.Split(group,"+")
	}

	// fmt.Print(conditions, len(conditions))
	// os.Exit(1)

	return Settings{ args[0], conditions, *batchSize }
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

	db := openLevelDB("db")
	defer db.Close()

	pg_db, err := sql.Open("postgres", "host=localhost user=geo password=geo dbname=geo sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	run(decoder, db, config, pg_db)
}

func run(d *osmpbf.Decoder, db *leveldb.DB, config Settings, pg_db *sql.DB){

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

				// inc count
				nc++

				// ----------------
				// write to leveldb
				// ----------------

				// write immediately
				// cacheStore(db, v)

				// write in batches
				cacheQueue(batch, v)
				if batch.Len() > config.BatchSize {
					cacheFlush(db, batch)
				}

				// ----------------
				// handle tags
				// ----------------

				if !hasTags(v.Tags) { break }

				v.Tags = trimTags(v.Tags)
				if containsValidTags( v.Tags, config.Tags ) {
					onNode(v)
				}

			case *osmpbf.Way:

				// ----------------
				// write to leveldb
				// ----------------

				// flush outstanding batches

				if batch.Len() > 1 {
					cacheFlush(db, batch)
				}

				// inc count
				wc++

				if !hasTags(v.Tags) { break }

				v.Tags = trimTags(v.Tags)
				if containsValidTags( v.Tags, config.Tags ) {

					// lookup from leveldb
					latlons, err := cacheLookup(db, v)

					// skip ways which fail to denormalize
					if err != nil { break }

					// compute centroid
					var centroid = computeCentroid(latlons);

					onWay(v,latlons,centroid, pg_db)
				}

			case *osmpbf.Relation:
				if !hasTags(v.Tags) { break }
				v.Tags = trimTags(v.Tags)
				if containsValidTags( v.Tags, config.Tags ) {
//					fmt.Println(v);
				}

				rc++

			default:

				log.Fatalf("unknown type %T\n", v)

			}
		}
	}
	// fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
}

type JsonNode struct {
	ID        int64               `json:"id"`
	Type      string              `json:"type"`
	Lat       float64             `json:"lat"`
	Lon       float64             `json:"lon"`
	Tags      map[string]string   `json:"tags"`
	//	Timestamp time.Time           `json:"timestamp"`
}

type JsonRelation struct {
	ID        int64               `json:"id"`
	Type      string              `json:"type"`
	Tags      map[string]string   `json:"tags"`
	Centroid  map[string]string   `json:"centroid"`
	Nodes     []map[string]string `json:"nodes"`
	//	Timestamp time.Time           `json:"timestamp"`
}

func onNode(node *osmpbf.Node){
	marshall := JsonNode{ node.ID, "node", node.Lat, node.Lon, node.Tags}
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
}

type JsonWay struct {
	ID        int64               `json:"id"`
	Type      string              `json:"type"`
	Tags      map[string]string   `json:"tags"`
	Centroid  map[string]string   `json:"centroid"`
	Nodes     []map[string]string `json:"nodes"`
	//	Timestamp time.Time           `json:"timestamp"`
}
type Tags struct  {
	housenumber string
	street string
}

const insQuery  = `INSERT INTO addresses(node_id, housenumber, street, centroid, coords)
 values($1, $2, $3, ST_MakePoint($4, $5), ST_MakePolygon(ST_GeomFromText($6)));`
func onWay(way *osmpbf.Way, latlons []map[string]string, centroid map[string]string, pg_db *sql.DB){
	marshall := JsonWay{ way.ID, "way", way.Tags/*, way.NodeIDs*/, centroid, latlons, }
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
//	var linestring string;
//	linestring = "LINESTRING("
//	for _, latlon := range latlons{
//		linestring += fmt.Sprintf("%s %s,",latlon["lon"], latlon["lat"])
//	}
//	linestring = linestring[:len(linestring) - 1]
//	linestring += ")"
//	insert_query, err := pg_db.Prepare(insQuery)
//	defer insert_query.Close()
//	if err != nil {
//		fmt.Printf("Query preparation error -->%v\n", err)
//		panic("Test query error")
//	}
//	_, err = insert_query.Exec(way.ID, way.Tags["addr:housenumber"], way.Tags["addr:street"], centroid["lon"], centroid["lat"], linestring)
//
//	if err != nil {
//		fmt.Printf("Query execution error -->%v\n", err)
//		panic("Error")
//	}

}

func onRelation(way *osmpbf.Relation, latlons []map[string]string, centroid map[string]string, pg_db *sql.DB){
	// do nothing (yet)
	marshall := JsonRelation{ way.ID, "way", way.Tags/*, way.NodeIDs*/, centroid, latlons, }
	json, _ := json.Marshal(marshall)
	fmt.Println(string(json))
}

// write to leveldb immediately
func cacheStore(db *leveldb.DB, node *osmpbf.Node){
	id, val := formatLevelDB(node)
	err := db.Put([]byte(id), []byte(val), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// queue a leveldb write in a batch
func cacheQueue(batch *leveldb.Batch, node *osmpbf.Node){
	id, val := formatLevelDB(node)
	batch.Put([]byte(id), []byte(val))
}

// flush a leveldb batch to database and reset batch to 0
func cacheFlush(db *leveldb.DB, batch *leveldb.Batch){
	err := db.Write(batch, nil)
	if err != nil {
		log.Fatal(err)
	}
	batch.Reset()
}

func cacheLookup(db *leveldb.DB, way *osmpbf.Way) ([]map[string]string, error) {

	var container []map[string]string

	for _, each := range way.NodeIDs {
		stringid := strconv.FormatInt(each,10)

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



func formatLevelDB(node *osmpbf.Node) (id string, val []byte){

	stringid := strconv.FormatInt(node.ID,10)

	var bufval bytes.Buffer
	bufval.WriteString(strconv.FormatFloat(node.Lat,'f',6,64))
	bufval.WriteString(":")
	bufval.WriteString(strconv.FormatFloat(node.Lon,'f',6,64))
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

		feature := strings.Split(name,"~")
		foundVal, foundKey := tags[feature[0]]

		// key check
		if !foundKey {
			return false
		}

		// value check
		if len( feature ) > 1 {
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
		if matchTagsAgainstCompulsoryTagList( tags, list ){
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

	points := geo.PointSet{}
	for _, each := range latlons {
		var lon, _ = strconv.ParseFloat( each["lon"], 64 );
		var lat, _ = strconv.ParseFloat( each["lat"], 64 );
		points.Push( geo.NewPoint( lon, lat ))
	}

	var compute = getCentroid(points);

	var centroid = make(map[string]string)
	centroid["lat"] = strconv.FormatFloat(compute.Lat(),'f',6,64)
	centroid["lon"] = strconv.FormatFloat(compute.Lng(),'f',6,64)

	return centroid
}

// compute the centroid of a polygon set
// using a spherical co-ordinate system
func getCentroid(ps geo.PointSet) *geo.Point {

	X := 0.0
	Y := 0.0
	Z := 0.0

	var toRad = math.Pi / 180
	var fromRad = 180 / math.Pi

	for _, point := range ps {

		var lon = point[0] * toRad
		var lat = point[1] * toRad

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

	var centroid = geo.NewPoint(lon * fromRad, lat * fromRad)

	return centroid;
}