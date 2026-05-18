package horizon

import (
	"math"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

func Compute(point geo.Point, observerHeight float64, buildings []geo.Building) Profile {
	return computeCore(point, observerHeight, buildings, nil)
}

func ComputeWithTerrain(point geo.Point, observerHeight float64, buildings []geo.Building, terrain *[360]float64) Profile {
	return computeCore(point, observerHeight, buildings, terrain)
}

func computeCore(point geo.Point, observerHeight float64, buildings []geo.Building, terrain *[360]float64) Profile {
	var p Profile
	p.Lat = point.Lat
	p.Lng = point.Lng
	p.ObserverHeight = observerHeight
	p.UseDSM = terrain != nil

	if len(buildings) == 0 && terrain == nil {
		for az := 0; az < 360; az++ {
			p.Horizon[az] = 0
		}
		p.Confidence = 1.0
		p.BuildingCount = 0
		p.EstimatedCount = 0
		p.BuildingDataHash = ComputeDataHash(buildings)
		return p
	}

	proj := geo.NewProj(point)
	ox, oy := proj.ToLocal(point.Lat, point.Lng)

	totalBuildings := 0
	estimatedBuildings := 0
	grid := newSpatialGrid(ox, oy, 50.0)

	for _, b := range buildings {
		if len(b.Footprint.Points) < 3 {
			continue
		}
		prism := geo.Extrude(b, point)
		totalBuildings++
		if b.HeightEstimated {
			estimatedBuildings++
		}
		relHeight := b.Height - observerHeight
		if relHeight > 0 {
			grid.add(prism, relHeight)
		}
	}

	p.BuildingCount = totalBuildings
	p.EstimatedCount = estimatedBuildings
	if totalBuildings > 0 {
		p.Confidence = 1.0 - float64(estimatedBuildings)/float64(totalBuildings)
	}

	for az := 0; az < 360; az++ {
		azFloat := float64(az)
		maxAngle := 0.0

		if terrain != nil && terrain[az] > maxAngle {
			maxAngle = terrain[az]
		}

		entries := grid.query(azFloat)
		for _, e := range entries {
			for _, wall := range e.prism.Walls {
				hit, dist := geo.RaySegmentIntersect(ox, oy, azFloat, wall)
				if hit && dist > 0 {
					angle := math.Atan2(e.relHeight, dist) * 180 / math.Pi
					if angle > maxAngle {
						maxAngle = angle
					}
				}
			}
		}

		p.Horizon[az] = maxAngle
	}

	p.BuildingDataHash = ComputeDataHash(buildings)
	return p
}
