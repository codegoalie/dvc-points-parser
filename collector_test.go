package main

import (
	"fmt"
	"testing"

	tinydate "github.com/lane-c-wagner/go-tinydate"
)

func TestOne(t *testing.T) {
	expexctedCheckIn, err := tinydate.New(2020, 1, 1)
	if err != nil {
		err = fmt.Errorf("failed to create expected check in date: %w", err)
		t.Fatal(err)
	}
	expexctedCheckOut, err := tinydate.New(2020, 1, 31)
	if err != nil {
		err = fmt.Errorf("failed to create expected check out date: %w", err)
		t.Fatal(err)
	}

	coll := &collector{
		Dates: []dateRange{
			{CheckInAt: expexctedCheckIn, CheckOutAt: expexctedCheckOut},
		},
		Points: [2][]int{
			{17, 20, 33, 40, 46, 55, 112},
			{20, 24, 41, 48, 55, 66, 132},
		},
	}

	resort := baseResort()

	collectorToResort(coll, resort)

	// spew.Dump(resort)

	for i, roomType := range resort.RoomTypes {
		if len(roomType.PointChart) != len(coll.Dates) {
			t.Errorf("incorrect number of point blocks collected for %s: expected %d; got %d", roomType.Name, len(coll.Dates), len(roomType.PointChart))
		}

		block := roomType.PointChart[0]
		if !block.CheckInAt.Equal(expexctedCheckIn) {
			t.Errorf("collectorToResort expected check in date of %v; got %v", expexctedCheckIn, block.CheckInAt)
		}
		if !block.CheckOutAt.Equal(expexctedCheckOut) {
			t.Errorf("collectorToResort expected check out date of %v; got %v", expexctedCheckOut, block.CheckOutAt)
		}
		if block.WeekdayPoints != coll.Points[0][i] {
			t.Errorf("collectorToResort expected weekday poinnts of %d; got %d", coll.Points[0][i], block.WeekdayPoints)
		}
		if block.WeekendPoints != coll.Points[1][i] {
			t.Errorf("collectorToResort expected weekend poinnts of %d; got %d", coll.Points[1][i], block.WeekendPoints)
		}
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
