package horizon

import (
	"math"
	"testing"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

func TestOpenFieldHorizon(t *testing.T) {
	point := geo.Point{Lat: 48.8566, Lng: 2.3522}
	buildings := []geo.Building{}

	profile := Compute(point, 1.5, buildings)

	for az := 0; az < 360; az++ {
		if profile.Horizon[az] != 0 {
			t.Errorf("azimuth %d: expected 0 horizon in open field, got %v", az, profile.Horizon[az])
			break
		}
	}
	if profile.BuildingCount != 0 {
		t.Errorf("expected 0 buildings, got %d", profile.BuildingCount)
	}
	if profile.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0 for empty field, got %v", profile.Confidence)
	}
}

func TestBoxedInHorizon(t *testing.T) {
	point := geo.Point{Lat: 48.8566, Lng: 2.3522}
	proj := geo.NewProj(point)

	cx, cy := proj.ToLocal(point.Lat, point.Lng)

	buildingHeight := 30.0
	observerH := 1.5

	// North building
	northPts := ptsFromLocal(proj, []struct{ x, y float64 }{
		{cx - 25, cy + 15},
		{cx + 25, cy + 15},
		{cx + 25, cy + 25},
		{cx - 25, cy + 25},
	})
	// South building
	southPts := ptsFromLocal(proj, []struct{ x, y float64 }{
		{cx - 25, cy - 25},
		{cx + 25, cy - 25},
		{cx + 25, cy - 15},
		{cx - 25, cy - 15},
	})
	// East building
	eastPts := ptsFromLocal(proj, []struct{ x, y float64 }{
		{cx + 15, cy - 25},
		{cx + 25, cy - 25},
		{cx + 25, cy + 25},
		{cx + 15, cy + 25},
	})
	// West building
	westPts := ptsFromLocal(proj, []struct{ x, y float64 }{
		{cx - 25, cy - 25},
		{cx - 15, cy - 25},
		{cx - 15, cy + 25},
		{cx - 25, cy + 25},
	})

	buildings := []geo.Building{
		{OSMID: 1, Footprint: polygonFromPts(northPts), Height: buildingHeight, HeightEstimated: true},
		{OSMID: 2, Footprint: polygonFromPts(southPts), Height: buildingHeight, HeightEstimated: true},
		{OSMID: 3, Footprint: polygonFromPts(eastPts), Height: buildingHeight, HeightEstimated: true},
		{OSMID: 4, Footprint: polygonFromPts(westPts), Height: buildingHeight, HeightEstimated: true},
	}

	profile := Compute(point, observerH, buildings)

	// Check all four cardinal directions have high obstruction
	for _, az := range []int{0, 90, 180, 270} {
		if profile.Horizon[az] < 20 {
			t.Errorf("azimuth %d: expected high obstruction, got %v", az, profile.Horizon[az])
		}
	}

	if profile.Confidence == 1.0 {
		t.Errorf("expected low confidence for all-estimated buildings, got %v", profile.Confidence)
	}
}

func TestObserverAboveBuildings(t *testing.T) {
	point := geo.Point{Lat: 48.8566, Lng: 2.3522}

	buildings := []geo.Building{
		{
			OSMID: 1,
			Footprint: polygonFromPts([]geo.Point{
				{Lat: 48.8570, Lng: 2.3525},
				{Lat: 48.8570, Lng: 2.3530},
				{Lat: 48.8565, Lng: 2.3530},
				{Lat: 48.8565, Lng: 2.3525},
			}),
			Height: 20,
		},
	}

	profile := Compute(point, 100, buildings)

	for az := 0; az < 360; az++ {
		if profile.Horizon[az] > 1 {
			t.Errorf("azimuth %d: observer above all buildings, expected horizon near 0, got %v", az, profile.Horizon[az])
			break
		}
	}
}

func TestSingleBuildingShadow(t *testing.T) {
	point := geo.Point{Lat: 48.8566, Lng: 2.3522}
	proj := geo.NewProj(point)

	ox, oy := proj.ToLocal(point.Lat, point.Lng)

	// Building 45-55m east, 20m tall, observer at 1.5m
	// West edge is at x=45m; ray at az=90 hits closest edge at x=45m
	buildingHeight := 20.0
	observerH := 1.5
	relHeight := buildingHeight - observerH
	closestDist := 45.0

	expectedAngle := math.Atan2(relHeight, closestDist) * 180 / math.Pi

	buildings := []geo.Building{
		{
			OSMID: 1,
			Footprint: polygonFromPts(ptsFromLocal(proj, []struct{ x, y float64 }{
				{ox + 45, oy - 10},
				{ox + 55, oy - 10},
				{ox + 55, oy + 10},
				{ox + 45, oy + 10},
			})),
			Height: buildingHeight,
		},
	}

	profile := Compute(point, observerH, buildings)

	if profile.Horizon[90] < expectedAngle-0.5 || profile.Horizon[90] > expectedAngle+0.5 {
		t.Errorf("expected horizon at east (90deg) ~%.2f, got %.2f", expectedAngle, profile.Horizon[90])
	}

	// North/south should have near-zero horizon
	if profile.Horizon[0] > 1 || profile.Horizon[180] > 1 {
		t.Errorf("north/south should be unobstructed")
	}
}

func TestComputeDataHashConsistency(t *testing.T) {
	b1 := []geo.Building{
		{OSMID: 1, Footprint: polygonFromPts([]geo.Point{{Lat: 0, Lng: 0}, {Lat: 1, Lng: 1}, {Lat: 0, Lng: 2}}), Height: 10},
	}
	b2 := []geo.Building{
		{OSMID: 1, Footprint: polygonFromPts([]geo.Point{{Lat: 0, Lng: 0}, {Lat: 1, Lng: 1}, {Lat: 0, Lng: 2}}), Height: 10},
	}

	h1 := computeDataHash(b1)
	h2 := computeDataHash(b2)
	if h1 != h2 {
		t.Error("expected identical hashes for identical building data")
	}
}

func ptsFromLocal(proj *geo.Proj, pts []struct{ x, y float64 }) []geo.Point {
	result := make([]geo.Point, len(pts))
	for i, p := range pts {
		lat, lng := proj.ToLatLng(p.x, p.y)
		result[i] = geo.Point{Lat: lat, Lng: lng}
	}
	return result
}

func polygonFromPts(pts []geo.Point) geo.Polygon {
	return geo.Polygon{Points: pts}
}
