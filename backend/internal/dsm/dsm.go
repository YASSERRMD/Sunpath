package dsm

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

type ElevationClient struct {
	BaseURL   string
	Client    *http.Client
	maxAgeCache map[string]float64
}

type elevationResponse struct {
	Results []elevationResult `json:"results"`
}

type elevationResult struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`
}

func NewElevationClient(baseURL string) *ElevationClient {
	if baseURL == "" {
		baseURL = "https://api.open-elevation.com/api/v1/lookup"
	}
	return &ElevationClient{
		BaseURL:     baseURL,
		Client:      &http.Client{Timeout: 15 * time.Second},
		maxAgeCache: make(map[string]float64),
	}
}

func (c *ElevationClient) GetElevation(lat, lng float64) (float64, error) {
	key := fmt.Sprintf("%.5f_%.5f", lat, lng)
	if val, ok := c.maxAgeCache[key]; ok {
		return val, nil
	}

	params := url.Values{}
	params.Set("locations", fmt.Sprintf("%.5f,%.5f", lat, lng))
	reqURL := fmt.Sprintf("%s?%s", c.BaseURL, params.Encode())

	resp, err := c.Client.Get(reqURL)
	if err != nil {
		return 0, fmt.Errorf("elevation request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("reading elevation response: %w", err)
	}

	var result elevationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing elevation: %w", err)
	}

	if len(result.Results) == 0 {
		return 0, fmt.Errorf("no elevation data for %.5f, %.5f", lat, lng)
	}

	elev := result.Results[0].Elevation
	c.maxAgeCache[key] = elev
	return elev, nil
}

type TerrainHorizon struct {
	Horizon [360]float64
	Enabled bool
}

func ComputeTerrainHorizon(elevClient *ElevationClient, lat, lng, observerHeight, observerElev float64) (TerrainHorizon, error) {
	var th TerrainHorizon
	th.Enabled = true

	for az := 0; az < 360; az++ {
		azRad := float64(az) * math.Pi / 180
		maxAngle := 0.0

		for dist := 10.0; dist <= 500.0; dist += 20 {
			dLat := dist * math.Cos(azRad) / 111320
			dLng := dist * math.Sin(azRad) / (111320 * math.Cos(lat*math.Pi/180))

			sLat := lat + dLat
			sLng := lng + dLng

			elev, err := elevClient.GetElevation(sLat, sLng)
			if err != nil {
				continue
			}

			relElev := elev - observerElev - observerHeight
			angle := math.Atan2(relElev, dist) * 180 / math.Pi
			if angle > maxAngle {
				maxAngle = angle
			}
		}

		th.Horizon[az] = maxAngle
	}

	return th, nil
}
