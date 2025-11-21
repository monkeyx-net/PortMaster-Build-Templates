package main

import (
	"cmp"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// InitPlayer initializes the player entity for the rest of the game.
func (g *Game) InitPlayer() {
	g.Entities[PlayerID] = &Entity{
		Name: "You",
		Rune: '@',
		Role: NewActor(2, 1, 9, Player),
	}
	for i := PrimarySpiritID; i < FirstMapID; i++ {
		g.Entities[i] = emptySlot()
	}
	// g.debugBuild()
	g.ComputePlayerStats()
}

func (g *Game) debugBuild() {
	// g.Entities[PrimarySpiritID+1] = spiritEntity(secondarySpirits[0])
	// g.Entities[PrimarySpiritID+3] = comestibleEntity(ComestibleData[4])
}

func emptySlot() *Entity {
	return &Entity{Name: "(empty slot)", P: InvalidPos, Seen: true}
}

// GenEntities generates the entities for a new level, placing them on the map.
func (g *Game) GenEntities(mg *MapGen) {
	// Reset non-player actors.
	clear(g.Entities[FirstMapID+1:])
	g.Entities = g.Entities[:FirstMapID+1]
	g.Map.ActorCache = g.Map.ActorCache.New()
	g.Map.RuneCache = g.Map.RuneCache.New()
	// Generate new items and actors.
	g.genPlayerOnNewLevel(mg)
	g.genItems(mg)
	g.genMonsters(mg)
}

// genPlayerOnNewLevel places the player in the map with replenished charges and HP.
func (g *Game) genPlayerOnNewLevel(mg *MapGen) {
	p := g.PlayerEntity()
	// Sort vaults by distance to center.
	// center := gruid.Point{MapWidth / 2, MapHeight / 2}
	// slices.SortFunc(mg.vaults, func(vi, vj *vault) int {
	// 	pi := vi.p.Add(gruid.Point{vi.w / 2, vi.h / 2})
	// 	pj := vj.p.Add(gruid.Point{vj.w / 2, vj.h / 2})
	// 	return cmp.Compare(paths.DistanceManhattan(center, pi), paths.DistanceManhattan(center, pj))
	// })
	// Chose player's vault.
	var pvi int // player's vault index
	pvi, p.P = mg.RandomVaultsPlace(mg.vaults, PlaceWaypoint)
	pv := mg.vaults[pvi]
	// Sort vaults by distance to the chosen player's vault.
	slices.SortFunc(mg.vaults, func(vi, vj *vault) int {
		return cmp.Compare(vaultDistance(pv, vi), vaultDistance(pv, vj))
	})
	if p.P == InvalidPos {
		panic("invalid player position")
	}
	p.KnownP = p.P
	g.Map.ActorCache.SetU(p.P, PlayerID)
	for _, sp := range g.PlayerSpirits() {
		sp.Charges = sp.MaxCharges[sp.Level]
	}
	pa := g.PlayerActor()
	clear(pa.Statuses)
	pa.HP = pa.MaxHP
}

// spiritInfo gathers information for generating a specific kind of spirit
// entity.
type spiritInfo struct {
	Name         string     // name of the spirit
	MaxCharges   [2]int     // maximum number of charges per level
	Ability      [2]Ability // active ability effect when used
	BonusAttack  [2]int     // attack bonuses per level (may be zero)
	BonusDefense [2]int     // defense bonuses per level (may be zero)
	BonusTraits  [2]Traits  // bonus traits per level
	BonusHP      [2]int     // bonus hp per level (may be zero)
}

// primarySpirits represents the various kinds of primary spirits.
var primarySpirits = []spiritInfo{
	{
		Name:        "Four-Headed Hydra",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectFocus{}, EffectFocus{}},
		BonusHP:     [2]int{0, 3},
		BonusTraits: [2]Traits{PatternFourDirs, PatternFourDirs}},
	{
		Name:        "Rampaging Boar",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectDig{}, EffectDig{}},
		BonusAttack: [2]int{0, 1},
		BonusTraits: [2]Traits{PatternRampage | PushingCharge, PatternRampage | PushingCharge}},
	{
		Name:         "Jumping Frog",
		MaxCharges:   [2]int{3, 4},
		Ability:      [2]Ability{EffectJump{}, EffectJump{}},
		BonusDefense: [2]int{0, 1},
		BonusTraits:  [2]Traits{PatternCatch, PatternCatch}},
	{
		Name:        "Wind Fox",
		MaxCharges:  [2]int{2, 4},
		Ability:     [2]Ability{EffectPushingGale{}, EffectPushingGale{}},
		BonusTraits: [2]Traits{PatternRangedRecoil, PatternRangedRecoil}},
	{
		Name:        "Temporal Cat",
		MaxCharges:  [2]int{1, 2},
		Ability:     [2]Ability{EffectTimeStop{}, EffectTimeStop{}},
		BonusTraits: [2]Traits{PatternSwapDaze, PatternSwapDaze}},
}

// secondarySpirits represents the various kinds of secondary spirit totems
// that can be found.
var secondarySpirits = []spiritInfo{
	{
		Name:        "Thunder Porcupine",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectLightning{}, EffectLightning{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{DazingSpines, DazingSpines | ResistanceDaze}},
	{
		Name:        "Barking Hound",
		MaxCharges:  [2]int{2, 4},
		Ability:     [2]Ability{EffectBark{}, EffectBark{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{GoodHearing, GoodHearing | ResistanceFear}},
	{
		Name:        "Stinking Skunk",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectNoxiousSmell{}, EffectNoxiousSmell{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{GoodSmell, GoodSmell | ResistanceConfusion}},
	{
		Name:         "Walking Tree",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectLignify{}, EffectLignify{}},
		BonusDefense: [2]int{2, 2},
		BonusTraits:  [2]Traits{VulnerabilityFire, VulnerabilityFire | WoodyLegs}},
	{
		Name:         "Venomous Viper",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectPoisonCloud{}, EffectPoisonCloud{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{VenomousMelee, VenomousMelee | ResistancePoison}},
	{
		Name:         "Sprinting Gazelle",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectSprint{}, EffectSprint{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{TrailingCloud, TrailingCloud | ResistanceImbalance}},
	{
		Name:        "Fire Salamander",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectFireRetreat{}, EffectFireRetreat{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{BadHearing | BurningHits, BadHearing | BurningHits | ResistanceFire}},
}

// spiritEntity generates a new spirit entity using the given data.
func spiritEntity(si spiritInfo) *Entity {
	return &Entity{
		Name:   si.Name,
		KnownP: InvalidPos,
		Role: &Spirit{
			Charges:      si.MaxCharges[0],
			MaxCharges:   si.MaxCharges,
			Ability:      si.Ability,
			BonusAttack:  si.BonusAttack,
			BonusDefense: si.BonusDefense,
			BonusTraits:  si.BonusTraits,
			BonusHP:      si.BonusHP}}
}

// genEmptyTotemAt generates a new empty totem entity at the given point.
func (g *Game) genEmptyTotemAt(p gruid.Point) ID {
	return g.AddEntity(&Entity{Name: "empty totem", Rune: '!', P: p, Role: &EmptyTotem{}})
}

// itemInfo gathers information for item generation.
type itemInfo struct {
	Name   string
	Effect Effect
}

// ComestibleData provides information about the various kinds of comestibles.
var ComestibleData = []itemInfo{
	{"ambrosia berries", EffectAmbrosiaBerries{}},
	{"berserking flower", EffectBerserkingFlower{}},
	{"clarity leaves", EffectClarityLeaves{}},
	{"firebreath pepper", EffectFirebreathPepper{}},
	{"foggy-skin onion", EffectFoggySkinOnion{}},
	{"lignification fruit", EffectLignificationFruit{}},
	{"teleport mushroom", EffectTeleportMushroom{}},
}

// comestibleEntity generates a new comestible Entity entity using the given
// data.
func comestibleEntity(ci itemInfo) *Entity {
	return &Entity{
		Name:   ci.Name,
		Rune:   '%',
		KnownP: InvalidPos,
		Role:   &Comestible{Effect: ci.Effect}}
}

// menhirKind represents the various kinds of menhirs.
type menhirKind int

const (
	MenhirEarth menhirKind = iota
	MenhirFire
	MenhirPoison
	MenhirWarping
)

// MenhirData provides information about the various kinds of menhirs.
var MenhirData = []itemInfo{
	MenhirEarth:   {"earth menhir", EffectEarthMenhir{}},
	MenhirFire:    {"fire menhir", EffectFireMenhir{}},
	MenhirPoison:  {"poison menhir", EffectPoisonMenhir{}},
	MenhirWarping: {"warping menhir", EffectWarpingMenhir{}},
}

// menhirEntity generates a new menhir Entity entity using the given data.
func menhirEntity(ci itemInfo) *Entity {
	return &Entity{
		Name:   ci.Name,
		Rune:   '&',
		KnownP: InvalidPos,
		Role:   &Menhir{Effect: ci.Effect}}
}

// genItems generates items for the new level.
func (g *Game) genItems(mg *MapGen) {
	// Portal.
	portalIdxs := []int{}
	npvaults := mg.vaults[1:] // non-player vaults
	if g.Map.Level < MapLevels {
		var i int
		portal := &Entity{Name: "magic portal", Rune: '>', Role: &Portal{}}
		i, portal.P = mg.RandomVaultsPlaceWithFunc(npvaults[len(npvaults)/3:], PlaceStatic, g.IntNBiasedUp)
		portalIdxs = append(portalIdxs, i+1)
		g.Map.Portal = portal.P
		g.AddEntity(portal)
		if g.ProcInfo.FakePortal[g.Map.Level-1] {
			portal := &Entity{Name: "magic portal", Rune: '>', Role: &Portal{Fake: true}}
			i, portal.P = mg.RandomVaultsPlaceWithFunc(npvaults[len(npvaults)/3:], PlaceStatic, g.IntNBiasedUp)
			portalIdxs = append(portalIdxs, i+1)
			g.AddEntity(portal)
		}
	} else {
		var i int
		orb := &Entity{Name: "Orb of Corruption", Rune: '☼', Role: &CorruptionOrb{}}
		i, orb.P = mg.RandomVaultsPlaceWithFunc(npvaults[len(npvaults)/2:], PlaceStatic, g.IntNBiasedUp)
		g.Map.Orb = orb.P
		portalIdxs = append(portalIdxs, i+1)
		g.AddEntity(orb)
	}
	// Menhirs.
	nmenhirs := g.ProcInfo.NMenhirs[g.Map.Level-1]
	menhirIdxs := []int{}
	for range min(nmenhirs, len(MenhirData)) {
		idx := g.NextMenhirIdx()
		for slices.Contains(menhirIdxs, idx) {
			// Never generate twice the same menhir in a given level.
			idx = g.NextMenhirIdx()
		}
		menhirIdxs = append(menhirIdxs, idx)
		tvaults := mg.vaults
		if menhirKind(idx) == MenhirWarping {
			// Never generate warping menhir in vault containing a
			// portal or the orb of corruption.
			tvaults = []*vault{}
			for i, v := range mg.vaults {
				if !slices.Contains(portalIdxs, i) {
					tvaults = append(tvaults, v)
				}
			}
		}
		me := menhirEntity(MenhirData[idx])
		_, me.P = mg.RandomVaultsPlace(tvaults, PlaceStatic)
		g.AddEntity(me)
	}
	// Totemic spirit.
	if spiritIndex := g.ProcInfo.Spirits[g.Map.Level-1]; spiritIndex >= 0 {
		sp := spiritEntity(secondarySpirits[spiritIndex])
		spr := sp.Role.(*Spirit)
		spr.Charges = spr.MaxCharges[spr.Level]
		sp.Rune = '!'
		_, sp.P = mg.RandomVaultsPlace(npvaults, PlaceItem)
		g.Map.Totem = sp.P
		g.AddEntity(sp)
	} else if g.Map.Level < MapLevels {
		_, p := mg.RandomVaultsPlace(npvaults, PlaceItem)
		g.Map.Totem = p
		g.genEmptyTotemAt(p)
	}
	// Comestibles.
	nitems := g.ProcInfo.NComestibles[g.Map.Level-1]
	for i := range nitems {
		co := comestibleEntity(ComestibleData[g.NextComestibleIdx()])
		if i <= (nitems+1)/2 {
			co.P = mg.RandomPlace(PlaceItem)
		} else {
			co.P = g.randomFreeItemFloor()
		}
		g.AddEntity(co)
	}
	// Runic traps.
	rto, rtt := 1, 1 // runes in open areas and in tunnels
	switch g.IntN(4) {
	case 0:
		rto++
	case 1:
		rtt++
	}
	if g.Map.Level >= 6 {
		// Extra runic trap.
		if g.IntN(2) == 0 {
			rto++
		} else {
			rtt++
		}
	}
	if g.Map.Level == g.ProcInfo.TrapLevel {
		rto += 4
		rtt++
	}
	for range rto {
		g.genRunicTrap(mg)
	}
	for range rtt {
		g.genRunicTrapInTunnel(mg)
	}
}

// randomFreeItemFloor returns a random map floor position that is not occupied by
// an item.
func (g *Game) randomFreeItemFloor() gruid.Point {
	for {
		p := g.RandomPassableBiased()
		if i, _ := g.ItemAt(p); i >= 0 {
			continue
		}
		if p == g.PP() {
			// Avoid putting a comestible on the player's starting
			// position, for clarity.
			continue
		}
		// We replace foliage or rubble with floor for putting the
		// item. We could allow items on non-floor cells, but for
		// clarity, it's better this way.
		g.Map.Terrain.Set(p, Floor)
		return p
	}
}

// RandomPassableBiased is like RandomPassable, but biased toward positions
// next to walls.
func (g *Game) RandomPassableBiased() gruid.Point {
	m := g.Map.Terrain
	size := m.Size()
	var nbs paths.Neighbors
	for {
		p := gruid.Point{g.rand.IntN(size.X), g.rand.IntN(size.Y)}
		if t := m.At(p); Passable(t) {
			return p
		}
		ps := nbs.All(p, g.Map.Passable)
		if len(ps) == 0 {
			continue
		}
		// Return a passable neighbor (including diagonals to increase
		// bias toward corners and dead ends).
		return ps[g.IntN(len(ps))]
	}
}

// monsterInfo gathers information for generating a specific kind of monster.
type monsterInfo struct {
	Name    string // name of the monster
	R       rune   // rune (for display)
	Attack  int    // actor's attack
	Defense int    // actor's defense
	HP      int    // actor's max HP
	Traits  Traits // actor's traits
}

// monsterKind represents the various kinds of monsters.
type monsterKind int

const (
	AcidMound monsterKind = iota
	BarkingHound
	BerserkingSpider
	BlazingGolem
	BlinkButterfly
	BurningPhoenix
	ConfusingEye
	CrazyDruid
	EarthDragon
	ExplodingNadre
	FearsomeLich
	FireLlama
	FourHeadedHydra
	HungryRat
	LashingFrog
	MadOctopode
	NoisyImp
	RampagingBoar
	TemporalCat
	ThunderPorcupine
	TotemWasp
	UndeadKnight
	VenomousViper
	WalkingMushroom
	WalkingTree
	WarpingWraith
	WindFox
)

// MonsData provides information about the various kinds of monsters.
var MonsData = []monsterInfo{
	// 		  {name, rune, attack, defense, hp, traits}
	AcidMound:        {"acid mound", 'a', 2, 0, 5, MonsIgnoreDefense | MonsImmunityImbalance | MonsCreep},
	BarkingHound:     {"barking hound", 'h', 3, 0, 4, MonsBarking | GoodHearing | MonsImmunityFear},
	BerserkingSpider: {"berserking spider", 's', 2, 1, 3, MonsBerserking | MonsSilent},
	BlazingGolem:     {"blazing golem", 'G', 2, 3, 4, MonsExplodingDeath | MonsImmunityFire | MonsImmunityPoison | MonsImmunityFear | MonsImmunityConfusion | MonsHeavyFootsteps | MonsNotable},
	BlinkButterfly:   {"blinking butterfly", 'b', 2, 3, 2, MonsBlink | MonsImmunityImbalance | MonsImmunityLignification | MonsWingFlap},
	BurningPhoenix:   {"burning phoenix", 'P', 3, 1, 5, PatternRampage | BurningHits | MonsImmunityFire | MonsImmunityLignification | MonsWingFlap | MonsNotable},
	ConfusingEye:     {"confusing eye", 'e', 2, 0, 2, PatternRanged | MonsConfusion | MonsImmunityConfusion | MonsLightFootsteps},
	CrazyDruid:       {"crazy druid", 'C', 3, 1, 4, PatternSwap | BurningHits | MonsBerserking | MonsNotable},
	EarthDragon:      {"earth dragon", 'D', 4, 2, 5, Pushing | MonsDig | MonsImmunityFear | MonsScales | MonsHeavyFootsteps | MonsNotable},
	ExplodingNadre:   {"exploding nadre", 'n', 2, 3, 1, MonsExplodingDeath | MonsScales},
	FearsomeLich:     {"fearsome lich", 'L', 2, 2, 5, PatternRanged | MonsFear | MonsIgnoreDefense | MonsImmunityPoison | MonsImmunityConfusion | MonsNotable},
	FireLlama:        {"fire llama", 'l', 2, 0, 4, PatternRanged | MonsSpitFire},
	FourHeadedHydra:  {"four-headed hydra", 'H', 2, 0, 6, PatternFourDirsMons | MonsScales | MonsFourHeaded | MonsHeavyFootsteps | MonsNotable},
	HungryRat:        {"hungry rat", 'r', 2, 0, 3, MonsHungry | MonsLightFootsteps},
	NoisyImp:         {"noisy imp", 'I', 2, 2, 4, MonsMusic | MonsLightFootsteps | MonsNotable},
	MadOctopode:      {"mad octopode", 'O', 3, 2, 5, PatternCatch | MonsImmunityImbalance | VenomousMelee | MonsCreep | MonsNotable},
	LashingFrog:      {"lashing frog", 'F', 2, 1, 4, PatternCatch},
	RampagingBoar:    {"rampaging boar", 'B', 3, 0, 4, PatternRampage | PushingCharge | MonsDig},
	TemporalCat:      {"temporal cat", 'c', 2, 0, 4, PatternSwapDaze | MonsLightFootsteps},
	ThunderPorcupine: {"thunder porcupine", 'p', 2, 0, 3, DazingSpines | MonsImmunityDaze | MonsLightFootsteps},
	TotemWasp:        {"totem wasp", 'w', 2, 1, 2, PatternRampage | VenomousMelee | MonsImmunityLignification | MonsWingFlap | MonsNotable},
	UndeadKnight:     {"undead knight", 'K', 3, 3, 4, PatternFourDirsMons | MonsFear | MonsImmunityPoison | MonsImmunityDaze | MonsNotable},
	VenomousViper:    {"venomous viper", 'v', 2, 1, 4, VenomousMelee | MonsImmunityPoison | MonsScales | MonsCreep},
	WalkingMushroom:  {"walking mushroom", 'M', 2, 1, 5, MonsSpores | MonsConfusion | MonsImmunityConfusion | MonsImmunityLignification | VulnerabilityFire | MonsNotable},
	WalkingTree:      {"walking tree", 'T', 2, 2, 5, MonsLignify | MonsImmunityImbalance | MonsImmunityPoison | MonsImmunityLignification | VulnerabilityFire | MonsHeavyFootsteps | MonsNotable},
	WarpingWraith:    {"warping wraith", 'W', 2, 3, 3, MonsTeleport | MonsImmunityPoison | MonsImmunityLignification | MonsSilent},
	WindFox:          {"wind fox", 'f', 2, 0, 4, PatternRangedRecoil | MonsLightFootsteps},
}

// monsEarly represents early level monsters.
var monsEarly = []monsterKind{
	BerserkingSpider,
	ConfusingEye,
	HungryRat,
	ThunderPorcupine,
}

// monsMid represents monsters that appear throughout the game (but less often
// early).
var monsMid = []monsterKind{
	AcidMound,
	BarkingHound,
	BlinkButterfly,
	ExplodingNadre,
	FireLlama,
	LashingFrog,
	RampagingBoar,
	TemporalCat,
	VenomousViper,
	WindFox,
}

// monsLate represents dangerous monsters that appear mostly in mid-to-late
// levels.
var monsLate = []monsterKind{
	BurningPhoenix,
	EarthDragon,
	FourHeadedHydra,
	WalkingTree,
}

// genMonsters generates monster entities for the current level.
func (g *Game) genMonsters(mg *MapGen) {
	if g.Map.Orb != InvalidPos {
		g.genMonsterGuardian(FearsomeLich, g.Map.Orb)
		g.genMonsterGuardian(UndeadKnight, g.Map.Orb)
	}
	if g.ProcInfo.GuardianPortal2 == g.Map.Level {
		g.genMonsterGuardian(MadOctopode, g.Map.Portal)
	}
	if g.ProcInfo.GuardianTotem1 == g.Map.Level {
		g.genMonsterGuardian(CrazyDruid, g.Map.Totem)
	}
	// number of early, mid, late monsters.
	nE, nM, nL := nMonsE[g.Map.Level-1], nMonsM[g.Map.Level-1], nMonsL[g.Map.Level-1]
	if g.ProcInfo.WanderingUnique1 == g.Map.Level {
		nM--
		g.genMonster(WalkingMushroom)
	}
	if g.ProcInfo.WanderingUnique2 == g.Map.Level {
		nM--
		g.genMonster(NoisyImp)
	}
	if g.ProcInfo.GuardianTotem2 == g.Map.Level {
		nE--
		g.genMonsterGuardian(TotemWasp, g.Map.Totem)
		g.genMonsterGuardian(TotemWasp, g.Map.Totem)
	}
	if g.ProcInfo.GuardianPortal1 == g.Map.Level {
		nM--
		p := g.Map.Portal
		if g.IntN(4) == 0 {
			// Occasionally make golems guard a random non-player
			// vault.
			_, p = mg.RandomVaultsPlace(mg.vaults[1:], PlaceWaypoint)
		}
		g.genMonsterGuardian(BlazingGolem, p)
		g.genMonsterGuardian(BlazingGolem, p)
	}
	switch g.Map.Level {
	case g.ProcInfo.MonsEarly:
		// Special level with many early monsters of same kind.
		nE += 2
		nM--
		mk := monsEarly[g.IntN(len(monsEarly))]
		for range nE - 2 {
			g.genMonster(mk)
		}
		nE = 2 // two remaining
	case g.ProcInfo.MonsMid:
		// Special level with many mid monsters of same kind.
		nE -= 2
		nM += 2
		mk := monsMid[g.IntN(len(monsMid))]
		for range nM - 2 {
			g.genMonster(mk)
		}
		nM = 2 // two remaining
	case g.ProcInfo.MonsLate:
		// Special level with many late monsters of same kind.
		nE--
		nM -= 3
		nL += 3
		mk := monsLate[g.IntN(len(monsLate))]
		for range nL - 2 {
			g.genMonster(mk)
		}
		nL = 2 // two remaining
	case g.ProcInfo.MonsLateSwarm:
		// Special late level (8-9) with many early and mid-level
		// monsters but less late ones.
		n := 0 // adjustment to ensure at least nL == 1 on the 8th level
		if g.Map.Level == 8 {
			n++
		}
		nE += 3 - n
		nM += 6 - n
		nL -= 4 - n
		switch g.IntN(3) {
		case 0:
			// Special level with more early-level monsters and
			// less mid-ones. Also, use smaller selection of early
			// monsters.
			nE += 3
			nM -= 2
			md := g.monsterSelection(monsEarly, 2)
			for range nE - 1 {
				g.genMonster(md[g.IntN(len(md))])
			}
			nE = 1
		case 1:
			// Special level biased toward more mid monsters and
			// with a smaller selection of them.
			nE -= 2
			nM++
			md := g.monsterSelection(monsMid, 3)
			for range nM - 1 {
				g.genMonster(md[g.IntN(len(md))])
			}
			nM = 1
		}
	}
	if g.Map.Level == g.ProcInfo.MonsMidLate {
		// Special level with a smaller selection of mid monsters.
		md := g.monsterSelection(monsMid, 4)
		for range nM - 1 {
			g.genMonster(md[g.IntN(len(md))])
		}
		nM = 1
	}
	for range nE {
		g.genMonster(monsEarly[g.IntN(len(monsEarly))])
	}
	for range nM {
		g.genMonster(monsMid[g.IntN(len(monsMid))])
	}
	for range nL {
		g.genMonster(monsLate[g.IntN(len(monsLate))])
	}
}

var nMonsE = [MapLevels]int{3, 5, 6, 6, 5, 5, 4, 4, 3}
var nMonsM = [MapLevels]int{2, 3, 4, 6, 8, 10, 11, 10, 10}
var nMonsL = [MapLevels]int{0, 0, 1, 1, 2, 2, 3, 4, 6}

// monsterSelection returns a random monster kind subset of size n.
func (g *Game) monsterSelection(mks []monsterKind, n int) []monsterKind {
	md := slices.Clone(mks)
	g.rand.Shuffle(len(md), func(i, j int) {
		md[i], md[j] = md[j], md[i]
	})
	md = md[:min(n, len(md))]
	return md
}

// genMonster spawns a monster of the given kind on a random free floor.
func (g *Game) genMonster(mk monsterKind) (ID, *Entity) {
	p := g.PP()
	q := g.randomFreeFloor()
	// XXX: use walkable-path distance?
	for paths.DistanceManhattan(p, q) <= MaxFOVRange+1 {
		q = g.randomFreeFloor()
	}
	return g.genMonsterAt(mk, q)
}

// genMonsterGardian spawns a gardian monster of the given kind around a
// certain position.
func (g *Game) genMonsterGuardian(mk monsterKind, at gruid.Point) (ID, *Actor) {
	p := g.randomFreeNearby(at, 5)
	id, mons := g.genMonsterAt(mk, p)
	a := mons.Actor()
	a.Behavior.Guard = at
	return id, a
}

// genMonsterAt spawns a monster of the given kind on the given position
// (assumed to be free).
func (g *Game) genMonsterAt(mk monsterKind, p gruid.Point) (ID, *Entity) {
	mons := monster(MonsData[mk])
	mons.P = p
	id := ID(len(g.Entities))
	g.Map.ActorCache.SetU(mons.P, id)
	g.AddEntity(mons)
	return id, mons
}

// monster generates a new monster entity using the given data.
func monster(mi monsterInfo) *Entity {
	return &Entity{
		Name:   mi.Name,
		Rune:   mi.R,
		KnownP: InvalidPos,
		Role:   NewActor(mi.Attack, mi.Defense, mi.HP, mi.Traits),
	}
}

// randomFreeFloor returns a random map floor position that is not occuppied by
// an actor or a trap.
func (g *Game) randomFreeFloor() gruid.Point {
	for {
		p := g.RandomPassableWithoutTrap()
		if i, _ := g.ActorAt(p); i >= 0 {
			continue
		}
		return p
	}
}

// randomFreeNearby returns a random map floor position near the given position
// and that is not occuppied by an actor or trap.
func (g *Game) randomFreeNearby(at gruid.Point, maxdist int) gruid.Point {
	n := 0
	for {
		if n > 1000 {
			return g.randomFreeFloor()
		}
		p := g.RandomPassableWithin(at, maxdist)
		if i, _ := g.ActorAt(p); i >= 0 || !g.NoTrapAt(p) {
			n++
			continue
		}
		return p
	}
}

// runicTrap generates a new runic trap entity using the given data.
func runicTrap(mr MagicRune) *Entity {
	return &Entity{
		Name:   mr.String(),
		Rune:   '=',
		KnownP: InvalidPos,
		Role:   &RunicTrap{Rune: mr},
	}
}

// genRunicTrap generates a runic trap at a position that doesn't truly block
// passage, somewhat biased toward vaults.
func (g *Game) genRunicTrap(mg *MapGen) (ID, *Entity) {
	for {
		p := g.RandomPassableWithoutTrap()
		inVault := mg.vault.At(p)
		if !inVault && g.IntN(6) > 0 {
			// Make vault cells more likely candidates.
			continue
		}
		if i, _ := g.ItemAt(p); i >= 0 {
			continue
		}
		if inVault && slices.Contains(g.Map.Waypoints, p) {
			// Waypoints should be free of traps, because monsters
			// may chose nearby locations reachable from them
			// without passing traps.
			continue
		}
		if p == g.PP() {
			// Don't put a trap on the player's starting position.
			continue
		}
		n, nfree := g.neighborInfo(p)
		if inVault && nfree >= 2 && nfree < 4 && g.suitableTrapLocation(p) {
			return g.genRunicTrapAt(p)
		}
		if n > 2 || nfree < 2 {
			// Not connex or not connex enough or not enough free
			// space.
			continue
		}
		if !inVault && n == 2 && nfree == 2 {
			// Likely an uninteresting corner.
			continue
		}
		return g.genRunicTrapAt(p)
	}
}

// genRunicTrapInTunnel generates a runic trap in an extra tunnel, if possible.
// It blocks locally the passage, but there's always an alternative path.
func (g *Game) genRunicTrapInTunnel(mg *MapGen) (ID, *Entity) {
	if len(mg.xtunnel) == 0 {
		return g.genRunicTrap(mg)
	}
	// We shuffle extra tunnel tiles.
	g.rand.Shuffle(len(mg.xtunnel), func(i, j int) {
		mg.xtunnel[i], mg.xtunnel[j] = mg.xtunnel[j], mg.xtunnel[i]
	})
	for _, p := range mg.xtunnel {
		if i, _ := g.ItemAt(p); i >= 0 {
			continue
		}
		if !g.suitableTrapLocation(p) {
			continue
		}
		return g.genRunicTrapAt(p)
	}
	return g.genRunicTrap(mg)
}

// suitableTrapLocation reports whether p's neighborhood makes it suitable for
// a new trap location.
func (g *Game) suitableTrapLocation(p gruid.Point) bool {
	var nbs paths.Neighbors
	ps := nbs.Cardinal(p, g.Map.PassableWithoutTraps)
	if len(ps) <= 1 {
		// Not an interesting trap location.
		return false
	}
	passable := func(q gruid.Point) bool {
		return g.Map.Passable(q) && g.NoTrapAt(q) && q != p
	}
	for i := range len(ps) - 1 {
		path := g.PR.JPSPath(nil, ps[i], ps[i+1], passable, false)
		if len(path) == 0 {
			// No alternative path.
			return false
		}
	}
	return true
}

// genRunicTrapAt adds a new runic trap entity at the given position. It
// replaces foliage or rubble with floor on that tile.
func (g *Game) genRunicTrapAt(p gruid.Point) (ID, *Entity) {
	g.Map.Terrain.Set(p, Floor)
	r := g.NextRune()
	e := runicTrap(r)
	e.P = p
	id := ID(len(g.Entities))
	g.Map.RuneCache.SetU(p, id)
	g.AddEntity(e)
	return id, e
}

// neighborInfo returns the number of passability toggles (non-passable/passable
// alternations) and passable neighbors without traps.
func (g *Game) neighborInfo(p gruid.Point) (n int, nfree int) {
	passable := g.Map.PassableWithoutTraps
	pass := false
	directions := append(slices.Clone(Directions), gruid.Point{1, 0})
	var pdir gruid.Point // previous direction
	for i, dir := range directions {
		q := p.Add(dir)
		qpass := passable(q)
		if i == 0 {
			pdir = dir
			pass = qpass
			continue
		}
		if qpass {
			if pass && !passable(q.Add(pdir)) {
				// Not connex or not connex enough.
				return -1, -1
			}
			nfree++
		}
		if pass != qpass {
			n++
		}
		pass = qpass
		pdir = dir
	}
	return n, nfree
}
