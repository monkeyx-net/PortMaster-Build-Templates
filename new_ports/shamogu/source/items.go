package main

import (
	"fmt"
	"log"
	"strings"

	"codeberg.org/anaseto/gruid"
)

// Number of inventory items by type.
const (
	NSpirits     = 3 // number of primary and secondary spirits
	NComestibles = 5 // number of comestibles
)

// Effect represents any kind of applicable effect.
type Effect interface {
	Desc(*Game) string
	Apply(*Game) bool
}

// Ability represents a spirit ability effect.
type Ability interface {
	Effect
	Name() string // name of the ability as displayed in the inventory
}

// Item holds an item that can be used and equipped.
type Item interface {
	// Desc returns a description of the item.
	Desc(*Game) string
	// Use takes a game state and the item's id and reports whether a
	// turn-taking action was performed.
	Use(*Game, ID) bool
}

// IsComestible reports whether the item is a comestible.
func IsComestible(it Item) bool {
	_, ok := it.(*Comestible)
	return ok
}

// Spirit represent an equipable spirit.
type Spirit struct {
	Level        int        // spirit's level (in range 0-1)
	Charges      int        // number of active use charges
	MaxCharges   [2]int     // maximum number of charges per level
	Ability      [2]Ability // active ability effect when used
	BonusAttack  [2]int     // attack bonuses per level (may be zero)
	BonusDefense [2]int     // defense bonuses per level (may be zero)
	BonusTraits  [2]Traits  // bonus traits per level
	BonusHP      [2]int     // bonus hp per level (may be zero)
	Uses         int        // number of times it was used (for statistics)
	Advanced     bool       // advanced secondary spirit
	Condition    *SpiritCond
}

func (sp *Spirit) ExamineDesc(g *Game) string {
	return sp.eitherDesc(g, true)
}

func (sp *Spirit) Desc(g *Game) string {
	return sp.eitherDesc(g, false)
}

func (sp *Spirit) eitherDesc(g *Game, descUpgrade bool) string {
	var sb strings.Builder
	if descUpgrade {
		sb.WriteString("@CTotemic spirit.@N\n")
	}
	if cond := sp.Condition; cond != nil {
		if descUpgrade {
			udesc := cond.Upgrade.Desc()
			if udesc != "" {
				sb.WriteString(udesc)
				sb.WriteByte('\n')
			}
		}
		sb.WriteString(cond.Desc())
		sb.WriteByte('\n')
	}
	bonuses := []string{}
	if attack := sp.BonusAttack[sp.Level]; attack > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d Attack", attack))
	}
	if defense := sp.BonusDefense[sp.Level]; defense > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d Defense", defense))
	}
	if hp := sp.BonusHP[sp.Level]; hp > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d HP", hp))
	}
	if len(bonuses) > 0 {
		sb.WriteString("@CStats:@N")
		for i, bonus := range bonuses {
			sb.WriteByte(' ')
			sb.WriteString(bonus)
			if i < len(bonuses)-1 {
				sb.WriteString(", ")
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "@CAbility @S(%s %d/%d charges)@C:@N %s",
		sp.GetAbility().Name(), sp.Charges, sp.GetMaxCharges(), sp.GetAbility().Desc(g))
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "@CTraits:@N %s.", g.TraitDesc(Player, sp.BonusTraits[sp.Level]))
	return sb.String()
}

func (sp *Spirit) UpgradeDesc(g *Game, enew *Entity) string {
	desc := sp.Desc(g)
	var sb strings.Builder
	sb.WriteString("@MThe spirit essence can be used to upgrade this spirit.@N\n")
	hasCond := sp.Condition != nil
	if cond := enew.Role.(*Spirit).Condition; cond != nil && cond.Upgrade.Applies(hasCond) {
		if hasCond {
			sb.WriteString("@RReplacing@N ")
		}
		sb.WriteString(cond.Desc())
		sb.WriteByte('\n')
	}
	bonuses := []string{}
	if attack := sp.BonusAttack[1] - sp.BonusAttack[0]; attack > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d Attack", attack))
	}
	if defense := sp.BonusDefense[1] - sp.BonusDefense[0]; defense > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d Defense", defense))
	}
	if hp := sp.BonusHP[1] - sp.BonusHP[0]; hp > 0 {
		bonuses = append(bonuses, fmt.Sprintf("%+d HP", hp))
	}
	if len(bonuses) > 0 {
		sb.WriteString("@BStats:@N")
		for i, bonus := range bonuses {
			sb.WriteByte(' ')
			sb.WriteString(bonus)
			if i < len(bonuses)-1 {
				sb.WriteString(", ")
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "@BCharges:@N %+d", sp.MaxCharges[1]-sp.MaxCharges[0])
	if sp.BonusTraits[1] != sp.BonusTraits[0] {
		sb.WriteByte('\n')
		fmt.Fprintf(&sb, "@BNew traits:@N %s.", g.TraitDesc(Player, sp.BonusTraits[1]&^sp.BonusTraits[0]))
	}
	return desc + "\n\n" + sb.String()
}

func (sp *Spirit) Use(g *Game, i ID) bool {
	if !sp.Condition.CanInvoke() {
		g.Log("This spirit cannot yet be invoked!")
		return false
	}
	if sp.Charges == 0 {
		g.Log("No charges remaining.")
		return false
	}
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) && !pa.DoesAny(ResistanceDaze) {
		g.Log("You cannot invoke a spirit while dazed.")
		return false
	}
	confused := pa.Has(StatusConfusion)
	if confused && pa.HP <= 2 {
		g.Log("You cannot invoke a spirit while confused and almost dead (HP < 3).")
		return false
	}
	ohp := pa.HP
	done := sp.GetAbility().Apply(g)
	if !done {
		return false
	}
	sp.Charges = max(0, sp.Charges-1)
	sp.Uses++
	g.Stats.SpiritUses++
	g.Stats.MapSpiritUses[g.Map.Level-1]++
	name := sp.GetAbility().Name()
	if confused {
		g.LogfStyled("Invoking the spirit while confused hurts you (%d dmg).", logHurtPlayer, ConfusionDamage)
		g.InflictDamage(PlayerID, pa, ConfusionDamage, AttackOther)
		g.StoryLogf("Used the “%s” ability while confused (HP: %d/%d, charges: %d/%d)",
			name, pa.HP, pa.GetMaxHP(), sp.Charges, sp.GetMaxCharges())
	} else if hp := pa.HP; hp == ohp {
		g.StoryLogf("Used the “%s” ability (charges: %d/%d)", name, sp.Charges, sp.GetMaxCharges())
	} else {
		g.StoryLogf("Used the “%s” ability (HP: %d/%d, charges: %d/%d)", name, hp, pa.GetMaxHP(), sp.Charges, sp.GetMaxCharges())
	}
	return true
}

// GetMaxCharges returns the maximum number of charges given the current spirit
// upgrade level.
func (sp *Spirit) GetMaxCharges() int {
	return sp.MaxCharges[sp.Level]
}

// GetAbility returns the maximum number of charges given the current spirit
// upgrade level.
func (sp *Spirit) GetAbility() Ability {
	return sp.Ability[sp.Level]
}

const ConfusionDamage = 1

// EmptyTotem represents a totem without spirit.
type EmptyTotem struct{}

func (sp *EmptyTotem) Desc(_ *Game) string {
	return "The totem does not have a spirit or it already left."
}

func (sp *EmptyTotem) Use(g *Game, i ID) bool {
	g.Log("You cannot equip a spirit from an empty totem.")
	return false
}

// Comestible represents a collectable comestible item.
type Comestible struct {
	Effect Effect // effect on use
}

func (c *Comestible) Desc(g *Game) string {
	return c.Effect.Desc(g)
}

func (c *Comestible) ExamineDesc(g *Game) string {
	return fmt.Sprintf("@CComestible.@N\n%s", c.Effect.Desc(g))
}

func (c *Comestible) Use(g *Game, i ID) bool {
	ef := c.Effect
	if g.Mod(ModTotemConditions) {
		co, cond := ef.(ConditionalComestible)
		if cond && !g.snack && !g.PlayerActor().DoesAny(Gluttony) && !co.CanApply(g, i) {
			// Unless the player is the Bear, forbid eating unless
			// condition is met.
			// NOTE: We have to check both snack and Gluttony above
			// because some curses may disable the latter, but we
			// want snack to still ignore requirements even then.
			g.Log(CantEatMsg(ef))
			return false
		}
	}
	pp := g.PP()
	done := ef.Apply(g)
	if !done {
		// Should never happen with comestibles.
		log.Printf("BUG: did not eat the %s.", g.Entity(i).Name)
		return false
	}
	g.Stats.Comestibles[g.Entity(i).Name]++
	g.Entities[i] = emptySlot()
	g.Stats.EatenComestibles++
	g.Stats.MapEatenComestibles[g.Map.Level-1]++
	g.handleGluttonyOnEating()
	for i, ai := range g.Monsters() {
		if ai.Is(HungryRat) && g.InFOV(g.Entity(i).P) {
			g.PutStatus(i, ai, StatusBerserk, DurationBerserk)
			// Make monster hunt in case it was still wandering
			// (with no extra log message).
			ai.Behavior.State = Hunting
			ai.Behavior.Target = pp
		}
	}
	return true
}

func (g *Game) handleGluttonyOnEating() {
	if GluttonyRework {
		return
	}
	switch pa := g.PlayerActor(); {
	case pa.Has(StatusGluttony):
		pa.Statuses[StatusGluttony] = 0
		g.LogStyled("You feel full.", logStatusEnd)
	case pa.DoesAny(Gluttony) && !g.snack:
		g.PutStatus1(PlayerID, pa, StatusGluttony, DurationGluttony)
	}
}

// EquipItemAt equips the item currently in the ground at index i.
func (g *Game) EquipItemAt(i ID) bool {
	ei := g.Entity(i)
	j, it := g.ItemAt(g.PP())
	if j <= 0 {
		// should not happen
		g.LogStyled("Cannot equip item (BUG).", logError)
		return false
	}
	ej := g.Entity(j)
	switch it := it.(type) {
	case *Spirit:
		pa := g.PlayerActor()
		ohpmax := pa.MaxHP
		itc := it.Condition // nil if ModTotemConditions is not enabled
		if spi, ok := ei.Role.(*Spirit); ok {
			if spi.Level >= 1 {
				// Should not happen.
				g.Log("Spirit cannot be upgraded anymore.")
				return false
			}
			spi.Level++
			spi.Charges += spi.MaxCharges[1] - spi.MaxCharges[0]
			if itc != nil && itc.Upgrade.Applies(spi.Condition != nil) {
				g.recordCurse(ei.Name, itc)
				spi.Condition = itc
			}
			g.Logf("You upgrade your %s spirit.", ei.Name)
			g.StoryLogf("Upgraded your %s spirit", ei.Name)
		} else {
			if itc != nil {
				g.recordCurse(ej.Name, itc)
			}
			g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
			g.Logf("You choose the %s spirit.", ej.Name)
			g.StoryLogf("Chose the %s spirit", ej.Name)
		}
		g.ComputePlayerStats()
		if hpmaxDelta := pa.MaxHP - ohpmax; hpmaxDelta > 0 {
			g.AdjustHP(PlayerID, pa, hpmaxDelta)
		}
		ej.P = InvalidPos
		et := g.Entity(g.genEmptyTotemAt(g.PP()))
		et.Seen = true
		g.totemPickupBerserk()
		return true
	case *Comestible:
		if ei.IsItem() {
			g.Logf("You pick the %s (replacing %s).", ej.Name, ei.Name)
			g.StoryLogf("Picked %s (replacing %s)", One(ej.Name), ei.Name)
			ei.P = ej.P
		} else {
			g.Logf("You pick the %s.", ej.Name)
			g.StoryLogf("Picked %s (comestibles: %d)", One(ej.Name), g.ComestibleCount()+1)
		}
		ej.P = InvalidPos
		g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
		TotemEvents(g, &EventHappened{EvType: EventPickup})
		return true
	default:
		panic("unexpected item type")
	}
}

// totemPickupBerserk makes monsters in view berserk (after you pick a totemic
// spirit).
func (g *Game) totemPickupBerserk() {
	berserk := false
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		if g.PutStatus(i, ai, StatusBerserk, DurationBerserkTotem) {
			berserk = true
		}
	}
	if berserk {
		g.LogStyled("Picking a spirit makes monsters in view berserk!", logSpecial)
	}
}

// Menhir represent a static magical menhir stone.
type Menhir struct {
	Used   bool   // whether already activated
	Effect Effect // effect on use
}

func (m *Menhir) Desc(g *Game) string {
	if !m.Used {
		return fmt.Sprintf("@CActivable@N.\n%s", m.Effect.Desc(g))
	}
	return m.Effect.Desc(g)
}

func (m *Menhir) Use(g *Game, i ID) bool {
	if m.Used {
		g.Log("The menhir is already inert.")
		return false
	}
	if g.WarnLeave(EventMenhir) {
		g.md.mode = modeUseConfirmation
		return false
	}
	done := m.Effect.Apply(g)
	if !done {
		return false
	}
	g.StoryLogf("Activated %s", One(g.Entity(i).Name))
	m.Used = true
	TotemEvents(g, &EventHappened{EvType: EventMenhir})
	g.Stats.ActivatedMenhirs++
	g.Stats.MapActivatedMenhirs[g.Map.Level-1]++
	return true
}

// Portal represent a magical portal leading to the next level.
type Portal struct {
	Fake bool // whether the portal is fake
	Used bool // whether fake but already activated
}

func (p *Portal) Desc(_ *Game) string {
	if p.Used {
		return "Malfunctioning magical portal that doesn’t lead anywhere, despite looking normal."
	}
	return fmt.Sprintf("@CActivable@N.\n%s", "A magical portal leading to the next level. Replenishes spirit charges and HP.")
}

func (p *Portal) Use(g *Game, i ID) bool {
	if p.Used {
		g.Log("This portal is malfunctional and cannot be triggered anymore.")
		return false
	}
	if p.Fake {
		p.Used = true
		g.useFakePortal()
		return true
	}
	if err := g.Save(); err != nil {
		log.Printf("error saving game before new level: %v", err)
	}
	g.StoryLog("Entered a magical portal!")
	g.LogStyled("You enter the portal.", logSpecial)
	g.NextLevel()
	return false
}

func (g *Game) useFakePortal() {
	g.LogStyled("You try to activate the magic portal…", logSpecial)
	g.StoryLog("Triggered a malfunctioning portal")
	const maxdist = MaxFOVRange + 3
	mp := &MapPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	g.MakeNoise(g.PP(), NoiseFakePortal)
	g.LogStyled("An eerie sound came out, but nothing happened!", logSpecial)
	TotemEvents(g, &EventHappened{EvType: EventFakePortal})
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearFakePortal)
		}
	}
}

// NextLevel proceeds to the next level (assuming we're not at the last one).
func (g *Game) NextLevel() {
	TotemEvents(g, &EventHappened{EvType: EventPortal})
	// We return false because after generating a new level we don't want
	// to end the turn. However, we still upgrade the turn counter, as it
	// wouldn't be intuitive otherwise.
	g.IncrTurn()   // call before InitLevel
	g.LevelStats() // call before InitLevel
	clear(g.Map.Noise)
	g.InitLevel()
	g.Logs.NextTick = g.Logs.Index
	g.md.targ.CancelExamine()
	TotemEvents(g, &EventHappened{EvType: EventNewLevel})
}

// CorruptionOrb represents the orb of corruption.
type CorruptionOrb struct {
	Broken bool // whether you broke the orb of corruption or not
}

func (o *CorruptionOrb) Desc(_ *Game) string {
	if o.Broken {
		// Should not happen in practice for now, unless we add extra
		// content after destroying the orb.
		return "The source of beast corruption that you destroyed."
	}
	return fmt.Sprintf("@CActivable@N.\n%s", "The source of beast corruption that you need to destroy. Beware that the Orb may sometimes actively drive you further away when you teleport.")
}

func (o *CorruptionOrb) Use(g *Game, i ID) bool {
	o.Broken = true
	if g.Wizard.Mode == WizardNone {
		g.LogStyled("You destroy the Orb of Corruption… You win!", logSpecial)
	} else {
		g.LogStyled("You destroy the Orb of Corruption… **WIZARD**!", logSpecial)
	}
	g.StoryLog("Destroyed the Orb of Corruption!")
	g.md.OrbDestructionAnimation()
	g.LevelStats()
	g.md.logConfirmContinue()
	g.IncrTurn()
	g.win = true
	g.md.mode = modeEnd
	return false
}
