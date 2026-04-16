package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
)

// RunicTrap represents a triggerable runic trap.
type RunicTrap struct {
	Used      bool      // whether the rune has already been triggered
	KnownUsed bool      // whether the rune is known to have already been triggered
	Rune      MagicRune // magic rune
}

// MagicRune represents the various kinds of runes used in runic traps.
type MagicRune int

const (
	RuneBerserk MagicRune = iota
	RuneFire
	RuneLignification
	RunePoison
	RuneWarp
)

// NRunes is the number of non-inert distinct magical runes.
const NRunes = 1 + int(RuneWarp)

// String returns the name of the runic trap.
func (mr MagicRune) String() string {
	switch mr {
	case RuneBerserk:
		return "berserk rune"
	case RuneFire:
		return "fire rune"
	case RuneLignification:
		return "lignification rune"
	case RunePoison:
		return "poison rune"
	case RuneWarp:
		return "warp rune"
	default:
		// Should not happen.
		return "unknown rune"
	}
}

// Color returns the color of the runic trap.
func (rt *RunicTrap) Color() gruid.Color {
	if rt.KnownUsed {
		return ColorForeground
	}
	switch rt.Rune {
	case RuneBerserk:
		return ColorMagenta
	case RuneFire:
		return ColorRed
	case RuneLignification:
		return ColorOrange
	case RunePoison:
		return ColorGreen
	case RuneWarp:
		return ColorViolet
	default:
		// Should not happen.
		return ColorForeground
	}
}

// Desc returns the description of the runic trap.
func (rt *RunicTrap) Desc() string {
	if rt.KnownUsed {
		return "The rune has already been triggered."
	}
	switch rt.Rune {
	case RuneBerserk:
		return fmt.Sprintf("Applies @BBerserk@N status for %d turns when stepped on, unless already berserk.", DurationBerserkTrap)
	case RuneFire:
		return fmt.Sprintf("Inflicts @OConfusion@N for %d turns and spawns a fire cloud lasting 6-12 turns when stepped on.",
			DurationConfusionFireTrap)
	case RuneLignification:
		return fmt.Sprintf("Applies @BLignification@N status for %d turns when stepped on, unless already lignified.", DurationLignificationTrap)
	case RunePoison:
		return "Spawns a poison cloud lasting 7-11 turns when stepped on."
	case RuneWarp:
		return "Teleports away any actor that steps on it."
	default:
		// Should not happen.
		return "Unknown effect."
	}
}

// RunicTraps are items but they cannot be used like other items.
func (rt *RunicTrap) Use(g *Game, id ID) bool {
	g.Log("You cannot interact with a trap.")
	return false
}

// NoTrapAt reports whether there's no trap at the given position.
func (g *Game) NoTrapAt(p gruid.Point) bool {
	return g.Map.RuneCache.At(p) <= 0
}

// TriggerTrap makes the given actor trigger a runic trap at its current
// position, if any.
func (g *Game) TriggerTrap(i ID, ai *Actor) {
	if ai.DoesAny(RunicChicken) {
		return
	}
	ei := g.Entity(i)
	pi := ei.P
	rt, ok := g.RunicTrapAt(pi)
	if !ok || rt.Used {
		return
	}
	switch {
	case rt.Rune == RuneLignification && (ai.Has(StatusLignification) || ai.Is(WalkingTree)):
		// Do not trigger lignification traps while lignified.
		return
	case rt.Rune == RuneBerserk && ai.Has(StatusBerserk):
		// Do not trigger berserk traps while berserk.
		return
	}
	switch {
	case i == PlayerID:
		g.LogfStyled("You trigger the %s.", logSpecial, rt.Rune.String())
		g.StoryLogf("Triggered %s", One(rt.Rune.String()))
		g.Stats.PlayerTrapTriggers++
		rt.KnownUsed = true
	case g.InFOV(pi):
		g.LogfStyled("The %s triggers the %s.", logSpecial, ei.Name, rt.Rune.String())
		rt.KnownUsed = true
	case SensingRange(pi, g.PP()) && g.PlayerActor().DoesAny(RunicChicken):
		g.LogfStyled("The %s triggers.", logSpecial, ei.Name)
		rt.KnownUsed = true
	}
	if i != PlayerID {
		g.Stats.MonsterTrapTriggers++
	}
	g.Stats.MapTriggeredTraps[g.Map.Level-1]++
	rt.Used = true
	g.Map.RuneCache.SetU(pi, 0)
	// Apply runic effect.
	switch rt.Rune {
	case RuneBerserk:
		g.PutStatus1(i, ai, StatusBerserk, DurationBerserkTrap)
	case RuneFire:
		g.PutStatus1(i, ai, StatusConfusion, DurationConfusionFireTrap)
		g.FireCloudAt(pi)
	case RuneLignification:
		g.PutStatus1(i, ai, StatusLignification, DurationLignificationTrap)
	case RunePoison:
		g.PoisonCloudAt(pi)
	case RuneWarp:
		g.TeleportActor(i, ai, 1)
	}
}

// RunicTrapAt returns any runic trap at the given position.
func (g *Game) RunicTrapAt(at gruid.Point) (*RunicTrap, bool) {
	i := g.Map.RuneCache.At(at)
	if i <= 0 {
		return nil, false
	}
	ei := g.Entity(i)
	rt, ok := ei.Role.(*RunicTrap)
	return rt, ok
}
