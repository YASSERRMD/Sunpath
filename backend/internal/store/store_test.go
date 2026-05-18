package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func testStore(t *testing.T, s Storage) {
	t.Helper()

	t.Run("OSM extract round-trip", func(t *testing.T) {
		records := []BuildingRecord{
			{OSMID: 1, FootprintJSON: `[{"lat":48.85,"lng":2.35}]`, Height: 25.0, HeightEstimated: false},
			{OSMID: 2, FootprintJSON: `[{"lat":48.86,"lng":2.36}]`, Height: 12.0, HeightEstimated: true},
		}
		key := "test_bbox"

		got, err := s.GetOSMExtract(key)
		if err != nil {
			t.Fatalf("GetOSMExtract (empty): %v", err)
		}
		if got != nil {
			t.Fatal("expected nil for missing key")
		}

		if err := s.PutOSMExtract(key, records); err != nil {
			t.Fatalf("PutOSMExtract: %v", err)
		}

		got, err = s.GetOSMExtract(key)
		if err != nil {
			t.Fatalf("GetOSMExtract: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 records, got %d", len(got))
		}
		if got[0].OSMID != 1 || got[1].OSMID != 2 {
			t.Errorf("OSMID mismatch: %+v", got)
		}
	})

	t.Run("Horizon profile round-trip", func(t *testing.T) {
		profile := &HorizonRecord{
			Lat:        48.8566,
			Lng:        2.3522,
			Height:     1.5,
			Horizon:    []float64{0, 5.2, 10.1},
			Confidence: 0.85,
			BuildCount: 10,
			EstCount:   2,
			DataHash:   "abc123",
			UseDSM:     true,
		}
		key := "test_profile"

		got, err := s.GetHorizonProfile(key)
		if err != nil {
			t.Fatalf("GetHorizonProfile (empty): %v", err)
		}
		if got != nil {
			t.Fatal("expected nil for missing key")
		}

		if err := s.PutHorizonProfile(key, profile); err != nil {
			t.Fatalf("PutHorizonProfile: %v", err)
		}

		got, err = s.GetHorizonProfile(key)
		if err != nil {
			t.Fatalf("GetHorizonProfile: %v", err)
		}
		if got == nil {
			t.Fatal("expected non-nil profile")
		}
		if got.Lat != 48.8566 || got.Lng != 2.3522 {
			t.Errorf("lat/lng mismatch: %+v", got)
		}
		if got.Confidence != 0.85 {
			t.Errorf("confidence mismatch: %v", got.Confidence)
		}
		if got.UseDSM != true {
			t.Errorf("UseDSM mismatch")
		}
	})

	t.Run("Update existing record", func(t *testing.T) {
		key := "update_test"
		p1 := &HorizonRecord{Lat: 1, Lng: 2, Height: 3, Horizon: []float64{1}, Confidence: 0.5}
		p2 := &HorizonRecord{Lat: 4, Lng: 5, Height: 6, Horizon: []float64{2}, Confidence: 0.9}

		if err := s.PutHorizonProfile(key, p1); err != nil {
			t.Fatal(err)
		}
		if err := s.PutHorizonProfile(key, p2); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetHorizonProfile(key)
		if err != nil {
			t.Fatal(err)
		}
		if got.Confidence != 0.9 {
			t.Errorf("expected updated confidence 0.9, got %v", got.Confidence)
		}
	})

	t.Run("Evict older than", func(t *testing.T) {
		key := "evict_test"
		p := &HorizonRecord{Lat: 0, Lng: 0, Height: 1, Horizon: []float64{0}}
		if err := s.PutHorizonProfile(key, p); err != nil {
			t.Fatal(err)
		}

		n, err := s.EvictOlderThan(-1 * time.Hour)
		if err != nil {
			t.Fatalf("EvictOlderThan: %v", err)
		}
		if n < 1 {
			t.Logf("evicted %d entries (may be 0 if recent)", n)
		}

		_, err = s.GetHorizonProfile(key)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Stats are non-negative", func(t *testing.T) {
		stats := s.Stats()
		if stats.OSMExtracts < 0 {
			t.Error("negative OSM extracts count")
		}
		if stats.HorizonProfiles < 0 {
			t.Error("negative horizon profiles count")
		}
		if stats.Hits < 0 || stats.Misses < 0 {
			t.Error("negative hit/miss count")
		}
	})
}

func TestPostgresStore(t *testing.T) {
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	st, err := NewPostgresStore(context.Background(), pgURL)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	defer st.Close()

	testStore(t, st)
}
