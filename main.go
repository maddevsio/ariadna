package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	"github.com/olivere/elastic"
	"github.com/qedus/osmpbf"

	"github.com/maddevsio/ariadna/common"
	"github.com/maddevsio/ariadna/importer"
	"github.com/maddevsio/ariadna/updater"
)

var (
	Roads                   []importer.JsonWay
	CitiesAndTowns          []importer.JsonWay
	Version                 = "dev"
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

func getCurrentIndexName(client *elastic.Client) (string, error) {
	res, err := client.Aliases().Index("_all").Do(context.Background())
	if err != nil {
		return "", err
	}
	for _, index := range res.IndicesByAlias(common.AC.IndexName) {
		if strings.HasPrefix(index, common.AC.IndexName) {
			return index, nil
		}
	}
	return "", nil
}

func getEnvOrDefault(key string, _default string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return _default
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
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "For testing",
			Action:  actionTest,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "index_settings",
			Usage:       "ElasticSearch Index settings",
			Destination: &indexSettingsPath,
			Value:       getEnvOrDefault("INDEX_SETTINGS", "index.json"),
		},
		cli.StringFlag{
			Name:        "custom_data",
			Usage:       "Custom data file path",
			Destination: &customDataPath,
			Value:       getEnvOrDefault("CUSTOM_DATA", "custom.json"),
		},
		cli.StringFlag{
			Name:        "es_index_name",
			Usage:       "Specify custom elasticsearch index name",
			Value:       getEnvOrDefault("ARIADNA_ES_INDEX_NAME", "addresses"),
			EnvVar:      "ARIADNA_ES_INDEX_NAME",
			Destination: &ElasticSearchIndexName,
		},
		cli.StringFlag{
			Name:        "es_pg_conn_url",
			Usage:       "Specify custom PG connection URL",
			Destination: &PGConnString,
			Value:       getEnvOrDefault("ARIADNA_PG_CONN_URL", "host=ariadna-pg user=ariadna password=ariadna dbname=ariadna sslmode=disable"),
			EnvVar:      "ARIADNA_PG_CONN_URL",
		},
		cli.StringFlag{
			Name:        "es_url",
			Usage:       "Custom url for elasticsearch e.g http://192.168.0.1:9200",
			Destination: &ElasticSearchHost,
			Value:       getEnvOrDefault("ARIADNA_ES_HOST", "http://localhost:9200/"),
			EnvVar:      "ARIADNA_ES_HOST",
		},
		cli.StringFlag{
			Name:        "es_index_type",
			Usage:       "ElasticSearch index type",
			Destination: &IndexType,
			Value:       getEnvOrDefault("ARIADNA_INDEX_TYPE", "address"),
			EnvVar:      "ARIADNA_INDEX_TYPE",
		},
		cli.StringFlag{
			Name:        "filename",
			Usage:       "filename for storing osm.pbf file",
			Destination: &FileName,
			Value:       getEnvOrDefault("ARIADNA_FILE_NAME", "xxx"),
			EnvVar:      "ARIADNA_FILE_NAME",
		},
		cli.StringFlag{
			Name:        "download_url",
			Usage:       "Geofabrik url to download file",
			Destination: &DownloadUrl,
			Value:       getEnvOrDefault("ARIADNA_DOWNLOAD_URL", "xxx"),
			EnvVar:      "ARIADNA_DOWNLOAD_URL",
		},
		cli.BoolFlag{
			Name:        "dont_import_intersections",
			Usage:       "if checked, then ariadna won't import intersections",
			Destination: &DontImportIntersections,
			EnvVar:      "ARIADNA_DONT_IMPORT_INTERSECTIONS",
		},
	}

	app.Before = func(context *cli.Context) (err error) {
		common.AC = common.AppConfig{
			IndexType:               IndexType,
			PGConnString:            PGConnString,
			ElasticSearchHost:       ElasticSearchHost,
			IndexName:               ElasticSearchIndexName,
			FileName:                FileName,
			DownloadUrl:             DownloadUrl,
			DontImportIntersections: DontImportIntersections,
		}

		common.ES, err = elastic.NewClient(
			elastic.SetURL(common.AC.ElasticSearchHost),
			elastic.SetSniff(false),
		)
		return err
	}

	if err := app.Run(os.Args); err != nil {
		importer.Logger.Fatal("error on run app, %v", err)
	}

}

func actionImport(_ *cli.Context) error {
	indexSettings, err := ioutil.ReadFile(indexSettingsPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder := getDecoder(file)

	indexVersion := fmt.Sprintf("%s_%d", common.AC.IndexName, time.Now().Unix())
	importer.Logger.Info("Creating index with name %s", indexVersion)
	_, err = common.ES.CreateIndex(indexVersion).BodyString(string(indexSettings)).Do(context.Background())

	common.AC.ElasticSearchIndexUrl = indexVersion

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
	importer.JsonWaysToES(AddressWays, CitiesAndTowns, common.ES)
	importer.JsonNodesToEs(AddressNodes, CitiesAndTowns, common.ES)
	file = importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder = getDecoder(file)

	if !common.AC.DontImportIntersections {
		tags = importer.BuildTags("highway+name")
		Roads, _ = importer.Run(decoder, db, tags)
		importer.RoadsToPg(Roads)
		importer.Logger.Info("Searching all roads intersecitons")
		Intersections := importer.GetRoadIntersectionsFromPG()
		importer.JsonNodesToEs(Intersections, CitiesAndTowns, common.ES)
	}

	importer.Logger.Info("Removing indices from alias")
	_, err = common.ES.Alias().Add(indexVersion, common.AC.IndexName).Do(context.Background())
	if err != nil {
		return err
	}
	res, err := common.ES.Aliases().Index("_all").Do(context.Background())
	if err != nil {
		return err
	}
	for _, index := range res.IndicesByAlias(common.AC.IndexName) {
		if strings.HasPrefix(index, common.AC.IndexName) && index != indexVersion {
			_, err = common.ES.Alias().Remove(index, common.AC.IndexName).Do(context.Background())
			if err != nil {
				importer.Logger.Error("Failed to delete index alias: %s", err.Error())
			}
			_, err = common.ES.DeleteIndex(index).Do(context.Background())
			if err != nil {
				importer.Logger.Error("Failed to delete index: %s", err.Error())
			}
		}
	}
	return nil
}

func actionUpdate(ctx *cli.Context) error {
	err := updater.DownloadOSMFile(common.AC.DownloadUrl, common.AC.FileName)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	actionImport(ctx)
	return nil
}

func actionTest(_ *cli.Context) error {
	currentIndex, err := getCurrentIndexName(common.ES)
	if err != nil {
		return err
	}

	fmt.Println(currentIndex)
	return nil
}

func actionHttp(_ *cli.Context) error {
	router := httprouter.New()
	router.GET("/api/search/:query", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		qs := elastic.NewQueryStringQuery(ps.ByName("query"))
		qs.Field("name")
		qs.Field("street")
		qs.Field("housenumber")
		qs.Field("district")
		qs.Field("old_name")
		qs.Field("town")
		qs.Field("city")
		qs.Analyzer("map_synonyms")
		result, err := common.ES.Search().Index("addresses").Query(qs).Do(context.Background())
		if err != nil {
			resp, _ := json.Marshal(common.BadRequest{Error: err.Error()})
			w.Write(resp)
			return
		}
		var results []importer.JsonEsIndex
		var res importer.JsonEsIndex
		for _, item := range result.Each(reflect.TypeOf(res)) {
			t := item.(importer.JsonEsIndex)
			results = append(results, t)
		}
		data, _ := json.Marshal(results)
		w.Write(data)
	})
	router.GET("/api/reverse/:lat/:lon", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		qs := elastic.NewGeoDistanceQuery("location")
		fmt.Printf("Got %s and %s\n", ps.ByName("lat"), ps.ByName("lon"))
		lat, _ := strconv.ParseFloat(ps.ByName("lat"), 64)
		lon, _ := strconv.ParseFloat(ps.ByName("lon"), 64)
		qs.GeoPoint(elastic.GeoPointFromLatLon(lat, lon))
		qs.Distance("10m")
		qs.QueryName("filtered")

		result, err := common.ES.Search().Index("addresses").Query(qs).Do(context.Background())
		if err != nil {
			resp, _ := json.Marshal(common.BadRequest{Error: err.Error()})
			w.Write(resp)
			return
		}
		var results []importer.JsonEsIndex
		var res importer.JsonEsIndex
		for _, item := range result.Each(reflect.TypeOf(res)) {
			t := item.(importer.JsonEsIndex)
			results = append(results, t)
		}
		data, _ := json.Marshal(results)
		w.Write(data)
	})
	router.NotFound = http.FileServer(http.Dir("public"))
	return http.ListenAndServe(":8080", router)
}

func actionCustom(_ *cli.Context) error {
	var custom common.Custom
	data, err := ioutil.ReadFile(customDataPath)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}
	err = json.Unmarshal(data, &custom)
	if err != nil {
		importer.Logger.Fatal(err.Error())
	}

	bulkClient := common.ES.Bulk()
	indexVersion, err := getCurrentIndexName(common.ES)
	if err != nil {
		return err
	}
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
			Index(indexVersion).
			Type(common.AC.IndexType).
			Id(strconv.FormatInt(item.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	_, err = bulkClient.Do(context.Background())
	if err != nil {
		importer.Logger.Error(err.Error())
	}
	return nil
}

func actionIntersection(_ *cli.Context) error {
	db := importer.OpenLevelDB("db")
	defer db.Close()

	file := importer.OpenFile(common.AC.FileName)
	defer file.Close()
	decoder := getDecoder(file)

	indexVersion, err := getCurrentIndexName(common.ES)
	if err != nil {
		return err
	}

	common.AC.ElasticSearchIndexUrl = indexVersion
	importer.Logger.Info("Creating index with name %s", common.AC.ElasticSearchIndexUrl)

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
	importer.JsonNodesToEs(Intersections, CitiesAndTowns, common.ES)
	return nil
}
