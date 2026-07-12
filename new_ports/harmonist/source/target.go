package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
	"codeberg.org/anaseto/gruid/ui"
)

var invalidPos = gruid.Point{-1, -1}

// examination holds all the targeting information.
type examination struct {
	p            gruid.Point   // position
	nmonster     int           // index of current monster
	objects      []gruid.Point // points for next object
	nobject      int           // index of current object
	sortedStairs []gruid.Point // sorted stair points
	stairIndex   int           // index of stair
	info         posInfo       // extra examine information
	scroll       int           // scrolling amount (descriptions)
}

// HideCursor hides the target cursor.
func (md *model) HideCursor() {
	md.targ.ex.p = invalidPos
}

// SetCursor sets the target cursor.
func (md *model) SetCursor(p gruid.Point) {
	md.targ.ex.p = p
}

// CancelExamine cancels current targeting.
func (md *model) CancelExamine() {
	md.g.highlight = nil
	md.g.MonsterTargFOV = nil
	if md.targ.ex != nil {
		*md.targ.ex = examination{}
		md.HideCursor()
	}
	md.targ.kbTargeting = false
}

// Examine targets a given position with the cursor.
func (md *model) Examine(p gruid.Point) {
	if md.targ.ex.p == p {
		// Avoid re-examining the same position.
		return
	}
	md.examine(p)
}

func (md *model) examine(p gruid.Point) {
	if !InMap(p) {
		return
	}
	md.SetCursor(p)
	md.computeHighlight()
	m := md.g.MonsterAt(p)
	if m.Exists() && (md.g.Player.Sees(p) || md.g.Wizard == WizardSeeAll) {
		md.g.ComputeMonsterCone(m)
	} else {
		md.g.MonsterTargFOV = nil
	}
	md.updatePosInfo()
	md.targ.ex.scroll = 0
}

// KeyboardExamine toggels keyboard examination mode on the player.
func (md *model) KeyboardExamine() {
	if md.targ.kbTargeting {
		md.CancelExamine()
		return
	}
	md.targ.kbTargeting = true
	p := md.g.Player.P
	md.targ.ex.p = p
	md.targ.ex.objects = []gruid.Point{}
	md.examine(p)
}

// posInfo gathers information about a target position.
type posInfo struct {
	P       gruid.Point
	Unknown bool
	Sees    bool
	Player  bool
	Monster *Monster
	Cell    rl.Cell
	Cloud   string
	Lighted bool
	MFOV    bool
}

func (md *model) drawPosInfo() {
	g := md.g
	p := gruid.Point{}
	if md.targ.ex.p.X < MapWidth/2 {
		p.X += MapWidth/2 + 1
	}
	info := md.targ.ex.info
	n := md.targ.ex.scroll
	y := 2
	stt := ui.StyledText{}.WithMarkups(Markups)
	formatBox := func(title, s string, fg gruid.Color) {
		if n > 0 {
			// Scrolling.
			n--
			return
		}
		md.description.Content = stt.WithText(s).Format(MapWidth/2 - 3)
		if md.description.Content.Size().Y+y > MapHeight {
			md.description.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg}),
				Footer: ui.NewStyledText("scroll/page down for more...", gruid.Style{Fg: ColorMagenta})}
		} else {
			md.description.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg})}
		}
		y += md.description.Draw(md.gd.Slice(gruid.NewRange(0, y, MapWidth/2-1, 2+MapHeight).Add(p))).Size().Y
	}

	features := []string{}
	if info.Unknown {
		if info.Cell == UnknownPassable {
			features = append(features, "passable terrain")
		} else {
			features = append(features, "unknown terrain")
		}
		features = append(features, "unexplored")
	} else {
		features = append(features, g.TerrainName(info.Cell, info.P))
		if info.Cloud != "" && info.Sees {
			features = append(features, info.Cloud)
		}
		if !info.Sees {
			features = append(features, "seen")
		} else if info.MFOV {
			features = append(features, "@Ounhidden@N")
		} else if info.Lighted {
			features = append(features, "@Ylighted@N")
		}
	}
	t := features[0]
	if len(features) > 1 {
		t += " (" + strings.Join(features[1:], ", ") + ")"
	}
	fg := ColorForeground
	desc := ""
	if info.Unknown {
		if info.Cell == UnknownPassable {
			desc = "Passable terrain of unknown nature."
		} else {
			desc = "Terrain of unknown nature."
		}
	} else {
		desc = g.TerrainDesc(info.Cell, info.P)
	}
	formatBox(t, desc, fg)

	if info.Player {
		formatBox("Syu", "This is you, the monkey named Syu.", ColorBlue)
		return
	}

	noiseBox := func() {
		if mons := g.MonsterAt(info.P); mons.Exists() && mons.Noise {
			fg := mons.noiseColor()
			var noiseMsg string
			if mons.State == Resting {
				noiseMsg = "You heard some breathing around there."
			} else {
				noiseMsg = fmt.Sprintf("You heard %s around there.", monsterNoise(mons.Kind))
			}
			formatBox("monster noise", noiseMsg, fg)
		} else if g.Map.Noise[info.P] > 0 {
			formatBox("noise", "You heard noise around there.", ColorCyan)
		}
	}

	mons := info.Monster
	if !mons.Exists() {
		noiseBox()
		return
	}
	title := fmt.Sprintf("%s (%s %s)", mons.Kind, mons.State, DirString(mons.Dir))
	if !info.Sees {
		title = fmt.Sprint(mons.Kind)
	}
	var mfg gruid.Color
	if g.Player.Sees(mons.P) || g.Wizard == WizardSeeAll {
		mfg = mons.color(g)
	} else {
		_, mfg = mons.memoryStyle()
	}
	var mdesc []string

	mdesc = append(mdesc, mons.Kind.Desc())
	if info.Sees {
		statuses := mons.statusesText()
		if statuses != "" {
			mdesc = append(mdesc, fmt.Sprintf("@CStatuses:@N %s", statuses))
		}
	}
	mdesc = append(mdesc, "@CTraits:@N "+mons.traits())
	if !mons.Seen && g.Wizard != WizardSeeAll {
		formatBox("unknown monster", "You sensed a monster there.", mfg)
	} else {
		if !mons.Peaceful(g) && mons.State != Resting && !mons.Status(MonsDazed) && info.MFOV {
			mdesc = append(mdesc, "@OCan see you from there!@N")
		}
		formatBox(title, strings.Join(mdesc, "\n"), mfg)
	}
	noiseBox()
}

func (m *Monster) color(g *Game) gruid.Color {
	var fg gruid.Color
	switch {
	case m.Status(MonsLignified) && m.Status(MonsConfused) || m.Status(MonsDazed):
		fg = ColorCyan
	case m.Status(MonsLignified):
		fg = ColorYellow
	case m.Status(MonsConfused):
		fg = ColorGreen
	case m.State == Resting:
		fg = ColorViolet
	case m.Peaceful(g):
		fg = ColorBlue
	case m.State == Hunting:
		fg = ColorRed
	default:
		fg = ColorOrange
	}
	return fg
}

func (m *Monster) statusesText() string {
	infos := []string{}
	for st, i := range m.Statuses {
		ms := monsterStatus(st)
		if i > 0 {
			infos = append(infos, fmt.Sprintf("@%c%s(%d)@N", ms.Color(), ms, m.Statuses[monsterStatus(st)]))
		}
	}
	return strings.Join(infos, ", ")
}

func (m *Monster) traits() string {
	var info string
	info += fmt.Sprintf("Their size is %s.", m.Kind.Size())
	if m.Kind.Peaceful() {
		info += " " + "They are peaceful."
	}
	if m.Kind.CanOpenDoors() {
		info += " " + "They can open doors."
	}
	if m.Kind.CanFly() {
		info += " " + "They can fly."
	}
	if m.Kind.CanSwim() {
		info += " " + "They can swim."
	}
	if m.Kind.ShallowSleep() {
		info += " " + "They have very shallow sleep."
	}
	if m.Kind.ResistsLignification() {
		info += " " + "They are unaffected by lignification."
	}
	if m.Kind.ReflectsTeleport() {
		info += " " + "They partially reflect back oric teleport magic."
	}
	if m.Kind.GoodSmell() {
		info += " " + "They have good sense of smell."
	}
	return info
}

func (md *model) updatePosInfo() {
	g := md.g
	pi := posInfo{}
	p := md.targ.ex.p
	pi.P = p
	switch {
	case g.Wizard == WizardMap:
		pi.Cell = g.Map.At(p)
		pi.Sees = true
		if g.Illuminated(p) && IsIlluminable(pi.Cell) {
			pi.Lighted = true
		}
		md.targ.ex.info = pi
		return
	case !IsKnown(g.Map.KnownTerrain.At(p)) && g.Wizard != WizardSeeAll:
		pi.Cell = g.Map.KnownTerrain.At(p)
		pi.Unknown = true
		if mons := g.KnownMonsterAt(p); mons.Exists() {
			pi.Monster = mons
		}
		md.targ.ex.info = pi
		return
	}
	if p == g.Player.P {
		pi.Player = true
	}
	infov := g.Player.Sees(p) || g.Wizard == WizardSeeAll
	pi.Sees = infov
	t := g.Map.KnownTerrain.At(p)
	if g.Wizard == WizardSeeAll {
		t = g.Map.At(p)
	}
	if mons := g.MonsterAt(p); mons.Exists() && infov {
		pi.Monster = mons
		if infov && mons.Sees(g, g.Player.P) {
			pi.MFOV = true
		}
	} else if mons := g.KnownMonsterAt(p); mons.Exists() && g.Wizard != WizardSeeAll {
		pi.Monster = mons
	}
	if cld, ok := g.Map.Clouds[p]; ok && infov {
		pi.Cloud = cld.String()
	}
	pi.Cell = t
	if infov && g.Illuminated(p) && IsIlluminable(t) {
		pi.Lighted = true
	}
	md.targ.ex.info = pi
}

func (md *model) computeHighlight() {
	g := md.g
	p := md.targ.ex.p
	path := g.PlayerPath(g.Player.P, p)
	g.highlight = map[gruid.Point]bool{}
	for _, p := range path {
		g.highlight[p] = true
	}
}

// UpdateTarget updates travel target.
func (md *model) UpdateTarget() error {
	g := md.g
	p := md.targ.ex.p
	t := g.Map.KnownTerrain.At(p)
	if g.Wizard == WizardSeeAll {
		t = g.Map.At(p)
	}
	expl := IsKnown(t)
	if !expl {
		return errors.New("You do not know this place.")
	}
	if t == WallCell && !g.Player.HasStatus(StatusDig) {
		return errors.New("You cannot travel into a wall.")
	}
	path := g.PlayerPath(g.Player.P, p)
	if len(path) == 0 {
		return errors.New("There is no safe path to this place.")
	}
	if expl && t != WallCell {
		g.autoTarget = p
		return nil
	}
	return errors.New("Invalid destination.")
}

// NextMonster examines next/previous known monster (depending on key).
// Visible monsters get priority.
func (md *model) NextMonster(key gruid.Key) {
	data := md.targ.ex
	ms := md.targetMonsters()
	if len(ms) == 0 {
		return
	}
	if !md.targ.kbTargeting {
		md.targ.kbTargeting = true
		data.nmonster = -1
	}
	nmonster := data.nmonster
	if key == "+" {
		nmonster++
	} else {
		nmonster--
	}
	if nmonster > len(ms)-1 {
		nmonster = 0
	} else if nmonster < 0 {
		nmonster = len(ms) - 1
	}
	mons := ms[nmonster]
	p := mons.P
	data.nmonster = nmonster
	md.Examine(p)
}

func (md *model) targetMonsters() []*Monster {
	if md.g.Wizard == WizardMap {
		return nil
	}
	tms := []*Monster{}
	for _, m := range md.g.Monsters {
		if !m.Exists() || !md.g.Player.Sees(m.P) && md.g.Wizard != WizardSeeAll {
			continue
		}
		tms = append(tms, m)
	}
	for p, i := range md.g.KnownMonster {
		m := md.g.Monsters[i]
		if !m.Exists() || md.g.Player.Sees(m.P) || md.g.Wizard == WizardSeeAll {
			continue
		}
		mons := &Monster{}
		*mons = *m
		mons.P = p
		tms = append(tms, mons)
	}
	sort.Slice(tms, func(i, j int) bool {
		mi := tms[i]
		mj := tms[j]
		if md.g.Player.Sees(mi.P) && !md.g.Player.Sees(mj.P) {
			return true
		}
		if !md.g.Player.Sees(mi.P) && md.g.Player.Sees(mj.P) {
			return false
		}
		return Distance(md.g.Player.P, mi.P) < Distance(md.g.Player.P, mj.P)
	})
	return tms
}

// NextStair examines the next known stair.
func (md *model) NextStair() {
	data := md.targ.ex
	g := md.g
	if data.sortedStairs == nil {
		stairs := g.getStairs()
		data.sortedStairs = g.SortedNearestTo(stairs, g.Player.P)
	}
	if len(data.sortedStairs) == 0 {
		return
	}
	if !md.targ.kbTargeting {
		md.targ.kbTargeting = true
		data.stairIndex = 0
	}
	md.targ.kbTargeting = true
	if data.stairIndex >= len(data.sortedStairs) {
		data.stairIndex = 0
	}
	if len(data.sortedStairs) > 0 {
		md.Examine(data.sortedStairs[data.stairIndex])
		data.stairIndex++
	}
}

func (g *Game) getStairs() []gruid.Point {
	stairs := []gruid.Point{}
	it := g.Map.KnownTerrain.Iterator()
	if g.Wizard == WizardSeeAll || g.Wizard == WizardMap {
		it = g.Map.Terrain.Iterator()
	}
	for it.Next() {
		t := it.Cell()
		if t != StairCell {
			continue
		}
		stairs = append(stairs, it.P())
	}
	return stairs
}

// NextObject examines the next magical object (stone, magara, item).
func (md *model) NextObject(np gruid.Point) {
	data := md.targ.ex
	g := md.g
	nobject := data.nobject
	if len(data.objects) == 0 {
		for p := range g.Objects.Stones {
			data.objects = append(data.objects, p)
		}
		for p := range g.Objects.Magaras {
			data.objects = append(data.objects, p)
		}
		for p := range g.Objects.Items {
			data.objects = append(data.objects, p)
		}
		data.objects = g.SortedNearestTo(data.objects, g.Player.P)
	}
	seeObjects := g.Wizard == WizardSeeAll || g.Wizard == WizardMap
	for i := 0; i < len(data.objects); i++ {
		p := data.objects[nobject]
		nobject++
		if nobject > len(data.objects)-1 {
			nobject = 0
		}
		if IsKnown(g.Map.KnownTerrain.At(p)) || seeObjects {
			np = p
			break
		}
	}
	if len(data.objects) == 0 || !InMap(np) || !IsKnown(g.Map.KnownTerrain.At(np)) && !seeObjects {
		return
	}
	if !md.targ.kbTargeting {
		md.targ.kbTargeting = true
	}
	data.nobject = nobject
	md.Examine(np)
}
