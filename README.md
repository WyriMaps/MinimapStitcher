MinimapSticher
==============

World of Warcraft MinimapSticher

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

