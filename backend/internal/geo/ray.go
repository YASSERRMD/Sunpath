package geo

import (
	"math"
)

func RaySegmentIntersect(ox, oy float64, azimuthDeg float64, seg Segment) (hit bool, dist float64) {
	azRad := azimuthDeg * math.Pi / 180
	dx := math.Sin(azRad)
	dy := -math.Cos(azRad)

	sx := seg.X2 - seg.X1
	sy := seg.Y2 - seg.Y1

	denom := dx*sy - dy*sx
	if math.Abs(denom) < 1e-12 {
		return false, 0
	}

	t := ((seg.X1-ox)*sy - (seg.Y1-oy)*sx) / denom
	u := ((seg.X1-ox)*dy - (seg.Y1-oy)*dx) / denom

	if t >= 0 && u >= 0 && u <= 1 {
		ix := ox + t*dx
		iy := oy + t*dy
		dist = math.Sqrt((ix-ox)*(ix-ox) + (iy-oy)*(iy-oy))
		return true, dist
	}

	return false, 0
}
