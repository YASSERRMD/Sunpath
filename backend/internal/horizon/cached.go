package horizon

import (
	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

type CachedComputer struct {
	Store store.Storage
}

func NewCachedComputer(st store.Storage) *CachedComputer {
	return &CachedComputer{Store: st}
}

func (c *CachedComputer) Compute(point geo.Point, observerHeight float64, buildings []geo.Building) (Profile, error) {
	return c.computeCore(point, observerHeight, buildings, nil, nil)
}

func (c *CachedComputer) ComputeWithTerrain(point geo.Point, observerHeight float64, buildings []geo.Building, terrain *[360]float64) (Profile, error) {
	return c.computeCore(point, observerHeight, buildings, terrain, nil)
}

func (c *CachedComputer) ComputeWithVegetation(point geo.Point, observerHeight float64, buildings []geo.Building, terrain *[360]float64, vegHorizon *[360]float64) (Profile, error) {
	return c.computeCore(point, observerHeight, buildings, terrain, vegHorizon)
}

func (c *CachedComputer) computeCore(point geo.Point, observerHeight float64, buildings []geo.Building, terrain *[360]float64, vegHorizon *[360]float64) (Profile, error) {
	suffix := ""
	if terrain != nil {
		suffix = "_dsm"
	}
	if vegHorizon != nil {
		suffix += "_veg"
	}
	dataHash := ComputeDataHash(buildings) + suffix
	ck := CacheKey(point.Lat, point.Lng, observerHeight, dataHash)

	existing, err := c.Store.GetHorizonProfile(ck)
	if err != nil {
		return Profile{}, err
	}
	if existing != nil {
		return profileFromRecord(*existing), nil
	}

	var profile Profile
	if vegHorizon != nil {
		profile = ComputeWithVegetation(point, observerHeight, buildings, terrain, vegHorizon)
	} else if terrain != nil {
		profile = ComputeWithTerrain(point, observerHeight, buildings, terrain)
	} else {
		profile = Compute(point, observerHeight, buildings)
	}
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
		UseDSM:     p.UseDSM,
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
	p.UseDSM = r.UseDSM
	return p
}
