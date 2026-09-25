// This file implements conditional comestible effects (with or without
// non-conditional variant).

package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Conditional comestible effect, only usable if conditions are met.
type ConditionalComestible interface {
	CanApply(*Game, ID) bool
}

type EffectPolymorphFungus struct{}

func (eff EffectPolymorphFungus) CanApply(g *Game, i ID) bool {
	var n int
	for i := range g.Monsters() {
		ei := g.Entity(i)
		if g.InFOV(ei.P) {
			n++
		}
	}
	return n == 2
}

func (eff EffectPolymorphFungus) Apply(g *Game) bool {
	g.eatMsg(PolymorphFungus.Name(), "")
	// Collect IDs from monsters in view.
	var mids []ID
	for i := range g.Monsters() {
		if g.InFOV(g.Entity(i).P) {
			mids = append(mids, i)
		}
	}
	if len(mids) == 0 {
		g.PutStatus1(PlayerID, g.PlayerActor(), StatusVampirism, DurationVampFung)
		return true
	}
	mks := []ActorKind{BlinkButterfly, ConfusingEye}
	p := g.rand.Perm(len(mids))[:min(len(mids), 2)]
	g.md.PolymorphAnimation(mids, p)
	for i, pi := range p {
		ei := g.Entity(mids[pi])
		ai := ei.Role.(*Actor)
		mki := mks[i]
		e := monster(mki)
		a := e.Role.(*Actor)
		g.Logf("The %s becomes %s.", ei.Name, One(e.Name))
		if ai.DoesAny(MonsNotable) {
			g.StoryLogf("Polymorphed %s into %s", One(ei.Name), One(e.Name))
		}
		// Update name and rune only in entity.
		ei.Name = e.Name
		ei.Rune = e.Rune
		// Update actor stats and traits, but keep behavior, statuses,
		// and knowledge info.
		ai.MaxHP = a.MaxHP
		ai.HP = min(a.HP, ai.MaxHP)
		ai.Defense = a.Defense
		ai.Attack = a.Attack
		ai.Kind = a.Kind
		ai.Traits = a.Traits
		switch mki {
		case BlinkButterfly:
			ai.Statuses[StatusLignification] = 0
			ai.Statuses[StatusImbalance] = 0
		case ConfusingEye:
			ai.Statuses[StatusConfusion] = 0
		}
	}
	g.md.PolymorphAnimation(mids, p)
	return true
}

func (eff EffectPolymorphFungus) Desc(g *Game) string {
	if g.HasCondComestibles() {
		return "Eat only if there are exactly two monsters in view. Polymorphs them into a blinking butterfly and a confusing eye."
	}
	return fmt.Sprintf("Polymorphs a random monster in view into a blinking butterfly, and a second monster, if any, into a confusing eye.\nIf no monsters are in view, gives @BVampirism@N for %d turns instead, preparing you for a vampiric ambush.", DurationVampFung)
}

type EffectRunethornRose struct{}

func (g *Game) anyRuneAt(p gruid.Point) (*RunicTrap, bool) {
	if i, it := g.ItemAt(p); i >= 0 {
		rt, ok := it.(*RunicTrap)
		return rt, ok
	}
	return nil, false
}

func (eff EffectRunethornRose) CanApply(g *Game, i ID) bool {
	_, ok := g.anyRuneAt(g.PP())
	return ok
}

func (eff EffectRunethornRose) Apply(g *Game) bool {
	g.eatMsg(RunethornRose.Name(), "")
	pp := g.PP()
	var t MagicRune // type of rune
	if r, ok := g.anyRuneAt(pp); ok {
		t = r.Rune
	} else {
		t = MagicRune(g.IntN(NRunes))
	}
	size := g.Map.Terrain.Size()
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			q := pp
			for range 2 {
				q = q.Shift(i, j)
				p := q
				for !g.validRunePlace(p) || p == pp {
					p = gruid.Point{g.rand.IntN(size.X), g.rand.IntN(size.Y)}
				}
				i, _ := g.genRunicTrapAt(t, p)
				g.SenseEntity(i, "sense")
				// Trigger the trap if there's already a
				// monster on top of it.
				if k, ak := g.ActorAt(p); k >= 0 {
					g.TriggerTrap(k, ak)
				}
			}
		}
	}
	return true
}

func (g *Game) validRunePlace(p gruid.Point) bool {
	m := g.Map
	if m.Terrain.At(p) != Floor || m.RuneCache.At(p) > 0 {
		return false
	}
	i, _ := g.ItemAt(p)
	return i < 0
}

func (eff EffectRunethornRose) Desc(g *Game) string {
	if !g.HasCondComestibles() {
		return "Creates 8 runes of same type: up to two at each diagonal on floor tiles, or randomly around the map if there’s no room.\nThe rune type is chosen randomly unless you are standing on a trap, in which case its rune type will be used."
	}
	return "Eat only if you are standing on a runic trap. Creates 8 runes of its type: up to two at each diagonal on floor tiles, or randomly around the map if there’s no room."
}

type EffectTunnelingAcorn struct{}

func (eff EffectTunnelingAcorn) CanApply(g *Game, i ID) bool {
	pp := g.PP()
	n := pp.Add(g.Dir)
	return inMap(n) && !g.Map.Passable(n) && g.Map.AdjacentNonPassableCount(pp) >= 3
}

func (eff EffectTunnelingAcorn) Apply(g *Game) bool {
	g.eatMsg(TunnelingAcorn.Name(), "")
	p := g.PP().Add(g.Dir)
	from := p
	for g.Map.Passable(p) {
		// Skip over floor tiles for non-conditional variant, and
		// reveal terrain on the path.
		p = p.Add(g.Dir)
		g.SeeTerrain(p)
		if i, _ := g.ItemAt(p); i >= 0 {
			g.SenseEntity(i, "sense")
		}
	}
	for inMap(p) && !g.Map.Passable(p) {
		g.disintegrateWall(p)
		p = p.Add(g.Dir)
	}
	if from != p {
		g.md.BeamAnimation(from, p)
	}
	return true
}

func (g *Game) disintegrateWall(at gruid.Point) {
	g.Stats.Digs++
	TotemEvents(g, &EventHappened{EvType: EventWallBreak})
	g.Map.Terrain.Set(at, Floor)
	g.Map.KnownTerrain.Set(at, Floor) // reveal tunnel
	g.NormalCloudAt(at, 10+g.IntN(30))
}

func (eff EffectTunnelingAcorn) Desc(g *Game) string {
	if !g.HasCondComestibles() {
		return "Makes you emit a beam that silently disintegrates any walls in front of you, until reaching a non-wall. Reveals terrain on the path."
	}
	return "Eat only if you are next to 3 walls, and facing one of them. Makes you silently disintegrate all walls in front of you, until reaching a non-wall."
}

type EffectVanishingSnail struct{}

func (eff EffectVanishingSnail) CanApply(g *Game, i ID) bool {
	count := 0
	for _, a := range g.Monsters() {
		if a.Behavior.State == Hunting {
			count++
		}
	}
	return count >= 4
}

func (eff EffectVanishingSnail) Apply(g *Game) bool {
	g.eatMsg(VanishingSnail.Name(), "")
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if ai.Behavior.State == Hunting {
			// Make monster lose track of you and wander again to a
			// nearby location, as if hit by a player with Shadow.
			ai.Behavior.State = Wandering
			ai.Behavior.Target = g.RandomPassableWithin(ei.P, MaxFOVRange)
		}
	}
	pa := g.PlayerActor()
	for p := range g.Map.PassableNeighbors(g.PP()) {
		if k, ak := g.ActorAt(p); k >= 0 && !ak.ResistsMove() {
			ek := g.Entity(k)
			if to, ok := g.BlinkPos(ek.P); ok {
				g.MoveActor(k, ak, to, MovTeleport)
			}
		}
	}
	g.PutStatus1(PlayerID, pa, StatusShadow, DurationShadow)
	return true
}

func (eff EffectVanishingSnail) Desc(g *Game) string {
	cond := "Eat only if at least 4 monsters are hunting you, including any out of view ones. "
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + fmt.Sprintf(
		"Makes all monsters lose track of you and blinks away any that are next to you. Gives @BShadow@N for %d turns.",
		DurationShadow)
}

type EffectWarpingBean struct{}

func (eff EffectWarpingBean) CanApply(g *Game, i ID) bool {
	s := g.PlayerActor().Statuses
	return s.Has(StatusLignification) || s.Has(StatusDaze) || s.Has(StatusPoison)
}

func (eff EffectWarpingBean) Apply(g *Game) bool {
	if GluttonyRework {
		// In case the bear teleported or warped first.
		g.UpdateFOV()
		g.UpdateKnowledge()
	}
	pa := g.PlayerActor()
	if pa.Has(StatusBerserk) {
		pa.HP = max(1, pa.HP-pa.HPBonus())
	}
	if pa.Has(StatusLignification) {
		pa.HP = max(min(3, pa.HP), pa.HP-pa.HPBonus())
	}
	for i := range pa.Statuses {
		pa.Statuses[i] = 0
	}
	var hpstr string
	if !g.Mod(ModHealingCombat) {
		hpstr = newHPMsg(pa, HealWarpingBean)
	}
	g.eatMsg(WarpingBean.Name(), hpstr)
	if !g.Mod(ModHealingCombat) {
		g.healPlayer(pa, HealFirebreathPepper)
	}
	pp := g.PP()
	ps := g.BlinkPosSlice(pp)
	to := pp
	for range 6 {
		p := ps[g.IntN(len(ps))]
		if paths.DistanceManhattan(to, pp) < paths.DistanceManhattan(p, pp) {
			to = p
		}
	}
	if to != pp {
		g.MoveActor(PlayerID, pa, to, MovTeleport)
	}
	return true
}

func (eff EffectWarpingBean) Desc(g *Game) string {
	cond := "Eat only if you are lignified, dazed, or poisoned. "
	if !g.HasCondComestibles() {
		cond = ""
	}
	return fmt.Sprintf("%sRemoves all your statuses without the usual after-effects, then warps you away in view with greater chance to distant tiles.%s",
		cond, g.healMsg(HealWarpingBean))
}

type EffectFortressOyster struct{}

func (eff EffectFortressOyster) CanApply(g *Game, i ID) bool {
	count := 0
	for range g.AdjacentActors(g.PP()) {
		count++
	}
	return count >= 3
}

func (eff EffectFortressOyster) Apply(g *Game) bool {
	const d = 5 // common status duration
	pa := g.PlayerActor()
	g.eatMsg(FortressOyster.Name(), "")
	g.RemoveStatus(PlayerID, pa, StatusClarity)
	if pa.Has(StatusBerserk) {
		pa.Statuses[StatusBerserk] = max(pa.Statuses[StatusBerserk], d+1)
		bonusHP := pa.HP - max(1, pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
	} else {
		g.PutStatus1(PlayerID, pa, StatusBerserk, d)
	}
	g.RemoveStatus(PlayerID, pa, StatusFoggySkin)
	if pa.Has(StatusLignification) {
		// We handle lignification after berserk, because at low HP it
		// can result in higher final bonuses due to the HP floor of 3.
		pa.Statuses[StatusLignification] = max(pa.Statuses[StatusLignification], d+1)
		bonusHP := pa.HP - max(min(3, pa.HP), pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
	} else {
		g.PutStatus1(PlayerID, pa, StatusLignification, d)
	}
	g.PutStatus1(PlayerID, pa, StatusVampirism, d)
	return true
}

func (eff EffectFortressOyster) Desc(g *Game) string {
	cond := "Eat only if you are next to at least 3 monsters. "
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + "Gives @BBerserk@N, @BLignification@N, and @BVampirism@N for 5 turns. Removes incompatible statuses. Renews duration and temporary HP bonuses if applicable."
}

type EffectSeeingPotato struct{}

func (eff EffectSeeingPotato) CanApply(g *Game, i ID) bool {
	j, _ := g.ItemAt(g.PP())
	return i >= 0 && i == j
}

func (eff EffectSeeingPotato) Apply(g *Game) bool {
	g.eatMsg("seeing potato’s stalk", "")
	const maxdist = 2 * MaxFOVRange
	mp := &MappingPath{passable: func(p gruid.Point) bool { return inMap(p) }}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	var draw bool
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Floor {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != t {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
	pp := g.PP()
	for i, ei := range g.NPMapEntities() {
		p := ei.P
		if paths.DistanceManhattan(pp, p) <= maxdist && g.Map.Terrain.At(p) == Floor {
			g.SenseEntity(i, "sense")
		}
	}
	return true
}

func (eff EffectSeeingPotato) Desc(g *Game) string {
	cond := "Eat only from the ground and not your inventory. "
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + "Makes you sense nearby floor tiles and any entities on them (2× view range)."
}

type EffectTrickyChestnut struct{}

func (eff EffectTrickyChestnut) CanApply(g *Game, i ID) bool {
	if i, it := g.ItemAt(g.PP()); i >= 0 {
		switch o := it.(type) {
		case *Portal:
			return o.Used
		case *Spirit:
			return true
		}
	}
	return false
}

func (eff EffectTrickyChestnut) Apply(g *Game) bool {
	g.eatMsg(TrickyChestnut.Name(), "")
	pp := g.PP()
	var tot *Spirit
	if i, it := g.ItemAt(pp); i >= 0 {
		switch o := it.(type) {
		case *Portal:
			o.Fake = false
			o.Used = false
			g.LogStyled("The portal is fixed!", logSpecial)
			return true
		case *Spirit:
			tot = o
		}
	}
	// Recharge spirits by default.
	g.LogStyled("Your spirits are recharged!", logSpecial)
	if tot != nil { // turn totem to empty if present
		g.Entity(g.TotemID()).P = InvalidPos
		et := g.Entity(g.genEmptyTotemAt(pp))
		et.Seen = true
		for _, sp := range g.PlayerSpirits() {
			if g.Mod(ModNoRecharges) {
				sp.Charges = min(sp.Charges+2, sp.GetMaxCharges())
			} else {
				sp.Charges = sp.GetMaxCharges()
			}
		}
	} else {
		// Bear outside of conditions.
		for _, sp := range g.PlayerSpirits() {
			sp.Charges = min(sp.Charges+1, sp.GetMaxCharges())
		}
	}
	return true
}

func (eff EffectTrickyChestnut) Desc(g *Game) string {
	nocond := !g.HasCondComestibles()
	recharges := "Eat if you are standing on a full totem, to empty the totem, and restore all your spirit charges."
	if g.Mod(ModNoRecharges) {
		recharges = "Eat if you are standing on a full totem, to empty the totem, and restore two charges per spirit."
		if nocond {
			recharges = "If you are standing on a full totem, empties it, and restores two charges per spirit."
		}
	} else if nocond {
		recharges = "If you are standing on a full totem, empties it, and restores all your spirit charges."
	}
	if nocond {
		return fmt.Sprintf("If you are standing on a broken portal, makes it work again.\n%s\nOtherwise, restores one charge per spirit.",
			recharges)
	}
	return "Eat if you are standing on a broken portal, to make it work again.\n" + recharges
}

// CantEatMsg returns a message suitable for when you try to eat a conditional
// comestible when the necessary conditions aren't fulfilled.
func CantEatMsg(ef Effect) string {
	switch ef.(type) {
	case EffectVanishingSnail, EffectWarpingBean:
		return "You can’t grab it!"
	case EffectFortressOyster, EffectTunnelingAcorn, EffectTrickyChestnut:
		return "It won’t open!"
	case EffectRunethornRose:
		return "You can’t get past its thorns!"
	case EffectSeeingPotato:
		return "It has to be planted in the ground!"
	case EffectPolymorphFungus:
		return "It can only be eaten with exactly two monsters in view!"
	default:
		panic("unexpected inedible comestible")
	}
}

// HasCondComestibles reports whether conditional comestibles are enabled.
func (g *Game) HasCondComestibles() bool {
	return g.Mod(ModTotemConditions) && !g.PlayerActor().DoesAny(Gluttony)
}
