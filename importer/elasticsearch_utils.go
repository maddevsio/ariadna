package importer

import (
	"fmt"
	"github.com/gen1us2k/ariadna/common"
	"github.com/gen1us2k/go-translit"
	"github.com/kellydunn/golang-geo"
	gj "github.com/paulmach/go.geojson"
	"gopkg.in/olivere/elastic.v3"
	"strconv"
	"strings"
)

func JsonWaysToES(Addresses []JsonWay, CitiesAndTowns []JsonWay, client *elastic.Client) {
	Logger.Info("Populating elastic search index")
	bulkClient := client.Bulk()
	Logger.Info("Creating bulk client")
	for _, address := range Addresses {
		cityName, villageName, suburbName, townName := "", "", "", ""
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
				case "town":
					townName = city.Tags["name"]
				case "neighbourhood":
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
		marshall := JsonEsIndex{
			Country:           "KG",
			City:              cityName,
			Village:           villageName,
			Town:              townName,
			District:          suburbName,
			Street:            cleanAddress(address.Tags["addr:street"]),
			HouseNumber:       housenumber,
			Name:              name,
			OldName:           address.Tags["old_name"],
			HouseName:         address.Tags["housename"],
			PostCode:          address.Tags["postcode"],
			LocalName:         address.Tags["loc_name"],
			AlternativeName:   address.Tags["alt_name"],
			InternationalName: address.Tags["int_name"],
			NationalName:      address.Tags["nat_name"],
			OfficialName:      address.Tags["official_name"],
			RegionalName:      address.Tags["reg_name"],
			ShortName:         address.Tags["short_name"],
			SortingName:       address.Tags["sorting_name"],
			TranslatedName:    translated,
			Centroid:          centroid,
			Geom:              pg,
			Custom:            false,
		}
		index := elastic.NewBulkIndexRequest().
			Index(common.C.CurrentIndex).
			Type(common.C.IndexType).
			Id(strconv.FormatInt(address.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	Logger.Info("Starting to insert many data to elasticsearch")
	_, err := bulkClient.Do()
	Logger.Info("Data insert")
	if err != nil {
		Logger.Error(err.Error())
	}
}

func JsonNodesToEs(Addresses []JsonNode, CitiesAndTowns []JsonWay, client *elastic.Client) {
	Logger.Info("Populating elastic search index with Nodes")
	bulkClient := client.Bulk()
	Logger.Info("Created bulk request to elasticsearch")
	for _, address := range Addresses {
		cityName, villageName, suburbName, townName := "", "", "", ""
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
				case "town":
					townName = city.Tags["name"]
				case "neighbourhood":
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

		marshall := JsonEsIndex{
			Country:           "KG",
			City:              cityName,
			Village:           villageName,
			Town:              townName,
			District:          suburbName,
			Street:            cleanAddress(address.Tags["addr:street"]),
			HouseNumber:       housenumber,
			Name:              name,
			TranslatedName:    translated,
			OldName:           address.Tags["old_name"],
			HouseName:         address.Tags["housename"],
			PostCode:          address.Tags["postcode"],
			LocalName:         address.Tags["loc_name"],
			AlternativeName:   address.Tags["alt_name"],
			InternationalName: address.Tags["int_name"],
			NationalName:      address.Tags["nat_name"],
			OfficialName:      address.Tags["official_name"],
			RegionalName:      address.Tags["reg_name"],
			ShortName:         address.Tags["short_name"],
			SortingName:       address.Tags["sorting_name"],
			Centroid:          centroid,
			Geom:              nil,
			Custom:            false,
		}

		index := elastic.NewBulkIndexRequest().
			Index(common.C.CurrentIndex).
			Type(common.C.IndexType).
			Id(strconv.FormatInt(address.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	Logger.Info("Started to bulk insert to elasticsearch")
	_, err := bulkClient.Do()
	Logger.Info("Data inserted")
	if err != nil {
		Logger.Error(err.Error())
	}

}
