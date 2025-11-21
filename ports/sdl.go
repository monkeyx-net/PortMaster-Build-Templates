//go:build sdl

package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
)

var driver gruid.Driver
var isFullscreen bool

func initDriver(fullscreen bool, sx, sy float64) {
	isFullscreen = fullscreen
	icon, err := base64pngToRGBA(TileImgs["favicon"])
	if err != nil {
		log.Printf("decoding window icon: %v", err)
	}
	dr := sdl.NewDriver(sdl.Config{
		TileManager: &monochromeTileManager{},
		Fullscreen:  fullscreen,
		WindowTitle: "Shamogu",
		WindowIcon:  icon,
	})
	if sx != 1 || sy != 1 {
		dr.SetScale(float32(sx), float32(sy))
	}
	driver = dr
}

func clearCache() {
	dr := driver.(*sdl.Driver)
	dr.ClearCache()
}
