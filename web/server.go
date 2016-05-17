package web

import (
	"github.com/gen1us2k/ariadna/common"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/olivere/elastic.v3"
	"net/http"
)

var es *elastic.Client

func StartServer() error {
	var err error
	es, err = elastic.NewClient(
		elastic.SetURL(common.AC.ElasticSearchHost),
	)
	if err != nil {
		return err
	}
	router := httprouter.New()
	router.GET("/api/search/:query", geoCoder)
	router.GET("/api/reverse/:lat/:lon", reverseGeoCode)
	router.NotFound = http.FileServer(http.Dir("public"))
	http.ListenAndServe(":8080", router)
	return nil
}
