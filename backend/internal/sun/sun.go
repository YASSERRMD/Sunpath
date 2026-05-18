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

	// Mean anomaly
	g := (357.528 + 0.9856003*n) * math.Pi / 180

	// Mean solar longitude
	L := (280.460 + 0.9856474*n) * math.Pi / 180

	// Ecliptic longitude
	lambda := L + (1.915*math.Sin(g)+0.020*math.Sin(2*g))*math.Pi/180

	// Obliquity of the ecliptic
	epsilon := earthTilt

	// Solar declination
	sinDecl := math.Sin(epsilon) * math.Sin(lambda)
	decl := math.Asin(sinDecl)

	// Equation of time (minutes)
	eot := 229.18 * (0.000075 + 0.001868*math.Cos(g) - 0.032077*math.Sin(g) - 0.014615*math.Cos(2*g) - 0.04089*math.Sin(2*g))

	// Time correction (minutes) = 4 * longitude + eot
	timeCorrection := eot + 4*lng

	// Solar time (hours)
	fracDay := jd - 0.5 - math.Floor(jd-0.5)
	solarTime := (fracDay*24*60 + timeCorrection) / 60

	// Hour angle (degrees), 0 at solar noon
	hourAngleDeg := (solarTime - 12) * 15
	haRad := hourAngleDeg * math.Pi / 180

	latRad := lat * math.Pi / 180

	// Solar elevation
	sinAlt := math.Sin(latRad)*math.Sin(decl) + math.Cos(latRad)*math.Cos(decl)*math.Cos(haRad)
	if sinAlt > 1 {
		sinAlt = 1
	}
	if sinAlt < -1 {
		sinAlt = -1
	}
	elevation = math.Asin(sinAlt) * 180 / math.Pi

	// Solar azimuth from standard formula (measured from south, positive west)
	// Then convert to 0=N, 90=E, 180=S, 270=W convention
	azRad := math.Atan2(
		math.Sin(haRad),
		math.Cos(haRad)*math.Sin(latRad)-math.Tan(decl)*math.Cos(latRad),
	)
	azimuth = (azRad*180/math.Pi + 180)
	for azimuth >= 360 {
		azimuth -= 360
	}
	for azimuth < 0 {
		azimuth += 360
	}

	return
}
