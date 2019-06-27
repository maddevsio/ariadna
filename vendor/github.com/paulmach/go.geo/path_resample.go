package geo

// Resample converts the path into totalPoints-1 evenly spaced segments.
// Assumes euclidean geometry.
func (p *Path) Resample(totalPoints int) *Path {
	if totalPoints <= 0 {
		p.PointSet = make([]Point, 0)
		return p
	}

	if p.resampleEdgeCases(totalPoints) {
		return p
	}

	// precomputes the total distance and intermediate distances
	total, dists := precomputeDistances(p.PointSet)
	p.resample(dists, total, totalPoints)
	return p
}

// ResampleWithInterval coverts the path into evenly spaced points of
// about the given distance. The total distance is computed using euclidean
// geometry and then divided by the given distance to get the number of segments.
func (p *Path) ResampleWithInterval(dist float64) *Path {
	if dist <= 0 {
		p.PointSet = make([]Point, 0)
		return p
	}

	// precomputes the total distance and intermediate distances
	total, dists := precomputeDistances(p.PointSet)

	totalPoints := int(total/dist) + 1
	if p.resampleEdgeCases(totalPoints) {
		return p
	}

	p.resample(dists, total, totalPoints)
	return p
}

// ResampleWithGeoInterval converts the path into about evenly spaced points of
// about the given distance. The total distance is computed using spherical (lng/lat) geometry
// and divided by the given distance. The new points are chosen by linearly interpolating
// between two given points. This may not make sense in some contexts, especially if
// the path covers a large range of latitude.
func (p *Path) ResampleWithGeoInterval(meters float64) *Path {
	if meters <= 0 {
		p.PointSet = make([]Point, 0)
		return p
	}

	// precomputes the total geo distance and intermediate distances
	totalDistance := 0.0
	distances := make([]float64, len(p.PointSet)-1)
	for i := 0; i < len(p.PointSet)-1; i++ {
		distances[i] = p.PointSet[i].GeoDistanceFrom(&p.PointSet[i+1])
		totalDistance += distances[i]
	}

	totalPoints := int(totalDistance/meters) + 1
	if p.resampleEdgeCases(totalPoints) {
		return p
	}

	p.resample(distances, totalDistance, totalPoints)
	return p
}

func (p *Path) resample(distances []float64, totalDistance float64, totalPoints int) {
	points := make([]Point, 1, totalPoints)
	points[0] = p.PointSet[0] // start stays the same

	step := 1
	distance := 0.0

	currentDistance := totalDistance / float64(totalPoints-1)
	currentLine := &Line{} // declare here and update has nice performance benefits
	for i := 0; i < len(p.PointSet)-1; i++ {
		currentLine.a = p.PointSet[i]
		currentLine.b = p.PointSet[i+1]

		currentLineDistance := distances[i]
		nextDistance := distance + currentLineDistance

		for currentDistance <= nextDistance {
			// need to add a point
			percent := (currentDistance - distance) / currentLineDistance
			points = append(points, Point{
				currentLine.a[0] + percent*(currentLine.b[0]-currentLine.a[0]),
				currentLine.a[1] + percent*(currentLine.b[1]-currentLine.a[1]),
			})

			// move to the next distance we want
			step++
			currentDistance = totalDistance * float64(step) / float64(totalPoints-1)
			if step == totalPoints-1 { // weird round off error on my machine
				currentDistance = totalDistance
			}
		}

		// past the current point in the original line, so move to the next one
		distance = nextDistance
	}

	// end stays the same, to handle round off errors
	if totalPoints != 1 { // for 1, we want the first point
		points[totalPoints-1] = p.PointSet[len(p.PointSet)-1]
	}

	(&p.PointSet).SetPoints(points)
	return
}

// resampleEdgeCases is used to handle edge case for
// resampling like not enough points and the path is all the same point.
// will return nil if there are no edge cases. If return true if
// one of these edge cases was found and handled.
func (p *Path) resampleEdgeCases(totalPoints int) bool {
	// degenerate case
	if len(p.PointSet) <= 1 {
		return true
	}

	// if all the points are the same, treat as special case.
	equal := true
	for _, point := range p.PointSet {
		if !p.PointSet[0].Equals(&point) {
			equal = false
			break
		}
	}

	if equal {
		if totalPoints > p.Length() {
			// extend to be requested length
			for p.Length() != totalPoints {
				p.PointSet = append(p.PointSet, p.PointSet[0])
			}

			return true
		}

		// contract to be requested length
		p.PointSet = p.PointSet[:totalPoints]
		return true
	}

	return false
}

// precomputeDistances precomputes the total distance and intermediate distances.
func precomputeDistances(p PointSet) (float64, []float64) {
	total := 0.0
	dists := make([]float64, len(p)-1)
	for i := 0; i < len(p)-1; i++ {
		dists[i] = p[i].DistanceFrom(&p[i+1])
		total += dists[i]
	}

	return total, dists
}
