package sun

import (
	"math"
	"time"
)

type Sample struct {
	Time      time.Time
	Azimuth   float64
	Elevation float64
}

func SampleDay(t time.Time, lat, lng float64, intervalSec int) []Sample {
	if intervalSec <= 0 {
		intervalSec = 60
	}
	y := t.Year()
	m := t.Month()
	d := t.Day()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	var samples []Sample
	for cursor := start; cursor.Before(end); cursor = cursor.Add(time.Duration(intervalSec) * time.Second) {
		az, el := SolarPosition(cursor, lat, lng)
		el = math.Max(el, -90)
		samples = append(samples, Sample{
			Time:      cursor,
			Azimuth:   az,
			Elevation: el,
		})
	}
	return samples
}
