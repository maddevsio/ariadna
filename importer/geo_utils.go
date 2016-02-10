package importer

import (
	"github.com/kellydunn/golang-geo"
	"math"
	"strconv"
)

func computeCentroid(latlons []map[string]string) map[string]string {
	var points []geo.Point
	for _, each := range latlons {
		var lon, _ = strconv.ParseFloat(each["lon"], 64)
		var lat, _ = strconv.ParseFloat(each["lat"], 64)
		point := geo.NewPoint(lat, lon)
		points = append(points, *point)
	}

	var compute = getCentroid(points)

	var centroid = make(map[string]string)
	centroid["lat"] = strconv.FormatFloat(compute.Lat(), 'f', 16, 64)
	centroid["lon"] = strconv.FormatFloat(compute.Lng(), 'f', 16, 64)

	return centroid
}

// compute the centroid of a polygon set
// using a spherical co-ordinate system
func getCentroid(ps []geo.Point) *geo.Point {

	X := 0.0
	Y := 0.0
	Z := 0.0

	var toRad = math.Pi / 180
	var fromRad = 180 / math.Pi

	for _, point := range ps {

		var lon = point.Lng() * toRad
		var lat = point.Lat() * toRad

		X += math.Cos(lat) * math.Cos(lon)
		Y += math.Cos(lat) * math.Sin(lon)
		Z += math.Sin(lat)
	}

	numPoints := float64(len(ps))
	X = X / numPoints
	Y = Y / numPoints
	Z = Z / numPoints

	var lon = math.Atan2(Y, X)
	var hyp = math.Sqrt(X*X + Y*Y)
	var lat = math.Atan2(Z, hyp)

	var centroid = geo.NewPoint(lat*fromRad, lon*fromRad)

	return centroid
}
