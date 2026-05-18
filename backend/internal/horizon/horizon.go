package horizon

import (
	"crypto/sha256"
	"fmt"
	"math"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

type Profile struct {
	Horizon           [360]float64 `json:"horizon"`
	Lat               float64      `json:"lat"`
	Lng               float64      `json:"lng"`
	ObserverHeight    float64      `json:"observer_height"`
	Confidence        float64      `json:"confidence"`
	BuildingCount     int          `json:"building_count"`
	EstimatedCount    int          `json:"estimated_count"`
	BuildingDataHash  string       `json:"data_hash"`
	UseDSM            bool         `json:"use_dsm"`
}

func computeDataHash(buildings []geo.Building) string {
	h := sha256.New()
	for _, b := range buildings {
		for _, pt := range b.Footprint.Points {
			s := fmt.Sprintf("%.6f,%.6f,%.2f,%t;", pt.Lat, pt.Lng, b.Height, b.HeightEstimated)
			h.Write([]byte(s))
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func cacheKey(lat, lng, h float64, dataHash string) string {
	rLat := math.Round(lat*1000) / 1000
	rLng := math.Round(lng*1000) / 1000
	return fmt.Sprintf("%.3f_%.3f_%.1f_%s", rLat, rLng, h, dataHash[:16])
}
