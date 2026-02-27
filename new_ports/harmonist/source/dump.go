package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"runtime/debug"
)

type statusSlice []status

func (sts statusSlice) Len() int           { return len(sts) }
func (sts statusSlice) Swap(i, j int)      { sts[i], sts[j] = sts[j], sts[i] }
func (sts statusSlice) Less(i, j int) bool { return sts[i] < sts[j] }

type monsSlice []monsterKind

func (ms monsSlice) Len() int      { return len(ms) }
func (ms monsSlice) Swap(i, j int) { ms[i], ms[j] = ms[j], ms[i] }
func (ms monsSlice) Less(i, j int) bool {
	return ms[i].Dangerousness() > ms[j].Dangerousness()
}

// Dump produces the game statistics full summary.
func (g *Game) Dump() string {
	sb := &strings.Builder{}
	g.writeDumpHead(sb)
	fmt.Fprintf(sb, "\nInventory:\n")
	if g.Player.Inventory.Body.Kind != NoItem {
		fmt.Fprintf(sb, "- %s (body)\n", g.Player.Inventory.Body.ShortDesc(g))
	}
	if g.Player.Inventory.Neck.Kind != NoItem {
		fmt.Fprintf(sb, "- %s (neck)\n", g.Player.Inventory.Neck.ShortDesc(g))
	}
	fmt.Fprintf(sb, "\nLast messages:\n")
	for i := len(g.Logs) - 10; i < len(g.Logs); i++ {
		if i >= 0 {
			fmt.Fprintf(sb, "%s\n", g.Logs[i].dumpString())
		}
	}
	fmt.Fprintf(sb, "\nDungeon:\n")
	fmt.Fprintf(sb, "┌%s┐\n", strings.Repeat("─", MapWidth))
	sb.WriteString(g.dumpDungeon())
	fmt.Fprintf(sb, "└%s┘\n", strings.Repeat("─", MapWidth))
	if g.Stats.Killed > 0 {
		fmt.Fprintf(sb, "\n")
		fmt.Fprint(sb, g.dumpedKilledMonsters())
	}
	if g.Stats.MagarasUsed > 0 {
		fmt.Fprintf(sb, "\nMagara uses:\n")
		magaraUses := []string{}
		for mag, n := range g.Stats.UsedMagaras {
			if mag != NoMagara {
				magaraUses = append(magaraUses, fmt.Sprintf("- %s: %d\n", mag, n))
			}
		}
		sort.Strings(magaraUses)
		for _, s := range magaraUses {
			sb.WriteString(s)
		}
	}
	g.dumpStatistics(sb)
	fmt.Fprintf(sb, "\nTimeline:\n")
	fmt.Fprint(sb, g.dumpStory())
	return sb.String()
}

func (g *Game) dumpStatuses() string {
	sts := sort.StringSlice{}
	for st, c := range g.Player.Statuses {
		if c <= 0 {
			continue
		}
		if st.Flag() {
			sts = append(sts, st.Short())
		} else {
			sts = append(sts, fmt.Sprintf("%s(%d)", st.Short(), c))
		}
	}
	sort.Sort(sts)
	if len(sts) == 0 {
		return ""
	}
	return " " + strings.Join(sts, " ")
}

func (g *Game) sortedKilledMonsters() monsSlice {
	var ms monsSlice
	for mk, p := range g.Stats.KilledMons {
		if p == 0 {
			continue
		}
		ms = append(ms, mk)
	}
	sort.Sort(ms)
	return ms
}

func (g *Game) dumpStatistics(sb *strings.Builder) {
	fmt.Fprintf(sb, "\nStatistics:\n")
	fmt.Fprintf(sb, "You evoked magaras %s (%d oric, %d harmonic, %d other).\n",
		times(g.Stats.MagarasUsed), g.Stats.OricMagUse,
		g.Stats.HarmonicMagUse, g.Stats.MagarasUsed-g.Stats.OricMagUse-g.Stats.HarmonicMagUse)
	fmt.Fprintf(sb, "You activated %s.\n", CountNoun("magical stone", g.Stats.UsedStones))
	fmt.Fprintf(sb, "You rested %s.\n", times(g.Stats.Rest))
	fmt.Fprintf(sb, "You drank %s.\n", CountNoun("potion", g.Stats.DrankPotions))
	fmt.Fprintf(sb, "You were spotted by %s, %s.\n", CountNoun("monster", g.Stats.NUSpotted), times(g.Stats.NSpotted))
	fmt.Fprintf(sb, "You endured %d damage.\n", g.Stats.Damage)
	if n := g.Stats.Statuses[StatusConfusion]; n > 0 {
		fmt.Fprintf(sb, "You were confused %s.\n", times(n))
	}
	if n := g.Stats.Statuses[StatusLignification]; n > 0 {
		fmt.Fprintf(sb, "You were lignified %s.\n", times(n))
	}
	if n := g.Stats.Statuses[StatusIlluminated]; n > 0 {
		fmt.Fprintf(sb, "You were illuminated by a harmonic celmist %s.\n", times(n))
	}
	if n := g.Stats.TimesBlocked; n > 0 {
		fmt.Fprintf(sb, "You were blocked by an oric celmist barrier %s.\n", times(n))
	}
	if n := g.Stats.TimesPushed; n > 0 {
		fmt.Fprintf(sb, "You were pushed %s by monsters.\n", times(n))
	}
	if n := g.Stats.TimesBlinked; n > 0 {
		fmt.Fprintf(sb, "You were blinked %d %s by blinking frogs.\n", n, times(n))
	}
	if g.Stats.StolenBananas > 0 {
		fmt.Fprintf(sb, "You were stolen %d bananas by harpies.\n", g.Stats.StolenBananas)
	}
	fmt.Fprintf(sb, "You jumped %s over monsters.\n", times(g.Stats.Jumps))
	fmt.Fprintf(sb, "You jumped %s by propelling yourself against walls.\n", times(g.Stats.WallJumps))
	fmt.Fprintf(sb, "You hid in %s.\n", CountNoun("barrel", g.Stats.BarrelHides))
	fmt.Fprintf(sb, "You crawled through %s.\n", CountNoun("holed wall", g.Stats.HoledWallsCrawled))
	fmt.Fprintf(sb, "You climbed %s.\n", CountNoun("tree", g.Stats.ClimbedTree))
	fmt.Fprintf(sb, "You hid under %s.\n", CountNoun("table", g.Stats.TableHides))
	fmt.Fprintf(sb, "You opened %s.\n", CountNoun("door", g.Stats.DoorsOpened))
	fmt.Fprintf(sb, "You moved %s.\n", times(g.Stats.Moves))
	fmt.Fprintf(sb, "You waited %s.\n", times(g.Stats.Waits))
	fmt.Fprintf(sb, "You extinguished %s.\n", CountNoun("campfire", g.Stats.Extinguishments))
	fmt.Fprintf(sb, "You read %s out of %d.\n", CountNoun("lore message", len(g.Stats.Lore)), len(g.Params.Lore))
	if g.Stats.Burns > 0 {
		fmt.Fprintf(sb, "%s.\n", there("fire", g.Stats.Burns))
	}
	if g.Stats.Digs > 0 {
		fmt.Fprintf(sb, "%s.\n", there("destroyed wall", g.Stats.Digs))
	}
	if g.Stats.NadreExplosions > 0 {
		fmt.Fprintf(sb, "%s.\n", there("nadre explosion", g.Stats.NadreExplosions))
	}
	if g.Stats.Killed > 0 {
		fmt.Fprintf(sb, "You murdered %s.\n", CountNoun("monster", g.Stats.Killed))
	}
	fmt.Fprintf(sb, "You spent %d%% turns wounded.\n", g.Stats.TWounded*100/(g.Stats.Turns+1))
	fmt.Fprintf(sb, "You spent %d%% turns with monsters in sight.\n", g.Stats.TMonsFOV*100/(g.Stats.Turns+1))
	fmt.Fprintf(sb, "You spent %d%% turns wounded with monsters in sight.\n", g.Stats.TMWounded*100/(g.Stats.Turns+1))
	maxDepth := g.Map.Depth
	if g.Player.HP <= 0 || g.win {
		// Player won or died, so include last depth.
		maxDepth++
	}
	fmt.Fprintf(sb, "\n")
	hfmt := "%-21s"
	perLevel := func(title string, x []int) {
		fmt.Fprintf(sb, hfmt, title)
		for i := 1; i < len(x); i++ {
			fmt.Fprintf(sb, " %4d", x[i])
		}
		sb.WriteByte('\n')
	}
	if maxDepth > 1 {
		fmt.Fprintf(sb, hfmt, "Quantity/Depth")
		for i := 1; i < maxDepth; i++ {
			fmt.Fprintf(sb, " %4d", i)
		}
		sb.WriteByte('\n')
		perLevel("Turns", g.Stats.DTurns[:maxDepth])
		perLevel("Explored (%)", g.Stats.DExplPerc[:maxDepth])
		perLevel("Sleeping monsters (%)", g.Stats.DSleepingPerc[:maxDepth])
		perLevel("Alerted monsters (%)", g.Stats.DUSpottedPerc[:maxDepth])
		perLevel("Total alerts", g.Stats.DSpotted[:maxDepth])
		perLevel("Rests", g.Stats.DRests[:maxDepth])
		perLevel("Received damage", g.Stats.DDamage[:maxDepth])
		perLevel("Magara uses", g.Stats.DMagaraUses[:maxDepth])
	}
	fmt.Fprintf(sb, "\nAchievements:\n")
	achvs := []string{}
	for achv := range g.Stats.Achievements {
		achvs = append(achvs, string(achv))
	}
	sort.Strings(achvs)
	for _, achv := range achvs {
		fmt.Fprintf(sb, "- %s (turn %d)\n", achv, g.Stats.Achievements[Achievement(achv)])
	}
}

func (g *Game) dumpStory() string {
	return strings.Join(g.Stats.Story, "\n")
}

func (g *Game) dumpDungeon() string {
	sb := strings.Builder{}
	it := g.Map.KnownTerrain.Iterator()
	i := 0
	for it.Next() {
		p := it.P()
		kt := it.Cell()
		if i%MapWidth == 0 {
			if i == 0 {
				sb.WriteRune('│')
			} else {
				sb.WriteString("│\n│")
			}
		}
		if !IsKnown(kt) {
			sb.WriteRune(' ')
			if i == MapSize-1 {
				sb.WriteString("│\n")
			}
			i++
			continue
		}
		var r rune
		switch kt {
		case WallCell:
			r = '#'
		default:
			switch p {
			case g.Player.P:
				r = '@'
			default:
				r, _ = g.TerrainStyle(kt, p)
				if _, ok := g.Map.Clouds[p]; ok && g.Player.AFOV[p] {
					r = '≡'
				}
				m := g.MonsterAt(p)
				if m.Exists() && g.Player.AFOV[m.P] {
					r = m.Kind.Letter()
				}
			}
		}
		sb.WriteRune(r)
		if i == MapSize-1 {
			sb.WriteString("│\n")
		}
		i++
	}
	return sb.String()
}

func (g *Game) dumpedKilledMonsters() string {
	sb := &strings.Builder{}
	fmt.Fprint(sb, "Killed Monsters:\n")
	ms := g.sortedKilledMonsters()
	for _, mk := range ms {
		fmt.Fprintf(sb, "- %s: %d\n", mk, g.Stats.KilledMons[mk])
	}
	return sb.String()
}

// Simplified dump produces the short summary.
func (g *Game) SimplifedDump(msg string, err error) string {
	sb := &strings.Builder{}
	g.writeDumpHead(sb)
	fmt.Fprintf(sb, "\n")
	if err != nil {
		fmt.Fprintf(sb, "Error writing dump: %v.\n", err)
		log.Printf("error writing dump: %v", err)
	} else {
		fmt.Fprintln(sb, msg)
	}
	return sb.String()
}

func (g *Game) writeDumpHead(sb *strings.Builder) {
	info, ok := debug.ReadBuildInfo()
	var version string
	if ok {
		version = info.Main.Version
	} else {
		version = Version
	}
	fmt.Fprintf(sb, " ♣ Game Summary — Harmonist %s ♣\n\n", version)
	if g.Wizard != WizardNone {
		fmt.Fprintf(sb, "ENABLED WIZARD MODE.\n")
	}
	if g.Player.HP > 0 && g.win {
		fmt.Fprintf(sb, "You escaped from Dayoriah Clan’s domain alive!\n")
	} else if g.Player.HP <= 0 {
		fmt.Fprintf(sb, "You died while exploring depth %d of Dayoriah Clan’s domain.\n", g.Map.Depth)
	} else {
		fmt.Fprintf(sb, "You are exploring depth %d of Dayoriah Clan’s domain.\n", g.Map.Depth)
	}
	if g.LiberatedShaedra {
		fmt.Fprint(sb, "You rescued Shaedra.\n")
	} else {
		fmt.Fprint(sb, "You did not rescue Shaedra.\n")
	}
	if g.LiberatedArtifact {
		fmt.Fprint(sb, "You recovered the Gem Portal Artifact.\n")
	} else {
		fmt.Fprint(sb, "You did not recover the Gem Portal Artifact.\n")
	}
	fmt.Fprintf(sb, "You spent %d turns in Hareka’s Underground.\n", g.Turn)
	fmt.Fprintf(sb, "\n")
	// fmt.Fprintf(sb, "You have %d/%d HP, %d/%d MP and %d/%d bananas.\n", g.Player.HP, g.Player.HPMax(), g.Player.MP, g.Player.MPMax(), g.Player.Bananas, MaxBananas)
	fmt.Fprintf(sb, "HP:%d/%d MP:%d/%d ):%d/%d%s\n", g.Player.HP, g.Player.HPMax(), g.Player.MP, g.Player.MPMax(), g.Player.Bananas, MaxBananas, g.dumpStatuses())
	fmt.Fprintf(sb, "\nMagaras:\n")
	for _, mag := range g.Player.Magaras {
		if mag.Kind != NoMagara {
			fmt.Fprintf(sb, "- %s (charges: %d)\n", mag, mag.Charges)
		}
	}
}

func there(s string, n int) string {
	if n == 1 {
		return "There was " + CountNoun(s, n)
	}
	return "There were " + CountNoun(s, n)
}

func times(n int) string {
	return CountNoun("time", n)
}
