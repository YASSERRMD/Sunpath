package osm

import (
	"encoding/json"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

type CachedClient struct {
	Overpass *OverpassClient
	Store    store.Storage
	Config   ParseConfig
}

func NewCachedClient(overpass *OverpassClient, st store.Storage, cfg ParseConfig) *CachedClient {
	return &CachedClient{
		Overpass: overpass,
		Store:    st,
		Config:   cfg,
	}
}

func (c *CachedClient) FetchBuildingsInBBox(minLat, minLng, maxLat, maxLng float64) ([]geo.Building, error) {
	key := BBoxKey(minLat, minLng, maxLat, maxLng)

	records, err := c.Store.GetOSMExtract(key)
	if err != nil {
		return nil, err
	}

	if records != nil {
		return recordsToBuildings(records), nil
	}

	resp, err := c.Overpass.FetchBuildings(minLat, minLng, maxLat, maxLng)
	if err != nil {
		return nil, err
	}

	buildings := ParseBuildings(resp, c.Config)
	records = buildingsToRecords(buildings)

	if err := c.Store.PutOSMExtract(key, records); err != nil {
		return nil, err
	}

	return buildings, nil
}

func recordsToBuildings(records []store.BuildingRecord) []geo.Building {
	buildings := make([]geo.Building, 0, len(records))
	for _, r := range records {
		var pts []geo.Point
		if err := json.Unmarshal([]byte(r.FootprintJSON), &pts); err != nil {
			continue
		}
		buildings = append(buildings, geo.Building{
			OSMID:           r.OSMID,
			Height:          r.Height,
			HeightEstimated: r.HeightEstimated,
			Footprint:       geo.Polygon{Points: pts},
		})
	}
	return buildings
}

func buildingsToRecords(buildings []geo.Building) []store.BuildingRecord {
	records := make([]store.BuildingRecord, 0, len(buildings))
	for _, b := range buildings {
		ptsJSON, _ := json.Marshal(b.Footprint.Points)
		records = append(records, store.BuildingRecord{
			OSMID:           b.OSMID,
			FootprintJSON:   string(ptsJSON),
			Height:          b.Height,
			HeightEstimated: b.HeightEstimated,
		})
	}
	return records
}
