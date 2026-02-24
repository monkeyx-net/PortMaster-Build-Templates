package main

import (
	//"log"
	"sort"
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// These constants represent various common durations for animation frames.
const (
	AnimDurExtremelyShort = 15 * time.Millisecond
	AnimDurVeryShort      = 30 * time.Millisecond
	AnimDurShort          = 40 * time.Millisecond
	AnimDurShortMedium    = 60 * time.Millisecond
	AnimDurMedium         = 80 * time.Millisecond
	AnimDurMediumLong     = 100 * time.Millisecond
	AnimDurLong           = 150 * time.Millisecond
)

// Animations provides support for animations.
type Animations struct {
	frames []AnimFrame // frame queue
	grid   gruid.Grid  // drawing grid
	mgrid  gruid.Grid  // map grid (slice of grid)
	pgrid  gruid.Grid  // previous drawing grid
	idx    int         // animation sequence counter
	draw   bool        // whether to draw an animation frame or not
}

// AnimFrame represents an animation frame.
type AnimFrame struct {
	Cells    []gruid.FrameCell
	Duration time.Duration
}

// Finish clears the frame queue.
func (a *Animations) Finish() {
	a.idx++
	a.frames = nil
}

// Done reports whether there are no more animations in the queue.
func (a *Animations) Done() bool {
	return len(a.frames) == 0
}

// Draw sets a given rune and foreground color at a given position.
func (a *Animations) Draw(p gruid.Point, r rune, fg gruid.Color) {
	c := a.mgrid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	a.mgrid.Set(p, c)
}

// DrawWith sets a given rune and foreground and background color and
// attributes at a given position.
func (a *Animations) DrawWith(p gruid.Point, r rune, fg, bg gruid.Color, attrs gruid.AttrMask) {
	c := a.mgrid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	c.Style.Bg = bg
	c.Style.Attrs = attrs
	a.mgrid.Set(p, c)
}

// DrawRune sets a given rune at a given position.
func (a *Animations) DrawRune(p gruid.Point, r rune) {
	c := a.mgrid.At(p)
	c.Rune = r
	a.mgrid.Set(p, c)
}

// DrawColor sets a given color at a given position.
func (a *Animations) DrawColor(p gruid.Point, fg gruid.Color) {
	c := a.mgrid.At(p)
	c.Style.Fg = fg
	a.mgrid.Set(p, c)
}

// DrawReverse sets a given rune and foreground color with reverse attribute.
func (a *Animations) DrawReverse(p gruid.Point, r rune, fg gruid.Color) {
	c := a.mgrid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	c.Style.Attrs |= AttrReverse
	a.mgrid.Set(p, c)
}

// DrawReverse sets a given rune and foreground color with reverse attribute.
func (a *Animations) DrawReverseChar(p gruid.Point, r rune, fg gruid.Color) {
	c := a.mgrid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	c.Style.Attrs ^= AttrInMap
	c.Style.Attrs |= AttrReverse
	a.mgrid.Set(p, c)
}

// Frame emits a new animation frame with given duration.
func (a *Animations) Frame(d time.Duration) {
	frame := AnimFrame{}
	frame.Duration = d
	it := a.grid.Iterator()
	itp := a.pgrid.Iterator()
	for it.Next() && itp.Next() {
		if it.Cell() == itp.Cell() {
			continue
		}
		frame.Cells = append(frame.Cells, gruid.FrameCell{P: it.P(), Cell: it.Cell()})
	}
	a.frames = append(a.frames, frame)
	a.pgrid.Copy(a.grid)
}

type msgAnim int

// InitAnimations initializes drawing grids for animations.
func (md *model) InitAnimations() {
	sz := md.gd.Size()
	md.anims.grid = gruid.NewGrid(sz.X, sz.Y)
	md.anims.mgrid = md.anims.grid.Slice(md.anims.grid.Range().Shift(0, 2, 0, -1))
	md.anims.pgrid = gruid.NewGrid(sz.X, sz.Y)
	md.anims.grid.Copy(md.gd)
	md.anims.pgrid.Copy(md.gd)
}

// AnimNext returns a ticker command appropriately timed for the next frame.
func (md *model) AnimNext() gruid.Cmd {
	d := md.anims.frames[0].Duration
	idx := md.anims.idx
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAnim(idx)
	}
}

// AnimCmd returns a command with current animation sequence counter, which is
// used when updating to check whether the current animation is still relevant
// or not.
func (md *model) AnimCmd() gruid.Cmd {
	if len(md.anims.frames) == 0 {
		return nil
	}
	idx := md.anims.idx
	return func() gruid.Msg {
		return msgAnim(idx)
	}
}

// startAnimSeq is a helper function for starting a new animation sequence from
// a freshly drawn map representing the current state.
func (md *model) startAnimSeq() {
	if md.anims.Done() {
		// Compute new frame with respect to last drawn grid.
		md.anims.pgrid.Copy(md.gd)
	}
	agrid := md.anims.grid
	rg := agrid.Range()
	// Draw log.
	gdlog := agrid.Slice(rg.Lines(0, 2))
	gdlog.Fill(gruid.Cell{Rune: ' '})
	md.log.Content = md.DrawLog()
	md.log.Draw(gdlog)
	// Draw map.
	md.drawMap(md.anims.mgrid)
	// Draw status.
	md.updateStatusInfo()
	gdstatus := agrid.Slice(rg.Line(UIHeight - 1))
	gdstatus.Fill(gruid.Cell{Rune: ' '})
	gdstatus.Copy(md.status.Draw())
}

// AnimationFrameFast updates the map and waits very briefly.
func (md *model) AnimationFrameFast() {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	md.anims.Frame(AnimDurExtremelyShort)
}

func (md *model) SwappingAnimation(mp, pp gruid.Point) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	md.anims.DrawRune(mp, 'Φ')
	md.anims.Draw(pp, 'Φ', ColorBlue)
	md.anims.Frame(AnimDurMedium)
	md.anims.Draw(mp, 'Φ', ColorBlue)
	md.anims.DrawRune(pp, 'Φ')
	md.anims.Frame(AnimDurMedium)
}

func (md *model) TeleportAnimation(from, to gruid.Point, showto bool) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	if showto {
		md.anims.Draw(from, 'Φ', ColorBlue)
		md.anims.Draw(to, 'Φ', ColorCyan)
		md.anims.Frame(AnimDurMediumLong)
	} else {
		md.anims.Draw(from, 'Φ', ColorCyan)
		md.anims.Frame(AnimDurMediumLong)
	}
}

func (md *model) MonsterProjectileAnimation(ray []gruid.Point, r rune, fg gruid.Color) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	for i := range len(ray) {
		p := ray[i]
		md.anims.Draw(p, r, fg)
		md.anims.Frame(AnimDurVeryShort)
	}
}

func (md *model) WaveDrawAt(p gruid.Point, fg gruid.Color) {
	c := md.anims.mgrid.At(p)
	md.anims.DrawReverse(p, c.Rune, fg)
}

func (md *model) NoiseAnimation(noises []gruid.Point) {
	if DisableAnimations {
		return
	}
	md.FOVWavesAnimation(DefaultFOVRange, WaveHarmonicNoise, md.g.Player.P)
	colors := []gruid.Color{ColorViolet, ColorCyan}
	for i := range 2 {
		for _, p := range noises {
			r := '♫'
			md.anims.DrawReverse(p, r, colors[i])
		}
		md.anims.DrawReverse(md.g.Player.P, '@', colors[i])
		md.anims.Frame(AnimDurShort)
	}

}

func (md *model) OricExplosionAnimation(ps []gruid.Point) {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	colors := [3]gruid.Color{ColorForeground, ColorForegroundSecondary, ColorForegroundEmph}
	for range 4 {
		for _, p := range ps {
			fg := colors[g.rand.IntN(3)]
			if !g.Player.Sees(p) {
				continue
			}
			md.ExplosionDrawAt(p, fg)
		}
		md.anims.Frame(AnimDurMediumLong)
	}
}

func (md *model) NadreExplosionAnimation(p gruid.Point) {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	colors := [2]gruid.Color{ColorYellow, ColorOrange}
	nb := g.CardinalNeighbors(p)
	nb = append(nb, p)
	for range 4 {
		for _, q := range nb {
			fg := colors[g.rand.IntN(2)]
			if !g.Player.Sees(q) {
				continue
			}
			md.ExplosionDrawAt(q, fg)
		}
		md.anims.Frame(AnimDurMediumLong)
	}
}

// ExplosionDrawAt draws a random colored explosion character at the given
// point.
func (md *model) ExplosionDrawAt(p gruid.Point, fg gruid.Color) {
	g := md.g
	r := ';'
	switch RandInt(9) {
	case 0, 6:
		r = ','
	case 1:
		r = '}'
	case 2:
		r = '%'
	case 3, 7:
		r = ':'
	case 4:
		r = '\\'
	case 5:
		r = '~'
	}
	mons := g.MonsterAt(p)
	if mons.Exists() || g.Player.P == p {
		r = '√'
	}
	md.anims.DrawReverseChar(p, r, fg)
}

func (g *Game) Waves(maxCost int, ws wavestyle, center gruid.Point) (dists []int, cdists map[int][]int) {
	var mp paths.Pather
	switch ws {
	case WaveHarmonicNoise:
		mp = &GridPath{}
	default:
		mp = &NoisePath{g: g}
	}
	nodes := g.PRanim.BreadthFirstMap(mp, []gruid.Point{center}, maxCost)
	cdists = make(map[int][]int)
	for _, n := range nodes {
		cdists[n.Cost] = append(cdists[n.Cost], Pos2Idx(n.P))
	}
	for dist := range cdists {
		dists = append(dists, dist)
	}
	sort.Ints(dists)
	return dists, cdists
}

func (md *model) FOVWavesAnimation(r int, ws wavestyle, center gruid.Point) {
	dists, cdists := md.g.Waves(r, ws, center)
	for _, d := range dists {
		wave := cdists[d]
		if len(wave) == 0 {
			break
		}
		md.WaveAnimation(wave, ws)
	}
}

type wavestyle int

const (
	WaveHarmonicNoise wavestyle = iota
	WaveConfusion
	WaveDaze
	WaveTree
	WaveSleeping
)

func (md *model) WaveAnimation(wave []int, ws wavestyle) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	for _, i := range wave {
		p := Idx2Pos(i)
		switch ws {
		case WaveConfusion:
			fg := ColorGreen
			if md.g.Player.Sees(p) {
				md.WaveDrawAt(p, fg)
			}
		case WaveSleeping:
			fg := ColorViolet
			if md.g.Player.Sees(p) {
				md.WaveDrawAt(p, fg)
			}
		case WaveDaze:
			fg := ColorCyan
			if md.g.Player.Sees(p) {
				md.WaveDrawAt(p, fg)
			}
		case WaveTree:
			fg := ColorYellow
			if md.g.Player.Sees(p) {
				md.WaveDrawAt(p, fg)
			}
		case WaveHarmonicNoise:
			fg := ColorCyan
			md.WaveDrawAt(p, fg)
		}
	}
	md.anims.Frame(AnimDurVeryShort)
}

type beamstyle int

const (
	BeamSleeping beamstyle = iota
	BeamLignification
	BeamObstruction
)

func (md *model) BeamsAnimation(ray []gruid.Point, bs beamstyle) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	md.anims.Frame(AnimDurVeryShort)
	// change colors depending on effect
	var fg gruid.Color
	switch bs {
	case BeamSleeping:
		fg = ColorViolet
	case BeamLignification:
		fg = ColorYellow
	case BeamObstruction:
		fg = ColorCyan
	}
	for range 3 {
		for i := len(ray) - 1; i >= 0; i-- {
			p := ray[i]
			r := '*'
			if RandInt(2) == 0 {
				r = '×'
			}
			md.anims.DrawReverseChar(p, r, fg)
		}
		md.anims.Frame(AnimDurShort)
	}
}

func (md *model) JavelinThrowAnimation(ray []gruid.Point) {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	for i := range len(ray) {
		p := ray[i]
		md.anims.Draw(p, md.projectileSymbol(DirNorm(g.Player.P, p)), ColorRed)
		md.anims.Frame(AnimDurVeryShort)
	}
}

func (md *model) projectileSymbol(dir gruid.Point) (r rune) {
	switch dir {
	case gruid.Point{1, 0}, gruid.Point{-1, 0}:
		r = '—'
	case gruid.Point{1, -1}, gruid.Point{-1, 1}:
		r = '/'
	case gruid.Point{0, 1}, gruid.Point{0, -1}:
		r = '|'
	case gruid.Point{1, 1}, gruid.Point{-1, -1}:
		r = '\\'
	}
	return r
}

func (md *model) WoundedAnimation() {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	md.anims.DrawColor(g.Player.P, ColorOrange)
	md.anims.Frame(AnimDurMedium)
}

func (md *model) PlayerGoodEffectAnimation() {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	md.anims.DrawColor(g.Player.P, ColorGreen)
	md.anims.Frame(AnimDurMediumLong)
}

func (md *model) StatusEndAnimation() {
	md.EffectAtPPAnimation(ColorViolet)
}

func (md *model) EffectAtPPAnimation(cl gruid.Color) {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	md.anims.DrawColor(g.Player.P, cl)
	md.anims.Frame(AnimDurMedium)
}

func (md *model) FoundFakeStairsAnimation() {
	if DisableAnimations {
		return
	}
	g := md.g
	md.anims.DrawColor(g.Player.P, ColorForeground)
	md.anims.Frame(AnimDurMediumLong)
}

func (md *model) PushAnimation(path []gruid.Point) {
	if DisableAnimations {
		return
	}
	if len(path) == 0 {
		// should not happen
		return
	}
	md.startAnimSeq()
	for _, p := range path[:len(path)-1] {
		md.anims.Draw(p, '×', ColorBlue)
	}
	md.anims.Draw(path[len(path)-1], '@', ColorBlue)
	md.anims.Frame(AnimDurMediumLong)
}

// MoveAnimation animates an actor's straight movement between two positions
// (possibly adjacent).
func (md *model) MoveAnimation(from, to gruid.Point) {
	if DisableAnimations {
		return
	}
	if from == to {
		// should not happen
		return
	}
	done := false
	delta := to.Sub(from)
	delta = delta.Div(max(delta.X, -delta.X, delta.Y, -delta.Y))
	for p := from; p != to; p = p.Add(delta) {
		if md.g.Player.Sees(p) {
			if !done {
				done = true
				md.startAnimSeq()
			}
			md.anims.Draw(p, '×', ColorForeground)
		}
	}
	if done {
		md.anims.Frame(AnimDurMedium)
	}
}

func (md *model) AbyssFallAnimation() {
	if DisableAnimations {
		return
	}
	gd := rl.NewGrid(MapWidth, MapHeight)
	n := 5
	gd.FillFunc(func() rl.Cell {
		return rl.Cell(RandInt(n))
	})
	for i := range n {
		it := gd.Iterator()
		for it.Next() {
			if it.Cell() == rl.Cell(i) {
				md.anims.DrawWith(it.P(), ' ', ColorForeground, ColorBackground, AttrInMap)
			}
		}
		md.anims.Frame(AnimDurVeryShort)
	}
}

// NextStoryPart makes a story sequence progress.
func (md *model) NextStoryPart() {
	switch md.g.Map.Depth {
	case WinDepth:
		md.FreeingShaedra()
	case MaxDepth:
		md.TakingArtifact()
	default:
		md.mode = modeNormal // should not happen
	}
}

// FreeingShaedra handles Shaedra's liberation story sequence.
func (md *model) FreeingShaedra() {
	g := md.g
	switch md.story {
	case 1:
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Draw(g.Places.Monolith, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Objects.Stairs[g.Places.Monolith] = WinStair
		g.Map.Set(g.Places.Monolith, StairCell)
		g.Map.KnownTerrain.Set(g.Places.Monolith, StairCell)
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Frame(AnimDurMediumLong)
			md.anims.Draw(g.Places.Marevor, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Objects.Story[g.Places.Marevor] = StoryMarevor
		g.LogStyled("Marevor: “And what about the mission? Take that magara!”", logSpecial)
		g.LogStyled("Shaedra: “Pff, don’t be reckless!”", logSpecial)
		g.LogStyled("[SPACE/ESC to continue]", logConfirm)
		// Progress to next story part.
		md.story++
	default:
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Draw(g.Places.Marevor, 'Φ', ColorCyan)
			md.anims.Draw(g.Places.Shaedra, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Map.Set(g.Places.Shaedra, GroundCell)
		g.Map.KnownTerrain.Set(g.Places.Shaedra, GroundCell)
		g.Map.Set(g.Places.Marevor, ScrollCell)
		g.Map.KnownTerrain.Set(g.Places.Marevor, ScrollCell)
		g.Objects.Scrolls[g.Places.Marevor] = ScrollExtended
		g.RescuedShaedra()
		// We're done with the story sequence.
		md.story = 0
		md.mode = modeNormal
	}
}

// RescuedShaedra finalizes Shaedra's liberation, including Marevor giving the
// player a special magara.
func (g *Game) RescuedShaedra() {
	g.Player.Magaras = append(g.Player.Magaras, Magara{})
	g.Player.MP++ // gain one MP point when obtaining the new magara slot and returning the artifact
	g.Player.Inventory.Misc = Item{Kind: NoItem}
	g.LogStyled("You equip the new magara in the artifact’s old place.", logSpecial)
	if RandInt(2) == 0 {
		g.Player.Magaras[len(g.Player.Magaras)-1] = Magara{Kind: DispersalMagara, Charges: DispersalMagara.DefaultCharges()}
	} else {
		g.Player.Magaras[len(g.Player.Magaras)-1] = Magara{Kind: OricExplosionMagara, Charges: OricExplosionMagara.DefaultCharges()}
	}
	AchRescuedShaedra.Get(g)
}

// TakingArtifact handles the artifact's activation final story sequence.
func (md *model) TakingArtifact() {
	g := md.g
	switch md.story {
	case 1:
		g.Map.Set(g.Places.Artifact, GroundCell)
		g.Map.KnownTerrain.Set(g.Places.Artifact, GroundCell)
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Draw(g.Places.Monolith, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Objects.Stairs[g.Places.Monolith] = WinStair
		g.Map.Set(g.Places.Monolith, StairCell)
		g.Map.KnownTerrain.Set(g.Places.Monolith, StairCell)
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Draw(g.Places.Marevor, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Objects.Story[g.Places.Marevor] = StoryMarevor
		g.Map.KnownTerrain.Set(g.Places.Marevor, StoryCell)
		g.LogStyled("Marevor: “Great! Let’s escape and find some bones to celebrate!”", logSpecial)
		g.LogStyled("Syu: “Sorry, but I prefer bananas!”", logSpecial)
		g.LogStyled("[SPACE/ESC to continue]", logConfirm)
		// Progress to next story part.
		md.story++
	default:
		if !DisableAnimations {
			md.startAnimSeq()
			md.anims.Draw(g.Places.Marevor, 'Φ', ColorCyan)
			md.anims.Frame(AnimDurMediumLong)
		}
		g.Map.Set(g.Places.Marevor, GroundCell)
		g.Map.KnownTerrain.Set(g.Places.Marevor, GroundCell)
		AchRetrievedArtifact.Get(g)
		// We're done with the story sequence.
		md.story = 0
		md.mode = modeNormal
	}
}

// PortalEscapeAnimation animates the portal escape.
func (md *model) PortalEscapeAnimation() {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	g := md.g
	c := gruid.Cell{Style: gruid.Style{Attrs: AttrInMap}}
	for k := range 7 {
		for p := range g.Map.Terrain.Points() {
			c.Style.Bg = ColorBackground
			if p.X >= 6 && p.X < 6+62 && p.Y >= 7 && p.Y < 7+5 {
				c.Rune = ' '
				c.Style.Fg = ColorForeground
				c.Style.Bg = ColorBackgroundSecondary
			} else if p.X >= 5 && p.X < 7+62 && p.Y >= 6 && p.Y < 8+5 {
				c.Rune = ')'
				c.Style.Fg = ColorYellow
				c.Style.Bg = ColorBackgroundSecondary
			} else if p.X >= 4 && p.X < 8+62 && p.Y >= 5 && p.Y < 9+5 {
				c.Rune = '♣'
				c.Style.Fg = ColorGreen
			} else {
				c.Rune = '"'
				c.Style.Fg = ColorForeground
			}
			switch g.rand.IntN(4) {
			case 0:
				md.anims.mgrid.Set(p, c)
			case 1:
				if k >= 4 {
					md.anims.mgrid.Set(p, c)
				}
			}
		}
		md.anims.Frame(AnimDurMediumLong)
	}
}
