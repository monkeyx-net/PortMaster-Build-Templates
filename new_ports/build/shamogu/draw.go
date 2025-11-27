// This file defines the Draw method for the model.

package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// RenderOrder is a type representing the priority of an entity rendering.
type RenderOrder int

// Those constants represent distinct kinds of rendering priorities. In case
// two entities are at a given position, only the one with the highest priority
// gets displayed.
const (
	RONone RenderOrder = iota
	ROItem
	ROActor
)

// RenderOrder returns the rendering priority of an entity.
func (g *Game) RenderOrder(i ID) (ro RenderOrder) {
	e := g.Entity(i)
	if i == PlayerID {
		return ROActor
	}
	if a, ok := e.Role.(*Actor); ok && !a.KnownDead {
		return ROActor
	}
	if e.IsItem() && i >= FirstMapID {
		return ROItem
	}
	return RONone
}

// RenderOrderWizard returns the rendering priority of an entity for the wizard
// monster revealing mode.
func (g *Game) RenderOrderWizard(i ID) (ro RenderOrder) {
	e := g.Entity(i)
	if i == PlayerID {
		return ROActor
	}
	if a, ok := e.Role.(*Actor); ok && a.IsAlive() {
		return ROActor
	}
	if e.IsItem() && i >= FirstMapID {
		return ROItem
	}
	return RONone
}

// Draw implements Draw() for gruid.Model.
func (md *model) Draw() gruid.Grid {
	if !md.anims.Done() {
		return md.drawAnimationFrame()
	}
	md.gd.Fill(gruid.Cell{Rune: ' '})
	switch md.mode {
	case modeQuitting:
		return md.gd.Slice(gruid.Range{})
	case modePager:
		if md.pager.mode == modeDump {
			md.gd.Copy(md.pager.pg.Draw())
			return md.gd
		}
	case modeNewGame, modeNewGameAdvanced:
		md.drawSpiritSelection()
		return md.gd
	case modeNewGameMods:
		md.drawModSelection()
		return md.gd
	case modeLoadGame:
		md.drawLoadGameScreen()
		return md.gd
	}
	// Log drawing.
	md.log.Content = md.DrawLog()
	md.log.Draw(md.gd.Slice(md.gd.Range().Lines(0, 2)))
	// Map drawing.
	if md.g.win {
		md.drawWin(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	} else {
		md.drawMap(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	}
	// Some extra widgets may appear in some modes: they're drawn over log
	// lines and map.
	switch md.mode {
	case modeNormal:
		if md.status.focus {
			md.drawStatusDesc()
		} else if md.targ.p != InvalidPos {
			// Target description label drawing (over log lines and map).
			md.drawTargInfo()
		}
	case modePager:
		pgd := md.pager.pg.Draw()
		rg := md.pager.pg.View()
		sz := pgd.Size()
		if rg.Min.Y > 0 {
			pgd.Set(gruid.Point{sz.X - 1, 1}, gruid.Cell{Rune: '↑'})
		}
		if rg.Max.Y < md.pager.pg.Lines() {
			pgd.Set(gruid.Point{sz.X - 1, sz.Y - 2}, gruid.Cell{Rune: '↓'})
		}
		md.gd.Copy(pgd)
	case modeMenu:
		switch md.menu.mode {
		case modeInventory, modeInventoryBound, modeEquip:
			md.drawInventory()
		case modeGameMenu, modeConfigMenu, modeHelpMenu:
			md.gd.Copy(md.menu.main.Draw())
			if md.desc.Content.Text() != "" {
				md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
			}
		case modeKeysView, modeKeysChange:
			md.drawKeySettings()
		}
	}
	// We draw the status line last: it should always be visible and no
	// other widgets should ever need that space.
	md.gd.Slice(md.gd.Range().Line(UIHeight - 1)).Copy(md.status.menu.Draw())
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

func (md *model) drawWin(gd gruid.Grid) {
	gd.Fill(gruid.Cell{
		Rune:  '♣',
		Style: gruid.Style{Fg: ColorGreen, Bg: ColorBackground, Attrs: AttrInMap}})
	textgd := gd.Slice(gruid.NewRange(6, 7, 6+62, 7+5))
	textgd.Fill(gruid.Cell{Rune: ' '})
	stt := ui.StyledText{}.WithMarkups(Markups)
	stt.WithText(`You destroyed the Orb of Corruption! ` +
		`The dungeon is vanishing and you feel teleported away back to your mountain! `).
		Format(60).Draw(textgd.Slice(textgd.Range().Shift(1, 1, -1, -1)))
}

func (md *model) drawMap(gd gruid.Grid) {
	g := md.g
	switch g.Wizard.Mode {
	case WizardReveal:
		md.drawMapWizardReveal(gd)
		return
	case WizardRevealTerrain:
		md.drawMapWizardRevealTerrain(gd)
		return
	}
	// Terrain drawing.
	for p := range g.Map.KnownTerrain.Points() {
		r := g.Map.KnownTerrainRuneAt(p)
		g.drawTerrainRuneAt(gd, r, p)
	}
	for _, cl := range g.Map.Clouds.Clouds {
		if !g.InFOV(cl.P) {
			continue
		}
		drawCloud(gd, cl)
	}
	// Entity drawing (with rendering order).
	SortedIDs := make([]ID, len(g.Entities))
	for i := range SortedIDs {
		SortedIDs[i] = ID(i)
	}
	slices.SortFunc(SortedIDs, func(i, j ID) int {
		return cmp.Compare(g.RenderOrder(i), g.RenderOrder(j))
	})
	for _, i := range SortedIDs {
		o := g.RenderOrder(i)
		if o == RONone {
			continue
		}
		e := md.g.Entity(i)
		if o == ROItem && g.CloudAt(e.P) && g.InFOV(e.P) {
			// Favor clouds over items, but in case of normal
			// clouds, show the color of the item.
			if g.Map.Clouds.At(e.P).Kind == CloudNormal {
				c := gd.At(e.P)
				c.Style.Fg = e.Color()
				gd.Set(e.P, c)
			}
			continue
		}
		if e.Noise && !g.InFOV(e.P) {
			c := gd.At(e.P)
			c.Rune = '♫'
			c.Style.Fg = noiseColor(e)
			gd.Set(e.P, c)
			if e.KnownP == e.P {
				continue
			}
		}
		p := e.KnownP
		if p == InvalidPos {
			continue
		}
		g.drawEntityAt(gd, e, p)
	}
	// Draw non-movement noise.
	for at := range g.Map.Noise {
		if !g.InFOV(at) {
			c := gd.At(at)
			c.Rune = '♪'
			c.Style.Fg = ColorCyan
			gd.Set(at, c)
		}
	}
	// Travel path highlighting.
	md.drawTravelPath(gd)
}

func (g *Game) sortedIDsByRenderOrder() []ID {
	sortedIDs := make([]ID, len(g.Entities))
	for i := range sortedIDs {
		sortedIDs[i] = ID(i)
	}
	slices.SortFunc(sortedIDs, func(i, j ID) int {
		return cmp.Compare(g.RenderOrder(i), g.RenderOrder(j))
	})
	return sortedIDs
}

func (md *model) drawMapWizardReveal(gd gruid.Grid) {
	g := md.g
	// Terrain drawing.
	for p, t := range g.Map.Terrain.All() {
		r := g.Map.TerrainRuneAt(p, t)
		g.drawTerrainRuneAt(gd, r, p)
	}
	for _, cl := range g.Map.Clouds.Clouds {
		drawCloud(gd, cl)
	}
	// Entity drawing (with rendering order).
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		o := g.RenderOrderWizard(i)
		if o == RONone {
			continue
		}
		e := md.g.Entity(i)
		switch {
		case o == ROActor:
			if e.Actor().IsDead() {
				continue
			}
		case o == ROItem && g.CloudAt(e.P):
			// Favor clouds over items, but in case of normal
			// clouds, show the color of the item.
			if g.Map.Clouds.At(e.P).Kind == CloudNormal {
				c := gd.At(e.P)
				c.Style.Fg = e.Color()
				gd.Set(e.P, c)
			}
			continue
		}
		g.drawEntityAt(gd, e, e.P)
	}
	// Travel path highlighting.
	md.drawTravelPath(gd)
}

func (md *model) drawMapWizardRevealTerrain(gd gruid.Grid) {
	g := md.g
	// Terrain drawing.
	for p, t := range g.Map.Terrain.All() {
		r := g.Map.TerrainRuneAt(p, t)
		g.drawTerrainRuneAt(gd, r, p)
	}
	// Entity drawing (with rendering order).
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		if g.RenderOrder(i) != ROItem {
			continue
		}
		e := md.g.Entity(i)
		g.drawEntityAt(gd, e, e.P)
	}
	it := gd.Iterator()
	for it.Next() {
		c := it.Cell()
		if c.Style.Fg == ColorForegroundSecondary {
			c.Style.Fg = ColorForeground
		}
		if c.Rune == ' ' {
			c.Style.Bg = ColorBackground
		} else {
			c.Style.Bg = ColorBackgroundSecondary
		}
		it.SetCell(c)
	}
	// Target highlighting.
	if p := md.targ.p; p != InvalidPos {
		c := gd.At(p)
		c.Style.Attrs |= AttrReverse
		gd.Set(p, c)
	}
}

func (g *Game) drawTerrainRuneAt(gd gruid.Grid, r rune, p gruid.Point) {
	c := gruid.Cell{Rune: r}
	if g.InFOV(p) {
		c.Style.Bg = ColorBackgroundSecondary
	} else {
		if c.Style.Fg == ColorForeground {
			c.Style.Fg = ColorForegroundSecondary
		}
	}
	c.Style.Attrs = AttrInMap
	if r == '#' || r == '◊' {
		c.Style.Attrs |= AttrBold
	}
	gd.Set(p, c)
}

func drawCloud(gd gruid.Grid, cl Cloud) {
	c := gd.At(cl.P)
	c.Rune = '§'
	c.Style.Fg = cl.Kind.Color()
	gd.Set(cl.P, c)
}

func (md *model) drawTravelPath(gd gruid.Grid) {
	if md.targ.p != InvalidPos {
		for _, p := range md.auto.path {
			c := gd.At(p)
			c.Style.Attrs |= AttrReverse
			gd.Set(p, c)
		}
	}
}

func (g *Game) drawEntityAt(gd gruid.Grid, e *Entity, p gruid.Point) {
	c := gd.At(p)
	c.Rune = e.Rune
	if !e.IsActor() || g.InFOV(p) ||
		g.PlayerActor().Has(StatusClarity) && clarityRange(g.PP(), p) || g.Wizard.Mode.Reveal() {
		c.Style.Fg = e.Color()
	} else {
		c.Style.Fg = ColorForeground
	}
	gd.Set(p, c)
}

func (md *model) drawStatusDesc() {
	rg := md.status.menu.ActiveBounds()
	x := (rg.Min.X + rg.Max.X) / 2
	sz := md.status.desc.Content.Size()
	w, h := sz.X, sz.Y
	x -= w / 2
	if x+w > UIWidth {
		x = UIWidth - w
	}
	if x < 0 {
		x = 0
	}
	md.status.desc.Draw(md.gd.Slice(md.gd.Range().Lines(UIHeight-3-h, UIHeight-1).Shift(x, 0, 0, 0)))
}

// Markups contains the styling markup-characters we use for StyledText.
var Markups = map[rune]gruid.Style{
	'B': {Fg: ColorBlue},
	'C': {Fg: ColorCyan},
	'G': {Fg: ColorGreen},
	'M': {Fg: ColorMagenta},
	'O': {Fg: ColorOrange},
	'R': {Fg: ColorRed},
	'V': {Fg: ColorViolet}, // unused
	'Y': {Fg: ColorYellow},
	'b': {Fg: ColorBlue, Attrs: AttrInMap},
	'c': {Fg: ColorCyan, Attrs: AttrInMap},
	'g': {Fg: ColorGreen, Attrs: AttrInMap},
	'm': {Fg: ColorMagenta, Attrs: AttrInMap},
	'o': {Fg: ColorOrange, Attrs: AttrInMap},
	'r': {Fg: ColorRed, Attrs: AttrInMap},
	'v': {Fg: ColorViolet, Attrs: AttrInMap}, // unused
	'y': {Fg: ColorYellow, Attrs: AttrInMap},
}

func (md *model) drawTargInfo() {
	g := md.g
	p := gruid.Point{}
	if md.targ.p.X < MapWidth/2 {
		p.X += MapWidth/2 + 1
	}
	info := md.targ.info

	y := 2
	stt := ui.StyledText{}.WithMarkups(Markups)
	formatBox := func(title, s string, fg gruid.Color) {
		md.desc.Content = stt.WithText(s).Format(MapWidth/2 - 3)
		if md.desc.Content.Size().Y+y > MapHeight {
			md.desc.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg}),
				Footer: ui.NewStyledText("scroll/page down for more…", gruid.Style{Fg: ColorMagenta})}
		} else {
			md.desc.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg})}
		}
		y += md.desc.Draw(md.gd.Slice(gruid.NewRange(0, y, MapWidth/2-1, 2+MapHeight).Add(p))).Size().Y
	}

	features := []string{}
	switch {
	case g.Wizard.Mode.Reveal():
		features = append(features, TerrainName(g.Map.Terrain.At(md.targ.p)))
	default:
		features = append(features, TerrainName(g.Map.KnownTerrain.At(md.targ.p)))
	}
	if info.cloud != NoCloud {
		clstr := info.cloud.String()
		features = append(features, fmt.Sprintf("@%c%s@N", info.cloud.ColorRune(), clstr))
	}
	if !info.sees && !info.unknown {
		features = append(features, "seen")
	} else if info.unknown {
		features = append(features, "unexplored")
	}
	t := features[0]
	if len(features) > 1 {
		t += " (" + strings.Join(features[1:], ", ") + ")"
	}
	var fg gruid.Color
	desc := ""
	switch {
	case g.Wizard.Mode.Reveal():
		desc = TerrainDesc(g.Map.Terrain.At(md.targ.p))
	default:
		desc = TerrainDesc(g.Map.KnownTerrain.At(md.targ.p))
	}
	if info.cloud != NoCloud {
		desc += "\n" + info.cloud.Desc()
	}
	if noise, ok := g.Map.Noise[md.targ.p]; ok && !g.InFOV(md.targ.p) && !md.targ.scroll {
		msg := "noise"
		switch noise {
		case NoiseBark:
			msg = "barking"
		case NoiseCombat:
			msg = "smashing"
		case NoiseDig:
			msg = "rocks falling"
		case NoiseExplosion:
			msg = "an explosion"
		case NoiseEarthMenhir:
			msg = "noisy echo" // should not happen (player-only now)
		case NoiseFakePortal:
			msg = "an eerie sound" // should not happen (player-only now)
		case NoiseLightning:
			msg = "a spark" // should not happen (player-only now)
		case NoiseRoar:
			msg = "a roar" // should not happen (player-only now)
		case NoiseWind:
			msg = "a whistle" // should not happen (player-only now)
		}
		formatBox("noise", fmt.Sprintf("You heard %s around there.", msg), ColorCyan)
	}
	if !md.targ.scroll {
		formatBox(t, desc, fg)
	}
	for i, id := range info.entities {
		if md.targ.scroll && i < len(info.entities)-1 {
			continue
		}
		e := g.Entity(id)
		if !info.sees && e.Noise && e.P == md.targ.p && !g.Wizard.Mode.Reveal() {
			noise := "footsteps"
			if a, ok := e.Role.(*Actor); ok {
				noise = monsterNoiseStringStyled(a)
			}
			formatBox("monster noise", fmt.Sprintf("You heard %s around there.", noise), ColorForeground)
			continue
		}
		if e.KnownP == InvalidPos && !g.Wizard.Mode.Reveal() {
			continue
		}
		var sb strings.Builder
		name := e.Name
		cl := e.Color()
		switch r := e.Role.(type) {
		case *Actor:
			switch {
			case g.Wizard.Mode == WizardRevealTerrain:
				continue
			case info.sees ||
				g.PlayerActor().Has(StatusClarity) && clarityRange(g.PP(), e.P) ||
				g.Wizard.Mode == WizardReveal:
				fmt.Fprintf(&sb, "HP:%s A:%s D:%s",
					r.fmtHP(),
					fmtAD(r.GetAttack(), r.Attack), fmtAD(r.GetDefense(), r.Defense))
				for i, turns := range r.Statuses {
					if turns <= 0 {
						continue
					}
					st := Status(i)
					fmt.Fprintf(&sb, " @%c%s(%d)@N ", statusColor(st, turns), st.Abbr(), turns)
				}
				if s := g.MindStateString(r); s != "" {
					name += fmt.Sprintf(" (%s)", s)
				}
			default:
				cl = ColorForeground
				fmt.Fprintf(&sb, "HP:?/%d A:%d D:%d", r.MaxHP, r.Attack, r.Defense)
			}
			sb.WriteByte('\n')
			fmt.Fprintf(&sb, "@CTraits:@N %s.", r.Traits)
		case *Spirit:
			sb.WriteString("@CTotemic spirit.@N\n" + r.Desc())
		case *Menhir:
			sb.WriteString(r.Desc())
			if r.Used {
				name += " (inert)"
			}
		case *RunicTrap:
			sb.WriteString(r.Desc())
			if r.KnownUsed || r.Used && g.Wizard.Mode.Reveal() {
				name += " (inert)"
			}
		case Item:
			sb.WriteString(r.Desc())
		}
		formatBox(name, sb.String(), cl)
	}
}

const spiritSelectionMargin = 5

func (md *model) drawSpiritSelection() {
	menugd := md.menu.main.Draw()
	gdslice := md.gd.Slice(gruid.NewRange(0, spiritSelectionMargin, UIWidth, UIHeight))
	gdslice.Copy(menugd)
	md.desc.Draw(gdslice.Slice(gdslice.Range().Columns(UIWidth/2, UIWidth)))
	ui.Textf("Shamanic Mountain Guardian - Shamogu %s", Version).WithStyle(gruid.Style{}.WithFg(ColorMagenta)).
		Draw(md.gd.Slice(gruid.NewRange(-25+UIWidth/2, 3, UIWidth, UIHeight)))
	y := 13
	if md.mode == modeNewGameAdvanced {
		y += len(primarySpiritsAdvanced)
	}
	ui.Text(`(TAB for advanced new game settings)`).WithStyle(gruid.Style{Fg: ColorGreen}).
		Draw(md.gd.Slice(gruid.NewRange(0, y, UIWidth/2, y+1)))
	y = 16
	if md.mode == modeNewGameAdvanced {
		y += len(primarySpiritsAdvanced) - 1
	}
	ui.Text(`All life in your mountain is being corrupted. Beasts are losing their minds and ` +
		`becoming aggressive. These disturbing events started to happen after a dungeon portal suddenly ` +
		`appeared at the top of the mountain. As the guardian, you decide to explore ` +
		`that dungeon in order to find and destroy the source of corruption… ` +
		`You hope luck will be on your side!`).Format(72).
		Draw(md.gd.Slice(gruid.NewRange(4, y, UIWidth-4, UIHeight)))
}

func (md *model) drawModSelection() {
	menugd := md.menu.main.Draw()
	md.gd.Copy(menugd)
	md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
}

func (md *model) drawLoadGameScreen() {
	ui.Textf("Shamanic Mountain Guardian - Shamogu %s", Version).WithStyle(gruid.Style{}.WithFg(ColorMagenta)).
		Draw(md.gd.Slice(gruid.NewRange(-25+UIWidth/2, 5, UIWidth, UIHeight)))
	ui.Text("—Press any key to load saved game—").
		Draw(md.gd.Slice(gruid.NewRange(-22+UIWidth/2, 18, UIWidth, UIHeight)))
	gd := md.gd.Slice(gruid.NewRange(-17+UIWidth/2, 7, UIWidth, UIHeight))
	drawGamePicture(gd)
}

func drawGamePicture(gd gruid.Grid) {
	markups := maps.Clone(Markups)
	for i, st := range markups {
		markups[i] = st.WithAttrs(AttrInMap)
	}
	st := gruid.Style{}.WithAttrs(AttrInMap)
	markups['w'] = st.WithFg(ColorForeground).WithBg(ColorBackgroundSecondary).WithAttrs(AttrBold | AttrInMap)
	markups['W'] = st.WithFg(ColorForegroundSecondary).WithBg(ColorBackground).WithAttrs(AttrBold | AttrInMap)
	markups['l'] = st.WithFg(ColorForeground).WithBg(ColorBackgroundSecondary)
	markups['y'] = st.WithFg(ColorYellow).WithBg(ColorBackgroundSecondary)
	markups['c'] = st.WithFg(ColorCyan).WithBg(ColorBackgroundSecondary)
	markups['b'] = st.WithFg(ColorBlue).WithBg(ColorBackgroundSecondary)
	markups['r'] = st.WithFg(ColorRed).WithBg(ColorBackgroundSecondary)
	markups['d'] = st.WithFg(ColorForegroundSecondary).WithBg(ColorBackground)
	markups['t'] = gruid.Style{}.WithFg(ColorGreen)
	stt := ui.StyledText{}.WithMarkups(markups)
	text := `@t───────────────────────
@l""@d...@w#@W##########@d...@W#@d^@W##
@d"@l.@w◊@d.@l@rF@w#@t SHAMOGU @W#@d.....^@W#
@d.@l^@c&@w#@l..@w#####@W####@d..@W##@d....
@l^..@b@@@l.....@rC@l..@d....@W###@d..."
@l..^@w#@l.@rc@l.@w◊@c>@w#@d^..@W#####@d.@Y%@d.""
@d.@O♫@d"@W#@w#@l...@w◊@y!@d..""@W###@d..""""
@d."""@W#@w#@l...@d..."""@W#@d..""""@W#
@d"""""@W#@w#@l.@d.........."""@W##
@t───────────────────────
`
	stt.WithText(text).Draw(gd)
}

func (md *model) drawInventory() {
	menugd := md.menu.main.Draw()
	md.gd.Copy(menugd)
	descgd := md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
	if md.menu.mode != modeEquip || !md.drawGroundDesc {
		return
	}
	// We draw the description of the item on the ground below the other
	// one, unless we use it to upgrade (then description doesn't matter).
	md.equipLabel.Draw(md.gd.Slice(gruid.NewRange(UIWidth/2, descgd.Size().Y, UIWidth, UIHeight-1)))
}

func (md *model) drawKeySettings() {
	gd := md.menu.keys.Draw()
	max := gd.Size()
	t := ui.Text("(R) reset (Enter) change").WithStyle(gruid.Style{}.WithFg(ColorCyan))
	if md.menu.mode == modeKeysChange {
		t = t.WithText(" Press new key… ")
	}
	t.Draw(gd.Slice(gd.Range().Line(max.Y-1).Shift(2, 0, 0, 0)))
	md.gd.Copy(gd)
}
