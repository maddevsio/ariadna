package main

import (
	"flag"
	"github.com/gen1us2k/osm-geogoder/osm-importer"
	"github.com/qedus/osmpbf"
	"gopkg.in/olivere/elastic.v3"
	"os"
	"runtime"
	"fmt"
)

var CitiesAndTowns, Roads []importer.JsonWay

func getSettings() importer.Settings {

	// command line flags
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")
	configPath := flag.String("config", "config.json", "config file path")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		importer.Logger.Fatal("invalid args, you must specify a PBF file")
	}

	return importer.Settings{args[0], *batchSize, *configPath}
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

	settings := getSettings()
	importer.ReadConfig(settings.ConfigPath)
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(settings.PbfPath)
	fmt.Println(importer.C.IndexName)
	defer file.Close()
	decoder := getDecoder(file)

	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}
	_, err = client.CreateIndex(importer.C.IndexName).BodyString(importer.ESSettings).Do()
	if err != nil {
		// Handle error
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(settings.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	importer.Logger.Info("Searching addresses")
	tags = importer.BuildTags("addr:street+addr:housenumber,amenity,shop")
	AddressWays, AddressNodes := importer.Run(decoder, db, tags)
	importer.Logger.Info("Addresses found")
	importer.JsonWaysToES(AddressWays, CitiesAndTowns, client)
	importer.JsonNodesToEs(AddressNodes, CitiesAndTowns, client)
	file = importer.OpenFile(settings.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	tags = importer.BuildTags("highway")
	Roads, _ = importer.Run(decoder, db, tags)
	importer.RoadsToPg(Roads)
	importer.Logger.Info("Searching all roads intersecitons")
	Intersections := importer.GetRoadIntersectionsFromPG()
	importer.JsonNodesToEs(Intersections,CitiesAndTowns, client)
}