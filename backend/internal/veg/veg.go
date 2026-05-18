package veg

import (
	"fmt"
	"math"
)

type Tree struct {
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Height     float64 `json:"height"`
	CrownWidth float64 `json:"crown_width"` // diameter of the crown in metres
	Estimated  bool    `json:"estimated"`
}

func DefaultTreeHeight() float64 {
	return 10.0
}

func DefaultCrownWidth(height float64) float64 {
	return height * 0.6
}

func ComputeVegetationHorizon(trees []Tree, observerLat, observerLng, observerHeight float64) [360]float64 {
	var vegHorizon [360]float64

	originLatRad := observerLat * math.Pi / 180

	for _, tree := range trees {
		dLat := (tree.Lat - observerLat) * 111320.0
		dLng := (tree.Lng - observerLng) * 111320.0 * math.Cos(originLatRad)
		dist := math.Sqrt(dLat*dLat + dLng*dLng)

		azRad := math.Atan2(dLng, dLat)
		az := int((azRad*180/math.Pi + 360 + 90)) % 360

		crownRadius := tree.CrownWidth / 2.0
		crownTop := tree.Height
		trunkHeight := crownTop * 0.3

		for d := dist - crownRadius; d <= dist+crownRadius; d += 1.0 {
			if d <= 0 {
				continue
			}

			frac := (d - (dist - crownRadius)) / (2 * crownRadius)
			if frac < 0 {
				frac = 0
			}
			if frac > 1 {
				frac = 1
			}

			coneHeight := trunkHeight + (crownTop-trunkHeight)*(1-frac)

			var relHeight float64
			if dist <= crownRadius {
				relHeight = coneHeight - observerHeight
			} else {
				crownFrac := 1.0 - frac
				if crownFrac > 0 {
					coneContribution := (crownTop - observerHeight) * crownFrac
					trunkContribution := trunkHeight - observerHeight
					if trunkContribution < 0 {
						trunkContribution = 0
					}
					if coneContribution > trunkContribution {
						relHeight = coneContribution
					} else {
						relHeight = trunkContribution
					}
				} else {
					relHeight = trunkHeight - observerHeight
				}
				if relHeight < 0 {
					relHeight = 0
				}
			}

			elev := math.Atan2(relHeight, d) * 180 / math.Pi
			if elev > vegHorizon[az] {
				vegHorizon[az] = elev
			}

			adjAz := (az + 1) % 360
			if elev > vegHorizon[adjAz] {
				vegHorizon[adjAz] = elev * 0.85
			}
			adjAz = (az + 359) % 360
			if elev > vegHorizon[adjAz] {
				vegHorizon[adjAz] = elev * 0.85
			}
		}
	}

	return vegHorizon
}

func ParseTreeFromOSM(tags map[string]string, lat, lng float64) Tree {
	t := Tree{Lat: lat, Lng: lng}

	if hStr, ok := tags["height"]; ok {
		if h, err := parseMetre(hStr); err == nil {
			t.Height = h
		}
	}
	if t.Height <= 0 {
		t.Height = DefaultTreeHeight()
		t.Estimated = true
	}

	if wStr, ok := tags["width"]; ok {
		if w, err := parseMetre(wStr); err == nil {
			t.CrownWidth = w
		}
	}
	if t.CrownWidth <= 0 {
		t.CrownWidth = DefaultCrownWidth(t.Height)
		t.Estimated = true
	}

	return t
}

func parseMetre(s string) (float64, error) {
	v, _, err := parseMeasurement(s)
	return v, err
}

func parseMeasurement(s string) (float64, string, error) {
	val := 0.0
	unit := ""
	for i, c := range s {
		if c >= '0' && c <= '9' || c == '.' || c == '-' {
			continue
		}
		v, err := fmtParseFloat(s[:i])
		if err != nil {
			return 0, "", err
		}
		val = v
		unit = s[i:]
		break
	}
	if unit == "" {
		v, err := fmtParseFloat(s)
		return v, "", err
	}
	switch unit {
	case "m", "meter", "metre", "meters", "metres":
		return val, "m", nil
	case "ft", "feet", "foot":
		return val * 0.3048, "m", nil
	case "in", "inch":
		return val * 0.0254, "m", nil
	default:
		return val, unit, nil
	}
}

func fmtParseFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}
