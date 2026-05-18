package sun

import (
	"math"
	"time"
)

type DayTimes struct {
	SolarNoon time.Time
	Sunrise   time.Time
	Sunset    time.Time
}

func SolarNoon(t time.Time, lng float64) time.Time {
	y := t.Year()
	m := t.Month()
	d := t.Day()
	noon := time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
	return noon.Add(-time.Duration(lng*240) * time.Second)
}

func ComputeDayTimes(t time.Time, lat, lng float64) DayTimes {
	jd := julianDate(t)
	n := jd - 2451545.0
	L := (280.460 + 0.9856474*n) * math.Pi / 180
	g := (357.528 + 0.9856003*n) * math.Pi / 180
	lambda := L + (1.915*math.Sin(g)+0.020*math.Sin(2*g))*math.Pi/180
	sinDecl := math.Sin(earthTilt) * math.Sin(lambda)
	decl := math.Asin(sinDecl)
	declDeg := decl * 180 / math.Pi

	latDeg := lat
	cosHA := -(math.Sin(0.8333*math.Pi/180) + math.Sin(latDeg*math.Pi/180)*math.Sin(declDeg*math.Pi/180)) /
		(math.Cos(latDeg*math.Pi/180) * math.Cos(declDeg*math.Pi/180))

	if cosHA < -1 {
		cosHA = -1
	}
	if cosHA > 1 {
		cosHA = 1
	}

	ha := math.Acos(cosHA) * 180 / math.Pi

	y := t.Year()
	m := t.Month()
	d := t.Day()
	noon := time.Date(y, m, d, 12, 0, 0, 0, time.UTC)

	solarNoon := noon.Add(-time.Duration(lng*240) * time.Second)
	sunrise := solarNoon.Add(-time.Duration(ha*240) * time.Second)
	sunset := solarNoon.Add(time.Duration(ha*240) * time.Second)

	return DayTimes{
		SolarNoon: solarNoon,
		Sunrise:   sunrise,
		Sunset:    sunset,
	}
}
