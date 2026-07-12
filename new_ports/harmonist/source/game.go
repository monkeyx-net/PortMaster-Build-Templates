package main

import (
	"log"
	"math/rand/v2"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

var Version string = "v1.0.2"

// Game contains the game logic's state, without ui stuff. Everything could be
// in the model struct instead, with only the game logic's fiend exported, as
// some game functions need the model anyway (like animations), but this allows
// to differentiate a bit things that are mainly game-logic from the stuff that
// is more about ui.
//
// NOTE: the handling of map entities is a bit ad hoc and not extensible, but
// it'd be a lot of work to update it to use Shamogu's approach.
type Game struct {
	Map              *Map                 // terrain, cloud, noise map data
	Player           *Player              // player entity
	Monsters         []*Monster           // monster entities
	MonstersPosCache []int                // monster (dungeon index + 1) / no monster (0)
	MonsterBands     []bandInfo           // monster bands
	Objects          Objects              // object entities
	KnownMonster     map[gruid.Point]int  // monster memory cache by position
	MonsterFOV       map[gruid.Point]bool // whether a position is in monster view
	MonsterTargFOV   map[gruid.Point]bool // whether a position is in target monster view
	LightFOV         *rl.FOV              // light map

	Events            *rl.EventQueue // event queue
	LiberatedArtifact bool           // whether we got the artifact
	LiberatedShaedra  bool           // whether Shaedra was liberated
	Turn              int            // current turn

	Logs        []logEntry  // log entries
	LogIndex    int         // current log entry index
	LogNextTick int         // log entry index at turn start
	Params      startParams // various procinfo params
	Places      places      // special place placements
	ProcInfo    *ProcInfo   // procedural generation information
	Stats       Stats       // game statistics
	Version     string      // game version
	Wizard      wizardMode  // wizard mode info

	PR      *paths.PathRange // path range for pathfinding
	PRnoise *paths.PathRange // path range for noise
	PRanim  *paths.PathRange // path range for noise
	PRauto  *paths.PathRange // path range for auto-explore

	autosources []gruid.Point        // cache for auto-explore sources
	highlight   map[gruid.Point]bool // highlighted positions (e.g. targeted ray)
	md          *model               // reference to the UI model
	mfov        *rl.FOV              // monster field of view
	rand        *rand.Rand

	autoDir               gruid.Point  // direction for auto-movement
	autoDirChanged        bool         // whether direction changed during auto-movement
	autoDirNeighbors      dirNeighbors // configuration for auto-movement in direction
	autoHalt              bool         // whether auto-movement needs to stop
	autoTarget            gruid.Point  // target for auto-movement
	autoexploreMapRebuild bool         // whether auto-explore maps needs rebuilding
	autoexploring         bool         // whether currently auto-exploring
	resting               bool         // whether currently resting
	restingTurns          int          // number of turns resting
	win                   bool         // whether the player already won
}

// ProcInfo contains procgen information for specific map levels.
type ProcInfo struct {
	GeneratedLore    map[int]bool             // whether there's a lore text
	GeneratedMagaras []magaraKind             // magara kind
	GeneratedCloaks  []ItemKind               // cloak kind
	GeneratedAmulets []ItemKind               // amulet kind
	GenPlan          [MaxDepth + 1]genFlavour // cloak/amulet generation
}

// specialEvent represents special events that may happen in some map levels.
type specialEvent int

const (
	NormalLevel specialEvent = iota
	UnstableLevel
	EarthquakeLevel
	MistLevel
)

const spEvMax = int(MistLevel)

// startParams represents various procinfo informations for specific levels.
// TODO: should probably be merged with ProcInfo.
type startParams struct {
	Lore         map[int]bool
	Blocked      map[int]bool
	Special      []specialVault
	Event        map[int]specialEvent
	Windows      map[int]bool
	Trees        map[int]bool
	Holes        map[int]bool
	Stones       map[int]bool
	Tables       map[int]bool
	NoMagara     map[int]bool
	FakeStair    map[int]bool
	ExtraBanana  map[int]int
	HealthPotion map[int]bool
	CrazyImp     int
}

// wizardMode represents the various wizard modes.
type wizardMode int

const (
	WizardNone wizardMode = iota
	WizardNormal
	WizardMap
	WizardSeeAll
)

// FreePassableCell returns a free passable cell without any actors.
// Also avoids some special locations like stairs, bananas, magical stones.
func (g *Game) FreePassableCell() gruid.Point {
	d := g.Map
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("FreePassableCell")
		}
		x := RandInt(MapWidth)
		y := RandInt(MapHeight)
		p := gruid.Point{x, y}
		t := d.At(p)
		if !Passable(t) || t == StairCell || t == StoneCell || t == BananaCell {
			continue
		}
		if g.Player != nil && g.Player.P == p {
			continue
		}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		return p
	}
}

const (
	MaxDepth  = 11                   // maximum map level
	WinDepth  = 8                    // shaedra's map level
	MapHeight = 21                   // map height
	MapWidth  = 80                   // map width
	MapSize   = MapWidth * MapHeight // map size (in tiles)
)

// GenMap generates the map.
func (g *Game) GenMap() {
	ml := AutomataCave
	switch g.Map.Depth {
	case 2, 6, 7:
		ml = RandomWalkCave
		if RandInt(3) == 0 {
			ml = NaturalCave
		}
	case 4, 10, 11:
		ml = RandomWalkTreeCave
		if RandInt(4) == 0 && g.Map.Depth < 11 {
			ml = RandomSmallWalkCaveUrbanised
		} else if g.Map.Depth == 11 && RandInt(2) == 0 {
			ml = RandomSmallWalkCaveUrbanised
		}
	case 9:
		switch RandInt(4) {
		case 0:
			ml = NaturalCave
		case 1:
			ml = RandomWalkCave
		}
	default:
		if RandInt(10) == 0 {
			ml = RandomSmallWalkCaveUrbanised
		} else if RandInt(10) == 0 {
			ml = NaturalCave
		}
	}
	g.GenMapWithLayout(ml)
}

// InitPlayer initializes the player, including magaras and inventory.
func (g *Game) InitPlayer() {
	g.Player = &Player{
		HP:      DefaultHealth,
		Bananas: 1,
	}
	g.Player.AFOV = map[gruid.Point]bool{}
	g.Player.Statuses = map[status]int{}
	g.Player.Expire = map[status]int{}
	g.Player.Magaras = []Magara{
		{},
		{},
		{},
		{},
	}
	g.Player.MP = g.Player.MPMax()
	g.ProcInfo.GeneratedMagaras = []magaraKind{}
	g.Player.Magaras[0] = g.RandomStartingMagara()
	g.ProcInfo.GeneratedMagaras = append(g.ProcInfo.GeneratedMagaras, g.Player.Magaras[0].Kind)
	g.Player.Inventory.Misc = Item{Kind: MarevorMagara}
	g.Player.FOV = rl.NewFOV(visionRange(g.Player.P, TreeRange))
	// Testing
	// g.Player.Inventory.Body = Item{Kind: CloakHear}
	// g.Player.Inventory.Neck = Item{Kind: AmuletDazing}
	// g.Player.Magaras[1] = Magara{Kind: SwiftnessMagara, Charges: 10}
	// g.Player.Magaras[2] = Magara{Kind: NoiseMagara, Charges: 10}
	// g.Player.Magaras[3] = Magara{Kind: DelayedOricExplosionMagara, Charges: 10}
}

// genFlavour represents special map flavours (amulet, cloak).
type genFlavour int

const (
	GenNothing genFlavour = iota
	GenAmulet             // generate an amulet
	GenCloak              // generate a cloak
)

// InitFirstLevel performs initialization for the first map level, including
// initializing procinfo for the whole dungeon.
func (g *Game) InitFirstLevel() {
	g.Version = Version
	g.Map.Depth++ // start at 1
	g.InitPlayer()
	g.autoTarget = invalidPos
	g.ProcInfo.GeneratedLore = map[int]bool{}
	g.Stats.KilledMons = map[monsterKind]int{}
	g.Stats.UsedMagaras = map[magaraKind]int{}
	g.Stats.Achievements = map[Achievement]int{}
	g.Stats.Lore = map[int]bool{}
	g.Stats.Statuses = map[status]int{}
	g.ProcInfo.GenPlan = [MaxDepth + 1]genFlavour{
		1:  GenNothing,
		2:  GenCloak,
		3:  GenNothing,
		4:  GenAmulet,
		5:  GenNothing,
		6:  GenCloak,
		7:  GenNothing,
		8:  GenAmulet,
		9:  GenNothing,
		10: GenCloak,
		11: GenNothing,
	}
	g.Params.Lore = map[int]bool{}
	putRandomLevels(g.Params.Lore, 8)
	g.Params.HealthPotion = map[int]bool{}
	putRandomLevels(g.Params.HealthPotion, 5)
	g.Params.Blocked = map[int]bool{}
	if RandInt(10) > 0 {
		g.Params.Blocked[2+RandInt(WinDepth-2)] = true
	}
	if RandInt(10) == 0 {
		// a second one sometimes!
		g.Params.Blocked[2+RandInt(WinDepth-2)] = true
	}
	g.Params.Special = []specialVault{
		noSpecialVault, // unused (depth 0)
		noSpecialVault,
		noSpecialVault,
		vaultMilfids,
		vaultCelmists,
		vaultVampires,
		vaultHarpies,
		vaultTreeMushrooms,
		vaultShaedra,
		vaultCelmists,
		vaultMirrorSpecters,
		vaultArtifact,
	}
	if RandInt(2) == 0 {
		g.Params.Special[5] = vaultNixes
	}
	if RandInt(4) == 0 {
		if g.Params.Special[5] == vaultNixes {
			g.Params.Special[9] = vaultVampires
		} else {
			g.Params.Special[9] = vaultNixes
		}
	}
	if RandInt(4) == 0 {
		if RandInt(2) == 0 {
			g.Params.Special[3] = vaultFrogs
		} else {
			g.Params.Special[7] = vaultFrogs
		}
	}
	if RandInt(4) == 0 {
		g.Params.Special[10], g.Params.Special[5] = g.Params.Special[5], g.Params.Special[10]
	}
	if RandInt(4) == 0 {
		g.Params.Special[6], g.Params.Special[7] = g.Params.Special[7], g.Params.Special[6]
	}
	if RandInt(4) == 0 {
		g.Params.Special[3], g.Params.Special[4] = g.Params.Special[4], g.Params.Special[3]
	}
	g.Params.Event = map[int]specialEvent{}
	for i := range 2 {
		g.Params.Event[2+5*i+RandInt(5)] = specialEvent(1 + RandInt(spEvMax))
	}
	g.Params.Event[2+RandInt(MaxDepth-1)] = NormalLevel
	g.Params.FakeStair = map[int]bool{}
	if RandInt(MaxDepth) > 0 {
		g.Params.FakeStair[2+RandInt(MaxDepth-2)] = true
		if RandInt(MaxDepth) > MaxDepth/2 {
			g.Params.FakeStair[2+RandInt(MaxDepth-2)] = true
			if RandInt(MaxDepth) == 0 {
				g.Params.FakeStair[2+RandInt(MaxDepth-2)] = true
			}
		}
	}
	g.Params.ExtraBanana = map[int]int{}
	// Guaranteed extra banana in late game.
	g.Params.ExtraBanana[6+RandInt(6)]++
	// Extra banana in random level (except first).
	g.Params.ExtraBanana[2+RandInt(10)]++
	for range 2 {
		// Less bananas in early-mid game.
		g.Params.ExtraBanana[1+RandInt(7)]--
	}
	g.Params.Windows = map[int]bool{}
	if RandInt(MaxDepth) > MaxDepth/2 {
		g.Params.Windows[2+RandInt(MaxDepth-1)] = true
		if RandInt(MaxDepth) == 0 {
			g.Params.Windows[2+RandInt(MaxDepth-1)] = true
		}
	}
	g.Params.Holes = map[int]bool{}
	if RandInt(MaxDepth) > MaxDepth/2 {
		g.Params.Holes[2+RandInt(MaxDepth-1)] = true
		if RandInt(MaxDepth) == 0 {
			g.Params.Holes[2+RandInt(MaxDepth-1)] = true
		}
	}
	g.Params.Trees = map[int]bool{}
	if RandInt(MaxDepth) > MaxDepth/2 {
		g.Params.Trees[2+RandInt(MaxDepth-1)] = true
		if RandInt(MaxDepth) == 0 {
			g.Params.Trees[2+RandInt(MaxDepth-1)] = true
		}
	}
	g.Params.Tables = map[int]bool{}
	if RandInt(MaxDepth) > MaxDepth/2 {
		g.Params.Tables[2+RandInt(MaxDepth-1)] = true
		if RandInt(MaxDepth) == 0 {
			g.Params.Tables[2+RandInt(MaxDepth-1)] = true
		}
	}
	g.Params.NoMagara = map[int]bool{}
	g.Params.NoMagara[WinDepth] = true
	g.Params.Stones = map[int]bool{}
	if RandInt(MaxDepth) > MaxDepth/2 {
		g.Params.Stones[2+RandInt(MaxDepth-1)] = true
		if RandInt(MaxDepth) == 0 {
			g.Params.Stones[2+RandInt(MaxDepth-1)] = true
		}
	}
	permi := RandInt(WinDepth - 1)
	switch permi {
	case 0, 1, 2, 3:
		g.ProcInfo.GenPlan[permi+1], g.ProcInfo.GenPlan[permi+2] = g.ProcInfo.GenPlan[permi+2], g.ProcInfo.GenPlan[permi+1]
	}
	if RandInt(4) == 0 {
		g.ProcInfo.GenPlan[6], g.ProcInfo.GenPlan[7] = g.ProcInfo.GenPlan[7], g.ProcInfo.GenPlan[6]
	}
	if RandInt(4) == 0 {
		g.ProcInfo.GenPlan[MaxDepth-1], g.ProcInfo.GenPlan[MaxDepth] = g.ProcInfo.GenPlan[MaxDepth], g.ProcInfo.GenPlan[MaxDepth-1]
	}
	g.Params.CrazyImp = 2 + RandInt(MaxDepth-2)
	g.PR = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	g.PRnoise = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	g.PRanim = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
	g.PRauto = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
}

func putRandomLevels(m map[int]bool, n int) {
	for i := 0; i < n; i++ {
		j := 1 + RandInt(MaxDepth)
		if !m[j] {
			m[j] = true
		} else {
			i--
		}
	}
}

// InitLevelStructures (re-)initializes various level-specific structures.
func (g *Game) InitLevelStructures() {
	g.Map.Noise = map[gruid.Point]int{}
	g.Map.NoiseIllusion = map[gruid.Point]int{}
	g.Map.MagicalBarriers = map[gruid.Point]rl.Cell{}
	g.Map.Clouds = map[gruid.Point]Cloud{}
	g.MonstersPosCache = make([]int, MapSize)
	g.KnownMonster = map[gruid.Point]int{}
	g.MonsterFOV = map[gruid.Point]bool{}
	g.Objects.Magaras = map[gruid.Point]Magara{}
	g.Objects.Lore = map[gruid.Point]int{}
	g.Objects.Items = map[gruid.Point]Item{}
	g.Objects.Scrolls = map[gruid.Point]Scroll{}
	g.Objects.Stairs = map[gruid.Point]Stair{}
	g.Objects.Bananas = make(map[gruid.Point]bool, 2)
	g.Objects.Barrels = map[gruid.Point]bool{}
	g.Objects.Lights = map[gruid.Point]bool{}
	g.Objects.Potions = map[gruid.Point]Potion{}
	g.Stats.AtNotablePos = map[gruid.Point]bool{}
}

var Testing = false

// Init initializes some data structures.
func (g *Game) Init() {
	if g.rand == nil {
		g.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	if g.Map == nil {
		g.Map = &Map{
			Terrain:      rl.NewGrid(MapWidth, MapHeight),
			KnownTerrain: rl.NewGrid(MapWidth, MapHeight),
		}
	}
	if g.ProcInfo == nil {
		g.ProcInfo = &ProcInfo{}
	}
}

// InitLevel initializes states for a new map level.
func (g *Game) InitLevel() {
	// Starting data
	if g.Map.Depth == 0 {
		g.InitFirstLevel()
	}

	g.InitLevelStructures()

	// Dungeon terrain
	g.GenMap()

	// Events
	if g.Map.Depth == 1 {
		g.StoryPrintf("Started with %s", g.Player.Magaras[0])
		g.Events = rl.NewEventQueue()
	} else {
		g.CleanEvents()
		for st := range g.Player.Statuses {
			if st.Clean() {
				g.Player.Statuses[st] = 0
			}
		}
	}
	monsters := make([]*Monster, len(g.Monsters))
	copy(monsters, g.Monsters)
	rand.Shuffle(len(monsters), func(i, j int) {
		monsters[i], monsters[j] = monsters[j], monsters[i]
	})
	for _, m := range monsters {
		g.PushEvent(&MonsterTurnEvent{Index: m.Index}, g.Turn)
	}
	switch g.Params.Event[g.Map.Depth] {
	case UnstableLevel:
		g.LogStyled("Uncontrolled oric magic fills the air on this level.", logSpecial)
		g.StoryPrint("Special event: magically unstable level")
		for range 7 {
			g.PushEvent(&PosEvent{Action: ObstructionProgression},
				g.Turn+DurationObstructionProgression+RandInt(DurationObstructionProgression/2))
		}
	case MistLevel:
		g.LogStyled("The air seems dense on this level.", logSpecial)
		g.StoryPrint("Special event: foggy level")
		for range 20 {
			g.PushEvent(&PosEvent{Action: FogProgression},
				g.Turn+DurationPurpleMistProgression+RandInt(DurationPurpleMistProgression/2))
		}
	case EarthquakeLevel:
		g.PushEvent(&PosEvent{P: gruid.Point{MapWidth/2 - 15 + RandInt(30), MapHeight/2 - 5 + RandInt(10)}, Action: Earthquake},
			g.Turn+10+RandInt(50))
	}

	// initialize FOV
	switch g.Map.Depth {
	case WinDepth:
		g.LogStyled("Finally! Shaedra should be imprisoned somewhere around here.", logSpecial)
	case MaxDepth:
		g.LogStyled("This the bottom floor, you now have to look for the artifact.", logSpecial)
	}
	g.ComputeFOV()
	g.ComputeMonsterFOV()
	if !Testing { // disable when testing
		g.md.updateStatusInfo()
	}
	if g.Map.Depth > 1 {
		g.md.CancelExamine()
	}
	if g.Map.Depth == 1 {
		g.LogStyled("Press SPACE for menu and ? for help. Good luck!", logSpecial)
	}
}

// CleanEvents clears events between map levels.
func (g *Game) CleanEvents() {
	g.Events.Filter(func(ev rl.Event) bool {
		switch ev.(type) {
		case *MonsterTurnEvent, *PosEvent, *MonsterStatusEvent, AbyssFallEvent:
			return false
		default:
			// keep player statuses events
			return true
		}
	})
	// finish current turn's other effects (like status progression)
	turn := g.Turn
	for !g.Events.Empty() {
		ev, r := g.Events.PopR()
		if r == turn {
			e, ok := ev.(Event)
			if ok {
				e.Handle(g)
			}
			continue
		}
		g.Events.PushFirst(ev, r)
		break
	}
	g.Turn++
}

// descendStyle represents the various forms of going to the next map level.
type descendstyle int

const (
	DescendNormal descendstyle = iota
	DescendJump
	DescendFall
)

// Descend goes to the next map level. Returns true on win (last map level).
func (g *Game) Descend(style descendstyle) bool {
	g.LevelStats()
	if g.Stats.DUSpotted[g.Map.Depth] < 3 {
		AchStealthNovice.Get(g)
	}
	if g.Map.Depth >= 3 {
		if g.Stats.DRests[g.Map.Depth] == 0 && g.Stats.DRests[g.Map.Depth-1] == 0 {
			AchInsomniaNovice.Get(g)
		}
	}
	if g.Map.Depth >= 5 {
		if g.Stats.DRests[g.Map.Depth] == 0 && g.Stats.DRests[g.Map.Depth-1] == 0 && g.Stats.DRests[g.Map.Depth-2] == 0 &&
			g.Stats.DRests[g.Map.Depth-3] == 0 {
			AchInsomniaInitiate.Get(g)
		}
	}
	if g.Map.Depth >= 8 {
		if g.Stats.DRests[g.Map.Depth] == 0 && g.Stats.DRests[g.Map.Depth-1] == 0 && g.Stats.DRests[g.Map.Depth-2] == 0 &&
			g.Stats.DRests[g.Map.Depth-3] == 0 && g.Stats.DRests[g.Map.Depth-4] == 0 && g.Stats.DRests[g.Map.Depth-5] == 0 {
			AchInsomniaMaster.Get(g)
		}
	}
	if g.Map.Depth >= 3 {
		if g.Stats.DMagaraUses[g.Map.Depth] == 0 && g.Stats.DMagaraUses[g.Map.Depth-1] == 0 {
			AchAntimagicNovice.Get(g)
		}
	}
	if g.Map.Depth >= 5 {
		if g.Stats.DMagaraUses[g.Map.Depth] == 0 && g.Stats.DMagaraUses[g.Map.Depth-1] == 0 && g.Stats.DMagaraUses[g.Map.Depth-2] == 0 &&
			g.Stats.DMagaraUses[g.Map.Depth-3] == 0 {
			AchAntimagicInitiate.Get(g)
		}
	}
	if g.Map.Depth >= 8 {
		if g.Stats.DMagaraUses[g.Map.Depth] == 0 && g.Stats.DMagaraUses[g.Map.Depth-1] == 0 && g.Stats.DMagaraUses[g.Map.Depth-2] == 0 &&
			g.Stats.DMagaraUses[g.Map.Depth-3] == 0 && g.Stats.DMagaraUses[g.Map.Depth-4] == 0 && g.Stats.DMagaraUses[g.Map.Depth-5] == 0 {
			AchAntimagicMaster.Get(g)
		}
	}
	if g.Map.Depth >= 5 {
		if g.Stats.DUSpotted[g.Map.Depth]+g.Stats.DUSpotted[g.Map.Depth-1]+g.Stats.DUSpotted[g.Map.Depth-2] < 6 {
			AchStealthInitiate.Get(g)
		}
	}
	if g.Map.Depth >= 8 {
		if g.Stats.DUSpotted[g.Map.Depth] < 3 &&
			g.Stats.DUSpotted[g.Map.Depth-1]+g.Stats.DUSpotted[g.Map.Depth-2]+g.Stats.DUSpotted[g.Map.Depth-3] < 6 {
			AchStealthMaster.Get(g)
		}
	}
	t := g.Map.At(g.Player.P)
	if t == StairCell && g.Objects.Stairs[g.Player.P] == WinStair {
		g.StoryPrint("Escaped!")
		g.win = true
		return true
	}
	if style != DescendNormal {
		g.LogStyled("You fall into the abyss. It hurts!", logDamage)
		g.StoryPrint("Fell into the abyss")
		g.md.AbyssFallAnimation()
	} else {
		g.LogStyled("You descend deeper in the dungeon.", logSpecial)
		g.StoryPrint("Descended stairs")
	}
	g.Map.Depth++
	g.InitLevel()
	if err := g.Save(); err != nil {
		log.Printf("error saving game: %v", err)
	}
	return false
}

// EnterWizardMode enters wizard mode.
func (g *Game) EnterWizardMode() {
	g.Wizard = WizardNormal
	g.LogStyled("Wizard mode enabled.", logSpecial)
	g.StoryPrint("Entered wizard mode")
}

// ApplyRest handles hp and mp regeneration.
func (g *Game) ApplyRest() {
	g.Player.HP = g.Player.HPMax()
	g.Player.HPbonus = 0
	g.Player.MP = g.Player.MPMax()
	g.Stats.Rest++
	g.Stats.DRests[g.Map.Depth]++
	g.LogStyled("You feel fresh again after eating banana and sleeping.", logStatusEnd)
	g.StoryPrintf("Rested in barrel (bananas: %d)", g.Player.Bananas)
	if g.Stats.Rest == 10 {
		AchSleepy.Get(g)
	}
}

// AutoPlayer reports whether auto-playing should remain on.
func (g *Game) AutoPlayer() bool {
	switch {
	case g.resting:
		const enoughRestTurns = 25
		if g.restingTurns < enoughRestTurns {
			g.restingTurns++
			return true
		}
		if g.restingTurns >= enoughRestTurns {
			g.ApplyRest()
		}
		g.resting = false
	case g.autoexploring:
		switch {
		case g.autoHalt:
			// stop exploring
		default:
			var n *gruid.Point
			var finished bool
			if g.autoexploreMapRebuild {
				if g.AllExplored() {
					g.Log("You finished exploring.")
					break
				}
				sources := g.AutoexploreSources()
				g.BuildAutoexploreMap(sources)
			}
			n, finished = g.NextAuto()
			if finished {
				n = nil
			}
			if finished && g.AllExplored() {
				g.Log("You finished exploring.")
			} else if n == nil {
				g.Log("You could not safely reach some places.")
			}
			if n != nil {
				again, err := g.PlayerBump(*n)
				if err != nil {
					g.Log(err.Error())
					break
				}
				return !again
			}
		}
		g.autoexploring = false
	case InMap(g.autoTarget):
		if g.MoveToTarget() {
			return true
		}
		g.autoTarget = invalidPos
	case g.autoDir != ZP:
		if g.AutoToDir() {
			return true
		}
		g.autoDir = ZP
	}
	return false
}

// Died reports whether the player died.
func (g *Game) Died() bool {
	if g.Player.HP <= 0 {
		if g.Wizard != WizardNone {
			g.Player.HP = g.Player.HPMax()
			g.LogStyled("*** Wizard mode resurrects you! ***", logSpecial)
			g.StoryPrint("Resurrected by the Wizard!")
		} else {
			g.LevelStats()
			return true
		}
	}
	return false
}

// EndTurn finalizes a turn.
func (g *Game) EndTurn() {
	g.Events.Push(EndTurnEvent{}, g.Turn+DurationTurn)
	for {
		if g.Died() {
			return
		}
		if g.Events.Empty() {
			return
		}
		ev, r := g.Events.PopR()
		g.Turn = r
		switch ev := ev.(type) {
		case AbyssFallEvent:
			depth := g.Map.Depth
			ev.Handle(g)
			if g.Map.Depth > depth {
				return
			}
		case EndTurnEvent:
			g.ProgressNoise()
			return
		case Event:
			ev.Handle(g)
		default:
			log.Printf("bad event: %v", ev)
		}
	}
}

// ProgressNoise decrease noise counters.
func (g *Game) ProgressNoise() {
	for p, n := range g.Map.Noise {
		if n <= 1 {
			delete(g.Map.Noise, p)
		} else {
			g.Map.Noise[p]--
		}
	}
	for p, n := range g.Map.NoiseIllusion {
		if n <= 1 {
			delete(g.Map.NoiseIllusion, p)
		} else {
			g.Map.NoiseIllusion[p]--
		}
	}
}

// CheckForStorySequence checks for Shaedra liberation conditions and
// starts the story sequence if appropriate.
func (g *Game) CheckForStorySequence() {
	if g.Map.Depth != WinDepth || g.LiberatedShaedra {
		return
	}
	p := g.Places.Shaedra
	if Distance(g.Player.P, p) > 1 &&
		Distance(g.Player.P, g.Places.Marevor) > 1 &&
		Distance(g.Player.P, g.Places.Monolith) > 1 {
		// If the player is not adjacent to one of the important
		// places, we cannot free Shaedra yet.
		return
	}
	if !g.Player.Sees(p) {
		return
	}
	if g.Player.P == g.Places.Marevor {
		// If somehow the player got on Marevor's tile, exchange
		// positions of Marevor's tile and the portal's, so that
		// Marevor is visible during story sequence.
		g.Places.Marevor, g.Places.Monolith = g.Places.Monolith, g.Places.Marevor
	}
	// We liberate Shaedra and start story sequence.
	g.LiberatedShaedra = true
	g.LogStyled("Shaedra: “Oh, it’s you, Syu! Let’s flee with Marevor’s magara!”", logSpecial)
	g.LogStyled("[SPACE/ESC to continue]", logConfirm)
	g.md.mode = modeStory
	g.md.confirm = false
	g.md.story = 1 // first part of the story sequence
}

// checks performs some sanity checks.
func (g *Game) checks() {
	if !Testing {
		return
	}
	for _, m := range g.Monsters {
		mons := g.MonsterAt(m.P)
		if !mons.Exists() && m.Exists() {
			log.Printf("does not exist")
			continue
		}
		if mons != m {
			log.Printf("bad monster: %v vs %v", mons.Index, m.Index)
		}
	}
}

// randInt returns a random integer like rand.IntN but returns zero for
// non-positive values.
func (g *Game) randInt(n int) int {
	if n <= 0 {
		return 0
	}
	return g.rand.IntN(n)
}
