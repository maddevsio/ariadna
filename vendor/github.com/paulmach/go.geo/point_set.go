package geo

import (
	"bytes"
	"fmt"
	"math"

	"github.com/paulmach/go.geojson"
)

// A PointSet represents a set of points in the 2D Eucledian or Cartesian plane.
type PointSet []Point

// NewPointSet simply creates a new point set with points array of the given size.
func NewPointSet() *PointSet {
	return &PointSet{}
}

// NewPointSetPreallocate simply creates a new point set with points array of the given size.
func NewPointSetPreallocate(length, capacity int) *PointSet {
	if length > capacity {
		capacity = length
	}

	ps := make([]Point, length, capacity)
	p := PointSet(ps)
	return &p
}

// Clone returns a new copy of the point set.
func (ps PointSet) Clone() *PointSet {
	points := make([]Point, len(ps))
	copy(points, ps)

	nps := PointSet(points)
	return &nps
}

// Centroid returns the average x and y coordinate of the point set.
// This can also be used for small clusters of lat/lng points.
func (ps PointSet) Centroid() *Point {
	x := 0.0
	y := 0.0
	numPoints := float64(len(ps))
	for _, point := range ps {
		x += point[0]
		y += point[1]
	}
	return &Point{x / numPoints, y / numPoints}
}

// GeoCentroid uses a more advanced algorithm to compute the centroid of points
// on the earth's surface. The points are first projected into 3D space then
// averaged. The result is projected back onto the sphere. This method is about
// 6x slower than the Centroid function, which may be adequate for some datasets.
// NOTE: Points with longitude outside the standard -180:180 range will be remapped to
// within the range. The result will always have longitude between -180 and 180 degrees.
func (ps PointSet) GeoCentroid() *Point {

	// Implementation sourced from Geolib
	// https://github.com/manuelbieh/Geolib/blob/74593bf93f9a99d5ce7e6bcefa367c5a78f5321b/src/geolib.js#L416
	var x, y, z float64

	for _, p := range ps {
		lngSin, lngCos := math.Sincos(deg2rad(p[0]))
		latSin, latCos := math.Sincos(deg2rad(p[1]))

		x += latCos * lngCos
		y += latCos * lngSin
		z += latSin
	}

	np := float64(len(ps))
	x /= np
	y /= np
	z /= np

	return NewPoint(rad2deg(math.Atan2(y, x)), rad2deg(math.Atan2(z, math.Sqrt(x*x+y*y))))
}

// DistanceFrom returns the minimum euclidean distance from the point set.
func (ps PointSet) DistanceFrom(point *Point) (float64, int) {
	dist := math.Inf(1)
	index := 0

	for i := range ps {
		if d := ps[i].SquaredDistanceFrom(point); d < dist {
			dist = d
			index = i
		}
	}

	return math.Sqrt(dist), index
}

// GeoDistanceFrom returns the minimum geo distance from the point set,
// along with the index of the point with minimum index.
func (ps PointSet) GeoDistanceFrom(point *Point) (float64, int) {
	dist := math.Inf(1)
	index := 0

	for i := range ps {
		if d := ps[i].GeoDistanceFrom(point); d < dist {
			dist = d
			index = i
		}
	}

	return dist, index
}

// Bound returns a bound around the point set. Simply uses rectangular coordinates.
func (ps PointSet) Bound() *Bound {
	if len(ps) == 0 {
		return NewBound(0, 0, 0, 0)
	}

	minX := math.Inf(1)
	minY := math.Inf(1)

	maxX := math.Inf(-1)
	maxY := math.Inf(-1)

	for _, v := range ps {
		minX = math.Min(minX, v.X())
		minY = math.Min(minY, v.Y())

		maxX = math.Max(maxX, v.X())
		maxY = math.Max(maxY, v.Y())
	}

	return NewBound(maxX, minX, maxY, minY)
}

// SetAt updates a position at i in the point set
func (ps *PointSet) SetAt(index int, point *Point) *PointSet {
	deref := *ps
	if index >= len(deref) || index < 0 {
		panic(fmt.Sprintf("geo: set index out of range, requested: %d, length: %d", index, len(deref)))
	}
	deref[index] = *point
	*ps = deref
	return ps
}

// GetAt returns the pointer to the Point in the page.
// This function is good for modifying values in place.
// Returns nil if index is out of range.
func (ps *PointSet) GetAt(i int) *Point {
	deref := *ps
	if i >= len(deref) || i < 0 {
		return nil
	}

	return &deref[i]
}

// First returns the first point in the point set.
// Will return nil if there are no points in the set.
func (ps PointSet) First() *Point {
	if len(ps) == 0 {
		return nil
	}

	return &ps[0]
}

// Last returns the last point in the point set.
// Will return nil if there are no points in the set.
func (ps PointSet) Last() *Point {
	if len(ps) == 0 {
		return nil
	}

	return &ps[len(ps)-1]
}

// InsertAt inserts a Point at i in the point set.
// Panics if index is out of range.
func (ps *PointSet) InsertAt(index int, point *Point) *PointSet {
	deref := *ps
	if index > len(deref) || index < 0 {
		panic(fmt.Sprintf("geo: insert index out of range, requested: %d, length: %d", index, len(deref)))
	}

	if index == len(deref) {
		deref = append(deref, *point)
		*ps = deref
		return ps
	}

	deref = append(deref, Point{})
	copy(deref[index+1:], deref[index:])
	deref[index] = *point
	*ps = deref
	return ps
}

// RemoveAt removes a Point at i in the point set.
// Panics if index is out of range.
func (ps *PointSet) RemoveAt(index int) *PointSet {
	deref := *ps
	if index >= len(deref) || index < 0 {
		panic(fmt.Sprintf("geo: remove index out of range, requested: %d, length: %d", index, len(deref)))
	}

	deref = append(deref[:index], deref[index+1:]...)
	*ps = deref
	return ps
}

// Push appends a point to the end of the point set.
func (ps *PointSet) Push(point *Point) *PointSet {
	*ps = append(*ps, *point)
	return ps
}

// Pop removes and returns the last point in the point set
func (ps *PointSet) Pop() *Point {
	deref := *ps
	if len(deref) == 0 {
		return nil
	}

	x := deref[len(deref)-1]
	*ps = deref[:len(deref)-1]

	return &x
}

//SetPoints sets the points in the point set
func (ps *PointSet) SetPoints(points []Point) *PointSet {
	*ps = points
	return ps
}

// Length returns the number of points in the point set.
func (ps PointSet) Length() int {
	return len(ps)
}

// Equals compares two point sets. Returns true if lengths are the same
// and all points are Equal
func (ps PointSet) Equals(pointSet *PointSet) bool {
	if (ps).Length() != (*pointSet).Length() {
		return false
	}

	for i, v := range ps {
		if !v.Equals((*pointSet).GetAt(i)) {
			return false
		}
	}

	return true
}

// ToGeoJSON creates a new geojson feature with a multipoint geometry
// containing all the points.
func (ps PointSet) ToGeoJSON() *geojson.Feature {
	f := geojson.NewMultiPointFeature()
	for _, v := range ps {
		f.Geometry.MultiPoint = append(f.Geometry.MultiPoint, []float64{v[0], v[1]})
	}

	return f
}

// ToWKT returns the point set in WKT format,
// eg. MULTIPOINT(30 10, 10 30, 40 40)
func (ps PointSet) ToWKT() string {
	return ps.String()
}

// String returns a string representation of the path.
// The format is WKT, e.g. MULTIPOINT(30 10,10 30,40 40)
// For empty paths the result will be 'EMPTY'.
func (ps PointSet) String() string {
	if len(ps) == 0 {
		return "EMPTY"
	}

	buff := bytes.NewBuffer(nil)
	fmt.Fprintf(buff, "MULTIPOINT(%g %g", ps[0][0], ps[0][1])

	for i := 1; i < len(ps); i++ {
		fmt.Fprintf(buff, ",%g %g", ps[i][0], ps[i][1])
	}

	buff.Write([]byte(")"))
	return buff.String()
}
