package resorts

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

	tinydate "github.com/lane-c-wagner/go-tinydate"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	CheckInAt     tinydate.TinyDate
	CheckOutAt    tinydate.TinyDate
	WeekdayPoints int
	WeekendPoints int
}

type collector struct {
	Dates  []dateRange
	Points [2][]int
}

type dateRange struct {
	CheckInAt  tinydate.TinyDate
	CheckOutAt tinydate.TinyDate
}

const dateParseFormat = "Jan 2 2006"

var (
	monthDayRegexp  = regexp.MustCompile(`^[a-zA-z]{3} \d`)
	yearRegexp      = regexp.MustCompile(`(\d{4})`)
	dateSplitRegexp = regexp.MustCompile(`--| - `)
)

func ParseFile(filename string) (Resort, error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open '%s': %w", filename, err)
		return Resort{}, err
	}
	defer file.Close()

	year := yearRegexp.FindStringSubmatch(filename)[1]
	// fmt.Println("Parsing", filename, year)

	resort, err := ParseUneditedFile(file, year)
	if err != nil {
		err = fmt.Errorf("failed to parse file %s: %w", filename, err)
		return resort, err
	}

	return resort, nil
}

func parseFile(file *os.File, year string) (Resort, error) {
	resort := Resort{}
	state := 0
	roomTypes := []RoomType{}
	viewLegend := map[string]string{}
	stateInnterIndex := 0
	coll := collector{}

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
			roomTypes = append(roomTypes, parseRoomType(line))
		case 2:
			parseRoomDescriptions(&roomTypes[stateInnterIndex], line)
		case 3:
			parseViewLegend(&viewLegend, line)
		case 4:
			parseRoomViews(&resort, roomTypes[stateInnterIndex], viewLegend, line)
		case 5:
			parseDates(&coll, year, line)
		case 6:
			parsePoints(&coll, line)
		default:
			collectorToResort(&coll, &resort)
			coll = collector{}
			parseDates(&coll, year, line)
			state = 5
		}
		stateInnterIndex++
	}

	collectorToResort(&coll, &resort)

	err := scanner.Err()
	if err != nil && !errors.Is(err, io.EOF) {
		err = fmt.Errorf("failed to read from file: %w", err)
		return resort, err
	}

	return resort, nil
}

func ParseUneditedFile(file io.Reader, year string) (Resort, error) {
	resort := Resort{}
	state := 0
	roomTypes := []RoomType{}
	viewLegend := map[string]string{}
	stateInnterIndex := 0
	coll := collector{}

	r := regexp.MustCompile(`Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept|Oct|Nov|Dec`)

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
			roomTypes = append(roomTypes, parseRoomType(line))
		case 2:
			parseRoomDescriptions(&roomTypes[stateInnterIndex], line)
		case 3:
			parseViewLegend(&viewLegend, line)
		case 4:
			parseRoomViews(&resort, roomTypes[stateInnterIndex], viewLegend, line)
		default:
			tokens := strings.Fields(line)
			fmt.Println("t", tokens)
			if len(tokens) < 1 {
				continue
			}

			if strings.Contains(line, "SUN") {
				if len(coll.Dates) > 0 {
					collectorToResort(&coll, &resort)
					coll = collector{}
				}

				pointsLine := line[strings.Index(line, "SUN"):]
				// fmt.Println("p", pointsLine)
				parsePoints(&coll, pointsLine)
				// fmt.Println("c", coll)
			}

			// fmt.Println("f", tokens[0], tokens[5])
			if strings.HasPrefix(tokens[0], "FRI") || (len(tokens) > 5 && strings.HasPrefix(tokens[5], "FRI")) {
				parsePoints(&coll, strings.Join(tokens[5:], " "))
				// fmt.Println("cf", coll)
			}

			if r.MatchString(tokens[0]) {
				parseDates(&coll, year, strings.Join(tokens[0:5], " "))
				// fmt.Println("cd", coll)
			}
		}
		stateInnterIndex++
	}

	collectorToResort(&coll, &resort)

	err := scanner.Err()
	if err != nil && !errors.Is(err, io.EOF) {
		err = fmt.Errorf("failed to read from file: %w", err)
		return resort, err
	}

	return resort, nil
}

func freeParsePoints(scanner *bufio.Scanner) collector {
	coll := collector{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Fields(line)
		fmt.Println(tokens)
		if len(tokens) == 0 {
			continue
		}

	}
	return coll
}

func parseName(resort *Resort, line string) {
	resort.Name = line
}

func parseRoomType(line string) RoomType {
	caser := cases.Title(language.AmericanEnglish)
	fmt.Println("rt", caser.String(strings.ToLower(line)))
	return RoomType{Name: caser.String(strings.ToLower(line))}
}

func parseRoomDescriptions(roomType *RoomType, line string) {
	roomType.Description = line
}

func parseViewLegend(legend *map[string]string, line string) {
	fields := strings.Split(line, " - ")
	(*legend)[strings.TrimSpace(fields[0])] = strings.TrimSpace(fields[1])
}

func parseRoomViews(resort *Resort, roomType RoomType, viewLegend map[string]string, line string) {
	for _, viewKey := range strings.Fields(line) {
		roomType.ViewType = viewLegend[viewKey]
		resort.RoomTypes = append(resort.RoomTypes, roomType)
	}
}

func parseDates(coll *collector, year string, line string) {
	dates := dateSplitRegexp.Split(line, -1)

	checkInAt, err := parseADate(dates[0] + " " + year)
	if err != nil {
		err = fmt.Errorf("failed to parse check in date '%s %s': %w", dates[0], year, err)
		log.Fatal(err)
	}

	// One date means check in and out are the same
	if len(dates) < 2 {
		coll.Dates = append(coll.Dates, dateRange{
			CheckInAt:  checkInAt,
			CheckOutAt: checkInAt,
		})
		return
	}

	checkOutString := ""
	if strings.Index(dates[1], " ") == -1 {
		parts := strings.Fields(dates[0])
		checkOutString = parts[0] + " "
	}
	checkOutString += dates[1]
	checkOutAt, err := parseADate(checkOutString + " " + year)
	if err != nil {
		err = fmt.Errorf("failed to parse check out date '%s %s': %w", checkOutString, year, err)
		log.Fatal(err)
	}

	coll.Dates = append(coll.Dates, dateRange{
		CheckInAt:  checkInAt,
		CheckOutAt: checkOutAt,
	})
}

func parseADate(in string) (tinydate.TinyDate, error) {
	pieces := strings.Split(in, " ")

	if len(pieces) < 3 {
		return tinydate.TinyDate{}, errors.New("not enough parts of a date")
	}

	if len(pieces[0]) > 3 {
		pieces[0] = pieces[0][:3]
	}

	return tinydate.Parse(dateParseFormat, strings.Join(pieces, " "))
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

func collectorToResort(coll *collector, resort *Resort) {
	for i := 0; i < len(resort.RoomTypes); i++ {
		for _, dates := range coll.Dates {
			resort.RoomTypes[i].PointChart = append(resort.RoomTypes[i].PointChart,
				PointBlock{
					CheckInAt:     dates.CheckInAt,
					CheckOutAt:    dates.CheckOutAt,
					WeekdayPoints: coll.Points[0][i],
					WeekendPoints: coll.Points[1][i],
				})
		}
	}
}
