package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	imgsPath := "."
	flag.Parse()
	if len(flag.Args()) > 0 {
		if len(flag.Args()) == 1 {
			imgsPath = flag.Arg(0)
		} else {
			log.Fatal("Too many arguments")
			os.Exit(1)
		}
	}
	filepath.Walk(imgsPath, walkpath)

}

type TransparencyTestResult struct {
	hasTransparency    bool
	percentTransparent float64
	croppableColumns   []int
	croppableRows      []int
}

func walkpath(path string, f os.FileInfo, err error) error {
	splitPath := strings.Split(path, ".")
	if splitPath[len(splitPath)-1] == "png" {
		transparencyTest, decodeErr := testFile(path)
		if decodeErr != nil {
			fmt.Printf("Bad file: %s\n", path)
		}
		fmt.Printf("%s,%t,%d,%d,%f\n", path, transparencyTest.hasTransparency, len(transparencyTest.croppableColumns), len(transparencyTest.croppableRows), transparencyTest.percentTransparent)
	}
	return nil
}

func testFile(fileName string) (TransparencyTestResult, error) {
	pngFile, err := os.Open(fileName)
	defer pngFile.Close()

	if err != nil {
		panic("File does not exist")
	}

	img, _, imgErr := image.Decode(pngFile)

	if imgErr != nil {
		return TransparencyTestResult{}, imgErr
	}

	transparentPixels := 0
	hasTransparency := false
	imgBounds := img.Bounds()
	croppableColumns := make([]int, 0, imgBounds.Max.Y)
	croppableRows := make([]int, 0, imgBounds.Max.Y)
	totalPixels := (imgBounds.Max.Y - imgBounds.Min.Y + 1) * (imgBounds.Max.X - imgBounds.Min.X + 1)
	for i := imgBounds.Min.Y; i <= imgBounds.Max.Y; i++ {
		for j := imgBounds.Min.X; j <= imgBounds.Max.X; j++ {
			transparentColumn := true
			px := img.At(i, j)
			_, _, _, a := px.RGBA()
			if a < 65535 {
				hasTransparency = true
				transparentPixels += 1
			}
			if a > 0 {
				transparentColumn = false
			}
			if j == imgBounds.Max.X && transparentColumn {
				croppableColumns = append(croppableColumns, i)
			}
		}
	}

	for j := imgBounds.Min.X; j <= imgBounds.Max.X; j++ {
		for i := imgBounds.Min.Y; i <= imgBounds.Max.Y; i++ {
			transparentRow := true
			px := img.At(i, j)
			_, _, _, a := px.RGBA()
			if a < 65535 {
				hasTransparency = true
			}
			if a > 0 {
				transparentRow = false
			}
			if j == imgBounds.Max.X && transparentRow {
				croppableRows = append(croppableRows, j)
			}
		}
	}

	percentTransparent := float64(transparentPixels) / float64(totalPixels)

	result := TransparencyTestResult{
		hasTransparency:    hasTransparency,
		percentTransparent: percentTransparent,
		croppableColumns:   croppableColumns,
		croppableRows:      croppableRows,
	}

	return result, nil
}
