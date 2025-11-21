package main

import (
	"cmp"
	"log"
	"math/rand/v2"
	"slices"
	"sort"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// MapLayout represents the various kinds of map base algorithms.
type MapLayout int

const (
	CellularAutomataCave MapLayout = iota
	RandomWalkCave
	RandomWalkTreeCave
	MixedAutomataWalkCave
	MixedAutomataWalkTreeCave
	MixedWalkCaveWalkTreeCave
)

// NLayouts represents the number of different kinds of map layouts.
const NLayouts = 3

// MapGen gathers all terrain and vault information for generating a new map.
type MapGen struct {
	terrain rl.Grid         // map terrain
	vaults  []*vault        // list of special vaults
	tunnel  CacheGrid[bool] // in tunnel
	xtunnel []gruid.Point   // points belonging to extra tunnels
	vault   CacheGrid[bool] // in vault
	PR      *paths.PathRange
	rand    *rand.Rand
}

// GenerateMap produces a new map level using the given map layout as base
// generation algorithm. It generates map entities too.
func (g *Game) GenerateMap(ml MapLayout) {
	var ok bool
	var mg *MapGen // map generator (contains extra information)
	for {
		if mg, ok = g.generateMap(ml); ok {
			break
		}
	}
	if g.ProcInfo.Earthquake == g.Map.Level {
		g.earthquake()
	}
	g.computeWaypoints(mg)
	g.GenEntities(mg)
	g.Map.Clouds.Reset()
}

// generateMap produces a new map (without entities).
func (g *Game) generateMap(ml MapLayout) (*MapGen, bool) {
	mt := g.Map.Terrain
	mt.Fill(Wall)
	mg := &MapGen{terrain: mt, rand: g.rand, PR: g.PR,
		tunnel: make(CacheGrid[bool], MapWidth*MapHeight),
		vault:  make(CacheGrid[bool], MapWidth*MapHeight),
	}
	halves := func() (rl.Grid, rl.Grid) {
		left := mg.terrain.Slice(mg.terrain.Range().Columns(0, 3+MapWidth/2))
		right := mg.terrain.Slice(mg.terrain.Range().Columns(-3+MapWidth/2, MapWidth))
		if mg.rand.IntN(2) == 0 {
			return right, left
		}
		return left, right
	}
	switch ml {
	case CellularAutomataCave:
		mg.genCellularAutomataCaveMap(mg.terrain)
	case RandomWalkCave:
		mg.genCaveMap(mg.terrain)
	case RandomWalkTreeCave:
		mg.genTreeCaveMap(mg.terrain)
	case MixedAutomataWalkCave:
		gd1, gd2 := halves()
		mg.genCellularAutomataCaveMap(gd1)
		mg.genCaveMap(gd2)
	case MixedAutomataWalkTreeCave:
		gd1, gd2 := halves()
		mg.genCellularAutomataCaveMap(gd1)
		mg.genTreeCaveMap(gd2)
	case MixedWalkCaveWalkTreeCave:
		gd1, gd2 := halves()
		mg.genCaveMap(gd1)
		mg.genTreeCaveMap(gd2)
	default:
		// should not happen.
		mg.genCellularAutomataCaveMap(mg.terrain)
	}

	mg.genFoliage()
	// Vault placement.
	bigCenter := g.IntN(2) == 0
	if bigCenter {
		mg.genVaults(BigVaultTemplates, 1, PlacementCenter)
		mg.genVaults(SmallVaultTemplates, 1, PlacementEdge)
	} else {
		mg.genVaults(BigVaultTemplates, 1, PlacementEdge)
		mg.genVaults(SmallVaultTemplates, 1, PlacementCenter)
	}
	mg.genVaults(BigVaultTemplates, 1, PlacementRandom)
	mg.genVaults(SmallVaultTemplates, 4+mg.rand.IntN(2), PlacementRandom)
	mg.connectAllVaults()
	// Ensure that all the cells are connected.
	pass := func(p gruid.Point) bool {
		return Passable(mt.At(p))
	}
	freep := mg.randomFloor()
	mg.PR.CCMap(&MappingPath{passable: pass}, freep)
	mgen := rl.MapGen{Rand: mg.rand, Grid: mg.terrain}
	ntiles := mgen.KeepCC(mg.PR, freep, Wall)
	// Report whether the generated map has enough passable terrain.
	const minCaveSize = 1000
	return mg, ntiles > minCaveSize
}

// genCellularAutomataCaveMap generates a floor/wall map with a cellular automata.
func (mg *MapGen) genCellularAutomataCaveMap(gd rl.Grid) {
	// map generator using the rl package from gruid
	mgen := rl.MapGen{Rand: mg.rand, Grid: gd}
	// cellular automata map generation with rules that give a cave-like
	// map.
	rules := []rl.CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 3, WallsOutOfRange: true},
	}
	n := 0.42
	switch mg.rand.IntN(3) {
	case 1:
		n = 0.45
	case 2:
		n = 0.48
	}
	mgen.CellularAutomataCave(Wall, Floor, n, rules)
}

// genCaveMap generates a floor/wall map with a simple random walk. The grid gd
// should not be too small.
func (mg *MapGen) genCaveMap(gd rl.Grid) {
	sz := gd.Size()
	w := ((37 + mg.rand.IntN(6)) * sz.X) / MapWidth
	size := sz.Y * w
	mgen := rl.MapGen{
		Rand: mg.rand,
		Grid: gd,
	}
	wlk := walker{rand: mg.rand}
	walks := ((7 + mg.rand.IntN(3)) * sz.X) / MapWidth
	mgen.RandomWalkCave(wlk, rl.Cell(Floor), float64(size)/float64(sz.X*sz.Y), walks)
}

// walker implements rl.RandomWalker.
type walker struct {
	rand *rand.Rand
}

// Neighbor returns a random neighbor position, favoring horizontal directions
// (because the maps we use are longer in that direction).
func (w walker) Neighbor(p gruid.Point) gruid.Point {
	switch w.rand.IntN(6) {
	case 0, 1:
		return p.Shift(1, 0)
	case 2, 3:
		return p.Shift(-1, 0)
	case 4:
		return p.Shift(0, 1)
	default:
		return p.Shift(0, -1)
	}
}

// genTreeCaveMap generates a floor/wall map with random walk (repeatedly
// starting from wall until new block is connected to the center).
func (mg *MapGen) genTreeCaveMap(gd rl.Grid) {
	sz := gd.Size()
	center := gruid.Point{sz.X / 2, sz.Y / 2}
	center.X += -2 + mg.rand.IntN(3)
	center.Y += -1 + mg.rand.IntN(2)
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			gd.Set(center.Add(gruid.Point{x, y}), Floor)
		}
	}
	w := (((30 + mg.rand.IntN(5)) * sz.X) / MapWidth)
	maxCells := sz.Y * w
	cells := 1
	block := make([]gruid.Point, 0, 64)
	for cells < maxCells {
		block = mg.digBlock(gd, block)
		if len(block) == 0 {
			continue
		}
		for _, p := range block {
			if gd.At(p) != Floor {
				gd.Set(p, Floor)
				cells++
			}
		}
	}
}

func (mg *MapGen) digBlock(gd rl.Grid, block []gruid.Point) []gruid.Point {
	const maxIterations = 10000
	p := mg.randomWallCell(gd)
	block = block[:0]
	count := 0
	for {
		count++
		if count >= 1000 && count%500 == 0 {
			// We haven't found a connected part yet, so try
			// digging again from another wall.
			p = mg.randomWallCell(gd)
			block = block[:0]
		}
		if count > maxIterations {
			panic("infinite loop")
		}
		block = append(block, p)
		if passableNeighbor(gd, p) {
			break
		}
		p = mg.randomNeighbor(p)
		if !inMap(p) {
			block = block[:0]
			p = mg.randomWallCell(gd)
			continue
		}
		if !inMap(p) {
			return nil
		}
	}
	return block
}

func (mg *MapGen) randomWallCell(gd rl.Grid) gruid.Point {
	sz := gd.Size()
	const maxIterations = 10000
	for range maxIterations {
		x := mg.rand.IntN(sz.X)
		y := mg.rand.IntN(sz.Y)
		p := gruid.Point{x, y}
		if gd.At(p) == Wall {
			return p
		}
	}
	panic("no random wall found")
}

func (mg *MapGen) randomNeighbor(p gruid.Point) gruid.Point {
	switch mg.rand.IntN(6) {
	case 0, 1:
		return p.Add(gruid.Point{1, 0})
	case 2, 3:
		return p.Add(gruid.Point{-1, 0})
	case 4:
		return p.Add(gruid.Point{0, 1})
	default:
		return p.Add(gruid.Point{0, -1})
	}
}

func passableNeighbor(gd rl.Grid, p gruid.Point) bool {
	for range NeighborsFunc(p, func(q gruid.Point) bool { return Passable(gd.At(q)) }) {
		return true
	}
	return false
}

// genFoliage generates foliage with a cellular automata.
func (mg *MapGen) genFoliage() {
	size := mg.terrain.Size()
	// We use a new grid layer for drawing the foliage.
	gd := rl.NewGrid(size.X, size.Y)
	rules := []rl.CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 2, WallsOutOfRange: true},
	}
	rlmg := rl.MapGen{Rand: mg.rand, Grid: gd}
	winit := 0.54
	switch mg.rand.IntN(3) {
	case 1:
		winit = 0.53
	case 2:
		winit = 0.55
	}
	rlmg.CellularAutomataCave(Wall, Foliage, winit, rules)
	// We now draw the foliage on floor cells of the real map.
	it := mg.terrain.Iterator()
	itgd := gd.Iterator()
	for it.Next() && itgd.Next() {
		if it.Cell() == Floor && itgd.Cell() == Foliage {
			it.SetCell(Foliage)
		}
	}
}

// randomFloor returns a random floor cell in the map. It assumes that such a
// floor cell exists (otherwise the function does not end).
func (mg *MapGen) randomFloor() gruid.Point {
	size := mg.terrain.Size()
	for {
		p := gruid.Point{mg.rand.IntN(size.X), mg.rand.IntN(size.Y)}
		if mg.terrain.At(p) == Floor {
			return p
		}
	}
}

// Placement represents various methods of vault Placement in map.
type Placement int

const (
	PlacementRandom Placement = iota
	PlacementCenter
	PlacementEdge
)

// genVaults generates a given number of vaults following a set of templates
// and a placement rule.
func (mg *MapGen) genVaults(templates []string, n int, pl Placement) (ps []gruid.Point, ok bool) {
	if len(templates) == 0 {
		return nil, false
	}
	ok = true
	for i := 0; i < n; i++ {
		var vi *vault
		var p gruid.Point
		var tpl string
		for range 500 {
			tpl = templates[mg.rand.IntN(len(templates))]
			for range 10 {
				// We try to place the randomly selected vault
				// template several times before trying a new
				// one.
				vi = mg.newVault(tpl, pl)
				if vi != nil {
					break
				}
			}
			if vi != nil {
				break
			}
		}
		if vi != nil {
			mg.vaults = append(mg.vaults, vi)
			ps = append(ps, p)
		} else {
			log.Printf("mapgen: found no space for vault")
			ok = false
		}
	}
	return ps, ok
}

// newVault returns a new vault at the given position following a
// string template. It may rotate/reflect the template.
func (mg *MapGen) newVault(tpl string, pl Placement) *vault {
	var p gruid.Point
	switch pl {
	case PlacementRandom:
		p = gruid.Point{mg.rand.IntN(MapWidth - 1), mg.rand.IntN(MapHeight - 1)}
	case PlacementCenter:
		p = gruid.Point{MapWidth/2 - 4 + mg.rand.IntN(5), MapHeight/2 - 3 + mg.rand.IntN(4)}
	case PlacementEdge:
		if mg.rand.IntN(2) == 0 {
			p = gruid.Point{mg.rand.IntN(MapWidth / 4), mg.rand.IntN(MapHeight - 1)}
		} else {
			p = gruid.Point{3*MapWidth/4 + mg.rand.IntN(MapWidth/4) - 1, mg.rand.IntN(MapHeight - 1)}
		}
	}
	v := &vault{p: p, vault: &rl.Vault{}}
	err := v.vault.Parse(tpl)
	if err != nil {
		log.Printf("mapgen: bad vault:%v", err)
		return nil
	}
	v.w = v.vault.Size().X
	v.h = v.vault.Size().Y
	drev := 2
	if v.w > v.h+2 {
		drev += v.w - v.h - 2
		if drev > 4 {
			drev = 4
		}
	}
	if mg.rand.IntN(drev) == 0 {
		switch mg.rand.IntN(2) {
		case 0:
			v.vault.Reflect()
			v.vault.Rotate(1 + 2*mg.rand.IntN(2))
		default:
			v.vault.Rotate(1 + 2*mg.rand.IntN(2))
		}
	} else {
		switch mg.rand.IntN(2) {
		case 0:
			v.vault.Reflect()
			v.vault.Rotate(2 * mg.rand.IntN(2))
		default:
			v.vault.Rotate(2 * mg.rand.IntN(2))
		}
	}
	v.w = v.vault.Size().X
	v.h = v.vault.Size().Y
	if !mg.canFitVault(v) {
		return nil
	}
	mg.DigVault(v)
	if v.w == 0 || v.h == 0 {
		log.Printf("mapgen: bad vault size: %v", v.vault.Size())
	}
	return v
}

// canFitVault reports whether the vault fits in the map without intersecting
// with another vault.
func (mg *MapGen) canFitVault(v *vault) bool {
	if MapWidth-v.p.X < v.w || MapHeight-v.p.Y < v.h {
		return false
	}
	for i := v.p.X - 1; i <= v.p.X+v.w; i++ {
		for j := v.p.Y - 1; j <= v.p.Y+v.h; j++ {
			p := gruid.Point{i, j}
			if inMap(p) && mg.vault.AtU(p) {
				return false
			}
		}
	}
	return true
}

// nearestConnectedVault returns the nearest vault among the ones that
// have already been connected (less than the given index).
func (mg *MapGen) nearestConnectedVault(i int) (vr *vault) {
	// Assumes i > 0.
	vi := mg.vaults[i]
	d := unreachable
	for j, vj := range mg.vaults[:i] {
		if j == i {
			continue
		}
		nd := vaultDistance(vi, vj)
		if nd < d {
			d = nd
			vr = vj
		}
	}
	return vr
}

// nearVault returns a random vault near the given one.
func (mg *MapGen) nearVault(v *vault) (vr *vault) {
	vaults := slices.Clone(mg.vaults)
	slices.SortFunc(vaults, func(vi, vj *vault) int {
		return cmp.Compare(vaultDistance(v, vi), vaultDistance(v, vj))
	})
	for j, vi := range vaults {
		if v == vi {
			continue
		}
		if vr != nil && j > 0 && vaultDistance(v, vi) > 40 {
			break
		}
		vr = vi
		if j > 2 || j == 2 && mg.rand.IntN(4) > 0 || mg.rand.IntN(2) == 0 {
			break
		}
	}
	return vr
}

// vaultSlice is a slice of vaults implementing sort.Interface.
type vaultSlice []*vault

func (rs vaultSlice) Len() int      { return len(rs) }
func (rs vaultSlice) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs vaultSlice) Less(i, j int) bool {
	center := gruid.Point{MapWidth / 2, MapHeight / 2}
	ipos := rs[i].p
	ipos.X += rs[i].w / 2
	ipos.Y += rs[i].h / 2
	jpos := rs[j].p
	jpos.X += rs[j].w / 2
	jpos.Y += rs[j].h / 2
	return paths.DistanceManhattan(ipos, center) <= paths.DistanceManhattan(jpos, center)
}

// connectAllVaults connects all the vaults using tunnels as necessary.
func (mg *MapGen) connectAllVaults() {
	sort.Sort(vaultSlice(mg.vaults))
	for i, vi := range mg.vaults {
		if i == 0 {
			continue
		}
		if vi.tunnels > 0 {
			continue
		}
		nearest := mg.nearestConnectedVault(i)
		ok := mg.connectVaultPair(nearest, vi, false)
		if !ok {
			panic("connect rooms")
		}
	}
	extraTunnels := 4
	switch mg.rand.IntN(6) {
	case 0:
		extraTunnels = 3
	case 1:
		extraTunnels = 5
	}
	count := 0
	for n := range 2 {
		for _, vi := range mg.vaults {
			if count >= extraTunnels {
				break
			}
			if vi.tunnels > n+1 {
				continue
			}
			count++
			mg.connectVaultPair(vi, mg.nearVault(vi), true)
		}
	}
}

// connectVaultPair connects a pair of vaults using a shortest path according
// to tunnelPath.
func (mg *MapGen) connectVaultPair(v1, v2 *vault, extra bool) bool {
	if v1 == v2 {
		return false
	}
	var e1pos, e2pos gruid.Point
	var e1i, e2i int
	e1i = mg.UnusedEntry(v1)
	e1pos = v1.entries[e1i].p
	e2i = mg.UnusedEntry(v2)
	e2pos = v2.entries[e2i].p
	tp := &tunnelPath{mg: mg}
	path := mg.PR.AstarPath(tp, e1pos, e2pos)
	if len(path) == 0 {
		// Should not happen.
		log.Print("mapgen: no path connecting vaults")
		return false
	}
	// log.Printf("path length: %d", len(path))
	fill := Floor
	switch mg.rand.IntN(8) {
	case 0:
		fill = Rubble
	case 1:
		fill = Foliage
	}
	for _, p := range path {
		if t := mg.terrain.At(p); !Passable(t) {
			if mg.rand.IntN(2) == 0 {
				mg.terrain.Set(p, Floor)
			} else {
				mg.terrain.Set(p, fill)
			}
			if extra {
				// This tile belongs to an extra tunnel
				// exclusively.
				mg.xtunnel = append(mg.xtunnel, p)
			}
		}
		mg.tunnel.Set(p, true)
	}
	v1.entries[e1i].used = true
	v2.entries[e2i].used = true
	v1.tunnels++
	v2.tunnels++
	return true
}

// earthquake produces extra rubble.
func (g *Game) earthquake() {
	n := g.Map.Terrain.Count(Wall)
	walls := make([]gruid.Point, 0, n)
	it := g.Map.Terrain.Iterator()
	for it.Next() {
		if it.Cell() == Wall {
			walls = append(walls, it.P())
		}
	}
	for range len(walls)/6 + g.IntN(len(walls)/15) {
		i := g.IntN(len(walls))
		p := walls[i]
		for !g.Map.PassableNeighbor(p) {
			i = g.IntN(len(walls))
			p = walls[i]
		}
		walls[i], walls[len(walls)-1] = walls[len(walls)-1], walls[i]
		walls = walls[:len(walls)-1]
		g.Map.Terrain.Set(p, Rubble)
	}
}

// computeWaypoints computes the list of vault patrol points.
func (g *Game) computeWaypoints(mg *MapGen) {
	g.Map.Waypoints = g.Map.Waypoints[:0]
	for _, vi := range mg.vaults {
		for _, pl := range vi.places {
			if pl.kind == PlaceWaypoint && g.Map.Passable(pl.p) {
				g.Map.Waypoints = append(g.Map.Waypoints, pl.p)
			}
		}
	}
}
