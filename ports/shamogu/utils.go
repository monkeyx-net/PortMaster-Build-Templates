package main

import (
	"iter"

	"codeberg.org/anaseto/gruid"
)

// Directions contains the four cardinal directions in direct order (east,
// north, west, south).
var Directions = []gruid.Point{
	{1, 0},
	{0, -1},
	{-1, 0},
	{0, 1},
}

// toDir normalizes each coordinate.
func toDir(p gruid.Point) gruid.Point {
	if p.X != 0 {
		p.X /= abs(p.X)
	}
	if p.Y != 0 {
		p.Y /= abs(p.Y)
	}
	return p
}

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

// One adds "a" or "an" to a string.
func One(s string) (text string) {
	if len(s) > 0 {
		if s[len(s)-1] == 's' {
			return s
		}
		switch s[0] {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			text = "an " + s
		default:
			text = "a " + s
		}
	}
	return text
}

// Countable adds an "s" depending on quantity n.
func Countable(s string, n int) string {
	if len(s) == 0 || s[len(s)-1] == 's' || n == 1 {
		return s
	}
	return s + "s"
}

// Neighbors returns an iterator over cardinal neighbors of the given position.
func Neighbors(p gruid.Point) iter.Seq[gruid.Point] {
	return func(yield func(gruid.Point) bool) {
		for i := -1; i <= 1; i += 2 {
			q := p.Shift(i, 0)
			if inMap(q) && !yield(q) {
				return
			}
			q = p.Shift(0, i)
			if inMap(q) && !yield(q) {
				return
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

// PassableNeighbors returns an iterator over passable cardinal neighbors of
// the given position.
func (m *Map) PassableNeighbors(p gruid.Point) iter.Seq[gruid.Point] {
	return func(yield func(gruid.Point) bool) {
		for i := -1; i <= 1; i += 2 {
			q := p.Shift(i, 0)
			if m.Passable(q) && !yield(q) {
				return
			}
			q = p.Shift(0, i)
			if m.Passable(q) && !yield(q) {
				return
			}
		}
	}
}

// PassableNeighbor reports whether p has a passable neighbor.
func (m *Map) PassableNeighbor(p gruid.Point) bool {
	for range m.PassableNeighbors(p) {
		return true
	}
	return false
}

// RandomPassableWithoutTrap returns a random floor, rubble or foliage cell in
// the map without any traps. It assumes that such a cell exists (otherwise the
// function does not end).
func (g *Game) RandomPassableWithoutTrap() gruid.Point {
	passable := g.Map.PassableWithoutTraps
	m := g.Map.Terrain
	size := m.Size()
	for {
		p := gruid.Point{g.rand.IntN(size.X), g.rand.IntN(size.Y)}
		if passable(p) {
			return p
		}
	}
}

// IntN is a wrapper around rand.IntN that allows for non-positive values.
func (g *Game) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return g.rand.IntN(n)
}

// IntNBiaseUp is a rand.IntN-like wrapper biased toward high values.
func (g *Game) IntNBiasedUp(n int) int {
	m, n := g.IntN(n), g.IntN(n)
	return max(m, n)
}
