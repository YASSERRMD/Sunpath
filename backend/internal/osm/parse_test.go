package osm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

func loadFixture(t *testing.T, name string) *OverpassResponse {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var resp OverpassResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatal(err)
	}
	return &resp
}

func TestParseBuildingsHeightTag(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{
				Type: "way",
				ID:   1,
				Tags: map[string]string{"building": "yes", "height": "12.5"},
				Geometry: []OverpassGeoPoint{
					{Lat: 48.8566, Lon: 2.3522},
					{Lat: 48.8567, Lon: 2.3523},
					{Lat: 48.8565, Lon: 2.3525},
					{Lat: 48.8566, Lon: 2.3522},
				},
			},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	if len(buildings) != 1 {
		t.Fatalf("expected 1 building, got %d", len(buildings))
	}
	if buildings[0].Height != 12.5 {
		t.Errorf("expected height 12.5 from height tag, got %v", buildings[0].Height)
	}
	if buildings[0].HeightEstimated {
		t.Error("expected heightEstimated=false for explicit height tag")
	}
}

func TestParseBuildingsLevelsTag(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{
				Type: "way",
				ID:   2,
				Tags: map[string]string{"building": "yes", "building:levels": "5"},
				Geometry: []OverpassGeoPoint{
					{Lat: 48.8566, Lon: 2.3522},
					{Lat: 48.8567, Lon: 2.3523},
					{Lat: 48.8565, Lon: 2.3525},
					{Lat: 48.8566, Lon: 2.3522},
				},
			},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	expected := 5*3.2 + 1.0
	if buildings[0].Height != expected {
		t.Errorf("expected height %v from levels, got %v", expected, buildings[0].Height)
	}
	if !buildings[0].HeightEstimated {
		t.Error("expected heightEstimated=true for levels-based height")
	}
}

func TestParseBuildingsDefaultHeight(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{
				Type: "way",
				ID:   3,
				Tags: map[string]string{"building": "yes"},
				Geometry: []OverpassGeoPoint{
					{Lat: 48.8566, Lon: 2.3522},
					{Lat: 48.8567, Lon: 2.3523},
					{Lat: 48.8565, Lon: 2.3525},
					{Lat: 48.8566, Lon: 2.3522},
				},
			},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	if buildings[0].Height != defaultBuildingHeight {
		t.Errorf("expected default height %v, got %v", defaultBuildingHeight, buildings[0].Height)
	}
	if !buildings[0].HeightEstimated {
		t.Error("expected heightEstimated=true for default height")
	}
}

func TestParseSkipsNonBuilding(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{
				Type: "way",
				ID:   4,
				Tags: map[string]string{"highway": "primary"},
				Geometry: []OverpassGeoPoint{
					{Lat: 48.8566, Lon: 2.3522},
					{Lat: 48.8567, Lon: 2.3523},
				},
			},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	if len(buildings) != 0 {
		t.Errorf("expected 0 non-building elements, got %d", len(buildings))
	}
}

func TestParseSkipsNodes(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{Type: "node", ID: 100, Lat: 48.8566, Lon: 2.3522, Tags: map[string]string{"building": "yes"}},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	if len(buildings) != 0 {
		t.Errorf("expected 0 node elements parsed as buildings, got %d", len(buildings))
	}
}

func TestParseSkipsUnder3Points(t *testing.T) {
	resp := &OverpassResponse{
		Elements: []OverpassElement{
			{
				Type: "way",
				ID:   5,
				Tags: map[string]string{"building": "yes"},
				Geometry: []OverpassGeoPoint{
					{Lat: 48.8566, Lon: 2.3522},
					{Lat: 48.8567, Lon: 2.3523},
				},
			},
		},
	}

	buildings := ParseBuildings(resp, DefaultConfig())
	if len(buildings) != 0 {
		t.Errorf("expected 0 buildings with <3 points, got %d", len(buildings))
	}
}

func TestRoundTripViaStore(t *testing.T) {
	buildings := []geo.Building{
		{
			OSMID:           1,
			Height:          10,
			HeightEstimated: false,
			Footprint: geo.Polygon{
				Points: []geo.Point{
					{Lat: 48.8566, Lng: 2.3522},
					{Lat: 48.8567, Lng: 2.3523},
					{Lat: 48.8565, Lng: 2.3525},
				},
			},
		},
	}

	records := buildingsToRecords(buildings)
	restored := recordsToBuildings(records)

	if len(restored) != 1 {
		t.Fatalf("expected 1 building, got %d", len(restored))
	}
	if restored[0].OSMID != 1 {
		t.Errorf("expected OSMID 1, got %d", restored[0].OSMID)
	}
	if restored[0].Height != 10 {
		t.Errorf("expected height 10, got %v", restored[0].Height)
	}
	if len(restored[0].Footprint.Points) != 3 {
		t.Errorf("expected 3 footprint points, got %d", len(restored[0].Footprint.Points))
	}
}

func TestBBoxKey(t *testing.T) {
	key := BBoxKey(48.8566, 2.3522, 48.8600, 2.3600)
	if key != "48.857_2.352_48.860_2.360" {
		t.Errorf("unexpected bbox key: %s", key)
	}
}

func TestFixtureFile(t *testing.T) {
	resp := loadFixture(t, "simple_building.json")
	buildings := ParseBuildings(resp, DefaultConfig())
	if len(buildings) != 2 {
		t.Fatalf("expected 2 buildings from fixture, got %d", len(buildings))
	}
	if buildings[0].Height != 15.0 {
		t.Errorf("fixture building 1: expected height 15, got %v", buildings[0].Height)
	}
	if buildings[0].HeightEstimated {
		t.Error("fixture building 1: expected heightEstimated=false")
	}
	if !buildings[1].HeightEstimated {
		t.Error("fixture building 2: expected heightEstimated=true from levels")
	}
}
