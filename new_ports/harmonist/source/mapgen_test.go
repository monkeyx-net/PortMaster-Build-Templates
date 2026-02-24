package main

import (
	"bytes"
	"fmt"
	"testing"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

var Rounds = 40

func (m *Map) FreePassableCell() gruid.Point {
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("FreeCell")
		}
		x := RandInt(MapWidth)
		y := RandInt(MapHeight)
		p := gruid.Point{x, y}
		t := m.At(p)
		if Passable(t) {
			return p
		}
	}
}

func (m *Map) connex(pr *paths.PathRange) bool {
	pos := m.FreePassableCell()
	passable := func(p gruid.Point) bool {
		return m.At(p) != WallCell
	}
	cp := NewMappingPath(passable)
	pr.CCMap(cp, pos)
	it := m.Terrain.Iterator()
	for it.Next() {
		if Passable(it.Cell()) && pr.CCMapAt(it.P()) == -1 {
			return false
		}
	}
	return true
}

func (m *Map) String() string {
	b := &bytes.Buffer{}
	it := m.Terrain.Iterator()
	for i := 0; it.Next(); i++ {
		t := it.Cell()
		if i > 0 && i%MapWidth == 0 {
			fmt.Fprint(b, "\n")
		}
		if t == WallCell {
			fmt.Fprint(b, "#")
		} else {
			fmt.Fprint(b, ".")
		}
	}
	return b.String()
}

func TestAutomataCave(t *testing.T) {
	for i := 0; i < Rounds; i++ {
		g := &Game{}
		g.Init()
		g.InitFirstLevel()
		g.InitLevelStructures()
		g.GenMapWithLayout(AutomataCave)
		if !g.Map.connex(g.PR) {
			t.Errorf("Not connex:\n%s\n", g.Map.String())
		}
	}
}

func TestRandomWalkCave(t *testing.T) {
	for i := 0; i < Rounds; i++ {
		g := &Game{}
		g.Init()
		g.InitFirstLevel()
		g.InitLevelStructures()
		g.GenMapWithLayout(RandomWalkCave)
		if !g.Map.connex(g.PR) {
			t.Errorf("Not connex:\n%s\n", g.Map.String())
		}
	}
}

func TestRandomWalkTreeCave(t *testing.T) {
	for i := 0; i < Rounds; i++ {
		g := &Game{}
		g.Init()
		g.InitFirstLevel()
		g.InitLevelStructures()
		g.GenMapWithLayout(RandomWalkTreeCave)
		if !g.Map.connex(g.PR) {
			t.Errorf("Not connex:\n%s\n", g.Map.String())
		}
	}
}

func TestRandomSmallWalkCaveUrbanised(t *testing.T) {
	for i := 0; i < Rounds; i++ {
		g := &Game{}
		g.Init()
		g.InitFirstLevel()
		g.InitLevelStructures()
		g.GenMapWithLayout(RandomSmallWalkCaveUrbanised)
		if !g.Map.connex(g.PR) {
			t.Errorf("Not connex:\n%s\n", g.Map.String())
		}
	}
}

func TestNaturalCave(t *testing.T) {
	for i := 0; i < Rounds; i++ {
		g := &Game{}
		g.Init()
		g.InitFirstLevel()
		g.InitLevelStructures()
		g.GenMapWithLayout(NaturalCave)
		if !g.Map.connex(g.PR) {
			t.Errorf("Not connex:\n%s\n", g.Map.String())
		}
	}
}
