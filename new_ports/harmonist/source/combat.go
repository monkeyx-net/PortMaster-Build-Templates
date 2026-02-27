// combat utility functions

package main

import (
	"errors"
	"fmt"

	"codeberg.org/anaseto/gruid"
)

// InflictDamage makes the monster inflict 1 damage on the player.
func (m *Monster) InflictDamage(g *Game) {
	oldHP := g.Player.HP
	g.damagePlayer()
	g.md.WoundedAnimation()
	if oldHP > 1 && g.Player.HP <= 1 {
		g.StoryPrintf("Critical hit by %s (HP: %d)", m.Kind, g.Player.HP)
		g.md.WoundedAnimation() // twice
		g.md.criticalHPWarning()
	} else if g.Player.HP > 0 {
		g.StoryPrintf("Hit by %s (HP: %d)", m.Kind, g.Player.HP)
	} else {
		g.StoryPrintf("Killed by %s", m.Kind)
	}
	if g.Player.HP > 0 && g.Player.Inventory.Body.Kind == CloakConversion && g.Player.MP < g.Player.MPMax() {
		g.Player.MP++
	}
}

// damagePlayer handles 1 damage for the player.
func (g *Game) damagePlayer() {
	g.Stats.Damage += 1
	g.Stats.DDamage[g.Map.Depth] += 1
	g.Player.HPbonus -= 1
	if g.Player.HPbonus < 0 {
		g.Player.HP += g.Player.HPbonus
		g.Player.HPbonus = 0
	}
}

func (md *model) criticalHPWarning() {
	if md.mode == modeNormal {
		md.mode = modeHPCritical
	}
	md.critical = true
}

// MakeNoise makes a given amount of noise at the given position.
func (g *Game) MakeNoise(noise int, at gruid.Point) {
	mp := &NoisePath{g: g}
	g.PRnoise.BreadthFirstMap(mp, []gruid.Point{at}, noise)
	for _, m := range g.Monsters {
		if !m.Exists() {
			continue
		}
		if m.State == Hunting {
			continue
		}
		d := g.PRnoise.BreadthFirstMapAt(m.P)
		if d > noise {
			continue
		}
		if m.State == Resting && 3*d > 2*noise || m.Status(MonsExhausted) && m.State == Resting && 3*d > noise {
			continue
		}
		m.MakeWanderAt(at)
		m.GatherBand(g)
	}
}

// LeaveRoomForPlayer moves a peaceful monster to make room for the player.
func (m *Monster) LeaveRoomForPlayer(g *Game) gruid.Point {
	dij := &MonPath{g: g, monster: m}
	nodes := g.PR.DijkstraMap(dij, []gruid.Point{m.P}, 10)
	free := invalidPos
	dist := unreachable
	for _, n := range nodes {
		if !m.CanPass(g, n.P) {
			continue
		}
		if n.P == g.Player.P || n.P == m.P {
			continue
		}
		mons := g.MonsterAt(n.P)
		if mons.Exists() {
			continue
		}
		if Distance(n.P, m.P) < dist {
			free = n.P
			dist = Distance(n.P, m.P)
		}
	}
	// free should be valid except in really rare cases.
	return free
}

// FindJumpTarget finds a jump target for cases that need jumping with turning
// (or else the player would be blocked). Thematically, we can think that the
// gawalt propulses itself using the monsters (so several small jumps in a
// row).
func (g *Game) FindJumpTarget(p gruid.Point) gruid.Point {
	mp := &JumpPath{g: g}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{p}, 10)
	free := invalidPos
	dist := unreachable
	for _, n := range nodes {
		if !g.PlayerCanJumpTo(n.P) {
			continue
		}
		if n.P == g.Player.P || n.P == p {
			continue
		}
		mons := g.MonsterAt(n.P)
		if mons.Exists() {
			continue
		}
		if Distance(n.P, p) < dist {
			free = n.P
			dist = Distance(n.P, p)
		}
	}
	if !g.PlayerCanPass(p) {
		// Very rarely, if there is no passable position nearby, we
		// just choose a teleport one, which makes no sense, but is a
		// simple way to handle an extremely rare edge-case.
		free = g.TeleportPos()
	}
	// free should be valid except in really rare cases
	return free
}

// Jump handles jumping over a monster.
func (g *Game) Jump(mons *Monster) (bool, error) {
	if mons.Peaceful(g) && mons.Kind != MonsEarthDragon {
		op := mons.P
		if g.Map.At(op) == ChasmCell && !g.Player.HasStatus(StatusLevitation) {
			return true, g.AbyssJump()
		}
		if !mons.CanPass(g, g.Player.P) {
			p := mons.LeaveRoomForPlayer(g)
			if p != invalidPos {
				if mons.MoveTo(g, p) {
					mons.Swapped = true
					g.PlacePlayerAt(op)
					return false, nil
				}
			}
			// otherwise (which should not happen in practice), swap anyways
		}
		g.PlacePlayerAt(mons.P)
		return false, nil
	}
	if g.Player.HasStatus(StatusExhausted) {
		return false, errors.New("You cannot jump while exhausted.")
	}
	dir := DirNorm(g.Player.P, mons.P)
	p := g.Player.P
	for {
		p = p.Add(dir)
		if !g.PlayerCanPass(p) {
			break
		}
		m := g.MonsterAt(p)
		if !m.Exists() {
			break
		}
	}
	if !g.PlayerCanJumpTo(p) {
		p = g.FindJumpTarget(mons.P)
		if !InMap(p) {
			// should not happen in practice, but better safe than sorry
			return false, errors.New("You cannot jump there.")
		}
	}
	if !g.Player.HasStatus(StatusSwift) && g.Player.Inventory.Body.Kind != CloakAcrobat {
		g.PutStatus(StatusExhausted, DurationExhaustion)
	}
	if mons.Kind == MonsEarthDragon {
		g.PutConfusion()
	}
	g.PlacePlayerAt(p)
	g.Stats.Jumps++
	g.Logf("You jump over %s", mons.Kind.Definite(false))
	g.StoryPrintf("Jumped over %s", mons.Kind)
	if g.Stats.Jumps+g.Stats.WallJumps == 15 {
		AchAcrobat.Get(g)
	}
	return false, nil
}

// WallJump performs a jump by propulsing off a wall.
func (g *Game) WallJump(p gruid.Point) (bool, error) {
	from := g.Player.P
	t := g.Map.At(from)
	if IsEnclosing(t) {
		return false, fmt.Errorf("You cannot jump from %s.", g.TerrainShortDesc(t, from))
	}
	if g.Player.HasStatus(StatusExhausted) {
		return false, errors.New("You cannot jump while exhausted.")
	}
	dir := DirNorm(p, from)
	jp, free, reach := from, from, from // last jump-passable/free/reachable tile
	for range 3 {
		reach = reach.Add(dir)
		if !g.PlayerCanJumpPass(reach) {
			break
		}
		jp = reach
		if m := g.MonsterAt(jp); m.Exists() || !g.PlayerCanPass(jp) {
			continue
		}
		free = jp
	}
	abyssJump := func() (bool, error) {
		if g.Map.At(jp) != ChasmCell || g.Player.HasStatus(StatusLevitation) {
			return false, nil
		}
		if g.Player.Sees(jp) {
			return true, g.AbyssJump()
		}
		// Fall into (unseen) the abyss.
		g.PushEventFirst(AbyssFallEvent{}, g.Turn)
		return true, nil
	}
	switch {
	case jp == from:
		return false, errors.New("There’s not enough room for a wall-jump.")
	case free == from:
		if m := g.MonsterAt(from.Add(dir)); m.Exists() {
			// If there's no free tile to jump and a monster is
			// next to you, perform a classic over-monster jump.
			return g.Jump(m)
		}
		if ok, err := abyssJump(); ok {
			return ok, err
		}
		// Very rarely, just freely find a jump target, nearby if
		// possible, like when jumping over a monster.
		free = g.FindJumpTarget(from.Add(dir))
	case Distance(free, from) == 1 && g.Player.Sees(reach):
		if ok, err := abyssJump(); ok {
			return ok, err
		}
		return false, errors.New("Wall-jumping from there is useless.")
	}
	if !g.Player.HasStatus(StatusSwift) && g.Player.Inventory.Body.Kind != CloakAcrobat {
		g.PutStatus(StatusExhausted, DurationExhaustion)
	}
	g.Log("You jump by propelling yourself against the wall.")
	g.PlacePlayerAt(free)
	g.md.MoveAnimation(from, free)
	g.Stats.WallJumps++
	if g.Stats.Jumps+g.Stats.WallJumps == 15 {
		AchAcrobat.Get(g)
	}
	return false, nil
}

// HitNoise returns the amount of noise corresponding to a hit.
func (g *Game) HitNoise(clang bool) int {
	noise := BaseHitNoise
	if clang {
		noise += 5
	}
	return noise
}

// HandleKill handles a monster death (like when falling into the abyss or
// drowning).
func (g *Game) HandleKill(mons *Monster) {
	g.Stats.Killed++
	g.Stats.KilledMons[mons.Kind]++
	if g.Player.Sees(mons.P) {
		AchMurderer.Get(g)
	}
	if g.Map.At(mons.P) == DoorCell {
		g.ComputeFOV()
	}
	g.StoryPrintf("Death of %s", mons.Kind.Indefinite(false))
}

// The following constants represents various kinds of noise amounts.
const (
	WallNoise              = 10
	ExplosionNoise         = 14
	BarkNoise              = 10
	MagicCastNoise         = 5
	HarmonicNoise          = 7
	BaseHitNoise           = 2
	QueenStoneNoise        = 10
	SingingNoise           = 12
	EarthquakeNoise        = 35
	QueenRockFootstepNoise = 7
	DelayedHarmonicNoise   = 25
	OricExplosionNoise     = 20
)

// ClangMsg returns text for a noisy hit.
func (g *Game) ClangMsg() (sclang string) {
	return "Smash!"
}
