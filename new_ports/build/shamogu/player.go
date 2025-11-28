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
	dirChange := dir != g.Dir
	to := pl.P.Add(dir)
	tto := g.Map.Terrain.At(to)
	passto := Passable(tto)
	if pa.CanMove() {
		switch {
		case pa.Has(StatusShadow) && tto == TranslucentWall:
			return g.WallThrough(dir, to)
		case pa.DoesAny(Gawalt) && !passto && !pa.Has(StatusDig):
			return g.WallJump(dir)
		case pa.Has(StatusSprint):
			return g.Sprint(dir)
		}
	}
	if pa.DoesAny(PatternCrocodile) && passto && dir == g.Dir.Mul(-1) {
		// Crocodile needs to spend a turn turning before moving
		// backwards.
		g.Dir = dir
		return true
	}
	if pa.DoesAny(Elephanty) && passto && dir != g.Dir && !g.Map.Passable(pl.P.Add(g.Dir)) {
		// Elephant players need to spend a turn turning when facing a
		// wall.
		// NOTE: Not sure if it makes sense thematically to allow
		// turning toward a fearsome foe (same for croco when reversing
		// directions), but it's helpful to not always lose a turn
		// turning once you become cornered and berserk.
		g.Dir = dir
		return true
	}
	j, aj := g.RangedTargetInDir(pl.P, dir)
	if j >= 0 {
		if g.PlayerFears(j, aj) {
			g.Log("You’re afraid to go in that direction.")
			return false
		}
		ej := g.Entity(j)
		dist := paths.DistanceManhattan(pl.P, ej.P)
		if dist == 1 {
			switch {
			case pa.DoesAny(PatternCrocodile):
				if dirChange {
					g.Dir = dir
					break
				}
				g.BumpAttackActor(PlayerID, j, pa, aj, AttackDrag, dist, 0)
			case pa.DoesAny(PatternFourDirs) && !pa.Has(StatusFocus):
				g.Dir = dir
				g.FourDirectionalAttack(PlayerID, pa, AttackPlain, dist)
			default:
				g.Dir = dir
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
		case pa.DoesAny(PatternRampage|PatternBat) && pa.CanMove() && (pa.HP > 1 || !pa.Has(StatusPoison)):
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
		if pa.DoesAny(PatternCrocodile) && dirChange && passto {
			g.Dir = dir
			return true
		}
		g.Log("You cannot move while lignified.")
		return false
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		if pa.DoesAny(PatternCrocodile) && dirChange && passto {
			g.Dir = dir
			return true
		}
		g.Log("You cannot move while poisoned and almost dead.")
		return false
	}
	if !passto {
		if !inMap(to) {
			g.Log("You cannot walk out of the map.")
			return false
		}
		switch {
		case pa.Has(StatusDig):
			g.digAt(to)
		default:
			g.Logf("You cannot walk into the %s.", TerrainName(tto))
			return false
		}
	}
	g.Dir = dir
	g.BumpMoveActor(PlayerID, pa, to)
	// Try charge.
	if pa.DoesAny(PatternCrocodile) && dirChange {
		return true
	}
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

// PlayerFears reports whether the players fears the given actor.
func (g *Game) PlayerFears(j ID, aj *Actor) bool {
	pa := g.PlayerActor()
	return pa.Has(StatusFear) || pa.DoesAny(Elephanty) && aj.IsMonster(HungryRat) && !pa.Has(StatusBerserk)
}

func (g *Game) logDaze() {
	g.Log("You’re dazed: you need to wait (ENTER) or eat.")
}

// Sprint performs a sprinting bump-movement in the given direction.
func (g *Game) Sprint(dir gruid.Point) bool {
	pl, pa := g.Player()
	if j, aj := g.RangedTargetInDir(pl.P, dir); j >= 0 {
		if g.PlayerFears(j, aj) {
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

// WallThrough performs a move through a translucent wall.
func (g *Game) WallThrough(dir, at gruid.Point) bool {
	to := at.Add(dir)
	if !g.IsFree(to) {
		g.Log("There is no room behind the translucent wall.")
		return false
	}
	from := g.PP()
	g.BumpMoveActor(PlayerID, g.PlayerActor(), to)
	g.md.MoveAnimation(from, to)
	return true
}

// WallJump performs a jump by propulsing against a wall.
func (g *Game) WallJump(dir gruid.Point) bool {
	pl, pa := g.Player()
	at := pl.P.Add(dir)
	if !inMap(at) {
		g.Log("You cannot walk out of the map.")
		return false
	}
	// if dir == g.Dir || dir == g.Dir.Mul(-1) {
	// 	g.Logf("You cannot walk into the %s.", TerrainName(g.Map.Terrain.At(at)))
	// 	return false
	// }
	over := pl.P.Sub(dir)
	to := over.Sub(dir)
	if !g.Map.Passable(over) {
		g.Log("There’s no room for a wall jump.")
		return false
	}
	if !g.InFOV(to) {
		g.Log("You cannot jump out of view.")
		return false
	}
	if !g.IsFree(to) {
		g.Log("There’s no room for a wall jump.")
		return false
	}
	g.Dir = dir
	from := pl.P
	g.BumpMoveActor(PlayerID, pa, to)
	g.md.MoveAnimation(from, to)
	if i, ai := g.ActorAt(over); i >= 0 {
		if pa.Has(StatusSprint) {
			g.PutStatus(i, ai, StatusImbalance, DurationImbalanceSprint)
		} else {
			g.BumpAttackActor(PlayerID, i, pa, ai, AttackPlain, 1, 0)
		}
	}
	if pa.Has(StatusImbalance) && g.IntN(10) == 0 {
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
	p.HP = min(p.HP, p.GetMaxHP())
}
