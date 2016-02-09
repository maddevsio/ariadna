package main

import (
	"flag"
	"fmt"
	"github.com/gen1us2k/osm-geogoder/osm-importer"
	"github.com/qedus/osmpbf"
	"gopkg.in/olivere/elastic.v3"
	"io/ioutil"
	"os"
	"runtime"
)

var CitiesAndTowns, Roads []importer.JsonWay

func getSettings() importer.Settings {

	// command line flags
	batchSize := flag.Int("batch", 50000, "batch leveldb writes in batches of this size")
	configPath := flag.String("config", "config.json", "config file path")
	indexPath := flag.String("index", "index.json", "ES index settings file path")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		importer.Logger.Fatal("invalid args, you must specify a PBF file")
	}

	return importer.Settings{args[0], *batchSize, *configPath, *indexPath}
}

func getDecoder(file *os.File) *osmpbf.Decoder {
	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	return decoder
}

func main() {

	settings := getSettings()
	importer.ReadConfig(settings.ConfigPath)
	indexSettings, err := ioutil.ReadFile(settings.IndexPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
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
	_, err = client.CreateIndex(importer.C.IndexName).BodyString(string(indexSettings)).Do()
	if err != nil {
		// Handle error
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town,place~neighbourhood")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(settings.PbfPath)
	defer file.Close()
	decoder = getDecoder(file)

	importer.Logger.Info("Searching addresses")
	tags = importer.BuildTags("addr:street+addr:housenumber,amenity,shop,addr:housenumber")
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
	importer.JsonNodesToEs(Intersections, CitiesAndTowns, client)
}
