// many ideas here from articles found at http://www.roguebasin.com/

package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// Map represents a map level.
type Map struct {
	Depth           int                     // dungeon depth
	Terrain         rl.Grid                 // map terrain
	KnownTerrain    rl.Grid                 // player's terrain knowledge
	MagicalBarriers map[gruid.Point]rl.Cell // original terrain (under the barrier)
	Clouds          map[gruid.Point]Cloud   // clouds
	Noise           map[gruid.Point]int     // noise sources
	NoiseIllusion   map[gruid.Point]int     // harmonic noise from noise magara
}

// At returns the terrain cell at the given point.
func (m *Map) At(p gruid.Point) rl.Cell {
	return m.Terrain.At(p)
}

// Border reports whether the given point is at map boundaries.
func (m *Map) Border(p gruid.Point) bool {
	return p.X == MapWidth-1 || p.Y == MapHeight-1 || p.X == 0 || p.Y == 0
}

// Set updates terrain at given map point.
func (m *Map) Set(p gruid.Point, t rl.Cell) {
	m.Terrain.Set(p, t)
}

// HasFreeNeighbor reports whether the given position has a free neighbor.
func (m *Map) HasFreeNeighbor(p gruid.Point) bool {
	for q := range Neighbors(p) {
		if Passable(m.At(q)) {
			return true
		}
	}
	return false
}

// IsKnown reports whether the terrain is known.
func IsKnown(t rl.Cell) bool {
	return t != Unknown && t != UnknownPassable
}

// Passable reports whether the given position is always passable. In other
// words, that it's not some kind of obstacle.
func (m *Map) Passable(p gruid.Point) bool {
	return Passable(m.Terrain.At(p))
}

// NoWallAt reports whether the given position is not a (regular) wall.
func (m *Map) NoWallAt(p gruid.Point) bool {
	// Cells outside the map are considered as walls, so not walkable.
	return m.Terrain.At(p) != WallCell
}
