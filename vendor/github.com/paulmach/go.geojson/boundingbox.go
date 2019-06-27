package geojson

import "fmt"

func decodeBoundingBox(bb interface{}) ([]float64, error) {
	if bb == nil {
		return nil, nil
	}

	switch f := bb.(type) {
	case []float64:
		return f, nil
	case []interface{}:
		bb := make([]float64, 0, 4)
		for _, v := range f {
			switch c := v.(type) {
			case float64:
				bb = append(bb, c)
			default:
				return nil, fmt.Errorf("bounding box coordinate not usable, got %T", v)
			}

		}
		return bb, nil
	default:
		return nil, fmt.Errorf("bounding box property not usable, got %T", bb)
	}
}
