package main

import (
	"math/rand/v2"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// MapPath implements the paths.Pather interface and is used to provide pathing
// information in map generation.
type MapPath struct {
	passable func(gruid.Point) bool
	nbs      paths.Neighbors
}

func (mp *MapPath) Neighbors(p gruid.Point) []gruid.Point {
	return mp.nbs.Cardinal(p, mp.passable)
}

// MappingPath implements the paths.Pather interface and is used to provide pathing
// information in magic mapping and for connected components (filling interior
// walls).
type MappingPath struct {
	passable func(p gruid.Point) bool
	nbs      paths.Neighbors
}

func (mp *MappingPath) Neighbors(p gruid.Point) []gruid.Point {
	if !mp.passable(p) {
		// We don't want to mark interior walls as explored.
		return nil
	}
	return mp.nbs.Cardinal(p, inMap)
}

// PlayerPath returns a player path between two map positions.
func (g *Game) PlayerPath(from, to gruid.Point) []gruid.Point {
	var ppassNT, ppass func(gruid.Point) bool
	switch g.Wizard.Mode {
	case WizardRevealTerrain:
		return nil
	case WizardReveal:
		ppassNT, ppass = g.WizardPlayerPassableNoTrapsFunc(), g.WizardPlayerPassableFunc()
	default:
		ppassNT, ppass = g.PlayerPassableNoTrapsFunc(), g.PlayerPassableFunc()
	}
	var path []gruid.Point
	if g.NoTrapAt(to) && !g.PlayerActor().DoesAny(RunicChicken) {
		path = g.PR.JPSPath(path, from, to, ppassNT, false)
	}
	if len(path) == 0 {
		// If there is a trap in the destination or there's no other
		// available known path, we don't avoid traps, as that's what
		// the player is asking for.
		path = g.PR.JPSPath(path, from, to, ppass, false)
	}
	return path
}

// PlayerPassableNoTrapsFunc returns a passability function suited for
// computing a player path. It considers tiles containing traps.
func (g *Game) PlayerPassableNoTrapsFunc() func(gruid.Point) bool {
	knownAt := g.Map.KnownTerrain.At
	if g.PlayerActor().Statuses.Has(StatusDig) {
		return func(p gruid.Point) bool {
			return inMap(p) && IsKnown(knownAt(p)) && g.NoTrapAt(p)
		}
	}
	passable := g.Map.Passable
	return func(p gruid.Point) bool {
		return passable(p) && IsKnown(knownAt(p)) && g.NoTrapAt(p)
	}
}

// PlayerPassableFunc returns a passability function suited for computing a
// player path. It considers tiles containing traps as passable.
func (g *Game) PlayerPassableFunc() func(gruid.Point) bool {
	knownAt := g.Map.KnownTerrain.At
	if g.PlayerActor().Statuses.Has(StatusDig) {
		return func(p gruid.Point) bool {
			return inMap(p) && knownAt(p) != Unknown
		}
	}
	passable := g.Map.Passable
	return func(p gruid.Point) bool {
		return passable(p) && knownAt(p) != Unknown
	}
}

// PlayerPassableNoTraps reports whether a given position is passable for the player.
func (g *Game) PlayerPassableNoTraps(p gruid.Point) bool {
	return g.PlayerPassableNoTrapsFunc()(p)
}

// WizardPlayerPassableNoTrapsFunc returns a passability function suited for
// computing a wizard player path. It considers tiles containing traps.
func (g *Game) WizardPlayerPassableNoTrapsFunc() func(gruid.Point) bool {
	if g.PlayerActor().Statuses.Has(StatusDig) {
		return func(p gruid.Point) bool { return inMap(p) && g.NoTrapAt(p) }
	}
	passable := g.Map.Passable
	return func(p gruid.Point) bool { return passable(p) && g.NoTrapAt(p) }
}

// WizardPlayerPassableFunc returns a passability function suited for computing
// a wizard player path. It considers tiles containing traps as passable.
func (g *Game) WizardPlayerPassableFunc() func(gruid.Point) bool {
	if g.PlayerActor().Statuses.Has(StatusDig) {
		return func(p gruid.Point) bool { return inMap(p) }
	}
	passable := g.Map.Passable
	return func(p gruid.Point) bool { return passable(p) }
}

// MonPath implements paths.Astar for monster movement.
type MonPath struct {
	passable func(gruid.Point) bool // map passability for given monster
	nbs      paths.Neighbors
	actorAt  func(gruid.Point) bool // whether there's an actor at a given position
	pp       gruid.Point            // player's position
	rand     *rand.Rand
	confused bool // confused monsters won't move around other ones
}

func (aip *MonPath) Neighbors(p gruid.Point) []gruid.Point {
	nbs := aip.nbs.Cardinal(p, aip.passable)
	// Shuffle so that monster movement is not unnaturally predictable.
	aip.rand.Shuffle(len(nbs), func(i, j int) { nbs[i], nbs[j] = nbs[j], nbs[i] })
	return nbs
}

func (aip *MonPath) Cost(from, to gruid.Point) int {
	if !aip.confused && aip.actorAt(to) && to != aip.pp {
		return 5
	}
	return 1
}

func (aip *MonPath) Estimation(from, to gruid.Point) int {
	return paths.DistanceManhattan(from, to)
}

// MonPath returns a path between two positions for a monster actor.
func (g *Game) MonPath(i ID, ai *Actor, from, to gruid.Point) []gruid.Point {
	pp := g.PP()
	ei := g.Entity(i)
	beh := ei.Role.(*Actor).Behavior
	passable := g.Map.Passable
	if beh.State != Hunting {
		passable = g.Map.PassableWithoutTraps
	}
	aip := &MonPath{
		actorAt:  g.Map.BoolCache.At,
		pp:       pp,
		passable: passable,
		rand:     g.rand,
		confused: ai.Has(StatusConfusion)}
	var path []gruid.Point
	if beh.State != Hunting && paths.DistanceManhattan(from, pp) > MaxFOVRange &&
		paths.DistanceManhattan(from, to) > MaxFOVRange {
		path = g.PR.JPSPath(beh.Path, from, to, aip.passable, false)
	}
	if len(path) == 0 {
		// We use a boolean cache marking actor positions for faster
		// cost computations during AstarPath.
		g.Map.BoolCache = g.Map.BoolCache.New()
		for _, ne := range g.Entities {
			if ne.IsAlive() {
				g.Map.BoolCache.Set(ne.P, true)
			}
		}
		if beh.State == Hunting && ai.DoesAny(MonsDig) {
			aip.passable = func(p gruid.Point) bool { return true }
		}
		path = g.PR.AstarPath(aip, from, to)
	}
	if len(path) == 0 && beh.State != Hunting {
		// Ignore traps if there's no other way.
		aip.passable = g.Map.Passable
		path = g.PR.AstarPath(aip, from, to)
	}
	if len(path) == 0 {
		return nil
	}
	return path
}

// tunnelPath computes tunnel paths in map generation.
type tunnelPath struct {
	mg  *MapGen
	nbs paths.Neighbors
}

func (tp *tunnelPath) Neighbors(p gruid.Point) []gruid.Point {
	return tp.nbs.Cardinal(p, inMap)
}

func (tp *tunnelPath) Cost(from, to gruid.Point) int {
	if tp.mg.vault.At(to) {
		if !tp.mg.tunnel.At(to) {
			// Tunnels should not go through non-entry vault cells.
			return 100
		}
		// Tunnels should rarely go through another vault entry.
		return 10
	}
	t := tp.mg.terrain.At(to)
	if Passable(t) {
		if !tp.mg.tunnel.At(to) && t != Floor {
			// Give non-Floor non-tunnel cells a little extra cost.
			return 2
		}
		// Making a tunnel to a free cell is cheap.
		return 1
	}
	// When we have a wall, we favor tunnels that continue through internal
	// walls.
	cost := 3
	tf := tp.mg.terrain.At(from)
	if from.X == MapWidth-1 || from.X == 0 || from.Y == 0 || from.Y == MapHeight-1 || Passable(tf) {
		// Slight penalty for tunnels that are on the edge of the map
		// because they don't look very nice.
		// Starting a tunnel costs a little too.
		cost++
	}
	wc := countLateralWalls(tp.mg.terrain, from, to)
	return cost - wc
}

func countLateralWalls(gd rl.Grid, from, to gruid.Point) int {
	n := 0
	dir := to.Sub(from)
	for p := range Neighbors(to) {
		if p.Sub(to) == dir || p == from {
			continue
		}
		if gd.At(p) == rl.Cell(Wall) {
			n++
		}
	}
	return n
}

func (tp *tunnelPath) Estimation(from, to gruid.Point) int {
	return paths.DistanceManhattan(from, to)
}
