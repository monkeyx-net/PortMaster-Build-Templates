package main

import (
	"codeberg.org/anaseto/gruid"
)

// Event represents the various kinds events.
type Event interface {
	Handle(*Game)
}

// PushEvent pushes an event with the given priority (corresponding to a game
// turn).
func (g *Game) PushEvent(ev Event, r int) {
	g.Events.Push(ev, r)
}

// PushEventD pushes an event with the given delay (relative to current turn).
func (g *Game) PushEventD(ev Event, delay int) {
	g.Events.Push(ev, g.Turn+delay)
}

// PushEventFirst pushes an event with the given priority (corresponding to a
// game turn), ensuring it's ranked first among events with same priority.
func (g *Game) PushEventFirst(ev Event, r int) {
	g.Events.PushFirst(ev, r)
}

// EndTurnEvent represents an event corresponding to the end of the current
// turn.
type EndTurnEvent struct{}

// AbyssFallEvent represents an abyss fall event for the player.
type AbyssFallEvent struct{}

func (ev AbyssFallEvent) Handle(g *Game) {
	g.FallAbyss(DescendFall)
}

// StatusEvent represents a status progression event.
type StatusEvent struct {
	Status status
}

// StatusEndMsg contains end-of-status messages for the various statuses.
var StatusEndMsgs = [...]string{
	StatusExhausted:     "You no longer feel exhausted.",
	StatusSwift:         "You no longer feel speedy.",
	StatusLignification: "You no longer feel attached to the ground.",
	StatusConfusion:     "You no longer feel confused.",
	StatusDig:           "You no longer feel like an earth dragon.",
	StatusLevitation:    "You no longer levitate.",
	StatusShadows:       "You are no longer surrounded by shadows.",
	StatusIlluminated:   "You are no longer illuminated.",
	StatusTransparent:   "You are no longer transparent.",
	StatusDisguised:     "You are no longer disguised.",
	StatusDispersal:     "You are no longer unstable.",
}

func (sev *StatusEvent) Handle(g *Game) {
	st := sev.Status
	g.Player.Statuses[st] -= DurationStatusStep
	if g.Player.Statuses[st] <= 0 {
		g.Player.Statuses[st] = 0
		g.LogStyled(StatusEndMsgs[st], logStatusEnd)
		g.md.StatusEndAnimation()
		switch st {
		case StatusLevitation:
			if g.Map.At(g.Player.P) == ChasmCell {
				g.PushEventFirst(AbyssFallEvent{}, g.Turn)
			}
		case StatusLignification:
			g.Player.HPbonus -= LignificationHPbonus
			if g.Player.HPbonus < 0 {
				g.Player.HPbonus = 0
			}
		}
	} else {
		g.PushEventD(sev, DurationStatusStep)
	}
}

// MonsterTurnEvent represents an event corresponding to a monster's turn.
type MonsterTurnEvent struct {
	Index int // monster index
}

func (mev *MonsterTurnEvent) Handle(g *Game) {
	mons := g.Monsters[mev.Index]
	if mons.Exists() {
		mons.HandleTurn(g)
	}
}

// MonsterStatusEvent represents a status progression event for a given
// monster.
type MonsterStatusEvent struct {
	Index  int           // monster index
	Status monsterStatus // status effect
}

func (mev *MonsterStatusEvent) Handle(g *Game) {
	mons := g.Monsters[mev.Index]
	st := mev.Status
	mons.Statuses[st] -= DurationStatusStep
	if mons.Statuses[st] <= 0 {
		mons.Statuses[st] = 0
		if g.Player.Sees(mons.P) {
			g.Logf("%s is no longer %s.", mons.Kind.Definite(true), st)
		}
		switch st {
		case MonsConfused, MonsLignified:
			mons.Path = mons.APath(g, mons.P, mons.Target)
		}
	} else {
		g.PushEventD(&MonsterStatusEvent{Index: mev.Index, Status: st}, DurationStatusStep)
	}
}

// PosEvent represents a positional event.
type PosEvent struct {
	P      gruid.Point // event's position
	Action PosAction   // event's action
	Timer  int         // number of turns until activation
}

// PosAction represents various kinds of actions for positional events.
type PosAction int

const (
	CloudEnd PosAction = iota
	ObstructionEnd
	ObstructionProgression
	FireProgression
	PurpleMistProgression
	FogProgression
	Earthquake
	DelayedHarmonicNoiseEvent
	DelayedOricExplosionEvent
)

func (cev *PosEvent) Handle(g *Game) {
	switch cev.Action {
	case CloudEnd:
		delete(g.Map.Clouds, cev.P)
		g.ComputeFOV()
	case ObstructionEnd:
		if t, ok := g.Map.MagicalBarriers[cev.P]; ok {
			g.Map.Set(cev.P, t)
		}
	case ObstructionProgression:
		p := g.FreePassableCell()
		g.MagicalBarrierAt(p)
		if g.Player.Sees(p) {
			g.Logf("You see an oric barrier appear out of thin air.")
			g.StopAuto()
		}
		g.PushEvent(&PosEvent{Action: ObstructionProgression},
			g.Turn+DurationObstructionProgression+RandInt(DurationObstructionProgression/4))
	case FireProgression:
		if _, ok := g.Map.Clouds[cev.P]; !ok {
			break
		}
		for _, p := range g.PlayerPassableNeighbors(cev.P) {
			if RandInt(10) == 0 {
				continue
			}
			g.Burn(p)
		}
		delete(g.Map.Clouds, cev.P)
		g.PurpleMist(cev.P, 1)
		g.ComputeFOV()
	case PurpleMistProgression:
		if _, ok := g.Map.Clouds[cev.P]; !ok {
			break
		}
		if cev.Timer <= 0 {
			delete(g.Map.Clouds, cev.P)
			g.ComputeFOV()
			break
		}
		g.applyPurpleMistEffectsAt(cev.P)
		cev.Timer--
		g.PushEventD(cev, DurationTurn)
	case FogProgression:
		p := g.FreePassableCell()
		g.Fog(p, 1)
		g.PushEvent(&PosEvent{Action: FogProgression},
			g.Turn+DurationPurpleMistProgression+RandInt(DurationPurpleMistProgression/4))
	case Earthquake:
		g.LogStyled("The earth suddenly shakes with force!", logSpecial)
		g.StoryPrint("Special event: earthquake!")
		g.LogStyled("CREAK-CRACK!", logNotable)
		it := g.Map.Terrain.Iterator()
		for it.Next() {
			p := it.P()
			t := it.Cell()
			if !IsDiggable(t) || !g.Map.HasFreeNeighbor(p) {
				continue
			}
			if Distance(cev.P, p) > RandInt(35) || RandInt(2) == 0 {
				continue
			}
			g.Map.Set(p, RubbleCell)
			g.Fog(p, 1)
		}
		g.MakeNoise(EarthquakeNoise, cev.P)
		g.Map.Noise[cev.P] = 3
	case DelayedHarmonicNoiseEvent:
		if cev.Timer <= 1 {
			g.Player.Statuses[StatusDelay] = 0
			g.LogStyled("RING-RING!", logNotable)
			g.Map.Noise[cev.P] = 3
			g.MakeNoise(DelayedHarmonicNoise, cev.P)
		} else {
			cev.Timer--
			g.Player.Statuses[StatusDelay] = cev.Timer
			g.PushEventD(cev, DurationTurn)
		}
	case DelayedOricExplosionEvent:
		if cev.Timer <= 1 {
			g.OricExplosion(cev.P)
		} else {
			cev.Timer--
			g.Player.Statuses[StatusDelay] = cev.Timer
			g.PushEventD(cev, DurationTurn)
		}
	}
}

// OricExplosion produces an oric explosion centered around the given position.
func (g *Game) OricExplosion(at gruid.Point) {
	g.Player.Statuses[StatusDelay] = 0
	g.LogStyled(g.CrackSound(), logNotable)
	mp := &OricExplosionPath{m: g.Map}
	nodes := g.PR.DijkstraMap(mp, []gruid.Point{at}, 9)
	ps := []gruid.Point{} // destroyed diggable tiles
	for _, n := range nodes {
		t := g.Map.At(n.P)
		if !IsDiggable(t) {
			continue
		}
		g.Map.Set(n.P, RubbleCell)
		g.NewDigStats()
		ps = append(ps, n.P)
	}
	g.md.OricExplosionAnimation(ps)
	for _, p := range ps {
		if IsKnown(g.Map.KnownTerrain.At(p)) {
			g.Map.KnownTerrain.Set(p, RubbleCell)
		}
		g.Fog(p, 1)
	}
	g.MakeNoise(OricExplosionNoise, at)
	g.Map.Noise[at] = 3
	g.ComputeFOV()
}

// PurpleMist produces purple mist around the given position.
func (g *Game) PurpleMist(at gruid.Point, radius int) {
	mp := &NoisePath{g: g}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{at}, radius)
	for _, n := range nodes {
		_, ok := g.Map.Clouds[n.P]
		if !ok {
			g.Map.Clouds[n.P] = CloudPurpleMist
			g.PushEventD(&PosEvent{Action: PurpleMistProgression,
				P: n.P, Timer: DurationPurpleMist}, DurationCloudProgression)

			g.applyPurpleMistEffectsAt(n.P)
		}
	}
	g.ComputeFOV()
}

func (g *Game) applyPurpleMistEffectsAt(p gruid.Point) {
	if p == g.Player.P {
		if g.PutStatus(StatusConfusion, DurationConfusionPlayer) {
			g.Log("The purple mist confuses you.")
		}
		return
	}
	mons := g.MonsterAt(p)
	if !mons.Exists() || (RandInt(2) == 0 && mons.Status(MonsExhausted)) {
		// do not always make already exhausted monsters sleep (they were probably awaken)
		return
	}
	mons.PutConfusion(g)
	if mons.State != Resting && g.Player.Sees(mons.P) {
		g.Logf("%s falls asleep.", mons.Kind.Definite(true))
	}
	mons.State = Resting
	mons.Dir = ZP
	mons.ExhaustTime(g, 4+RandInt(2))
}

// Burn starts a magical fire at the given position.
func (g *Game) Burn(p gruid.Point) {
	if _, ok := g.Map.Clouds[p]; ok {
		return
	}
	t := g.Map.At(p)
	if !Flammable(t) {
		return
	}
	g.Stats.Burns++
	switch t {
	case DoorCell:
		g.Log("The door vanishes in magical flames.")
	case TableCell:
		g.Log("The table vanishes in magical flames.")
	case BarrelCell:
		g.Log("The barrel vanishes in magical flames.")
		delete(g.Objects.Barrels, p)
	case TreeCell:
		g.Log("The tree vanishes in magical flames.")
	}
	g.Map.Set(p, GroundCell)
	g.Map.Clouds[p] = CloudFire
	if g.Player.Sees(p) {
		g.ComputeFOV()
	}
	g.PushEventD(&PosEvent{P: p, Action: FireProgression}, DurationCloudProgression)
}

// The following constants represent various kinds of durations.
const (
	DurationCloudProgression       = 1
	DurationConfusionMonster       = 12
	DurationConfusionPlayer        = 5
	DurationDazeMonster            = 6
	DurationDigging                = 8
	DurationDisguise               = 9
	DurationDispersal              = 12
	DurationExhaustion             = 5
	DurationExhaustionMonster      = 10
	DurationFog                    = 15
	DurationFogSmokingCloak        = 2
	DurationHarmonicNoiseDelay     = 13
	DurationIlluminated            = 7
	DurationLevitation             = 18
	DurationLignificationMonster   = 12
	DurationLignificationPlayer    = 4
	DurationMagicalBarrier         = 15
	DurationObstructionProgression = 15
	DurationOricExplosionDelay     = 8
	DurationPurpleMist             = 15
	DurationPurpleMistProgression  = 12
	DurationSatiationMonster       = 40
	DurationShadows                = 15
	DurationStatusStep             = 1
	DurationSwiftness              = 5
	DurationSwiftnessAmulet        = 3
	DurationTransparency           = 15
	DurationTurn                   = 1
)
