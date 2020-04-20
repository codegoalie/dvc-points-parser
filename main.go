package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	StartDate     time.Time
	EndDate       time.Time
	WeekdayPoints int
	WeekendPoints int
}

func main() {
	var files []string
	files = append(files, "converted-charts/2020/GFV_PointsChart.pdf-2020.pdf.txt")

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

		resorts[i], err = parseFile(file)
		if err != nil {
			err = fmt.Errorf("failed to parse file %s: %w", filename, err)
			log.Fatal(err)
		}
	}

	spew.Dump(resorts)
}

func parseFile(file *os.File) (Resort, error) {
	resort := Resort{}
	state := 0
	roomTypes := []RoomType{}
	viewLegend := map[string]string{}
	stateInnterIndex := 0

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
		default:
			break
		}
		stateInnterIndex++
	}

	if err := scanner.Err(); err != nil {
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
