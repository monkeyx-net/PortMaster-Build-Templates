package main

import (
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// The following constants represent the various kinds of terrains.
const (
	WallCell rl.Cell = iota
	GroundCell
	DoorCell
	FoliageCell
	BarrelCell
	StairCell
	StoneCell
	MagaraCell
	BananaCell
	LightCell
	ExtinguishedLightCell
	TableCell
	TreeCell
	HoledWallCell
	ScrollCell
	StoryCell
	ItemCell
	BarrierCell
	WindowCell
	ChasmCell
	WaterCell
	RubbleCell
	CavernCell
	PotionCell
	QueenRockCell

	UnknownPassable // cell is known to be passable (for KnownTerrain field of Map)
	Unknown         // cell is unknown (for KnownTerrain field of Map)
)

// Passable reports whether a given terrain type is Passable (for any actor).
func Passable(t rl.Cell) bool {
	switch t {
	case WallCell, DoorCell, BarrelCell, TableCell, TreeCell, HoledWallCell, BarrierCell, WindowCell, StoryCell, ChasmCell, WaterCell:
		return false
	default:
		return true
	}
}

// IsJumpPassable reports whether a given terrain type is passable, including
// jump-passability.
func IsJumpPassable(t rl.Cell) bool {
	switch t {
	case TableCell, ChasmCell, WaterCell, StoryCell:
		return true
	default:
		return Passable(t)
	}
}

// IsNormalPatrolWay reports whether patrols may usually go through that
// terrain.
func IsNormalPatrolWay(t rl.Cell) bool {
	switch t {
	case GroundCell, ScrollCell, DoorCell, StairCell, LightCell, ItemCell, ExtinguishedLightCell, StoneCell, MagaraCell:
		return true
	default:
		return false
	}
}

// IsLevitatePassable reports whether the terrain is passable when levitating.
func IsLevitatePassable(t rl.Cell) bool {
	switch t {
	case ChasmCell, WaterCell:
		return true
	default:
		return Passable(t)
	}
}

// IsDoorPassable reports whether the terrain is passable or a door.
func IsDoorPassable(t rl.Cell) bool {
	switch t {
	case DoorCell:
		return true
	default:
		return Passable(t)
	}
}

// IsSwimPassable reports whether the terrain is passable or water.
func IsSwimPassable(t rl.Cell) bool {
	switch t {
	case WaterCell:
		return true
	default:
		return Passable(t)
	}
}

// IsJumpPropulsion reports whether the terrain may be used as jumping
// propulsion.
func IsJumpPropulsion(t rl.Cell) bool {
	switch t {
	case WallCell, WindowCell, BarrierCell:
		return true
	default:
		return false
	}
}

// IsEnclosing reports whether the terrain is closed and prevents jumping.
func IsEnclosing(t rl.Cell) bool {
	switch t {
	case BarrelCell, TableCell, HoledWallCell:
		return true
	default:
		return false
	}
}

// AllowsFog reports whether the terrain may be filled with fog.
func AllowsFog(t rl.Cell) bool {
	switch t {
	case WallCell, HoledWallCell, WindowCell, StoryCell, RubbleCell:
		return false
	default:
		return true
	}
}

// CoversPlayer reports whether the terrain covers for the player (drawn on top).
func CoversPlayer(t rl.Cell) bool {
	switch t {
	case BarrelCell, TableCell, TreeCell, HoledWallCell, BarrierCell:
		return true
	default:
		return false
	}
}

// IsPlayerPassable reports whether the terrain is always passable by the
// player.
func IsPlayerPassable(t rl.Cell) bool {
	switch t {
	case WallCell, BarrierCell, WindowCell, ChasmCell:
		return false
	default:
		return true
	}
}

// IsDiggable reports whether the terrain is some kind of diggable wall.
func IsDiggable(t rl.Cell) bool {
	switch t {
	case WallCell, WindowCell, HoledWallCell:
		return true
	default:
		return false
	}
}

// BlocksRange reports whether the t errain blocks ranged attacks.
func BlocksRange(t rl.Cell) bool {
	switch t {
	case WallCell, TreeCell, BarrierCell, WindowCell, StoryCell:
		return true
	default:
		return false
	}
}

// Hides reports whether the terrain provides hiding for the player if behind.
func Hides(t rl.Cell) bool {
	switch t {
	case WallCell, BarrelCell, TableCell, TreeCell, WindowCell, StoryCell:
		return true
	default:
		return false
	}
}

// IsIlluminable reports whether the terrain is illuminable.
func IsIlluminable(t rl.Cell) bool {
	switch t {
	case WallCell, BarrelCell, TableCell, TreeCell, HoledWallCell, BarrierCell, WindowCell, ChasmCell, RubbleCell:
		return false
	}
	return true
}

// IsDestructible reports whether the terrain is destructible (diggable or
// flammable).
func IsDestructible(t rl.Cell) bool {
	switch t {
	case WallCell, BarrelCell, DoorCell, TableCell, TreeCell, HoledWallCell, WindowCell:
		return true
	default:
		return false
	}
}

// Flammable reports whether the terrain is flammable.
func Flammable(t rl.Cell) bool {
	switch t {
	case FoliageCell, DoorCell, BarrelCell, TableCell, TreeCell:
		return true
	default:
		return false
	}
}

// IsGround reports whether the terrain is some kind of ground cell
// (with no special features).
func IsGround(t rl.Cell) bool {
	switch t {
	case GroundCell, CavernCell, QueenRockCell:
		return true
	default:
		return false
	}
}

// IsNotable reports whether the terrain is notable (for log purposes).
func IsNotable(t rl.Cell) bool {
	switch t {
	case StairCell, StoneCell, BarrelCell, MagaraCell, BananaCell,
		ScrollCell, ItemCell, PotionCell:
		return true
	default:
		return false
	}
}

// TerrainName returns the name of the terrain assuming it's at the given
// positon.
func (g *Game) TerrainName(t rl.Cell, p gruid.Point) (desc string) {
	switch t {
	case WallCell:
		desc = "wall"
	case GroundCell:
		desc = "paved ground"
	case DoorCell:
		desc = "door"
	case FoliageCell:
		desc = "foliage"
	case BarrelCell:
		desc = "barrel"
	case StoneCell:
		desc = g.Objects.Stones[p].String()
	case StairCell:
		desc = g.Objects.Stairs[p].ShortString(g)
	case MagaraCell:
		desc = g.Objects.Magaras[p].String()
	case BananaCell:
		desc = "banana"
	case LightCell:
		desc = "campfire"
	case ExtinguishedLightCell:
		desc = "extinguished campfire"
	case TableCell:
		desc = "table"
	case TreeCell:
		desc = "banana tree"
	case HoledWallCell:
		desc = "holed wall"
	case ScrollCell:
		desc = g.Objects.Scrolls[p].String()
	case StoryCell:
		desc = g.Objects.Story[p].String()
	case ItemCell:
		desc = g.Objects.Items[p].String()
	case BarrierCell:
		desc = "magical barrier"
	case WindowCell:
		desc = "closed window"
	case ChasmCell:
		if g.Map.Depth == MaxDepth || g.Map.Depth == WinDepth {
			desc = "deep chasm"
		} else {
			desc = "chasm"
		}
	case WaterCell:
		desc = "shallow water"
	case RubbleCell:
		desc = "rubblestone"
	case CavernCell:
		desc = "cave ground"
	case PotionCell:
		desc = g.Objects.Potions[p].String()
	case QueenRockCell:
		desc = "queen rock"
	case UnknownPassable:
		desc = "passable terrain"
	default:
		desc = "unknown terrain"
	}
	return desc
}

// TerrainShortDesc returns the naming in a sentence of the terrain assuming
// it's at the given positon.
func (g *Game) TerrainShortDesc(t rl.Cell, p gruid.Point) (desc string) {
	switch t {
	case WallCell:
		desc = "a wall"
	case GroundCell:
		desc = "paved ground"
	case DoorCell:
		desc = "a door"
	case FoliageCell:
		desc = "foliage"
	case BarrelCell:
		desc = "a barrel"
	case StoneCell:
		desc = g.Objects.Stones[p].ShortDesc(g)
	case StairCell:
		desc = g.Objects.Stairs[p].ShortDesc(g)
	case MagaraCell:
		desc = g.Objects.Magaras[p].ShortDesc()
	case BananaCell:
		desc = "a banana"
	case LightCell:
		desc = "a campfire"
	case ExtinguishedLightCell:
		desc = "an extinguished campfire"
	case TableCell:
		desc = "a table"
	case TreeCell:
		desc = "a banana tree"
	case HoledWallCell:
		desc = "a holed wall"
	case ScrollCell:
		desc = g.Objects.Scrolls[p].ShortDesc()
	case StoryCell:
		desc = g.Objects.Story[p].String()
	case ItemCell:
		desc = g.Objects.Items[p].ShortDesc(g)
	case BarrierCell:
		desc = "a magical barrier"
	case WindowCell:
		desc = "a closed window"
	case ChasmCell:
		if g.Map.Depth == MaxDepth || g.Map.Depth == WinDepth {
			desc = "a deep chasm"
		} else {
			desc = "a chasm"
		}
	case WaterCell:
		desc = "shallow water"
	case RubbleCell:
		desc = "rubblestone"
	case CavernCell:
		desc = "cave ground"
	case PotionCell:
		desc = g.Objects.Potions[p].ShortDesc(g)
	case QueenRockCell:
		desc = "queen rock"
	case UnknownPassable:
		desc = "unknown passable terrain"
	default:
		desc = "unknown terrain"
	}
	return desc
}

// TerrainShortDesc returns the description of the terrain assuming it's at the
// given positon.
func (g *Game) TerrainDesc(t rl.Cell, p gruid.Point) (desc string) {
	switch t {
	case WallCell:
		desc = "A wall is an obstructing pile of rocks."
	case GroundCell:
		desc = "Paved ground is a way constructed with small stones, often used by patrols between buildings."
	case DoorCell:
		desc = "A closed door blocks your line of sight. Doors open automatically when you or a creature stand on them."
	case FoliageCell:
		desc = "Blue dense foliage grows in Hareka’s Underground. It is difficult to see through."
	case BarrelCell:
		desc = "A barrel. You can hide yourself inside it when no creatures see you. It is a safe place for resting and recovering."
	case StoneCell:
		desc = g.Objects.Stones[p].Desc(g)
	case StairCell:
		desc = g.Objects.Stairs[p].Desc(g)
	case MagaraCell:
		desc = g.Objects.Magaras[p].Desc(g)
	case BananaCell:
		desc = "A gawalt monkey cannot enter a healthy sleep without eating one of those bananas before."
	case LightCell:
		desc = "A campfire illuminates surrounding cells. Creatures can spot you in illuminated cells from a greater range. You can extinguish it, but guards may light it up again."
	case ExtinguishedLightCell:
		desc = "An extinguished campfire can be lighted again by guards."
	case TableCell:
		desc = "You can hide under the table so that only adjacent creatures can see you. Most creatures cannot walk across the table."
	case TreeCell:
		desc = "Underground banana trees grow with nearly no light sources. Their rare bananas are very appreciated by many creatures, specially some harpy species. You may find some bananas dropped by them while exploring. You can climb trees to see farther. Moreover, only big, flying or jumping creatures will be able to attack you while you stand on a tree. The top is never illuminated."
	case HoledWallCell:
		desc = "Only very small creatures can pass there. It is difficult to see through."
	case ScrollCell:
		desc = g.Objects.Scrolls[p].Desc(g)
	case StoryCell:
		desc = g.Objects.Story[p].Desc(g, p)
	case ItemCell:
		desc = g.Objects.Items[p].Desc(g)
	case BarrierCell:
		desc = "A temporal magical barrier created by oric energies. It may have been created by an oric magara or an oric celmist. Sometimes, natural oric energies may produce such barriers too in energetically unstable Underground areas."
	case WindowCell:
		desc = "A transparent window in the wall."
	case ChasmCell:
		if g.Map.Depth == MaxDepth || g.Map.Depth == WinDepth {
			desc = "A deep chasm. If you jump into it, you’ll be dead."
		} else {
			desc = "A chasm. If you jump into it, you’ll reach the next level, but you’ll be seriously injured."
		}
	case WaterCell:
		desc = "This is shallow water. Only monsters that can swim will follow you there."
	case RubbleCell:
		desc = "Rubblestone is a collection of rocks broken into smaller stones. They block vision and are never well illuminated."
	case CavernCell:
		desc = "This is natural cave ground."
	case PotionCell:
		desc = g.Objects.Potions[p].Desc(g)
	case QueenRockCell:
		desc = "Queen rock amplifies sounds. Even though you are usually very silent, monsters may hear your footsteps when walking on those rocks."
	case UnknownPassable:
		desc = "You heard some monster noise around there in the past."
	case Unknown:
		desc = "You do not known what is in there."
	}
	var autodesc string
	if !IsPlayerPassable(t) {
		autodesc += " It is impassable."
	}
	if Flammable(t) {
		autodesc += " It is flammable."
	}
	if IsLevitatePassable(t) && !Passable(t) {
		autodesc += " It can be traversed with levitation."
	}
	if IsDiggable(t) && !Passable(t) {
		autodesc += " It is diggable by oric destructive magic."
	}
	if IsSwimPassable(t) && !Passable(t) {
		autodesc += " It is possible to traverse by swimming."
	}
	if BlocksRange(t) {
		autodesc += " It blocks ranged attacks from foes."
	}
	if Hides(t) && t != WallCell {
		autodesc += " You can hide just behind it."
	}
	if IsJumpPropulsion(t) {
		autodesc += " You can use it to propel yourself with a jump."
	}
	if autodesc != "" {
		desc += "\n\n" + strings.TrimSpace(autodesc)
	}
	return desc
}

// TerrainStyle returns the styling information of the terrain assuming it's at
// the given positon.
func (g *Game) TerrainStyle(t rl.Cell, p gruid.Point) (r rune, fg gruid.Color) {
	switch t {
	case WallCell:
		r, fg = '#', ColorForeground
	case GroundCell:
		r, fg = '.', ColorForeground
	case DoorCell:
		r, fg = '+', ColorMagenta
	case FoliageCell:
		r, fg = '"', ColorForeground
	case BarrelCell:
		r, fg = '&', ColorYellow
	case StoneCell:
		r, fg = g.Objects.Stones[p].Style(g)
	case StairCell:
		st := g.Objects.Stairs[p]
		r, fg = st.Style(g)
	case MagaraCell:
		r, fg = '/', ColorCyan
	case BananaCell:
		r, fg = ')', ColorYellow
	case LightCell:
		r, fg = '☼', ColorYellow
	case ExtinguishedLightCell:
		r, fg = '○', ColorForeground
	case TableCell:
		r, fg = 'π', ColorYellow
	case TreeCell:
		r, fg = '♣', ColorGreen
	case HoledWallCell:
		r, fg = 'Π', ColorViolet
	case ScrollCell:
		r, fg = g.Objects.Scrolls[p].Style(g)
	case StoryCell:
		r, fg = g.Objects.Story[p].Style(g)
	case ItemCell:
		r, fg = g.Objects.Items[p].Style(g)
	case BarrierCell:
		r, fg = 'Ξ', ColorCyan
	case WindowCell:
		r, fg = 'Θ', ColorViolet
	case ChasmCell:
		r, fg = '◊', ColorForeground
		if g.Map.Depth == MaxDepth || g.Map.Depth == WinDepth {
			fg = ColorViolet
		}
	case WaterCell:
		r, fg = '≈', ColorForeground
	case RubbleCell:
		r, fg = '^', ColorForeground
	case CavernCell:
		r, fg = ',', ColorForeground
	case PotionCell:
		r, fg = g.Objects.Potions[p].Style(g)
	case QueenRockCell:
		r, fg = '‗', ColorForeground
	case UnknownPassable:
		r, fg = '♫', ColorForegroundSecondary
	case Unknown:
		r, fg = ' ', ColorForegroundSecondary
	}
	return r, fg
}
