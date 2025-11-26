package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// PlayerBump performs a player bump action in the given direction. It returns
// true if the turn was consumed.
func (g *Game) PlayerBump(dir gruid.Point) bool {
	pl := g.PlayerEntity()
	pa := pl.Actor()
	if pa.Has(StatusDaze) {
		g.logDaze()
		return false
	}
	if pa.Has(StatusSprint) && !pa.Has(StatusLignification) {
		return g.Sprint(dir)
	}
	j, aj := g.RangedTargetInDir(pl.P, dir)
	if j >= 0 {
		if pa.Has(StatusFear) {
			g.Log("You’re afraid to go in that direction.")
			return false
		}
		ej := g.Entity(j)
		dist := paths.DistanceManhattan(pl.P, ej.P)
		if dist == 1 {
			g.Dir = dir
			if pa.DoesAny(PatternFourDirs) && !pa.Has(StatusFocus) {
				g.FourDirectionalAttack(PlayerID, pa, AttackPlain, dist)
			} else {
				g.BumpAttackActor(PlayerID, j, pa, aj, AttackPlain, dist, 0)
			}
			return true
		}
		switch {
		case pa.DoesAny(PatternRanged|PatternRangedRecoil) ||
			pa.DoesAny(PatternCatch) && !aj.ResistsMove():
			g.Dir = dir
			g.md.RangedAttackAnimation(pa, pl.P, ej.P)
			g.BumpAttackActor(PlayerID, j, pa, aj, AttackRanged, dist, 0)
			return true
		case pa.DoesAny(PatternSwap|PatternSwapDaze) && !pa.Has(StatusLignification) && !aj.ResistsMove():
			g.Dir = dir
			g.md.RangedAttackAnimation(pa, pl.P, ej.P)
			g.BumpAttackActor(PlayerID, j, pa, aj, AttackRanged, dist, 0)
			return true
		case pa.DoesAny(PatternRampage) && pa.CanMove() && (pa.HP > 1 || !pa.Has(StatusPoison)):
			g.Dir = dir
			to := ej.P.Sub(dir)
			from := pl.P
			g.BumpMoveActor(PlayerID, pa, to)
			g.md.MoveAnimation(from, to)
			g.BumpAttackActor(PlayerID, j, pa, aj, AttackCharge, dist, 0)
			return true
		}
	}
	// We're not ranged and next position is not a monster: move.
	if !pa.CanMove() {
		g.Log("You cannot move while lignified.")
		return false
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("You cannot move while poisoned and almost dead.")
		return false
	}
	to := pl.P.Add(dir)
	if !g.Map.Passable(to) {
		if !inMap(to) {
			g.Log("You cannot walk out of the map.")
			return false
		}
		switch {
		case pa.Has(StatusDig):
			g.dig(to)
		default:
			g.Logf("You cannot walk into the %s.", TerrainName(g.Map.Terrain.At(to)))
			return false
		}
	}
	g.Dir = dir
	g.BumpMoveActor(PlayerID, pa, to)
	// Try charge.
	nto := to.Add(dir)
	if pa.DoesAny(PatternFourDirs) && !pa.Has(StatusFocus) {
		g.FourDirectionalAttack(PlayerID, pa, AttackCharge, 2)
		return true
	}
	if j, aj := g.ActorAt(nto); j >= 0 {
		// Charge against actor at nto. An ambushing charge outside FOV
		// can happen and is allowed even while afraid.
		g.BumpAttackActor(PlayerID, j, pa, aj, AttackCharge, 2, 0)
	}
	return true
}

func (g *Game) logDaze() {
	g.Log("You’re dazed: you need to wait (ENTER) or eat.")
}

// Sprint performs a sprinting bump-movement in the given direction.
func (g *Game) Sprint(dir gruid.Point) bool {
	pl, pa := g.Player()
	if j, _ := g.RangedTargetInDir(pl.P, dir); j >= 0 {
		if pa.Has(StatusFear) {
			g.Log("You’re afraid to go in that direction.")
			return false
		}
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("You cannot sprint while poisoned and almost dead.")
		return false
	}
	n := 2 // length of step
	switch {
	case dir == g.Dir:
		// Still sprinting in the same dir: advance up to 3 tiles.
		n = 3
	case dir == g.Dir.Mul(-1):
		// Turning backwards while sprinting moves one tile only.
		n = 1
	}
	jump := false // whether there's a monster we're going to jump over
	from := pl.P
	p, to := from, from
	for range n {
		// We check the next n tiles in FOV, to know whether there's
		// room for sprinting. We update destination as far as possible
		// before a wall and without a monster.
		p = p.Add(dir)
		if !g.InFOV(p) {
			break
		}
		if !g.Map.Passable(p) && !g.Dig(PlayerID, pa, p) {
			break
		}
		if i, _ := g.ActorAt(p); i < 0 {
			to = p
		} else {
			jump = true
		}
	}
	if from == to {
		g.Log("You cannot sprint in that direction.")
		return false
	}
	g.Dir = dir
	g.BumpMoveActor(PlayerID, pa, to)
	if jump {
		// Path contains a monster or two: jumping over them unbalances them.
		for p := from; p != to; p = p.Add(dir) {
			if i, ai := g.ActorAt(p); i >= 0 {
				g.PutStatus(i, ai, StatusImbalance, DurationImbalanceSprint)
			}
		}
	}
	g.md.MoveAnimation(from, to)
	if n > 1 && pa.Has(StatusImbalance) && g.IntN(DurationSprint) == 0 {
		g.Log("You fall due to imbalance.")
		g.PutStatus(PlayerID, pa, StatusDaze, DurationDazeFall)
	}
	return true
}

// ComputePlayerStats recomputes the player attack, defense, and trait
// statistics. It should be called after equipping or improving spirits.
func (g *Game) ComputePlayerStats() {
	p := g.PlayerActor()
	p.Attack = 2
	p.Defense = 1
	p.MaxHP = 9
	p.Traits = Player
	for _, sp := range g.PlayerSpirits() {
		p.Attack += sp.BonusAttack[sp.Level]
		p.Defense += sp.BonusDefense[sp.Level]
		p.MaxHP += sp.BonusHP[sp.Level]
		p.Traits |= sp.BonusTraits[sp.Level]
	}
}
