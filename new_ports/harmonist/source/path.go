package main

import (
	"slices"
	"sort"

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

// MappingPath implements the paths.Pather interface and is used to provide
// pathing for connected components during map generation.
type MappingPath struct {
	passable  func(gruid.Point) bool
	neighbors paths.Neighbors
}

// NewMappingPath returns a new mapping pather.
func NewMappingPath(passable func(gruid.Point) bool) *MappingPath {
	return &MappingPath{passable: passable}
}

func (sp *MappingPath) Neighbors(p gruid.Point) []gruid.Point {
	if !sp.passable(p) {
		return nil
	}
	return sp.neighbors.Cardinal(p, sp.passable)
}

// GridPath is a pather for the grid (no obstacles).
type GridPath struct {
	nbs paths.Neighbors
}

func (gp *GridPath) Neighbors(p gruid.Point) []gruid.Point {
	return gp.nbs.Cardinal(p, InMap)
}

// OricExplosionPath is a pather for oric explosions.
type OricExplosionPath struct {
	m   *Map
	nbs paths.Neighbors
}

func (op *OricExplosionPath) Neighbors(p gruid.Point) []gruid.Point {
	return op.nbs.Cardinal(p, InMap)
}

func (op *OricExplosionPath) Cost(from, to gruid.Point) int {
	if IsDiggable(op.m.At(from)) {
		// Double cost for walls, to avoid destroying too much interior
		// walls.
		return 3
	}
	return 1
}

// EarthStonePath is used for earth stone.
type EarthStonePath struct {
	m   *Map
	nbs paths.Neighbors
}

func (dp *EarthStonePath) Neighbors(p gruid.Point) []gruid.Point {
	if dp.m.At(p) == WallCell {
		return nil
	}
	return dp.nbs.Cardinal(p, InMap)
}

// TunnelPath represents a pather for mapgen tunneling.
type TunnelPath struct {
	dg  *MapGen
	nbs paths.Neighbors
}

func (tp *TunnelPath) Neighbors(p gruid.Point) []gruid.Point {
	return tp.nbs.Cardinal(p, InMap)
}

func (tp *TunnelPath) Cost(from, to gruid.Point) int {
	if tp.dg.vault[from] && !tp.dg.tunnel[from] {
		return 50
	}
	cost := 1
	t := tp.dg.m.At(from)
	if tp.dg.vault[from] {
		cost += 7
	} else if !tp.dg.tunnel[from] && t != GroundCell {
		cost++
	}
	if Passable(t) {
		return cost
	}
	wc := countWalls(tp.dg.m.Terrain, from, 1, true)
	return cost + 8 - wc
}

func countWalls(gd rl.Grid, p gruid.Point, radius int, countOut bool) int {
	count := 0
	rg := gruid.Range{
		gruid.Point{p.X - radius, p.Y - radius},
		gruid.Point{p.X + radius + 1, p.Y + radius + 1},
	}
	if countOut {
		osize := rg.Size()
		rg = rg.Intersect(gd.Range())
		size := rg.Size()
		count += osize.X*osize.Y - size.X*size.Y
	} else {
		rg = rg.Intersect(gd.Range())
	}
	gd = gd.Slice(rg)
	count += gd.Count(rl.Cell(WallCell))
	return count
}

func (tp *TunnelPath) Estimation(from, to gruid.Point) int {
	return Distance(from, to)
}

// JumpPath represents a pather for jump-target purposes.
type JumpPath struct {
	g   *Game
	nbs paths.Neighbors
}

func (jp *JumpPath) Neighbors(p gruid.Point) []gruid.Point {
	nbs := jp.nbs.Cardinal(p, jp.g.PlayerCanPass)
	nbs = jp.g.ShufflePos(nbs)
	return nbs
}

// NoisePath represents a pather for noise (which passes through everything
// except plain walls).
type NoisePath struct {
	g   *Game
	nbs paths.Neighbors
}

func (fp *NoisePath) Neighbors(p gruid.Point) []gruid.Point {
	d := fp.g.Map
	keep := func(q gruid.Point) bool {
		return InMap(q) && d.At(q) != WallCell
	}
	return fp.nbs.Cardinal(p, keep)
}

// AutoexplorePath is a pather for auto-exploring.
type AutoexplorePath struct {
	g   *Game
	nbs paths.Neighbors
}

func (ap *AutoexplorePath) Neighbors(p gruid.Point) []gruid.Point {
	d := ap.g.Map.KnownTerrain
	keep := func(q gruid.Point) bool {
		kt := d.At(q)
		if cld, ok := ap.g.Map.Clouds[q]; ok && cld == CloudFire {
			// XXX little info leak
			return false
		}
		if !InMap(q) {
			return false
		}
		return IsPlayerPassable(kt) && kt != WallCell
	}
	nbs := ap.nbs.Cardinal(p, keep)
	return nbs
}

func (ap *AutoexplorePath) Cost(from, to gruid.Point) int {
	return 1
}

// MonPath represents a pather for monsters.
type MonPath struct {
	g       *Game
	monster *Monster
	nbs     paths.Neighbors
}

func (mp *MonPath) CanPassDestruct(p gruid.Point) bool {
	m := mp.monster
	if m.Kind != MonsEarthDragon {
		return mp.CanPass(p)
	}
	if !InMap(p) {
		return false
	}
	g := mp.g
	t := g.Map.At(p)
	return m.CanPass(g, p) || IsDestructible(t)
}

func (mp *MonPath) CanPatrolPass(p gruid.Point) bool {
	m := mp.monster
	if !m.CanPass(mp.g, p) {
		return false
	}
	t := mp.g.Map.At(p)
	return IsNormalPatrolWay(t)
}

func (mp *MonPath) CanPass(p gruid.Point) bool {
	return mp.monster.CanPass(mp.g, p)
}

func (mp *MonPath) Neighbors(p gruid.Point) []gruid.Point {
	keep := func(q gruid.Point) bool {
		return mp.CanPassDestruct(q)
	}
	nbs := mp.nbs.Cardinal(p, keep)
	// shuffle so that monster movement is not unnaturally predictable
	nbs = mp.g.ShufflePos(nbs)
	return nbs
}

// ShufflePos shuffles a slice of positions.
func (g *Game) ShufflePos(ps []gruid.Point) []gruid.Point {
	for i := range len(ps) {
		j := i + g.randInt(len(ps)-i)
		ps[i], ps[j] = ps[j], ps[i]
	}
	return ps
}

func (mp *MonPath) Cost(from, to gruid.Point) int {
	g := mp.g
	mons := g.MonsterAt(to)
	if !mons.Exists() {
		t := g.Map.At(to)
		if mp.monster.Kind == MonsEarthDragon && IsDestructible(t) && !mp.monster.Status(MonsConfused) {
			return 5
		}
		if to == g.Player.P && mp.monster.Kind.Peaceful() {
			switch mp.monster.Kind {
			case MonsEarthDragon:
				return 1
			default:
				return 4
			}
		}
		if mp.monster.Kind.Patrolling() && mp.monster.State != Hunting && !IsNormalPatrolWay(t) {
			return 4
		}
		return 1
	}
	if mons.Status(MonsLignified) {
		return 8
	}
	return 6
}

func (mp *MonPath) Estimation(from, to gruid.Point) int {
	return Distance(from, to)
}

// APath returns a monster path between two positions.
func (m *Monster) APath(g *Game, from, to gruid.Point) []gruid.Point {
	mp := &MonPath{g: g, monster: m}
	var path []gruid.Point
	if mp.monster.State != Hunting && Distance(from, mp.g.Player.P) > DefaultMonsterFOVRange &&
		Distance(from, to) > DefaultFOVRange && m.Kind != MonsEarthDragon {
		if mp.monster.Kind.Patrolling() {
			path = g.PR.JPSPath(m.Path, from, to, mp.CanPatrolPass, false)
		} else {
			path = g.PR.JPSPath(m.Path, from, to, mp.CanPass, false)
		}
	}
	if len(path) == 0 {
		path = g.PR.AstarPath(mp, from, to)
	}
	if len(path) == 0 {
		return nil
	}
	return path
}

// PlayerPath returns a player-passable path between two positions.
func (g *Game) PlayerPath(from, to gruid.Point) []gruid.Point {
	path := []gruid.Point{}
	path = g.PR.JPSPath(path, from, to, g.ppPassable, false)
	if len(path) == 0 {
		return nil
	}
	return path
}

func (g *Game) ppPassable(p gruid.Point) bool {
	if cld, ok := g.Map.Clouds[p]; ok && cld == CloudFire && g.Player.Sees(p) {
		return false
	}
	kt := g.Map.KnownTerrain.At(p)
	if g.Wizard == WizardSeeAll {
		kt = g.Map.Terrain.At(p)
	}
	return InMap(p) && IsKnown(kt) && (IsPlayerPassable(kt) ||
		g.Player.HasStatus(StatusLevitation) && (kt == BarrierCell || kt == ChasmCell) ||
		g.Player.HasStatus(StatusDig) && IsDiggable(kt))
}

// SortedNearestTo sorts some positions by distance.
func (g *Game) SortedNearestTo(cells []gruid.Point, to gruid.Point) []gruid.Point {
	ps := slices.Clone(cells)
	sort.Slice(ps, func(i, j int) bool {
		return paths.DistanceManhattan(ps[i], to) < paths.DistanceManhattan(ps[j], to)
	})
	return ps
}
