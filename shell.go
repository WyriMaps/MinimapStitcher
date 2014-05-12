package main

import (
	"fmt"
	//"os"
	"io/ioutil"
	"strings"
	"strconv"
)

const (
	BASE_DIR = "./Minimaps/"
)

func main() {
	files, _ := ioutil.ReadDir(BASE_DIR)
	for _, f := range files {
		compileMinimap(BASE_DIR + f.Name(), f.Name(), false)
	}
}

func compileMinimap(resultFileName string, minimapName string, noLiquid bool) {
	var foundNoLiquid = false
	var tiles = make(map[string]string);
	files, _ := ioutil.ReadDir(BASE_DIR + minimapName)
	for _, f := range files {
		var fullFileName = BASE_DIR + minimapName + "/" + f.Name()
		if strings.Contains(f.Name(), ".png") {
			var fName = strings.TrimRight(f.Name(), ".png")
			var fNameNoLiquid = strings.Contains(fName, "noLiquid")
			if fNameNoLiquid && !noLiquid {
				foundNoLiquid = true
			} else if !fNameNoLiquid && !noLiquid {
				tiles[strings.TrimLeft(fName, "map")] = fullFileName
			} else if fNameNoLiquid && noLiquid {
				tiles[strings.TrimLeft(fName, "noLiquid_map")] = fullFileName
			} else if !fNameNoLiquid && noLiquid {
				var trimmerFName = strings.TrimLeft(fName, "map")
				if _, ok := tiles[trimmerFName]; !ok {
					tiles[trimmerFName] = fullFileName
				}
			}
		}
	}

	if foundNoLiquid && !noLiquid {
		compileMinimap(resultFileName + "NoLiquid", minimapName, true)
	}

	buildMinimap(resultFileName, tiles)
}

func buildMinimap(resultFileName string, tiles map[string]string)  {
	//fmt.Printf("%v", tiles)
	fmt.Println("")
	fmt.Printf("%v", resultFileName)
	fmt.Println("")
	var hc, lc, hr, lr = calculateMinimapSize(tiles)
	fmt.Printf("%v", hc)
	fmt.Println("")
	fmt.Printf("%v", lc)
	fmt.Println("")
	fmt.Printf("%v", hr)
	fmt.Println("")
	fmt.Printf("%v", lr)
	fmt.Println("")
	var files = calculateMinimapTilePlacement(tiles, hc, lc, hr, lr)
	fmt.Printf("%v", files)
	fmt.Println("")
}

func calculateMinimapSize(tiles map[string]string) (hc, lc, hr, lr int) {
	hc = 1000
	lc = 0
	hr = 1000
	lr = 0
	for tile, _ := range tiles {
		var tileParts = strings.Split(tile, "_")
		var col, _ = strconv.Atoi(tileParts[0])
		var row, _ = strconv.Atoi(tileParts[1])
		if hc > col {
			hc = col
		}
		if lc < col {
			lc = col
		}
		if hr > row {
			hr = row
		}
		if lr < row {
			lr = row
		}
	}
	return
}

func calculateMinimapTilePlacement(tiles map[string]string, hc int, lc int, hr int, lr int) (map[string]string) {
	var files = make(map[string]string)

	var width = 0;
	var height = 0;

	for i := hc; i < lc; i++ {
		width += 256
		for j := hr; j < lr; j++ {
			if i == hc {
				height += 256
			}

			var si = strconv.Itoa(i)
			var sj = strconv.Itoa(j)
			fmt.Printf("%v", si + "_" + sj)
			if _, ok := tiles[si + "_" + sj]; !ok {
				files[si + "_" + sj] = tiles[si + "_" + sj]
			} else {
				files[si + "_" + sj] = "empty.png"
			}

			fmt.Printf("%v", files)
		}
	}

	return files
}
