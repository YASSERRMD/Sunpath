package geo

import "math"

type Point struct {
	Lat float64
	Lng float64
}

type Proj struct {
	Origin Point
}

func NewProj(origin Point) *Proj {
	return &Proj{Origin: origin}
}

func (p *Proj) ToLocal(lat, lng float64) (x, y float64) {
	dLat := (lat - p.Origin.Lat) * math.Pi / 180
	dLng := (lng - p.Origin.Lng) * math.Pi / 180
	latRad := p.Origin.Lat * math.Pi / 180
	x = dLng * math.Cos(latRad) * 6371000
	y = dLat * 6371000
	return
}

func (p *Proj) ToLatLng(x, y float64) (lat, lng float64) {
	latRad := p.Origin.Lat * math.Pi / 180
	lat = p.Origin.Lat + (y/6371000)*180/math.Pi
	lng = p.Origin.Lng + (x/(6371000*math.Cos(latRad)))*180/math.Pi
	return
}
