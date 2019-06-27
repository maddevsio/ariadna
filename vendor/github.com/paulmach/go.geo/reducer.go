package geo

// A Reducer reduces a path using any simplification algorithm.
// It should return a copy of the path, not modify the original.
type Reducer interface {
	Reduce(*Path) *Path
}

// A GeoReducer reduces a path in EPSG:4326 (lng/lat) using any simplification algorithm.
// It should return a copy of the path, also in EPSG:4326, and not modify the original.
type GeoReducer interface {
	GeoReduce(*Path) *Path
}
