package api

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/osm"
)

const maxRadius = 1000.0

func (s *Server) handleHorizon(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, 405, "method not allowed")
		return
	}

	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	hStr := r.URL.Query().Get("h")

	if latStr == "" || lngStr == "" {
		s.writeError(w, 400, "lat and lng are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		s.writeError(w, 400, "invalid lat")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		s.writeError(w, 400, "invalid lng")
		return
	}

	observerH := 1.5
	if hStr != "" {
		observerH, err = strconv.ParseFloat(hStr, 64)
		if err != nil || observerH < 0 {
			s.writeError(w, 400, "invalid h (observer height in metres)")
			return
		}
	}

	point := geo.Point{Lat: lat, Lng: lng}

	buildings, err := fetchBuildingsAround(point, s.cachedClient)
	if err != nil {
		log.Printf("fetching buildings: %v", err)
		s.writeError(w, 502, "failed to fetch building data")
		return
	}

	profile, err := s.horizonComp.Compute(point, observerH, buildings)
	if err != nil {
		log.Printf("computing horizon: %v", err)
		s.writeError(w, 500, "failed to compute horizon")
		return
	}

	writeJSON(w, 200, envelope{Data: profile})
}

func fetchBuildingsAround(point geo.Point, client *osm.CachedClient) ([]geo.Building, error) {
	dLat := maxRadius / 111320.0
	dLng := maxRadius / (111320.0 * cosDeg(point.Lat))

	return client.FetchBuildingsInBBox(
		point.Lat-dLat, point.Lng-dLng,
		point.Lat+dLat, point.Lng+dLng,
	)
}

func cosDeg(deg float64) float64 {
	return math.Cos(deg * math.Pi / 180)
}
