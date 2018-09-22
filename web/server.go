package web

import (
	"github.com/maddevsio/ariadna/common"
	"github.com/julienschmidt/httprouter"
	"github.com/maddevsio/ariadna/utils"
	"gopkg.in/olivere/elastic.v3"
	"net/http"
)

var es *elastic.Client

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func StartServer() error {
	var err error
	if es, err = elastic.NewClient(elastic.SetURL(common.AC.ElasticSearchHost), ); err != nil {
		return err
	}

	router := httprouter.New()
	router.GET("/api/search/:query", geoCoder)
	router.GET("/api/reverse/:lat/:lon", reverseGeoCode)
	router.NotFound = http.FileServer(http.Dir("public"))

	addr := utils.GetAddress()
	logger.Info("Server is starting on: %s", addr)
	if err = http.ListenAndServe(addr, logHandler(router)); err != nil {
		logger.Fatal(err.Error())
	}
	return nil
}
