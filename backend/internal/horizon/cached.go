package horizon

import (
	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

type CachedComputer struct {
	Store *store.Store
}

func NewCachedComputer(st *store.Store) *CachedComputer {
	return &CachedComputer{Store: st}
}

func (c *CachedComputer) Compute(point geo.Point, observerHeight float64, buildings []geo.Building) (Profile, error) {
	dataHash := computeDataHash(buildings)
	ck := cacheKey(point.Lat, point.Lng, observerHeight, dataHash)

	existing, err := c.Store.GetHorizonProfile(ck)
	if err != nil {
		return Profile{}, err
	}
	if existing != nil {
		return profileFromRecord(*existing), nil
	}

	profile := Compute(point, observerHeight, buildings)
	profile.BuildingDataHash = dataHash

	rec := profileToRecord(profile)
	if err := c.Store.PutHorizonProfile(ck, &rec); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func profileToRecord(p Profile) store.HorizonRecord {
	return store.HorizonRecord{
		Lat:        p.Lat,
		Lng:        p.Lng,
		Height:     p.ObserverHeight,
		Horizon:    p.Horizon[:],
		Confidence: p.Confidence,
		BuildCount: p.BuildingCount,
		EstCount:   p.EstimatedCount,
		DataHash:   p.BuildingDataHash,
	}
}

func profileFromRecord(r store.HorizonRecord) Profile {
	var p Profile
	p.Lat = r.Lat
	p.Lng = r.Lng
	p.ObserverHeight = r.Height
	copy(p.Horizon[:], r.Horizon)
	p.Confidence = r.Confidence
	p.BuildingCount = r.BuildCount
	p.EstimatedCount = r.EstCount
	p.BuildingDataHash = r.DataHash
	return p
}
