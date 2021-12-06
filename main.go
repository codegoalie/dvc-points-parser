package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/codegoalie/dvc-points-parser/resorts"
)

var parsedResorts []resorts.Resort

func main() {

	//root := "vgf/"
	root := "converted-charts/2023/"
	// // root := "converted-charts/2022/FINAL_2022_DVC_VGF_Pt_Chts.pdf.txt"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		resort, err := resorts.ParseFile(path)
		if err != nil {
			err = fmt.Errorf("failed to parse files: %w", err)
			log.Println(err)
			return err
		}

		fmt.Println(len(resort.RoomTypes))

		return nil
	})
	if err != nil {
		panic(err)
	}
}
