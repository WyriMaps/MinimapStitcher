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
	"encoding/json"
)

const (
	TILE_SIZE = 256
)

type reportCallbackMessage map[string]string
type reportCallback func(reportCallbackMessage)
type reportCallbackWrapper func(string, reportCallbackMessage)

func Stitch(sourceDirectory string, destinationDirectory string) {
	var callback = func(message reportCallbackMessage) {
		b, _ := json.Marshal(message)
		os.Stdout.Write(b)
		print("\r\n")
	}

	var wg sync.WaitGroup;
	tasks := make(chan [5]string);

	setupWaitGroup(wg, tasks, callback);
	listMapsFound(callback, sourceDirectory);
	addMapsToWaitGroup(tasks, sourceDirectory, destinationDirectory);

	wg.Wait()
}

func listMapsFound(callback reportCallback, sourceDirectory string) {
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
				message := make(map [string]string)
				message["minimap"] = f.Name()
				message["type"] = "found"
				callback(message)
			}
		}
		defer fd.Close()
	}
}

func setupWaitGroup(wg sync.WaitGroup, tasks chan [5]string, callback reportCallback, ) {
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			for arguments := range tasks {
				message := make(map [string]string)
				message["minimap"] = arguments[4]
				message["tile"] = "0"
				message["tiles"] = "0"
				var callbackWrapper = func(messageText string, extras reportCallbackMessage) {
					message["type"] = messageText

					if _, ok := extras["tile"]; ok {
						message["tile"] = extras["tile"]
					}
					if _, ok := extras["tiles"]; ok {
						message["tiles"] = extras["tiles"]
					}

					callback(message)
				}
				callbackWrapper("start_compile", make(map [string]string))
				compileMinimap(callbackWrapper, tasks, arguments[0], arguments[1], arguments[2], arguments[3])
				callbackWrapper("complete_compile", make(map [string]string))
			}
			wg.Done()
		}()
	}
}

func addMapsToWaitGroup(tasks chan [5]string, sourceDirectory string, destinationDirectory string) {
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
				var array [5]string
				array[0] = destinationDirectory + f.Name()
				array[1] = sourceDirectory + f.Name()
				array[2] = f.Name()
				array[3] = "false"
				array[4] = f.Name()
				tasks <- array
			}
		}
		defer fd.Close()
	}
}

func compileMinimap(callback reportCallbackWrapper, tasks chan [5]string, resultFileName string, sourceDirectory string, minimapName string, noLiquidString string) {
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
		var array [5]string
		array[0] = resultFileName + "NoLiquid"
		array[1] = sourceDirectory
		array[2] = minimapName
		array[3] = "true"
		array[4] = minimapName + "NoLiquid"
		tasks <- array
	}

	callback("start_build", make(map [string]string))
	buildMinimap(callback, resultFileName, tiles)
	callback("start_build", make(map [string]string))
}

func buildMinimap(callback reportCallbackWrapper, resultFileName string, tiles map[string]string)  {
	callback("calculate_minimap_size", make(map [string]string))
	var hc, lc, hr, lr = calculateMinimapSize(tiles)

	callback("calculate_minimap_tileplacement", make(map [string]string))
	var files, width, height = calculateMinimapTilePlacement(tiles, hc, lc, hr, lr)

	createMinimapImage(callback, resultFileName, width, height, files)
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

func createMinimapImage(callback reportCallbackWrapper, resultFileName string, width int, height int, files map[string]string) {
	extras := make(map [string]string)
	extras["tiles"] = strconv.Itoa(len(files))
	callback("start_stitch", extras)

	m := image.NewRGBA(image.Rect(0, 0, width, height))
	for coords, fileName := range files {
		extras := make(map [string]string)
		extras["tile"] = coords
		callback("stitch_tile", extras)

		var coordParts = strings.Split(coords, "_")
		var x, _ = strconv.Atoi(coordParts[0])
		var y, _ = strconv.Atoi(coordParts[1])

		applyTileToImage(m, fileName, x, y)
	}

	toimg, _ := os.Create(resultFileName + ".png")
	png.Encode(toimg, m)

	defer toimg.Close()

	callback("finish_stitch", make(map [string]string))
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
