package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// AttackKind represents different kinds of attack kinds.
type AttackKind int

const (
	AttackPlain  AttackKind = iota // Attack against adjacent foe without moving
	AttackCharge                   // Attack against adjacent foe after moving
	AttackRanged                   // Attack against non-adjacent foe without moving (swapping occurs after the attack)
	AttackDrag                     // Attack against adjacent foe that may move backwards (crocodile)
	AttackSlap                     // Attack against adjacent foes (crocodile's tail slap)

	AttackGale      // pushing gale damage
	AttackLightning // lightning damage
	AttackFire      // fire damage
	AttackFireCatch // catching fire damage
	AttackPoison    // poison damage
	AttackOther     // other damage sources (not logged by InflictDamage)
)

const (
	// HPCritical represents the critical HP threshold.
	HPCritical = 3

	// MaxAttackDamage represents the maximum damage that can be inflicted
	// in one attack.
	MaxAttackDamage = 3

	BerserkAttackBonus            = 2
	VampirismAttackBonus          = 2
	ConfusedImbalanceDefenseBonus = 2
	LignificationDefenseBonus     = 2
)

// GetAttack returns the actor's current attack with modifiers.
func (a *Actor) GetAttack() int {
	attack := a.Attack
	if a.Has(StatusBerserk) {
		attack += BerserkAttackBonus
	}
	if a.Has(StatusVampirism) {
		attack += VampirismAttackBonus
	}
	if a.Has(StatusImbalance) {
		attack /= 2
	}
	return attack
}

// GetDefense returns the actor's current defense with modifiers.
func (a *Actor) GetDefense() int {
	defense := a.Defense
	if a.Has(StatusLignification) {
		defense += LignificationDefenseBonus
	}
	if a.Has(StatusImbalance) && a.Has(StatusConfusion) {
		defense += ConfusedImbalanceDefenseBonus
	}
	return defense
}

// GetMaxHP returns the actor's current max HP with modifiers.
func (a *Actor) GetMaxHP() int {
	hp := a.MaxHP
	if a.Has(StatusBerserk) {
		hp += a.HPBonus()
	}
	if a.Has(StatusLignification) {
		hp += a.HPBonus()
	}
	return hp
}

// InflictDamage deals the given amount of non-bump damage to the actor, after
// applying maximum restrictions. It handles critical mode and extra effects.
func (g *Game) InflictDamage(i ID, ai *Actor, dmg int, ap AttackKind) int {
	return g.InflictDamageGeneric(-1, i, nil, ai, dmg, ap)
}

// InflictDamageGeneric deals the given amount of damage to the actor, after
// applying maximum restrictions. It handles critical mode and extra effects.
func (g *Game) InflictDamageGeneric(i, j ID, ai, aj *Actor, dmg int, ap AttackKind) int {
	if dmg <= 0 || aj.IsDead() {
		return 0
	}
	dmg = min(dmg, aj.HP)
	aj.Statuses[StatusDaze] = 0 // no logging as it's already clear from the damage
	aj.HP -= dmg
	// Logging.
	ej := g.Entity(j)
	switch {
	case i >= 0 && ai != nil:
		g.logBumpDamage(i, j, ai, aj, dmg, ap)
		if ai.Has(StatusVampirism) {
			g.AdjustHP(i, ai, dmg)
		}
		if ai.DoesAny(PatternSwap|PatternSwapDaze) || ai.Has(StatusShadow) || aj.Has(StatusShadow) || g.IntN(dmg) == 0 {
			break
		}
		switch {
		case ai.DoesAny(PatternCrocodile) && ap != AttackSlap:
			g.MakeNoise(ej.P, NoiseChomp)
		default:
			g.MakeNoise(ej.P, NoiseCombat)
		}
	case !g.InFOV(ej.P):
		break
	case ap == AttackGale:
		if aj.IsDead() {
			g.LogfStyled("The gale kills the %s (%d dmg).", logHurtMons, ej.Name, dmg)
		} else {
			g.LogfStyled("The gale hurts the %s (%d dmg).", logHurtMons, ej.Name, dmg)
		}
	case ap == AttackLightning:
		if aj.IsDead() {
			g.LogfStyled("The lightning kills the %s (%d dmg).", logHurtMons, ej.Name, dmg)
		} else {
			g.LogfStyled("The lightning hurts the %s (%d dmg).", logHurtMons, ej.Name, dmg)
		}
	case ap == AttackFire:
		if j == PlayerID {
			g.LogfStyled("The fire burns you (%d dmg).", logHurtPlayer, FireDamage)
			g.StoryLogf("Burnt by fire (HP: %d/%d)", aj.HP, aj.GetMaxHP())
			break
		}
		if aj.IsDead() {
			g.LogfStyled("The fire kills the %s.", logHurtMons, ej.Name)
		}
	case ap == AttackFireCatch:
		if j == PlayerID {
			g.LogfStyled("Catching fire burns you (%d dmg).", logHurtPlayer, FireDamage)
			g.StoryLogf("Caught fire (HP: %d/%d)", aj.HP, aj.GetMaxHP())
			break
		}
		if aj.IsDead() {
			g.LogfStyled("Catching fire kills the %s.", logHurtMons, ej.Name)
		}
	case ap == AttackPoison:
		if j == PlayerID {
			g.LogfStyled("Moving while poisoned hurts you (%d dmg).", logHurtPlayer, dmg)
			g.StoryLogf("Moved while poisoned (HP: %d/%d)", aj.HP, aj.GetMaxHP())
			break
		}
		if aj.IsDead() {
			g.LogfStyled("The %s dies from poison.", logHurtMons, ej.Name)
		}
	case j != PlayerID && aj.IsDead() && !aj.DoesAny(MonsExplodingDeath):
		g.Logf("The %s dies.", ej.Name)
	}
	// Extra effects (gathering stats & monster death/explosion).
	switch {
	case j == PlayerID:
		g.Stats.Hurt++
		g.Stats.Damage += dmg
		g.Stats.MapDamage[g.Map.Level-1] += dmg
		g.md.WoundedAnimation()
		if g.Wizard.Mode != WizardNone && aj.IsDead() {
			aj.HP = aj.GetMaxHP()
			g.LogStyled("*** Wizard mode resurrects you! ***", logSpecial)
			g.StoryLog("Resurrected by the Wizard!")
			g.Wizard.Resurrections++
		}
		if aj.HP+dmg > HPCritical && aj.HP <= HPCritical {
			g.md.mode = modeCritical
		}
		g.md.auto.mode = noAuto
	case aj.IsDead():
		g.monsterDeath(i, j, aj)
	}
	return dmg
}

func (g *Game) logBumpDamage(i, j ID, _, aj *Actor, dmg int, ap AttackKind) {
	ei, ej := g.Entity(i), g.Entity(j)
	switch {
	case i == PlayerID:
		switch {
		case aj.IsDead():
			if ap == AttackCharge {
				g.LogfStyled("Your charge kills the %v (%d dmg).", logHurtMons, ej.Name, dmg)
			} else {
				g.LogfStyled("You kill the %v (%d dmg).", logHurtMons, ej.Name, dmg)
			}
			g.Stats.Hits++
		default:
			if ap == AttackCharge {
				g.LogfStyled("You charge the %v (%d dmg).", logHurtMons, g.Entity(j).Name, dmg)
			} else {
				g.LogfStyled("You hit the %v (%d dmg).", logHurtMons, g.Entity(j).Name, dmg)
			}
			g.Stats.Hits++
		}
	case j == PlayerID:
		hpstr := fmt.Sprintf("HP: %d/%d", aj.HP, aj.GetMaxHP())
		if ap == AttackCharge {
			g.LogfStyled("The %s charges you (%d dmg).", logHurtPlayer, ei.Name, dmg)
			g.StoryLogf("Charged at by %s for %d dmg (%s)", ei.Name, dmg, hpstr)
		} else {
			g.LogfStyled("The %v hits you (%d dmg).", logHurtPlayer, ei.Name, dmg)
			g.StoryLogf("Hit by %s for %d dmg (%s)", ei.Name, dmg, hpstr)
		}
	default:
		if !g.InFOV(ej.P) {
			return
		}
		switch {
		case aj.IsDead():
			g.LogfStyled("The %s kills the %s (%d dmg).", logHurtMons, ei.Name, ej.Name, dmg)
		default:
			g.LogfStyled("The %s hits the %s (%d dmg).", logHurtMons, ei.Name, ej.Name, dmg)
		}
	}
}

// monsterDeath handles extra effects on monster death (like for exploding
// nadres). The first optional id is the id of the actor that was acting when
// the death happened.
func (g *Game) monsterDeath(i, j ID, aj *Actor) {
	ej := g.Entity(j)
	g.Stats.Death(ej.Name)
	g.md.DeathAnimation(ej.P)
	if g.InFOV(ej.P) {
		// Mark as known dead here, because due to clouds the player
		// might not see it later otherwise.
		aj.KnownDead = true
	}
	if g.Mod(ModHealingCombat) && !g.HasVampirism() {
		pa := g.PlayerActor()
		if pa.HP < pa.GetMaxHP() && g.IntN(pa.HP) < 4 {
			aj.KnownDead = true // You feel the monster's death as you heal.
			g.AdjustHP(PlayerID, g.PlayerActor(), 1)
		}
	}
	if !aj.DoesAny(MonsExplodingDeath) {
		return
	}
	if g.InFOV(ej.P) {
		g.Logf("The %s explodes.", ej.Name)
	}
	g.ExplosionAt(i, ej.P)
}

// ExplosionAt makes a fire explosion at the given position. The optional id
// is the id of the actor that was acting when the explosion was triggered.
func (g *Game) ExplosionAt(i ID, at gruid.Point) {
	g.MakeNoise(at, NoiseExplosion)
	var nbs paths.Neighbors
	ps := append(nbs.Cardinal(at, inMap), at)
	g.md.ExplosionAnimation(ps)
	for _, q := range ps {
		if g.Map.Burnable(q) {
			g.FireCloudAt(q)
		} else if !g.digAt(q) {
			g.NormalCloudAt(q, 2+g.IntN(3))
		}
		if j, aj := g.ActorAt(q); j >= 0 {
			if j == i {
				g.PutStatus1(j, aj, StatusFire, DurationFire)
			} else {
				g.PutStatus(j, aj, StatusFire, DurationFire)
			}
		}
	}
}

// RangedTargetInDir returns the first visible actor target in the given
// direction.
func (g *Game) RangedTargetInDir(from, dir gruid.Point) (ID, *Actor) {
	for p := from.Add(dir); g.Map.Passable(p) && g.InFOV(p); p = p.Add(dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			return i, ai
		}
	}
	return -1, nil
}

// ActorInRange returns an actor at a given position reachable from a given one
// (for a ranged or attraction attack).
func (g *Game) ActorInRange(from, at gruid.Point) ID {
	if at.X != from.X && at.Y != from.Y {
		return -1
	}
	dir := toDir(at.Sub(from))
	j, _ := g.RangedTargetInDir(from, dir)
	return j
}

// FourDirectionalAttack handles extra attacks behind and on the sides for
// actors with PatternFourDirs or PatternFourDirsMons trait. Those kinds of
// attacks can be performed even while afraid as long as they're done laterally
// while moving.
func (g *Game) FourDirectionalAttack(i ID, ai *Actor, ap AttackKind, dist int) bool {
	ei := g.Entity(i)
	var actors []*Actor
	var ids []ID
	for j, aj := range g.AdjacentActors(ei.P) {
		// Either attacker or defender should be the player. Confused
		// monsters fighting each other in discord don't perform
		// four-directional attacks.
		if i == PlayerID || j == PlayerID {
			actors = append(actors, aj)
			ids = append(ids, j)
		}
	}
	// Shuffle foes so that attack order is not predictable in an
	// artificial way.
	g.rand.Shuffle(len(actors), func(i, j int) {
		actors[i], actors[j] = actors[j], actors[i]
		ids[i], ids[j] = ids[j], ids[i]
	})
	bonus := len(actors)
	if bonus <= 1 {
		// We only give +2, +3, +4 bonuses when there's at least two
		// adjacent foes or more.
		bonus = 0
	}
	for k, j := range ids {
		g.BumpAttackActor(i, j, ai, actors[k], ap, dist, bonus)
	}
	return len(actors) > 0
}

// BumpAttackActor does an attack against an actor at a given position.
// It returns the damage done.
func (g *Game) BumpAttackActor(i, j ID, ai, aj *Actor, ap AttackKind, dist, bonus int) {
	redirected := false
	if j == PlayerID {
		// Ensure that an attacking monster becomes a hunting one even
		// if the player is shadowed.
		ai.Behavior.State = Hunting
		if aj.DoesAny(Dazzling) && (ap == AttackCharge || ap == AttackPlain ||
			ap == AttackRanged && !ai.DoesAny(PatternCatch)) {
			if k, ak := g.dazzlingRedirect(i); k >= 0 {
				j, aj = k, ak
				redirected = true
			}
		}
	}
	dazed := aj.Has(StatusDaze)
	ei, ej := g.Entity(i), g.Entity(j)
	dmg := g.AttackDamage(i, j, ai, aj, ap, bonus)
	if dmg <= 0 {
		if ai.DoesAny(PatternSwap|PatternSwapDaze) && ai.CanMove() && ap == AttackRanged {
			pi, pj := ei.P, ej.P
			g.BumpMoveActor(i, ai, pi.Add(toDir(pj.Sub(pi))))
		}
		return
	}
	afraid := aj.Has(StatusFear)
	// Extra effects by attack pattern.
	switch {
	case ai.DoesAny(Pushing|PushingCharge) && ai.CanMove() && !aj.ResistsMove() &&
		(ap == AttackCharge || !ai.Has(StatusPoison)) && !redirected:
		if ai.Has(StatusDig) || ai.DoesAny(PushingCharge) && ap == AttackCharge && dist > 2 ||
			ai.DoesAny(Pushing) && g.IntN(5-dmg) < 2 {
			g.push(i, j, ai, aj, ap, dist, bonus)
		}
	case ai.DoesAny(PatternSwap|PatternSwapDaze) && ap != AttackCharge && !ai.Has(StatusLignification) && !aj.ResistsMove():
		g.swapSpaceDistortion(i, j, ai, aj, dmg, dazed)
	case ai.DoesAny(PatternCatch) && ap == AttackRanged && !aj.ResistsMove():
		g.catch(i, j, ai, aj, dist)
	case ai.DoesAny(PatternRangedRecoil) && ap != AttackCharge && ai.CanMove():
		if g.recoil(i, j, ai, dist) && i != PlayerID &&
			ai.Has(StatusConfusion) && g.PutStatus1(i, ai, StatusImbalance, 1+dmg) && g.InFOV(ei.P) {
			g.Logf("The confused %s lost balance due to recoil!", ei.Name)
		}
	case ai.DoesAny(PatternCrocodile) && ai.CanMove() && !aj.ResistsMove() && ap == AttackDrag && !ai.Has(StatusPoison) && aj.IsAlive():
		g.drag(i, j, ai, aj)
	case ai.DoesAny(PatternBat) && ap == AttackPlain && ai.CanMove() && !ai.Has(StatusPoison):
		g.retreat(i, j, ai, aj)
	case ai.DoesAny(PatternBat) && ap == AttackCharge:
		g.PutStatus(j, aj, StatusConfusion, 2)
	}
	// Extra attack effects (shared by both player and monsters).
	if ai.DoesAny(VenomousMelee) && (ap != AttackRanged || g.AdjacentEntities(i, j)) && g.IntN(6-dmg) < 2 {
		if g.PutStatus(j, aj, StatusPoison, DurationPoisonBite) && i != PlayerID && ai.Has(StatusConfusion) && ai.IsAlive() {
			if g.InFOV(ei.P) {
				g.LogfStyled("The confused %s bites itself too!", logHurtMons, ei.Name)
			}
			g.InflictDamage(i, ai, ConfusionDamage, AttackOther)
		}
	}
	if ai.DoesAll(BurningHits) && g.IntN(6-dmg) < 2 {
		if g.PutStatus(j, aj, StatusFire, DurationFire) {
			if p := ej.P; g.Map.Burnable(p) {
				g.FireCloudAt(p)
			}
		}
	}
	// Monster-specific extra attack effects (at most one per monster).
	switch {
	case ai.DoesAll(MonsSpitFire) && ap == AttackRanged:
		if g.PutStatus(j, aj, StatusFire, DurationFire) {
			if p := ej.P; g.Map.Burnable(p) {
				g.FireCloudAt(p)
			}
			if ai.Has(StatusConfusion) && i != PlayerID && g.PutStatus(i, ai, StatusFire, DurationFire) {
				g.Logf("The confused %s burned itself too!", ei.Name)
			}
		}
	case ai.DoesAll(MonsFear) && g.IntN(5-dmg) < 2:
		if g.PutStatus(j, aj, StatusFear, DurationFearUndead) {
			if ai.Has(StatusConfusion) && g.PutStatus1(i, ai, StatusFear, DurationFearUndead) && g.InFOV(ei.P) {
				g.Logf("The confused %s frightened itself too!", ei.Name)
			}
		}
	case ai.DoesAny(MonsConfusion) && g.IntN(6-dmg) < 2:
		if g.PutStatus(j, aj, StatusConfusion, DurationConfusionHit) {
			g.monsConfusionPenalty(i, ai)
		}
	case ai.DoesAny(MonsLignify) && g.IntN(6-dmg) < 2:
		if g.PutStatus(j, aj, StatusLignification, DurationLignificationHit) {
			g.monsConfusionPenalty(i, ai)
		}
	case ai.DoesAny(MonsBerserking) && g.IntN(6-dmg) < 2:
		if g.PutStatus(j, aj, StatusBerserk, DurationBerserkHit) {
			g.monsConfusionPenalty(i, ai)
		}
	case ai.DoesAny(MonsBlink) && aj.IsAlive():
		g.blinkOther(i, j, ai, aj)
	case ai.DoesAny(MonsTeleport) && aj.IsAlive() && !aj.ResistsMove():
		if j == PlayerID {
			g.Logf("The %s teleports you away!", ei.Name)
		} else {
			g.Logf("The confused %s teleports the %s away!", ei.Name, ej.Name)
		}
		g.TeleportActor(j, aj, 0)
		if ai.Has(StatusConfusion) && ai.IsAlive() {
			g.TeleportActor(i, ai, 1)
		}
	}
	// Extra defense effects.
	switch {
	case aj.DoesAny(DazingSpines) && (ap != AttackRanged || g.AdjacentEntities(i, j)) && g.IntN(5-dmg) == 0:
		g.PutStatus1(i, ai, StatusDaze, DurationDazeSpines)
	}
	if afraid && aj.Has(StatusFear) && aj.Has(StatusLignification) && !aj.DoesAny(WoodyLegs) {
		// When hurt, afraid lignified actors become berserk.
		g.PutStatus(j, aj, StatusBerserk, DurationBerserkAfraid)
	}
}

// Percentage chances for damage and absorption rolls based on attack
// and defense. Currently, maximum attack is 10 (surrounded berserk
// hydra with offensive spirits) and maximum defense is 7 (lignified
// frog with the most defensive spirits), but we add some extra values
// in case those limits go up.
var probA = [12]int{0, 35, 44, 52, 59, 66, 72, 78, 84, 90, 95, 100}
var probD = [10]int{0, 29, 42, 53, 62, 70, 76, 81, 86, 90}

// computeDamage computes a damage roll with given attack and defense values.
func (g *Game) computeDamage(a, d int) int {
	if g.IntN(100) < 5 {
		// Ensure at least 5% miss chance.
		return 0
	}
	var dmg int
	// Attack-based damage dealing rolls.
	for range 3 {
		if g.IntN(100) < probA[a] {
			dmg++
		}
	}
	// Defense-based damage absorption rolls.
	for range min(2, dmg) {
		if g.IntN(100) < probD[d] {
			dmg--
		}
	}
	return max(0, min(a, dmg))
}

// AttackDamage inflicts bump attack damage by an actor on another, and
// returns the damage done (zero for a miss).
func (g *Game) AttackDamage(i, j ID, ai, aj *Actor, ap AttackKind, bonus int) int {
	switch {
	case ap == AttackCharge && ai.DoesAny(PatternRampage) && !aj.ResistsMove():
		// Rampage has an attack bonus when charging, except against
		// walking trees and lignified oponents.
		bonus++
	case ap == AttackRanged && ai.DoesAny(PatternCatch):
		// Catching ranged attacks get an attack bonus.
		bonus++
	}
	attack, defense := ai.GetAttack()+bonus, aj.GetDefense()
	if attack <= 0 {
		// Ensure at least 1 offense.
		attack = 1
	}
	if ap == AttackRanged && (aj.Has(StatusLignification) || aj.DoesAny(MonsLignify|MonsScales)) {
		defense++
	}
	if ai.DoesAny(MonsIgnoreDefense) {
		defense = 0
	}
	n := 1 // default number of attacks
	if ai.Has(StatusFocus) || ai.DoesAny(MonsFourHeaded) {
		n += 3 // extra attacks for each extra head
	}
	var dmg int
	for range n {
		// Damage per attack.
		dmg += g.computeDamage(attack, defense)
	}
	if aj.Has(StatusLignification) || aj.DoesAny(MonsLignify) {
		// Lignification caps attack damage at 1.
		dmg = min(1, dmg)
	}
	if i == PlayerID && ai.DoesAny(Gawalt) && !ai.Has(StatusShadow) && dmg > 1 {
		// Gawalt monkey reduces damage of non-weak hits.
		dmg--
	}
	if ai.Has(StatusDig) && (ap == AttackCharge || dmg == 0) || ap == AttackSlap || dmg == 0 && ai.Has(StatusVampirism) {
		// Dig status gives extra *effective* damage on charge and
		// ensures at least 1 damage.
		// Crocodile's tail-slapping has extra effective damage, too.
		// Vampirism ensures at least 1 damage.
		dmg++
	}
	if ai.Has(StatusBerserk) {
		// Berserk status gives extra *effective* damage.
		dmg++
	}
	dmg = min(dmg, aj.HP)
	if j == PlayerID {
		if dmg == aj.HP &&
			(dmg >= 2 && g.IntN(2+dmg) > 0 || dmg == 1 && !ai.Has(StatusBerserk) && g.IntN(2) == 0) {
			// Lucky roll adjustment for increased game tension
			// before death and more newbie-friendliness.
			dmg--
		}
		// Compute reasonable maximum expected damage for “times lucky”
		// purposes. We don't take everything into account (like fire
		// from phoenix or llama). More than 5 damage, even from hydra,
		// is always considered unexpected.
		maxdmg := 5
		switch {
		case aj.Has(StatusLignification):
			maxdmg = 1
		case defense > 3:
			// If the player has 4 or more defense, more than 3
			// damage even from a hydra is unlikely.
			maxdmg = 3
		case defense == 3:
			// If the player has 3 defense, more than 4 damage from
			// a hydra is unlikely.
			maxdmg = 4
		}
		m := min(maxdmg, min(attack, 3)*n)
		if ai.Has(StatusBerserk) {
			m++
		}
		if m >= aj.HP && dmg < aj.HP {
			// We avoided death with a lucky roll.
			g.Stats.Lucky++
		}
	}
	if dmg <= 0 {
		g.logBumpMiss(i, j, ap)
		return 0
	}
	pj := g.Entity(j).P
	if j != PlayerID && dmg < aj.HP {
		g.md.HitAnimation(pj)
	}
	dmg = g.InflictDamageGeneric(i, j, ai, aj, dmg, ap) // dmg actually remains the same
	if aj.Behavior != nil && aj.Behavior.State != Hunting && ai.IsPlayer() {
		// Surprise charging attacks make monsters hunt immediately.
		aj.Behavior.State = Hunting
		aj.Behavior.Target = g.PP()
	}
	return dmg
}

func (g *Game) logBumpMiss(i, j ID, ap AttackKind) {
	ei, ej := g.Entity(i), g.Entity(j)
	switch {
	case i == PlayerID:
		if ap == AttackCharge {
			g.Logf("Your charge misses the %v.", ej.Name)
		} else {
			g.Logf("You miss the %v.", ej.Name)
		}
		g.Stats.Misses++
	case j == PlayerID:
		if ap == AttackCharge {
			g.Logf("The %s’s charge misses you.", ei.Name)
		} else {
			g.Logf("The %v misses you.", ei.Name)
		}
	}
	// NOTE: we don't log misses between monsters, as that's not very
	// interesting information.
}

func (g *Game) monsConfusionPenalty(i ID, ai *Actor) {
	if ai.Has(StatusConfusion) && ai.IsAlive() {
		if g.InFOV(g.Entity(i).P) {
			g.LogfStyled("The confused %s hurts itself too!", logHurtMons, g.Entity(i).Name)
		}
		g.InflictDamage(i, ai, ConfusionDamage, AttackOther)
	}
}

// push handles extra penetrating and pushing effects for actors with
// Pushing trait.  It should be called only on pairs of adjacent actors.
func (g *Game) push(i, j ID, ai, aj *Actor, ap AttackKind, dist, bonus int) {
	ei, ej := g.Entity(i), g.Entity(j)
	pi, pj := ei.P, ej.P
	delta := pj.Sub(pi)
	npj := pj.Add(delta)
	d := 2 + 2*(dist/5) // imbalance duration (2 for dist 2-4, 4 for dist 5-8)
	if g.PutStatus(j, aj, StatusImbalance, d) && i != PlayerID &&
		ai.Has(StatusConfusion) && g.PutStatus1(i, ai, StatusImbalance, d) && g.InFOV(ei.P) {
		g.Logf("The confused %s lost balance while pushing too!", ei.Name)
	}
	k, ak := g.ActorAt(npj)
	if ak != nil {
		// There's another actor behind: extra piercing attack (without
		// extra effects, but inheriting any charge attack bonuses).
		g.AttackDamage(i, k, ai, ak, ap, bonus)
	}
	move := true // whether to move pushing actor to pj
	switch {
	case !g.Map.Passable(npj) || ak != nil && ak.IsAlive():
		// No room for pushing, so move only if the foe is dead.
		move = aj.IsDead()
	case aj.IsAlive():
		// Push.
		if j != PlayerID && (i == PlayerID || g.InFOV(pj)) {
			g.Logf("The %v is pushed away.", ej.Name)
		} else {
			g.Log("You’re pushed away.")
			g.StoryLogf("Pushed away by %s", ei.Name)
		}
		g.MoveActor(j, aj, npj)
		g.md.MoveAnimation(pj, npj)
	}
	if i == PlayerID {
		switch g.Map.Clouds.At(pj).Kind {
		case CloudFire, CloudPoison:
			// Player is smarter than corrupted monsters: we don't
			// move when foe is on a dangerous cloud.
			move = false
		}
	}
	if !move || ai.IsDead() {
		return
	}
	if ap == AttackCharge {
		// Already bump-moved, so extra movement effects were processed
		// already.
		g.MoveActor(i, ai, pj)
	} else {
		g.BumpMoveActor(i, ai, pj)
	}
}

// catch handles extra catching effects for actors with PatternCatch trait.
func (g *Game) catch(i, j ID, ai, aj *Actor, dist int) {
	ei, ej := g.Entity(i), g.Entity(j)
	dir := toDir(ej.P.Sub(ei.P))
	to := ei.P.Add(dir)
	from := ej.P
	g.MoveActor(j, aj, to)
	g.md.MoveAnimation(from, to)
	d := 2 + 2*(dist/5) // imbalance duration (2 for dist 2-4, 4 for dist 5-8)
	if g.PutStatus(j, aj, StatusImbalance, d) && i != PlayerID &&
		ai.Has(StatusConfusion) && g.PutStatus1(i, ai, StatusImbalance, d) {
		g.Logf("The confused %s lost balance too!", ei.Name)
	}
}

// swapSpaceDistortion handles extra swapping effects for actors with
// PatternSwap trait, after inflicting damage on target.
func (g *Game) swapSpaceDistortion(i, j ID, ai, aj *Actor, dmg int, dazed bool) {
	pj := g.Entity(j).P
	g.SwapClouds(g.Entity(i).P, pj)
	g.MoveActor(i, ai, pj) // swaps positions
	if !dazed && ai.DoesAny(PatternSwapDaze) && g.IntN(6-dmg) < 2 {
		duration := 2 + dmg/2
		if g.PutStatus(j, aj, StatusDaze, duration) && i != PlayerID &&
			ai.Has(StatusConfusion) && g.PutStatus1(i, ai, StatusDaze, duration) && g.InFOV(pj) {
			g.LogfStyled("The confused %s dazed itself too!",
				logHurtMons, g.Entity(i).Name)
		}
	}
	for p := range g.Map.PassableNeighbors(pj) {
		if k, ak := g.ActorAt(p); k >= 0 && k != j && !ak.ResistsMove() {
			ek := g.Entity(k)
			if to, ok := g.BlinkPos(ek.P); ok {
				if k == PlayerID {
					g.Log("The space distortion blinks you away.")
				} else {
					g.Logf("The space distortion blinks the %s away.", ek.Name)
				}
				g.md.TeleportAnimation(ek.P, to, true)
				g.MoveActor(k, ak, to)
			}
		}
	}
	if i == PlayerID || j == PlayerID {
		// Reverse direction too, because it makes sense that after
		// swapping positions both actors are still facing each other.
		g.Dir = g.Dir.Mul(-1)
	}
}

// recoil handles extra recoil effects for actors with the PatternRangedRecoil
// trait.
func (g *Game) recoil(i, j ID, ai *Actor, dist int) bool {
	if dist > 2 || dist == 2 && g.IntN(3) == 0 {
		return false
	}
	ei, ej := g.Entity(i), g.Entity(j)
	dir := toDir(ei.P.Sub(ej.P))
	npi := ei.P.Add(dir)
	if !g.IsFree(npi) {
		return false
	}
	if dist == 1 {
		p := npi.Add(dir)
		if g.IsFree(p) {
			npi = p
			g.MakeNoise(ei.P, NoiseWind)
		}
	}
	pi := ei.P
	g.MoveActor(i, ai, npi)
	g.md.MoveAnimation(pi, npi)
	return true
}

// drag attemps to drag a foe backwards.
func (g *Game) drag(i, j ID, ai, aj *Actor) {
	ei, ej := g.Entity(i), g.Entity(j)
	pi, pj := ei.P, ej.P
	delta := pj.Sub(pi)
	to := pi.Sub(delta)
	if !g.IsFree(to) {
		return
	}
	if i == PlayerID {
		switch g.Map.Clouds.At(to).Kind {
		case CloudFire, CloudPoison:
			// Player is smarter than corrupted monsters: we don't
			// move onto a dangerous cloud.
			return
		}
	}
	g.PutStatus(j, aj, StatusImbalance, DurationImbalanceDragging)
	// Move backwards.
	g.BumpMoveActor(i, ai, to)
	g.MoveActor(j, aj, pi)
	g.md.MoveAnimation(pj, pi)
}

// retreat attemps to a sneaky retreat backwards.
func (g *Game) retreat(i, j ID, ai, aj *Actor) {
	ei, ej := g.Entity(i), g.Entity(j)
	dir := toDir(ei.P.Sub(ej.P))
	from, to := ei.P, ei.P
	for range 2 {
		p := to.Add(dir)
		if !g.IsFree(p) {
			break
		}
		to = p
	}
	if from == to {
		return
	}
	g.BumpMoveActor(i, ai, to)
	g.md.MoveAnimation(from, to)
}

// AdjacentEntities reports whether two entities are adjacent.
func (g *Game) AdjacentEntities(i, j ID) bool {
	return paths.DistanceManhattan(g.Entity(i).P, g.Entity(j).P) == 1
}

// Dig makes an actor attempt digging at a given position. It reports whether
// the position is now passable.
func (g *Game) Dig(i ID, ai *Actor, at gruid.Point) bool {
	if !inMap(at) {
		return false
	}
	if !ai.Has(StatusDig) {
		return false
	}
	g.digAt(at)
	return true
}

// digAt performs digging at a given position. It reports whether any walls were
// dug.
func (g *Game) digAt(at gruid.Point) bool {
	t := g.Map.Terrain.At(at)
	if Passable(t) {
		return false
	}
	g.Stats.Digs++
	g.Map.Terrain.Set(at, Rubble)
	if t == TranslucentWall {
		g.PoisonCloudAt(at)
	}
	g.MakeNoise(at, NoiseDig)
	pp := g.PP()
	if t == TranslucentWall && paths.DistanceManhattan(pp, at) <= MaxFOVRange {
		g.UpdateFOV()
		g.UpdateKnowledge()
		if pp != at {
			g.md.AnimationFrame()
		}
	}
	return true
}

// blinkOther implements the blinking bump effect.
func (g *Game) blinkOther(i, j ID, ai, aj *Actor) {
	if aj.IsDead() || aj.ResistsMove() {
		return
	}
	ej := g.Entity(j)
	if to, ok := g.BlinkPos(ej.P); ok {
		g.md.TeleportAnimation(ej.P, to, true)
		g.MoveActor(j, aj, to)
		ei := g.Entity(i)
		if j == PlayerID {
			g.Logf("The %s’s attack made you blink away.", ei.Name)
			g.StoryLogf("Blinked away by %s", ei.Name)
		}
		if ai.Has(StatusConfusion) && !ai.ResistsMove() {
			if to, ok := g.BlinkPos(ei.P); ok {
				g.md.TeleportAnimation(ei.P, to, g.InFOV(to))
				g.MoveActor(i, ai, to)
				g.LogfStyled("The confused %s hurt and blinked itself too!", logHurtMons, ei.Name)
				g.InflictDamage(i, ai, ConfusionDamage, AttackOther)
			}
		}
	}
}

// BlinkPos returns a suitable blinking position in FOV from the given
// position.
func (g *Game) BlinkPos(from gruid.Point) (gruid.Point, bool) {
	if !g.InFOV(from) {
		// Disable blinking when out of FOV, like it can happen when
		// confused monsters fight among themselves. Computing a
		// separate FOV for the monster would not be worth the
		// complexity, and not even really consistent with how it works
		// when in FOV.
		return InvalidPos, false
	}
	pp := g.PP()
	ps := []gruid.Point{}
	for _, p := range g.Map.FOVPts {
		if paths.DistanceManhattan(p, pp) > MaxFOVRange || p == from {
			continue
		}
		if cost, ok := g.Map.FOV.At(p); ok && cost <= MaxFOVRange && g.IsFree(p) {
			ps = append(ps, p)
		}
	}
	if len(ps) == 0 {
		return InvalidPos, false
	}
	return ps[g.IntN(len(ps))], true
}

// dazzlingRedirect may return a new target for the given monster, suitable for
// cases where the the player has the Dazzling trait.
func (g *Game) dazzlingRedirect(i ID) (ID, *Actor) {
	ei := g.Entity(i)
	pp := g.PP()
	dir := toDir(pp.Sub(ei.P))
	return g.ActorAt(pp.Add(dir))
}
