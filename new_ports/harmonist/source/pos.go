package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Distance is an abbreviation for paths.DistanceManhattan.
func Distance(from, to gruid.Point) int {
	return paths.DistanceManhattan(from, to)
}

// ZP is the zero value for gruid.Point.
var ZP gruid.Point = gruid.Point{}

// DirString returns the string representation of the given delta direction.
func DirString(dir gruid.Point) (s string) {
	switch dir {
	case ZP:
		s = ""
	case gruid.Point{1, 0}:
		s = "E"
	case gruid.Point{1, -1}:
		s = "NE"
	case gruid.Point{0, -1}:
		s = "N"
	case gruid.Point{-1, -1}:
		s = "NW"
	case gruid.Point{-1, 0}:
		s = "W"
	case gruid.Point{-1, 1}:
		s = "SW"
	case gruid.Point{0, 1}:
		s = "S"
	case gruid.Point{1, 1}:
		s = "SE"
	}
	return s
}

// KeyToDir maps directional action to delta direction.
func KeyToDir(k Action) (p gruid.Point) {
	switch k {
	case ActionW, ActionRunW:
		p = gruid.Point{-1, 0}
	case ActionE, ActionRunE:
		p = gruid.Point{1, 0}
	case ActionS, ActionRunS:
		p = gruid.Point{0, 1}
	case ActionN, ActionRunN:
		p = gruid.Point{0, -1}
	}
	return p
}

// DirNorm returns a normalized direction between two points, so that
// directions that aren't cardinal nor diagonal are transformed into the
// cardinal part (this corresponds to pruned intermediate nodes in diagonal
// jump).
func DirNorm(p, q gruid.Point) gruid.Point {
	dir := q.Sub(p)
	dx := Abs(dir.X)
	dy := Abs(dir.Y)
	dir = gruid.Point{sign(dir.X), sign(dir.Y)}
	switch {
	case dx == dy:
	case dx > dy:
		dir.Y = 0
	default:
		dir.X = 0
	}
	return dir
}

func sign(n int) int {
	var i int
	switch {
	case n > 0:
		i = 1
	case n < 0:
		i = -1
	}
	return i
}

// Idx2Pos transforms a flat index into a map point.
func Idx2Pos(i int) gruid.Point {
	return gruid.Point{i % MapWidth, i / MapWidth}
}

// Pos2Idx transfors a map point into a flat index.
func Pos2Idx(p gruid.Point) int {
	return p.Y*MapWidth + p.X
}

// InMap reports whether the given position is within map boundaries.
func InMap(p gruid.Point) bool {
	return p.Y >= 0 && p.Y < MapHeight && p.X >= 0 && p.X < MapWidth
}

// InVieCone reports whether a given position is visible from another one
// following a direction.
func InViewCone(dir, from, to gruid.Point) bool {
	if to == from || Distance(from, to) <= 1 {
		return true
	}
	d := DirNorm(from, to)
	return d == dir || LeftDir(d) == dir || RightDir(d) == dir
}

// LeftDir returns the direction to the left (counter-clockwise).
func LeftDir(dir gruid.Point) gruid.Point {
	switch {
	case dir.X == 0 || dir.Y == 0:
		return Left(dir, dir)
	default:
		return gruid.Point{(dir.Y + dir.X) / 2, (dir.Y - dir.X) / 2}
	}
}

// RightDir returns the direction to the right (clockwise).
func RightDir(dir gruid.Point) gruid.Point {
	switch {
	case dir.X == 0 || dir.Y == 0:
		return Right(dir, dir)
	default:
		return gruid.Point{(dir.X - dir.Y) / 2, (dir.Y + dir.X) / 2}
	}
}

// Right returns the position to the Right (according to direction).
func Right(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X - dir.Y, p.Y + dir.X}
}

// Left returns the position to the Left (according to direction).
func Left(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X + dir.Y, p.Y - dir.X}
}
