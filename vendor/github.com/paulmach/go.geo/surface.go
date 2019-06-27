package geo

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// Surface is the 2d version of path.
type Surface struct {
	bound         *Bound
	Width, Height int

	// represents the underlying data, as [x][y]
	// where x in [0:Width] and y in [0:Height]
	Grid [][]float64 // x,y
}

// NewSurface build and allocates all the memory to create
// a surface defined by the bound represented by width*height discrete points.
// Note that surface.Grid[width-1][height-1] will be on the boundary of the bound.
func NewSurface(bound *Bound, width, height int) *Surface {
	s := &Surface{
		bound:  bound.Clone(),
		Width:  width,
		Height: height,
	}

	s.Grid = make([][]float64, width)
	points := make([]float64, width*height)

	for i := range s.Grid {
		s.Grid[i], points = points[:height], points[height:]
	}

	return s
}

// Bound returns the same bound given at creation time.
func (s *Surface) Bound() *Bound {
	return s.bound
}

// PointAt returns the point, in the bound, corresponding to
// this grid coordinate. x in [0, s.Width()-1], y in [0, s.Height()-1]
func (s *Surface) PointAt(x, y int) *Point {
	if x >= s.Width || y >= s.Height {
		panic("geo: x, y outside of grid range")
	}

	p := NewPoint(0, 0)

	p[0] = s.bound.sw.X() + float64(x)*s.gridBoxWidth()
	p[1] = s.bound.sw.Y() + float64(y)*s.gridBoxHeight()

	return p
}

// ValueAt returns the bi-linearly interpolated value for
// the given point. Returns 0 if the point is out of surface bounds
// TODO: cleanup and optimize this code
func (s *Surface) ValueAt(point *Point) float64 {
	if !s.bound.Contains(point) {
		return 0
	}

	// find height and width
	xi, yi, w, h := s.gridCoordinate(point)

	xi1 := xi + 1
	if limit := s.Width - 1; xi1 > limit {
		xi1 = limit
	}

	yi1 := yi + 1
	if limit := s.Height - 1; yi1 > limit {
		yi1 = limit
	}

	w1 := s.Grid[xi][yi]*(1-w) + s.Grid[xi1][yi]*w
	w2 := s.Grid[xi][yi1]*(1-w) + s.Grid[xi1][yi1]*w

	return w1*(1-h) + w2*h
}

// GradientAt returns the surface gradient at the given point.
// Bilinearlly interpolates the grid cell to find the gradient.
func (s *Surface) GradientAt(point *Point) *Point {
	if !s.bound.Contains(point) {
		return NewPoint(0, 0)
	}

	xi, yi, deltaX, deltaY := s.gridCoordinate(point)

	xi1 := xi + 1
	if limit := s.Width - 1; xi1 > limit {
		xi = limit - 1
		xi1 = limit
		deltaX = 1.0
	}

	yi1 := yi + 1
	if limit := s.Height - 1; yi1 > limit {
		yi = limit - 1
		yi1 = limit
		deltaY = 1.0
	}

	u1 := s.Grid[xi][yi]*(1-deltaX) + s.Grid[xi1][yi]*deltaX
	u2 := s.Grid[xi][yi1]*(1-deltaX) + s.Grid[xi1][yi1]*deltaX

	w1 := (1 - deltaY) * (s.Grid[xi1][yi] - s.Grid[xi][yi])
	w2 := deltaY * (s.Grid[xi1][yi1] - s.Grid[xi][yi1])

	return NewPoint((w1+w2)/s.gridBoxWidth(), (u2-u1)/s.gridBoxHeight())
}

// WriteOffFile writes an Object File Format representation of
// the surface to the writer provided. This is for viewing
// in MeshLab or something like that. You should close the
// writer yourself after this function returns.
// http://segeval.cs.princeton.edu/public/off_format.html
func (s *Surface) WriteOffFile(w io.Writer) {
	var i, j int

	facesCount := 0
	var faces bytes.Buffer

	for i = 0; i < s.Width-1; i++ {
		for j := i % 2; j < s.Height-1; j += 2 {
			face := fmt.Sprintf("4 %d %d %d %d\n", i*s.Height+j, i*s.Height+j+1, (i+1)*s.Height+j+1, (i+1)*s.Height+j)
			faces.WriteString(face)
			facesCount++
		}
	}

	w.Write([]byte("OFF\n"))
	w.Write([]byte(fmt.Sprintf("%d %d 0\n", s.Height*s.Width, facesCount)))

	// vertexes
	for i = 0; i < s.Width; i++ {
		for j = 0; j < s.Height; j++ {
			p := s.PointAt(i, j)
			// weirdness is to things will be colored correctly in meshlab 1.3.2-OS X
			w.Write([]byte(fmt.Sprintf("%.8f %.8f %.8f\n", p[0], p[1], s.Grid[(s.Width-1)-i][j])))
		}
	}

	w.Write(faces.Bytes())
}

// gridBoxWidth returns the width of a grid element in the units of s.Bound.
func (s Surface) gridBoxWidth() float64 {
	return s.bound.Width() / float64(s.Width-1)
}

// gridBoxHeight returns the height of a grid element in the units of s.Bound.
func (s Surface) gridBoxHeight() float64 {
	return s.bound.Height() / float64(s.Height-1)
}

func (s Surface) gridCoordinate(point *Point) (x, y int, deltaX, deltaY float64) {
	w := (point[0] - s.bound.sw[0]) / s.gridBoxWidth()
	h := (point[1] - s.bound.sw[1]) / s.gridBoxHeight()

	x = int(math.Floor(w))
	y = int(math.Floor(h))

	deltaX = w - math.Floor(w)
	deltaY = h - math.Floor(h)

	return
}
