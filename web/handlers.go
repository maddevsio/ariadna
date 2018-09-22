package web

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type BadRequest struct {
	Error string `json:"error"`
}

func geoCoder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := ps.ByName("query")

	result, err := getGeoCoderResult(query)
	if err != nil {
		resp, _ := json.Marshal(BadRequest{err.Error()})
		w.Write(resp)
	}
	data, _ := esResultToJson(result)
	w.Write(data)
}

func reverseGeoCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	lat, _ := strconv.ParseFloat(ps.ByName("lat"), 64)
	lon, _ := strconv.ParseFloat(ps.ByName("lon"), 64)

	result, err := getReverseGeoCodeQuery(lat, lon)
	if err != nil {
		resp, _ := json.Marshal(BadRequest{err.Error()})
		w.Write(resp)
	}
	data, _ := esResultToJson(result)
	w.Write(data)
}
