package geo

import (
	"fmt"
	"math"
)

// A Projector is a function that converts the given point to a different space.
type Projector func(p *Point)

// A Projection is a set of projectors to map forward and backwards to the projected space.
type Projection struct {
	Project Projector
	Inverse Projector
}

const mercatorPole = 20037508.34

// Mercator projection, performs EPSG:3857, sometimes also described as EPSG:900913.
var Mercator = Projection{
	Project: func(p *Point) {
		p.SetX(mercatorPole / 180.0 * p.Lng())

		y := math.Log(math.Tan((90.0+p.Lat())*math.Pi/360.0)) / math.Pi * mercatorPole
		p.SetY(math.Max(-mercatorPole, math.Min(y, mercatorPole)))
	},
	Inverse: func(p *Point) {
		p.SetLng(p.X() * 180.0 / mercatorPole)
		p.SetLat(180.0 / math.Pi * (2*math.Atan(math.Exp((p.Y()/mercatorPole)*math.Pi)) - math.Pi/2.0))
	},
}

// MercatorScaleFactor returns the mercator scaling factor for a given degree latitude.
func MercatorScaleFactor(degreesLatitude float64) float64 {
	if degreesLatitude < -90.0 || degreesLatitude > 90.0 {
		panic(fmt.Sprintf("geo: latitude out of range, given %f", degreesLatitude))
	}

	return 1.0 / math.Cos(degreesLatitude/180.0*math.Pi)
}

// BuildTransverseMercator builds a transverse Mercator projection
// that automatically recenters the longitude around the provided centerLng.
// Works correctly around the anti-meridian.
// http://en.wikipedia.org/wiki/Transverse_Mercator_projection
func BuildTransverseMercator(centerLng float64) Projection {
	return Projection{
		Project: func(p *Point) {
			lng := p.Lng() - centerLng
			if lng < 180 {
				lng += 360.0
			}

			if lng > 180 {
				lng -= 360.0
			}

			p.SetLng(lng)
			TransverseMercator.Project(p)
		},
		Inverse: func(p *Point) {
			TransverseMercator.Inverse(p)

			lng := p.Lng() + centerLng
			if lng < 180 {
				lng += 360.0
			}

			if lng > 180 {
				lng -= 360.0
			}

			p.SetLng(lng)
		},
	}
}

// TransverseMercator implements a default transverse Mercator projector
// that will only work well +-10 degrees around longitude 0.
var TransverseMercator = Projection{
	Project: func(p *Point) {
		radLat := deg2rad(p.Lat())
		radLng := deg2rad(p.Lng())

		sincos := math.Sin(radLng) * math.Cos(radLat)
		p.SetX(0.5 * math.Log((1+sincos)/(1-sincos)) * EarthRadius)

		p.SetY(math.Atan(math.Tan(radLat)/math.Cos(radLng)) * EarthRadius)
	},
	Inverse: func(p *Point) {
		x := p.X() / EarthRadius
		y := p.Y() / EarthRadius

		lng := math.Atan(math.Sinh(x) / math.Cos(y))
		lat := math.Asin(math.Sin(y) / math.Cosh(x))

		p.SetLng(rad2deg(lng))
		p.SetLat(rad2deg(lat))
	},
}

// ScalarMercator converts from lng/lat float64 to x,y uint64.
// This is the same as Google's world coordinates.
var ScalarMercator struct {
	Level   uint64
	Project func(lng, lat float64, level ...uint64) (x, y uint64)
	Inverse func(x, y uint64, level ...uint64) (lng, lat float64)
}

func init() {
	ScalarMercator.Level = 31

	ScalarMercator.Project = func(lng, lat float64, level ...uint64) (x, y uint64) {
		l := ScalarMercator.Level
		if len(level) != 0 {
			l = level[0]
		}
		return scalarMercatorProject(lng, lat, l)
	}

	ScalarMercator.Inverse = func(x, y uint64, level ...uint64) (lng, lat float64) {
		l := ScalarMercator.Level
		if len(level) != 0 {
			l = level[0]
		}
		return scalarMercatorInverse(x, y, l)
	}
}

func scalarMercatorProject(lng, lat float64, level uint64) (x, y uint64) {
	var factor uint64

	factor = 1 << level
	maxtiles := float64(factor)

	lng = lng/360.0 + 0.5
	x = uint64(lng * maxtiles)

	// bound it because we have a top of the world problem
	siny := math.Sin(lat * math.Pi / 180.0)

	if siny < -0.9999 {
		lat = 0.5 + 0.5*math.Log((1.0+siny)/(1.0-siny))/(-2*math.Pi)
		y = 0
	} else if siny > 0.9999 {
		lat = 0.5 + 0.5*math.Log((1.0+siny)/(1.0-siny))/(-2*math.Pi)
		y = factor - 1
	} else {
		lat = 0.5 + 0.5*math.Log((1.0+siny)/(1.0-siny))/(-2*math.Pi)
		y = uint64(lat * maxtiles)
	}

	return
}

func scalarMercatorInverse(x, y, level uint64) (lng, lat float64) {
	var factor uint64

	factor = 1 << level
	maxtiles := float64(factor)

	lng = 360.0 * (float64(x)/maxtiles - 0.5)
	lat = (2.0*math.Atan(math.Exp(math.Pi-(2*math.Pi)*(float64(y))/maxtiles)))*(180.0/math.Pi) - 90.0

	return
}
