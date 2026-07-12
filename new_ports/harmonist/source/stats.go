package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// Stats gathers various game statistics.
type Stats struct {
	Story             []string             // timeline messages
	Killed            int                  // killed monsters
	NadreExplosions   int                  // number of nadre explosions
	KilledMons        map[monsterKind]int  // killed monsters (by kind)
	Moves             int                  // times you moved
	Waits             int                  // times you waited
	Jumps             int                  // times you jumped over monsters
	WallJumps         int                  // times you jumped off walls
	MagarasUsed       int                  // times you used magaras
	DMagaraUses       [MaxDepth + 1]int    // magaras used per level
	UsedMagaras       map[magaraKind]int   // magaras uses (by kind)
	UsedStones        int                  // times you activated stones
	Damage            int                  // received damage
	DDamage           [MaxDepth + 1]int    // received damage per level
	DExplPerc         [MaxDepth + 1]int    // exploration percentage per level
	DSleepingPerc     [MaxDepth + 1]int    // sleeping monster percentage per level
	Burns             int                  // number of terrain burns
	Digs              int                  // number of digs
	Rest              int                  // times you rested
	DRests            [MaxDepth + 1]int    // times you rested per level
	Turns             int                  // number of turns
	DTurns            [MaxDepth + 1]int    // number of turns per level
	TWounded          int                  // turns wounded
	TMWounded         int                  // turns wounded with monsters in sight
	TMonsFOV          int                  // turns with monsters in sight
	NSpotted          int                  // total number of alerts
	NUSpotted         int                  // total number of unique alerted monsters
	DSpotted          [MaxDepth + 1]int    // alerts per level
	DUSpotted         [MaxDepth + 1]int    // unique alerts per level
	DUSpottedPerc     [MaxDepth + 1]int    // unique alerts percentage per level
	Achievements      map[Achievement]int  // achievements
	AtNotablePos      map[gruid.Point]bool // whether notable position has been reached already
	HarmonicMagUse    int                  // harmonic magara uses
	OricMagUse        int                  // oric magara uses
	FireUse           int                  // fire uses
	DestructionUse    int                  // destruction uses
	OricTelUse        int                  // oric teleport uses
	ClimbedTree       int                  // trees you climbed
	TableHides        int                  // tables you hid under
	HoledWallsCrawled int                  // holed walls you crawled under
	DoorsOpened       int                  // doors you opened
	BarrelHides       int                  // barrels you hid within
	Extinguishments   int                  // campfires you extinguished
	Lore              map[int]bool         // lore you read
	Statuses          map[status]int       // times you got each status
	StolenBananas     int                  // bananas stolen from you
	DrankPotions      int                  // potions you drank
	TimesPushed       int                  // times you were pushed
	TimesBlinked      int                  // times you were blinked away
	TimesBlocked      int                  // times you were blocked by an oric barrier
}

// TurnStats computes some turn-specific stats.
func (g *Game) TurnStats() {
	g.Stats.Turns++
	if g.Map.Depth > 0 && g.Map.Depth <= MaxDepth {
		g.Stats.DTurns[g.Map.Depth]++
	}
	if g.Player.HP < g.Player.HPMax() {
		g.Stats.TWounded++
	}
	if g.MonsterInFOV() != nil {
		g.Stats.TMonsFOV++
		if g.Player.HP < g.Player.HPMax() {
			g.Stats.TMWounded++
		}
	}
}

// LevelStats computes level-specific stats.
func (g *Game) LevelStats() {
	free := 0
	exp := 0
	it := g.Map.Terrain.Iterator()
	for it.Next() {
		t := it.Cell()
		if t == WallCell || t == ChasmCell {
			continue
		}
		free++
		if IsKnown(g.Map.KnownTerrain.At(it.P())) {
			exp++
		}
	}
	g.Stats.DExplPerc[g.Map.Depth] = exp * 100 / free
	if g.Stats.DExplPerc[g.Map.Depth] > 93 {
		AchNoviceExplorer.Get(g)
	}
	if g.Map.Depth >= 5 && g.Stats.DExplPerc[g.Map.Depth] > 93 && g.Stats.DExplPerc[g.Map.Depth-1] > 93 && g.Stats.DExplPerc[g.Map.Depth-2] > 93 {
		AchInitiateExplorer.Get(g)
	}
	if g.Map.Depth >= 8 && g.Stats.DExplPerc[g.Map.Depth] > 93 && g.Stats.DExplPerc[g.Map.Depth-1] > 93 && g.Stats.DExplPerc[g.Map.Depth-2] > 93 &&
		g.Stats.DExplPerc[g.Map.Depth-3] > 93 && g.Stats.DExplPerc[g.Map.Depth-4] > 93 {
		AchMasterExplorer.Get(g)
	}
	nmons := len(g.Monsters)
	kmons := 0
	smons := 0
	for _, mons := range g.Monsters {
		if !mons.Exists() {
			kmons++
			continue
		}
		if mons.State == Resting {
			smons++
		}
	}
	g.Stats.DSleepingPerc[g.Map.Depth] = smons * 100 / nmons
	g.Stats.DUSpottedPerc[g.Map.Depth] = g.Stats.DUSpotted[g.Map.Depth] * 100 / nmons
}

// Achievement represents an achievement.
type Achievement string

// List of achievements.
const (
	NoAchievement          Achievement = "Pitiful Death"
	AchBananaCollector     Achievement = "Banana Collector"
	AchStoneEnthousiast    Achievement = "Magic Stone Enthousiast"
	AchHarmonistNovice     Achievement = "Harmonist Novice"
	AchHarmonistInitiate   Achievement = "Harmonist Initiate"
	AchHarmonistMaster     Achievement = "Harmonist Master"
	AchNoviceOricCelmist   Achievement = "Oric Celmist Novice"
	AchInitiateOricCelmist Achievement = "Oric Celmist Initiate"
	AchMasterOricCelmist   Achievement = "Oric Celmist Master"
	AchUnstealthy          Achievement = "Unstealthy Gawalt"
	AchStealthNovice       Achievement = "Stealth Novice"
	AchStealthInitiate     Achievement = "Stealth Initiate"
	AchStealthMaster       Achievement = "Stealth Master"
	AchPyromancerNovice    Achievement = "Pyromancer Novice"
	AchPyromancerInitiate  Achievement = "Pyromancer Initiate"
	AchPyromancerMaster    Achievement = "Pyromancer Master"
	AchDestructorNovice    Achievement = "Destructor Novice"
	AchDestructorInitiate  Achievement = "Destructor Initiate"
	AchDestructorMaster    Achievement = "Destructor Master"
	AchTeleport            Achievement = "Oric Teleport Maniac"
	AchCloak               Achievement = "Dressed Gawalt"
	AchAmulet              Achievement = "Protective Charm"
	AchRescuedShaedra      Achievement = "Rescued Shaedra"
	AchRetrievedArtifact   Achievement = "Recovered Gem Portal Artifact"
	AchAcrobat             Achievement = "Acrobat"
	AchTree                Achievement = "Tree Climber"
	AchTable               Achievement = "Under Table Gawalt"
	AchHole                Achievement = "Hole Crawler"
	AchDoors               Achievement = "Door Opener"
	AchBarrels             Achievement = "Barrel Enthousiast"
	AchExtinguisher        Achievement = "Light Extinguisher"
	AchLoreStudent         Achievement = "Lore Student"
	AchLoremaster          Achievement = "Loremaster"
	AchNoviceExplorer      Achievement = "Novice Explorer"
	AchInitiateExplorer    Achievement = "Initiate Explorer"
	AchMasterExplorer      Achievement = "Master Explorer"
	AchMurderer            Achievement = "Murderer"
	AchInsomniaNovice      Achievement = "Insomnia Novice"
	AchInsomniaInitiate    Achievement = "Insomnia Initiate"
	AchInsomniaMaster      Achievement = "Insomnia Master"
	AchSleepy              Achievement = "Sleepy Gawalt"
	AchAntimagicNovice     Achievement = "Antimagic Novice"
	AchAntimagicInitiate   Achievement = "Antimagic Initiate"
	AchAntimagicMaster     Achievement = "Antimagic Master"
)

// Get makes you get the achievement.
func (ach Achievement) Get(g *Game) {
	if g.Stats.Achievements[ach] == 0 {
		g.Stats.Achievements[ach] = g.Turn
		g.LogfStyled("Achievement: %s.", logSpecial, ach)
		g.StoryPrintf("Achievement: %s", ach)
	}
}

// AchivementTerrain reports whether reaching that terrain is notable for
// achievement purposes.
func AchivementTerrain(t rl.Cell) bool {
	switch t {
	case TreeCell, TableCell, HoledWallCell, DoorCell, BarrelCell:
		return true
	default:
		return false
	}
}

// Reach handles any achievement for reaching the given position.
func (g *Game) Reach(p gruid.Point) {
	if g.Stats.AtNotablePos[p] {
		return
	}
	g.Stats.AtNotablePos[p] = true
	t := g.Map.At(p)
	switch t {
	case TreeCell:
		g.Stats.ClimbedTree++
		if g.Stats.ClimbedTree == 12 {
			AchTree.Get(g)
		}
	case TableCell:
		g.Stats.TableHides++
		if g.Stats.TableHides == 12 {
			AchTable.Get(g)
		}
	case HoledWallCell:
		g.Stats.HoledWallsCrawled++
		if g.Stats.HoledWallsCrawled == 12 {
			AchHole.Get(g)
		}
	case DoorCell:
		g.Stats.DoorsOpened++
		if g.Stats.DoorsOpened == 100 {
			AchDoors.Get(g)
		}
	case BarrelCell:
		g.Stats.BarrelHides++
		if g.Stats.BarrelHides == 20 {
			AchBarrels.Get(g)
		}
	}
}

// NewDigStats handle digging (caused by the player) for statistics and
// achievement purposes.
func (g *Game) NewDigStats() {
	g.Stats.Digs++
	g.Stats.DestructionUse++
	if g.Stats.DestructionUse == 25 {
		AchDestructorNovice.Get(g)
	}
	if g.Stats.DestructionUse == 100 {
		AchDestructorInitiate.Get(g)
	}
	if g.Stats.DestructionUse == 200 {
		AchDestructorMaster.Get(g)
	}
}
