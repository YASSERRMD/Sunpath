package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

type BuildingOutline struct {
	OSMID  int64       `json:"osm_id"`
	Height float64     `json:"height"`
	Shape  [][2]float64 `json:"shape"`
}

func (s *Server) handleBuildings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, 405, "method not allowed")
		return
	}

	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

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

	point := geo.Point{Lat: lat, Lng: lng}
	buildings, err := fetchBuildingsAround(point, s.cachedClient)
	if err != nil {
		log.Printf("fetching buildings: %v", err)
		s.writeError(w, 502, "failed to fetch building data")
		return
	}

	outlines := make([]BuildingOutline, 0, len(buildings))
	for _, b := range buildings {
		shape := make([][2]float64, len(b.Footprint.Points))
		for i, pt := range b.Footprint.Points {
			shape[i] = [2]float64{pt.Lng, pt.Lat}
		}
		outlines = append(outlines, BuildingOutline{
			OSMID:  b.OSMID,
			Height: b.Height,
			Shape:  shape,
		})
	}

	writeJSON(w, 200, envelope{Data: outlines})
}
