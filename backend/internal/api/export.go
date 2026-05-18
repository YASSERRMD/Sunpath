package api

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/horizon"
	"github.com/yasserrmd/sunpath/backend/internal/sun"
)

func computeDayStats(lat, lng, h float64, profile [360]float64, year, month, day int) (totalMinutes int, firstSun, lastSun float64) {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	firstSun = -1
	lastSun = -1

	for min := 0; min < 1440; min++ {
		ct := t.Add(time.Duration(min) * time.Minute)
		az, el := sun.SolarPosition(ct, lat, lng)
		azIdx := int(az + 0.5)
		if azIdx < 0 {
			azIdx = 0
		}
		if azIdx > 359 {
			azIdx = 359
		}
		if el > 0 && el > profile[azIdx] {
			totalMinutes++
			if firstSun < 0 {
				firstSun = float64(min) / 60.0
			}
			lastSun = float64(min) / 60.0
		}
	}
	return
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	hStr := r.URL.Query().Get("h")
	year := 2025

	if latStr == "" || lngStr == "" {
		s.writeError(w, 400, "lat and lng are required")
		return
	}

	lat := parseFloat(latStr)
	if math.IsNaN(lat) || lat < -90 || lat > 90 {
		s.writeError(w, 400, "invalid lat")
		return
	}
	lng := parseFloat(lngStr)
	if math.IsNaN(lng) || lng < -180 || lng > 180 {
		s.writeError(w, 400, "invalid lng")
		return
	}
	h := 1.5
	if hStr != "" {
		h = parseFloat(hStr)
		if math.IsNaN(h) || h < 0 {
			s.writeError(w, 400, "invalid h")
			return
		}
	}

	p := geo.Point{Lat: lat, Lng: lng}
	buildings, err := s.cachedClient.FetchBuildingsInBBox(
		lat-0.1, lng-0.1, lat+0.1, lng+0.1,
	)
	if err != nil {
		s.writeError(w, 500, "failed to fetch buildings")
		return
	}

	profile, err := s.horizonComp.Compute(p, h, buildings)
	if err != nil {
		s.writeError(w, 500, "failed to compute horizon")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=sunpath_%.4f_%.4f.csv", lat, lng))

	fmt.Fprintf(w, "date,day_of_year,sun_minutes,first_sun_hour,last_sun_hour\n")
	for doy := 1; doy <= 365; doy++ {
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, doy-1)
		total, firstSun, lastSun := computeDayStats(lat, lng, h, profile.Horizon, t.Year(), int(t.Month()), t.Day())

		firstStr := fmt.Sprintf("%.2f", firstSun)
		lastStr := fmt.Sprintf("%.2f", lastSun)
		if firstSun < 0 {
			firstStr = "none"
			lastStr = "none"
		}
		fmt.Fprintf(w, "%s,%d,%d,%s,%s\n", t.Format("2006-01-02"), doy, total, firstStr, lastStr)
	}
}

func (s *Server) handleExportPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	hStr := r.URL.Query().Get("h")

	if latStr == "" || lngStr == "" {
		s.writeError(w, 400, "lat and lng are required")
		return
	}

	lat := parseFloat(latStr)
	if math.IsNaN(lat) || lat < -90 || lat > 90 {
		s.writeError(w, 400, "invalid lat")
		return
	}
	lng := parseFloat(lngStr)
	if math.IsNaN(lng) || lng < -180 || lng > 180 {
		s.writeError(w, 400, "invalid lng")
		return
	}
	h := 1.5
	if hStr != "" {
		h = parseFloat(hStr)
		if math.IsNaN(h) || h < 0 {
			s.writeError(w, 400, "invalid h")
			return
		}
	}

	p := geo.Point{Lat: lat, Lng: lng}
	buildings, err := s.cachedClient.FetchBuildingsInBBox(
		lat-0.1, lng-0.1, lat+0.1, lng+0.1,
	)
	if err != nil {
		s.writeError(w, 500, "failed to fetch buildings")
		return
	}

	profile, err := s.horizonComp.Compute(p, h, buildings)
	if err != nil {
		s.writeError(w, 500, "failed to compute horizon")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=sunpath_%.4f_%.4f.pdf", lat, lng))

	html := buildPDFHTML(lat, lng, h, profile)
	w.Write([]byte(html))
}

func buildPDFHTML(lat, lng, h float64, profile horizon.Profile) string {
	totalYearMins := 0
	bestDay := 0
	worstDay := 365
	bestMins := 0
	worstMins := 1440

	rows := ""
	for doy := 1; doy <= 365; doy++ {
		t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, doy-1)
		total, firstSun, lastSun := computeDayStats(lat, lng, h, profile.Horizon, t.Year(), int(t.Month()), t.Day())
		totalYearMins += total
		if total > bestMins {
			bestMins = total
			bestDay = doy
		}
		if total < worstMins {
			worstMins = total
			worstDay = doy
		}

		firstStr := "none"
		lastStr := "none"
		if firstSun >= 0 {
			firstStr = fmt.Sprintf("%.1fh", firstSun)
			lastStr = fmt.Sprintf("%.1fh", lastSun)
		}
		rows += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%s</td></tr>\n",
			t.Format("Jan 2"), doy, total, firstStr, lastStr)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Sunpath Report</title>
<style>
  body { font-family: Helvetica, Arial, sans-serif; font-size: 10pt; margin: 20px; }
  h1 { font-size: 16pt; }
  table { border-collapse: collapse; width: 100%%; }
  th, td { border: 1px solid #ccc; padding: 3px 6px; text-align: center; }
  th { background: #f0f0f0; }
  .summary { margin: 10px 0; padding: 10px; background: #f8f8f8; border: 1px solid #ddd; }
</style></head><body>
<h1>Sunpath Solar Exposure Report</h1>
<div class="summary">
  <p><strong>Location:</strong> %.4f, %.4f</p>
  <p><strong>Observer height:</strong> %.1f m</p>
  <p><strong>Total annual sun:</strong> %d hours</p>
  <p><strong>Best day:</strong> day %d (%d min)</p>
  <p><strong>Worst day:</strong> day %d (%d min)</p>
</div>
<table>
<tr><th>Date</th><th>Day</th><th>Sun (min)</th><th>First sun</th><th>Last sun</th></tr>
%s
</table>
</body></html>`,
		lat, lng, h, totalYearMins/60, bestDay, bestMins, worstDay, worstMins, rows)
}
