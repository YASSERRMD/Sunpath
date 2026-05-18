package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GeocodeResult struct {
	DisplayName string  `json:"display_name"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	Type        string  `json:"type"`
	Importance  float64 `json:"importance"`
}

func (s *Server) handleGeocode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, 400, "q (query) is required")
		return
	}

	params := url.Values{}
	params.Set("q", q)
	params.Set("format", "json")
	params.Set("limit", "10")
	params.Set("addressdetails", "0")

	nominatimURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?%s", params.Encode())

	req, err := http.NewRequest("GET", nominatimURL, nil)
	if err != nil {
		writeError(w, 500, "failed to create geocode request")
		return
	}
	req.Header.Set("User-Agent", "Sunpath/1.0 (solar analysis tool)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeError(w, 502, "geocode service unavailable")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, 502, "failed to read geocode response")
		return
	}

	var results []GeocodeResult
	if err := json.Unmarshal(body, &results); err != nil {
		writeError(w, 502, "failed to parse geocode response")
		return
	}

	if results == nil {
		results = []GeocodeResult{}
	}

	writeJSON(w, 200, envelope{Data: results})
}
