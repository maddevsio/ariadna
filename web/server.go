package web

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func StartServer() {
	router := httprouter.New()
	router.GET("/api/search/:query", geoCoder)
	router.GET("/api/reverse/:lat/:lon", reverseGeoCode)
	router.NotFound = http.FileServer(http.Dir("public"))
	http.ListenAndServe(":8080", router)
}
