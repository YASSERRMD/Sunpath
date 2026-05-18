package osm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

const (
	defaultBuildingHeight = 8.0
	levelsPerFloor       = 3.2
	baseHeight           = 1.0
)

type ParseConfig struct {
	DefaultHeight float64
}

func DefaultConfig() ParseConfig {
	return ParseConfig{
		DefaultHeight: defaultBuildingHeight,
	}
}

func ParseBuildings(resp *OverpassResponse, cfg ParseConfig) []geo.Building {
	nodeMap := make(map[int64]OverpassElement)
	for _, el := range resp.Elements {
		if el.Type == "node" {
			nodeMap[el.ID] = el
		}
	}

	var buildings []geo.Building
	for _, el := range resp.Elements {
		if el.Type == "node" {
			continue
		}
		if _, ok := el.Tags["building"]; !ok {
			continue
		}

		b := geo.Building{
			OSMID:     el.ID,
			Footprint: geo.Polygon{},
		}

		resolveHeight(el.Tags, &b, cfg)

		switch el.Type {
		case "way":
			pts := make([]geo.Point, 0, len(el.Geometry))
			for _, gp := range el.Geometry {
				pts = append(pts, geo.Point{Lat: gp.Lat, Lng: gp.Lon})
			}
			b.Footprint.Points = pts

		case "relation":
			pts := make([]geo.Point, 0, len(el.Members))
			for _, m := range el.Members {
				if m.Type == "way" {
					if node, ok := nodeMap[m.Ref]; ok {
						pts = append(pts, geo.Point{Lat: node.Lat, Lng: node.Lon})
					}
				}
				if m.Lat != 0 || m.Lon != 0 {
					pts = append(pts, geo.Point{Lat: m.Lat, Lng: m.Lon})
				}
			}
			b.Footprint.Points = pts
		}

		if len(b.Footprint.Points) >= 3 {
			buildings = append(buildings, b)
		}
	}

	return buildings
}

func resolveHeight(tags map[string]string, b *geo.Building, cfg ParseConfig) {
	if hStr, ok := tags["height"]; ok {
		h, err := strconv.ParseFloat(hStr, 64)
		if err == nil && h > 0 {
			b.Height = h
			b.HeightEstimated = false
			return
		}
	}

	if lStr, ok := tags["building:levels"]; ok {
		levels, err := strconv.ParseFloat(lStr, 64)
		if err == nil && levels > 0 {
			if roofHeight, ok := tags["roof:height"]; ok {
				rh, err := strconv.ParseFloat(roofHeight, 64)
				if err == nil && rh > 0 {
					b.Height = levels*levelsPerFloor + rh
					b.HeightEstimated = true
					return
				}
			}
			b.Height = levels*levelsPerFloor + baseHeight
			b.HeightEstimated = true
			return
		}
	}

	b.Height = cfg.DefaultHeight
	b.HeightEstimated = true
}

func (c *OverpassClient) FetchAndParse(minLat, minLng, maxLat, maxLng float64, cfg ParseConfig) ([]geo.Building, error) {
	resp, err := c.FetchBuildings(minLat, minLng, maxLat, maxLng)
	if err != nil {
		return nil, err
	}
	return ParseBuildings(resp, cfg), nil
}

func BBoxKey(minLat, minLng, maxLat, maxLng float64) string {
	r := func(v float64) float64 { return math.Round(v*1000) / 1000 }
	return fmt.Sprintf("%.3f_%.3f_%.3f_%.3f", r(minLat), r(minLng), r(maxLat), r(maxLng))
}
