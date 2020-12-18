package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/codegoalie/dvc-points-parser/resorts"
	gonanoid "github.com/matoous/go-nanoid"
	_ "github.com/mattn/go-sqlite3"
)

var parsedResorts []resorts.Resort

func main() {
	dbFile := "./dvc-points.sqlite3"
	os.Remove(dbFile)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createResorts(db)
	if err != nil {
		log.Fatal(err)
	}

	err = createRoomTypes(db)
	if err != nil {
		log.Fatal(err)
	}

	err = createPoints(db)
	if err != nil {
		log.Fatal(err)
	}

	//root := "vgf/"
	root := "converted-charts/"
	// root := "converted-charts/2022/FINAL_2022_DVC_VGF_Pt_Chts.pdf.txt"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		resort, err := resorts.ParseFile(path)
		if err != nil {
			err = fmt.Errorf("failed to parse files: %w", err)
			log.Println(err)
			return err
		}

		err = insertResort(db, resort)
		if err != nil {
			err = fmt.Errorf("failed to insert resort: %w", err)
			log.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	// for _, resort := range parsedResorts {
	// 	if strings.Contains(resort.Name, "Floridian") {
	// 		spew.Dump(resort)
	// 		return
	// 	}
	// }
	// spew.Dump(resorts[0].Name, resorts[0].RoomTypes[0])

	// for _, resort := range parsedResorts {
	// 	if strings.Contains(resort.Name, "Floridian") {
	// 		fmt.Println(resort.RoomTypes[0].PointChart[0].CheckInAt)
	// 	}
	// }

	rows, err := db.Query(`
  SELECT resorts.name, room_types.name, view_type, SUM(amount)
    FROM points
    JOIN room_types ON points.room_type_id = room_types.id
    JOIN resorts ON room_types.resort_id = resorts.id
GROUP BY resorts.id, room_types.id
	;
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var resort string
		var name string
		var viewType string
		var points int
		err = rows.Scan(&resort, &name, &viewType, &points)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resort, name, viewType, points)
	}
	err = rows.Err()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal(err)
	}

	// stmt, err := db.Prepare("select name from resorts where id = ?")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stmt.Close()
	// var name string
	// err = stmt.QueryRow("1").Scan(&name)
	// if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 	log.Fatal(err)
	// }
	// fmt.Println(name)
}

func createResorts(db *sql.DB) error {
	sqlStmt := `
	create table resorts (id text not null primary key, name text);
	delete from resorts;
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		err = fmt.Errorf("failed to create resorts table: %w", err)
		return err
	}

	return nil
}

func insertResort(db *sql.DB, resort resorts.Resort) error {
	stmt, err := db.Prepare("insert into resorts(id, name) values(?, ?)")
	if err != nil {
		err = fmt.Errorf("failed to prepare insert resorts statement: %w", err)
		return err
	}
	defer stmt.Close()

	// findStmt, err := db.Prepare("select id from resorts where name = ?")
	// if err != nil {
	// 	err = fmt.Errorf("failed to prepare select resort statement: %w", err)
	// 	return err
	// }
	// defer findStmt.Close()

	var resortID string
	resortID, err = gonanoid.Nanoid()
	err = db.QueryRow("select id from resorts where name = ?", resort.Name).Scan(&resortID)
	fmt.Println("ID", resortID, err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resortID, _ = gonanoid.Nanoid()
			_, err = stmt.Exec(resortID, resort.Name)
			if err != nil {
				err = fmt.Errorf("failed to insert resort record: %w", err)
				return err
			}
		} else {
			err = fmt.Errorf("failed to fetch existing resort: %w", err)
			return err
		}
	}
	fmt.Println("ID2", resortID, err)

	err = insertRoomTypes(db, resortID, resort.RoomTypes)
	if err != nil {
		return err
	}

	return nil
}

func createRoomTypes(db *sql.DB) error {
	sqlStmt := `
	create table room_types (id text not null primary key, resort_id text, name text, description text, view_type text);
	delete from room_types;
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		err = fmt.Errorf("failed to create room_types table: %w", err)
		return err
	}

	return nil
}

func insertRoomTypes(db *sql.DB, resortID string, roomTypes []resorts.RoomType) error {
	stmt, err := db.Prepare("insert into room_types(id, resort_id, name, description, view_type) values(?, ?, ?, ?, ?)")
	if err != nil {
		err = fmt.Errorf("failed to prepare insert room_types statement: %w", err)
		return err
	}
	defer stmt.Close()

	for _, roomType := range roomTypes {

		var roomTypeID string
		err = db.QueryRow(
			"select id from room_types where resort_id = ? and name = ? and view_type = ? and description = ?",
			resortID,
			roomType.Name,
			roomType.ViewType,
			roomType.Description,
		).Scan(&roomTypeID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				err = fmt.Errorf("failed to fetch existing roomType: %w", err)
				return err
			}

			roomTypeID, err = gonanoid.Nanoid()
			fmt.Println("creating room type", roomTypeID)
			if err != nil {
				err = fmt.Errorf("failed to generate roomTypeID: %w", err)
				return err
			}

			_, err = stmt.Exec(roomTypeID, resortID, roomType.Name, roomType.Description, roomType.ViewType)
			if err != nil {
				err = fmt.Errorf("failed to insert room_types record: %w", err)
				return err
			}
		}

		err = insertPoints(db, roomTypeID, roomType.PointChart)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPoints(db *sql.DB) error {
	sqlStmt := `
	create table points (id text not null primary key, room_type_id text, stay_on date, amount int);
	delete from points;
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		err = fmt.Errorf("failed to create points table: %w", err)
		return err
	}

	return nil

}

func insertPoints(db *sql.DB, roomTypeID string, pointChart []resorts.PointBlock) error {
	stmt, err := db.Prepare("insert into points(id, room_type_id, stay_on, amount) values(?, ?, ?, ?)")
	if err != nil {
		err = fmt.Errorf("failed to prepare insert points statement: %w", err)
		return err
	}
	defer stmt.Close()

	var points int
	for _, pointBlock := range pointChart {
		for {
			if pointBlock.CheckInAt.After(pointBlock.CheckOutAt) {
				break
			}

			pointsID, err := gonanoid.Nanoid()
			if err != nil {
				err = fmt.Errorf("failed to genrate points ID: %w", err)
				return err
			}
			points = pointBlock.WeekdayPoints
			if pointBlock.CheckInAt.Weekday() == time.Saturday || pointBlock.CheckInAt.Weekday() == time.Friday {
				points = pointBlock.WeekendPoints
			}
			_, err = stmt.Exec(pointsID, roomTypeID, pointBlock.CheckInAt.Format("2006-01-02"), points)
			if err != nil {
				err = fmt.Errorf("failed to insert points record: %w", err)
				return err
			}

			pointBlock.CheckInAt = pointBlock.CheckInAt.AddDate(0, 0, 1)
		}
	}

	return nil
}
