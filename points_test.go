package main

import "testing"

func TestWeekdayPoints(t *testing.T) {
	coll := &collector{}
	expected := []int{17, 20, 33, 40, 46, 55, 112}
	line := "SUN--THU  17  20  33  40  46   55  112"

	parsePoints(coll, line)

	actual := coll.Points[0]
	if len(actual) != len(expected) {
		t.Errorf("incorrect number of points parsed from '%s': expected : %d -- got: %d", line, len(expected), len(actual))
	}

	for i, actualPts := range actual {
		if actualPts != expected[i] {
			t.Errorf("incorrect points parsed from '%s' at %d: expected : %d -- got: %d", line, i, expected[i], actualPts)
		}
	}
}

func TestWeekendPoints(t *testing.T) {
	coll := &collector{}
	expected := []int{17, 20, 33, 40, 46, 55, 112}
	line := "FRI--SAT  17  20  33  40  46   55  112"

	parsePoints(coll, line)

	actual := coll.Points[1]
	if len(actual) != len(expected) {
		t.Errorf("incorrect number of points parsed from '%s': expected : %d -- got: %d", line, len(expected), len(actual))
	}

	for i, actualPts := range actual {
		if actualPts != expected[i] {
			t.Errorf("incorrect points parsed from '%s' at %d: expected : %d -- got: %d", line, i, expected[i], actualPts)
		}
	}
}

func TestOneSetOfPoints(t *testing.T) {
	coll := &collector{}
	expected := []int{17, 20, 33, 40, 46, 55, 112}
	line := "SUN--SAT  17  20  33  40  46   55  112"

	parsePoints(coll, line)

	actual := coll.Points[0]
	if len(actual) != len(expected) {
		t.Errorf("incorrect number of points parsed from '%s': expected : %d -- got: %d", line, len(expected), len(actual))
	}

	for i, actualPts := range actual {
		if actualPts != expected[i] {
			t.Errorf("incorrect points parsed from '%s' at %d: expected : %d -- got: %d", line, i, expected[i], actualPts)
		}
	}

	actual = coll.Points[1]
	if len(actual) != len(expected) {
		t.Errorf("incorrect number of points parsed from '%s': expected : %d -- got: %d", line, len(expected), len(actual))
	}

	for i, actualPts := range actual {
		if actualPts != expected[i] {
			t.Errorf("incorrect points parsed from '%s' at %d: expected : %d -- got: %d", line, i, expected[i], actualPts)
		}
	}
}
