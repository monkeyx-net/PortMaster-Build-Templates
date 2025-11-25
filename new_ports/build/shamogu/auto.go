package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// auto represents information related to all kinds of auto-travel.
type auto struct {
	mode       autoMode
	delta      gruid.Point      // direction
	dirChange  bool             // change of directions (smart running)
	dirn       dirNeighbors     // neighbors passability configuration
	path       []gruid.Point    // travelling path
	sources    []gruid.Point    // auto-exploration sources
	PRauto     *paths.PathRange // path range for auto-exploration (cached)
	aemRebuild bool             // rebuild auto-explore map
}

// autoMode represents the various kinds of auto-movement.
type autoMode int

const (
	noAuto autoMode = iota
	autoRun
	autoTravel
	autoExplore
)

// UpdateRun updates auto dir movement info.
func (md *model) UpdateRun(delta gruid.Point) {
	g := md.g
	a := md.auto
	a.delta = delta
	a.mode = autoRun
	p := g.PP()
	nbs := g.dirNeighbors(p, a.delta)
	blocked := g.dirBlocked(p, a.delta)
	a.dirChange = false
	if blocked && nbs > 0 {
		if g.PlayerPassableNoTraps(left(p, a.delta)) {
			a.delta = left(p, a.delta).Sub(p)
			a.dirChange = true
		} else if g.PlayerPassableNoTraps(right(p, a.delta)) {
			a.delta = right(p, a.delta).Sub(p)
			a.dirChange = true
		}
		a.dirn = g.dirNeighbors(p, a.delta)
	} else {
		a.dirn = nbs
	}
}

// KeepRunning reports whether we should keep auto-running.
func (md *model) KeepRunning() bool {
	g := md.g
	if g.DangerInProximity() {
		return false
	}
	p := g.PP().Add(md.auto.delta)
	if !g.PlayerPassableNoTraps(p) {
		return false
	}
	nbs := g.dirNeighbors(p, md.auto.delta)
	if md.auto.dirn == nbs {
		return true
	}
	blocked := g.dirBlocked(p, md.auto.delta)
	return nbs != dirFreeLaterals && blocked ||
		nbs == dirBlockedLaterals &&
			(md.auto.dirChange ||
				g.dirNeighbors(p, md.auto.delta) == dirBlockedLaterals)
}

// dirNeighbors represents various kinds of neighbor-configurations (for
// auto-running purposes).
type dirNeighbors int

const (
	dirFreeLaterals dirNeighbors = iota
	dirBlockedLeft
	dirBlockedRight
	dirBlockedLaterals
)

func (g *Game) dirNeighbors(p, dir gruid.Point) (dn dirNeighbors) {
	if !g.PlayerPassableNoTraps(left(p, dir)) {
		dn += dirBlockedLeft
	}
	if !g.PlayerPassableNoTraps(right(p, dir)) {
		dn += dirBlockedRight
	}
	return dn
}

func (g *Game) dirBlocked(p, dir gruid.Point) bool {
	return !g.PlayerPassableNoTraps(p.Add(dir))
}

func right(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X - dir.Y, p.Y + dir.X}
}

func left(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X + dir.Y, p.Y - dir.X}
}

// UpdateTravel updates auto travel movement info, and returns the next
// position, if any.
func (md *model) UpdateTravel() (next gruid.Point, ok bool) {
	g := md.g
	if g.DangerInProximity() {
		return next, false
	}
	if len(md.auto.path) <= 1 {
		return next, false
	}
	md.auto.path = md.auto.path[1:]
	return md.auto.path[0], true
}

// UpdateAutoExplore updates auto exploration movement info, and returns next
// position to explore, if any.
func (md *model) UpdateAutoExplore() (next gruid.Point, ok bool) {
	g := md.g
	if md.auto.aemRebuild {
		if g.allExplored() {
			g.Log("You finished exploring.")
			return next, false
		}
		sources := md.autoExploreSources()
		if len(sources) == 0 {
			g.Log("You cannot reach some places.")
			return next, false
		}
		md.buildAutoExploreMap(sources)
	}
	if g.DangerInProximity() {
		if md.auto.mode == noAuto {
			g.Log("You cannot auto-explore while danger is nearby.")
		}
		return next, false
	}
	ap := &MapPath{passable: g.PlayerPassableNoTrapsFunc()}
	pp := g.PP()
	if md.auto.PRauto.BreadthFirstMapAt(pp) > unreachable {
		g.Log("You cannot reach some places.")
		return next, false
	}
	neighbors := ap.Neighbors(pp)
	if len(neighbors) == 0 {
		g.Log("You cannot reach some places.")
		return next, false
	}
	next = neighbors[0]
	ncost := md.auto.PRauto.BreadthFirstMapAt(next)
	for _, p := range neighbors[1:] {
		cost := md.auto.PRauto.BreadthFirstMapAt(p)
		if cost < ncost {
			next = p
			ncost = cost
		}
	}
	if ncost >= md.auto.PRauto.BreadthFirstMapAt(pp) {
		g.Log("You cannot reach some places.")
		return next, false
	}
	return next, true
}

func (g *Game) allExplored() bool {
	mp := &MapPath{passable: g.Map.Passable}
	for p, t := range g.Map.KnownTerrain.All() {
		if !Passable(t) {
			if len(mp.Neighbors(p)) == 0 {
				continue
			}
		}
		if !IsKnown(t) {
			return false
		}
	}
	return true
}

func (md *model) autoExploreSources() []gruid.Point {
	g := md.g
	md.auto.sources = md.auto.sources[:0]
	mp := &MapPath{passable: g.Map.Passable}
	for p, t := range g.Map.KnownTerrain.All() {
		if !Passable(t) {
			if len(mp.Neighbors(p)) == 0 {
				continue
			}
		}
		if !IsKnown(t) {
			md.auto.sources = append(md.auto.sources, p)
		}

	}
	return md.auto.sources
}

const unreachable = 9999

func (md *model) buildAutoExploreMap(sources []gruid.Point) {
	g := md.g
	ap := &MapPath{passable: g.PlayerPassableNoTrapsFunc()}
	md.auto.PRauto.BreadthFirstMap(ap, sources, unreachable)
	md.auto.aemRebuild = false
}
