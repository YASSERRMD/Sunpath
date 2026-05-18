package api

import (
	"encoding/json"
	"net/http"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/horizon"
)

type batchPoint struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Height float64 `json:"height"`
	UseDSM bool    `json:"use_dsm"`
}

type batchResponseItem struct {
	Lat    float64          `json:"lat"`
	Lng    float64          `json:"lng"`
	Height float64          `json:"height"`
	Error  string           `json:"error,omitempty"`
	Data   *horizon.Profile `json:"data,omitempty"`
}

func (s *Server) handleBatchHorizon(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Points []batchPoint `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, 400, "invalid JSON body")
		return
	}
	if len(req.Points) == 0 {
		s.writeError(w, 400, "at least one point is required")
		return
	}
	if len(req.Points) > 20 {
		s.writeError(w, 400, "maximum 20 points per request")
		return
	}

	results := make([]batchResponseItem, 0, len(req.Points))
	for _, pt := range req.Points {
		item := batchResponseItem{
			Lat:    pt.Lat,
			Lng:    pt.Lng,
			Height: pt.Height,
		}

		p := geo.Point{Lat: pt.Lat, Lng: pt.Lng}
		buildings, err := s.cachedClient.FetchBuildingsInBBox(
			p.Lat-0.1, p.Lng-0.1, p.Lat+0.1, p.Lng+0.1,
		)
		if err != nil {
			item.Error = "failed to fetch buildings"
			results = append(results, item)
			continue
		}

		var profile horizon.Profile
		if pt.UseDSM {
			profile, err = s.horizonComp.ComputeWithTerrain(p, pt.Height, buildings, nil)
		} else {
			profile, err = s.horizonComp.Compute(p, pt.Height, buildings)
		}
		if err != nil {
			item.Error = "failed to compute horizon"
			results = append(results, item)
			continue
		}

		item.Data = &profile
		results = append(results, item)
	}

	writeJSON(w, 200, envelope{Data: results})
}
