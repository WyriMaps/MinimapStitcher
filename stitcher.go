package MinimapStitcher

import (
	"fmt"
	"runtime"
	"os"
	"log"
	"sync"
	"io/ioutil"
	"strings"
	"strconv"
	"image/draw"
	"image"
	"image/png"
)

const (
	TILE_SIZE = 256
)

func Stitch(sourceDirectory string, destinationDirectory string) {
	var wg sync.WaitGroup
	tasks := make(chan [4]string)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			for arguments := range tasks {
				compileMinimap(tasks, arguments[0],arguments[1],arguments[2],arguments[3])
			}
			wg.Done()
		}()
	}

	files, _ := ioutil.ReadDir(sourceDirectory)
	for _, f := range files {
		fd, err := os.Open(sourceDirectory + f.Name())
		if err != nil {
			fmt.Println(err)
			return
		}
		fi, err := fd.Stat()
		if err != nil {
			fmt.Println(err)
			return
		}
		mode := fi.Mode()
		if mode.IsDir() {
			if f.Name() != "WMO" {
				var array [4]string
				array[0] = destinationDirectory + f.Name()
				array[1] = sourceDirectory + f.Name()
				array[2] = f.Name()
				array[3] = "false"
				tasks <- array
			}
		}
		defer fd.Close()
	}

	wg.Wait()
}

func compileMinimap(tasks chan [4]string, resultFileName string, sourceDirectory string, minimapName string, noLiquidString string) {
	var falseString = "false"
	var foundNoLiquid = false
	var tiles = make(map[string]string);
	files, _ := ioutil.ReadDir(sourceDirectory)
	for _, f := range files {
		var fullFileName = sourceDirectory + "/" + f.Name()
		if strings.Contains(f.Name(), ".png") {
			var fName = strings.TrimRight(f.Name(), ".png")
			var fNameNoLiquid = strings.Contains(fName, "noLiquid")
			if fNameNoLiquid && noLiquidString == falseString {
				foundNoLiquid = true
			} else if !fNameNoLiquid && noLiquidString == falseString {
				tiles[strings.TrimLeft(fName, "map")] = fullFileName
			} else if fNameNoLiquid && noLiquidString != falseString {
				tiles[strings.TrimLeft(fName, "noLiquid_map")] = fullFileName
			} else if !fNameNoLiquid && noLiquidString != falseString {
				var trimmerFName = strings.TrimLeft(fName, "map")
				if _, ok := tiles[trimmerFName]; !ok {
					tiles[trimmerFName] = fullFileName
				}
			}
		}
	}

	if foundNoLiquid && noLiquidString == falseString {
		var array [4]string
		array[0] = resultFileName + "NoLiquid"
		array[1] = sourceDirectory
		array[2] = minimapName
		array[3] = "true"
		tasks <- array
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

		applyTileToImage(m, fileName, x, y)
	}

	toimg, _ := os.Create(resultFileName + ".png")
	png.Encode(toimg, m)

	defer toimg.Close()
}

func applyTileToImage(m *image.RGBA, fileName string, x int, y int) {
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
