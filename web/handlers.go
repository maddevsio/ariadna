package web
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/olivere/elastic.v3"
	"reflect"
	"github.com/gen1us2k/ariadna/importer"
	"encoding/json"
	"strconv"
	"fmt"
)

type BadRequest struct {
	Error string `json:"error"`
}

func geoCoder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	qs := elastic.NewQueryStringQuery(ps.ByName("query"))
	qs.Field("name")
	qs.Field("street")
	qs.Field("housenumber")
	qs.Field("district")
	qs.Field("old_name")
	qs.Field("town")
	qs.Field("city")
	qs.Analyzer("map_synonyms")
	result, err := es.Search().Index("addresses").Query(qs).Do()
	if err != nil {
		resp, _ := json.Marshal(BadRequest{err.Error()})
		w.Write(resp)
	}
	var results []importer.JsonEsIndex
	var res importer.JsonEsIndex
	for _, item := range result.Each(reflect.TypeOf(res)) {
		t := item.(importer.JsonEsIndex)
		results = append(results, t)
	}
	data, _ :=json.Marshal(results)
	w.Write(data)

}

func reverseGeoCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
	qs := elastic.NewGeoDistanceQuery("centroid")
	fmt.Printf("Got %s and %s\n", ps.ByName("lat"), ps.ByName("lon"))
	lat, _ := strconv.ParseFloat(ps.ByName("lat"), 64)
	lon, _ := strconv.ParseFloat(ps.ByName("lon"), 64)
	qs.Lat(lat)
	qs.Lon(lon)
	qs.Distance("200m")

	result, err := es.Search().Index("addresses").Query(qs).Do()
	if err != nil {
		resp, _ := json.Marshal(BadRequest{err.Error()})
		w.Write(resp)
	}
	var results []importer.JsonEsIndex
	var res importer.JsonEsIndex
	for _, item := range result.Each(reflect.TypeOf(res)) {
		t := item.(importer.JsonEsIndex)
		results = append(results, t)
	}
	data, _ :=json.Marshal(results)
	w.Write(data)

}
func index(w http.ResponseWriter, r *http.Request) {

}
