package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tinydate "github.com/lane-c-wagner/go-tinydate"
)

// Resort models a WDW resort
type Resort struct {
	Name      string
	RoomTypes []RoomType
}

// RoomType models a room size and view combination
type RoomType struct {
	Name        string
	Description string
	ViewType    string

	PointChart []PointBlock
}

// PointBlock models the points needed to stay in a RoomType over a range of dates
type PointBlock struct {
	StartDate     tinydate.TinyDate
	EndDate       tinydate.TinyDate
	WeekdayPoints int
	WeekendPoints int
}

type collector struct {
	Dates  []dateRange
	Points [2][]int
}

type dateRange struct {
	CheckInAt  time.Time
	CheckOutAt time.Time
}

const dateParseFormat = "Jan 2 2006"

var monthDayRegexp = regexp.MustCompile(`^[a-zA-z]{3} \d`)
var yearRegexp = regexp.MustCompile(`(\d{4})`)

func main() {
	var files []string
	files = append(files, "converted-charts/2020/GFV_PointsChart-2020.txt")

	// root := "converted-charts/"
	// err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
	// 	if info.IsDir() {
	// 		return nil
	// 	}

	// 	files = append(files, path)

	// 	return nil
	// })
	// if err != nil {
	// 	panic(err)
	// }

	resorts := make([]Resort, len(files))
	for i, filename := range files {
		fmt.Println(filename)
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		year := yearRegexp.FindStringSubmatch(filename)[1]

		resorts[i], err = parseFile(file, year)
		if err != nil {
			err = fmt.Errorf("failed to parse file %s: %w", filename, err)
			log.Fatal(err)
		}
	}

	// spew.Dump(resorts)
}

func parseFile(file *os.File, year string) (Resort, error) {
	resort := Resort{}
	state := 0
	roomTypes := []RoomType{}
	viewLegend := map[string]string{}
	stateInnterIndex := 0
	coll := &collector{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			state++
			stateInnterIndex = 0
			continue
		}

		switch state {
		case 0:
			parseName(&resort, line)
		case 1:
			parseRoomType(&roomTypes, line)
		case 2:
			parseRoomDescriptions(&roomTypes[stateInnterIndex], line)
		case 3:
			parseViewLegend(&viewLegend, line)
		case 4:
			parseRoomViews(&resort, roomTypes[stateInnterIndex], viewLegend, line)
		case 5:
			parseDates(coll, year, line)
		case 6:
			parsePoints(coll, line)
		default:
			// collectorToResort(coll, &resort)
			coll = &collector{}
			parseDates(coll, year, line)
			state = 5
		}
		stateInnterIndex++
	}

	err := scanner.Err()
	if err != nil && !errors.Is(err, io.EOF) {
		err = fmt.Errorf("failed to read from file: %w", err)
		return resort, err
	}

	return resort, nil
}

func parseName(resort *Resort, line string) {
	resort.Name = line
}

func parseRoomType(roomTypes *[]RoomType, line string) {
	*roomTypes = append(*roomTypes, RoomType{
		Name: line,
	})
}

func parseRoomDescriptions(roomType *RoomType, line string) {
	roomType.Description = line
}

func parseViewLegend(legend *map[string]string, line string) {
	fields := strings.Fields(line)
	(*legend)[fields[0]] = fields[2]
}

func parseRoomViews(resort *Resort, roomType RoomType, viewLegend map[string]string, line string) {
	for _, viewKey := range strings.Fields(line) {
		roomType.ViewType = viewLegend[viewKey]
		resort.RoomTypes = append(resort.RoomTypes, roomType)
	}
}

func parseDates(coll *collector, year string, line string) {
	dates := strings.Split(line, "--")

	checkInAt, err := time.Parse(dateParseFormat, dates[0]+" "+year)
	if err != nil {
		err = fmt.Errorf("failed to parse check in date '%s %s': %w", dates[0], year, err)
		log.Fatal(err)
	}

	checkOutString := ""
	if strings.Index(dates[1], " ") == -1 {
		parts := strings.Fields(dates[0])
		checkOutString = parts[0] + " "
	}
	checkOutString += dates[1]
	checkOutAt, err := time.Parse(dateParseFormat, checkOutString+" "+year)
	if err != nil {
		err = fmt.Errorf("failed to parse check out date '%s %s': %w", checkOutString, year, err)
		log.Fatal(err)
	}

	coll.Dates = append(coll.Dates, dateRange{
		CheckInAt:  checkInAt,
		CheckOutAt: checkOutAt,
	})
}

func parsePoints(coll *collector, line string) {
	fields := strings.Fields(line)
	days := fields[0]
	points := []int{}

	for i := 1; i < len(fields); i++ {
		pts, err := strconv.Atoi(fields[i])
		if err != nil {
			err = fmt.Errorf("failed to parse points '%s': %w", fields[i], err)
			log.Fatal(err)
		}
		points = append(points, pts)
	}
	if days == "SUN--THU" || days == "SUN--SAT" {
		coll.Points[0] = points
	}
	if days == "FRI--SAT" || days == "SUN--SAT" {
		coll.Points[1] = points
	}
}
