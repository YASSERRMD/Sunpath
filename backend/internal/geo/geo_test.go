package geo

import (
	"math"
	"testing"
)

func TestProjectionRoundTrip(t *testing.T) {
	origin := Point{Lat: 48.8566, Lng: 2.3522}
	proj := NewProj(origin)

	pts := []Point{
		{Lat: 48.8566, Lng: 2.3522},
		{Lat: 48.8570, Lng: 2.3530},
		{Lat: 48.8550, Lng: 2.3500},
	}

	for _, pt := range pts {
		x, y := proj.ToLocal(pt.Lat, pt.Lng)
		lat2, lng2 := proj.ToLatLng(x, y)
		if math.Abs(lat2-pt.Lat) > 1e-6 || math.Abs(lng2-pt.Lng) > 1e-6 {
			t.Errorf("round trip failed for (%v, %v): got (%v, %v)", pt.Lat, pt.Lng, lat2, lng2)
		}
	}
}

func TestPointInPolygon(t *testing.T) {
	poly := Polygon{
		Points: []Point{
			{Lat: 0, Lng: 0},
			{Lat: 0, Lng: 2},
			{Lat: 2, Lng: 2},
			{Lat: 2, Lng: 0},
		},
	}

	if !poly.Contains(Point{Lat: 1, Lng: 1}) {
		t.Error("expected (1,1) to be inside square")
	}
	if poly.Contains(Point{Lat: 3, Lng: 3}) {
		t.Error("expected (3,3) to be outside square")
	}
	if poly.Contains(Point{Lat: -1, Lng: 0}) {
		t.Error("expected (-1,0) to be outside square")
	}
}

func TestCentroid(t *testing.T) {
	poly := Polygon{
		Points: []Point{
			{Lat: 0, Lng: 0},
			{Lat: 0, Lng: 4},
			{Lat: 4, Lng: 4},
			{Lat: 4, Lng: 0},
		},
	}
	c := poly.Centroid()
	if math.Abs(c.Lat-2) > 1e-6 || math.Abs(c.Lng-2) > 1e-6 {
		t.Errorf("centroid expected (2,2), got (%v,%v)", c.Lat, c.Lng)
	}
}

func TestBoundingBox(t *testing.T) {
	poly := Polygon{
		Points: []Point{
			{Lat: 1, Lng: 3},
			{Lat: 5, Lng: 7},
			{Lat: 2, Lng: 1},
		},
	}
	minLat, minLng, maxLat, maxLng := poly.BoundingBox()
	if math.Abs(minLat-1) > 1e-6 || math.Abs(minLng-1) > 1e-6 || math.Abs(maxLat-5) > 1e-6 || math.Abs(maxLng-7) > 1e-6 {
		t.Errorf("bbox expected (1,1,5,7), got (%v,%v,%v,%v)", minLat, minLng, maxLat, maxLng)
	}
}

func TestExtrusionVertexCount(t *testing.T) {
	b := Building{
		OSMID:     1,
		Height:    10,
		Footprint: Polygon{Points: []Point{{Lat: 48.8566, Lng: 2.3522}, {Lat: 48.8567, Lng: 2.3523}, {Lat: 48.8565, Lng: 2.3525}}},
	}
	origin := Point{Lat: 48.8566, Lng: 2.3522}
	prism := Extrude(b, origin)
	if len(prism.Walls) != 3 {
		t.Errorf("expected 3 walls for triangle, got %d", len(prism.Walls))
	}
	if len(prism.Roof.Points) != 3 {
		t.Errorf("expected 3 roof points for triangle, got %d", len(prism.Roof.Points))
	}
}

func TestRaySegmentIntersection(t *testing.T) {
	ox, oy := 0.0, 0.0
	seg := Segment{X1: 5, Y1: -1, X2: 5, Y2: 1}
	az := 90.0
	hit, dist := RaySegmentIntersect(ox, oy, az, seg)
	if !hit {
		t.Error("expected hit on vertical wall at x=5")
	}
	if math.Abs(dist-5) > 1e-6 {
		t.Errorf("expected distance 5, got %v", dist)
	}

	az2 := 0.0
	hit2, _ := RaySegmentIntersect(ox, oy, az2, seg)
	if hit2 {
		t.Error("expected miss when ray points north away from vertical wall")
	}

	az3 := 180.0
	seg3 := Segment{X1: -1, Y1: 5, X2: 1, Y2: 5}
	hit3, dist3 := RaySegmentIntersect(ox, oy, az3, seg3)
	if !hit3 {
		t.Error("expected hit on horizontal wall at y=5 when facing south")
	}
	if math.Abs(dist3-5) > 1e-6 {
		t.Errorf("expected distance 5, got %v", dist3)
	}
}
