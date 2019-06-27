package geo

import (
	"encoding/json"
	"errors"
)

// MarshalJSON enables lines to be encoded as JSON using the encoding/json package.
func (l *Line) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]Point{l.a, l.b})
}

// UnmarshalJSON enables lines to be decoded as JSON using the encoding/json package.
func (l *Line) UnmarshalJSON(data []byte) error {
	var points [][2]float64

	err := json.Unmarshal(data, &points)
	if err != nil {
		return err
	}

	if len(points) > 2 {
		return errors.New("geo: too many points to unmarshal into line")
	}

	if len(points) < 2 {
		return errors.New("geo: not enough points to unmarshal into line")
	}

	l.a = Point{points[0][0], points[0][1]}
	l.b = Point{points[1][0], points[1][1]}

	return nil
}

// MarshalJSON enables paths to be encoded as JSON using the encoding/json package.
func (p *Path) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.PointSet)
}

// UnmarshalJSON enables paths to be decoded as JSON using the encoding/json package.
func (p *Path) UnmarshalJSON(data []byte) error {
	var pointSet PointSet

	err := json.Unmarshal(data, &pointSet)
	if err != nil {
		return err
	}
	p.PointSet = pointSet
	return nil
}

// MarshalJSON enables bounds to be encoded as JSON using the encoding/json package.
func (b *Bound) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]*Point{b.sw, b.ne})
}

// UnmarshalJSON enables bounds to be decoded as JSON using the encoding/json package.
func (b *Bound) UnmarshalJSON(data []byte) error {
	var points []*Point

	err := json.Unmarshal(data, &points)
	if err != nil {
		return err
	}

	if len(points) > 2 {
		return errors.New("geo: too many points to unmarshal into bound")
	}

	if len(points) < 2 {
		return errors.New("geo: not enough points to unmarshal into bound")
	}

	b.sw = points[0]
	b.ne = points[0].Clone()
	b.Extend(points[1])

	return nil
}

// MarshalJSON enables surfaces to be encoded as JSON using the encoding/json package.
func (s *Surface) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"bound":  s.bound,
		"values": s.Grid,
	})
}

// UnmarshalJSON enables surfaces to be decoded as JSON using the encoding/json package.
func (s *Surface) UnmarshalJSON(data []byte) error {
	surface := struct {
		Bound  *Bound
		Values [][]float64
	}{}

	err := json.Unmarshal(data, &surface)
	if err != nil {
		return err
	}

	// if len(points) > 2 {
	// 	return errors.New("geo: too many points to unmarshal into bound")
	// }

	// if len(points) < 2 {
	// 	return errors.New("geo: not enough points to unmarshal into bound")
	// }

	s.bound = surface.Bound
	s.Grid = surface.Values

	return nil
}
