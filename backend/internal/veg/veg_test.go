package veg

import (
	"math"
	"testing"
)

func TestParseMeasurementMetres(t *testing.T) {
	v, unit, err := parseMeasurement("10m")
	if err != nil || math.Abs(v-10) > 0.01 || unit != "m" {
		t.Errorf("expected 10m, got %f %s", v, unit)
	}
}

func TestParseMeasurementFeet(t *testing.T) {
	v, unit, err := parseMeasurement("20ft")
	if err != nil || math.Abs(v-6.096) > 0.01 || unit != "m" {
		t.Errorf("expected 6.096m, got %f %s", v, unit)
	}
}

func TestParseTreeFromOSM(t *testing.T) {
	tags := map[string]string{"height": "15m", "width": "8m"}
	tree := ParseTreeFromOSM(tags, 48.85, 2.35)

	if math.Abs(tree.Height-15) > 0.1 {
		t.Errorf("expected height 15, got %f", tree.Height)
	}
	if math.Abs(tree.CrownWidth-8) > 0.1 {
		t.Errorf("expected crown width 8, got %f", tree.CrownWidth)
	}
	if tree.Estimated {
		t.Error("expected not estimated")
	}
}

func TestParseTreeFromOSMDefault(t *testing.T) {
	tags := map[string]string{}
	tree := ParseTreeFromOSM(tags, 48.85, 2.35)

	if !tree.Estimated {
		t.Error("expected estimated")
	}
	if tree.Height <= 0 {
		t.Error("expected default height")
	}
	if tree.CrownWidth <= 0 {
		t.Error("expected default crown width")
	}
}

func TestComputeVegetationHorizonEmpty(t *testing.T) {
	horizon := ComputeVegetationHorizon(nil, 48.85, 2.35, 1.5)
	for az := 0; az < 360; az++ {
		if horizon[az] != 0 {
			t.Errorf("expected 0 at az %d, got %f", az, horizon[az])
			break
		}
	}
}

func TestComputeVegetationHorizonSingleTree(t *testing.T) {
	trees := []Tree{
		{Lat: 48.851, Lng: 2.351, Height: 15, CrownWidth: 8},
	}
	horizon := ComputeVegetationHorizon(trees, 48.85, 2.35, 1.5)

	hasNonZero := false
	for az := 0; az < 360; az++ {
		if horizon[az] > 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("expected some obstruction from tree")
	}
}
