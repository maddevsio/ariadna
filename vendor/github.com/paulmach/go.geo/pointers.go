package geo

// A Pointer is the interface for something that has a point.
type Pointer interface {
	// Point should return the "center" or other canonical point
	// for the object. The caller is expected to Clone
	// the point if changes need to be make.
	Point() *Point
}

// TODO: add some functionality around sets of pointers,
// ie. `type PointerSlice []Pointer`
