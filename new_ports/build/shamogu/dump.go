package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"runtime/debug"
)

// Stats gathers various game statistics from a run.
type Stats struct {
	Lucky               int            // number of times you were lucky (could die to bad roll but got lucky roll)
	NDeaths             int            // number of monster deaths
	Deaths              map[string]int // monster deaths by name
	Statuses            []int          // times you got each status
	Hits                int            // number of times you hit with a bump-attack
	Misses              int            // number of times you missed a bump-attack
	Hurt                int            // number of times you got hurt
	Damage              int            // total damage you got
	FireClouds          int            // number of spawn fire clouds
	PoisonClouds        int            // number of spawn poison clouds
	SpiritUses          int            // number of ability uses
	Comestibles         map[string]int // number of eaten comestibles (per type)
	EatenComestibles    int            // number of eaten comestibles
	ActivatedMenhirs    int            // number of activated menhirs
	MonsterTrapTriggers int            // number of triggered traps by the monsters
	PlayerTrapTriggers  int            // number of triggered traps by the player
	Digs                int            // number of digs (destroyed walls)
	MapTurns            [MapLevels]int // number of turns per level
	MapExplorePerc      [MapLevels]int // exploration percent per level
	MapDeathPerc        [MapLevels]int // monster death percent per level
	MapDamage           [MapLevels]int // endured damage per level
	MapEatenComestibles [MapLevels]int // eaten comestibles per level
	MapSpiritUses       [MapLevels]int // spirit uses per level
	MapActivatedMenhirs [MapLevels]int // activated menhirs per level
	MapTriggeredTraps   [MapLevels]int // number of triggered traps per level
}

// newStats returns newly initialized structure for statistics.
func newStats() *Stats {
	return &Stats{
		Deaths:      map[string]int{},
		Comestibles: map[string]int{},
		Statuses:    make([]int, NStatuses),
	}
}

// Death registers the death of a monster with given name.
func (gs *Stats) Death(name string) {
	gs.NDeaths++
	gs.Deaths[name]++
}

// LevelStats collects stats about current level (just before leaving).
func (g *Game) LevelStats() {
	gs := g.Stats
	// Explored percentage.
	mp := &MapPath{passable: g.Map.Passable}
	n, total := 0, 0
	for p, t := range g.Map.KnownTerrain.All() {
		if !g.Map.Passable(p) && len(mp.Neighbors(p)) == 0 {
			continue
		}
		total++
		if IsKnown(t) {
			n++
		}
	}
	gs.MapExplorePerc[g.Map.Level-1] = int(100 * float64(n) / float64(total))
	// Death percentage.
	n, total = 0, 0
	for _, e := range g.NPMapEntities() {
		if a, ok := e.Role.(*Actor); ok {
			total++
			if a.IsDead() {
				n++
			}
		}
	}
	gs.MapDeathPerc[g.Map.Level-1] = int(100 * float64(n) / float64(total))

}

// DumpSummary produces the game statistics short summary displayed at the end
// of the game.
func (g *Game) DumpSummary() string {
	var sb strings.Builder
	var version string
	info, ok := debug.ReadBuildInfo()
	if ok {
		version = info.Main.Version
	} else {
		version = Version
	}
	fmt.Fprintf(&sb, " ♣ Game Summary — Shamogu %s ♣\n\n", version)
	if g.Wizard.Mode != WizardNone {
		fmt.Fprintf(&sb, "ENABLED WIZARD MODE (%d %s",
			g.Wizard.Resurrections, Countable("resurrection", g.Wizard.Resurrections))
		if g.Wizard.Extra {
			sb.WriteString(" & EXTRA CHEATS")
		}
		sb.WriteString(").\n")
	}
	pa := g.PlayerActor()
	if g.win {
		fmt.Fprintf(&sb, "You destroyed the Orb of Corruption!\n")
	} else if pa.IsDead() {
		fmt.Fprintf(&sb, "You died while exploring level %d of the dungeon.\n", g.Map.Level)
	} else {
		fmt.Fprintf(&sb, "You are exploring level %d of the dungeon.\n", g.Map.Level)
	}
	fmt.Fprintf(&sb, "You spent %d turns in the dungeon.\n", g.Turn)
	fmt.Fprintf(&sb, "Your adventure resulted in %d monster %s.\n",
		g.Stats.NDeaths, Countable("death", g.Stats.NDeaths))
	return sb.String()
}

// Dump produces the game statistics full summary.
func (g *Game) Dump() string {
	var sb strings.Builder
	summary := g.DumpSummary()
	sb.WriteString(summary)
	g.dumpPlayer(&sb)
	sb.WriteString("\nLast messages:\n")
	for _, e := range g.Logs.Entries[max(0, len(g.Logs.Entries)-20):] {
		fmt.Fprintf(&sb, "%s\n", e.dumpString())
	}
	sb.WriteString("\nMap:\n")
	g.dumpDungeon(&sb)
	if g.Stats.NDeaths > 0 {
		sb.WriteString("\nMonster deaths:\n")
		g.dumpKilledMonsters(&sb)
	}
	if g.Stats.EatenComestibles > 0 {
		sb.WriteString("\nEaten comestibles:\n")
		g.dumpEatenComestibles(&sb)
	}
	sb.WriteString("\nStatistics:\n")
	g.dumpStatistics(&sb)
	sb.WriteString("\nTimeline:\n")
	for _, s := range g.Logs.Story {
		sb.WriteString(s)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (g *Game) dumpPlayer(sb *strings.Builder) {
	pa := g.PlayerActor()
	fmt.Fprintf(sb, "\nHP:%d/%d A:%d D:%d",
		pa.HP, pa.GetMaxHP(), pa.GetAttack(), pa.GetDefense())
	for i, turns := range pa.Statuses {
		if turns <= 0 {
			continue
		}
		sb.WriteByte(' ')
		st := Status(i)
		fmt.Fprintf(sb, "%s(%d)", st.Abbr(), turns)
	}
	sb.WriteByte('\n')
	sb.WriteString("\nSpirits:\n")
	for i, sp := range g.PlayerSpirits() {
		ei := g.Entity(i)
		fmt.Fprintf(sb, "- %s (%s %d/%d, used %d %s)\n",
			ei.Name, sp.Ability[sp.Level].Name(),
			sp.Charges, sp.MaxCharges[sp.Level],
			sp.Uses, times(sp.Uses))
	}
	sb.WriteString("\nComestibles:\n")
	for i := range g.PlayerComestibles() {
		ei := g.Entity(i)
		fmt.Fprintf(sb, "- %s\n", ei.Name)
	}
}

func (g *Game) dumpDungeon(sb *strings.Builder) {
	// Entities.
	var ids CacheGrid[ID]
	ids = ids.New()
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		ei := g.Entity(i)
		if a, ok := ei.Role.(*Actor); ok && i != PlayerID && a.IsDead() {
			continue
		}
		ids.Set(ei.P, i)
	}
	// Monster noise.
	var noise CacheGrid[ID]
	noise = noise.New()
	for _, i := range sortedIDs {
		ei := g.Entity(i)
		if ai, ok := ei.Role.(*Actor); ok && ei.Noise && ai.IsAlive() {
			noise.Set(ei.P, i)
			continue
		}
	}
	// Draw map.
	for p := range g.Map.KnownTerrain.Points() {
		if p.X%MapWidth == 0 {
			if p.Y == 0 {
				sb.WriteRune('|')
			} else {
				sb.WriteString("|\n|")
			}
		}
		r := g.Map.RuneAt(p)
		if r == '♫' {
			r = ' '
		}
		if !g.InFOV(p) {
			if i := noise.At(p); i > 0 {
				r = '♫'
			} else if _, ok := g.Map.Noise[p]; ok {
				r = '♪'
			}
			sb.WriteRune(r)
			continue
		}
		if g.Map.Clouds.At(p).Kind != NoCloud && (r == '.' || r == '"' || r == '^') {
			r = '§'
		}
		if i := ids.At(p); i > 0 {
			r = g.Entity(i).Rune
		}
		sb.WriteRune(r)
	}
	sb.WriteString("|\n")
}

func (g *Game) dumpKilledMonsters(sb *strings.Builder) {
	monsters := slices.Sorted(maps.Keys(g.Stats.Deaths))
	for _, mons := range monsters {
		fmt.Fprintf(sb, "- %s: %d\n", mons, g.Stats.Deaths[mons])
	}
}

func (g *Game) dumpEatenComestibles(sb *strings.Builder) {
	comestibles := slices.Sorted(maps.Keys(g.Stats.Comestibles))
	for _, co := range comestibles {
		fmt.Fprintf(sb, "- %s: %d\n", co, g.Stats.Comestibles[co])
	}
}

func (g *Game) dumpStatistics(sb *strings.Builder) {
	// NOTE: if the player dumps game statistics while playing, there are
	// some minor information leaks (like number of destroyed walls or trap
	// triggers including out of view ones you might not still be aware
	// of). Avoiding those is not worth the effort.
	timesPer100 := func(n int) string {
		return fmt.Sprintf("%d %s (%.1f per 100 turns)",
			n, times(n), float64(n)*100/float64(max(1, g.Turn)))
	}
	fmt.Fprintf(sb, "You used spirits %d %s, ate %d %s, activated %d %s.\n",
		g.Stats.SpiritUses, times(g.Stats.SpiritUses),
		g.Stats.EatenComestibles, Countable("comestible", g.Stats.EatenComestibles),
		g.Stats.ActivatedMenhirs, Countable("menhir", g.Stats.ActivatedMenhirs))
	fmt.Fprintf(sb, "You triggered %d %s.\n", g.Stats.PlayerTrapTriggers, Countable("trap", g.Stats.PlayerTrapTriggers))
	fmt.Fprintf(sb, "You hit foes %s.\n", timesPer100(g.Stats.Hits))
	fmt.Fprintf(sb, "You missed foes %s.\n", timesPer100(g.Stats.Misses))
	fmt.Fprintf(sb, "You got hurt %d %s and endured %d damage.\n", g.Stats.Hurt, times(g.Stats.Hurt), g.Stats.Damage)
	fmt.Fprintf(sb, "You were lucky %d %s.\n", g.Stats.Lucky, times(g.Stats.Lucky))
	for i, n := range g.Stats.Statuses {
		st := Status(i)
		if n > 0 {
			fmt.Fprintf(sb, "You were %s %d %s.\n", st, n, times(n))
		}
	}
	fmt.Fprintf(sb, "There were %d fire %s and %d poison %s.\n",
		g.Stats.FireClouds, Countable("cloud", g.Stats.FireClouds),
		g.Stats.PoisonClouds, Countable("cloud", g.Stats.PoisonClouds))
	fmt.Fprintf(sb, "There were %d destroyed walls.\n", g.Stats.Digs)
	fmt.Fprintf(sb, "There were %d traps triggered by monsters.\n", g.Stats.MonsterTrapTriggers)
	// Statistics per map level.
	levels := g.Map.Level - 1
	if g.win || g.PlayerActor().IsDead() {
		levels++
	}
	if levels == 0 {
		return
	}
	sb.WriteByte('\n')
	hfmt := "%-20s"
	fmt.Fprintf(sb, hfmt, "Quantity/Level")
	for i := range levels {
		fmt.Fprintf(sb, " %4d", i+1)
	}
	sb.WriteByte('\n')
	perLevel := func(title string, x []int) {
		fmt.Fprintf(sb, hfmt, title)
		for _, n := range x {
			fmt.Fprintf(sb, " %4d", n)
		}
		sb.WriteByte('\n')
	}
	perLevel("Turns", g.Stats.MapTurns[:levels])
	perLevel("Explored (%)", g.Stats.MapExplorePerc[:levels])
	perLevel("Dead monsters (%)", g.Stats.MapDeathPerc[:levels])
	perLevel("Endured damage", g.Stats.MapDamage[:levels])
	perLevel("Eaten comestibles", g.Stats.MapEatenComestibles[:levels])
	perLevel("Spirit uses", g.Stats.MapSpiritUses[:levels])
	perLevel("Activated menhirs", g.Stats.MapActivatedMenhirs[:levels])
	perLevel("Triggered traps", g.Stats.MapTriggeredTraps[:levels])
}

func times(n int) string {
	if n == 1 {
		return "time"
	}
	return "times"
}
