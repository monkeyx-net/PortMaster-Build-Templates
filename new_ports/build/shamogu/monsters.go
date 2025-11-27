package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// HandleMonsterTurn handles AI for a monster actor with given ID. The ID
// should then correspond to a non-player actor.
func (g *Game) HandleMonsterTurn(i ID, ai *Actor) {
	beh := ai.Behavior
	if beh.SkipTurn {
		beh.SkipTurn = false
		return
	}
	if ai.Has(StatusDaze) {
		return
	}
	ei := g.Entity(i)
	if ai.Has(StatusConfusion) && g.monsterDiscord(i, ai) {
		return
	}
	pl, pa := g.Player()
	if g.InFOV(ei.P) && g.MonsterUpdateTarget(i, ai, pl.P, false) {
		// A wandering monster spends one turn noticing the player,
		// unless you have the Dazzling Zebra.
		if !pa.DoesAny(Dazzling) {
			return
		}
	}
	if pa.DoesAny(Dazzling) {
		// A dazzling player is easily noticeable by monsters at the
		// end of their turn.
		defer func() { _ = g.InFOV(ei.P) && g.MonsterUpdateTarget(i, ai, pl.P, false) }()
	}
	if beh.Target == ei.P || beh.Target == InvalidPos {
		// The monster's target has either been reached or is invalid:
		// we spend a turn chosing a new target.
		if ai.DoesAny(MonsHungry) && g.HungryHunt(ei.P) || g.Marked() && beh.Guard == InvalidPos {
			beh.Target = pl.P
			beh.State = Hunting
			return
		}
		from, maxdist := ei.P, unreachable
		if beh.Guard != InvalidPos {
			maxdist = MaxFOVRange - 1
			from = beh.Guard
		} else if beh.State == Hunting {
			maxdist = MaxFOVRange + 5
		}
		if maxdist < unreachable {
			beh.Target = g.RandomPassableWithin(from, maxdist)
		} else {
			beh.Target = g.RandomPassableNearbyTarget(from)
		}
		beh.State = Wandering
		if maxdist > MaxFOVRange && g.IntN(3) == 0 || maxdist <= MaxFOVRange && g.IntN(10) == 0 {
			g.CallToCommonTarget(i, ai)
		}
		return
	}
	if pa.Has(StatusDisorient) && g.monsterBumpDisoriented(i, ai) {
		return
	}
	if g.monsterFlee(i, ai) {
		return
	}
	g.monsterBumpNext(i, ai)
}

// monsterDiscord makes a confused monster possibly attack a random adjacent
// monster, instead of their normal action. It reports whether such an attack
// was performed.
func (g *Game) monsterDiscord(i ID, ai *Actor) bool {
	if ai.Has(StatusFear) {
		return false
	}
	ei := g.Entity(i)
	var ids []ID
	var actors []*Actor
	for p := range Neighbors(ei.P) {
		if j, aj := g.ActorAt(p); j >= 0 {
			ids = append(ids, j)
			actors = append(actors, aj)
		}
	}
	if len(ids) == 0 {
		return false
	}
	k := g.IntN(len(ids))
	g.BumpAttackActor(i, ids[k], ai, actors[k], AttackPlain, 1, 0)
	return true
}

// monsterBumpDisoriented handles movement for a monster disoriented by the
// dazzling zebra. It assumes the monster is in view. It reports whether the
// turn was consumed.
func (g *Game) monsterBumpDisoriented(i ID, ai *Actor) bool {
	ei := g.Entity(i)
	if !g.InFOV(ei.P) {
		// Monsters cannot be disoriented by the dazzling zebra outside
		// of view.
		return false
	}
	if ai.Behavior.State != Hunting && g.PlayerIsHidden() {
		// Wandering monsters that haven't noticed the player aren't
		// affected by disorientation (should be very rare with
		// the zebra, but can happen in combination with gawalt).
		return false
	}
	pl := g.PlayerEntity()
	to := ei.P.Add(g.Dir.Mul(-1)) // redirection target due to disorientation
	if to == pl.P || paths.DistanceManhattan(pl.P, to) < paths.DistanceManhattan(pl.P, ei.P) &&
		g.ActorInRange(to, pl.P) == PlayerID {
		if ai.Has(StatusFear) || (ai.DoesAny(MonsMusic) && !ai.Has(StatusBerserk)) {
			// The disoriented and afraid monster loses a turn when
			// it wants to go toward the player.
			return true
		}
		// The direction is the player's, so the monster can act as
		// usual.
		return false
	}
	if !g.Map.Passable(to) && !g.Dig(i, ai, to) {
		// The disoriented monster is facing a wall: lose a turn or dig
		// (if possible).
		return true
	}
	if j, aj := g.ActorAt(to); j >= 0 {
		// The disoriented monster may attack other adjacent monsters.
		g.BumpAttackActor(i, j, ai, aj, AttackPlain, 1, 0)
		return true
	}
	// We move to the free tile, without performing any charge.
	g.BumpMoveActor(i, ai, to)
	return true
}

// MonsterUpdateTarget makes an ai entity aware about a new target position.
// It reports whether the monster's mindstate changed. The noise argument
// indicates whether the trigger for updating the target came from noise.
func (g *Game) MonsterUpdateTarget(i ID, ai *Actor, at gruid.Point, noise bool) bool {
	pp := g.PP()
	ei, beh := g.Entity(i), ai.Behavior
	if beh.State == Hunting {
		if g.InFOV(ei.P) {
			// If player's in FOV, then PP gets priority.
			beh.Target = pp
		} else {
			beh.Target = at
		}
		return false
	}
	unhidden := !g.PlayerIsHidden()
	if g.InFOV(ei.P) && at == pp && unhidden {
		g.Logf("The %s notices you.", ei.Name)
		beh.State = Hunting
		switch {
		case ai.DoesAny(MonsBarking):
			g.Logf("The %s barks.", ei.Name)
			g.md.NoiseAnimation(ei.P)
			g.MakeNoise(ei.P, NoiseBark)
			g.PutStatus(PlayerID, g.PlayerActor(), StatusFear, DurationFearBark)
			if ai.Has(StatusConfusion) {
				g.LogfStyled("The confused %s bit its tongue while barking!", logHurtMons, g.Entity(i).Name)
				g.InflictDamage(i, ai, ConfusionDamage, AttackOther)
			}
		case ai.DoesAny(MonsSpores):
			pa := g.PlayerActor()
			if !pa.Has(StatusLignification) {
				g.releaseSpores(i)
			}
		}
	}
	if noise || unhidden {
		beh.Target = at
	}
	return beh.State == Hunting
}

// PlayerIsHidden reports whether the player is hidden and cannot be noticed by
// wandering monsters.
func (g *Game) PlayerIsHidden() bool {
	pl, pa := g.Player()
	return pa.DoesAny(Gawalt) && (pa.Has(StatusShadow) || g.menhirAt(pl.P)) ||
		pa.DoesAny(Elephanty) && g.Map.AdjacentNonPassableCount(pl.P) >= 3
}

func (g *Game) menhirAt(p gruid.Point) bool {
	if i, it := g.ItemAt(p); i >= 0 {
		_, ok := it.(*Menhir)
		return ok
	}
	return false
}

// releaseSpores releases lignifying spores in sight (as done by the wandering
// mushroom).
func (g *Game) releaseSpores(i ID) {
	const maxdist = 2*MaxFOVRange - 1
	mp := &MapPath{passable: g.Map.Passable}
	ei := g.Entity(i)
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{ei.P}, maxdist)
	if g.PR.BreadthFirstMapAt(g.PP()) > maxdist {
		return
	}
	g.Logf("The %s releases lignifying spores.", ei.Name)
	g.md.FloodAnimation(nodes, ColorYellow)
	for j, aj := range g.Actors() {
		if i == j {
			continue
		}
		ej := g.Entity(j)
		if g.PR.BreadthFirstMapAt(ej.P) > maxdist {
			continue
		}
		g.PutStatus(j, aj, StatusLignification, DurationLignificationSpores)
	}
}

// monsterBumpNext handles movement of a monster actor toward the next position
// in the precomputed path. It assumes the monster isn't afraid to go in that
// direction.
func (g *Game) monsterBumpNext(i ID, ai *Actor) {
	ei, beh := g.Entity(i), ai.Behavior
	if beh.State != Wandering || len(beh.Path) < 2 || beh.Path[0] != ei.P || beh.Path[len(beh.Path)-1] != ai.Behavior.Target {
		// Monster is hunting or was moved from its old path position,
		// so we recompute the current best path.
		beh.Path = g.MonPath(i, ai, ei.P, beh.Target)
		if beh.State != Hunting || len(beh.Path) < 2 {
			// A wandering monster spends a turn updating its path.
			// A path should always exist but, if for some reason
			// we don't find a path, lose a turn and try again on
			// next turn.
			return
		}
	}
	to := beh.Path[1] // next position in path
	pl, pa := g.Player()
	dist := paths.DistanceManhattan(ei.P, pl.P)
	// We consider the player is unnoticeable if distance is beyond FOV
	// range (to avoid redundant computations for monsters outside of view)
	// or if the monster is not hunting and the player is hidden.
	// Extra visibility considerations apply then in various cases.
	unnoticeable := dist > MaxFOVRange || beh.State != Hunting && g.PlayerIsHidden()
	if pl.P == to || dist == 1 && !unnoticeable {
		g.BumpAttackActor(i, PlayerID, ai, pa, AttackPlain, dist, 0)
		return
	}
	if g.InFOV(ei.P) && !unnoticeable && g.ActorInRange(ei.P, pl.P) == PlayerID {
		pa := g.PlayerActor()
		switch {
		case ai.DoesAny(PatternRanged|PatternRangedRecoil) ||
			ai.DoesAny(PatternCatch) && !pa.ResistsMove():
			g.md.RangedAttackAnimation(ai, ei.P, pl.P)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackRanged, dist, 0)
			return
		case ai.DoesAny(PatternSwap|PatternSwapDaze) && !ai.Has(StatusLignification) && !pa.ResistsMove():
			g.md.RangedAttackAnimation(ai, pl.P, pl.P)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackRanged, dist, 0)
			return
		case ai.DoesAny(PatternRampage) && ai.CanMove():
			dir := toDir(pl.P.Sub(ei.P))
			to := pl.P.Sub(dir)
			from := ei.P
			g.BumpMoveActor(i, ai, to)
			g.md.MoveAnimation(from, to)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackCharge, dist, 0)
			return
		}
	}
	if !ai.CanMove() {
		return
	}
	j, aj := g.ActorAt(to)
	if j < 0 {
		// Move and then try charge. An ambushing charge outside FOV
		// can happen and is allowed even while afraid.
		delta := to.Sub(ei.P)
		nto := to.Add(delta)
		g.BumpMoveActor(i, ai, to)
		if unnoticeable {
			// Non-hunting monsters don't perform charge against
			// hidden player.
			return
		}
		if ai.DoesAny(PatternFourDirsMons) {
			g.FourDirectionalAttack(i, ai, AttackCharge, 2)
		} else if j, aj := g.ActorAt(nto); j == PlayerID {
			g.BumpAttackActor(i, j, ai, aj, AttackCharge, 2, 0)
		}
		return
	}
	// Next position is obstructed by another monster.
	if g.InFOV(ei.P) && !unnoticeable {
		return
	}
	// For out of view monsters and wandering monsters in view but that
	// haven't noticed the player, we simplify logic: they may lose a turn
	// or swap positions (if both want to exchange their positions or a
	// random roll decides it). This is cheaper and avoids the need for
	// complex logic to prevent blockages.
	ej := g.Entity(j)
	if aj.CanMove() && !aj.Behavior.SkipTurn && (aj.Behavior.State != Hunting || !g.InFOV(ej.P)) &&
		(len(aj.Behavior.Path) >= 2 && aj.Behavior.Path[1] == ei.P || g.IntN(5) == 0) {
		// Swap monster positions and skip next turn.
		g.BumpMoveActor(i, ai, to)
		g.ApplyBumpMoveEffects(j, aj)
		aj.Behavior.SkipTurn = true
	}
}

// monsterFlee handles movement for an afraid hunting monster. When they're in
// a straight direction from the player, they'll flee randomly in a direction
// that avoids them (if any is free). Reports whether turn is finished.
func (g *Game) monsterFlee(i ID, ai *Actor) bool {
	if !ai.Has(StatusFear) && (!ai.DoesAny(MonsMusic) || ai.Has(StatusBerserk)) {
		return false
	}
	ei, pl := g.Entity(i), g.PlayerEntity()
	if !g.InFOV(ei.P) {
		return false
	}
	if ai.DoesAny(MonsMusic) && g.IntN(3) > 0 {
		g.MakeNoise(ei.P, NoiseMusic)
	}
	if g.ActorInRange(ei.P, pl.P) != PlayerID {
		// If the monster is almost adjacent to the player, they'll
		// just wait and lose a turn (true), otherwise they'll try to
		// approach (false).
		return paths.DistanceManhattan(ei.P, pl.P) == 2
	}
	if !ai.CanMove() {
		// The monster wants to flee but cannot move: lose a turn doing
		// nothing.
		return true
	}
	var nbs paths.Neighbors
	ps := nbs.Cardinal(ei.P, func(p gruid.Point) bool {
		if !g.IsFree(p) {
			return false
		}
		return paths.DistanceManhattan(p, pl.P) > paths.DistanceManhattan(ei.P, pl.P)
	})
	if len(ps) == 0 {
		// No free positions: lose a turn doing nothing.
		return true
	}
	to := ps[g.IntN(len(ps))]
	if to.X != pl.P.X && to.Y != pl.P.Y {
		// Extra chance of going in the opposite direction.
		to = ps[g.IntN(len(ps))]
	}
	g.BumpMoveActor(i, ai, to)
	return true
}

// HungryHunt reports whether the player is within a hungry hunt's range.
func (g *Game) HungryHunt(from gruid.Point) bool {
	pp := g.PP()
	const maxdist = (3 * MaxFOVRange) / 2
	if paths.DistanceManhattan(from, pp) > maxdist {
		return false
	}
	dij := &MapPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(dij, []gruid.Point{from}, maxdist)
	return g.PR.BreadthFirstMapAt(pp) <= maxdist
}

// RandomPassableWithin returns a random free passable point within a maximal
// path distance maxdist from the given point, with some bias toward the origin
// point.
func (g *Game) RandomPassableWithin(from gruid.Point, maxdist int) gruid.Point {
	dij := &MapPath{passable: g.Map.PassableWithoutTraps}
	nodes := g.PR.BreadthFirstMap(dij, []gruid.Point{from}, maxdist)
	if len(nodes) == 0 {
		// Should not happen.
		return from
	}
	p := nodes[g.IntN(len(nodes))].P
	q := nodes[g.IntN(len(nodes))].P
	if paths.DistanceManhattan(from, p) > paths.DistanceManhattan(from, q) {
		return q
	}
	return p
}

// RandomPassableNearbyTarget returns a random free passable point somewhat
// biased toward the given origin point.
func (g *Game) RandomPassableNearbyTarget(from gruid.Point) gruid.Point {
	p := g.RandomPassableTarget()
	q := g.RandomPassableTarget()
	if paths.DistanceManhattan(from, p) > paths.DistanceManhattan(from, q) {
		return q
	}
	return p
}

// RandomPassableTarget provides a suitable next target position for wandering
// monsters, somewhat biased toward vault waypoints.
func (g *Game) RandomPassableTarget() gruid.Point {
	if g.rand.IntN(2) == 0 && len(g.Map.Waypoints) > 0 {
		p := g.Map.Waypoints[g.IntN(len(g.Map.Waypoints))]
		return g.RandomPassableWithin(p, 5)
	}
	return g.RandomPassableWithoutTrap()
}

// CallToCommonTarget makes a few nearby monsters chose the same target.  The
// idea is to make monsters occasionally naturally form small groups in a
// simple way (without requiring more complex herd mechanics).
func (g *Game) CallToCommonTarget(i ID, ai *Actor) {
	const maxdist = MaxFOVRange - 1
	maxcalls := 2
	if g.IntN(5) < 2 {
		maxcalls = 1
	}
	k := 0
	from := g.Entity(i).P
	dij := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(dij, []gruid.Point{from}, maxdist)
	for _, n := range nodes {
		j, aj := g.ActorAt(n.P)
		if j < 0 || j == i || j == PlayerID || aj.Behavior.State == Hunting || aj.Behavior.Guard != InvalidPos {
			continue
		}
		aj.Behavior.Target = ai.Behavior.Target
		k++
		if k >= maxcalls {
			break
		}
	}
}
