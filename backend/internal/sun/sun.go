package sun

import (
	"math"
	"time"
)

const earthTilt = 23.4397 * math.Pi / 180

func julianDate(t time.Time) float64 {
	return float64(t.Unix())/86400.0 + 2440587.5
}

func solarPosition(t time.Time, lat, lng float64) (azimuth, elevation float64) {
	jd := julianDate(t)
	n := jd - 2451545.0
	L := (280.460 + 0.9856474*n) * math.Pi / 180
	g := (357.528 + 0.9856003*n) * math.Pi / 180
	lambda := L + (1.915*math.Sin(g)+0.020*math.Sin(2*g)) * math.Pi / 180
	epsilon := earthTilt
	sinDecl := math.Sin(epsilon) * math.Sin(lambda)
	decl := math.Asin(sinDecl)

	hourAngle := ((jd-0.5-math.Floor(jd-0.5))*360 + lng) * math.Pi / 180
	latRad := lat * math.Pi / 180
	altitude := math.Asin(math.Sin(latRad)*math.Sin(decl) + math.Cos(latRad)*math.Cos(decl)*math.Cos(hourAngle))

	if altitude < -1 {
		altitude = -1
	}
	if altitude > 1 {
		altitude = 1
	}
	elevation = altitude * 180 / math.Pi

	azRad := math.Atan2(
		-math.Sin(hourAngle),
		math.Tan(decl)*math.Cos(latRad)-math.Sin(latRad)*math.Cos(hourAngle),
	)
	azimuth = (azRad*180/math.Pi + 360)
	for azimuth >= 360 {
		azimuth -= 360
	}
	for azimuth < 0 {
		azimuth += 360
	}

	return
}
