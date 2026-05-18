package sun

import (
	"math"
	"testing"
	"time"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestSummerSolstice2025(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 6, 20, 12, 0, 0, 0, loc)
	lat := 48.8566
	lng := 2.3522

	_, el := solarPosition(tm, lat, lng)
	if !almostEqual(el, 64.6, 0.5) {
		t.Errorf("summer solstice elevation: expected ~64.6, got %v", el)
	}
}

func TestWinterSolstice2025(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 12, 21, 12, 0, 0, 0, loc)
	lat := 48.8566
	lng := 2.3522

	_, el := solarPosition(tm, lat, lng)
	if !almostEqual(el, 17.8, 0.5) {
		t.Errorf("winter solstice elevation: expected ~17.8, got %v", el)
	}
}

func TestSunsAltitudeAtNoon(t *testing.T) {
	// At solar noon, azimuth should be ~180 (south in northern hemisphere)
	loc := time.UTC
	lat := 48.8566
	lng := 2.3522

	// Solar noon at longitude 2.3522E is roughly 12:00 - 2.3522*4min = ~11:50:36 UTC
	tm := time.Date(2025, 6, 20, 11, 50, 36, 0, loc)
	az, el := solarPosition(tm, lat, lng)
	if !almostEqual(az, 180, 1.0) {
		t.Errorf("solar noon azimuth: expected ~180 (south), got %v", az)
	}
	if !almostEqual(el, 64.6, 0.5) {
		t.Errorf("solar noon elevation: expected ~64.6, got %v", el)
	}
}

func TestEquinox2025(t *testing.T) {
	loc := time.UTC
	lat := 0.0
	lng := 0.0

	// At the equator during equinox, sun should be near zenith at solar noon
	// Solar noon at 0,0 is close to 12:00 UTC
	tm := time.Date(2025, 3, 20, 12, 0, 0, 0, loc)
	_, el := solarPosition(tm, lat, lng)
	if el < 80 || el > 90 {
		t.Errorf("equinox noon at equator: expected near 90, got %v", el)
	}
}

func TestPolarDay(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 6, 20, 12, 0, 0, 0, loc)
	lat := 70.0
	lng := 0.0

	_, el := solarPosition(tm, lat, lng)
	if el < 10 {
		t.Errorf("polar day at lat 70: expected high sun, got elevation %v", el)
	}
}

func TestPolarNight(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 12, 21, 12, 0, 0, 0, loc)
	lat := 70.0
	lng := 0.0

	_, el := solarPosition(tm, lat, lng)
	if el > 0 {
		t.Errorf("polar night at lat 70: expected sun below horizon, got elevation %v", el)
	}
}

func TestAzimuthConvention(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 6, 20, 6, 0, 0, 0, loc)
	lat := 48.8566
	lng := 2.3522

	az, _ := solarPosition(tm, lat, lng)
	if az < 45 || az > 135 {
		t.Errorf("morning azimuth expected ~east (90), got %v", az)
	}
}

func TestSampleDayCount(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 6, 20, 0, 0, 0, 0, loc)
	samples := SampleDay(tm, 48.8566, 2.3522, 60)
	if len(samples) != 1440 {
		t.Errorf("expected 1440 1-minute samples, got %d", len(samples))
	}
}

func TestSunriseSunsetOrder(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2025, 6, 20, 0, 0, 0, 0, loc)
	day := ComputeDayTimes(tm, 48.8566, 2.3522)

	if !day.Sunrise.Before(day.Sunset) {
		t.Errorf("expected sunrise before sunset: sunrise=%v, sunset=%v", day.Sunrise, day.Sunset)
	}
	if !day.Sunrise.Before(day.SolarNoon) {
		t.Errorf("expected sunrise before solar noon: sunrise=%v, noon=%v", day.Sunrise, day.SolarNoon)
	}
	if !day.SolarNoon.Before(day.Sunset) {
		t.Errorf("expected solar noon before sunset: noon=%v, sunset=%v", day.SolarNoon, day.Sunset)
	}
}
