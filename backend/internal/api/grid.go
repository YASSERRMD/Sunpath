package api

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/horizon"
	"github.com/yasserrmd/sunpath/backend/internal/sun"
)

type GridCell struct {
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	SunMinutes int     `json:"sun_minutes"`
}

func (s *Server) handleGrid(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	q := r.URL.Query()
	lat1 := parseFloat(q.Get("lat1"))
	lng1 := parseFloat(q.Get("lng1"))
	lat2 := parseFloat(q.Get("lat2"))
	lng2 := parseFloat(q.Get("lng2"))
	observerH := parseFloatWithDefault(q.Get("h"), 1.5)
	resolution := int(parseFloatWithDefault(q.Get("res"), 10))
	if resolution < 5 {
		resolution = 5
	}
	if resolution > 100 {
		resolution = 100
	}

	if math.IsNaN(lat1) || math.IsNaN(lng1) || math.IsNaN(lat2) || math.IsNaN(lng2) {
		writeError(w, 400, "lat1, lng1, lat2, lng2 are required")
		return
	}

	minLat := math.Min(lat1, lat2)
	maxLat := math.Max(lat1, lat2)
	minLng := math.Min(lng1, lng2)
	maxLng := math.Max(lng1, lng2)

	nCols := int(math.Ceil((maxLat - minLat) / float64(resolution) * 111320))
	nRows := int(math.Ceil((maxLng - minLng) / float64(resolution) * 111320))
	if nCols < 1 {
		nCols = 1
	}
	if nRows < 1 {
		nRows = 1
	}
	if nCols*nRows > 400 {
		nCols = 20
		nRows = 20
	}

	buildings, err := fetchBuildingsAround(geo.Point{Lat: (minLat + maxLat) / 2, Lng: (minLng + maxLng) / 2}, s.cachedClient)
	if err != nil {
		log.Printf("fetching buildings for grid: %v", err)
		writeError(w, 502, "failed to fetch building data")
		return
	}

	refDate := time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC)

	var cells []GridCell

	for i := 0; i < nCols; i++ {
		for j := 0; j < nRows; j++ {
			lat := minLat + (float64(i)+0.5)/float64(nCols)*(maxLat-minLat)
			lng := minLng + (float64(j)+0.5)/float64(nRows)*(maxLng-minLng)
			point := geo.Point{Lat: lat, Lng: lng}

			profile, err := s.horizonComp.Compute(point, observerH, buildings)
			if err != nil {
				continue
			}

			minutes := computeSunMinutes(refDate, lat, lng, profile)
			cells = append(cells, GridCell{Lat: lat, Lng: lng, SunMinutes: minutes})
		}
	}

	writeJSON(w, 200, envelope{Data: cells})
}

func computeSunMinutes(date time.Time, lat, lng float64, profile horizon.Profile) int {
	total := 0
	end := date.Add(24 * time.Hour)
	for cursor := date; cursor.Before(end); cursor = cursor.Add(time.Minute) {
		az, el := sun.SolarPosition(cursor, lat, lng)
		if el <= 0 {
			continue
		}
		azIdx := int(math.Round(az)) % 360
		if azIdx < 0 {
			azIdx += 360
		}
		if el > profile.Horizon[azIdx] {
			total++
		}
	}
	return total
}

func parseFloat(s string) float64 {
	if s == "" {
		return math.NaN()
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return v
}

func parseFloatWithDefault(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
