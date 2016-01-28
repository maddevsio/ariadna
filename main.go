// Useful tags for Bishkek
// addr:street+addr:housenumber - Get all known addresses
// place~city - Get all cities
// place~suburb - Get districts
// place~village - Get villages
// building,shop - get all buildings and shops
// highway - Get all roads

package main

import (
	"flag"
	"github.com/gen1us2k/osm-geogoder/osm-importer"
	"github.com/qedus/osmpbf"
	"gopkg.in/olivere/elastic.v3"
	"os"
	"runtime"
)

var PlaceSynonyms = map[string][]string{
	"American University of Central Asia": []string{"АУЦА", "Американский университет в центральной азии", "AUCA"},
}

var CitiesAndTowns, Roads []importer.JsonWay

func getSettings() importer.Settings {

	// command line flags
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		importer.Logger.Fatal("invalid args, you must specify a PBF file")
	}

	return importer.Settings{args[0], *batchSize}
}

func getDecoder(file *os.File) *osmpbf.Decoder {
	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	return decoder
}

func main() {

	config := getSettings()
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(config.PbfPath)
	defer file.Close()
	decoder := getDecoder(file)

	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}
	_, err = client.CreateIndex("addresses").BodyString(importer.ESSettings).Do()
	if err != nil {
		// Handle error
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(config.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	importer.Logger.Info("Searching addresses")
	tags = importer.BuildTags("addr:street+addr:housenumber,amenity,shop")
	AddressWays, AddressNodes := importer.Run(decoder, db, tags)
	importer.Logger.Info("Addresses found")
	importer.JsonWaysToES(AddressWays, CitiesAndTowns, client)
	importer.JsonNodesToEs(AddressNodes, CitiesAndTowns, client)
	//	file = openFile(config.PbfPath)
	//	defer file.Close()
	//	decoder = getDecoder(file)

	//	tags = buildTags("highway")
	//	Roads, _ = run(decoder, db, tags)
	//	RoadsToPg()
	//	fmt.Println("Searching all roads intersecitons")
	//	Intersections := GetRoadIntersectionsFromPG()
	//	JsonNodesToEs(Intersections, client)
}

var Translations []importer.Translate
