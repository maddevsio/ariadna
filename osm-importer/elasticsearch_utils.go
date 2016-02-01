package importer

import (
	"fmt"
	"github.com/gen1us2k/go-translit"
	"github.com/kellydunn/golang-geo"
	gj "github.com/paulmach/go.geojson"
	"gopkg.in/olivere/elastic.v3"
	"strconv"
	"strings"
)

func JsonWaysToES(Addresses []JsonWay, CitiesAndTowns []JsonWay, client *elastic.Client) {
	if Logger.IsInfo() {
		Logger.Info("Populating elastic search index")
	}
	bulkClient := client.Bulk()
	if Logger.IsInfo() {
		Logger.Info("Creating bulk client")
	}
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
		name := cleanAddress(address.Tags["name"])
		translated := ""

		if latinre.Match([]byte(name)) {
			word := make(map[string]string)
			word["original"] = name

			trans := strings.Split(name, " ")
			for _, k := range trans {
				s := synonims[k]
				if s == "" {
					s = translit.Translit(k)
				}
				translated += fmt.Sprintf("%s ", s)
			}

			word["trans"] = translated
		}
		housenumber := translit.Translit(address.Tags["addr:housenumber"])
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), housenumber, name, centroid, pg}
		index := elastic.NewBulkIndexRequest().Index(C.IndexName).Type(C.IndexType).Id(strconv.FormatInt(address.ID, 10)).Doc(marshall)
		bulkClient = bulkClient.Add(index)
		if translated != "" {
			marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), housenumber, translated, centroid, pg}
			index = elastic.NewBulkIndexRequest().Index(C.IndexName).Type(C.IndexType).Id(strconv.FormatInt(address.ID*2, 10)).Doc(marshall)
			bulkClient = bulkClient.Add(index)
		}

	}
	if Logger.IsInfo() {
		Logger.Info("Starting to insert many data to elasticsearch")
	}
	_, err := bulkClient.Do()
	if Logger.IsInfo() {
		Logger.Info("Data insert")
	}
	if err != nil {
		Logger.Error(err.Error())
	}
}

func JsonNodesToEs(Addresses []JsonNode, CitiesAndTowns []JsonWay, client *elastic.Client) {
	if Logger.IsInfo() {
		Logger.Info("Populating elastic search index with Nodes")
	}
	bulkClient := client.Bulk()
	if Logger.IsInfo() {
		Logger.Info("Created bulk request to elasticsearch")
	}
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
		name := cleanAddress(address.Tags["name"])
		translated := ""
		if latinre.Match([]byte(name)) {
			word := make(map[string]string)
			word["original"] = name

			trans := strings.Split(name, " ")
			for _, k := range trans {
				s := synonims[k]
				if s == "" {
					s = translit.Translit(k)
				}
				translated += fmt.Sprintf("%s ", s)
			}

			word["trans"] = translated
		}
		housenumber := translit.Translit(address.Tags["addr:housenumber"])
		marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), housenumber, name, centroid, nil}
		index := elastic.NewBulkIndexRequest().Index(C.IndexName).Type(C.IndexType).Id(strconv.FormatInt(address.ID, 10)).Doc(marshall)
		bulkClient = bulkClient.Add(index)
		if translated != "" {
			marshall := JsonEsIndex{"KG", cityName, villageName, suburbName, cleanAddress(address.Tags["addr:street"]), housenumber, translated, centroid, nil}
			index = elastic.NewBulkIndexRequest().Index(C.IndexName).Type(C.IndexType).Id(strconv.FormatInt(address.ID*2, 10)).Doc(marshall)
			bulkClient = bulkClient.Add(index)
		}

	}
	if Logger.IsInfo() {
		Logger.Info("Started to bulk insert to elasticsearch")
	}
	_, err := bulkClient.Do()
	if Logger.IsInfo() {
		Logger.Info("Data inserted")
	}
	if err != nil {
		Logger.Error(err.Error())
	}

}
