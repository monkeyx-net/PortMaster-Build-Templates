//go:build js || sdl

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"

	"codeberg.org/anaseto/gruid"
)

const Tiles = true

func init() {
	configActions = append(configActions, ActionToggleTiles{})
}

// ActionToggleTiles toggles tiles/ASCII.
type ActionToggleTiles struct{}

func (a ActionToggleTiles) Handle(md *model) (gruid.Effect, bool) {
	GameConfig.Tiles = !GameConfig.Tiles
	err := SaveConfig()
	if err != nil {
		md.g.Logf("Error saving config changes: %v", err)
	}
	clearCache()
	eff := gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
	md.mode = modeNormal
	return eff, false
}

func (a ActionToggleTiles) String() string {
	if GameConfig.Tiles {
		return "Switch to ASCII graphics"
	}
	return "Switch to Tiles graphics"
}

// ColorToRGBA  maps to colors from selenized palette:
//
//	https://github.com/jan-warchol/selenized
func ColorToRGBA(c gruid.Color, fg bool) color.Color {
	cl := color.RGBA{}
	opaque := uint8(255)
	switch c {
	case ColorBackgroundSecondary:
		if GameConfig.DarkColors {
			cl = color.RGBA{24, 73, 86, opaque}
		} else {
			cl = color.RGBA{236, 227, 204, opaque}
		}
	case ColorRed:
		if GameConfig.DarkColors {
			cl = color.RGBA{250, 87, 80, opaque}
		} else {
			cl = color.RGBA{210, 33, 45, opaque}
		}
	case ColorGreen:
		if GameConfig.DarkColors {
			cl = color.RGBA{117, 185, 56, opaque}
		} else {
			cl = color.RGBA{72, 145, 0, opaque}
		}
	case ColorYellow:
		if GameConfig.DarkColors {
			cl = color.RGBA{219, 179, 45, opaque}
		} else {
			cl = color.RGBA{173, 137, 0, opaque}
		}
	case ColorBlue:
		if GameConfig.DarkColors {
			// cl = color.RGBA{70, 149, 247, opaque}
			// We use the bright version for blue:
			cl = color.RGBA{88, 163, 255, opaque}
		} else {
			cl = color.RGBA{0, 114, 212, opaque}
		}
	case ColorMagenta:
		if GameConfig.DarkColors {
			cl = color.RGBA{242, 117, 190, opaque}
		} else {
			cl = color.RGBA{202, 72, 152, opaque}
		}
	case ColorCyan:
		if GameConfig.DarkColors {
			cl = color.RGBA{65, 199, 185, opaque}
		} else {
			cl = color.RGBA{0, 156, 143, opaque}
		}
	case ColorOrange:
		if GameConfig.DarkColors {
			cl = color.RGBA{237, 134, 73, opaque}
		} else {
			cl = color.RGBA{194, 93, 30, opaque}
		}
	case ColorViolet:
		if GameConfig.DarkColors {
			cl = color.RGBA{175, 136, 235, opaque}
		} else {
			cl = color.RGBA{135, 98, 198, opaque}
		}
	case ColorForegroundEmph:
		if GameConfig.DarkColors {
			cl = color.RGBA{202, 216, 217, opaque}
		} else {
			cl = color.RGBA{58, 77, 83, opaque}
		}
	case ColorForegroundSecondary:
		if GameConfig.DarkColors {
			cl = color.RGBA{114, 137, 143, opaque}
		} else {
			cl = color.RGBA{144, 153, 149, opaque}
		}
	default:
		if fg {
			if GameConfig.DarkColors {
				cl = color.RGBA{173, 188, 188, opaque}
			} else {
				cl = color.RGBA{83, 103, 109, opaque}
			}
			break
		}
		if GameConfig.DarkColors {
			cl = color.RGBA{16, 60, 72, opaque}
		} else {
			cl = color.RGBA{251, 243, 219, opaque}
		}
	}
	return cl
}

var TileImgs map[string][]byte

var MapNames = map[rune]string{
	'!': "totem",
	'"': "foliage",
	'#': "wall",
	'%': "food",
	'&': "menhir",
	'.': "ground",
	'=': "rune",
	'>': "portal",
	'@': "player",
	'^': "rubble",
	'¤': "frontier",
	'§': "cloud",
	'Φ': "magic",
	'√': "hit",
	'∞': "kill",
	'◊': "glass",
	'☼': "orb",
	'♣': "tree",
	'♪': "sound",
	'♫': "footsteps",
}

var LetterNames = map[rune]string{
	' ':  "space",
	'!':  "totem",
	'"':  "quotes",
	'#':  "wall",
	'%':  "food",
	'&':  "menhir",
	'(':  "lparen",
	')':  "rparen",
	'*':  "asterisc",
	'+':  "plus",
	',':  "comma",
	'-':  "hyphen",
	'.':  "ground",
	'/':  "slash",
	':':  "colon",
	';':  "semicolon",
	'<':  "less",
	'=':  "rune",
	'>':  "portal",
	'?':  "interrogation",
	'@':  "player",
	'[':  "lbracket",
	'\'': "quote",
	'\\': "backslash",
	']':  "rbracket",
	'^':  "rubble",
	'{':  "lbrace",
	'|':  "pipe",
	'}':  "rbrace",
	'~':  "tilde",
	'¤':  "frontier",
	'§':  "cloud",
	'«':  "ldiag",
	'»':  "rdiag",
	'×':  "times",
	'Φ':  "magic",
	'—':  "longhyphen",
	'’':  "apostrophe",
	'“':  "lquotes",
	'”':  "rquotes",
	'•':  "tick",
	'…':  "dots",
	'←':  "larrow",
	'↑':  "uarrow",
	'→':  "rarrow",
	'↓':  "darrow",
	'√':  "hit",
	'∞':  "kill",
	'─':  "hbar",
	'│':  "vbar",
	'┌':  "boxnw",
	'┐':  "boxne",
	'└':  "boxsw",
	'┘':  "boxse",
	'┤':  "boxe",
	'◊':  "glass",
	'☼':  "orb",
	'♣':  "tree",
	'♪':  "sound",
	'♫':  "footsteps",
}

type monochromeTileManager struct{}

func (tm *monochromeTileManager) TileSize() gruid.Point {
	return gruid.Point{16, 24}
}

func base64pngToRGBA(bs []byte) (*image.RGBA, error) {
	buf := make([]byte, len(bs))
	_, err := base64.StdEncoding.Decode(buf, bs)
	if err != nil {
		return nil, fmt.Errorf("could not decode base64-encoded png: %v", err)
	}
	br := bytes.NewReader(buf)
	img, err := png.Decode(br)
	if err != nil {
		return nil, fmt.Errorf("could not decode png: %v", err)
	}
	rect := img.Bounds()
	rgbaimg := image.NewRGBA(rect)
	draw.Draw(rgbaimg, rect, img, rect.Min, draw.Src)
	return rgbaimg, nil
}

func (tm *monochromeTileManager) GetImage(gc gruid.Cell) image.Image {
	var pngImg []byte
	hastile := false
	if gc.Style.Attrs&AttrInMap != 0 && GameConfig.Tiles {
		pngImg = TileImgs["map-notile"]
		if im, ok := TileImgs["map-"+string(gc.Rune)]; ok {
			pngImg = im
			hastile = true
		} else if im, ok := TileImgs["map-"+MapNames[gc.Rune]]; ok {
			pngImg = im
			hastile = true
		}
	}
	if !hastile {
		pngImg = TileImgs["map-notile"]
		if im, ok := TileImgs["letter-"+string(gc.Rune)]; ok {
			pngImg = im
		} else if im, ok := TileImgs["letter-"+LetterNames[gc.Rune]]; ok {
			pngImg = im
		}
	}
	rgbaimg, err := base64pngToRGBA(pngImg)
	if err != nil {
		log.Printf("Rune %s: %v", string(gc.Rune), err)
		return image.Black
	}
	bgc := ColorToRGBA(gc.Style.Bg, false)
	fgc := ColorToRGBA(gc.Style.Fg, true)
	if gc.Style.Attrs&AttrReverse != 0 {
		fgc, bgc = bgc, fgc
	}
	rect := rgbaimg.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			c := rgbaimg.At(x, y)
			r, _, _, _ := c.RGBA()
			if r == 0 {
				rgbaimg.Set(x, y, bgc)
			} else {
				rgbaimg.Set(x, y, fgc)
			}
		}
	}
	return rgbaimg
}
