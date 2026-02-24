package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"sort"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// maxIterations is used for loops that theoretically may not finish but do in
// practice, just in case.
const maxIterations = 1500

// WallCell returns a random wall position.
func (mg *MapGen) WallCell() gruid.Point {
	d := mg.m
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		t := d.At(p)
		if t == WallCell {
			return p
		}
	}
}

// MapGen is used for map generation.
type MapGen struct {
	m         *Map
	tunnel    map[gruid.Point]bool
	vault     map[gruid.Point]bool
	vaults    []*vault
	rndPos    []gruid.Point
	spl       places
	special   specialVault
	layout    maplayout
	PR        *paths.PathRange
	rand      *rand.Rand
	neighbors paths.Neighbors
}

// GenMapWithLayout generates the map using the given layout.
func (g *Game) GenMapWithLayout(ml maplayout) {
	mg := MapGen{}
	mg.rndPos = make([]gruid.Point, 0, MapWidth*MapHeight)
	for y := range MapHeight {
		for x := range MapWidth {
			mg.rndPos = append(mg.rndPos, gruid.Point{x, y})
		}
	}
	rand.Shuffle(len(mg.rndPos), func(i, j int) {
		mg.rndPos[i], mg.rndPos[j] = mg.rndPos[j], mg.rndPos[i]
	})
	mg.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	mg.PR = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	mg.layout = ml
	m := g.Map
	m.Terrain.Fill(WallCell)
	m.KnownTerrain.Fill(Unknown)
	mg.m = m
	mg.tunnel = make(map[gruid.Point]bool)
	mg.vault = make(map[gruid.Point]bool)
	mg.vaults = []*vault{}
	switch ml {
	case AutomataCave:
		mg.genCellularAutomataCaveMap()
	case RandomWalkCave:
		mg.genCaveMap(21 * 40)
	case RandomWalkTreeCave:
		mg.genTreeCaveMap()
	case RandomSmallWalkCaveUrbanised:
		mg.genCaveMap(20 * 10)
	case NaturalCave:
		if mg.rand.IntN(3) == 0 {
			mg.genCellularAutomataCaveMap()
		} else {
			mg.genCaveMap(21 * 47)
		}
	}
	var places []gruid.Point
	var nspecial = 4
	if sr := g.Params.Special[g.Map.Depth]; sr != noSpecialVault {
		nspecial--
		pl := PlacementEdge
		if mg.rand.IntN(3) == 0 {
			pl = PlacementCenter
		}
		mg.special = sr
		var ok bool
		count := 0
		for {
			places, ok = mg.genVaults(sr.Templates(), 1, pl)
			count++
			if count > 150 {
				if g.Map.Depth == WinDepth || g.Map.Depth == MaxDepth {
					panic("special room")
				}
				break
			}
			if ok {
				break
			}
		}
	}
	switch g.Map.Depth {
	case WinDepth:
		mg.genShaedraCell(g)
		nspecial--
	case MaxDepth:
		mg.genArtifactPlace(g)
		nspecial--
	}
	normal := 5
	if g.Map.Depth < 3 {
		nspecial--
		normal--
	}
	barrels := 4
	switch ml {
	case RandomWalkCave:
		barrels--
		mg.genVaults(vaultBigTemplates, nspecial-1, PlacementRandom)
		mg.genVaults(vaultNormalTemplates, normal, PlacementRandom)
	case RandomWalkTreeCave:
		mg.genVaults(vaultBigTemplates, nspecial+1, PlacementRandom)
		mg.genVaults(vaultNormalTemplates, normal+2, PlacementRandom)
	case RandomSmallWalkCaveUrbanised:
		nspecial += 3
		mg.genVaults(vaultBigTemplates, nspecial, PlacementRandom)
		mg.genVaults(vaultNormalTemplates, normal+5, PlacementRandom)
	case NaturalCave:
		barrels--
		nspecial++
		if g.Map.Depth == WinDepth {
			nspecial += 2
		}
		mg.genVaults(vaultBigTemplates, nspecial, PlacementRandom)
		mg.genVaults(vaultNormalTemplates, normal-3, PlacementRandom)
	default:
		mg.genVaults(vaultBigTemplates, nspecial, PlacementRandom)
		mg.genVaults(vaultNormalTemplates, normal+2, PlacementRandom)
	}
	mg.connectVaults()
	g.Map = m
	mg.putDoors()
	mg.playerStart(g, places)
	mg.clearUnconnected(g)
	if mg.rand.IntN(10) > 0 {
		var t rl.Cell
		if mg.rand.IntN(5) > 1 {
			t = ChasmCell
		} else {
			t = WaterCell
		}
		mg.genLake(t)
		if mg.rand.IntN(5) == 0 {
			mg.genLake(t)
		}
	}
	if g.Map.Depth < MaxDepth {
		if g.Params.Blocked[g.Map.Depth] {
			mg.genStairs(g, SealedStair)
		} else {
			mg.genStairs(g, NormalStair)
		}
		mg.genFakeStairs(g)
		if g.Map.Depth != WinDepth && mg.rand.IntN(4) == 0 {
			mg.genFakeStairs(g)
		}
	}
	for i := 0; i < barrels+mg.rand.IntN(2); i++ {
		mg.genBarrel(g)
	}
	mg.addSpecialFeatures(g, ml)
	mg.PR.CCMapAll(NewMappingPath(func(p gruid.Point) bool {
		return Passable(g.Map.At(p))
	}))
	mg.GenMonsters(g)
	mg.putCavernCells()
	if mg.rand.IntN(2) == 0 {
		mg.genQueenRock()
	}
}

// placement represents various kinds of vault placements.
type placement int

const (
	PlacementRandom placement = iota
	PlacementCenter
	PlacementEdge
)

// genVaults generates the vaults.
func (mg *MapGen) genVaults(templates []string, n int, pl placement) (ps []gruid.Point, ok bool) {
	if len(templates) == 0 {
		return nil, false
	}
	ok = true
	for range n {
		var r *vault
		count := 500
		var p gruid.Point
		var tpl string
		for r == nil && count > 0 {
			count--
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
			tpl = templates[mg.rand.IntN(len(templates))]
			r = mg.newVault(p, tpl)
		}
		if r != nil {
			switch pl {
			case PlacementCenter, PlacementEdge:
				r.special = true
			}
			mg.vaults = append(mg.vaults, r)
			ps = append(ps, p)
		} else {
			ok = false
		}
	}
	return ps, ok
}

// Newroom creates a new room at the given position of the given kind.
func (mg *MapGen) newVault(rpos gruid.Point, kind string) *vault {
	r := &vault{p: rpos, vault: &rl.Vault{}}
	err := r.vault.Parse(kind)
	if err != nil {
		log.Printf("bad vault:%v", err)
		return nil
	}
	r.w = r.vault.Size().X
	r.h = r.vault.Size().Y
	drev := 2
	if r.w > r.h+2 {
		drev += r.w - r.h - 2
		if drev > 4 {
			drev = 4
		}
	}
	if mg.rand.IntN(drev) == 0 {
		switch mg.rand.IntN(2) {
		case 0:
			r.vault.Reflect()
			r.vault.Rotate(1 + 2*mg.rand.IntN(2))
		default:
			r.vault.Rotate(1 + 2*mg.rand.IntN(2))
		}
	} else {
		switch mg.rand.IntN(2) {
		case 0:
			r.vault.Reflect()
			r.vault.Rotate(2 * mg.rand.IntN(2))
		default:
			r.vault.Rotate(2 * mg.rand.IntN(2))
		}
	}
	r.w = r.vault.Size().X
	r.h = r.vault.Size().Y
	if !r.HasSpace(mg) {
		return nil
	}
	r.Dig(mg)
	if r.w == 0 || r.h == 0 {
		log.Printf("bad vault size: %v", r.vault.Size())
	}
	return r
}

// connectVaults connects the rooms through tunnels to ensure connexity.
func (mg *MapGen) connectVaults() {
	sort.Sort(roomSlice(mg.vaults))
	for i, r := range mg.vaults {
		if i == 0 {
			continue
		}
		if r.tunnels > 0 {
			continue
		}
		nearest := mg.nearestConnectedVault(i)
		ok := mg.connectVaultsShortestPath(nearest, i)
		for !ok {
			panic("connect rooms")
		}
	}
	extraTunnels := 6
	switch mg.layout {
	case RandomSmallWalkCaveUrbanised:
		extraTunnels = 7
	case NaturalCave:
		extraTunnels = 4
	}
	count := 0
	for i, r := range mg.vaults {
		if count >= extraTunnels {
			break
		}
		if r.tunnels > 1 {
			continue
		}
		count++
		mg.connectVaultsShortestPath(i, mg.nearVault(i))
	}
	if count >= extraTunnels {
		return
	}
	for i, r := range mg.vaults {
		if count >= extraTunnels {
			break
		}
		if r.tunnels > 2 {
			continue
		}
		count++
		mg.connectVaultsShortestPath(i, mg.nearVault(i))
	}
}

// connectVaultsShortestPath connects two rooms with a paved ground tunnel.
func (mg *MapGen) connectVaultsShortestPath(i, j int) bool {
	if i == j {
		return false
	}
	r1 := mg.vaults[i]
	r2 := mg.vaults[j]
	var e1pos, e2pos gruid.Point
	var e1i, e2i int
	e1i = r1.UnusedEntry()
	e1pos = r1.entries[e1i].p
	e2i = r2.UnusedEntry()
	e2pos = r2.entries[e2i].p
	tp := &TunnelPath{dg: mg}
	path := mg.PR.AstarPath(tp, e1pos, e2pos)
	if len(path) == 0 {
		log.Printf("no path from %v to %v", e1pos, e2pos)
		return false
	}
	for _, p := range path {
		if !InMap(p) {
			panic(fmt.Sprintf("gruid.Point %v from %v to %v", p, e1pos, e2pos))
		}
		t := mg.m.At(p)
		if t == WallCell || t == ChasmCell || t == GroundCell || t == FoliageCell {
			mg.m.Set(p, GroundCell)
			mg.tunnel[p] = true
		}
	}
	r1.entries[e1i].used = true
	r2.entries[e2i].used = true
	r1.tunnels++
	r2.tunnels++
	return true
}

// roomSlice is used for sorting vaults/rooms.
type roomSlice []*vault

func (rs roomSlice) Len() int      { return len(rs) }
func (rs roomSlice) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs roomSlice) Less(i, j int) bool {
	center := gruid.Point{MapWidth / 2, MapHeight / 2}
	ipos := rs[i].p
	ipos.X += rs[i].w / 2
	ipos.Y += rs[i].h / 2
	jpos := rs[j].p
	jpos.X += rs[j].w / 2
	jpos.Y += rs[j].h / 2
	return rs[i].special || !rs[j].special && Distance(ipos, center) <= Distance(jpos, center)
}

func (mg *MapGen) nearestConnectedVault(i int) (k int) {
	r := mg.vaults[i]
	d := unreachable
	for j, nextVault := range mg.vaults[:i] {
		if j == i {
			continue
		}
		nd := vaultDistance(r, nextVault)
		if nd < d {
			d = nd
			k = j
		}
	}
	return k
}

func (mg *MapGen) nearVault(i int) (k int) {
	r := mg.vaults[i]
	d := unreachable
	for j, nextVault := range mg.vaults {
		if j == i {
			continue
		}
		nd := vaultDistance(r, nextVault)
		if nd < d {
			n := mg.rand.IntN(5)
			if n > 0 {
				d = nd
				k = j
			}
		}
	}
	return k
}

// putDoors places doors.
func (mg *MapGen) putDoors() {
	for _, r := range mg.vaults {
		for _, e := range r.entries {
			if e.used && !e.nodoor {
				r.places = append(r.places, place{p: e.p, kind: PlaceDoor})
			}
		}
		for _, pl := range r.places {
			if pl.kind != PlaceDoor {
				continue
			}
			mg.m.Set(pl.p, DoorCell)
			r.places = append(r.places, place{p: pl.p, kind: PlaceDoor})
		}
	}
}

// putHoledWalls adds n holed walls (if possible).
func (mg *MapGen) putHoledWalls(g *Game, n int) {
	candidates := []gruid.Point{}
	it := mg.m.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		if mg.vault[p] && g.holedWallCandidate(p) {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return
	}
	for range n {
		p := candidates[mg.rand.IntN(len(candidates))]
		g.Map.Set(p, HoledWallCell)
	}
}

// putWindows adds n windows (if possible).
func (mg *MapGen) putWindows(g *Game, n int) {
	candidates := []gruid.Point{}
	it := mg.m.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		if mg.vault[p] && g.holedWallCandidate(p) {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return
	}
	for range n {
		p := candidates[mg.rand.IntN(len(candidates))]
		g.Map.Set(p, WindowCell)
	}
}

// holedWallCandidate reports whether the given position is suitable for a
// holed wall.
func (g *Game) holedWallCandidate(p gruid.Point) bool {
	d := g.Map
	if !InMap(p) || d.At(p) != WallCell {
		return false
	}
	return InMap(p.Add(gruid.Point{-1, 0})) && InMap(p.Add(gruid.Point{1, 0})) &&
		d.At(p.Add(gruid.Point{-1, 0})) == WallCell && d.At(p.Add(gruid.Point{1, 0})) == WallCell &&
		InMap(p.Add(gruid.Point{0, -1})) && Passable(d.At(p.Add(gruid.Point{0, -1}))) &&
		InMap(p.Add(gruid.Point{0, 1})) && Passable(d.At(p.Add(gruid.Point{0, 1}))) ||
		(InMap(p.Add(gruid.Point{-1, 0})) && InMap(p.Add(gruid.Point{1, 0})) &&
			Passable(d.At(p.Add(gruid.Point{-1, 0}))) && Passable(d.At(p.Add(gruid.Point{1, 0}))) &&
			InMap(p.Add(gruid.Point{0, -1})) && d.At(p.Add(gruid.Point{0, -1})) == WallCell &&
			InMap(p.Add(gruid.Point{0, 1})) && d.At(p.Add(gruid.Point{0, 1})) == WallCell)
}

// genShaedraCell places Shaedra, Marevor, and the monolith.
func (mg *MapGen) genShaedraCell(g *Game) {
	g.Objects.Story = map[gruid.Point]Story{}
	g.Places.Shaedra = mg.spl.Shaedra
	g.Objects.Story[g.Places.Shaedra] = StoryShaedra
	g.Places.Monolith = mg.spl.Monolith
	g.Objects.Story[g.Places.Monolith] = NoStory
	g.Places.Marevor = mg.spl.Marevor
	g.Objects.Story[g.Places.Marevor] = NoStory
}

// genArtifactPlace places the artifact and the monolith.
func (mg *MapGen) genArtifactPlace(g *Game) {
	g.Objects.Story = map[gruid.Point]Story{}
	g.Places.Artifact = mg.spl.Artifact
	g.Objects.Story[g.Places.Artifact] = StoryArtifactSealed
	g.Places.Monolith = mg.spl.Monolith
	g.Objects.Story[g.Places.Monolith] = NoStory
	g.Places.Marevor = mg.spl.Marevor
	g.Objects.Story[g.Places.Marevor] = NoStory
}

// mapLayout represents the various kinds of map layouts.
type maplayout int

const (
	AutomataCave maplayout = iota
	RandomWalkCave
	RandomWalkTreeCave
	RandomSmallWalkCaveUrbanised
	NaturalCave
)

// putCavernCells ensures distinction between paved ground and natural cave
// ground (for patrol purposes).
func (mg *MapGen) putCavernCells() {
	d := mg.m
	it := mg.m.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		if it.Cell() == GroundCell && !mg.vault[p] && !mg.tunnel[p] {
			d.Set(p, CavernCell)
		}
	}
}

// clearUnconnected ensures that no unconnected terrain remains (before adding
// lakes or chasm).
func (dg *MapGen) clearUnconnected(g *Game) {
	d := dg.m
	sp := NewMappingPath(func(p gruid.Point) bool { return IsPlayerPassable(d.At(p)) })
	dg.PR.CCMap(sp, g.Player.P)
	mg := rl.MapGen{Grid: dg.m.Terrain}
	mg.KeepCC(dg.PR, g.Player.P, rl.Cell(WallCell))
}

// addSpecialFeatures handles placement of stones, bananas, potions, tables,
// lights, trees, holed walls, windows, and lore.
func (mg *MapGen) addSpecialFeatures(g *Game, ml maplayout) {
	g.Objects.Stones = map[gruid.Point]Stone{}
	if g.Params.Blocked[g.Map.Depth] || g.Map.Depth == MaxDepth {
		mg.genSealStone(g)
	}
	bananas := 1
	bananas += g.Params.ExtraBanana[g.Map.Depth]
	for i := 0; i < bananas; i++ {
		mg.genBanana(g)
	}
	if !g.Params.NoMagara[g.Map.Depth] {
		mg.genMagara(g)
	}
	mg.genItem(g)
	mg.genPotion(g, MagicPotion)
	if g.Params.HealthPotion[g.Map.Depth] {
		mg.genPotion(g, HealthPotion)
	}
	mg.genStones(g)
	ntables := 4
	switch ml {
	case AutomataCave, RandomWalkCave, NaturalCave:
		if mg.rand.IntN(3) == 0 {
			ntables++
		} else if mg.rand.IntN(10) == 0 {
			ntables--
		}
	case RandomWalkTreeCave:
		if mg.rand.IntN(4) > 0 {
			ntables++
		}
		if mg.rand.IntN(4) > 0 {
			ntables++
		}
	case RandomSmallWalkCaveUrbanised:
		ntables += 2
		if mg.rand.IntN(4) > 0 {
			ntables++
		}
	}
	if g.Params.Tables[g.Map.Depth] {
		ntables += 2 + mg.rand.IntN(2)
	}
	for i := 0; i < ntables; i++ {
		mg.genTable(g)
	}
	mg.genLights(g)
	ntrees := 1
	switch ml {
	case AutomataCave:
		if mg.rand.IntN(4) == 0 {
			ntrees++
		} else if mg.rand.IntN(8) == 0 {
			ntrees--
		}
	case RandomWalkCave:
		if mg.rand.IntN(4) > 0 {
			ntrees++
		}
		if mg.rand.IntN(8) == 0 {
			ntrees++
		}
	case NaturalCave:
		ntrees++
		if mg.rand.IntN(2) > 0 {
			ntrees++
		}
	case RandomWalkTreeCave, RandomSmallWalkCaveUrbanised:
		if mg.rand.IntN(2) == 0 {
			ntrees--
		}
	}
	if g.Params.Trees[g.Map.Depth] {
		ntrees += 2 + mg.rand.IntN(2)
	}
	for i := 0; i < ntrees; i++ {
		mg.genTree(g)
	}
	nhw := 1
	if mg.rand.IntN(3) > 0 {
		nhw++
	}
	if g.Params.Holes[g.Map.Depth] {
		nhw += 3 + mg.rand.IntN(2)
	}
	switch ml {
	case RandomSmallWalkCaveUrbanised:
		if mg.rand.IntN(4) > 0 {
			nhw++
		}
	}
	mg.putHoledWalls(g, nhw)
	nwin := 1
	if nhw == 1 {
		nwin++
	}
	if g.Params.Windows[g.Map.Depth] {
		nwin += 4 + mg.rand.IntN(3)
	}
	switch ml {
	case RandomSmallWalkCaveUrbanised:
		if mg.rand.IntN(4) > 0 {
			nwin++
		}
	}
	mg.putWindows(g, nwin)
	if g.Params.Lore[g.Map.Depth] {
		mg.putLore(g)
	}
}

// putLore generates a lore message.
func (mg *MapGen) putLore(g *Game) {
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > 2000 {
			panic("infinite loop")
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceItem)
	}
	count = 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		i := mg.rand.IntN(len(LoreMessages))
		if g.ProcInfo.GeneratedLore[i] {
			continue
		}
		g.ProcInfo.GeneratedLore[i] = true
		g.Objects.Lore[p] = i
		g.Objects.Scrolls[p] = ScrollLore
		g.Map.Set(p, ScrollCell)
		break
	}
}

// genLights generates lights.
func (mg *MapGen) genLights(g *Game) {
	no := 2
	ni := 8
	switch mg.layout {
	case NaturalCave:
		no += mg.rand.IntN(2)
		ni += mg.rand.IntN(3)
	case AutomataCave, RandomWalkCave:
		ni += mg.rand.IntN(4)
	case RandomWalkTreeCave:
		no--
		ni += mg.rand.IntN(4)
	case RandomSmallWalkCaveUrbanised:
		no--
		no -= mg.rand.IntN(2)
		ni += 2
		ni += mg.rand.IntN(4)
	}
	for i := 0; i < no; i++ {
		p := mg.outsideGroundCell(g)
		g.Map.Set(p, LightCell)
	}
	for i := 0; i < ni; i++ {
		p := mg.vaults[RandInt(len(mg.vaults))].RandomPlaces(PlaceSpecialOrStatic)
		if p != invalidPos {
			g.Map.Set(p, LightCell)
		} else if mg.rand.IntN(10) > 0 {
			i--
		}
	}
	it := g.Map.Terrain.Iterator()
	for it.Next() {
		if it.Cell() == rl.Cell(LightCell) {
			g.Objects.Lights[it.P()] = true
		}
	}
	g.ComputeLights()
}

// playerStart generates the player's starting room features.
func (mg *MapGen) playerStart(g *Game, places []gruid.Point) {
	const far = 30
	r := mg.vaults[len(mg.vaults)-1]
loop:
	for i := len(mg.vaults) - 2; i >= 0; i-- {
		for _, p := range places {
			if Distance(r.p, p) < far && Distance(mg.vaults[i].p, p) >= far {
				r = mg.vaults[i]
				mg.vaults[len(mg.vaults)-1], mg.vaults[i] = mg.vaults[i], mg.vaults[len(mg.vaults)-1]
				break loop
			}
		}
	}
	g.Player.P = r.RandomPlace(PlacePatrol)
	switch g.Map.Depth {
	case 1, 4:
	default:
		return
	}
	itpos := invalidPos
	neighbors := g.PlayerPassableNeighbors(g.Player.P)
	for i := range len(neighbors) {
		j := RandInt(len(neighbors) - i)
		neighbors[i], neighbors[j] = neighbors[j], neighbors[i]
	}
loopnb:
	for _, npos := range neighbors {
		t := g.Map.At(npos)
		if IsGround(t) {
			for _, pl := range r.places {
				if npos == pl.p {
					continue loopnb
				}
			}
			itpos = npos
			break
		}
	}
	if itpos == invalidPos {
		itpos = r.RandomPlace(PlaceItem)
	}
	if itpos == invalidPos {
		itpos = r.RandomPlaces(PlaceSpecialOrStatic)
		if itpos == invalidPos {
			panic("no item")
		}
	}
	g.Map.Set(itpos, ScrollCell)
	switch g.Map.Depth {
	case 1:
		g.Objects.Scrolls[itpos] = ScrollStory
	case 4:
		g.Objects.Scrolls[itpos] = ScrollDayoriahMessage
	}
}

// genBanana places a banana at a suitable location.
func (mg *MapGen) genBanana(g *Game) {
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		t := mg.m.At(p)
		if t == GroundCell && !mg.vault[p] {
			mg.m.Set(p, BananaCell)
			g.Objects.Bananas[p] = true
			break
		}
	}
}

// genPotion places a potion at a suitable vault location.
func (mg *MapGen) genPotion(g *Game, ptn Potion) {
	count := 0
	p := invalidPos
	for p == invalidPos {
		count++
		if count > maxIterations {
			return
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceItem)
	}
	mg.m.Set(p, PotionCell)
	g.Objects.Potions[p] = ptn
}

// outsideGroundCell returns a random non-vault ground position.
func (mg *MapGen) outsideGroundCell(g *Game) gruid.Point {
	count := 0
	for {
		count++
		if count > 1500 {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if t == GroundCell && !mg.vault[p] {
			return p
		}
	}
}

// outsideCavernMiddleCell returns a random non-vault ground or foliage
// position (with various biases).
func (mg *MapGen) outsideCavernMiddleCell(g *Game) gruid.Point {
	count := 0
	for {
		count++
		if count > 2500 {
			return invalidPos
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if t == GroundCell && count < 400 || t == FoliageCell && count < 350 {
			continue
		}
		if IsGround(t) && !mg.vault[p] && !mg.m.HasTooManyWallNeighbors(p) {
			return p
		}
	}
}

// HasTooManyWallNeighbors reports whether the given position has more than 2
// wall neighbors (excluding boundaries).
func (m *Map) HasTooManyWallNeighbors(p gruid.Point) bool {
	count := 0
	for q := range AllNeighbors(p) {
		if !Passable(m.At(q)) {
			count++
		}
	}
	return count > 1
}

// genItem generates a magical equipment item (cloak or amulet).
func (mg *MapGen) genItem(g *Game) {
	plan := g.ProcInfo.GenPlan[g.Map.Depth]
	if plan != GenAmulet && plan != GenCloak {
		return
	}
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceItem)
	}
	g.Map.Set(p, ItemCell)
	var it Item
	switch plan {
	case GenCloak:
		it = g.RandomCloak()
		g.ProcInfo.GeneratedCloaks = append(g.ProcInfo.GeneratedCloaks, it.Kind)
	case GenAmulet:
		it = g.RandomAmulet()
		g.ProcInfo.GeneratedAmulets = append(g.ProcInfo.GeneratedAmulets, it.Kind)
	}
	g.Objects.Items[p] = it
}

// genSealStone handles steal stone generation.
func (mg *MapGen) genSealStone(g *Game) {
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlaces(PlaceSpecialOrStatic)
	}
	g.Map.Set(p, StoneCell)
	g.Objects.Stones[p] = SealStone
}

// genMagara handles magara generation.
func (mg *MapGen) genMagara(g *Game) {
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceItem)
	}
	g.Map.Set(p, MagaraCell)
	mag := g.RandomMagara()
	g.Objects.Magaras[p] = mag
	g.ProcInfo.GeneratedMagaras = append(g.ProcInfo.GeneratedMagaras, mag.Kind)
}

// genStairs generates real stairs (possibly sealed).
func (mg *MapGen) genStairs(g *Game, st Stair) {
	var ri, pj int
	best := 0
	for i, r := range mg.vaults {
		for j, pl := range r.places {
			score := Distance(pl.p, g.Player.P) + mg.rand.IntN(20)
			if !pl.used && pl.kind == PlaceSpecialStatic && score > best {
				ri = i
				pj = j
				best = Distance(pl.p, g.Player.P)
			}
		}
	}
	r := mg.vaults[ri]
	r.places[pj].used = true
	r.places[pj].used = true
	p := r.places[pj].p
	g.Map.Set(p, StairCell)
	g.Objects.Stairs[p] = st
}

// genFakeStairs generates fake stairs.
func (mg *MapGen) genFakeStairs(g *Game) {
	if !g.Params.FakeStair[g.Map.Depth] {
		return
	}
	var ri, pj int
	best := 0
loop:
	for i, r := range mg.vaults {
		for _, pl := range r.places {
			if mg.m.At(pl.p) == StairCell {
				continue loop
			}
		}
		for j, pl := range r.places {
			score := Distance(pl.p, g.Player.P) + mg.rand.IntN(20)
			if !pl.used && pl.kind == PlaceSpecialStatic && score > best {
				ri = i
				pj = j
				best = Distance(pl.p, g.Player.P)
			}
		}
	}
	r := mg.vaults[ri]
	r.places[pj].used = true
	r.places[pj].used = true
	p := r.places[pj].p
	g.Map.Set(p, StairCell)
	g.Objects.Stairs[p] = FakeStair
}

// genBarrel places a barrel in a free suitable location.
func (mg *MapGen) genBarrel(g *Game) {
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > 500 {
			return
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceSpecialStatic)
	}
	g.Map.Set(p, BarrelCell)
	g.Objects.Barrels[p] = true
}

// genTable places a table in a free suitable location.
func (mg *MapGen) genTable(g *Game) {
	p := invalidPos
	count := 0
	for p == invalidPos {
		count++
		if count > 500 {
			return
		}
		p = mg.vaults[RandInt(len(mg.vaults))].RandomPlaces(PlaceSpecialOrStatic)
	}
	g.Map.Set(p, TableCell)
}

// genTree places a tree in a free suitable outside location.
func (mg *MapGen) genTree(g *Game) {
	p := mg.outsideCavernMiddleCell(g)
	if p != invalidPos {
		g.Map.Set(p, TreeCell)
	}
}

// genStones places magical stones.
func (mg *MapGen) genStones(g *Game) {
	nstones := 2 // number of stones
	outside := 0 // number of stones outside vaults
	if g.rand.IntN(3) > 0 {
		outside++
	}
	if mg.rand.IntN(1+2*g.Map.Depth) >= 2 {
		nstones++
		if g.rand.IntN(3) == 0 {
			outside++
		}
	}
	if g.Params.Stones[g.Map.Depth] {
		nstones++
		if g.rand.IntN(3) == 0 {
			outside++
		}
	}
	genStones := []Stone{}
	for i := range nstones {
		p := invalidPos
		var st Stone
		if i < nstones-outside {
			// Place a stone inside a vault.
			count := 0
			for p == invalidPos {
				count++
				if count > 1500 {
					p = mg.caveGroundCell()
					break
				}
				p = mg.vaults[RandInt(len(mg.vaults))].RandomPlace(PlaceStatic)
			}
			st = mg.randomInStone(genStones)
			genStones = append(genStones, st)
		} else {
			// The map contains a magical stone outside from a
			// vault.
			p = mg.caveGroundCell()
			st = mg.randomOutStone(genStones, outside)
			genStones = append(genStones, st)
		}
		g.Objects.Stones[p] = st
		g.Map.Set(p, StoneCell)
	}
}

func (mg *MapGen) caveGroundCell() gruid.Point {
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		t := mg.m.At(p)
		if (t == GroundCell || t == CavernCell || t == QueenRockCell) && !mg.vault[p] {
			return p
		}
	}
}

func (mg *MapGen) randomInStone(sts []Stone) Stone {
	stones := []Stone{
		BarrelStone,
		EarthStone,
		QueenStone,
		TeleportStone,
	}
	for {
		st := stones[RandInt(len(stones))]
		if !slices.Contains(sts, st) {
			return st
		}
	}
}

func (mg *MapGen) randomOutStone(sts []Stone, outside int) Stone {
	stones := []Stone{
		NightStone,
		TreeStone,
	}
	if mg.rand.IntN(2) == 0 || outside > len(stones) {
		// Stones that usually appear in vaults may appear outside,
		// too, but less often.
		stones = append(stones, BarrelStone, EarthStone, QueenStone, TeleportStone)
	}
	for {
		st := stones[mg.rand.IntN(len(stones))]
		if !slices.Contains(sts, st) {
			return st
		}
	}
}

func (mg *MapGen) genCellularAutomataCaveMap() {
	mg.runCellularAutomataCave()
	mg.foliage(false)
}

func (dg *MapGen) runCellularAutomataCave() bool {
	rules := []rl.CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 3, WallsOutOfRange: true},
	}
	mg := rl.MapGen{Rand: dg.rand, Grid: dg.m.Terrain}
	if dg.rand.IntN(2) == 0 {
		mg.CellularAutomataCave(rl.Cell(WallCell), rl.Cell(GroundCell), 0.47, rules)
	} else {
		mg.CellularAutomataCave(rl.Cell(WallCell), rl.Cell(GroundCell), 0.43, rules)
	}
	return true
}

// genLake generates a lake using the given terrain for filling.
func (mg *MapGen) genLake(fillt rl.Cell) {
	walls := []gruid.Point{}
	xshift := 10 + mg.rand.IntN(5)
	yshift := 5 + mg.rand.IntN(3)
	it := mg.m.Terrain.Iterator()
	for it.Next() {
		p := it.P()
		if p.X < xshift || p.Y < yshift || p.X > MapWidth-xshift || p.Y > MapHeight-yshift {
			continue
		}
		t := it.Cell()
		if t == WallCell && !mg.vault[p] {
			walls = append(walls, p)
		}
	}
	count := 0
	var bestpos = walls[RandInt(len(walls))]
	var bestsize int
	d := mg.m
	passable := func(p gruid.Point) func(q gruid.Point) bool {
		return func(q gruid.Point) bool {
			return InMap(q) && mg.m.At(q) == WallCell && !mg.vault[q] && Distance(p, q) < 10+mg.rand.IntN(10)
		}
	}
	for {
		p := walls[RandInt(len(walls))]
		sp := NewMappingPath(passable(p))
		size := len(mg.PR.CCMap(sp, p))
		count++
		if Abs(bestsize-90) > Abs(size-90) {
			bestsize = size
			bestpos = p
		}
		if count > 15 || Abs(size-90) < 25 {
			break
		}
	}
	sp := NewMappingPath(passable(bestpos))
	for _, p := range mg.PR.CCMap(sp, bestpos) {
		d.Set(p, fillt)
	}
}

// genQueenRock places queenrock tiles.
func (mg *MapGen) genQueenRock() {
	cavern := []gruid.Point{}
	for i := range MapSize {
		p := Idx2Pos(i)
		t := mg.m.At(p)
		if t == CavernCell {
			cavern = append(cavern, p)
		}
	}
	if len(cavern) == 0 {
		return
	}
	for i := 0; i < 1+mg.rand.IntN(2); i++ {
		p := cavern[RandInt(len(cavern))]
		passable := func(q gruid.Point) bool {
			return InMap(q) && mg.m.At(q) == CavernCell && Distance(q, p) < 15+mg.rand.IntN(5)
		}
		cp := NewMappingPath(passable)
		for _, q := range mg.PR.CCMap(cp, p) {
			mg.m.Set(q, QueenRockCell)
		}
	}

}

// foliage generates foliage.
func (dg *MapGen) foliage(less bool) {
	gd := rl.NewGrid(MapWidth, MapHeight)
	rules := []rl.CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 2, WallsOutOfRange: true},
	}
	mg := rl.MapGen{Rand: dg.rand, Grid: gd}
	winit := 0.55
	if less {
		winit = 0.53
	}
	mg.CellularAutomataCave(rl.Cell(WallCell), rl.Cell(FoliageCell), winit, rules)
	it := dg.m.Terrain.Iterator()
	itgd := gd.Iterator()
	for it.Next() && itgd.Next() {
		if it.Cell() == GroundCell && itgd.Cell() == FoliageCell {
			it.SetCell(rl.Cell(FoliageCell))
		}
	}
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

func (dg *MapGen) genCaveMap(size int) {
	mg := rl.MapGen{
		Rand: dg.rand,
		Grid: dg.m.Terrain,
	}
	wlk := walker{rand: dg.rand}
	mg.RandomWalkCave(wlk, rl.Cell(GroundCell), float64(size)/float64(MapSize), 8)
	dg.foliage(false)
}

func (mg *MapGen) DigBlock(block []gruid.Point) []gruid.Point {
	d := mg.m
	p := mg.WallCell()
	block = block[:0]
	count := 0
	for {
		count++
		if count > 3000 && count%500 == 0 {
			p = mg.WallCell()
			block = block[:0]
		}
		if count > 10000 {
			panic("infinite loop")
		}
		block = append(block, p)
		if d.HasFreeNeighbor(p) {
			break
		}
		p = RandomNeighbor(p)
		if !InMap(p) {
			block = block[:0]
			p = mg.WallCell()
			continue
		}
		if !InMap(p) {
			return nil
		}
	}
	return block
}

func (mg *MapGen) genTreeCaveMap() {
	d := mg.m
	center := gruid.Point{40, 10}
	d.Set(center, GroundCell)
	d.Set(center.Add(gruid.Point{1, 0}), GroundCell)
	d.Set(center.Add(gruid.Point{1, -1}), GroundCell)
	d.Set(center.Add(gruid.Point{0, 1}), GroundCell)
	d.Set(center.Add(gruid.Point{1, 1}), GroundCell)
	d.Set(center.Add(gruid.Point{0, -1}), GroundCell)
	d.Set(center.Add(gruid.Point{-1, -1}), GroundCell)
	d.Set(center.Add(gruid.Point{-1, 0}), GroundCell)
	d.Set(center.Add(gruid.Point{-1, 1}), GroundCell)
	max := 21 * 21
	cells := 1
	block := make([]gruid.Point, 0, 64)
loop:
	for cells < max {
		block = mg.DigBlock(block)
		if len(block) == 0 {
			continue loop
		}
		for _, p := range block {
			if d.At(p) != GroundCell {
				d.Set(p, GroundCell)
				cells++
			}
		}
	}
	mg.foliage(true)
}
