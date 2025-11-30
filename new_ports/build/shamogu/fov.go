package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// MaxFOVRange is the maximum distance in player's field of view.
const MaxFOVRange = 8

// UpdateFOV updates the field of view. It has to be called each time the
// player moves or some terrain change impacts vision.
func (g *Game) UpdateFOV() {
	pp := g.PP()
	rg := gruid.NewRange(-MaxFOVRange, -MaxFOVRange, MaxFOVRange+1, MaxFOVRange+1)
	g.Map.FOV.SetRange(rg.Add(pp).Intersect(g.Map.Terrain.Range()))
	passable := func(p gruid.Point) bool {
		t := g.Map.Terrain.At(p)
		return t != Wall && t != Rubble
	}
	pa := g.PlayerActor()
	lt := &lighter{
		g:      g,
		flying: pa.DoesAny(NocturnalFlying) && !pa.Has(StatusLignification) && !pa.Has(StatusGardener),
	}
	g.Map.FOV.VisionMap(lt, pp)
	g.Map.FOVPts = g.Map.FOV.SSCVisionMap(pp, g.MaxFOVRange(), passable, false)
}

// MaxFOVRange returns the maximum FOV range adjusted for special traits (like
// nocturnal flying).
func (g *Game) MaxFOVRange() int {
	if g.PlayerActor().DoesAny(NocturnalFlying) {
		return MaxFOVRange - 3
	}
	return MaxFOVRange
}

// UpdateKnowledge updates player's knowledge based on FOV. It has to be called at
// the end of each turn.
func (g *Game) UpdateKnowledge() {
	pp := g.PP()
	maxRange := g.MaxFOVRange()
	for _, p := range g.Map.FOVPts {
		if paths.DistanceManhattan(p, pp) > maxRange {
			continue
		}
		cost, ok := g.Map.FOV.At(p)
		if !ok || cost > maxRange {
			continue
		}
		g.SeeTerrain(p)
	}
	g.SeeEntities()
}

// SeeTerrain handles knowledge changes about a given position during a FOV
// update.
func (g *Game) SeeTerrain(p gruid.Point) {
	t := g.Map.Terrain.At(p)
	if g.Map.KnownTerrain.At(p) != t {
		g.md.auto.aemRebuild = true
		g.Map.KnownTerrain.Set(p, t)
	}
}

// SensePassable handles knowledge changes about an unknown terrain's
// passability.
func (g *Game) SensePassable(p gruid.Point) {
	if g.Map.KnownTerrain.At(p) == Unknown {
		g.md.auto.aemRebuild = true
		g.Map.KnownTerrain.Set(p, UnknownPassable)
	}
}

// SeeEntities handles knowledge changes about the last known position of
// map entities that are in the field of view.
func (g *Game) SeeEntities() {
	for i, e := range g.NPMapEntities() {
		if g.InFOV(e.P) {
			g.SenseEntity(i, "see")
		} else if g.InFOV(e.KnownP) {
			e.KnownP = InvalidPos
		}
	}
	if g.PlayerActor().Has(StatusClarity) {
		g.senseMonsters()
	}
}

// SenseEntity updates known position of an entity to its real one. The given
// verb is used in the log message (one of "see", "sense").
func (g *Game) SenseEntity(i ID, verb string) {
	ei := g.Entity(i)
	if ei.Role == nil {
		return
	}
	if !ei.Seen {
		ei.Seen = true
		g.SensePassable(ei.P)
		g.md.auto.mode = noAuto
		switch ri := ei.Role.(type) {
		case *Actor:
			if !ri.IsAlive() {
				break
			}
			g.Logf("You %s %s.", verb, One(ei.Name))
			if ri.DoesAny(MonsNotable) {
				g.StoryLogf("Saw %s (distance: %d)", One(ei.Name), g.ppDist(ei.P))
			}
			if !g.InFOV(ei.P) {
				break
			}
			pa := g.PlayerActor()
			switch {
			case pa.DoesAny(Elephanty) && ri.IsMonster(HungryRat):
				g.Logf("You trumpet at the %s", ei.Name)
				g.MakeNoise(g.PP(), NoiseTrumpet)
			case pa.DoesAny(ScaryRoar):
				// Scary roar at foe on first sight.
				// NOTE: Skipped if you already sensed
				// the monster through Clarity. May
				// happen several times in a turn if
				// you see several monsters at once.
				g.Logf("You roar at the %s", ei.Name)
				g.MakeNoise(g.PP(), NoiseRoar)
				g.PutStatus(i, ri, StatusFear, DurationFearRoar)
			}
		case *Comestible:
			g.LogfStyled("You %s %s.", logNotable, verb, One(ei.Name))
			g.StoryLogf("Spotted %s (distance: %d)", One(ei.Name), g.ppDist(ei.P))
		case *Spirit:
			g.LogfStyled("You %s %s totem.", logNotable, verb, One(ei.Name))
			g.StoryLogf("Spotted %s totem (distance: %d)", One(ei.Name), g.ppDist(ei.P))
			g.cackleAt(ei.P)
		case *EmptyTotem:
			g.LogStyled("You see an empty totem.", logNotable)
			g.StoryLogf("Spotted an empty totem (distance: %d)", g.ppDist(ei.P))
		case *CorruptionOrb:
			g.LogfStyled("You %s the Orb of Corruption!", logNotable, verb)
			g.StoryLogf("Found the Orb of Corruption (distance: %d)", g.ppDist(ei.P))
			g.cackleAt(ei.P)
		case *RunicTrap:
			g.LogfStyled("You %s %s.", logNotable, verb, One(ei.Name))
			g.StoryLogf("Spotted %s (distance: %d)", One(ei.Name), g.ppDist(ei.P))
		case *Portal:
			g.LogfStyled("You %s %s.", logNotable, verb, One(ei.Name))
			g.StoryLogf("Discovered %s (distance: %d)", One(ei.Name), g.ppDist(ei.P))
			g.cackleAt(ei.P)
		default:
			g.LogfStyled("You %s %s.", logNotable, verb, One(ei.Name))
			g.StoryLogf("Discovered %s (distance: %d)", One(ei.Name), g.ppDist(ei.P))
		}
	}
	ei.KnownP = ei.P
	switch r := ei.Role.(type) {
	case *Actor:
		r.KnownDead = r.IsDead()
	case *RunicTrap:
		if r.Used {
			r.KnownUsed = true
		}
	}
}

func (g *Game) ppDist(p gruid.Point) int {
	return paths.DistanceManhattan(g.PP(), p)
}

// cackleAt emits a cackling sound if the player has the a runic chicken spirit.
func (g *Game) cackleAt(p gruid.Point) {
	if g.PlayerActor().DoesAny(RunicChicken) && g.InFOV(p) {
		g.MakeNoise(g.PP(), NoiseCackle)
	}
}

// senseMonsters implements the monster-sensing of the player-only status
// effect Clarity.
func (g *Game) senseMonsters() {
	pp := g.PP()
	// Sense monsters and terrain under them.
	for i, ei := range g.NPMapEntities() {
		if !ei.IsActor() {
			continue
		}
		if clarityRange(pp, ei.P) {
			g.SenseEntity(i, "sense")
			g.SensePassable(ei.P)
		}
	}
}

// clarityRange reports whether p and q are within Clarity's range from each
// other.
func clarityRange(p, q gruid.Point) bool {
	return paths.DistanceManhattan(p, q) <= 2*MaxFOVRange
}

// ResetKnowledge initializes player's knowledge of map terrain.
func (g *Game) ResetKnowledge() {
	g.Map.KnownTerrain.Fill(Unknown)
	for _, e := range g.NPMapEntities() {
		e.KnownP = InvalidPos
	}
}

// InFOV reports whether p is in the player's field of view.
func (g *Game) InFOV(p gruid.Point) bool {
	cost, ok := g.Map.FOV.At(p)
	return ok && cost <= g.MaxFOVRange() && g.Map.FOV.Visible(p)
}

// DangerInFOV reports whether there is a foe in the player's field of view.
func (g *Game) DangerInFOV() bool {
	for i, ai := range g.Monsters() {
		if ai.IsAlive() && g.InFOV(g.Entity(i).P) {
			return true
		}
	}
	return g.dangerousCloudInFOV()
}

// DangerInProximity reports whether there is a sensed foe in the player's
// proximity (either from seen in FOV or from footstep noise).
func (g *Game) DangerInProximity() bool {
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if ai.IsAlive() && (g.InFOV(ei.P) || ei.Noise) {
			return true
		}
	}
	return g.dangerousCloudInFOV()
}

func (g *Game) dangerousCloudInFOV() bool {
	for _, cl := range g.Map.Clouds.Clouds {
		if g.InFOV(cl.P) && cl.Kind != CloudNormal {
			return true
		}
	}
	return false
}

// lighter implements rl.Lighter
type lighter struct {
	g      *Game
	flying bool // owl flying a bit above the ground
}

func (lt *lighter) MaxCost(src gruid.Point) int {
	if lt.flying {
		return MaxFOVRange - 2
	}
	return MaxFOVRange + 1
}

func (lt *lighter) Cost(src, from, to gruid.Point) int {
	g := lt.g
	wallcost := lt.MaxCost(src)
	// no terrain cost on origin
	if src == from {
		// specific diagonal costs
		vis := lt.diagonalVisibility(from, to)
		switch vis {
		case Opaque:
			return wallcost
		case Fuzzy:
			return wallcost - 1
		default:
			return paths.DistanceManhattan(to, from)
		}
	}
	// from terrain specific costs
	t := g.Map.Terrain.AtU(from)
	if t == Wall || t == Rubble {
		return wallcost
	}
	clk := g.Map.Clouds.At(from).Kind
	if clk == CloudNormal {
		return wallcost
	}
	// specific diagonal costs
	vis := lt.diagonalVisibility(from, to)
	if vis == Opaque {
		return wallcost
	}
	if clk != NoCloud {
		return wallcost + paths.DistanceManhattan(to, from) - 3
	}
	if t == Foliage && !lt.flying {
		return wallcost + paths.DistanceManhattan(to, from) - 3
	}
	switch vis {
	case Fuzzy:
		cost := wallcost - paths.DistanceManhattan(from, src) - 1
		return max(cost, 1)
	default:
		return paths.DistanceManhattan(to, from)
	}
}

type visibility uint32

const (
	Opaque visibility = 0b00
	Fuzzy  visibility = 0b01
	Clear  visibility = 0b11
)

func (lt *lighter) diagonalVisibility(from, to gruid.Point) visibility {
	// The state uses cardinal movement only, so two diagonal walls should,
	// for example, block line of sight. This is in contrast with the main
	// mechanics of the line of sight algorithm, which for gameplay reasons
	// allows diagonals for light rays in normal circumstances.
	var ps [2]gruid.Point
	delta := to.Sub(from)
	switch delta {
	case gruid.Point{1, -1}:
		ps[0] = from.Shift(1, 0)
		ps[1] = from.Shift(0, -1)
	case gruid.Point{-1, -1}:
		ps[0] = from.Shift(-1, 0)
		ps[1] = from.Shift(0, -1)
	case gruid.Point{-1, 1}:
		ps[0] = from.Shift(-1, 0)
		ps[1] = from.Shift(0, 1)
	case gruid.Point{1, 1}:
		ps[0] = from.Shift(1, 0)
		ps[1] = from.Shift(0, 1)
	default:
		return Clear
	}
	vis := Opaque
	g := lt.g
	for _, p := range ps {
		clk := g.Map.Clouds.At(p).Kind
		if clk == CloudNormal {
			continue
		}
		switch g.Map.Terrain.AtU(p) {
		case Wall, Rubble:
		case Foliage:
			if lt.flying {
				return Clear
			}
			vis |= Fuzzy
		default:
			if clk != NoCloud {
				vis |= Fuzzy
				continue
			}
			return Clear
		}
	}
	return vis
}
