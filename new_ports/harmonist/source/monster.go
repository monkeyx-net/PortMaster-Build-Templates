package main

import (
	"fmt"
	"log"

	"codeberg.org/anaseto/gruid"
)

// Mindstate represents the various monster mindstates.
type Mindstate int

const (
	Resting Mindstate = iota
	Hunting
	Wandering
	Watching
)

func (m Mindstate) String() string {
	var st string
	switch m {
	case Resting:
		st = "resting"
	case Wandering:
		st = "wandering"
	case Hunting:
		st = "hunting"
	case Watching:
		st = "watching"
	}
	return st
}

// monsterStatus represents a monster status effect.
type monsterStatus int

const (
	MonsConfused monsterStatus = iota
	MonsExhausted
	MonsDazed
	MonsSatiated
	MonsLignified
)

// NMonsStatus is the number of distinct monster status effects.
const NMonsStatus = int(MonsLignified) + 1

func (st monsterStatus) String() (text string) {
	switch st {
	case MonsConfused:
		text = "confused"
	case MonsExhausted:
		text = "exhausted"
	case MonsDazed:
		text = "dazed"
	case MonsSatiated:
		text = "satiated"
	case MonsLignified:
		text = "lignified"
	}
	return text
}

func (st monsterStatus) Color() rune {
	switch st {
	case MonsConfused:
		return 'G'
	case MonsDazed:
		return 'C'
	case MonsSatiated:
		return 'B'
	case MonsLignified:
		return 'Y'
	default:
		return 'N'
	}
}

// monsterKind represents the various kinds of monsters.
type monsterKind int

const (
	MonsGuard monsterKind = iota
	MonsYack
	MonsSatowalgaPlant
	MonsMadNixe
	MonsBlinkingFrog
	MonsWorm
	MonsMirrorSpecter
	MonsTinyHarpy
	MonsOricCelmist
	MonsHarmonicCelmist
	MonsDog
	MonsHighGuard
	MonsSpider
	MonsWingedMilfid
	MonsEarthDragon
	MonsAcidMound
	MonsExplosiveNadre
	MonsVampire
	MonsTreeMushroom
	MonsButterfly
	MonsCrazyImp
	MonsHazeCat
)

func (mk monsterKind) String() string {
	return MonsData[mk].name
}

// Letter returns the monster's letter.
func (mk monsterKind) Letter() rune {
	return MonsData[mk].letter
}

// Dangerousness returns the relative danger level (for generation purposes).
func (mk monsterKind) Dangerousness() int {
	return MonsData[mk].dangerousness
}

// Ranged reports whether the monster has a normal ranged attack.
func (mk monsterKind) Ranged() bool {
	switch mk {
	case MonsHighGuard, MonsSatowalgaPlant, MonsMadNixe, MonsVampire, MonsTreeMushroom:
		return true
	default:
		return false
	}
}

// Smiting reports whether the monster has a ranged smiting attack.
func (mk monsterKind) Smiting() bool {
	switch mk {
	case MonsMirrorSpecter, MonsOricCelmist, MonsHarmonicCelmist:
		return true
	default:
		return false
	}
}

// Peaceful reports whether the monster is peaceful.
func (mk monsterKind) Peaceful() bool {
	switch mk {
	case MonsButterfly, MonsEarthDragon, MonsCrazyImp:
		return true
	default:
		return false
	}
}

// GoodSmell reports whether the monster has a good sense of smell.
func (mk monsterKind) GoodSmell() bool {
	switch mk {
	case MonsDog, MonsHazeCat, MonsVampire, MonsExplosiveNadre:
		return true
	default:
		return false
	}
}

// Notable reports whether the monster is notable (for logging purposes).
func (mk monsterKind) Notable() bool {
	switch mk {
	case MonsCrazyImp, MonsEarthDragon, MonsHazeCat:
		return true
	default:
		return false
	}
}

// CanOpenDoors reports whether the monster can open doors.
func (mk monsterKind) CanOpenDoors() bool {
	switch mk {
	case MonsGuard, MonsHighGuard, MonsMadNixe, MonsOricCelmist, MonsHarmonicCelmist, MonsVampire, MonsWingedMilfid:
		return true
	default:
		return false
	}
}

// Patrolling reports whether the monster is a patrolling monster.
func (mk monsterKind) Patrolling() bool {
	switch mk {
	case MonsGuard, MonsHighGuard, MonsMadNixe, MonsOricCelmist, MonsHarmonicCelmist:
		return true
	default:
		return false
	}
}

// CanFly reports whether the monster can fly.
func (mk monsterKind) CanFly() bool {
	switch mk {
	case MonsWingedMilfid, MonsMirrorSpecter, MonsButterfly, MonsTinyHarpy:
		return true
	default:
		return false
	}
}

// CanSwim reports whether the monster can swim.
func (mk monsterKind) CanSwim() bool {
	switch mk {
	case MonsBlinkingFrog, MonsVampire, MonsDog:
		return true
	default:
		return false
	}
}

// CanAttackOnTree reports whether the monster can attack a player that's on
// top of a banana tree.
func (mk monsterKind) CanAttackOnTree() bool {
	switch {
	case mk.Size() == MonsLarge:
		return true
	case mk.CanFly():
		return true
	case mk == MonsBlinkingFrog || mk == MonsHazeCat:
		return true
	default:
		return false
	}
}

// ShallowSleep reports whether the monster has a shallow sleep (naturally
// wakes up faster).
func (mk monsterKind) ShallowSleep() bool {
	switch mk {
	case MonsCrazyImp, MonsHazeCat:
		return true
	default:
		return false
	}
}

// ResistsLignification reports whether the monster resists lignification.
func (mk monsterKind) ResistsLignification() bool {
	switch mk {
	case MonsSatowalgaPlant, MonsTreeMushroom:
		return true
	default:
		return false
	}
}

// ReflectsTeleport reports whether the monster reflects any teleport effects.
func (mk monsterKind) ReflectsTeleport() bool {
	switch mk {
	case MonsBlinkingFrog:
		return true
	default:
		return false
	}
}

// Desc returns a description of the monster.
func (mk monsterKind) Desc() string {
	return MonsDesc[mk]
}

// Indefinite returns the name of the monster with an indefinite article
// before.
func (mk monsterKind) Indefinite(capital bool) (text string) {
	text = Indefinite(mk.String(), capital)
	return text
}

// Definite returns the name of the monster with a definite article before.
func (mk monsterKind) Definite(capital bool) (text string) {
	if capital {
		text = fmt.Sprintf("The %s", mk.String())
	} else {
		text = fmt.Sprintf("the %s", mk.String())
	}
	return text
}

// Size returns the size of the monster.
func (mk monsterKind) Size() monsize {
	return MonsData[mk].size
}

// monsize represents the possible monster sizes.
type monsize int

const (
	MonsSmall monsize = iota
	MonsMedium
	MonsLarge
)

func (ms monsize) String() (text string) {
	switch ms {
	case MonsSmall:
		text = "small"
	case MonsMedium:
		text = "average"
	case MonsLarge:
		text = "large"
	}
	return text
}

// monsterData gathers information about a monster kind.
type monsterData struct {
	size          monsize // monster size (used in some size checks)
	letter        rune    // monster letter
	name          string  // monster name
	dangerousness int     // amount of danger it represents (for monster generation)
}

// MonsData gathers data about the various monster kinds.
var MonsData = []monsterData{
	MonsGuard:           {MonsMedium, 'g', "guard", 3},
	MonsTinyHarpy:       {MonsSmall, 't', "tiny harpy", 3},
	MonsOricCelmist:     {MonsMedium, 'o', "oric celmist", 9},
	MonsHarmonicCelmist: {MonsMedium, 'h', "harmonic celmist", 9},
	MonsWorm:            {MonsSmall, 'w', "farmer worm", 4},
	MonsAcidMound:       {MonsSmall, 'a', "acid mound", 4},
	MonsDog:             {MonsMedium, 'd', "dog", 5},
	MonsYack:            {MonsMedium, 'y', "yack", 5},
	MonsHighGuard:       {MonsMedium, 'G', "high guard", 5},
	MonsSpider:          {MonsSmall, 's', "spider", 15},
	MonsWingedMilfid:    {MonsMedium, 'W', "winged milfid", 6},
	MonsBlinkingFrog:    {MonsMedium, 'F', "blinking frog", 6},
	MonsEarthDragon:     {MonsLarge, 'D', "earth dragon", 18},
	MonsMirrorSpecter:   {MonsMedium, 'm', "mirror specter", 11},
	MonsExplosiveNadre:  {MonsMedium, 'n', "explosive nadre", 8},
	MonsSatowalgaPlant:  {MonsLarge, 'P', "satowalga plant", 7},
	MonsMadNixe:         {MonsMedium, 'N', "mad nixe", 14},
	MonsVampire:         {MonsMedium, 'V', "vampire", 13},
	MonsTreeMushroom:    {MonsLarge, 'T', "tree mushroom", 17},
	MonsButterfly:       {MonsSmall, 'b', "kerejat", 2},
	MonsCrazyImp:        {MonsSmall, 'i', "Crazy Imp", 19},
	MonsHazeCat:         {MonsSmall, 'c', "haze cat", 16},
}

// MonsDesc holds description data for the various kinds of monsters.
var MonsDesc = []string{
	MonsGuard:           "Guards are low rank soldiers who patrol between Dayoriah Clan’s buildings.",
	MonsTinyHarpy:       "Tiny harpies are little humanoid flying creatures. They are aggressive when hungry, but peaceful when satiated. This Underground harpy species eats fruits (including bananas) and other vegetables.",
	MonsOricCelmist:     "Oric celmists are mages that can create magical barriers in cells adjacent to you, complicating your escape.\n\nDayoriah Clan’s oric celmists are famous for their knowledge of oric magic force manipulations. According to Marevor, they plan on doing some dangerous oric experiments with the Gem Portal Artifact, though that’s all you can say about it, because his boring explanations were a bit over your head.",
	MonsHarmonicCelmist: "Harmonic celmists are mages specialized in manipulation of sound and light. They can illuminate you with harmonic light, making it more difficult to hide from them. They also use alert harmonic sounds around you.\n\nAlthough harmonies are often considered as less prestigious magic energies than oric ones, the Dayoriah Clan knows how to make good use of them, as they clearly showed when they stole Marevor’s Gem Portal Artifact.",
	MonsWorm:            "Farmer worms are ugly creeping creatures. They furrow as they move, helping new foliage to grow.",
	MonsAcidMound:       "Acid mounds are acidic creatures. They can corrode your magaras, reducing their number of charges.",
	MonsDog:             "Dogs are carnivore quadrupeds. They can bark, and smell you from up to 5 tiles away when hunting or watching for you.",
	MonsYack:            "Yacks are quite large herbivorous quadrupeds. They tend to eat grass peacefully, but upon seeing you they may attack, pushing you up to 5 cells away.",
	MonsHighGuard:       "High guards watch over a particular location. They can throw javelins.",
	MonsSpider:          "Spiders are small creatures, with panoramic vision and whose bite can confuse you.",
	MonsWingedMilfid:    "Winged milfids are  humanoids that can fly over you and make you swap positions. They tend to be very aggressive creatures.",
	MonsBlinkingFrog:    "Blinking frogs are big frog-like creatures, whose bite can make you blink away. The science behind their attack is not clear, but many think it relies on some kind of oric deviation magic. They can jump to attack from below.",
	MonsEarthDragon:     "Earth dragons are big creatures from a dragon species that wander in the Underground. They are peaceful creatures, but they may hurt you inadvertently, pushing you up to 6 tiles away (3 if confused). They naturally emit powerful oric energies, allowing them to eat rocks and dig tunnels. Their oric energies can confuse you if you’re close enough, for example if they hurt you or you jump over them.",
	MonsMirrorSpecter:   "Mirror specters are very insubstantial creatures, which can absorb your mana.",
	MonsExplosiveNadre:  "Nadres are dragon-like biped creatures that are famous for exploding upon dying. Explosive nadres are a tiny nadre race that explode upon attacking. The explosion confuses any adjacent creatures and occasionally destroys walls.",
	MonsSatowalgaPlant:  "Satowalga Plants are immobile bushes that throw viscous acidic projectiles at you, destroying some of your magara charges. They attack at half normal speed.",
	MonsMadNixe:         "Nixes are magical humanoids. Usually, they specialize in illusion harmonic magic, but the so called mad nixes are a perverted variant who learned the oric arts to create a spell that can attract their foes to them, so that they can kill them without pursuing them.",
	MonsVampire:         "Vampires are humanoids that drink blood to survive. Their nauseous spitting can cause confusion, impeding the use of magaras for a few turns.",
	MonsTreeMushroom:    "Tree mushrooms are big clunky creatures. They can throw lignifying spores at you, leaving you unable to move for a few turns, though the spores will also provide some protection against harm.",
	MonsButterfly:       "Underground’s butterflies, called kerejats, wander peacefully around, illuminating their surroundings. They stop doing so when confused or dazed.",
	MonsCrazyImp:        "Crazy Imp is a crazy creature that likes to sing with its small guitar. It seems to be fond of monkeys and quite capable at finding them by smell. While singing it may attract unwanted attention.",
	MonsHazeCat:         "Haze cats are a special variety of cats found in the Underground. They have very good night vision and are always alert.",
}

// Monster represents a monster's state and characteristics.
type Monster struct {
	Kind           monsterKind          // monster kind
	Band           int                  // index of band it belongs
	Index          int                  // index in g.Monsters
	Dir            gruid.Point          // direction (for monster fov)
	Dead           bool                 // whether the monster is dead (abyss or water swap)
	State          Mindstate            // mindstate
	Statuses       [NMonsStatus]int     // monster statuses
	P              gruid.Point          // current position
	LastKnownPos   gruid.Point          // known position (for the player)
	Target         gruid.Point          // target destination for pathfinding
	Path           []gruid.Point        // cache
	Seen           bool                 // already seen by the player
	Noise          bool                 // made footstep noise on last turn
	FOV            map[gruid.Point]bool // field of view of the monster
	LastKnownState Mindstate            // last known mindstate (for out-of-view resting)
	Swapped        bool                 // just swapped by another monster
	Watching       int                  // turns watching
	Left           bool                 // turned left
	Search         gruid.Point          // position to search around
	Alerted        bool                 // monster is alert
	Waiting        int                  // turns waiting
}

// Initialize monster.
func (m *Monster) Init() {
	m.P = invalidPos
	m.FOV = make(map[gruid.Point]bool)
	m.LastKnownPos = invalidPos
	m.Search = invalidPos
	if RandInt(2) == 0 {
		m.Left = true
	}
	switch m.Kind {
	case MonsButterfly:
		m.MakeWander()
	case MonsSatowalgaPlant:
		m.StartWatching()
	}
}

// CanPass report whether the monster can pass through the given position.
func (m *Monster) CanPass(g *Game, p gruid.Point) bool {
	if !InMap(p) {
		return false
	}
	if mons := g.MonsterAt(p); mons.Exists() && (mons.Kind == MonsSatowalgaPlant || mons.Status(MonsLignified)) {
		// Monsters cannot pass through tiles that have a non-moving
		// monster.
		return false
	}
	t := g.Map.At(p)
	return Passable(t) ||
		IsDoorPassable(t) && m.Kind.CanOpenDoors() ||
		IsLevitatePassable(t) && m.Kind.CanFly() ||
		IsSwimPassable(t) && (m.Kind.CanSwim() || m.Kind.CanFly()) ||
		t == HoledWallCell && m.Kind.Size() == MonsSmall
}

// StartWatching makes a monster start watching around.
func (m *Monster) StartWatching() {
	m.State = Watching
	m.Watching = 0
}

// Status reports whether the monster has the given status effect.
func (m *Monster) Status(st monsterStatus) bool {
	return m.Statuses[st] > 0
}

// Exists reports whether the monster exists and is alive.
func (m *Monster) Exists() bool {
	// NOTE: In harmonist, we assume the player always knows whether a
	// known monster is alive or dead, which is ok, because the only source
	// of death for monsters is swapping (and you know when it happens).
	return m != nil && !m.Dead
}

// Alternate makes the monster turn around.
func (m *Monster) Alternate() {
	if m.Left {
		if RandInt(4) > 0 {
			m.Dir = LeftDir(m.Dir)
		} else {
			m.Dir = RightDir(m.Dir)
			m.Left = false
		}
	} else {
		if RandInt(3) > 0 {
			m.Dir = RightDir(m.Dir)
		} else {
			m.Dir = LeftDir(m.Dir)
			m.Left = true
		}
	}
}

// TeleportAway teleports the monster away. Reports whether teleportation
// happened.
func (m *Monster) TeleportAway(g *Game) bool {
	if m.Status(MonsLignified) {
		g.Logf("The lignified %s resists teleportation.", m.Kind)
		return false
	}
	var p gruid.Point
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		p = g.FreePassableCell()
		if Distance(p, m.P) < 15 {
			continue
		}
		if m.Kind == MonsSatowalgaPlant && count < maxIterations/2 && g.Map.HasTooManyWallNeighbors(p) {
			continue
		}
		break
	}

	switch m.State {
	case Hunting: // XXX: change the target or state?
	case Resting, Wandering:
		m.MakeWander()
		m.Target = m.P
	}
	if g.Player.Sees(m.P) {
		g.LogfStyled("%s teleports away.", logNotable, m.Kind.Definite(true))
	}
	opos := m.P
	m.MoveTo(g, p)
	if g.Player.Sees(opos) {
		g.md.TeleportAnimation(opos, p, false)
	}
	return true
}

// MoveTo attemps to move the monster to the given target point.
func (m *Monster) MoveTo(g *Game, p gruid.Point) bool {
	if mons := g.MonsterAt(p); mons != nil && (mons.Kind == MonsSatowalgaPlant || mons.Status(MonsLignified)) {
		// The monster cannot swap places with a monster that cannot
		// move.
		return false
	}
	if g.Player.Sees(p) {
		m.UpdateKnowledge(g, p)
	} else if g.Player.Sees(m.P) {
		if Distance(m.P, p) == 1 {
			// You know the direction, so you know where the
			// monster should be.
			m.UpdateKnowledge(g, p)
		} else {
			delete(g.KnownMonster, m.P)
			m.LastKnownPos = invalidPos
		}
	}
	if !g.Player.Sees(m.P) && g.Player.Sees(p) {
		if !m.Seen {
			m.Seen = true
			g.Logf("%s (%v) comes into view.", m.Kind.Indefinite(true), m.State)
		}
		g.StopAuto()
	}
	recomputeFOV := g.Player.Sees(m.P) && g.Map.At(m.P) == DoorCell ||
		g.Player.Sees(p) && g.Map.At(p) == DoorCell
	m.PlaceAt(g, p)
	if recomputeFOV {
		g.ComputeFOV()
	}
	t := g.Map.At(p)
	if t == ChasmCell && !m.Kind.CanFly() || t == WaterCell && !m.Kind.CanSwim() && !m.Kind.CanFly() {
		m.Dead = true
		g.HandleKill(m)
		if g.Player.Sees(m.P) {
			switch t {
			case ChasmCell:
				g.Logf("%s falls into the abyss.", m.Kind.Definite(true))
			case WaterCell:
				g.Logf("%s drowns.", m.Kind.Definite(true))
			}
		}
	}
	return true
}

// PlaceAtStart sets a monster at the given starting position.
func (m *Monster) PlaceAtStart(g *Game, p gruid.Point) {
	i := Pos2Idx(p)
	if g.MonstersPosCache[i] > 0 {
		log.Printf("used monster starting position: %v", p)
		// should not happen
		return
	}
	m.P = p
	g.MonstersPosCache[i] = m.Index + 1
	npos := m.RandomFreeNeighbor(g)
	if npos != m.P {
		m.Dir = DirNorm(m.P, npos)
	} else {
		m.Dir = gruid.Point{1, 0}
	}
}

// PlaceAt puts a monster at the given position.
func (m *Monster) PlaceAt(g *Game, p gruid.Point) {
	if !InMap(m.P) {
		log.Printf("monster place at: bad position %v", m.P)
		// should not happen
		return
	}
	if p == invalidPos {
		log.Printf("monster new place at invalid position")
		// should not happen
		return
	}
	if p == m.P {
		// should not happen
		log.Printf("monster place at: same position %v", m.P)
		return
	}
	if p == g.Player.P {
		log.Printf("monster place at: player position %v", p)
		// should not happen
		return
	}
	i := Pos2Idx(p)
	j := Pos2Idx(m.P)
	m.Dir = DirNorm(m.P, p)
	mons := g.MonsterAt(p)
	g.MonstersPosCache[i], g.MonstersPosCache[j] = g.MonstersPosCache[j], g.MonstersPosCache[i]
	if mons.Exists() {
		m.P, mons.P = mons.P, m.P
		mons.Swapped = true
	} else {
		m.P = p
	}
	m.Waiting = 0
}

// MonsterAt returns any monster at the given position.
func (g *Game) MonsterAt(p gruid.Point) *Monster {
	pi := Pos2Idx(p)
	if pi < 0 || pi >= len(g.MonstersPosCache) {
		log.Printf("monster at: bad index %v for pos %v", pi, p)
		// should not happen
		return nil
	}
	i := g.MonstersPosCache[pi]
	if i <= 0 {
		return nil
	}
	m := g.Monsters[i-1]
	return m
}

// EndTurn produces the next MonsterTurnEvent.
func (m *Monster) EndTurn(g *Game) {
	g.PushEventD(&MonsterTurnEvent{Index: m.Index}, 1)
}

// AttackAction handles an attack action from the monster.
func (m *Monster) AttackAction(g *Game) bool {
	m.Dir = DirNorm(m.P, g.Player.P)
	switch m.Kind {
	case MonsExplosiveNadre:
		g.StoryPrint("Nadre explosion")
		m.Explode(g)
		return false
	default:
		if g.Player.HasStatus(StatusDispersal) {
			m.Blink(g)
		} else {
			m.HitPlayer(g)
		}
	}
	return true
}

// Explode handles nadre explosions.
func (m *Monster) Explode(g *Game) {
	m.Dead = true
	neighbors := g.CardinalNeighbors(m.P)
	g.Logf("%s blows up!", m.Kind.Definite(true))
	g.LogStyled(g.ExplosionSound(), logNotable)
	g.Stats.NadreExplosions++
	g.md.NadreExplosionAnimation(m.P)
	g.MakeNoise(ExplosionNoise, m.P)
	for _, p := range append(neighbors, m.P) {
		t := g.Map.At(p)
		if Flammable(t) {
			g.Burn(p)
		}
		if g.Player.P == p {
			m.InflictDamage(g)
		} else if mons := g.MonsterAt(p); mons.Exists() && !mons.Status(MonsConfused) {
			mons.PutConfusion(g)
			if mons.State != Hunting && mons.State != Watching {
				mons.StartWatching()
			}
		}
		if IsDiggable(t) {
			g.Map.Set(p, RubbleCell)
			g.Stats.Digs++
			g.Fog(p, 1)
		}
	}
	g.ComputeFOV()
}

// NaturalAwake handles natural monster awakening.
func (m *Monster) NaturalAwake(g *Game) {
	m.Target = m.NextTarget(g)
	switch g.MonsterBands[m.Band].Beh {
	case BehGuard:
		m.StartWatching()
	default:
		m.MakeWander()
	}
	m.GatherBand(g)
}

// RandomFreeNeighbor returns a random free neighbor point.
func (m *Monster) RandomFreeNeighbor(g *Game) gruid.Point {
	p := m.P
	neighbors := [4]gruid.Point{p.Add(gruid.Point{1, 0}), p.Add(gruid.Point{-1, 0}), p.Add(gruid.Point{0, -1}), p.Add(gruid.Point{0, 1})}
	fnb := []gruid.Point{}
	for _, nbpos := range neighbors {
		if !InMap(nbpos) {
			continue
		}
		t := g.Map.At(nbpos)
		if Passable(t) {
			fnb = append(fnb, nbpos)
		}
	}
	if len(fnb) == 0 {
		return m.P
	}
	samedir := fnb[RandInt(len(fnb))]
	for _, p := range fnb {
		// invariant: pos != m.Pos
		if InViewCone(m.Dir, m.P, m.P.Add(p)) {
			samedir = p
			break
		}
	}
	if RandInt(4) > 0 {
		return samedir
	}
	return fnb[RandInt(len(fnb))]
}

// SearchAroundCache is a cache for monster searching.
var SearchAroundCache []gruid.Point

// SearchAround returns a search position.
func (m *Monster) SearchAround(g *Game, p gruid.Point, radius int) gruid.Point {
	dij := &MonPath{g: g, monster: m}
	nodes := g.PR.DijkstraMap(dij, []gruid.Point{p}, radius)
	SearchAroundCache = SearchAroundCache[:0]
	for _, n := range nodes {
		SearchAroundCache = append(SearchAroundCache, n.P)
	}
	if len(SearchAroundCache) > 0 {
		p := SearchAroundCache[RandInt(len(SearchAroundCache))]
		return p
	}
	return invalidPos
}

// NextTarget returns a new target for the monster.
func (m *Monster) NextTarget(g *Game) (p gruid.Point) {
	band := g.MonsterBands[m.Band]
	p = m.P
	switch band.Beh {
	case BehWander:
		if Distance(m.P, band.Path[0]) < 8+RandInt(8) {
			p = m.SearchAround(g, m.P, 4)
			if p != invalidPos {
				break
			}
		}
		if m.Search != invalidPos && RandInt(2) == 0 {
			p = m.SearchAround(g, m.Search, 7)
			if p != invalidPos {
				break
			}
		}
		p = m.SearchAround(g, band.Path[0], 7)
		if p != invalidPos {
			break
		}
		p = band.Path[0]
	case BehExplore:
		if m.Kind.CanOpenDoors() {
			if m.Search != invalidPos && RandInt(4) == 0 {
				p = m.SearchAround(g, m.Search, 7)
			} else {
				p = m.SearchAround(g, p, 5)
			}
			if p != invalidPos {
				break
			}
		}
		p = band.Path[RandInt(len(band.Path))]
	case BehGuard:
		if m.Search != invalidPos && Distance(m.Search, m.P) < 5 && RandInt(2) == 0 {
			p = m.SearchAround(g, m.Search, 3)
			if p != invalidPos {
				break
			}
		}
		p = band.Path[0]
	case BehPatrol:
		if m.Search != invalidPos && RandInt(4) > 0 {
			p = m.SearchAround(g, m.Search, 7)
			if p != m.P && p != invalidPos {
				break
			}
		}
		if band.Path[0] == m.Target {
			p = band.Path[1]
		} else if band.Path[1] == m.Target {
			p = band.Path[0]
		} else if Distance(band.Path[0], m.P) < Distance(band.Path[1], m.P) {
			p = band.Path[0]
			if RandInt(4) == 0 {
				p = band.Path[1]
			}
		} else {
			p = band.Path[1]
			if RandInt(4) == 0 {
				p = band.Path[0]
			}
		}
	case BehCrazyImp:
		path := m.APath(g, m.P, g.Player.P)
		if len(path) == 0 {
			p = m.SearchAround(g, m.P, 3)
		} else {
			p = g.Player.P
		}
	}
	return p
}

// HandleMonsSpecifics handles monster-specific behaviors.
func (m *Monster) HandleMonsSpecifics(g *Game) (done bool) {
	switch m.Kind {
	case MonsSatowalgaPlant:
		switch m.State {
		case Hunting:
			if m.Target != invalidPos && m.Target != m.P {
				m.Dir = DirNorm(m.P, m.Target)
			}
			if !m.SeesPlayer(g) {
				m.StartWatching()
			}
		default:
			if RandInt(4) > 0 {
				m.Alternate()
			}
		}
		// oklob plants are static ranged-only
		return true
	case MonsGuard, MonsHighGuard:
		if m.State != Wandering && m.State != Watching {
			break
		}
		for p, on := range g.Objects.Lights {
			if !on && Distance(p, m.P) <= DefaultMonsterFOVRange {
				m.ComputeFOV(g)
				break
			}
		}
		for p, on := range g.Objects.Lights {
			if !on && p == m.P {
				g.Map.Set(m.P, LightCell)
				g.Objects.Lights[m.P] = true
				if g.Player.Sees(m.P) {
					g.Logf("%s makes a new fire.", m.Kind.Definite(true))
				}
				return true
			} else if !on && m.SeesLight(g, p) {
				m.Target = p
			}
		}
	case MonsCrazyImp:
		if g.Player.Sees(m.P) && RandInt(2) == 0 && !m.Status(MonsConfused) && !m.Status(MonsExhausted) {
			g.LogStyled("Crazy Imp: “♫ larilon, larila ♫ ♪”", logNotable)
			g.MakeNoise(SingingNoise, m.P)
			m.Exhaust(g)
		}
	}
	return false
}

// DogSmellDist represents the distance at which dogs can smell you.
const DogSmellDist = 5

// DisguiseSmellDist represents the distance at which dogs can smell you when
// you're disguised.
const DisguiseSmellDist = 3

// HandleWatching handles watching for a monster.
func (m *Monster) HandleWatching(g *Game) {
	turns := 4
	if m.Kind == MonsHazeCat {
		turns = 3
	}
	if m.Watching+RandInt(2) < turns {
		m.Alternate()
		m.Watching++
		if m.Kind == MonsDog {
			dij := &MonPath{g: g, monster: m}
			g.PR.DijkstraMap(dij, []gruid.Point{m.P}, DogSmellDist)
			if c := g.PR.DijkstraMapAt(g.Player.P); c <= DogSmellDist {
				m.Target = g.Player.P
				m.MakeWander()
			}
		}
	} else {
		// pick a random cell: more escape strategies for the player
		m.Target = m.NextTarget(g)
		switch g.MonsterBands[m.Band].Beh {
		case BehGuard:
			m.Alternate()
			if m.P != m.Target {
				m.MakeWander()
				m.GatherBand(g)
			}
		default:
			m.MakeWander()
			m.GatherBand(g)
		}
	}
}

// ComputePath computes a new path to target for the monster.
func (m *Monster) ComputePath(g *Game) {
	if !(len(m.Path) > 0 && m.Path[len(m.Path)-1] == m.Target && m.Path[0] == m.P) {
		m.Path = m.APath(g, m.P, m.Target)
		if len(m.Path) == 0 && !m.Status(MonsConfused) {
			// if target is not accessible, try free neighbor cells
			for _, npos := range g.PlayerPassableNeighbors(m.Target) {
				m.Path = m.APath(g, m.P, npos)
				if len(m.Path) > 0 {
					m.Target = npos
					break
				}
			}
		}
	}
}

// Peaceful reports whether the monster is currently peaceful (like butterflies
// or dragons, but including also satiated harpies or monsters that cannot see
// through disguise).
func (m *Monster) Peaceful(g *Game) bool {
	if m.Kind.Peaceful() {
		return true
	}
	if m.State != Hunting && g.Player.HasStatus(StatusDisguised) && (!m.Kind.GoodSmell() || Distance(m.P, g.Player.P) > DisguiseSmellDist) {
		return true
	}
	switch m.Kind {
	case MonsTinyHarpy:
		if m.Status(MonsSatiated) || g.Player.Bananas == 0 {
			return true
		}
	}
	return false
}

// HandleEndPath handles behavior at the end of the path.
func (m *Monster) HandleEndPath(g *Game) {
	if len(m.Path) == 0 && m.Search != invalidPos && Distance(m.Search, m.Target) < 10 && m.P != m.Target {
		// the cell where the player was last noticed may not be recheable for the monster
		m.Search = invalidPos
	}
	switch m.State {
	case Wandering, Hunting:
		if !m.Peaceful(g) {
			if !m.SeesPlayer(g) {
				m.StartWatching()
				m.Alternate()
			}
		} else {
			m.Target = m.NextTarget(g)
		}
	}
}

// MakeWanderAt makes a monster wander with the given target.
func (m *Monster) MakeWanderAt(target gruid.Point) {
	m.Target = target
	if m.State == Hunting {
		return
	}
	if m.Kind == MonsSatowalgaPlant {
		m.State = Hunting
	} else {
		m.State = Wandering
	}
}

// MakeWander makes a monster wander (or watch in satowalga's case).
func (m *Monster) MakeWander() {
	if m.Kind == MonsSatowalgaPlant {
		m.State = Watching
	} else {
		m.State = Wandering
	}
}

// HandleMove handles movement for a monster.
func (m *Monster) HandleMove(g *Game) bool {
	target := m.Path[1]
	mons := g.MonsterAt(target)
	monstarget := invalidPos
	if mons.Exists() && len(mons.Path) >= 2 {
		monstarget = mons.Path[1]
	}
	t := g.Map.At(target)
	switch {
	case m.Peaceful(g) && target == g.Player.P:
		switch m.Kind {
		case MonsEarthDragon:
			return m.AttackAction(g)
		default:
			m.Path = m.APath(g, m.P, m.Target)
		}
	case !mons.Exists():
		if m.Kind == MonsEarthDragon && IsDestructible(t) {
			g.Map.Set(target, RubbleCell)
			if t == BarrelCell {
				delete(g.Objects.Barrels, target)
			}
			g.Stats.Digs++
			g.MakeNoise(WallNoise, m.P)
			g.Fog(m.P, 1)
			if Distance(g.Player.P, target) < 12 {
				// XXX use dijkstra distance ?
				switch t {
				case WallCell:
					g.LogfStyled("%s You hear an earth-splitting noise.", logNotable, g.CrackSound())
				case BarrelCell, DoorCell, TableCell:
					g.LogfStyled("%s You hear an wood-splitting noise.", logNotable, g.CrackSound())
				}
				g.StopAuto()
			}
			if m.MoveTo(g, target) {
				m.Path = m.Path[1:]
			}
		} else if !m.CanPass(g, target) {
			m.Path = m.APath(g, m.P, m.Target)
		} else {
			m.InvertFoliage(g)
			if m.MoveTo(g, target) {
				m.Path = m.Path[1:]
			}
		}
	case (mons.P == target && m.P == monstarget || m.Waiting > 5+RandInt(2)) && !mons.Status(MonsLignified):
		target := mons.P
		if m.MoveTo(g, target) {
			m.Path = m.Path[1:]
			if len(mons.Path) > 0 {
				mons.Path = mons.Path[1:]
			}
		}
	case m.State == Hunting && mons.State != Hunting:
		if m.Waiting > 2+RandInt(3) {
			if mons.Peaceful(g) {
				mons.MakeWander()
			} else {
				mons.MakeWanderAt(m.Target)
				mons.GatherBand(g)
			}
		} else {
			m.Path = m.APath(g, m.P, m.Target)
		}
		m.Waiting++
	case !mons.SeesPlayer(g) && mons.State != Hunting:
		if m.Waiting > 1+RandInt(2) && mons.Kind != MonsSatowalgaPlant {
			mons.MakeWanderAt(mons.RandomFreeNeighbor(g))
		} else {
			m.Path = m.APath(g, m.P, m.Target)
		}
		m.Waiting++
	default:
		m.Path = m.APath(g, m.P, m.Target)
		m.Waiting++
	}
	return true
}

// HandleTurn handles a monster's turn and ends it.
func (m *Monster) HandleTurn(g *Game) {
	m.handleTurn(g)
	if m.Exists() {
		m.EndTurn(g)
	}
}

// handleTurn handles a monster's turn.
func (m *Monster) handleTurn(g *Game) {
	if m.Status(MonsDazed) {
		return
	}
	if m.Swapped {
		m.Swapped = false
		m.MakeAware(g)
		return
	}
	ppos := g.Player.P
	mpos := m.P
	if m.MakeAware(g) {
		// Make monsters lose a turn when they notice you, unless
		// you're adjacent to them.
		return
	}
	if m.State == Resting {
		if RandInt(3000) == 0 || m.Kind.ShallowSleep() && RandInt(10) == 0 {
			m.NaturalAwake(g)
		}
		return
	}
	if m.State == Hunting && m.RangedAttack(g) {
		return
	}
	if m.State == Hunting && m.SmitingAttack(g) {
		return
	}
	if m.HandleMonsSpecifics(g) {
		return
	}
	if Distance(mpos, ppos) == 1 && g.Map.At(ppos) != BarrelCell && !m.Peaceful(g) {
		if m.Status(MonsConfused) {
			g.Logf("%s appears too confused to attack.", m.Kind.Definite(true))
			return
		}
		if g.Map.At(ppos) == TreeCell && !m.Kind.CanAttackOnTree() {
			g.Logf("%s watches you from below.", m.Kind.Definite(true))
			return
		}
		m.AttackAction(g)
		return
	}
	if m.Status(MonsLignified) {
		return
	}
	switch m.State {
	case Watching:
		m.HandleWatching(g)
		return
	}
	m.ComputePath(g)
	if len(m.Path) < 2 {
		m.HandleEndPath(g)
		return
	}
	m.HandleMove(g)
}

// InvertFoliage inverts foliage (for worms).
func (m *Monster) InvertFoliage(g *Game) {
	if m.Kind != MonsWorm {
		return
	}
	invert := false
	t := g.Map.At(m.P)
	switch t {
	case CavernCell:
		g.Map.Set(m.P, FoliageCell)
		invert = true
	case FoliageCell:
		g.Map.Set(m.P, CavernCell)
		invert = true
	}
	if invert {
		if g.Player.Sees(m.P) {
			g.ComputeFOV()
		}
	}
}

// Exhaust makes the monster exhausted for a default duration.
func (m *Monster) Exhaust(g *Game) {
	m.ExhaustTime(g, DurationExhaustionMonster+RandInt(DurationExhaustionMonster/2))
}

// ExhaustTime makes the monster exhausted for the given duration.
func (m *Monster) ExhaustTime(g *Game, t int) {
	m.PutStatus(g, MonsExhausted, t)
}

// HitPlayer makes the monster hit the player (if possible).
func (m *Monster) HitPlayer(g *Game) {
	if g.Player.HP <= 0 || Distance(g.Player.P, m.P) > 1 {
		return
	}
	clang := RandInt(4) == 0
	noise := g.HitNoise(clang)
	g.MakeNoise(noise, g.Player.P)
	var sclang string
	if clang {
		sclang = g.ClangMsg()
	}
	g.LogfStyled("%s hits you (1 dmg).", logDamage, m.Kind.Definite(true))
	if sclang != "" {
		g.LogStyled(sclang, logNotable)
	}
	m.InflictDamage(g)
	if g.Player.HP <= 0 {
		return
	}
	m.HitSideEffects(g)
	const HeavyWoundHP = 2
	if g.Player.HP >= HeavyWoundHP {
		return
	}
	switch g.Player.Inventory.Neck.Kind {
	case AmuletConfusion:
		g.LogfStyled("Your amulet releases confusing harmonies.", logNotable)
		m.PutConfusion(g)
	case AmuletFog:
		g.LogStyled("Your amulet feels warm.", logNotable)
		g.ActivateAmuletFog()
	case AmuletTeleport:
		g.LogStyled("Your amulet shines.", logNotable)
		m.TeleportAway(g)
		if m.Kind.ReflectsTeleport() {
			g.LogfStyled("The %s reflected back some energies.", logNotable, m.Kind)
			g.Teleportation()
		}
	case AmuletDazing:
		g.LogStyled("Your amulet glows.", logNotable)
		m.PutStatus(g, MonsDazed, DurationDazeMonster)
	}
}

// PutStatus gives the monster a status effect with the given duration.
func (m *Monster) PutStatus(g *Game, st monsterStatus, duration int) bool {
	if m.Status(st) {
		return false
	}
	m.Statuses[st] += duration
	g.PushEventD(&MonsterStatusEvent{
		Index:  m.Index,
		Status: st}, DurationStatusStep)

	return true
}

// PutConfusion confuses the monster.
func (m *Monster) PutConfusion(g *Game) bool {
	if m.PutStatus(g, MonsConfused, DurationConfusionMonster) {
		m.Path = m.Path[:0]
		if g.Player.Sees(m.P) {
			g.Logf("%s looks confused.", m.Kind.Definite(true))
		}
		return true
	}
	return false
}

// EnterLignification gives lignification to the monster.
func (m *Monster) EnterLignification(g *Game) {
	if m.PutStatus(g, MonsLignified, DurationLignificationMonster) {
		m.Path = m.Path[:0]
		if g.Player.Sees(m.P) {
			g.Logf("%s is rooted to the ground.", m.Kind.Definite(true))
		}
	}
}

// HitSideEffects handles any monster-specific hit side effects.
func (m *Monster) HitSideEffects(g *Game) {
	switch m.Kind {
	case MonsEarthDragon:
		if m.Status(MonsConfused) {
			m.PushPlayer(g, 3)
		} else {
			m.PushPlayer(g, 6)
		}
		g.PutConfusion()
	case MonsBlinkingFrog:
		if g.Blink() {
			g.StoryPrintf("Blinked away by %s", m.Kind)
			g.Stats.TimesBlinked++
		}
	case MonsYack:
		m.PushPlayer(g, 5)
	case MonsAcidMound:
		m.Corrode(g)
	case MonsSpider:
		g.PutConfusion()
	case MonsWingedMilfid:
		if m.Status(MonsExhausted) || g.Player.HasStatus(StatusLignification) {
			break
		}
		t := g.Map.At(g.Player.P)
		if !(Passable(t) || IsSwimPassable(t) || IsDoorPassable(t) || IsLevitatePassable(t)) {
			break
		}
		g.PlacePlayerAt(m.P)
		g.LogStyled("The flying milfid makes you swap positions.", logNotable)
		g.StoryPrintf("Position swap by %s", m.Kind)
		m.ExhaustTime(g, 5+RandInt(5))
		if g.Map.At(g.Player.P) == ChasmCell {
			g.PushEventFirst(AbyssFallEvent{}, g.Turn)
		}
	case MonsTinyHarpy:
		if m.Status(MonsSatiated) {
			return
		}
		g.Player.Bananas--
		if g.Player.Bananas < 0 {
			g.Player.Bananas = 0
		} else {
			m.PutStatus(g, MonsSatiated, DurationSatiationMonster)
			g.LogStyled("The tiny harpy steals a banana from you.", logNotable)
			g.StoryPrintf("Banana stolen by %s (bananas: %d)", m.Kind, g.Player.Bananas)
			g.Stats.StolenBananas++
			m.Target = m.NextTarget(g)
			m.MakeWander()
		}
	}
}

// PushPlayer pushes the player a given distance amount.
func (m *Monster) PushPlayer(g *Game, dist int) {
	if g.Player.HasStatus(StatusLignification) {
		return
	}
	dir := DirNorm(m.P, g.Player.P)
	p := g.Player.P
	q := p
	path := []gruid.Point{p}
	i := 0
	for {
		i++
		q = q.Add(dir)
		path = append(path, q)
		if !InMap(q) || BlocksRange(g.Map.At(q)) {
			path = path[:len(path)-1]
			break
		}
		mons := g.MonsterAt(q)
		if mons.Exists() {
			continue
		}
		p = q
		if i >= dist {
			break
		}
	}
	if p == g.Player.P {
		// XXX: pushing: do more interesting things, perhaps?
		return
	}
	g.Stats.TimesPushed++
	t := g.Map.At(p)
	var cs string
	if m.Kind == MonsEarthDragon {
		cs = " inadvertently"
		if m.Status(MonsConfused) {
			cs = " out of confusion"
		}
	}
	g.md.PushAnimation(path)
	g.PlacePlayerAt(p)
	g.Logf("%s pushes you%s.", m.Kind.Definite(true), cs)
	g.StoryPrintf("Pushed by %s%s", m.Kind.Definite(true), cs)
	if t == ChasmCell {
		g.PushEventFirst(AbyssFallEvent{}, g.Turn)
	}
}

// RangedAttack handles ranged attacks and reports whether there was one.
func (m *Monster) RangedAttack(g *Game) bool {
	if !m.Kind.Ranged() {
		return false
	}
	if m.Status(MonsConfused) {
		g.Logf("%s appears too confused to attack.", m.Kind.Definite(true))
		return false
	}
	if Distance(m.P, g.Player.P) <= 1 && m.Kind != MonsSatowalgaPlant {
		return false
	}
	if !m.SeesPlayer(g) {
		return false
	}
	if m.Status(MonsExhausted) {
		return false
	}
	switch m.Kind {
	case MonsHighGuard:
		return m.ThrowJavelin(g)
	case MonsSatowalgaPlant:
		return m.ThrowAcid(g)
	case MonsMadNixe:
		return m.NixeAttraction(g)
	case MonsVampire:
		return m.VampireSpit(g)
	case MonsTreeMushroom:
		return m.ThrowSpores(g)
	}
	return false
}

// RangeBlockde reports whether the monster cannot perform a ranged attack from
// its position due to some obstacle.
func (m *Monster) RangeBlocked(g *Game) bool {
	ray := g.Ray(m.P)
	if len(ray) == 1 {
		return false
	}
	if len(ray) == 0 {
		// should not happen
		return true
	}
	for _, p := range ray[1:] {
		t := g.Map.At(p)
		if BlocksRange(t) {
			return true
		}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			return true
		}
	}
	return false
}

// BarrierCandidates returns the list of possible candidates for an oric
// barrier.
func (g *Game) BarrierCandidates(p gruid.Point, dir gruid.Point) []gruid.Point {
	candidates := g.CardinalNeighbors(p)
	bestpos := p.Add(dir)
	if Distance(bestpos, p) > 1 {
		j := 0
		for i := 0; i < len(candidates); i++ {
			if Distance(candidates[i], bestpos) == 1 {
				candidates[j] = candidates[i]
				j++
			}
		}
		if len(candidates) > 2 {
			candidates = candidates[0:2]
		}
		return candidates
	}
	worstpos := p.Add(DirNorm(bestpos, p))
	for i := 1; i < len(candidates); i++ {
		if candidates[i] == bestpos {
			candidates[0], candidates[i] = candidates[i], candidates[0]
		}
	}
	for i := 1; i < len(candidates)-1; i++ {
		if candidates[i] == worstpos {
			candidates[len(candidates)-1], candidates[i] = candidates[i], candidates[len(candidates)-1]
		}
	}
	if len(candidates) == 4 && RandInt(2) == 0 {
		candidates[1], candidates[2] = candidates[2], candidates[1]
	}
	if len(candidates) == 4 {
		candidates = candidates[0:3]
	}
	return candidates
}

// CreateBarrier attemps to create a magical barrier.
func (m *Monster) CreateBarrier(g *Game) bool {
	dir := DirNorm(m.P, g.Player.P)
	candidates := g.BarrierCandidates(g.Player.P, dir)
	done := false
	for _, p := range candidates {
		t := g.Map.At(p)
		mons := g.MonsterAt(p)
		if mons.Exists() || !IsLevitatePassable(t) {
			continue
		}
		g.MagicalBarrierAt(p)
		done = true
		g.LogStyled("The oric celmist creates a magical barrier.", logNotable)
		g.StoryPrintf("Blocked by %s barrier", m.Kind)
		g.Stats.TimesBlocked++
		break
	}
	if !done {
		return false
	}
	m.Exhaust(g)
	return true
}

// Illuminate attemps to illuminate the player.
func (m *Monster) Illuminate(g *Game) bool {
	if g.PutStatus(StatusIlluminated, DurationIlluminated) {
		g.Log("The harmonic celmist casts magical harmonies on you.")
		g.LogStyled("Buzz!", logNotable)
		g.StoryPrintf("Illuminated by %s", m.Kind)
		g.md.EffectAtPPAnimation(ColorYellow)
		g.MakeNoise(HarmonicNoise, g.Player.P)
		m.Exhaust(g)
		return true
	}
	return false
}

// VampireSpit attempts to spit at the player.
func (m *Monster) VampireSpit(g *Game) bool {
	blocked := m.RangeBlocked(g)
	if blocked || g.Player.HasStatus(StatusConfusion) {
		return false
	}
	g.LogStyled("The vampire spits at you.", logNotable)
	g.StoryPrint("Spat at by vampire")
	g.md.MonsterProjectileAnimation(g.Ray(m.P), '*', ColorOrange)
	g.PutConfusion()
	m.Exhaust(g)
	return true
}

// ThrowSpores attempts to lignifying throw spores at the player.
func (m *Monster) ThrowSpores(g *Game) bool {
	blocked := m.RangeBlocked(g)
	if blocked || g.Player.HasStatus(StatusLignification) {
		return false
	}
	g.LogStyled("The tree mushroom releases spores at you.", logNotable)
	g.StoryPrintf("Lignified by %s", m.Kind)
	g.md.BeamsAnimation(g.Ray(m.P), BeamLignification)
	g.EnterLignification()
	m.Exhaust(g)
	return true
}

// ThrowJaveling attemps to throw a javelin a the player.
func (m *Monster) ThrowJavelin(g *Game) bool {
	blocked := m.RangeBlocked(g)
	if blocked {
		return false
	}
	clang := RandInt(4) == 0
	noise := g.HitNoise(clang)
	var sclang string
	if clang {
		sclang = g.ClangMsg()
	}
	g.Logf("%s throws a javelin at you (1 dmg).", m.Kind.Definite(true))
	if sclang != "" {
		g.LogStyled(sclang, logNotable)
	}
	g.StoryPrintf("Targeted by %s javelin", m.Kind)
	g.md.JavelinThrowAnimation(g.Ray(m.P))
	g.MakeNoise(noise, g.Player.P)
	m.InflictDamage(g)
	m.ExhaustTime(g, 10+RandInt(5))
	return true
}

// Corrode makes the player lose some magara charges through corrosion.
func (m *Monster) Corrode(g *Game) {
	count := 0
	for i := range g.Player.Magaras {
		n := RandInt(2)
		g.Player.Magaras[i].Charges -= n
		if g.Player.Magaras[i].Charges < 0 {
			g.Player.Magaras[i].Charges = 0
		} else {
			count += n
		}
	}
	if count > 0 {
		g.Logf("You lose %d magara charges by corrosion.", count)
		g.StoryPrintf("Corroded by %s (lost %d magara charges)", m.Kind, count)
	}
}

// ThrowAcid attemps to throw acid at the player.
func (m *Monster) ThrowAcid(g *Game) bool {
	blocked := m.RangeBlocked(g)
	if blocked {
		return false
	}
	noise := g.HitNoise(false) // no clang with acid projectiles
	g.Logf("%s throws acid at you (1 dmg).", m.Kind.Definite(true))
	g.md.MonsterProjectileAnimation(g.Ray(m.P), '*', ColorGreen)
	g.MakeNoise(noise, g.Player.P)
	m.InflictDamage(g)
	m.Corrode(g)
	m.ExhaustTime(g, 2)
	return true
}

// NixeAttraction attemps to attract the player.
func (m *Monster) NixeAttraction(g *Game) bool {
	blocked := m.RangeBlocked(g)
	if blocked {
		return false
	}
	ray := g.Ray(m.P)
	if len(ray) <= 1 || !IsPlayerPassable(g.Map.At(ray[1])) {
		return false
	}
	g.MakeNoise(MagicCastNoise, m.P)
	g.Logf("%s lures you to her.", m.Kind.Definite(true))
	g.StoryPrintf("Lured by %s", m.Kind)
	g.md.TeleportAnimation(g.Player.P, ray[1], true)
	g.PlacePlayerAt(ray[1])
	m.Exhaust(g)
	return true
}

// SmitingAttack attempts to perform a smiting attack against the player.
func (m *Monster) SmitingAttack(g *Game) bool {
	if !m.Kind.Smiting() {
		return false
	}
	if m.Status(MonsConfused) {
		g.Logf("%s appears too confused to attack.", m.Kind.Definite(true))
		return false
	}
	if !m.SeesPlayer(g) {
		return false
	}
	if m.Status(MonsExhausted) {
		return false
	}
	switch m.Kind {
	case MonsMirrorSpecter:
		return m.AbsorbMana(g)
	case MonsOricCelmist:
		return m.CreateBarrier(g)
	case MonsHarmonicCelmist:
		return m.Illuminate(g)
	}
	return false
}

// AbsorbMana attempts to absorb mana from the player.
func (m *Monster) AbsorbMana(g *Game) bool {
	if g.Player.MP == 0 {
		return false
	}
	g.Player.MP -= 1
	g.Logf("%s absorbs your mana.", m.Kind.Definite(true))
	g.StoryPrintf("Mana absorbed by %s (MP: %d)", m.Kind, g.Player.MP)
	m.ExhaustTime(g, 3+RandInt(2))
	return true
}

// Blink attemps to make the monster blink away.
func (m *Monster) Blink(g *Game) {
	if m.Status(MonsLignified) {
		return
	}
	npos := g.BlinkPos(true)
	if !InMap(npos) || npos == g.Player.P || npos == m.P {
		return
	}
	opos := m.P
	g.Logf("The %s blinks away.", m.Kind)
	g.md.TeleportAnimation(opos, npos, true)
	m.MoveTo(g, npos)
}

// MakeHunt makes the monster hunt the player.
func (m *Monster) MakeHunt(g *Game) (noticed bool) {
	if m.State != Hunting {
		m.State = Hunting
		g.Stats.NSpotted++
		g.Stats.DSpotted[g.Map.Depth]++
		if !m.Alerted {
			g.Stats.NUSpotted++
			g.Stats.DUSpotted[g.Map.Depth]++
			if g.Stats.DUSpotted[g.Map.Depth] >= 90*len(g.Monsters)/100 {
				AchUnstealthy.Get(g)
			}
		}
		m.Alerted = true
		noticed = true
	}
	m.Search = g.Player.P
	m.Target = g.Player.P
	return noticed
}

// MakeAware makes the monster aware (if appropriate).
func (m *Monster) MakeAware(g *Game) bool {
	if m.Peaceful(g) || m.Status(MonsSatiated) {
		if m.State == Resting && Distance(m.P, g.Player.P) == 1 {
			g.Logf("%s awakens.", m.Kind.Definite(true))
			m.MakeWander()
		}
		return false
	}
	if !m.SeesPlayer(g) {
		return false
	}
	switch m.State {
	case Resting:
		g.Logf("%s awakens.", m.Kind.Definite(true))
	case Wandering, Watching:
		g.Logf("%s notices you.", m.Kind.Definite(true))
	}
	noticed := m.MakeHunt(g)
	if noticed && m.Kind == MonsDog {
		g.Logf("%s barks.", m.Kind.Definite(true))
		g.LogStyled("WOOF!", logNotable)
		g.StoryPrintf("Barked at by %s", m.Kind)
		g.MakeNoise(BarkNoise, m.P)
	}
	return noticed
}

// GatherBand makes the monster call for any fellow monsters.
func (m *Monster) GatherBand(g *Game) {
	if !monsBands[g.MonsterBands[m.Band].Kind].Band {
		return
	}
	dij := &NoisePath{g: g}
	g.PR.BreadthFirstMap(dij, []gruid.Point{m.P}, 4)
	for _, mons := range g.Monsters {
		if mons.Band == m.Band {
			if mons.State == Hunting && m.State != Hunting {
				continue
			}
			c := g.PR.BreadthFirstMapAt(mons.P)
			if c > 4 || mons.State == Resting && mons.Status(MonsExhausted) {
				continue
			}
			mons.Target = m.Target
			if mons.State == Resting {
				mons.MakeWander()
			}
		}
	}
}

// MonsterInFOV returns the first monster in FOV (if any).
func (g *Game) MonsterInFOV() *Monster {
	for _, mons := range g.Monsters {
		if mons.Exists() && g.Player.Sees(mons.P) {
			return mons
		}
	}
	return nil
}
