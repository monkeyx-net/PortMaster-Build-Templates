package main

import (
	"errors"

	"codeberg.org/anaseto/gruid"
)

// Autoexplore attemps auto-exploration.
func (g *Game) Autoexplore() (again bool, err error) {
	if mons := g.MonsterInFOV(); mons.Exists() {
		return again, errors.New("You cannot auto-explore while there are monsters in view.")
	}
	if g.AllExplored() {
		return again, errors.New("Nothing left to explore.")
	}
	sources := g.AutoexploreSources()
	if len(sources) == 0 {
		return again, errors.New("Some unreachable places remain unexplored.")
	}
	g.BuildAutoexploreMap(sources)
	n, finished := g.NextAuto()
	if finished || n == nil {
		return again, errors.New("You cannot reach safely some places.")
	}
	g.autoexploring = true
	g.autoHalt = false
	return g.PlayerBump(*n)
}

// AllExplored reports whether all the reachable terrain has been explored.
func (g *Game) AllExplored() bool {
	np := &NoisePath{g: g}
	it := g.Map.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		t := it.Cell()
		if t == WallCell {
			if len(np.Neighbors(p)) == 0 {
				continue
			}
		}
		if !IsKnown(g.Map.KnownTerrain.At(p)) {
			return false
		}
	}
	return true
}

// AutoexploreSources returns source points for auto-exploration.
func (g *Game) AutoexploreSources() []gruid.Point {
	g.autosources = g.autosources[:0]
	np := &NoisePath{g: g}
	it := g.Map.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		t := it.Cell()
		if t == WallCell {
			if len(np.Neighbors(p)) == 0 {
				continue
			}
		}
		if !IsKnown(g.Map.KnownTerrain.At(p)) {
			g.autosources = append(g.autosources, p)
		}

	}
	if g.Player.Bananas < MaxBananas {
		for p := range g.Objects.Bananas {
			g.autosources = append(g.autosources, p)
		}
	}
	for p, pt := range g.Objects.Potions {
		if pt == HealthPotion && g.Player.HP < g.Player.HPMax() ||
			pt == MagicPotion && g.Player.MP < g.Player.MPMax() {
			g.autosources = append(g.autosources, p)
		}
	}
	return g.autosources
}

const unreachable = 9999

// BuildAutoexploreMap rebuilds auto-exploration data.
func (g *Game) BuildAutoexploreMap(sources []gruid.Point) {
	ap := &AutoexplorePath{g: g}
	g.PRauto.BreadthFirstMap(ap, sources, unreachable)
	g.autoexploreMapRebuild = false
}

// NextAuto returns the next auto-explore position and reports whether
// auto-explore finished.
func (g *Game) NextAuto() (next *gruid.Point, finished bool) {
	ap := &AutoexplorePath{g: g}
	if g.PRauto.BreadthFirstMapAt(g.Player.P) > unreachable {
		return nil, false
	}
	neighbors := ap.Neighbors(g.Player.P)
	if len(neighbors) == 0 {
		return nil, false
	}
	n := neighbors[0]
	ncost := g.PRauto.BreadthFirstMapAt(n)
	for _, p := range neighbors[1:] {
		cost := g.PRauto.BreadthFirstMapAt(p)
		if cost < ncost {
			n = p
			ncost = cost
		}
	}
	if ncost >= g.PRauto.BreadthFirstMapAt(g.Player.P) {
		finished = true
	}
	next = &n
	return next, finished
}

// StopAuto stops auto-movement.
func (g *Game) StopAuto() {
	if g.autoexploring && !g.autoHalt {
		g.Log("You stop exploring.")
	} else if g.autoDir != ZP {
		g.Log("You stop.")
	} else if g.autoTarget != invalidPos {
		g.Log("You stop.")
	}
	g.autoHalt = true
	g.autoDir = ZP
	g.autoTarget = invalidPos
}
