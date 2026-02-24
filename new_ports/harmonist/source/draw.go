package main

import (
	"fmt"
	//"log"
	//"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// Draw implements Draw() for gruid.Model.
func (md *model) Draw() gruid.Grid {
	if !md.anims.Done() {
		return md.drawAnimationFrame()
	}
	md.gd.Fill(gruid.Cell{Rune: ' '})
	switch md.mode {
	case modeQuit:
		return md.gd.Slice(gruid.Range{})
	case modeDump:
		md.gd.Copy(md.pager.Draw())
		return md.gd
	case modeWelcome:
		return drawWelcome(md.gd)
	}
	// Draw map in all other cases, as it may be covered only partially by
	// other modes.
	if md.g.win {
		md.drawWin(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	} else {
		md.drawMap(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	}
	md.log.Content = md.DrawLog()
	md.log.Draw(md.gd.Slice(md.gd.Range().Lines(0, 2)))
	switch md.mode {
	case modeNormal:
		if md.statusFocus {
			md.drawStatusDesc()
		} else if md.targ.ex.p != invalidPos {
			md.drawPosInfo()
		}
	case modePager:
		pgd := md.pager.Draw()
		rg := md.pager.View()
		sz := pgd.Size()
		if rg.Min.Y > 0 {
			pgd.Set(gruid.Point{sz.X - 1, 1}, gruid.Cell{Rune: '↑'})
		}
		if rg.Max.Y < md.pager.Lines() {
			pgd.Set(gruid.Point{sz.X - 1, sz.Y - 2}, gruid.Cell{Rune: '↓'})
		}
		md.gd.Copy(pgd)
	case modeSmallPager:
		md.gd.Slice(gruid.NewRange(10, 2, UIWidth, UIHeight-1)).Copy(md.smallPager.Draw())
	case modeMenu:
		switch md.menuMode {
		case modeInventory, modeEquip, modeEvocation:
			gd := md.menu.Draw()
			md.gd.Copy(gd)
			md.description.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
			if md.menuMode == modeEquip {
				it := md.g.Objects.Magaras[md.g.Player.P]
				stt := ui.StyledText{}.WithMarkups(Markups)
				md.equipLabel.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
				md.equipLabel.Box = &ui.Box{Title: stt.WithTextf("%s (ground)", it.String())}
				md.equipLabel.Draw(md.gd.Slice(gruid.NewRange(0, gd.Size().Y, UIWidth/2, UIHeight-1)))
			}
		case modeGameMenu, modeHelpMenu, modeSettings:
			md.gd.Copy(md.menu.Draw())
			if md.description.Content.Text() != "" {
				md.description.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
			}
		case modeKeys, modeKeysChange:
			gd := md.keysMenu.Draw()
			max := gd.Size()
			t := ui.Text("(R) reset (Enter) change").WithStyle(gruid.Style{Fg: ColorCyan})
			if md.menuMode == modeKeysChange {
				t = t.WithText(" Press new key... ")
			}
			t.Draw(gd.Slice(gd.Range().Line(max.Y-1).Shift(2, 0, 0, 0)))
			md.gd.Copy(gd)
		}
	}
	md.gd.Slice(md.gd.Range().Line(UIHeight - 1)).Copy(md.status.Draw())
	return md.gd
}

func (md *model) drawAnimationFrame() gruid.Grid {
	if !md.anims.draw {
		return md.gd.Slice(gruid.Range{})
	}
	gd := md.gd
	for _, fc := range md.anims.frames[0].Cells {
		gd.Set(fc.P, fc.Cell)
	}
	md.anims.frames = md.anims.frames[1:]
	return gd
}

func (md *model) drawStatusDesc() {
	rg := md.status.ActiveBounds()
	x := (rg.Min.X + rg.Max.X) / 2
	sz := md.statusDesc.Content.Size()
	w, h := sz.X, sz.Y
	x -= w / 2
	if x+w > UIWidth-2 {
		x = UIWidth - w - 2
	}
	if x < 0 {
		x = 0
	}
	md.statusDesc.Draw(md.gd.Slice(md.gd.Range().Lines(UIHeight-3-h, UIHeight-1).Shift(x, 0, 0, 0)))
}

func drawWelcome(gd gruid.Grid) gruid.Grid {
	tst := gruid.Style{}
	st := gruid.Style{}.WithAttrs(AttrInMap)
	stt := ui.StyledText{}.WithMarkups(map[rune]gruid.Style{
		't': tst.WithFg(ColorGreen),
		'W': st.WithFg(ColorViolet),
		'l': st.WithFg(ColorForeground).WithBg(ColorBackgroundSecondary),
		'm': st.WithFg(ColorForeground).WithBg(ColorBackground),
		'L': st.WithFg(ColorYellow).WithBg(ColorBackgroundSecondary),
		'p': st.WithFg(ColorBlue).WithBg(ColorBackgroundSecondary),
		'R': st.WithFg(ColorRed),
		'M': st.WithFg(ColorOrange).WithBg(ColorBackgroundSecondary),
		'P': st.WithFg(ColorMagenta),
		'h': tst.WithFg(ColorMagenta),
		'd': st.WithFg(ColorForegroundSecondary),
		's': st.WithFg(ColorCyan),
		'T': st.WithFg(ColorGreen),
		'b': st.WithFg(ColorYellow),
		'z': st.WithFg(ColorViolet),
		'r': st.WithFg(ColorOrange).WithBg(ColorBackgroundSecondary),
		'o': st.WithFg(ColorYellow),
	})
	rg := gd.Range()
	text := fmt.Sprintf("   @hHarmonist %s@N\n\n", Version)
	text += `@t───────────────────────
 @d#@l##@l#@d##############@d###
@d#,@L,""@l#@t HARMONIST @d#@d^,@b)@zt@d#
@d#,@pb@L,,"@l##@d#########@d^,,##
@d #@L,,,@r,,,@l#@d#@oπ@d.@P>@d##,,,,,#
@d @l#@L..@r.@Mg@r..+@d..@mG@d..@P+@d.,,,,#
@l#@p@@@L.@l#@d≈≈,.##@o☼@d.@o&@d##.,@T♣@d,",##
@l#@L.@l#@d#≈≈≈@R♫@d.##@P+@d##..@mh@d,"#,@s∩@d#
@l#@L.@l.@d##≈≈≈.......,,""""##
@t───────────────────────
`
	stt.WithText(text).Draw(gd.Slice(rg.Shift(23, 5, 0, 0)))
	ui.Text("—Press any key to continue—").Draw(gd.Slice(rg.Shift(21, 18, 0, 0)))
	return gd
}

func (md *model) drawWin(gd gruid.Grid) {
	st := gruid.Style{Attrs: AttrInMap}
	gd.Fill(gruid.Cell{Rune: '"', Style: st})
	treegd := gd.Slice(gruid.NewRange(4, 5, 8+62, 9+5))
	treegd.Fill(gruid.Cell{Rune: '♣', Style: st.WithFg(ColorGreen)})
	bananagd := gd.Slice(gruid.NewRange(5, 6, 7+62, 8+5))
	bananagd.Fill(gruid.Cell{Rune: ')', Style: st.WithFg(ColorYellow).WithBg(ColorBackgroundSecondary)})
	textgd := gd.Slice(gruid.NewRange(6, 7, 6+62, 7+5))
	textgd.Fill(gruid.Cell{Rune: ' ', Style: gruid.Style{Bg: ColorBackgroundSecondary}})
	stt := ui.StyledText{}.WithMarkups(Markups).WithStyle(gruid.Style{Bg: ColorBackgroundSecondary})
	var text string
	if md.g.LiberatedArtifact {
		text = `You escaped after both saving Shaedra and retrieving the artifact. ` +
			`What a journey! Shaedra and Marevor will surely give you some bananas!`
	} else {
		text = `You saved Shaedra, and that’s what mattered most to you! ` +
			`Marevor seems a bit disappointed, but he will surely still give you some bananas.`
	}
	stt.WithText(text).Format(60).Draw(textgd.Slice(textgd.Range().Shift(1, 1, -1, -1)))
}

func (md *model) drawMap(gd gruid.Grid) {
	it := md.g.Map.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		r, fg, bg := md.positionDrawing(p)
		attrs := AttrInMap
		if md.g.highlight[p] && md.g.Wizard != WizardMap || p == md.targ.ex.p {
			attrs |= AttrReverse
		}
		if r == '#' {
			attrs |= AttrBold
		}
		gd.Set(p, gruid.Cell{Rune: r, Style: gruid.Style{Fg: fg, Bg: bg, Attrs: attrs}})
	}
}

func (md *model) positionDrawing(p gruid.Point) (r rune, fgColor, bgColor gruid.Color) {
	g := md.g
	t := g.Map.At(p)
	kt := g.Map.KnownTerrain.At(p)
	if g.Wizard == WizardSeeAll || g.Wizard == WizardMap {
		kt = t
	}
	if g.Wizard == WizardMap {
		fgColor, bgColor = ColorForeground, ColorBackgroundSecondary
		if !g.HasNonWallNeighbors(p) {
			r = ' '
			bgColor = ColorBackground
			return
		}
		r, fgColor = g.TerrainStyle(kt, p)
		return
	}
	fgColor, bgColor = ColorForeground, ColorBackground
	if !IsKnown(kt) && (g.Wizard == WizardNone || g.Wizard == WizardNormal) {
		r, fgColor = g.TerrainStyle(kt, p)
		bgColor = ColorBackground
		if kt == Unknown && g.HasNonWallExploredNeighbor(p) {
			r = '¤'
			fgColor = ColorForegroundSecondary
		}
		if mons := g.KnownMonsterAt(p); mons.Exists() {
			r, fgColor = mons.memoryStyle()
		}
		if mons := g.MonsterAt(p); mons.Exists() && mons.Noise {
			r = '♫'
			fgColor = mons.noiseColor()
		} else if g.Map.Noise[p] > 0 {
			r = '♪'
			fgColor = ColorMagenta
		} else if g.Map.NoiseIllusion[p] > 0 {
			r = '♪'
			fgColor = ColorCyan
		}
		return
	}
	if g.Wizard == WizardSeeAll {
		if t == WallCell {
			if !g.HasNonWallNeighbors(p) {
				r = ' '
				return
			}
		}
	}
	if g.Player.Sees(p) {
		fgColor = ColorForeground
		bgColor = ColorBackgroundSecondary
	} else {
		fgColor = ColorForegroundSecondary
		bgColor = ColorBackground
	}
	var fgTerrain gruid.Color
	switch {
	case kt == WallCell || kt == WindowCell:
		r, fgTerrain = g.TerrainStyle(kt, p)
		if fgTerrain != ColorForeground {
			fgColor = fgTerrain
		}
	case CoversPlayer(kt):
		r, fgTerrain = g.TerrainStyle(kt, p)
		if p == g.Player.P {
			fgColor = ColorBlue
		} else if _, ok := g.Map.MagicalBarriers[p]; ok {
			fgColor = ColorCyan
		} else if fgTerrain != ColorForeground {
			fgColor = fgTerrain
		} else if g.Player.Sees(p) || g.Wizard == WizardSeeAll {
			mons := g.MonsterAt(p)
			if mons.Exists() {
				fgColor = mons.color(g)
			}
		}
	case p == g.Player.P:
		r = '@'
		fgColor = ColorBlue
	default:
		r, fgTerrain = g.TerrainStyle(kt, p)
		if fgTerrain != ColorForeground {
			fgColor = fgTerrain
		}
		if g.MonsterTargFOV != nil {
			if g.MonsterTargFOV[p] {
				fgColor = ColorOrange
			}
		} else if g.MonsterFOV[p] {
			fgColor = ColorOrange
		}
		infov := g.Player.Sees(p) || g.Wizard == WizardSeeAll
		if cld, ok := g.Map.Clouds[p]; ok && infov {
			r = '≡'
			switch cld {
			case CloudFire:
				fgColor = ColorOrange
			case CloudPurpleMist:
				fgColor = ColorViolet
			}
		}
		if infov {
			m := g.MonsterAt(p)
			if m.Exists() {
				r = m.Kind.Letter()
				fgColor = m.color(g)
			}
		} else if mons := g.MonsterAt(p); mons.Exists() && mons.Noise {
			r = '♫'
			fgColor = mons.noiseColor()
		} else if g.Map.Noise[p] > 0 {
			r = '♪'
			fgColor = ColorMagenta
		} else if g.Map.NoiseIllusion[p] > 0 {
			r = '♪'
			fgColor = ColorCyan
		} else if mons := g.KnownMonsterAt(p); mons.Exists() {
			r, fgColor = mons.memoryStyle()
		}
		if g.Player.Sees(p) && fgColor == ColorForeground && g.Illuminated(p) && IsIlluminable(t) && t != StoneCell {
			fgColor = ColorYellow
		}
	}
	return
}

func (m *Monster) noiseColor() gruid.Color {
	if m.State == Resting {
		return ColorViolet
	}
	switch m.Kind {
	case MonsSatowalgaPlant, MonsButterfly:
		// Satowalga plants and kerejats don't make any
		// noise, so shouldn't happen.
		return ColorForeground
	case MonsMirrorSpecter, MonsSpider:
		return ColorYellow
	case MonsWingedMilfid, MonsTinyHarpy:
		return ColorCyan
	case MonsEarthDragon, MonsTreeMushroom, MonsYack:
		return ColorMagenta
	case MonsWorm, MonsAcidMound:
		return ColorGreen
	case MonsDog, MonsBlinkingFrog, MonsHazeCat, MonsCrazyImp:
		return ColorOrange
	default:
		return ColorRed
	}

}

func (m *Monster) memoryStyle() (r rune, fg gruid.Color) {
	if !m.Seen {
		r = '☻'
	} else {
		r = m.Kind.Letter()
	}
	if m.LastKnownState == Resting {
		fg = ColorViolet
	} else if m.Kind.Peaceful() && m.Seen {
		fg = ColorBlue
	} else {
		fg = ColorForeground
	}
	return r, fg
}
