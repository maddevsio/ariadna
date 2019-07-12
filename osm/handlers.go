package osm

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type BadRequest struct {
	Error string `json:"error"`
}

func (i *Importer) geoCodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte{1})

}

func (i *Importer) reverseGeoCodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte{1})

}
