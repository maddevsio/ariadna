package web

import (
	"gopkg.in/olivere/elastic.v3"
)

func geoCoderQuery(query string) *elastic.QueryStringQuery {
	qs := elastic.NewQueryStringQuery(query)
	qs.Field("name")
	qs.Field("street")
	qs.Field("housenumber")
	qs.Field("district")
	qs.Field("old_name")
	qs.Field("town")
	qs.Field("city")
	qs.Analyzer("map_synonyms")
	return qs
}

func reverseGeoCodeQuery(lat, lon float64) *elastic.GeoDistanceQuery {
	qs := elastic.NewGeoDistanceQuery("centroid")
	qs.GeoPoint(elastic.GeoPointFromLatLon(lat, lon))
	qs.Distance("10m")
	qs.QueryName("filtered")
	return qs
}
