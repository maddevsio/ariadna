go.geo
======

Go.geo is a geometry/geography library in [Go](http://golang.org). The primary use
case is GIS geometry manipulation on the server side vs. in the browser using javascript.
This may be motivated by memory, computation time or data privacy constraints.
All objects are defined in a 2D context.

#### Imports as package name `geo`:

```go
import "github.com/paulmach/go.geo"
```

<br />
[![Build Status](https://travis-ci.org/paulmach/go.geo.png?branch=master)](https://travis-ci.org/paulmach/go.geo)
&nbsp; &nbsp;
[![Coverage Status](https://coveralls.io/repos/paulmach/go.geo/badge.png?branch=master)](https://coveralls.io/r/paulmach/go.geo?branch=master)
&nbsp; &nbsp;
[![Godoc Reference](https://godoc.org/github.com/paulmach/go.geo?status.png)](https://godoc.org/github.com/paulmach/go.geo)

## Exposed objects
* **Point** represents a 2D location, x/y or lng/lat.
	It is up to the programmer to know if the data is a lng/lat location, projection of that point, or a vector.
	Useful features:
	* Project between WGS84 (EPSG:4326) and Mercator (EPSG:3857) or Scalar Mercator (map tiles). See examples below.
	* [GeoHash](https://godoc.org/github.com/paulmach/go.geo#Point.GeoHash) and [Quadkey](https://godoc.org/github.com/paulmach/go.geo#Point.Quadkey) support.
	* Supports vector functions like add, scale, etc. 
* **Line** represents the shortest distance between two points in Euclidean space.
	In many cases the path object is more useful.

* **PointSet** represents a set of points 
	with methods such as `DistanceFrom()` and `Centroid()`.

* **Path** is an extention of PointSet with methods for working with a polyline.
	Functions for converting to/from
	[Google's polyline encoding](https://developers.google.com/maps/documentation/utilities/polylinealgorithm) are included.
* **Bound** represents a rectangular 2D area defined by North, South, East, West values.
	Computable for Line and Path objects, used by the Surface object.
* **Surface** is used to assign values to points in a 2D area, such as elevation.

## Library conventions

There are two big conventions that developers should be aware of:
**functions are chainable** and **operations modify the original object.**
For example:

```go
p := geo.NewPoint(0, 0)
p.SetX(10).Add(geo.NewPoint(10, 10))
p.Equals(geo.NewPoint(20, 10))  // == true
```

If you want to create a copy, all objects support the `Clone()` method.

```go
p1 := geo.NewPoint(10, 10)
p2 := p1.SetY(20)
p1.Equals(p2) // == true, in this case p1 and p2 point to the same memory

p2 := p1.Clone().SetY(30)
p1.Equals(p2) // == false
```

These conventions put a little extra load on the programmer,
but tests showed that making a copy every time was significantly slower.
Similar conventions are found in the [math/big](https://golang.org/pkg/math/big/)
package of the Golang standard library.

### Databases, WKT and WKB

To make it easy to get and set data from spatial databases, all geometries support direct 
[scanning](https://golang.org/pkg/database/sql/#Scanner) of query results.
However, they must be retrieved in WKB format using functions such as
PostGIS' [ST_AsBinary](http://postgis.net/docs/ST_AsBinary.html).

For example, this query from a Postgres/PostGIS database:

```go
row := db.QueryRow("SELECT ST_AsBinary(point_column) FROM postgis_table")

var p *geo.Point
row.Scan(&p)
```

For MySQL, Geometry data is stored as SRID+WKB and the library detects and works with
this prefixed WKB data. So fetching spatial data from a MySQL database is even simpler:

```go
row := db.QueryRow("SELECT point_column FROM mysql_table")

var p *geo.Point
row.Scan(&p)
```

Inserts and updates can be made using the `.ToWKT()` methods. For example:

```go
db.Exec("INSERT INTO mysql_table (point_column) VALUES (GeomFromText(?))", p.ToWKT())
```

This has been tested using MySQL 5.5, MySQL 5.6 and PostGIS 2.0 using the
Point, LineString, MultiPoint and Polygon 2d spatial data types. 

### Reducers

The reducers sub-package includes implementations for 
[Douglas-Peucker](http://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm),
[Visvalingam](http://bost.ocks.org/mike/simplify/) 
and
[Radial](http://psimpl.sourceforge.net/radial-distance.html) 
polyline reduction algorithms. See the [reducers godoc](http://godoc.org/github.com/paulmach/go.geo/reducers) for more information.

### GeoJSON

All geometries support `.ToGeoJSON()` that return [*geojson.Feature](https://github.com/paulmach/go.geojson)
objects with the correct sub-geometry. For example:

```go
feature := path.ToGeoJSON()
feature.SetProperty("type", "road")

encodedJSON, _ := features.MarshalJSON()
```

## Examples

The [GoDoc Documentation](https://godoc.org/github.com/paulmach/go.geo) provides a very readable list
of exported functions. Below are a few usage examples.

### Projections

```go
lnglatPoint := geo.NewPoint(-122.4167, 37.7833)

// Mercator, EPSG:3857
mercator := geo.Mercator.Project(latlngPoint)
backToLnglat := geo.Mercator.Inverse(mercator)

// ScalarMercator or Google World Coordinates
tileX, TileY := geo.ScalarMercator.Project(latlngPoint.Lng(), latlngPoint.Lat())
tileZ := geo.ScalarMercator.Level

// level 16 tile the point is in
tileX >>= (geo.ScalarMercator.Level - 16)
tileY >>= (geo.ScalarMercator.Level - 16)
tileZ = 16
```

### Encode/Decode polyline path

```go
// lng/lat data, in this case, is encoded at 6 decimal place precision
path := geo.NewPathFromEncoding("smsqgAtkxvhFwf@{zCeZeYdh@{t@}BiAmu@sSqg@cjE", 1e6)

// reduce using the Douglas Peucker line reducer from the reducers sub-package.
// Note the threshold distance is in the coordinates of the points,
// which in this case is degrees.
reducedPath := reducers.DouglasPeucker(path, 1.0e-5)

// encode with the default/typical 5 decimal place precision
encodedString := reducedPath.Encode() 

// encode as json [[lng1,lat1],[lng2, lat2],...]
// using encoding/json from the standard library.
encodedJSON, err := json.Marshal(reducedPath)
```

### Path, line intersection

```go
path := geo.NewPath()
path.Push(geo.NewPoint(0, 0))
path.Push(geo.NewPoint(1, 1))

line := geo.NewLine(geo.NewPoint(0, 1), geo.NewPoint(1, 0))

// intersects does a simpler check for yes/no
if path.Intersects(line) {
	// intersection will return the actual points and places on intersection
	points, segments := path.Intersection(line)

	for i, _ := range points {
		log.Printf("Intersection %d at %v with path segment %d", i, points[i], segments[i][0])
	}
}
```

## Surface

A surface object is defined by a bound (lng/lat georegion for example) and a width and height 
defining the number of discrete points in the bound. This allows for access such as:

```go
surface.Grid[x][y]         // the value at a location in the grid
surface.GetPoint(x, y)     // the point, which will be in the space as surface.bound,
                           // corresponding to surface.Grid[x][y]
surface.ValueAt(*Point)    // the bi-linearly interpolated grid value for any point in the bounds
surface.GradientAt(*Point) // the gradient of the surface a any point in the bounds,
                           // returns a point object which should be treated as a vector
```

A couple things about how the bound area is discretized in the grid:
 
* `surface.Grid[0][0]`
	corresponds to the surface.Bound.SouthWest() location, or bottom left corner or the bound
* `surface.Grid[0][surface.Height-1]`
	corresponds to the surface.Bound.NorthWest() location,
	the extreme points in the grid are on the edges of the bound

While these conventions are useful, they are different.
If you're using this object, your feedback on these choices would be appreciated.

## Performance <a id="performance">&nbsp;</a>

This code is meant to act as a core library to more advanced geo algorithms, 
like [slide](https://github.com/paulmach/slide) for example.
Thus, performance is very important. Included are a good set of benchmarks covering
the core functions and efforts have been made to optimize them. Recent improvements:

```
                                      old         new        delta
BenchmarkPointDistanceFrom             8.16        5.91      -27.57%
BenchmarkPointSquaredDistanceFrom      1.63        1.62       -0.61%
BenchmarkPointQuadKey                271         265          -2.21%
BenchmarkPointQuadKeyString         2888         522         -81.93%
BenchmarkPointGeoHash                302         308           1.99%
BenchmarkPointGeoHashInt64           165         158          -4.24%
BenchmarkPointNormalize               22.3        17.6       -21.08%
BenchmarkPointEquals                   1.65        1.29      -21.82%
BenchmarkPointClone                    7.46        0.97      -87.00%

BenchmarkLineDistanceFrom             15.5        13.2       -14.84%
BenchmarkLineSquaredDistanceFrom       9.3         9.24       -0.65%
BenchmarkLineProject                   8.75        8.73       -0.23%
BenchmarkLineMeasure                  21.3        20          -6.10%
BenchmarkLineInterpolate              44.9        44.6        -0.67%
BenchmarkLineMidpoint                 47.2         5.13      -89.13%
BenchmarkLineEquals                    9.38       10.4        10.87%
BenchmarkLineClone                    70.5         3.26      -95.38%

BenchmarkPathDistanceFrom           6190        4662         -24.68%
BenchmarkPathSquaredDistanceFrom    5076        4625          -8.88%
BenchmarkPathMeasure               10080        7626         -24.35%
BenchmarkPathResampleToMorePoints  69380       17255         -75.13%
BenchmarkPathResampleToLessPoints  26093        6780         -74.02%
```

Units are Nanoseconds per Operation and run using Golang 1.3.1 on a 2012 Macbook Air with a 2GHz Intel Core i7 processor.
The old version corresponds to a commit on [Sept. 22, 2014](https://github.com/paulmach/go.geo/tree/984bc95cceb5e8fd7c3b8e9fdb0b2066207790e5) and the new version corresponds to a commit on [Sept 24, 2014](https://github.com/paulmach/go.geo/tree/9eb57f27bd88cdb2c1c96e058fafc74bb9aeaffb). These benchmarks can be run using: 

```
go get github.com/paulmach/go.geo
go test github.com/paulmach/go.geo -bench .
```

## Projects making use of this package

* [Slide: Vector to Raster Map Conflation](https://github.com/paulmach/slide)
* [SF MUNI Transit Delays, Visualized](http://bdon.org/transit)
* [osm-rune](https://github.com/tmcw/osm-rune) and [osm-rune-viewer](https://github.com/tmcw/osm-rune-viewer)
* Internally at [Strava](http://www.strava.com) for data analysis and the [Segment Compare Tool](http://blog.strava.com/whats-your-best-effort-see-how-it-compares-8480/)

## Contributing

While this project started as the core of [Slide](https://github.com/paulmach/slide) it's now being used in many place.
So, if you have features you'd like to add or improvements to make, please submit a pull request.
A big thank you to those who have contributed so far:

* [@bdon](https://github.com/bdon)
* [@ericrwolfe](https://github.com/ericrwolfe)
* [@mlerner](https://github.com/mlerner)
