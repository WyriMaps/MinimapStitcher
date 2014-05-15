MinimapSticher
==============

MinimapSticher stitches exported minimap files from World of Warcraft. It ignores the WMO directory and when noLiquid files are found it makes a image with and without them.

## Installing

Install Minimap Stitcher with:

    go get github.com/WyriMaps/MinimapStitcher

## Example

```go
package main

import stitcher "github.com/WyriMaps/MinimapStitcher"

func main () {
	stitcher.Stitch("./Minimaps/", "./StitchedMaps/")
}

```

