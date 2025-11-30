package main

import (
	"math/rand/v2"
	"strings"
	"testing"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

const rounds = 100

const maxIterations = 10000

func passablePoint(mt rl.Grid, r *rand.Rand) gruid.Point {
	size := mt.Size()
	for range maxIterations {
		p := gruid.Point{r.IntN(size.X), r.IntN(size.Y)}
		if Passable(mt.At(p)) {
			return p
		}
	}
	panic("infinite loop")
}

func connex(mt rl.Grid, pr *paths.PathRange, r *rand.Rand) bool {
	pass := func(p gruid.Point) bool {
		return Passable(mt.At(p))
	}
	pr.CCMap(&MappingPath{passable: pass}, passablePoint(mt, r))
	for p, t := range mt.All() {
		if Passable(t) && pr.CCMapAt(p) == -1 {
			return false
		}
	}
	return true
}

func map2String(mt rl.Grid) string {
	var sb strings.Builder
	for p, t := range mt.All() {
		if p.X%MapWidth == 0 && p.Y > 0 {
			sb.WriteRune('\n')
		}
		switch t {
		case TranslucentWall:
			sb.WriteRune('$')
		case Wall:
			sb.WriteRune('#')
		case Floor:
			sb.WriteRune('.')
		case Foliage:
			sb.WriteRune('"')
		case Rubble:
			sb.WriteRune('^')
		}
	}
	return sb.String()
}

func TestGame(t *testing.T) {
	for range rounds {
		testGame(t)
	}
}

func testGame(t *testing.T) {
	gd := gruid.NewGrid(UIWidth, UIHeight)
	md := &model{gd: gd, g: &Game{}, targ: &targeting{}}
	md.initStructures()
	md.initWidgets()
	md.initKeys()
	g := md.g
	g.md = md
	g.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	g.Mods = make([]bool, NMods)
	g.Mods[ModCorruptedDungeon] = true
	g.Mods[ModAdvancedSpirits] = true
	// g.Mods[ModNoHealingComestibles] = true
	g.Init(spiritEntity(g.Mods, primarySpirits[g.rand.IntN(len(primarySpirits))]))
	g.ComputePlayerStats()
	g.InitLevel()
	if !connex(g.Map.Terrain, g.PR, g.rand) {
		t.Errorf("Not connex map level %d:\n%s\n", g.Map.Level, map2String(g.Map.Terrain))
	}
	for range MapLevels - 1 {
		g.NextLevel()
		if !connex(g.Map.Terrain, g.PR, g.rand) {
			t.Errorf("Not connex map level %d:\n%s\n", g.Map.Level, map2String(g.Map.Terrain))
		}
		switch {
		case g.Map.Level == MapLevels && g.Map.Orb == InvalidPos:
			t.Errorf("Missing Orb of Corruption:\n%s\n", map2String(g.Map.Terrain))
		case g.Map.Level < MapLevels && g.Map.Portal == InvalidPos:
			t.Errorf("Missing portal at level %d:\n%s\n", g.Map.Level, map2String(g.Map.Terrain))
		case g.ProcInfo.Spirits[g.Map.Level-1].Idx >= 0 && g.Map.Portal == InvalidPos:
			t.Errorf("Missing totem at level %d:\n%s\n", g.Map.Level, map2String(g.Map.Terrain))
		}
		b := map[gruid.Point]bool{}
		for i := range g.ItemEntities() {
			p := g.Entity(i).P
			if i >= FirstMapID && b[p] {
				t.Errorf("Two items in same place: %d", p)
			}
			b[p] = true
		}
		clear(b)
		for i := range g.Actors() {
			p := g.Entity(i).P
			if i >= FirstMapID && b[p] {
				t.Errorf("Two actors in same place: %d", p)
			}
			b[p] = true
		}
	}
}
