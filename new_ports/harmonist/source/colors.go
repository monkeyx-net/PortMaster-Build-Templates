package main

import (
	"codeberg.org/anaseto/gruid"
)

// Thoses are the colors of the main palette. They are given 16-palette color
// numbers compatible with terminals, though they are then mapped to more
// precise colors depending on options and the driver. Dark colorscheme is
// assumed by default, but it can be changed in configuration.
const (
	ColorBackground          gruid.Color = gruid.ColorDefault // background
	ColorBackgroundSecondary gruid.Color = 1 + 0              // black
	ColorForeground          gruid.Color = gruid.ColorDefault
	ColorForegroundSecondary gruid.Color = 1 + 7  // white
	ColorForegroundEmph      gruid.Color = 1 + 15 // bright white
	ColorYellow              gruid.Color = 1 + 3
	ColorOrange              gruid.Color = 1 + 1 // red
	ColorRed                 gruid.Color = 1 + 9 // bright red
	ColorMagenta             gruid.Color = 1 + 5
	ColorViolet              gruid.Color = 1 + 12 // bright blue
	ColorBlue                gruid.Color = 1 + 4
	ColorCyan                gruid.Color = 1 + 6
	ColorGreen               gruid.Color = 1 + 2
)

// Those constants represent available styling attributes.
const (
	AttrInMap gruid.AttrMask = 1 << iota
	AttrReverse
	AttrBold
)
