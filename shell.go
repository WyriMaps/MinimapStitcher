package main

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	"strings"
	"strconv"
	"image/draw"
	"image"
	"image/png"
)

const (
	SRC_DIR = "./Minimaps/"
	DST_DIR = "./StitchedMaps/"
	TILE_SIZE = 256
)

func main() {
	files, _ := ioutil.ReadDir(SRC_DIR)
	for _, f := range files {
		compileMinimap(DST_DIR + f.Name(), f.Name(), false)
	}
}

func compileMinimap(resultFileName string, minimapName string, noLiquid bool) {
	var foundNoLiquid = false
	var tiles = make(map[string]string);
	files, _ := ioutil.ReadDir(SRC_DIR + minimapName)
	for _, f := range files {
		var fullFileName = SRC_DIR + minimapName + "/" + f.Name()
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
	var hc, lc, hr, lr = calculateMinimapSize(tiles)
	var files, width, height = calculateMinimapTilePlacement(tiles, hc, lc, hr, lr)
	createMinimapImage(resultFileName, width, height, files)
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

func calculateMinimapTilePlacement(tiles map[string]string, hc int, lc int, hr int, lr int) (map[string]string, int, int) {
	var files = make(map[string]string)

	var width = 0;
	var height = 0;

	for i := hc; i < lc; i++ {
		width += TILE_SIZE
		for j := hr; j < lr; j++ {
			if i == hc {
				height += TILE_SIZE
			}

			var si = strconv.Itoa(i)
			var sj = strconv.Itoa(j)
			var oi = strconv.Itoa((i - hc) * TILE_SIZE)
			var oj = strconv.Itoa((j - hr) * TILE_SIZE)
			if _, ok := tiles[si + "_" + sj]; ok {
				files[oi + "_" + oj] = tiles[si + "_" + sj]
			}
		}
	}

	return files, width, height
}

func createMinimapImage(resultFileName string, width int, height int, files map[string]string) {
	m := image.NewRGBA(image.Rect(0, 0, width, height))
	for coords, fileName := range files {
		var coordParts = strings.Split(coords, "_")
		var x, _ = strconv.Atoi(coordParts[0])
		var y, _ = strconv.Atoi(coordParts[1])

		tile, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer tile.Close()

		tileImage, err := png.Decode(tile)
		if err != nil {
			fmt.Printf("%v", fileName)
			log.Fatal(err)
		}

		draw.Draw(m, image.Rect(x, y, x + TILE_SIZE, y + TILE_SIZE), tileImage, image.Point{0,0}, draw.Src)
	}

	toimg, _ := os.Create(resultFileName + ".png")
	png.Encode(toimg, m)
	defer toimg.Close()
}
