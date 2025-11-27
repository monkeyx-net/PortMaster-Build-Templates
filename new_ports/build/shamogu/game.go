package main

import (
	"math/rand/v2"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Version is the game's version string of the last release.
const Version = "v1.1.0"

// InvalidPos is a special variable containing an invalid position.
var InvalidPos = gruid.Point{-1, -1}

// ID represents an entity identifier. Valid identifiers start from 0.
// Negative may indicate that the entity no longer exists or never
// existed. The first IDs get some special use, see FirstMapID.
type ID int32

// PrimarySpiritID is the ID of the first item in the inventory.
const PrimarySpiritID = 0

// InventorySize defines the size of the inventory, accounting for different
// kind of items.
const InventorySize = NSpirits + NComestibles

// FirstMapID gives the first identifier that corresponds to a non-inventory
// entity. All the IDs preceding FirstMapID are reserved for inventory items.
const FirstMapID = InventorySize + PrimarySpiritID

// PlayerID is the player's entity index in game.Entities. This convention is
// enforced for convenience.
const PlayerID = FirstMapID

// Game represents information relevant the current Game's state. It contains
// all the information to map entities, as well as other relevant informations,
// such as statistics.
type Game struct {
	Entities []*Entity        // entity ID index: entity components
	Map      *Map             // current level's map
	PR       *paths.PathRange // path range for most pathfinding
	PRnoise  *paths.PathRange // path range for noise computations (to avoid conflict with PR)

	Dir  gruid.Point // last bump-direction
	Prev gruid.Point // previous position of player
	Turn int         // current game's turn

	Logs     *Logs     // game log
	ProcInfo *ProcInfo // information for procedural generation
	Stats    *Stats    // game statistics
	Version  string    // game's version
	Wizard   Wizard    // wizard mode (default: no)

	instant bool       // whether the last action was instant-effect
	win     bool       // whether we just won the game
	rand    *rand.Rand // random number generator
	md      *model     // reference to the UI model
}

// These constants define the common width and height of all game maps.
const (
	MapWidth  = 80
	MapHeight = 21
)

// Init initializes main game structures related to map and entities.
// It performs only general initialization not specific to a particular map.
func (g *Game) Init() {
	g.Version = Version
	g.InitProcInfo()
	g.Stats = newStats()
	g.Logs = &Logs{}
	g.Entities = make([]*Entity, InventorySize+1)
	g.Map = NewMap() // initialize empty map object
	g.PR = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	g.PRnoise = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	g.Dir = gruid.Point{1, 0}
	g.InitPlayer()
}

// InitLevel generates a new map level.
func (g *Game) InitLevel() {
	g.Map.Level++
	g.GenerateMap(g.ProcInfo.Layouts[g.Map.Level-1])
	g.ResetKnowledge()
	g.UpdateFOV()
	g.UpdateKnowledge()
}

// EndTurn processes everything that happens between the end of last player's
// turn and the beginning of the next turn.
func (g *Game) EndTurn() {
	clear(g.Map.Noise)
	if g.endTurnEarly() {
		return
	}
	pa := g.PlayerActor()
	// We collect actor ids and shuffle monster ones so that we process
	// monster turns in a non-predictable order. NOTE: alternatively, we
	// could maybe process them in some clearly predictable order, like
	// maybe distance to target, favoring tighter line formations, helping
	// various kinds of beam-like attacks, but it'd be a bit harder to
	// explain to the player (even though it might actually feel more
	// intuitive in some ways in practice as it'd often seem as if monsters
	// cooperate).
	ids := []ID{PlayerID}
	for i := range g.Monsters() {
		ids = append(ids, i)
	}
	mids := ids[1:] // monster ids (excluding player)
	g.rand.Shuffle(len(mids), func(i, j int) { mids[i], mids[j] = mids[j], mids[i] })
	for _, i := range ids {
		if pa.IsDead() {
			// If player is dead, make remaining monsters skip
			// their turn.
			break
		}
		p := g.PP()
		g.HandleActorTurn(i, g.Entity(i).Actor())
		if p != g.PP() || i == PlayerID {
			g.UpdateFOV()
			g.UpdateKnowledge()
			if i != PlayerID {
				// Extra short animation frame if the player
				// was moved during a monster turn.
				g.md.AnimationFrame()
			}
		}
	}
	if pa.IsDead() {
		// The player died before turn ended.
		return
	}
	// Clouds are updated after all actors acted.
	g.UpdateClouds()
	if pa.IsDead() {
		// The player died due to a fire cloud.
		return
	}
	// We update FOV again after all monsters moved and clouds were
	// updated.
	g.UpdateFOV()
	g.UpdateKnowledge()
	g.ComputeNoise()
	g.HandleTurnCount()
}

// HandleTurnCount() increments the turn count and handles watching.
func (g *Game) HandleTurnCount() {
	g.IncrTurn()
	nmt := g.Stats.MapTurns[g.Map.Level-1]
	switch nmt {
	case 950:
		g.LogStyled("You feel a presence searching for intruders.", logSpecial)
		if g.Map.Portal != InvalidPos {
			g.LogStyled("You need to hurry to the portal!", logSpecial)
		} else {
			g.LogStyled("You need to hurry!", logSpecial)
		}
		g.md.mode = modeCritical
	case 1000:
		g.LogStyled("You've been found by the dungeon core!", logSpecial)
		g.StoryLog("Found by the dungeon core!")
		g.md.mode = modeCritical
		g.genMonster(BlazingGolem)
		g.genMonster(BlazingGolem)
	}
}

// Marked reports whether you've been found by the Orb of Corruption.
func (g *Game) Marked() bool {
	return g.Stats.MapTurns[g.Map.Level-1] >= 1000
}

// IncrTurn() increments the turn count.
func (g *Game) IncrTurn() {
	g.Turn++
	g.Stats.MapTurns[g.Map.Level-1]++
}

// endTurnEarly handles early-return cases (instant effects and frozen time)
// and reports whether early return actually happened.
func (g *Game) endTurnEarly() bool {
	if g.instant {
		g.UpdateFOV()
		g.UpdateKnowledge()
		return true
	}
	pa := g.PlayerActor()
	if !pa.Has(StatusTimeStop) {
		return false
	}
	g.HandleActorTurn(PlayerID, pa)
	if pa.IsDead() {
		// The player died before turn ended.
		return true
	}
	g.UpdateFOV()
	g.UpdateKnowledge()
	return true
}

// ComputeNoise generates footstep noise information.
func (g *Game) ComputeNoise() {
	pa := g.PlayerActor()
	for _, ei := range g.NPMapEntities() {
		ei.Noise = false
	}
	if pa.Has(StatusClarity) {
		// When you have clarity, you have senses with strictly better
		// range.
		return
	}
	dij := &MapPath{passable: g.Map.Passable}
	goodhearing := pa.DoesAny(GoodHearing) && !pa.DoesAny(BadHearing)
	badhearing := !pa.DoesAny(GoodHearing) && pa.DoesAny(BadHearing)
	const maxdist = MaxFOVRange + 5
	g.PRnoise.BreadthFirstMap(dij, []gruid.Point{g.PP()}, maxdist)
	nheard := 0  // number of heard monsters
	var a *Actor // actor corresponding to heard monster (when there's only one)
	for _, ei := range g.NPMapEntities() {
		ai, ok := ei.Role.(*Actor)
		if !ok {
			continue
		}
		if ai.IsDead() || g.InFOV(ei.P) {
			continue
		}
		dist := g.PRnoise.BreadthFirstMapAt(ei.P)
		if dist > maxdist {
			continue
		}
		switch {
		case goodhearing:
			// Good hearing generally increases noise hearing range by 2.
			dist -= 2
		case badhearing && ai.DoesAny(MonsCreep|MonsWingFlap) && dist > 2:
			// Wing flapping and creeping can't be heard well with
			// bad (mostly ground vibration-based) hearing.
			continue
		}
		switch {
		case ai.DoesAny(MonsSilent):
			if !goodhearing {
				// Silent monsters like spiders cannot be
				// heared without good hearing.
				continue
			}
			// Good hearing allows to hear almost silent
			// monsters like spiders, but not too far.
			dist += 4
		case ai.DoesAny(MonsHeavyFootsteps):
			// Heavy footsteps can be heared from a greater distance.
			dist -= 3
		case ai.DoesAny(MonsLightFootsteps | MonsCreep):
			// Light footsteps and creeping are slightly harder to
			// hear.
			dist += 2
		}
		if ei.Noise = dist <= MaxFOVRange; !ei.Noise {
			continue
		}
		g.SensePassable(ei.P)
		a = ai
		nheard++
	}
	if nheard == 1 {
		g.Logf("You hear %s.", monsterNoiseString(a))
	} else if nheard > 1 {
		g.Logf("You hear %d monsters.", nheard)
	}
}

func monsterNoiseString(ai *Actor) string {
	switch {
	case ai.DoesAny(MonsSilent):
		return "an imperceptible air movement"
	case ai.DoesAny(MonsHeavyFootsteps):
		return "heavy footsteps"
	case ai.DoesAny(MonsWingFlap):
		return "the flapping of wings"
	case ai.DoesAny(MonsCreep):
		return "a creep noise"
	case ai.DoesAny(MonsLightFootsteps):
		return "light footsteps"
	default:
		return "footsteps"
	}
}

func monsterNoiseStringStyled(ai *Actor) string {
	switch {
	case ai.DoesAny(MonsSilent):
		return "an @Yimperceptible air movement@N"
	case ai.DoesAny(MonsHeavyFootsteps):
		return "@Mheavy footsteps@N"
	case ai.DoesAny(MonsWingFlap):
		return "the @Cflapping of wings@N"
	case ai.DoesAny(MonsCreep):
		return "a @Gcreep noise@N"
	case ai.DoesAny(MonsLightFootsteps):
		return "@Olight footsteps@N"
	default:
		return "@Rfootsteps@N"
	}
}

func noiseColor(e *Entity) gruid.Color {
	a, ok := e.Role.(*Actor)
	if !ok {
		// Should not happen.
		return ColorOrange
	}
	switch {
	case a.DoesAny(MonsSilent):
		return ColorYellow
	case a.DoesAny(MonsLightFootsteps):
		return ColorOrange
	case a.DoesAny(MonsCreep):
		return ColorGreen
	case a.DoesAny(MonsWingFlap):
		return ColorCyan
	case a.DoesAny(MonsHeavyFootsteps):
		return ColorMagenta
	default:
		return ColorRed
	}
}

// MakeNoise produces noise of a given intensity at the given position.
func (g *Game) MakeNoise(at gruid.Point, nt NoiseType) {
	noise := nt.Noise()
	dij := &MapPath{passable: g.Map.Passable}
	maxdist := (3 * noise) / 2
	g.PRnoise.BreadthFirstMap(dij, []gruid.Point{at}, maxdist)
	pp := g.PP()
	pa := g.PlayerActor()
	goodhearing := pa.DoesAny(GoodHearing) && !pa.DoesAny(BadHearing)
	d := g.PRnoise.BreadthFirstMapAt(pp)
	if d <= maxdist && (goodhearing || d <= noise) {
		if at != pp && g.Map.Noise != nil {
			g.Map.Noise[at] = nt
		}
		if nt == NoiseExplosion {
			g.StoryLogf("Explosion (distance: %d)", paths.DistanceManhattan(at, pp))
		}
		g.LogStyled(nt.Msg(), logNotable)
	}
	for i, ai := range g.Monsters() {
		d := g.PRnoise.BreadthFirstMapAt(g.Entity(i).P)
		if d > maxdist || d > noise && !ai.DoesAny(GoodHearing) {
			continue
		}
		g.MonsterUpdateTarget(i, ai, at)
	}
}

// NoiseType represents various kinds of non-movement noise sources.
type NoiseType int

const (
	NoiseBark NoiseType = iota
	NoiseCombat
	NoiseDig
	NoiseExplosion
	NoiseEarthMenhir
	NoiseFakePortal
	NoiseLightning
	NoiseMusic
	NoiseRoar
	NoiseWind
)

// Noise returns the amount of noise associated with the given type of noise.
func (nt NoiseType) Noise() int {
	switch nt {
	case NoiseCombat:
		return 4
	case NoiseExplosion, NoiseFakePortal:
		return MaxFOVRange + 4
	case NoiseMusic, NoiseEarthMenhir:
		return 2 * MaxFOVRange
	default:
		return MaxFOVRange
	}
}

// Msg returns the onomatopeya corresponding to the given  type of noise.
func (nt NoiseType) Msg() string {
	switch nt {
	case NoiseBark:
		return "WOOF!"
	case NoiseCombat:
		// Lowercase for combat, because noise is not as strong.
		return "Smash!"
	case NoiseDig:
		return "CRACK!"
	case NoiseExplosion:
		return "POP-BOOM!"
	case NoiseFakePortal:
		return "THRUM!"
	case NoiseLightning:
		return "PANG!"
	case NoiseEarthMenhir:
		return "RING-RING!"
	case NoiseMusic:
		return "Noisy Imp belts “♫ larilon, larila ♫ ♪”."
	case NoiseRoar:
		return "ROAR!"
	case NoiseWind:
		return "WHIZ!"
	default:
		// should not happen.
		return "NOISE!"
	}
}

// SmellFood performs item sensing of the GoodSmell trait.
func (g *Game) SmellFood() {
	pp := g.PP()
	for i := range g.Comestibles() {
		ei := g.Entity(i)
		if ei.KnownP == ei.P {
			continue
		}
		if paths.DistanceManhattan(ei.P, pp) <= 2*MaxFOVRange {
			// NOTE: using path-distance would leak information
			// about walls (if the player goes through the tedious
			// task of doing some math).
			g.SenseEntity(ei, "smell")
		}
	}
}
