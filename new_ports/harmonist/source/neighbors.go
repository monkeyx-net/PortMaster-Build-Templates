package main

import (
	"iter"
	"slices"

	"codeberg.org/anaseto/gruid"
)

// CardinalNeighbors returns cardinal neighbors as a slice.
func (g *Game) CardinalNeighbors(p gruid.Point) []gruid.Point {
	return slices.Collect(Neighbors(p))
}

// PlayerPassableNeighbors returns all neighbors (including diagonal ones) that
// are passable for the player.
func (g *Game) PlayerPassableNeighbors(p gruid.Point) []gruid.Point {
	var ps []gruid.Point
	for q := range AllNeighbors(p) {
		if InMap(q) && IsPlayerPassable(g.Map.At(q)) {
			ps = append(ps, q)
		}
	}
	return ps
}

// HasNonWallNeighbors returns all non-wall cardinal neighbors as a slice.
func (g *Game) HasNonWallNeighbors(p gruid.Point) bool {
	for range NeighborsFunc(p, func(q gruid.Point) bool {
		return InMap(q) && g.Map.At(q) != WallCell
	}) {
		return true
	}
	return false
}

// HasNonWallExploredNeighbor reports whether the given point is at the
// exploration frontier.
func (g *Game) HasNonWallExploredNeighbor(p gruid.Point) bool {
	for range NeighborsFunc(p, func(q gruid.Point) bool {
		if !InMap(q) {
			return false
		}
		kt := g.Map.KnownTerrain.At(q)
		return IsKnown(kt) && kt != WallCell
	}) {
		return true
	}
	return false
}

// FlammableNeighbors returns all flammable neighbors.
func (g *Game) FlammableNeighbors(p gruid.Point) []gruid.Point {
	return slices.Collect(NeighborsFunc(p, func(q gruid.Point) bool {
		return InMap(q) && Flammable(g.Map.At(q))
	}))
}

// RandomNeighbor returns a random neighbor with some bias toward horizontal
// directions.
func RandomNeighbor(p gruid.Point) gruid.Point {
	switch RandInt(6) {
	case 0, 1:
		return p.Add(gruid.Point{1, 0})
	case 2, 3:
		return p.Add(gruid.Point{-1, 0})
	case 4:
		return p.Add(gruid.Point{0, 1})
	default:
		return p.Add(gruid.Point{0, -1})
	}
}

// Neighbors returns an iterator over cardinal neighbors of the given position.
func Neighbors(p gruid.Point) iter.Seq[gruid.Point] {
	return func(yield func(gruid.Point) bool) {
		for i := -1; i <= 1; i += 2 {
			q := p.Shift(i, 0)
			if InMap(q) && !yield(q) {
				return
			}
			q = p.Shift(0, i)
			if InMap(q) && !yield(q) {
				return
			}
		}
	}
}

// AllNeighbors returns an iterator over cardinal and diagonal neighbors of the
// given position.
func AllNeighbors(p gruid.Point) iter.Seq[gruid.Point] {
	return func(yield func(gruid.Point) bool) {
		for y := -1; y <= 1; y++ {
			for x := -1; x <= 1; x++ {
				if x == 0 && y == 0 {
					continue
				}
				q := p.Shift(x, y)
				if InMap(q) && !yield(q) {
					return
				}
			}
		}
	}
}

// NeighborsFunc returns an iterator over cardinal neighbors of the given
// position that statisfy the given predicate.
func NeighborsFunc(p gruid.Point, f func(gruid.Point) bool) iter.Seq[gruid.Point] {
	return func(yield func(gruid.Point) bool) {
		for i := -1; i <= 1; i += 2 {
			q := p.Shift(i, 0)
			if f(q) && !yield(q) {
				return
			}
			q = p.Shift(0, i)
			if f(q) && !yield(q) {
				return
			}
		}
	}
}
