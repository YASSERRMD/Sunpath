package osm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OverpassClient struct {
	BaseURL      string
	RateLimited  *RateLimitedClient
	HTTPClient   *http.Client
	UserAgent    string
}

func NewClient(baseURL string) *OverpassClient {
	rl := NewRateLimitedClient(2)
	return &OverpassClient{
		BaseURL:     baseURL,
		RateLimited: rl,
		HTTPClient:  rl.Client,
		UserAgent:   "Sunpath/1.0 (solar analysis tool)",
	}
}

type OverpassResponse struct {
	Elements []OverpassElement `json:"elements"`
}

type OverpassElement struct {
	Type    string             `json:"type"`
	ID      int64              `json:"id"`
	Lat     float64            `json:"lat,omitempty"`
	Lon     float64            `json:"lon,omitempty"`
	Nodes   []int64            `json:"nodes,omitempty"`
	Members []OverpassMember   `json:"members,omitempty"`
	Tags    map[string]string  `json:"tags,omitempty"`
	Geometry []OverpassGeoPoint `json:"geometry,omitempty"`
	Bounds   *OverpassBounds   `json:"bounds,omitempty"`
}

type OverpassMember struct {
	Type string  `json:"type"`
	Ref  int64   `json:"ref"`
	Role string  `json:"role"`
	Lat  float64 `json:"lat,omitempty"`
	Lon  float64 `json:"lon,omitempty"`
}

type OverpassGeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type OverpassBounds struct {
	MinLat float64 `json:"minlat"`
	MaxLat float64 `json:"maxlat"`
	MinLon float64 `json:"minlon"`
	MaxLon float64 `json:"maxlon"`
}

func (c *OverpassClient) FetchBuildings(minLat, minLng, maxLat, maxLng float64) (*OverpassResponse, error) {
	query := fmt.Sprintf(`[out:json][timeout:45];
	(
		node["building"](%f,%f,%f,%f);
		way["building"](%f,%f,%f,%f);
		relation["building"](%f,%f,%f,%f);
	);
	out body geom;`, minLat, minLng, maxLat, maxLng, minLat, minLng, maxLat, maxLng, minLat, minLng, maxLat, maxLng)

	params := url.Values{}
	params.Set("data", query)
	reqURL := c.BaseURL

	resp, err := c.RateLimited.DoRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("overpass request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("overpass returned %d: %s", resp.StatusCode, bodyStr)
	}

	var result OverpassResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}
