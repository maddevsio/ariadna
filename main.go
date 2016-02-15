package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gen1us2k/ariadna/common"
	"github.com/gen1us2k/ariadna/importer"
	"github.com/gen1us2k/ariadna/updater"
	"github.com/gen1us2k/ariadna/web"
	"github.com/qedus/osmpbf"
	"gopkg.in/olivere/elastic.v3"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
)

var (
	CitiesAndTowns, Roads []importer.JsonWay
	Version               string = "dev"
	configPath            string
	indexSettingsPath     string
	customDataPath        string
)

func getDecoder(file *os.File) *osmpbf.Decoder {
	decoder := osmpbf.NewDecoder(file)
	err := decoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	return decoder
}

func main() {
	app := cli.NewApp()
	app.Name = "Ariadna"
	app.Usage = "OSM Geocoder"
	app.Version = Version

	app.Commands = []cli.Command{
		{
			Name:      "import",
			Aliases:   []string{"i"},
			Usage:     "Import OSM file to ElasticSearch",
			Action:    actionImport,
			ArgsUsage: "<filename>",
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Download OSM file and update index",
			Action:  actionUpdate,
		},
		{
			Name:    "http",
			Aliases: []string{"h"},
			Usage:   "Run http server",
			Action:  actionHttp,
		},
		{
			Name:    "custom",
			Aliases: []string{"c"},
			Usage:   "Import custom data",
			Action:  actionCustom,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "Config file path",
			Destination: &configPath,
		},
		cli.StringFlag{
			Name:        "index_settings",
			Usage:       "ElasticSearch Index settings",
			Destination: &indexSettingsPath,
		},
		cli.StringFlag{
			Name:        "custom_data",
			Usage:       "Custom data file path",
			Destination: &customDataPath,
		},
	}

	app.Before = func(context *cli.Context) error {
		if configPath == "" {
			configPath = "config.json"
		}
		if indexSettingsPath == "" {
			indexSettingsPath = "index.json"
		}
		if customDataPath == "" {
			customDataPath = "custom.json"
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		importer.Logger.Fatal("error on run app, %v", err)
	}

}
func actionImport(ctx *cli.Context) {
	common.ReadConfig(configPath)
	indexSettings, err := ioutil.ReadFile(indexSettingsPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(common.C.FileName)
	defer file.Close()
	decoder := getDecoder(file)

	client, err := elastic.NewClient()

	if err != nil {
		importer.Logger.Fatal("Failed to create Elastic search client with error %s", err)
	}

	indexVersion := fmt.Sprintf("%s_v%d", common.C.IndexName, common.C.IndexVersion+1)
	common.C.CurrentIndex = indexVersion
	importer.Logger.Info("Creating index with name %s", common.C.CurrentIndex)
	_, err = client.CreateIndex(common.C.CurrentIndex).BodyString(string(indexSettings)).Do()
	if err != nil {
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town,place~neighbourhood")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(common.C.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	importer.Logger.Info("Searching addresses")
	tags = importer.BuildTags("addr:street+addr:housenumber,amenity,shop,addr:housenumber")
	AddressWays, AddressNodes := importer.Run(decoder, db, tags)
	importer.Logger.Info("Addresses found")
	importer.JsonWaysToES(AddressWays, CitiesAndTowns, client)
	importer.JsonNodesToEs(AddressNodes, CitiesAndTowns, client)
	file = importer.OpenFile(common.C.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	if !common.C.DontImportIntersections {
		tags = importer.BuildTags("highway")
		Roads, _ = importer.Run(decoder, db, tags)
		importer.RoadsToPg(Roads)
		importer.Logger.Info("Searching all roads intersecitons")
		Intersections := importer.GetRoadIntersectionsFromPG()
		importer.JsonNodesToEs(Intersections, CitiesAndTowns, client)
	}
	common.C.LastIndexVersion = common.C.IndexVersion
	common.C.IndexVersion += 1
	data, err := json.Marshal(common.C)
	if err != nil {
		importer.Logger.Error("Failed to encode to json: %s", err)
	}
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		importer.Logger.Error("Failed to write file %s", err)
	}

	_, err = client.Alias().
		Remove(fmt.Sprintf("%s_v%d", common.C.IndexName, common.C.LastIndexVersion), common.C.IndexName).
		Add(common.C.CurrentIndex, common.C.IndexName).Do()
	if err != nil {
		importer.Logger.Error("Failed to change aliases because of %s", err)
		importer.Logger.Info("Creating index alias")
		_, err = client.Alias().
			Add(common.C.CurrentIndex, common.C.IndexName).Do()
		if err != nil {
			importer.Logger.Error("Failed to create index: %s", err)
		}
	}
	_, err = client.DeleteIndex(fmt.Sprintf("%s_v%d", common.C.IndexName, common.C.LastIndexVersion)).Do()

	if err != nil {
		importer.Logger.Error("Failed to delete index %s: %s", common.C.LastIndexVersion, err)
	}
}

func actionUpdate(ctx *cli.Context) {
	common.ReadConfig(configPath)
	err := updater.DownloadOSMFile(common.C.DownloadUrl, common.C.FileName)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	actionImport(ctx)
}

func actionHttp(ctx *cli.Context) {
	web.StartServer()
}

type CustomData struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}
type Custom []CustomData

func actionCustom(ctx *cli.Context) {
	common.ReadConfig(configPath)
	var custom Custom
	data, err := ioutil.ReadFile(customDataPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	err = json.Unmarshal(data, &custom)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	client, err := elastic.NewClient()
	bulkClient := client.Bulk()
	for _, item := range custom {
		centroid := make(map[string]float64)
		centroid["lat"] = item.Lat
		centroid["lon"] = item.Lon
		marshall := importer.JsonEsIndex{
			Name:     item.Name,
			Centroid: centroid,
			Custom:   true,
		}
		index := elastic.NewBulkIndexRequest().
			Index(common.C.CurrentIndex).
			Type(common.C.IndexType).
			Id(strconv.FormatInt(item.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	_, err = bulkClient.Do()
	if err != nil {
		importer.Logger.Error(err.Error())
	}
}
