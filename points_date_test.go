package main

import (
	"testing"
	"time"
)

func TestDatesSameMonth(t *testing.T) {
	line := "Jan 1--31"
	coll := &collector{}
	expexctedCheckIn, _ := time.Parse("Jan 2 2006", "Jan 1 2020")
	expexctedCheckOut, _ := time.Parse("Jan 2 2006", "Jan 31 2020")

	parseDates(coll, "2020", line)

	if len(coll.Dates) != 1 {
		t.Errorf("Wrong amount of dates parsed: expected 1; got %d: %s", len(coll.Dates), line)
	}

	firstRange := coll.Dates[0]
	if firstRange.CheckInAt != expexctedCheckIn {
		t.Errorf("parseDates(%s) = %+v; expected %+v", line, firstRange.CheckInAt, expexctedCheckIn)
	}
	if firstRange.CheckOutAt != expexctedCheckOut {
		t.Errorf("parseDates(%s) = %+v; expected %+v", line, firstRange.CheckOutAt, expexctedCheckOut)
	}
}

func TestDatesCrossMonth(t *testing.T) {
	line := "Jan 1--Feb 3"
	coll := &collector{}
	expexctedCheckIn, _ := time.Parse("Jan 2 2006", "Jan 1 2020")
	expexctedCheckOut, _ := time.Parse("Jan 2 2006", "Feb 3 2020")

	parseDates(coll, "2020", line)

	if len(coll.Dates) != 1 {
		t.Errorf("Wrong amount of dates parsed: expected 1; got %d: %s", len(coll.Dates), line)
	}

	firstRange := coll.Dates[0]
	if firstRange.CheckInAt != expexctedCheckIn {
		t.Errorf("parseDates(%s) = %+v; expected %+v", line, firstRange.CheckInAt, expexctedCheckIn)
	}
	if firstRange.CheckOutAt != expexctedCheckOut {
		t.Errorf("parseDates(%s) = %+v; expected %+v", line, firstRange.CheckOutAt, expexctedCheckOut)
	}
}

func TestPoints(t *testing.T) {
	// line := "SUN--THU  17  20  33  40  46   55  112"
	resort := baseResort()

	// parsePoints(resort, line)

	expected := resort.RoomTypes[0].PointChart
	if len(expected) < 1 {
		// t.Errorf("no points parsed from '%s': %+v", line, expected)
	}
}

func baseResort() *Resort {
	return &Resort{
		Name: "The Villas at Disney's Grand Floridian Resort & Spa",
		RoomTypes: []RoomType{
			{
				Name:     "Deluxe studio",
				ViewType: "Standard",
			},
			{
				Name:     "Deluxe studio",
				ViewType: "Lake",
			},
			{
				Name:     "1 bedroom villa",
				ViewType: "Standard",
			},
			{
				Name:     "1 bedroom villa",
				ViewType: "Lake",
			},
			{
				Name:     "2 bedroom villa",
				ViewType: "Standard",
			},
			{
				Name:     "2 bedroom villa",
				ViewType: "Lake",
			},
			{
				Name:     "3 bedroom villa",
				ViewType: "Lake",
			},
		},
	}
}
