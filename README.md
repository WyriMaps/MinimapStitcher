MinimapSticher
==============

MinimapSticher stitches exported minimap files from World of Warcraft. It ignores the WMO directory and when noLiquid files are found it makes a image with and without them. It is assume that the `BLP` images have been converted to `PNG` images.

## Installing

Install Minimap Stitcher with:

    go get github.com/WyriMaps/MinimapStitcher

## Example

```go
package main

import stitcher "github.com/WyriMaps/MinimapStitcher"
import "runtime"

func main () {
	runtime.GOMAXPROCS(runtime.NumCPU())
	stitcher.Stitch(report, "./Minimaps/", "./StitchedMaps/")
}

```

