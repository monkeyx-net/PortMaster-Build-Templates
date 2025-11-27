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
	Desc() string
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
	Desc() string
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
}

func (sp *Spirit) Desc() string {
	var sb strings.Builder
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
		sb.WriteString("Gives")
		for i, bonus := range bonuses {
			sb.WriteByte(' ')
			sb.WriteString(bonus)
			if i < len(bonuses)-1 {
				sb.WriteByte(',')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "@CAbility:@N %s (%d/%d charges). %s",
		sp.Ability[sp.Level].Name(), sp.Charges, sp.MaxCharges[sp.Level], sp.Ability[sp.Level].Desc())
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "@CTraits:@N %s.", sp.BonusTraits[sp.Level])
	return sb.String()
}

func (sp *Spirit) UpgradeDesc() string {
	desc := sp.Desc()
	var sb strings.Builder
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
	sb.WriteString("@MThe spirit essence can be used to upgrade this spirit.@N\n")
	if len(bonuses) > 0 {
		sb.WriteString("Gives ")
		for i, bonus := range bonuses {
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
		fmt.Fprintf(&sb, "@BNew traits:@N %s.", sp.BonusTraits[1]&^sp.BonusTraits[0])
	}
	return desc + "\n\n" + sb.String()
}

func (sp *Spirit) Use(g *Game, i ID) bool {
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
	done := sp.Ability[sp.Level].Apply(g)
	if !done {
		return false
	}
	sp.Charges--
	sp.Uses++
	g.Stats.SpiritUses++
	g.Stats.MapSpiritUses[g.Map.Level-1]++
	name := sp.Ability[sp.Level].Name()
	if confused {
		g.LogfStyled("Invoking the spirit while confused hurts you (%d dmg).", logHurtPlayer, ConfusionDamage)
		g.InflictDamage(PlayerID, pa, ConfusionDamage, AttackOther)
		g.StoryLogf("Used the “%s” ability while confused (HP: %d/%d)",
			name, pa.HP, pa.GetMaxHP())
	} else if hp := pa.HP; hp == ohp {
		g.StoryLogf("Used the “%s” ability", name)
	} else {
		g.StoryLogf("Used the “%s” ability (HP: %d/%d)", name, hp, pa.GetMaxHP())
	}
	return true
}

const ConfusionDamage = 1

// EmptyTotem represents a totem without spirit.
type EmptyTotem struct{}

func (sp *EmptyTotem) Desc() string {
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

func (c *Comestible) Desc() string {
	return fmt.Sprintf("@CComestible.@N\n%s", c.Effect.Desc())
}

func (c *Comestible) Use(g *Game, i ID) bool {
	pp := g.PP()
	done := c.Effect.Apply(g)
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
		if ai.DoesAny(MonsHungry) && g.InFOV(g.Entity(i).P) {
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
	if g.Mod(ModGluttonyRework) {
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
	switch it.(type) {
	case *Spirit:
		pa := g.PlayerActor()
		ohpmax := pa.MaxHP
		if ev, ok := ei.Role.(*Spirit); ok {
			if ev.Level >= 1 {
				// Should not happen.
				g.Log("Spirit cannot be upgraded anymore.")
				return false
			}
			ev.Level++
			ev.Charges += ev.MaxCharges[1] - ev.MaxCharges[0]
			g.Logf("You upgrade your %s spirit.", ei.Name)
			g.StoryLogf("Upgraded your %s spirit", ei.Name)
		} else {
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
			g.StoryLogf("Picked %s", One(ej.Name))
		}
		ej.P = InvalidPos
		g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
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

func (m *Menhir) Desc() string {
	if !m.Used {
		return fmt.Sprintf("@CActivable@N.\n%s", m.Effect.Desc())
	}
	return m.Effect.Desc()
}

func (m *Menhir) Use(g *Game, i ID) bool {
	if m.Used {
		g.Log("Magical stone is already inert.")
		return false
	}
	done := m.Effect.Apply(g)
	if !done {
		return false
	}
	g.StoryLogf("Activated %s", One(g.Entity(i).Name))
	m.Used = true
	g.Stats.ActivatedMenhirs++
	g.Stats.MapActivatedMenhirs[g.Map.Level-1]++
	return true
}

// Portal represent a magical portal leading to the next level.
type Portal struct {
	Fake bool // whether the portal is fake
	Used bool // whether fake but already activated
}

func (p *Portal) Desc() string {
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
		log.Printf("saving game before new level: %v", err)
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
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearFakePortal)
		}
	}
}

// NextLevel proceeds to the next level (assuming we're not at the last one).
func (g *Game) NextLevel() {
	// We return false because after generating a new level we don't want
	// to end the turn. However, we still upgrade the turn counter, as it
	// wouldn't be intuitive otherwise.
	g.IncrTurn()   // call before InitLevel
	g.LevelStats() // call before InitLevel
	clear(g.Map.Noise)
	g.InitLevel()
	g.Logs.NextTick = g.Logs.Index
	g.md.targ.CancelExamine()
}

// CorruptionOrb represents the orb of corruption.
type CorruptionOrb struct {
	Broken bool // whether you broke the orb of corruption or not
}

func (o *CorruptionOrb) Desc() string {
	if o.Broken {
		// Should not happen in practice for now, unless we add extra
		// content after destroying the orb.
		return "The source of beast corruption that you destroyed."
	}
	return fmt.Sprintf("@CActivable@N.\n%s", "The source of beast corruption that you need to destroy.")
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
