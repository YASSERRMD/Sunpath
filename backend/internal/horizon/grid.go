package horizon

import (
	"math"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
)

type cellKey struct {
	X, Y int
}

type spatialGrid struct {
	cells    map[cellKey][]gridEntry
	cellSize float64
	ox, oy   float64
}

type gridEntry struct {
	prism    geo.Prism
	relHeight float64
}

func newSpatialGrid(ox, oy float64, cellSize float64) *spatialGrid {
	return &spatialGrid{
		cells:    make(map[cellKey][]gridEntry),
		cellSize: cellSize,
		ox:       ox,
		oy:       oy,
	}
}

func (g *spatialGrid) add(prism geo.Prism, relHeight float64) {
	seen := make(map[cellKey]bool)
	for _, wall := range prism.Walls {
		minX := math.Min(wall.X1, wall.X2)
		maxX := math.Max(wall.X1, wall.X2)
		minY := math.Min(wall.Y1, wall.Y2)
		maxY := math.Max(wall.Y1, wall.Y2)

		cx1 := int(math.Floor(minX / g.cellSize))
		cx2 := int(math.Floor(maxX / g.cellSize))
		cy1 := int(math.Floor(minY / g.cellSize))
		cy2 := int(math.Floor(maxY / g.cellSize))

		for cx := cx1; cx <= cx2; cx++ {
			for cy := cy1; cy <= cy2; cy++ {
				ck := cellKey{cx, cy}
				if !seen[ck] {
					seen[ck] = true
					g.cells[ck] = append(g.cells[ck], gridEntry{prism: prism, relHeight: relHeight})
				}
			}
		}
	}
}

func (g *spatialGrid) query(az float64) []gridEntry {
	azRad := az * math.Pi / 180
	dx := math.Sin(azRad)
	dy := -math.Cos(azRad)

	maxDist := 2000.0
	steps := int(math.Ceil(maxDist / g.cellSize))

	var result []gridEntry
	seen := make(map[int]bool)

	for s := 0; s <= steps; s++ {
		cx := int(math.Floor((g.ox + dx*float64(s)*g.cellSize) / g.cellSize))
		cy := int(math.Floor((g.oy + dy*float64(s)*g.cellSize) / g.cellSize))

		ck := cellKey{cx, cy}
		if entries, ok := g.cells[ck]; ok {
			for _, e := range entries {
				if !seen[e.prism.Building.OSMID] {
					seen[e.prism.Building.OSMID] = true
					result = append(result, e)
				}
			}
		}
	}

	return result
}
