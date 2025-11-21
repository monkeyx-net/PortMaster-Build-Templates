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
func (rt RunicTrap) Color() gruid.Color {
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
func (rt RunicTrap) Desc() string {
	if rt.KnownUsed {
		return "The rune has already been triggered."
	}
	switch rt.Rune {
	case RuneBerserk:
		return fmt.Sprintf("Applies @BBerserk@N status for %d turns when stepped on.", DurationBerserkTrap)
	case RuneFire:
		return fmt.Sprintf("Inflicts @OConfusion@N for %d turns and spawns a fire cloud lasting 6-12 turns when stepped on.",
			DurationConfusionFireTrap)
	case RuneLignification:
		return fmt.Sprintf("Applies @BLignification@N status for %d turns when stepped on.", DurationLignificationTrap)
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
func (rt RunicTrap) Use(g *Game, id ID) bool {
	g.Log("You cannot interact with a trap.")
	return false
}

// NoTrapAt reports whether there's no trap at the given position.
func (g *Game) NoTrapAt(p gruid.Point) bool {
	return g.Map.RuneCache.At(p) <= 0
}

// TriggerTrapAt triggers any runic traps at the given point.
func (g *Game) TriggerTrapAt(p gruid.Point) {
	i := g.Map.RuneCache.At(p)
	if i <= 0 {
		return
	}
	ei := g.Entity(i)
	rt, ok := ei.Role.(*RunicTrap)
	if !ok || rt.Used {
		return
	}
	if j, _ := g.ActorAt(p); j >= 0 {
		if j == PlayerID {
			g.LogfStyled("You trigger the %s.", logSpecial, rt.Rune.String())
			g.StoryLogf("Triggered %s", One(rt.Rune.String()))
			g.Stats.PlayerTrapTriggers++
		} else if g.InFOV(p) {
			g.LogfStyled("The %s triggers the %s.", logSpecial, g.Entity(j).Name, rt.Rune.String())
			g.Stats.MonsterTrapTriggers++
		}
		g.Stats.MapTriggeredTraps[g.Map.Level-1]++
	}
	switch rt.Rune {
	case RuneBerserk:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.PutStatus1(i, ai, StatusBerserk, DurationBerserkTrap)
		}
	case RuneFire:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.PutStatus1(i, ai, StatusConfusion, DurationConfusionFireTrap)
		}
		g.FireCloudAt(p)
	case RuneLignification:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.PutStatus1(i, ai, StatusLignification, DurationLignificationTrap)
		}
	case RunePoison:
		g.PoisonCloudAt(p)
	case RuneWarp:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.TeleportActor(i, ai, 1)
		}
	}
	rt.Used = true
	if g.InFOV(p) {
		rt.KnownUsed = true
	}
	g.Map.RuneCache.SetU(p, 0)
}
