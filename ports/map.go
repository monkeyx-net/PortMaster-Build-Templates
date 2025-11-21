// This file contains map-related code.

package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// These constants represent the different kind of map tiles.
const (
	Wall            rl.Cell = iota // obstructing and blocks vision
	Floor                          // passable ground
	Foliage                        // passable but difficult to see through
	Rubble                         // passable, but blocks vision like walls
	TranslucentWall                // obstructing, but does not block vision, contains poison gas

	UnknownPassable // cell is known to be passable (for KnownTerrain field of Map)
	Unknown         // cell is unknown (for KnownTerrain field of Map)
)

func TerrainName(t rl.Cell) string {
	switch t {
	case Wall:
		return "wall"
	case Floor:
		return "floor"
	case Foliage:
		return "foliage"
	case Rubble:
		return "rubble"
	case TranslucentWall:
		return "translucent wall"
	case UnknownPassable:
		return "passable terrain"
	default:
		return "unknown terrain"
	}
}

func TerrainDesc(t rl.Cell) string {
	switch t {
	case Wall:
		return "An obstructing pile of rocks that blocks vision."
	case Floor:
		return "Passable plain ground."
	case Foliage:
		return "Dense foliage that is difficult to see through. Flammable."
	case Rubble:
		return "A collection of rocks broken into smaller ones. Passable, but blocks vision."
	case TranslucentWall:
		return "An obstructing pile of translucent rocks. Contains poison gas."
	case UnknownPassable:
		return "Passable terrain of unknown nature."
	default:
		return "Terrain of unknown nature."
	}
}

// IsKnown reports whether the terrain is known.
func IsKnown(t rl.Cell) bool {
	return t != Unknown && t != UnknownPassable
}

// Map represents the rectangular map of the game's level.
type Map struct {
	Terrain      rl.Grid                   // terrain
	KnownTerrain rl.Grid                   // player's terrain knowledge
	FOV          *rl.FOV                   // player's field of view
	FOVPts       []gruid.Point             // points in ssc field of view
	Clouds       *CloudGrid                // cloud map
	Level        int                       // current map level
	Noise        map[gruid.Point]NoiseType // non-movement noise sources
	Waypoints    []gruid.Point             // vault waypoints

	BoolCache  CacheGrid[bool] // for boolean caching in algorithms
	ActorCache CacheGrid[ID]   // for caching actor positions
	RuneCache  CacheGrid[ID]   // for caching runic trap positions

	Orb    gruid.Point // position of Orb of Corruption (if on the map)
	Portal gruid.Point // position of (true) portal
	Totem  gruid.Point // position of totem
}

// NewMap returns a new map object ready for use.
func NewMap() *Map {
	m := &Map{
		Terrain:      rl.NewGrid(MapWidth, MapHeight),
		KnownTerrain: rl.NewGrid(MapWidth, MapHeight),
		Clouds:       NewCloudGrid(),
		FOV:          rl.NewFOV(gruid.NewRange(-MaxFOVRange, -MaxFOVRange, MaxFOVRange+1, MaxFOVRange+1)),
		Noise:        map[gruid.Point]NoiseType{},
		Orb:          InvalidPos,
	}
	return m
}

// inMap reports whether a position is within map bounds (assumming they're
// relative to the upper-left map's corner).
func inMap(p gruid.Point) bool {
	return p.X >= 0 && p.X < MapWidth && p.Y >= 0 && p.Y < MapHeight
}

// Passable reports whether the given position is passable. In other words,
// that it's not an obstacle.
func (m *Map) Passable(p gruid.Point) bool {
	// Cells outside the map are considered as walls, so not walkable.
	return Passable(m.Terrain.At(p))
}

// PassableWithoutTraps reports whether the given position is passable without
// traps.
func (m *Map) PassableWithoutTraps(p gruid.Point) bool {
	return Passable(m.Terrain.At(p)) && m.RuneCache.At(p) <= 0
}

// NoWallAt reports whether the given position is not a (regular) wall.
func (m *Map) NoWallAt(p gruid.Point) bool {
	// Cells outside the map are considered as walls, so not walkable.
	return m.Terrain.At(p) != Wall
}

// Passable reports whether a given terrain type is Passable.
func Passable(t rl.Cell) bool {
	return t != Wall && t != TranslucentWall
}

// Burnable reports whether at the given position there is a burnable tile
// (foliage).
func (m *Map) Burnable(p gruid.Point) bool {
	return m.Terrain.At(p) == Foliage
}

// MapRune returns the character rune representing a given terrain.
func MapRune(t rl.Cell) (r rune) {
	switch t {
	case Wall:
		r = '#'
	case Floor:
		r = '.'
	case Foliage:
		r = '"'
	case Rubble:
		r = '^'
	case TranslucentWall:
		r = '◊'
	case UnknownPassable:
		r = '♫'
	default:
		r = '?'
	}
	return r
}

// RuneAt returns the character rune at a given position.
func (m *Map) RuneAt(p gruid.Point) rune {
	t := m.KnownTerrain.At(p)
	if t != Unknown {
		return MapRune(t)
	}
	for q := range m.PassableNeighbors(p) {
		if tq := m.KnownTerrain.At(q); IsKnown(tq) && Passable(tq) {
			return '¤'
		}
	}
	return ' '
}

// CacheGrid represents a map-sized grid of any type.
type CacheGrid[T any] []T

// At returns the value in the grid at a given position.
func (bs CacheGrid[T]) At(p gruid.Point) T {
	var zero T
	i := p.Y*MapWidth + p.X
	if i >= 0 && i < len(bs) && p.X >= 0 && p.X < MapWidth {
		return bs[i]
	}
	return zero
}

// AtU returns the value in the grid at a given position. It doesn't check
// boundaries.
func (bs CacheGrid[T]) AtU(p gruid.Point) T {
	i := p.Y*MapWidth + p.X
	return bs[i]
}

// Set puts a value at the given position in the grid.
func (bs CacheGrid[T]) Set(p gruid.Point, v T) {
	i := p.Y*MapWidth + p.X
	if i < 0 || i > len(bs) || p.X < 0 || p.X >= MapWidth {
		return
	}
	bs[i] = v
}

// SetU puts a value at the given position in the grid. It doesn't check
// boundaries.
func (bs CacheGrid[T]) SetU(p gruid.Point, v T) {
	i := p.Y*MapWidth + p.X
	bs[i] = v
}

// New prepares a map-sized grid of zero values. It uses bs if already
// initialized.
func (bs CacheGrid[T]) New() CacheGrid[T] {
	if bs == nil {
		return make(CacheGrid[T], MapWidth*MapHeight)
	}
	clear(bs)
	return bs
}
