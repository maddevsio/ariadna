package web

import (
	"gopkg.in/olivere/elastic.v3"
	"reflect"
	"github.com/maddevsio/ariadna/importer"
	"encoding/json"
)

//ToDo: get index name from config
func executeQuery(qs elastic.Query) (*elastic.SearchResult, error) {
	return es.Search().Index("addresses").Query(qs).Do()
}

func getGeoCoderResult(query string) (*elastic.SearchResult, error) {
	qs := geoCoderQuery(query)
	return executeQuery(qs)
}

func getReverseGeoCodeQuery(lat, lon float64) (*elastic.SearchResult, error) {
	qs := reverseGeoCodeQuery(lat, lon)
	return executeQuery(qs)
}

func esResultToJson(result *elastic.SearchResult) ([]byte, error) {
	var items []importer.JsonEsIndex
	var res importer.JsonEsIndex
	for _, elem := range result.Each(reflect.TypeOf(res)) {
		item := elem.(importer.JsonEsIndex)
		items = append(items, item)
	}
	return json.Marshal(items)
}
