package web
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func StartServer() {
	router := httprouter.New()
	router.GET("/api/search/:query", geoCoder)
	router.GET("/api/reverse/:lat/:lon", reverseGeoCode)

	http.ListenAndServe(":8080", router)
}