package main

import (
	"bufio"
	"errors"
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
	var err error
	viewLegend := map[string]string{}
	typeIndex := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			state++
			continue
		}

		switch state {
		case 0:
			parseName(&resort, line)
		case 1:
			parseRoomType(&resort, line)
		case 2:
			parseViewLegend(&viewLegend, line)
		case 3:
			err = parseRoomViews(&resort, typeIndex, viewLegend, line)
			if err != nil {
				err = fmt.Errorf("failed to parse room views: %w", err)
				return resort, err
			}
			typeIndex++
		default:
			break
		}
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

func parseRoomType(resort *Resort, line string) {
	resort.RoomTypes = append(resort.RoomTypes, RoomType{
		Name: line,
	})
}

func parseViewLegend(legend *map[string]string, line string) {
	fields := strings.Fields(line)
	(*legend)[fields[0]] = fields[2]
}

func parseRoomViews(resort *Resort, typeIndex int, viewLegend map[string]string, line string) error {
	if len(resort.RoomTypes) < 1 {
		return errors.New("cannot add room views without any room types")
	}

	currentRoomType := &resort.RoomTypes[typeIndex]

	for _, viewKey := range strings.Fields(line) {
		viewType := viewLegend[viewKey]
		fmt.Println(viewKey, viewLegend["L"])

		if currentRoomType.ViewType == "" {
			currentRoomType.ViewType = viewType
			continue
		}

		resort.RoomTypes = append(resort.RoomTypes, RoomType{
			Name:        currentRoomType.Name,
			Description: currentRoomType.Description,
			ViewType:    viewType,
		})
	}

	return nil
}
