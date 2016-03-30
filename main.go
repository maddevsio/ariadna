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
	CitiesAndTowns, Roads   []importer.JsonWay
	Version                 string = "dev"
	configPath              string
	indexSettingsPath       string
	customDataPath          string
	ElasticSearchIndexName  string
	PGConnString            string
	ElasticSearchHost       string
	IndexType               string
	FileName                string
	DownloadUrl             string
	DontImportIntersections bool
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
		{
			Name:    "intersections",
			Aliases: []string{"i"},
			Usage:   "Process intersections only",
			Action:  actionIntersection,
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
		cli.StringFlag{
			Name:        "es_index_name",
			Usage:       "Specify custom elasticsearch index name",
			Value:       "addresses",
			EnvVar:      "ARIADNA_ES_INDEX_NAME",
			Destination: &ElasticSearchIndexName,
		},
		cli.StringFlag{
			Name:        "es_pg_conn_url",
			Usage:       "Specify custom PG connection URL",
			Destination: &PGConnString,
			Value:       "host=localhost user=geo password=geo dbname=geo sslmode=disable",
			EnvVar:      "ARIADNA_PG_CONN_URL",
		},
		cli.StringFlag{
			Name:        "es_url",
			Usage:       "Custom url for elasticsearch e.g http://192.168.0.1:9200",
			Destination: &ElasticSearchHost,
			Value:       "http://localhost:9200",
			EnvVar:      "ARIADNA_ES_HOST",
		},
		cli.StringFlag{
			Name:        "es_index_type",
			Usage:       "ElasticSearch index type",
			Destination: &IndexType,
			Value:       "address",
			EnvVar:      "ARIADNA_INDEX_TYPE",
		},
		cli.StringFlag{
			Name:        "filename",
			Usage:       "filename for storing osm.pbf file",
			Destination: &FileName,
			Value:       "xxx",
			EnvVar:      "ARIADNA_FILE_NAME",
		},
		cli.StringFlag{
			Name:        "download_url",
			Usage:       "Geofabrik url to download file",
			Destination: &DownloadUrl,
			Value:       "xxx",
			EnvVar:      "ARIADNA_DOWNLOAD_URL",
		},
		cli.BoolFlag{
			Name:        "dont_import_intersections",
			Usage:       "if checked, then ariadna won't import intersections",
			Destination: &DontImportIntersections,
			EnvVar:      "ARIADNA_DONT_IMPORT_INTERSECTIONS",
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
	common.AC = common.AppConfig{
		IndexType:               IndexType,
		PGConnString:            PGConnString,
		ElasticSearchHost:       ElasticSearchHost,
		IndexName:               ElasticSearchIndexName,
		FileName:                FileName,
		DownloadUrl:             DownloadUrl,
		DontImportIntersections: DontImportIntersections,
	}
	indexSettings, err := ioutil.ReadFile(indexSettingsPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder := getDecoder(file)

	client, err := elastic.NewClient()

	if err != nil {
		importer.Logger.Fatal("Failed to create Elastic search client with error %s", err)
	}

	indexVersion := fmt.Sprintf("%s_v%d", common.AC.IndexName, common.IC.IndexVersion+1)
	common.IC.CurrentIndex = indexVersion
	importer.Logger.Info("Creating index with name %s", common.IC.CurrentIndex)
	_, err = client.CreateIndex(common.IC.CurrentIndex).BodyString(string(indexSettings)).Do()
	if err != nil {
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town,place~neighbourhood")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	importer.Logger.Info("Searching addresses")
	tags = importer.BuildTags("addr:street+addr:housenumber,amenity+name,building+name,addr:housenumber,shop+name,office+name,public_transport+name,cuisine+name,railway+name,sport+name,natural+name,tourism+name,leisure+name,historic+name,man_made+name,landuse+name,waterway+name,aerialway+name,aeroway+name,craft+name,military+name")
	AddressWays, AddressNodes := importer.Run(decoder, db, tags)
	importer.Logger.Info("Addresses found")
	importer.JsonWaysToES(AddressWays, CitiesAndTowns, client)
	importer.JsonNodesToEs(AddressNodes, CitiesAndTowns, client)
	file = importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	if !common.AC.DontImportIntersections {
		tags = importer.BuildTags("highway+name")
		Roads, _ = importer.Run(decoder, db, tags)
		importer.RoadsToPg(Roads)
		importer.Logger.Info("Searching all roads intersecitons")
		Intersections := importer.GetRoadIntersectionsFromPG()
		importer.JsonNodesToEs(Intersections, CitiesAndTowns, client)
	}
	common.IC.LastIndexVersion = common.IC.IndexVersion
	common.IC.IndexVersion += 1
	data, err := json.Marshal(common.IC)
	if err != nil {
		importer.Logger.Error("Failed to encode to json: %s", err)
	}
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		importer.Logger.Error("Failed to write file %s", err)
	}

	_, err = client.Alias().
		Remove(fmt.Sprintf("%s_v%d", common.AC.IndexName, common.IC.LastIndexVersion), common.AC.IndexName).
		Add(common.IC.CurrentIndex, common.AC.IndexName).Do()
	if err != nil {
		importer.Logger.Error("Failed to change aliases because of %s", err)
		importer.Logger.Info("Creating index alias")
		_, err = client.Alias().
			Add(common.IC.CurrentIndex, common.AC.IndexName).Do()
		if err != nil {
			importer.Logger.Error("Failed to create index: %s", err)
		}
	}
	_, err = client.DeleteIndex(fmt.Sprintf("%s_v%d", common.AC.IndexName, common.IC.LastIndexVersion)).Do()

	if err != nil {
		importer.Logger.Error("Failed to delete index %s: %s", common.IC.LastIndexVersion, err)
	}
}

func actionUpdate(ctx *cli.Context) {
	common.ReadConfig(configPath)
	common.AC = common.AppConfig{
		IndexType:               IndexType,
		PGConnString:            PGConnString,
		ElasticSearchHost:       ElasticSearchHost,
		IndexName:               ElasticSearchIndexName,
		FileName:                FileName,
		DownloadUrl:             DownloadUrl,
		DontImportIntersections: DontImportIntersections,
	}
	err := updater.DownloadOSMFile(common.AC.DownloadUrl, common.AC.FileName)
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
	common.AC = common.AppConfig{
		IndexType:               IndexType,
		PGConnString:            PGConnString,
		ElasticSearchHost:       ElasticSearchHost,
		IndexName:               ElasticSearchIndexName,
		FileName:                FileName,
		DownloadUrl:             DownloadUrl,
		DontImportIntersections: DontImportIntersections,
	}
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
	elastic.SetURL()
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
			Index(common.IC.CurrentIndex).
			Type(common.AC.IndexType).
			Id(strconv.FormatInt(item.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	_, err = bulkClient.Do()
	if err != nil {
		importer.Logger.Error(err.Error())
	}
}

func actionIntersection(ctx *cli.Context) {
	common.AC = common.AppConfig{
		IndexType:               IndexType,
		PGConnString:            PGConnString,
		ElasticSearchHost:       ElasticSearchHost,
		IndexName:               ElasticSearchIndexName,
		FileName:                FileName,
		DownloadUrl:             DownloadUrl,
		DontImportIntersections: DontImportIntersections,
	}

	common.ReadConfig(configPath)
	indexSettings, err := ioutil.ReadFile(indexSettingsPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder := getDecoder(file)

	client, err := elastic.NewClient()

	if err != nil {
		importer.Logger.Fatal("Failed to create Elastic search client with error %s", err)
	}

	indexVersion := fmt.Sprintf("%s_v%d", common.AC.IndexName, common.IC.IndexVersion)
	common.IC.CurrentIndex = indexVersion
	importer.Logger.Info("Creating index with name %s", common.IC.CurrentIndex)
	_, err = client.CreateIndex(common.IC.CurrentIndex).BodyString(string(indexSettings)).Do()
	if err != nil {
		importer.Logger.Error(err.Error())
	}

	importer.Logger.Info("Searching cities, villages, towns and districts")
	tags := importer.BuildTags("place~city,place~village,place~suburb,place~town,place~neighbourhood")
	CitiesAndTowns, _ = importer.Run(decoder, db, tags)

	importer.Logger.Info("Cities, villages, towns and districts found")

	file = importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	tags = importer.BuildTags("highway+name")
	Roads, _ = importer.Run(decoder, db, tags)
	importer.BuildIndex(Roads)
	Intersections := importer.SearchIntersections(Roads)
	importer.JsonNodesToEs(Intersections, CitiesAndTowns, client)
}
