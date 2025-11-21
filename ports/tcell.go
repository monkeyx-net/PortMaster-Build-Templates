//go:build !sdl && !js

package main

import (
	"codeberg.org/anaseto/gruid"
	tcell "codeberg.org/anaseto/gruid-tcell"
	tc "github.com/gdamore/tcell/v2"
)

const Tiles = false

var driver gruid.Driver

func initDriver(_ bool, _, _ float64) {
	st := styler{}
	dr := tcell.NewDriver(tcell.Config{StyleManager: st})
	//dr.PreventQuit()
	driver = dr
}

// styler implements the tcell.StyleManager interface.
type styler struct{}

func (sty styler) GetStyle(cst gruid.Style) tc.Style {
	st := tc.StyleDefault
	switch ColorMode {
	case ColorMode256:
		cst.Fg = map16ColorTo256(cst.Fg, true)
		cst.Bg = map16ColorTo256(cst.Bg, false)
		st = st.Background(tc.ColorValid + tc.Color(cst.Bg)).Foreground(tc.ColorValid + tc.Color(cst.Fg))
	case ColorMode24bit:
		st = st.Foreground(mapColor2TrueColor(cst.Fg, true))
		st = st.Background(mapColor2TrueColor(cst.Bg, false))
	default:
		// ColorMode16 & ColorMode8
		if !GameConfig.DarkColors {
			cst.Fg = map16ColorToLight(cst.Fg)
			cst.Bg = map16ColorToLight(cst.Bg)
		}
		if ColorMode == ColorMode8 {
			cst.Fg = map16ColorTo8Color(cst.Fg)
			cst.Bg = map16ColorTo8Color(cst.Bg)
		}
		fg := tc.Color(cst.Fg)
		bg := tc.Color(cst.Bg)
		if cst.Bg == gruid.ColorDefault {
			st = st.Background(tc.ColorDefault)
		} else {
			st = st.Background(tc.ColorValid + bg - 1)
		}
		if cst.Fg == gruid.ColorDefault {
			st = st.Foreground(tc.ColorDefault)
		} else {
			st = st.Foreground(tc.ColorValid + fg - 1)
		}
	}
	if cst.Attrs&AttrReverse != 0 {
		st = st.Reverse(true)
	}
	if cst.Attrs&AttrBold != 0 {
		st = st.Bold(true)
	}
	return st
}

func map16ColorTo8Color(c gruid.Color) gruid.Color {
	if c >= 1+8 {
		c -= 8
	}
	return c
}

func map16ColorToLight(c gruid.Color) gruid.Color {
	switch c {
	case ColorBackgroundSecondary:
		return ColorForegroundEmph
	case ColorForegroundSecondary:
		return ColorBackgroundSecondary
	case ColorForegroundEmph:
		return ColorBackgroundSecondary
	default:
		return c
	}
}

// xterm solarized colors: http://ethanschoonover.com/solarized
const (
	Color256Base03  gruid.Color = 234
	Color256Base02  gruid.Color = 235
	Color256Base01  gruid.Color = 240
	Color256Base00  gruid.Color = 241 // for dark on light background
	Color256Base0   gruid.Color = 244
	Color256Base1   gruid.Color = 245
	Color256Base2   gruid.Color = 254
	Color256Base3   gruid.Color = 230
	Color256Yellow  gruid.Color = 136
	Color256Orange  gruid.Color = 166
	Color256Red     gruid.Color = 160
	Color256Magenta gruid.Color = 125
	Color256Violet  gruid.Color = 61
	Color256Blue    gruid.Color = 33
	Color256Cyan    gruid.Color = 37
	Color256Green   gruid.Color = 64
)

func map16ColorTo256(c gruid.Color, fg bool) gruid.Color {
	switch c {
	case ColorBackground:
		if fg {
			if GameConfig.DarkColors {
				return Color256Base0
			}
			return Color256Base00
		}
		if GameConfig.DarkColors {
			return Color256Base03
		}
		return Color256Base3
	case ColorBackgroundSecondary:
		if GameConfig.DarkColors {
			return Color256Base02
		}
		return Color256Base2
	case ColorForegroundEmph:
		if GameConfig.DarkColors {
			return Color256Base1
		}
		return Color256Base01
	case ColorForegroundSecondary:
		if GameConfig.DarkColors {
			return Color256Base01
		}
		return Color256Base1
	case ColorYellow:
		return Color256Yellow
	case ColorOrange:
		return Color256Orange
	case ColorRed:
		return Color256Red
	case ColorMagenta:
		return Color256Magenta
	case ColorViolet:
		return Color256Violet
	case ColorBlue:
		return Color256Blue
	case ColorCyan:
		return Color256Cyan
	case ColorGreen:
		return Color256Green
	default:
		return c
	}
}

// mapColor2TrueColor  maps to colors from selenized palette:
//
//	https://github.com/jan-warchol/selenized
func mapColor2TrueColor(c gruid.Color, fg bool) tc.Color {
	var cl tc.Color
	switch c {
	case ColorBackgroundSecondary:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(24, 73, 86)
		} else {
			cl = tc.NewRGBColor(236, 227, 204)
		}
	case ColorRed:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(250, 87, 80)
		} else {
			cl = tc.NewRGBColor(210, 33, 45)
		}
	case ColorGreen:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(117, 185, 56)
		} else {
			cl = tc.NewRGBColor(72, 145, 0)
		}
	case ColorYellow:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(219, 179, 45)
		} else {
			cl = tc.NewRGBColor(173, 137, 0)
		}
	case ColorBlue:
		if GameConfig.DarkColors {
			// cl = color.RGBA{70, 149, 247, opaque}
			// We use the bright version for blue:
			cl = tc.NewRGBColor(88, 163, 255)
		} else {
			cl = tc.NewRGBColor(0, 114, 212)
		}
	case ColorMagenta:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(242, 117, 190)
		} else {
			cl = tc.NewRGBColor(202, 72, 152)
		}
	case ColorCyan:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(65, 199, 185)
		} else {
			cl = tc.NewRGBColor(0, 156, 143)
		}
	case ColorOrange:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(237, 134, 73)
		} else {
			cl = tc.NewRGBColor(194, 93, 30)
		}
	case ColorViolet:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(175, 136, 235)
		} else {
			cl = tc.NewRGBColor(135, 98, 198)
		}
	case ColorForegroundEmph:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(202, 216, 217)
		} else {
			cl = tc.NewRGBColor(58, 77, 83)
		}
	case ColorForegroundSecondary:
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(114, 137, 143)
		} else {
			cl = tc.NewRGBColor(144, 153, 149)
		}
	default:
		if fg {
			if GameConfig.DarkColors {
				cl = tc.NewRGBColor(173, 188, 188)
			} else {
				cl = tc.NewRGBColor(83, 103, 109)
			}
			break
		}
		if GameConfig.DarkColors {
			cl = tc.NewRGBColor(16, 60, 72)
		} else {
			cl = tc.NewRGBColor(251, 243, 219)
		}
	}
	return cl
}

func clearCache() {
	// do nothing
}
