package geo

type Segment struct {
	X1, Y1, X2, Y2 float64
}

type Prism struct {
	Building Building
	Walls    []Segment
	Roof     Polygon
}

func Extrude(b Building, origin Point) Prism {
	proj := NewProj(origin)
	localPts := make([]struct{ x, y float64 }, len(b.Footprint.Points))
	for i, pt := range b.Footprint.Points {
		x, y := proj.ToLocal(pt.Lat, pt.Lng)
		localPts[i] = struct{ x, y float64 }{x, y}
	}

	n := len(localPts)
	walls := make([]Segment, 0, n)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		walls = append(walls, Segment{
			X1: localPts[i].x,
			Y1: localPts[i].y,
			X2: localPts[j].x,
			Y2: localPts[j].y,
		})
	}

	roofPts := make([]Point, len(b.Footprint.Points))
	copy(roofPts, b.Footprint.Points)

	return Prism{
		Building: b,
		Walls:    walls,
		Roof:     Polygon{Points: roofPts},
	}
}
