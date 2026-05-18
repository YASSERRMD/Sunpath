package geo

import "math"

type Polygon struct {
	Points []Point
}

func (p Polygon) BoundingBox() (minLat, minLng, maxLat, maxLng float64) {
	minLat = math.MaxFloat64
	minLng = math.MaxFloat64
	maxLat = -math.MaxFloat64
	maxLng = -math.MaxFloat64
	for _, pt := range p.Points {
		if pt.Lat < minLat {
			minLat = pt.Lat
		}
		if pt.Lng < minLng {
			minLng = pt.Lng
		}
		if pt.Lat > maxLat {
			maxLat = pt.Lat
		}
		if pt.Lng > maxLng {
			maxLng = pt.Lng
		}
	}
	return
}

func (p Polygon) Area() float64 {
	n := len(p.Points)
	if n < 3 {
		return 0
	}
	var area float64
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += p.Points[i].Lat * p.Points[j].Lng
		area -= p.Points[j].Lat * p.Points[i].Lng
	}
	return math.Abs(area) / 2
}

func (p Polygon) Centroid() Point {
	n := len(p.Points)
	if n == 0 {
		return Point{}
	}
	var latSum, lngSum float64
	for _, pt := range p.Points {
		latSum += pt.Lat
		lngSum += pt.Lng
	}
	return Point{Lat: latSum / float64(n), Lng: lngSum / float64(n)}
}

func (p Polygon) Contains(point Point) bool {
	n := len(p.Points)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		if ((p.Points[i].Lng > point.Lng) != (p.Points[j].Lng > point.Lng)) &&
			(point.Lat < (p.Points[j].Lat-p.Points[i].Lat)*(point.Lng-p.Points[i].Lng)/(p.Points[j].Lng-p.Points[i].Lng)+p.Points[i].Lat) {
			inside = !inside
		}
		j = i
	}
	return inside
}
