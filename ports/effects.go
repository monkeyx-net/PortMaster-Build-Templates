package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// The following constants are used for various effects and statuses.
const (
	DurationBerserk                = 8
	DurationBerserkAfraid          = 4
	DurationBerserkHit             = 4
	DurationBerserkTotem           = 5
	DurationBerserkTrap            = 5
	DurationClarity                = 5
	DurationConfusionAmbrosia      = 6
	DurationConfusionFireTrap      = 3
	DurationConfusionHit           = 6
	DurationConfusionPepperBreath  = 2
	DurationConfusionSkunk         = 5
	DurationConfusionToxins        = 4
	DurationDazeFall               = 1
	DurationDazeJump               = 4
	DurationDazeLightning          = 4
	DurationDazeSpines             = 2
	DurationDazeTeleport           = 4
	DurationFearAmbrosia           = 4
	DurationFearBark               = 4
	DurationFearFakePortal         = 6
	DurationFearUndead             = 3
	DurationFire                   = 3
	DurationFoggySkin              = 10
	DurationImbalanceGale          = 5
	DurationImbalanceLignification = 5
	DurationImbalanceSprint        = 5
	DurationImbalanceTeleport      = 5
	DurationLignification          = 12
	DurationLignificationHit       = 6
	DurationLignificationRoots     = 6
	DurationLignificationSpores    = 6
	DurationLignificationTrap      = 8
	DurationPoisonBerserk          = 5
	DurationPoisonBite             = 4
	DurationPoisonCloud            = 3

	HealAmbrosia         = 8
	HealClarityLeaves    = 4
	HealFirebreathPepper = 2
	HealFoggySkinOnion   = 3

	DurationDig      = 5
	DurationFocus    = 3
	DurationSprint   = 6
	DurationTimeStop = 5
)

func newHP(a *Actor, hp int) string {
	return fmt.Sprintf("%d/%d", max(0, min(a.GetMaxHP(), a.HP+hp)), a.GetMaxHP())
}

func (g *Game) eatMsg(name string, hpstr string) {
	g.Logf("You eat the %s.", name)
	if hpstr != "" {
		g.StoryLogf("Ate %s (HP: %s)", One(name), hpstr)
	} else {
		g.StoryLogf("Ate %s", One(name))
	}
}

type EffectAmbrosiaBerries struct{}

func (eff EffectAmbrosiaBerries) Apply(g *Game) bool {
	pa := g.PlayerActor()
	g.eatMsg("ambrosia berries", newHP(pa, HealAmbrosia))
	g.AdjustHP(PlayerID, pa, HealAmbrosia)
	if g.IntN(2) == 0 {
		g.PutStatus1(PlayerID, pa, StatusConfusion, DurationConfusionAmbrosia)
	} else {
		g.PutStatus1(PlayerID, pa, StatusFear, DurationFearAmbrosia)
	}
	return true
}

func (eff EffectAmbrosiaBerries) Desc() string {
	return fmt.Sprintf("Restores %+d HP. Leaves you either @Oconfused@N or @Oafraid@N for %d or %d turns respectively.",
		HealAmbrosia, DurationConfusionAmbrosia, DurationFearAmbrosia)
}

type EffectBerserkingFlower struct{}

func (eff EffectBerserkingFlower) Apply(g *Game) bool {
	pa := g.PlayerActor()
	d := DurationBerserk
	if pa.Has(StatusClarity) {
		// Eating a berserking flower while clear-minded works but
		// reduces the duration by half.
		g.RemoveStatus(PlayerID, pa, StatusClarity)
		d /= 2
	}
	g.eatMsg("berserking flower", "")
	if pa.Has(StatusBerserk) {
		// The berserking flower can be used to prolong berserk
		// duration, but you'll still get the Poison debuff.
		pa.Statuses[StatusBerserk] = 0
		g.AdjustHP(PlayerID, pa, -pa.HPBonus())
		g.PutStatus1(PlayerID, pa, StatusPoison, DurationPoisonBerserk)
	}
	g.PutStatus1(PlayerID, pa, StatusBerserk, d)
	return true
}

func (eff EffectBerserkingFlower) Desc() string {
	return fmt.Sprintf("Makes you @BBerserk@N for %d turns. %s",
		DurationBerserk, statusDesc[StatusBerserk])
}

type EffectLignificationFruit struct{}

func (eff EffectLignificationFruit) Apply(g *Game) bool {
	pa := g.PlayerActor()
	d := DurationLignification
	if pa.Has(StatusFoggySkin) {
		// Eating a lignification fruit while foggy works but reduces
		// the duration by half.
		g.RemoveStatus(PlayerID, pa, StatusFoggySkin)
		d /= 2
	}
	g.eatMsg("lignification fruit", "")
	if pa.Has(StatusLignification) {
		// If already lignified, cancel previous effects.
		pa.Statuses[StatusLignification] = 0
		g.AdjustHP(PlayerID, pa, -pa.HPBonus())
	}
	g.PutStatus1(PlayerID, pa, StatusLignification, d)
	return true
}

func (eff EffectLignificationFruit) Desc() string {
	return fmt.Sprintf("@BLignifies@N you for %d turns. %s",
		DurationLignification, statusDesc[StatusLignification])
}

type EffectClarityLeaves struct{}

func (eff EffectClarityLeaves) Apply(g *Game) bool {
	pa := g.PlayerActor()
	var hpstr string
	if pa.Has(StatusBerserk) {
		hpstr = newHPWithoutBonus(pa, 1, HealClarityLeaves)
	} else {
		hpstr = newHP(pa, HealClarityLeaves)
	}
	g.eatMsg("clarity leaves", hpstr)
	g.PutStatus1(PlayerID, pa, StatusClarity, DurationClarity)
	g.AdjustHP(PlayerID, pa, HealClarityLeaves)
	return true
}

func newHPWithoutBonus(a *Actor, m, hp int) string {
	hpbonus := a.HPBonus()
	var nhp int
	if a.HP < m {
		nhp = a.HP
	} else {
		nhp = max(m, a.HP-hpbonus)
	}
	return fmt.Sprintf("%d/%d", nhp+hp, a.GetMaxHP()-hpbonus)
}

func (eff EffectClarityLeaves) Desc() string {
	return fmt.Sprintf("Gives @BClarity@N for %d turns. %s Restores %+d HP.",
		DurationClarity, statusDesc[StatusClarity], HealClarityLeaves)
}

type EffectFoggySkinOnion struct{ Beam bool }

func (eff EffectFoggySkinOnion) Apply(g *Game) bool {
	pa := g.PlayerActor()
	var hpstr string
	if pa.Has(StatusLignification) {
		hpstr = newHPWithoutBonus(pa, 3, HealFoggySkinOnion)
	} else {
		hpstr = newHP(pa, HealFoggySkinOnion)
	}
	g.eatMsg("foggy-skin onion", hpstr)
	g.PutStatus1(PlayerID, pa, StatusFoggySkin, DurationFoggySkin)
	g.AdjustHP(PlayerID, pa, HealFoggySkinOnion)
	return true
}

func (eff EffectFoggySkinOnion) Desc() string {
	return fmt.Sprintf("Gives @BFoggy-Skin@N for %d turns. %s Restores %+d HP.",
		DurationFoggySkin, statusDesc[StatusFoggySkin], HealFoggySkinOnion)
}

type EffectFirebreathPepper struct{ Beam bool }

func (eff EffectFirebreathPepper) Apply(g *Game) bool {
	pp := g.PP()
	pa := g.PlayerActor()
	g.eatMsg("firebreath pepper", newHP(pa, HealFirebreathPepper))
	g.MakeNoise(pp, NoiseRoar)
	g.AdjustHP(PlayerID, pa, HealFirebreathPepper)
	g.md.FirebreathAnimation(pp.Add(g.Dir), g.Dir)
	p := pp
	for range MaxFOVRange {
		p = p.Add(g.Dir)
		t := g.Map.Terrain.At(p)
		if !Passable(t) {
			g.ExplosionAt(PlayerID, p)
			return true
		}
		if t == Foliage {
			g.FireCloudAt(p)
		}
		if j, aj := g.ActorAt(p); j >= 0 {
			g.PutStatus(j, aj, StatusConfusion, DurationConfusionPepperBreath)
			g.PutStatus(j, aj, StatusFire, DurationFire)
		}
	}
	// Leave cloud at the last tile if passable.
	g.NormalCloudAt(p, 2+g.IntN(3))
	return true
}

func (eff EffectFirebreathPepper) Desc() string {
	return fmt.Sprintf("Makes you breathe a confusing fire beam in the current direction inflicting @OConfusion@N for %d turns and @OFire@N for %d turns. Burns foliage, producing fire clouds lasting 6-12 turns. Makes walls explode. Restores %+d HP.",
		DurationConfusionPepperBreath, DurationFire, HealFirebreathPepper)
}

type EffectTeleportMushroom struct{}

func (eff EffectTeleportMushroom) Apply(g *Game) bool {
	pa := g.PlayerActor()
	j := g.TotemID()
	if pa.Has(StatusLignification) {
		if j < 0 || g.Entity(j).Seen {
			g.Log("Eating that while lignified is useless now.")
			return false
		}
	}
	g.eatMsg("teleport mushroom", "")
	if !g.TeleportActor(PlayerID, pa, 1) {
		g.Log("You resist teleport while lignified.")
	}
	if j >= 0 {
		ej := g.Entity(j)
		g.SenseEntity(ej, "sense")
	}
	return true
}

func (g *Game) teleportStatuses(i ID, ai *Actor, n int) {
	if g.IntN(2) == 0 {
		g.PutStatusN(i, ai, StatusDaze, DurationDazeTeleport, n)
	} else {
		g.PutStatusN(i, ai, StatusImbalance, DurationImbalanceTeleport, n)
	}
}

func (eff EffectTeleportMushroom) Desc() string {
	return fmt.Sprintf("Teleports you away and makes you sense any totemic spirit. Leaves you either @Odazed@N or @Oimbalanced@N for %d or %d turns respectively.",
		DurationDazeTeleport, DurationImbalanceTeleport)
}

type EffectFocus struct{}

func (eff EffectFocus) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("Using focus while dazed would be useless.")
		return false
	}
	if pa.Has(StatusFocus) {
		g.Log("You’re already focusing your heads.")
		return false
	}
	if pa.Has(StatusFear) {
		g.Log("You cannot focus while afraid.")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusFocus, DurationFocus)
	g.instant = true
	return true
}

func (eff EffectFocus) Name() string {
	return "focus"
}

func (eff EffectFocus) Desc() string {
	return fmt.Sprintf("@BFocus@N for %d turns. %s @GInstant@N.", DurationFocus, statusDesc[StatusFocus])
}

type EffectDig struct{}

func (eff EffectDig) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("Using dig while dazed would be useless.")
		return false
	}
	if pa.Has(StatusDig) {
		g.Log("You’re already on a destructive charge.")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusDig, DurationDig)
	g.instant = true
	return true
}

func (eff EffectDig) Name() string {
	return "dig"
}

func (eff EffectDig) Desc() string {
	return fmt.Sprintf("@BDig@N for %d turns. %s @GInstant@N.",
		DurationDig, statusDesc[StatusDig])
}

type EffectJump struct{}

func (eff EffectJump) Apply(g *Game) bool {
	pl, pa := g.Player()
	if !pa.CanMove() {
		g.Log("You cannot jump while lignified.")
		return true
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("You cannot jump while poisoned and almost dead.")
		return false
	}
	dir, from, to := g.Dir, pl.P, pl.P
	for p := to.Add(dir); g.Map.Passable(p) && g.InFOV(p); p = p.Add(dir) {
		to = p
	}
	for to != from {
		if i, _ := g.ActorAt(to); i < 0 {
			break
		}
		to = to.Sub(dir)
	}
	if to == from {
		g.Log("There’s no room to jump in that direction.")
		return false
	}
	g.Log("You jump.")
	g.BumpMoveActor(PlayerID, pa, to)
	g.md.MoveAnimation(from, to)
	for p := from.Add(dir); p != to; p = p.Add(dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			if !pa.Has(StatusFear) && !pa.Has(StatusDaze) {
				if pa.Has(StatusSprint) {
					g.PutStatus(i, ai, StatusImbalance, DurationImbalanceSprint)
				} else {
					g.BumpAttackActor(PlayerID, i, pa, ai, AttackPlain, 1, 1)
				}
			}
			g.PutStatus(i, ai, StatusDaze, DurationDazeJump)
		}
	}
	if pa.Has(StatusImbalance) && g.IntN(2) == 0 {
		g.Log("You fall due to imbalance.")
		g.PutStatus1(PlayerID, pa, StatusDaze, DurationDazeFall)
	}
	return true
}

func (eff EffectJump) Name() string {
	return "dazing jump"
}

func (eff EffectJump) Desc() string {
	return fmt.Sprintf("Jump in the current direction to the furthest free position in view. You will attack any monsters on the path with +1 attack and @Odaze@N them for %d turns.",
		DurationDazeJump)
}

type EffectPushingGale struct{}

func (eff EffectPushingGale) Apply(g *Game) bool {
	pp := g.PP()
	g.Log("A strong gale pushes the air away from you.")
	g.MakeNoise(pp, NoiseWind)
	for _, dir := range Directions {
		g.pushingGale(pp, dir)
	}
	return true
}

func (g *Game) pushingGale(from, dir gruid.Point) {
	const maxdist = MaxFOVRange
	at := from
	for n, p := 0, at.Add(dir); n < maxdist && g.Map.Passable(p); p = p.Add(dir) {
		g.blowClouds(p, dir)
		at = p
		n++
	}
	for ; at != from; at = at.Sub(dir) {
		i, ai := g.ActorAt(at)
		if i < 0 {
			continue
		}
		ei := g.Entity(i)
		g.InflictDamage(i, ai, 1, AttackGale)
		if ai.IsDead() {
			continue
		}
		g.MonsterUpdateTarget(i, ai, g.PP()) // alert monster toward wind origin
		g.PutStatus(i, ai, StatusImbalance, DurationImbalanceGale)
		to := at.Add(dir)
		n := 0
		for p := to; g.IsFree(p) && n < 2; p = p.Add(dir) {
			to = p
			n++
		}
		if g.IsFree(to) {
			from := ei.P
			g.MoveActor(i, ai, to)
			g.md.MoveAnimation(from, to)
		}
	}
}

func (g *Game) blowClouds(p, dir gruid.Point) {
	if g.Map.Terrain.At(p) != Foliage || g.Map.Clouds.At(p).Kind != CloudFire {
		// The strong gale dissipates the cloud (except if it's a fire
		// cloud on a dense foliage cell).
		g.RemoveCloudAt(p)
		return
	}
	if p = p.Add(dir); g.Map.Burnable(p) && !g.FireAt(p) {
		// Make fire on foliage propagate in the given direction.
		g.FireCloudAt(p)
	}
}

func (eff EffectPushingGale) Name() string {
	return "pushing gale"
}

func (eff EffectPushingGale) Desc() string {
	return fmt.Sprintf("Hurt foes in the four cardinal directions for 1 damage, pushing them away and @Ounbalancing@N them for %d turns. Dissipates clouds except foliage fires that propagate instead.",
		DurationImbalanceGale)
}

type EffectTimeStop struct{}

func (eff EffectTimeStop) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusTimeStop) {
		g.Log("You already froze time.")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusTimeStop, DurationTimeStop)
	g.instant = true
	return true
}

func (eff EffectTimeStop) Name() string {
	return "stop time"
}

func (eff EffectTimeStop) Desc() string {
	return fmt.Sprintf("@BStop Time@N for %d turns. %s @GInstant@N.",
		DurationTimeStop, statusDesc[StatusTimeStop])
}

type EffectLightning struct{}

func (eff EffectLightning) Apply(g *Game) bool {
	const (
		maxdist = 3
		selfDmg = 1
	)
	pa := g.PlayerActor()
	if pa.HP <= 1 {
		g.Log("You’re too injured for using lightning.")
		return false
	}
	mp := &MapPath{passable: func(q gruid.Point) bool {
		if !inMap(q) {
			return false
		}
		i, _ := g.ActorAt(q)
		return i >= 0
	}}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	done := false
	lightning := func() {
		g.md.LightningAnimation(nodes)
		g.Logf("You emit a static discharge (-%d HP).", selfDmg)
		g.MakeNoise(g.PP(), NoiseLightning)
		g.InflictDamage(PlayerID, pa, selfDmg, AttackOther)
		done = true
	}
	for _, n := range nodes {
		i, ai := g.ActorAt(n.P)
		if i < 0 || i == PlayerID {
			continue
		}
		if !done {
			lightning()
		}
		dmg := maxdist + 1 - n.Cost
		if g.InflictDamage(i, ai, dmg, AttackLightning) > 0 {
			g.PutStatus(i, ai, StatusDaze, DurationDazeLightning)
		}
	}
	if !done && pa.Has(StatusDaze) && pa.DoesAny(ResistanceDaze) {
		lightning()
	}
	if !done {
		g.Log("No suitable targets for lightning.")
	}
	return done
}

func (eff EffectLightning) Name() string {
	return "lightning"
}

func (eff EffectLightning) Desc() string {
	return fmt.Sprintf("Inflict 1-3 damage on any chain of monsters adjacent to you. Damage decreases with distance. Affected monsters are @Odazed@N for %d turns. Hurts you too for 1 HP.",
		DurationDazeLightning)
}

type EffectBark struct{}

func (eff EffectBark) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	mp := &MapPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	g.Log("You bark.")
	pp := g.PP()
	g.md.NoiseAnimation(pp)
	g.MakeNoise(pp, NoiseBark)
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearBark)
		}
	}
	return true
}

func (eff EffectBark) Name() string {
	return "bark"
}

func (eff EffectBark) Desc() string {
	return fmt.Sprintf("Frighten all monsters in a %d-radius distance for %d turns.",
		MaxFOVRange, DurationFearBark)
}

type EffectNoxiousSmell struct{}

func (eff EffectNoxiousSmell) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	g.Log("You emit a confusing noxious smell.")
	g.confusingSmell(PlayerID, maxdist, DurationConfusionSkunk)
	return true
}

func (g *Game) confusingSmell(i ID, maxdist, duration int) {
	p := g.Entity(i).P
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{p}, maxdist)
	g.md.FloodAnimation(nodes, ColorGreen)
	for j, aj := range g.Actors() {
		if j == i {
			continue
		}
		ej := g.Entity(j)
		if g.PR.BreadthFirstMapAt(ej.P) <= maxdist {
			g.PutStatus(j, aj, StatusConfusion, duration)
		}
	}
}

func (eff EffectNoxiousSmell) Name() string {
	return "noxious smell"
}

func (eff EffectNoxiousSmell) Desc() string {
	return fmt.Sprintf("Confuse monsters in a %d-radius distance for %d turns.",
		MaxFOVRange, DurationConfusionSkunk)
}

type EffectLignify struct{}

func (eff EffectLignify) Apply(g *Game) bool {
	done := false
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		if ai.DoesAny(MonsImmunityLignification) || ai.Has(StatusLignification) {
			// Double-check before so that the "Growing roots"
			// message appears first and only if the status will
			// apply.
			continue
		}
		if !done {
			g.Log("Growing roots lignify your foes.")
			done = true
		}
		g.PutStatus(i, ai, StatusLignification, DurationLignificationRoots)
	}
	if done {
		mp := &MapPath{passable: inMap}
		nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, MaxFOVRange)
		g.md.FloodAnimation(nodes, ColorYellow)
	} else {
		g.Log("No lignifiable foes in view.")
	}
	return done
}

func (eff EffectLignify) Name() string {
	return "lignify"
}

func (eff EffectLignify) Desc() string {
	return fmt.Sprintf("Lignify monsters in view for %d turns.",
		DurationLignificationRoots)
}

type EffectPoisonCloud struct{}

func (eff EffectPoisonCloud) Apply(g *Game) bool {
	p := g.PP()
	behind := p.Sub(g.Dir)
	done := false
	for q := range g.Map.PassableNeighbors(p) {
		if q == behind {
			continue
		}
		nq := q.Add(q.Sub(p))
		if g.IsFree(q) && g.IsFree(nq) {
			q = nq
		}
		done = true
		g.PoisonCloudAt(q)
	}
	if done {
		g.Log("You spit poisonous fumes around.")
	} else {
		g.Log("There’s no room to spit poison.")
	}
	return done
}

func (eff EffectPoisonCloud) Name() string {
	return "poison cloud"
}

func (eff EffectPoisonCloud) Desc() string {
	return "Spit poisonous clouds on the sides and in front of you up to two free tiles away, lasting 7-11 turns."
}

type EffectSprint struct{}

func (eff EffectSprint) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("Using sprint while dazed would be useless.")
		return false
	}
	if pa.Has(StatusSprint) {
		g.Log("You’re already sprinting.")
		return false
	}
	if !pa.CanMove() {
		g.Log("You cannot sprint while lignified.")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusSprint, DurationSprint)
	g.instant = true
	return true
}

func (eff EffectSprint) Name() string {
	return "sprint"
}

func (eff EffectSprint) Desc() string {
	return fmt.Sprintf("@BSprint@N for %d turns. %s @GInstant@N.",
		DurationSprint, statusDesc[StatusSprint])
}

type EffectFireRetreat struct{}

func (eff EffectFireRetreat) Apply(g *Game) bool {
	pl, pa := g.Player()
	if !pa.CanMove() {
		g.Log("You cannot retreat while lignified.")
		return false
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("You cannot retreat while poisoned and almost dead.")
		return false
	}
	from := pl.P
	p, to := from, from
	for range 2 {
		p = p.Sub(g.Dir)
		if !g.InFOV(p) {
			break
		}
		if !g.Map.Passable(p) && !g.Dig(PlayerID, pa, p) {
			break
		}
		if i, _ := g.ActorAt(p); i >= 0 {
			break
		}
		to = p
	}
	if from == to {
		g.Log("There’s no room to retreat.")
		return false
	}
	g.Log("You retreat.")
	g.BumpMoveActor(PlayerID, pa, to)
	g.md.MoveAnimation(from, to)
	g.FireCloudAt(from)
	return true
}

func (eff EffectFireRetreat) Name() string {
	return "fire retreat"
}

func (eff EffectFireRetreat) Desc() string {
	return "Walk backwards up to two tiles while leaving a fire cloud lasting 6-12 turns."
}

func (g *Game) menhirMsg(name string) {
	g.Logf("You activate the %s.", name)
}

type EffectEarthMenhir struct{}

func (eff EffectEarthMenhir) Apply(g *Game) bool {
	const maxdist = MaxFOVRange + 4 // "earthquake" distance
	g.menhirMsg("earth menhir")
	// Reveal walls (but not the translucent ones) and destroy nearby ones.
	g.MakeNoise(g.PP(), NoiseEarthMenhir)
	mp := &MappingPath{passable: g.Map.NoWallAt}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, unreachable)
	var draw bool
	var nbs paths.Neighbors
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Wall {
			continue
		}
		if n.Cost <= maxdist && g.Map.PassableNeighbor(n.P) {
			g.earthMenhirDig(nbs, n.P)
			if g.Map.KnownTerrain.At(n.P) == Wall {
				g.Map.KnownTerrain.Set(n.P, Rubble)
			} else if g.Map.revelableWallAt(n.P) {
				g.Map.KnownTerrain.Set(n.P, t)
			}
			draw = true
		} else if g.Map.KnownTerrain.At(n.P) != t && g.Map.revelableWallAt(n.P) {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
	return true
}

// revelableWallAt reports whether there's a earth-menhir revelable wall at the
// given position (wall adjacent to an interior wall).
func (m *Map) revelableWallAt(p gruid.Point) bool {
	if m.Terrain.At(p) != Wall {
		return false
	}
	for q := range Neighbors(p) {
		if m.interiorWallAt(q) {
			return true
		}
	}
	return false
}

// interiorWallAt reports whether there's an interior wall at the given
// position (wall not connected to any passable tile).
func (m *Map) interiorWallAt(p gruid.Point) bool {
	if m.Terrain.At(p) != Wall {
		return false
	}
	for q := range Neighbors(p) {
		if m.Terrain.At(q) != Wall {
			return false
		}
	}
	return true
}

// earthMenhirDig is similar to dig but without the noise and extra behaviors
// (because we want a single noise source at the "earthquake" centre).
func (g *Game) earthMenhirDig(nbs paths.Neighbors, at gruid.Point) {
	g.Stats.Digs++
	g.Map.Terrain.Set(at, Rubble)
	ps := append(nbs.Cardinal(at, inMap), at)
	for _, q := range ps {
		g.NormalCloudAt(q, 2+g.IntN(3))
	}
}

func (eff EffectEarthMenhir) Desc() string {
	return "Emits a noisy echoing sound that disintegrates nearby walls and reveals the frontier formed by map walls adjacent to interior ones."
}

type EffectWarpingMenhir struct{}

func (eff EffectWarpingMenhir) Apply(g *Game) bool {
	g.menhirMsg("warping menhir")
	// Teleport nearby monsters away.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	for _, n := range nodes {
		if n.P == pp {
			continue
		}
		if i, ai := g.ActorAt(n.P); i > 0 {
			g.TeleportActor(i, ai, 0)
		}
	}
	// Reveal portal or orb of corruption.
	for _, ei := range g.NPMapEntities() {
		switch ei.Role.(type) {
		case *Portal:
			g.SenseEntity(ei, "sense")
		case *CorruptionOrb:
			g.SenseEntity(ei, "sense")
		}
	}
	if g.IntN(2+(MapLevels-g.Map.Level)/2) == 0 {
		// Dungeon sends a warping wraith to investigate: it first goes
		// to the menhir, then back to the portal.
		q := g.Map.Orb
		if q == InvalidPos {
			q = g.Map.Portal
		}
		_, a := g.genMonsterGuardian(WarpingWraith, q)
		a.Behavior.Target = pp
	}
	return true
}

func (eff EffectWarpingMenhir) Desc() string {
	return fmt.Sprintf("Teleports away monsters in a %d-radius distance. Reveals location of any portals. The dungeon might sense you and send a wraith to investigate: the deeper you are, the more likely.", MaxFOVRange)
}

type EffectPoisonMenhir struct{}

func (eff EffectPoisonMenhir) Apply(g *Game) bool {
	g.menhirMsg("poison menhir")
	// Spawn poison clouds randomly.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes[:min(max(3, len(nodes)/4), len(nodes))] {
		if n.P == pp {
			continue
		}
		g.PoisonCloudAt(n.P)
	}
	// Reveal translucent walls (which contain poison gas).
	for p, t := range g.Map.Terrain.All() {
		if t == TranslucentWall {
			g.Map.KnownTerrain.Set(p, t)
		} else if g.Map.KnownTerrain.At(p) == TranslucentWall {
			g.Map.KnownTerrain.Set(p, t)
		}
	}
	return true
}

func (eff EffectPoisonMenhir) Desc() string {
	return fmt.Sprintf("Makes many poisonous clouds appear in a %d-radius distance. Reveals translucent walls.", MaxFOVRange)
}

type EffectFireMenhir struct{}

func (eff EffectFireMenhir) Apply(g *Game) bool {
	g.menhirMsg("fire menhir")
	// Spawn fire clouds randomly.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes[:min(max(3, len(nodes)/3), len(nodes))] {
		if n.P == pp {
			continue
		}
		g.FireCloudAt(n.P)
	}
	// Reveal foliage.
	// Extra FOV update first to avoid info leak during animation. Just FOV
	// is enough, because it becomes strictly smaller.
	g.UpdateFOV()
	g.revealFoliage()
	return true
}

func (g *Game) revealFoliage() {
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, unreachable)
	var draw bool
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Foliage {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != t && g.IntN(5) < 2 {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
}

func (eff EffectFireMenhir) Desc() string {
	return fmt.Sprintf("Makes many fire clouds appear in a %d-radius distance, each one lasting 6-12 turns. Reveals foliage partially.", MaxFOVRange)
}
