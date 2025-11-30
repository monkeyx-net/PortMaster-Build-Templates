package main

import (
	"cmp"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// InitPlayer initializes the player entity for the rest of the game.
func (g *Game) InitPlayer(spe *Entity) {
	g.Entities[PlayerID] = &Entity{
		Name: "You",
		Rune: '@',
		Role: NewActor(2, 1, 9, Player),
	}
	for i := PrimarySpiritID; i < FirstMapID; i++ {
		g.Entities[i] = emptySlot()
	}
	g.Entities[PrimarySpiritID] = spe
	// g.debugBuild()
	g.ComputePlayerStats()
}

func (g *Game) debugBuild() {
	// g.Entities[PrimarySpiritID+1] = spiritEntity(g.Mods, challengeSpirits[4])
	// g.Entities[PrimarySpiritID+2] = spiritEntity(g.Mods, challengeSpirits[2])
	// g.Entities[PrimarySpiritID+4] = comestibleEntity(ComestibleData[BerserkingFlower])
	// g.Entities[PrimarySpiritID+5] = comestibleEntity(ComestibleData[ClarityLeaves])
	// g.Entities[PrimarySpiritID+6] = comestibleEntity(ComestibleData[FirebreathPepper])
	// g.Entities[PrimarySpiritID+7] = comestibleEntity(ComestibleData[FoggySkinOnion])
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
	// Chose player's vault.
	var pvi int // player's vault index
	pvi, p.P = g.RandomVaultsPlace(mg, mg.vaults, PlaceWaypoint)
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
	if g.Mod(ModHealingCombat) && g.Map.Level == 5 {
		sp := g.Entity(PrimarySpiritID).Role.(*Spirit)
		if _, ok := sp.Ability[sp.Level].(EffectVampirism); ok {
			sp.MaxCharges[0]++
			sp.MaxCharges[1]++
			sp.Charges++
		}
	}
	if !g.Mod(ModNoRecharges) {
		for _, sp := range g.PlayerSpirits() {
			sp.Charges = sp.MaxCharges[sp.Level]
		}
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
	Advanced     bool       // advanced secondary spirit
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

// primarySpiritsAdvanced represents the various kinds of advanced primary
// spirits.
var primarySpiritsAdvanced = []spiritInfo{
	{
		Name:         "Spinning Crocodile",
		MaxCharges:   [2]int{3, 4},
		Ability:      [2]Ability{EffectTailSlap{}, EffectTailSlap{}},
		BonusAttack:  [2]int{1, 1},
		BonusDefense: [2]int{0, 1},
		BonusTraits:  [2]Traits{PatternCrocodile, PatternCrocodile}},
	{
		Name:        "Vampiric Bat",
		MaxCharges:  [2]int{1, 2},
		Ability:     [2]Ability{EffectVampirism{DurationVampirism}, EffectVampirism{DurationVampirism}},
		BonusTraits: [2]Traits{PatternBat, PatternBat}},
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

// challengeSpirits represents the various kinds of secondary spirit totems
// that can be found.
var challengeSpirits = []spiritInfo{
	{
		Name:        "Gawalt Monkey",
		MaxCharges:  [2]int{1, 4},
		Ability:     [2]Ability{EffectShadows{}, EffectShadows{}},
		BonusTraits: [2]Traits{Gawalt, Gawalt},
		Advanced:    true},
	{
		Name:        "Gluttonous Bear",
		MaxCharges:  [2]int{1, 3},
		Ability:     [2]Ability{EffectSnack{}, EffectSnack{}},
		BonusHP:     [2]int{3, 3},
		BonusTraits: [2]Traits{Gluttony, Gluttony},
		Advanced:    true},
	{
		Name:        "Gardening Lion",
		MaxCharges:  [2]int{1, 3},
		Ability:     [2]Ability{EffectGarden{}, EffectGarden{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{ScaryRoar, ScaryRoar},
		Advanced:    true},
	{
		Name:         "Stomping Elephant",
		MaxCharges:   [2]int{1, 3},
		Ability:      [2]Ability{EffectStomp{}, EffectStomp{}},
		BonusAttack:  [2]int{1, 1},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{Elephanty, Elephanty},
		Advanced:     true},
	{
		Name:        "Staring Owl",
		MaxCharges:  [2]int{1, 3},
		Ability:     [2]Ability{EffectDeathStare{}, EffectDeathStare{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{NocturnalFlying, NocturnalFlying},
		Advanced:    true},
	{
		Name:         "Runic Chicken",
		MaxCharges:   [2]int{2, 5},
		Ability:      [2]Ability{EffectLayRune{}, EffectLayRune{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{RunicChicken, RunicChicken},
		Advanced:     true},
	{
		Name:         "Dazzling Zebra",
		MaxCharges:   [2]int{1, 3},
		Ability:      [2]Ability{EffectDisorient{}, EffectDisorient{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{Dazzling, Dazzling},
		Advanced:     true},
}

// spiritEntity generates a new spirit entity using the given data.
func spiritEntity(mods []bool, si spiritInfo) *Entity {
	if HasMod(mods, ModGluttonyRework) && si.BonusTraits[0].Any(Gluttony) {
		snack := EffectSnack{NoGluttonyStatus: true}
		si.Ability[0], si.Ability[1] = snack, snack
	}
	if HasMod(mods, ModHealingCombat) {
		if _, ok := si.Ability[0].(EffectVampirism); ok {
			si.MaxCharges[0]++
			si.MaxCharges[1]++
		}
	}
	if HasMod(mods, ModNoRecharges) {
		si.MaxCharges[0] *= 2
		si.MaxCharges[1] *= 2
	}
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
			BonusHP:      si.BonusHP,
			Advanced:     si.Advanced}}
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

// comestibleKind represents the various kinds of comestibles.
type comestibleKind int

const (
	AmbrosiaBerries comestibleKind = iota
	BerserkingFlower
	ClarityLeaves
	FirebreathPepper
	FoggySkinOnion
	LignificationFruit
	TeleportMushroom
)

// ComestibleData provides information about the various kinds of comestibles.
var ComestibleData = []itemInfo{
	AmbrosiaBerries:    {"ambrosia berries", EffectAmbrosiaBerries{}},
	BerserkingFlower:   {"berserking flower", EffectBerserkingFlower{}},
	ClarityLeaves:      {"clarity leaves", EffectClarityLeaves{}},
	FirebreathPepper:   {"firebreath pepper", EffectFirebreathPepper{}},
	FoggySkinOnion:     {"foggy-skin onion", EffectFoggySkinOnion{}},
	LignificationFruit: {"lignification fruit", EffectLignificationFruit{}},
	TeleportMushroom:   {"teleport mushroom", EffectTeleportMushroom{}},
}

// NoHealingComestibleData provides information about the various kinds of
// comestibles (for ModNoHealingComestibles).
var NoHealingComestibleData = []itemInfo{
	BerserkingFlower:   {"berserking flower", EffectBerserkingFlower{}},
	ClarityLeaves:      {"clarity leaves", EffectClarityLeaves{NoHeal: true}},
	FirebreathPepper:   {"firebreath pepper", EffectFirebreathPepper{NoHeal: true}},
	FoggySkinOnion:     {"foggy-skin onion", EffectFoggySkinOnion{NoHeal: true}},
	LignificationFruit: {"lignification fruit", EffectLignificationFruit{}},
	TeleportMushroom:   {"teleport mushroom", EffectTeleportMushroom{NoHeal: true}},
}

// GetComestibleData returns the appropiate comestible data for current mod
// selection.
func (g *Game) GetComestibleData() []itemInfo {
	if g.Mod(ModHealingCombat) {
		return NoHealingComestibleData
	}
	return ComestibleData
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
	// We keep count of static places (there's a limit of about 2 with
	// the current number of vaults).
	extraStatic := 0
	// Portal.
	portalIdxs := []int{}
	npvaults := mg.vaults[1:] // non-player vaults
	if g.Map.Level < MapLevels {
		var i int
		addPortal := func(fake bool) {
			portal := &Entity{Name: "magic portal", Rune: '>', Role: &Portal{Fake: fake}}
			i, portal.P = g.RandomVaultsPlaceWithFunc(mg, npvaults[len(npvaults)/3:], PlaceStatic, g.IntNBiasedUp)
			portalIdxs = append(portalIdxs, i+1)
			if !fake {
				g.Map.Portal = portal.P
			}
			g.AddEntity(portal)
		}
		addPortal(false)
		if g.ProcInfo.FakePortal[g.Map.Level-1] {
			extraStatic++
			addPortal(true)
			if g.Mod(ModCorruptedDungeon) && g.IntN(2) == 0 {
				extraStatic++
				addPortal(true)
				if g.IntN(2) == 0 {
					extraStatic++
					addPortal(true)
				}
			}
		}
	} else {
		var i int
		orb := &Entity{Name: "Orb of Corruption", Rune: '☼', Role: &CorruptionOrb{}}
		i, orb.P = g.RandomVaultsPlaceWithFunc(mg, npvaults[len(npvaults)/2:], PlaceStatic, g.IntNBiasedUp)
		g.Map.Orb = orb.P
		portalIdxs = append(portalIdxs, i+1)
		g.AddEntity(orb)
	}
	// Totemic spirit.
	totemPoint := func() gruid.Point {
		if g.Mod(ModCorruptedDungeon) && g.IntN(2*MapLevels) == 0 {
			return g.randomFreeItemFloor(mg)
		}
		_, p := g.RandomVaultsPlace(mg, npvaults, PlaceItem)
		return p
	}
	if spi := g.ProcInfo.Spirits[g.Map.Level-1]; spi.Idx >= 0 {
		var sp *Entity
		if spi.Advanced {
			sp = spiritEntity(g.Mods, challengeSpirits[spi.Idx])
		} else {
			sp = spiritEntity(g.Mods, secondarySpirits[spi.Idx])
		}
		spr := sp.Role.(*Spirit)
		spr.Charges = spr.MaxCharges[spr.Level]
		sp.Rune = '!'
		sp.P = totemPoint()
		g.Map.Totem = sp.P
		g.AddEntity(sp)
	} else if g.Map.Level < MapLevels {
		addEmptyTotem := func() {
			p := totemPoint()
			g.Map.Totem = p
			g.genEmptyTotemAt(p)
		}
		addEmptyTotem()
		if g.Mod(ModCorruptedDungeon) && g.IntN(5) == 0 && extraStatic < 2 {
			extraStatic++
			addEmptyTotem()
		}
	}
	// Menhirs.
	nmenhirs := g.ProcInfo.NMenhirs[g.Map.Level-1]
	if g.Mod(ModCorruptedDungeon) && g.IntN(10*MapLevels) == 0 {
		nmenhirs = 0
	}
	menhirIdxs := []menhirKind{}
	menhirIdx := func() menhirKind {
		idx := g.NextMenhirKind()
		if !g.Mod(ModCorruptedDungeon) || g.IntN(50) > 0 {
			for slices.Contains(menhirIdxs, idx) {
				// Never generate twice the same menhir in a given level.
				idx = g.NextMenhirKind()
			}
		}
		return idx
	}
	switch mg.theme {
	case ThemeWarp:
		menhirIdx = func() menhirKind { return MenhirWarping }
	case ThemeFire:
		menhirIdx = func() menhirKind { return MenhirFire }
	}
	for range min(nmenhirs, len(MenhirData)) {
		idx := menhirIdx()
		menhirIdxs = append(menhirIdxs, idx)
		tvaults := mg.vaults
		if menhirKind(idx) == MenhirWarping && mg.theme != ThemeWarp {
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
		if g.Mod(ModCorruptedDungeon) && g.IntN(5) == 0 || extraStatic >= 2 {
			me.P = g.randomFreeItemFloor(mg)
		} else {
			_, me.P = g.RandomVaultsPlace(mg, tvaults, PlaceStatic)
		}
		if g.Mod(ModCorruptedDungeon) && g.IntN(100) == 0 {
			menhir := me.Role.(*Menhir)
			menhir.Used = true
		}
		g.AddEntity(me)
	}
	// Comestibles.
	nitems := g.ProcInfo.NComestibles[g.Map.Level-1]
	addComestible := func(i int, cod itemInfo) {
		co := comestibleEntity(cod)
		if i <= max(4, (nitems+1)/2) {
			co.P = g.RandomPlace(mg, PlaceItem)
		} else {
			co.P = g.randomFreeItemFloor(mg)
		}
		g.AddEntity(co)
	}
	comestibleIdx := g.NextComestibleKind
	switch mg.theme {
	case ThemeBerserk:
		comestibleIdx = func() comestibleKind { return BerserkingFlower }
	case ThemeFire:
		comestibleIdx = func() comestibleKind { return FirebreathPepper }
	case ThemeLignification:
		comestibleIdx = func() comestibleKind { return LignificationFruit }
	case ThemeWarp:
		comestibleIdx = func() comestibleKind { return TeleportMushroom }
	}
	if g.Mod(ModCorruptedDungeon) && mg.theme == ThemeNone && g.IntN(3*MapLevels) == 0 {
		if g.IntN(2) == 0 {
			// Field of comestibles of same kind.
			cod := (g.GetComestibleData())[comestibleIdx()]
			for i := range nitems + 7 {
				addComestible(i, cod)
			}
		} else if g.IntN(3) > 0 {
			// Extra traps (of same kind) to compensate for lack of
			// comestibles.
			r := g.NextRune()
			for range 7 {
				g.genRunicTrap(mg, r)
			}
		}
		nitems = 0
	}
	switch mg.theme {
	case ThemeNone:
	case ThemeWarp:
		nitems += 6
	default:
		nitems += 4
	}
	for i := range nitems {
		addComestible(i, (g.GetComestibleData())[comestibleIdx()])
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
	if g.Map.Level == g.ProcInfo.TrapLevel || mg.theme != ThemeNone {
		rto += 4
		rtt++
		if g.Mod(ModCorruptedDungeon) {
			rto += g.IntN(6)
			rtt += g.IntN(2)
		}
	}
	trapRune := g.NextRune
	switch mg.theme {
	case ThemeBerserk:
		trapRune = func() MagicRune { return RuneBerserk }
	case ThemeFire:
		trapRune = func() MagicRune { return RuneFire }
	case ThemeLignification:
		trapRune = func() MagicRune { return RuneLignification }
	case ThemeWarp:
		trapRune = func() MagicRune { return RuneWarp }
	}
	for range rto {
		g.genRunicTrap(mg, trapRune())
	}
	for range rtt {
		g.genRunicTrapInTunnel(mg, trapRune())
	}
}

// randomFreeItemFloor returns a random passable position that is not occupied
// by an item. The returned position has some bias toward walls and sets the
// terrain to Floor.
func (g *Game) randomFreeItemFloor(mg *MapGen) gruid.Point {
	for {
		p := g.RandomPassableBiased()
		if i, _ := g.ItemAt(p); i >= 0 {
			continue
		}
		if mg.itemPlace.At(p) {
			// Avoid placement on a special item place, as we
			// couldn't have enough of them otherwise.
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
func (g *Game) genRunicTrap(mg *MapGen, r MagicRune) (ID, *Entity) {
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
			return g.genRunicTrapAt(r, p)
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
		return g.genRunicTrapAt(r, p)
	}
}

// genRunicTrapInTunnel generates a runic trap in an extra tunnel, if possible.
// It blocks locally the passage, but there's always an alternative path.
func (g *Game) genRunicTrapInTunnel(mg *MapGen, r MagicRune) (ID, *Entity) {
	if len(mg.xtunnel) == 0 {
		return g.genRunicTrap(mg, r)
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
		return g.genRunicTrapAt(r, p)
	}
	return g.genRunicTrap(mg, r)
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
func (g *Game) genRunicTrapAt(r MagicRune, p gruid.Point) (ID, *Entity) {
	g.Map.Terrain.Set(p, Floor)
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

var nMonsE = [MapLevels]int{3, 5, 6, 6, 5, 5, 4, 4, 3}
var nMonsM = [MapLevels]int{2, 3, 4, 6, 8, 10, 11, 10, 10}
var nMonsL = [MapLevels]int{0, 0, 1, 1, 2, 2, 3, 4, 6}

// genMonsters generates monster entities for the current level.
func (g *Game) genMonsters(mg *MapGen) {
	// Number of early, mid, late monsters.
	nE, nM, nL := nMonsE[g.Map.Level-1], nMonsM[g.Map.Level-1], nMonsL[g.Map.Level-1]
	// Guardians and wandering uniques.
	if g.Mod(ModCorruptedDungeon) {
		nE, nM, nL = g.genGuardiansCorrupted(mg, nE, nM, nL)
	} else {
		nE, nM, nL = g.genGuardians(nE, nM, nL)
	}
	// Special levels.
	if mg.theme != ThemeNone {
		// Rare themed level: special treatment.
		g.genThemedMonsters(mg, nE, nM, nL)
		return
	}
	switch g.Map.Level {
	case g.ProcInfo.MonsEarly:
		// Special level with many early monsters of same kind.
		nE += 2
		nM--
		mk := g.randMonsKind(monsEarly)
		for range nE - 2 {
			g.genMonster(mk)
		}
		nE = 2 // two remaining
	case g.ProcInfo.MonsMid:
		// Special level with many mid monsters of same kind.
		nE -= 2
		nM += 2
		mk := g.randMonsKind(monsMid)
		for range nM - 2 {
			g.genMonster(mk)
		}
		nM = 2 // two remaining
	case g.ProcInfo.MonsLate:
		// Special level with many late monsters of same kind.
		nE--
		nM -= 3
		nL += 3
		mk := g.randMonsKind(monsLate)
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
				g.genMonster(g.randMonsKind(md))
			}
			nE = 1
		case 1:
			// Special level biased toward more mid monsters and
			// with a smaller selection of them.
			nE -= 2
			nM++
			md := g.monsterSelection(monsMid, 3)
			for range nM - 1 {
				g.genMonster(g.randMonsKind(md))
			}
			nM = 1
		}
	}
	if g.Map.Level == g.ProcInfo.MonsMidLate {
		// Special level with a smaller selection of mid monsters.
		md := g.monsterSelection(monsMid, 4)
		for range nM - 1 {
			g.genMonster(g.randMonsKind(md))
		}
		nM = 1
	}
	// Mod adjustments.
	if g.Mod(ModCorruptedDungeon) {
		if nE > 0 && nM > 0 && g.IntN(5) == 0 {
			// Occasionally one extra late monster (even in early game).
			nE--
			nM--
			nL++
		}
		if g.Map.Level >= 4 && g.Map.Level < MapLevels && g.IntN(5*4) == 0 {
			switch g.IntN(4) {
			case 0:
				// Rare level with many imps.
				for range nE*2 + nM*2 + nL*3 {
					g.genMonster(NoisyImp)
				}
			case 1:
				// Rare spooky level with no remaining
				// wandering monsters with no drawback!
			default:
				// Rare level where all remaining monsters
				// become guardians scattered around.
				for range nE {
					g.genMonsterGuardian(g.randMonsKind(monsEarly), g.RandomWaypoint())
				}
				for range nM {
					g.genMonsterGuardian(g.randMonsKind(monsMid), g.RandomWaypoint())
				}
				for range nL {
					g.genMonsterGuardian(g.randMonsKind(monsLate), g.RandomWaypoint())
				}
			}
			nE, nM, nL = 0, 0, 0
		} else if g.IntN(6) == 0 {
			// Occasionally spawn extra uniques or guardians
			// (possibly not unique anymore).
			nE, nM, nL = g.genMonstersCorrupted(mg, nE, nM, nL)
		}
	}
	// Generate remaining monsters randomly from each category.
	for range nE {
		g.genMonster(g.randMonsKind(monsEarly))
	}
	for range nM {
		g.genMonster(g.randMonsKind(monsMid))
	}
	for range nL {
		g.genMonster(g.randMonsKind(monsLate))
	}
}

func (g *Game) genThemedMonsters(mg *MapGen, nE, nM, nL int) {
	// The idea behind each theme is that the generated monsters should
	// have some relation to it and, generally, immune monsters should be
	// rare, because it's not fun if the related consumables and runes
	// aren't useful. Some rare annoying monsters are fine.
	switch mg.theme {
	case ThemeBerserk:
		mks := []monsterKind{BerserkingSpider, HungryRat}
		for range 1 + nE {
			g.genMonster(g.randMonsKind(mks))
		}
		n := g.IntN(2)
		mksAnnoying := []monsterKind{CrazyDruid, ExplodingNadre}
		for range 1 + n {
			// Annoying monsters leading to poison+fire risk.
			g.genMonster(g.randMonsKind(mksAnnoying))
		}
		for range 1 + nM - n {
			g.genMonster(BarkingHound)
		}
		mksLate := []monsterKind{UndeadKnight, FearsomeLich}
		for range nL {
			g.genMonster(g.randMonsKind(mksLate))
		}
	case ThemeFire:
		for range 2 + nE {
			g.genMonster(ExplodingNadre)
		}
		mksAnnoying := []monsterKind{BerserkingSpider, VenomousViper}
		g.genMonster(g.randMonsKind(mksAnnoying)) // Annoying poison monster.
		for range 2 + nM {
			g.genMonster(FireLlama)
		}
		switch g.IntN(4) {
		case 0:
			// Rare annoying fire/berserk monster (risk of poison+fire).
			g.genMonster(CrazyDruid)
		case 1:
			// Rare annoying poison monster (risk of poison+fire).
			g.genMonster(MadOctopode)
		default:
			// Rare annoying fire immune monster.
			g.genMonster(BurningPhoenix)
		}
		for range nL + 2 {
			g.genMonster(WalkingTree)
		}
	case ThemeLignification:
		mks := []monsterKind{RampagingBoar, BlinkButterfly, TemporalCat, LashingFrog, AcidMound}
		for range nM + nE {
			g.genMonster(g.randMonsKind(mks))
		}
		switch g.IntN(4) {
		case 0:
			g.genMonster(BurningPhoenix)
		case 1:
			g.genMonster(FireLlama)
		default:
			g.genMonster(ExplodingNadre)
		}
		for range 2 {
			g.genMonster(WalkingTree)
		}
		mksLate := []monsterKind{FourHeadedHydra, EarthDragon}
		for range nL {
			g.genMonster(g.randMonsKind(mksLate))
		}
	case ThemeWarp:
		for range nE {
			g.genMonster(TemporalCat)
		}
		for range 2 + nM {
			g.genMonster(BlinkButterfly)
		}
		for range 1 + nL {
			g.genMonster(WarpingWraith)
		}
	}
}

// RandomWaypoint returns a random non-player waypoint, or a random free
// position if there is no such place. It doesn't reserve the place for
// exclusive use.
func (g *Game) RandomWaypoint() gruid.Point {
	if len(g.Map.Waypoints) > 0 {
		p := g.Map.Waypoints[g.IntN(len(g.Map.Waypoints))]
		if p == g.PP() {
			//
			return g.RandomPassableWithoutTrap()
		}
		return g.RandomPassableWithin(p, 5)
	}
	return g.RandomPassableWithoutTrap()
}

// randMonsKind returns a random monster kind among the given ones.
func (g *Game) randMonsKind(mks []monsterKind) monsterKind {
	return mks[g.IntN(len(mks))]
}

// genGuardians generates any monster guardians and special wandering uniques.
func (g *Game) genGuardians(nE, nM, nL int) (int, int, int) {
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
			p = g.RandomWaypoint()
		}
		g.genMonsterGuardian(BlazingGolem, p)
		g.genMonsterGuardian(BlazingGolem, p)
	}
	return nE, nM, nL
}

// genGuardiansCorrupted generates any monster guardians and special
// wandering uniques with special adjustments to account for extra corruption.
func (g *Game) genGuardiansCorrupted(mg *MapGen, nE, nM, nL int) (int, int, int) {
	if g.Map.Orb != InvalidPos {
		switch g.IntN(9) {
		case 0, 1:
			switch g.IntN(3) {
			case 0:
				n := min(nL, 2)
				for range n {
					g.genMonsterGuardian(g.randMonsKind(monsLate), g.Map.Orb)
				}
				nL -= n
			case 1:
				mk := g.randMonsKind(monsMid)
				n := min(nM, 4)
				for range n {
					g.genMonsterGuardian(mk, g.Map.Orb)
				}
				nM -= n
			}
			g.genMonster(FearsomeLich)
			g.genMonster(UndeadKnight)
		case 2, 3, 4:
			g.genMonsterGuardian(g.randMonsKind(monsLate), g.Map.Orb)
			nL--
			if g.IntN(4) == 0 {
				g.genMonsterGuardian(UndeadKnight, g.Map.Orb)
				g.genMonster(FearsomeLich)
			} else {
				g.genMonsterGuardian(FearsomeLich, g.Map.Orb)
				g.genMonster(UndeadKnight)
			}
		default:
			g.genMonsterGuardian(FearsomeLich, g.Map.Orb)
			g.genMonsterGuardian(UndeadKnight, g.Map.Orb)
		}
	}
	if g.ProcInfo.GuardianPortal2 == g.Map.Level {
		p := g.corruptedPortalGuardianPoint(mg)
		if g.IntN(3) > 0 {
			g.genMonsterGuardian(MadOctopode, p)
		} else {
			g.genMonster(MadOctopode)
			if nL > 0 && g.IntN(2) == 0 {
				g.genMonsterGuardian(g.randMonsKind(monsLate), p)
				nL--
			}
		}
	}
	if g.ProcInfo.GuardianTotem1 == g.Map.Level {
		p := g.corruptedTotemGuardianPoint(mg)
		if g.IntN(3) > 0 {
			g.genMonsterGuardian(CrazyDruid, p)
		} else {
			g.genMonster(CrazyDruid)
			if nM > 0 && g.IntN(2) == 0 {
				g.genMonsterGuardian(g.randMonsKind(monsMid), p)
				nM--
			}
		}
	}
	if g.ProcInfo.WanderingUnique1 == g.Map.Level {
		nM--
		if g.IntN(3) > 0 {
			g.genMonster(WalkingMushroom)
		} else {
			p := g.corruptedPortalGuardianPoint(mg)
			g.genMonsterGuardian(WalkingMushroom, p)
		}
	}
	if g.ProcInfo.WanderingUnique2 == g.Map.Level {
		nM--
		if g.IntN(3) > 0 {
			g.genMonster(NoisyImp)
		} else {
			p := g.corruptedPortalGuardianPoint(mg)
			g.genMonsterGuardian(NoisyImp, p)
		}
	}
	if g.ProcInfo.GuardianTotem2 == g.Map.Level {
		nE--
		p := g.corruptedTotemGuardianPoint(mg)
		if g.IntN(3) > 0 {
			g.genMonsterGuardian(TotemWasp, p)
			g.genMonsterGuardian(TotemWasp, p)
		} else {
			g.genMonster(TotemWasp)
			g.genMonster(TotemWasp)
			if nE >= 2 && g.IntN(3) == 0 {
				nmk := g.randMonsKind(monsEarly)
				g.genMonsterGuardian(nmk, p)
				g.genMonsterGuardian(nmk, p)
				nE -= 2
			}
		}
	}
	if g.ProcInfo.GuardianPortal1 == g.Map.Level {
		nM--
		mk := BlazingGolem
		if g.IntN(5) == 0 {
			mk = WarpingWraith
		}
		p := g.corruptedPortalGuardianPoint(mg)
		if g.IntN(3) > 0 {
			g.genMonsterGuardian(mk, p)
			g.genMonsterGuardian(mk, p)
		} else {
			g.genMonster(mk)
			g.genMonster(mk)
			if nM >= 2 && g.IntN(2) == 0 {
				nmk := g.randMonsKind(monsMid)
				g.genMonsterGuardian(nmk, p)
				g.genMonsterGuardian(nmk, p)
				nM -= 2
			}
		}
	}
	if g.IntN(2*MapLevels) == 0 && nE > 2 && nM > 2 {
		switch g.IntN(2) {
		case 0:
			p := g.RandomWaypoint()
			var mk monsterKind
			switch g.IntN(4) {
			case 0:
				mk = BerserkingSpider
			case 1:
				mk = ConfusingEye
			default:
				mk = HungryRat
			}
			for range 5 + g.IntN(4) {
				g.genMonsterGuardian(mk, p)
			}
			nE -= 2
			nM--
		default:
			p := g.RandomWaypoint()
			var mk monsterKind
			switch g.IntN(3) {
			case 0:
				mk = BlinkButterfly
			case 1:
				mk = ExplodingNadre
			default:
				mk = g.randMonsKind(monsMid)
			}
			for range 4 + g.IntN(2) {
				g.genMonsterGuardian(mk, p)
			}
			nE--
			nM -= 2
		}
	}
	return nE, nM, nL
}

func (g *Game) corruptedPortalGuardianPoint(mg *MapGen) gruid.Point {
	var p gruid.Point
	if g.Map.Orb != InvalidPos {
		p = g.Map.Orb
	} else {
		ids := g.AllPortalIDs()
		p = g.Entity(ids[g.IntN(len(ids))]).P
		if g.IntN(3) == 0 {
			// Higher chance of guarding the true portal.
			p = g.Map.Portal
		}
	}
	if g.IntN(4) == 0 {
		// Occasionally guard a random non-player vault.
		p = g.RandomWaypoint()
		if g.Map.Totem != InvalidPos && g.IntN(4) == 0 {
			p = g.Map.Totem
		}
	}
	return p
}

func (g *Game) corruptedTotemGuardianPoint(mg *MapGen) gruid.Point {
	if g.IntN(3) > 0 {
		return g.Map.Totem
	}
	return g.corruptedPortalGuardianPoint(mg)
}

// genMonstersCorrupted may generate extra special wandering monsters.
func (g *Game) genMonstersCorrupted(mg *MapGen, nE, nM, nL int) (int, int, int) {
	// Special early monsters.
	if g.IntN(3) == 0 {
		nEs := min(nE, 1+g.IntN(3))
		for range nEs {
			g.genMonster(TotemWasp)
		}
		nE -= nEs
	}
	// Special mid monsters.
	if g.IntN(3) == 0 {
		mks := []monsterKind{WalkingMushroom, BlazingGolem, CrazyDruid, NoisyImp}
		nMs := min(nM, 1+g.IntN(2), len(mks))
		monsSpecialMid := g.monsterSelection(mks, len(mks))
		for _, mk := range monsSpecialMid[:nMs] {
			g.genMonster(mk)
		}
		nM -= nMs
	}
	// Special late monsters.
	if g.IntN(3) == 0 {
		nLs := min(nL, 1+g.IntN(2))
		if g.IntN(3) == 0 && g.Map.Level < MapLevels {
			for range nLs {
				p := g.corruptedTotemGuardianPoint(mg)
				g.genMonsterGuardian(FearsomeLich, p)
			}
		} else {
			mks := []monsterKind{UndeadKnight, MadOctopode}
			for range nLs {
				g.genMonster(g.randMonsKind(mks))
			}
		}
		nL -= nLs
	}
	return nE, nM, nL
}

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
	mons := monster(mk)
	mons.P = p
	id := ID(len(g.Entities))
	g.Map.ActorCache.SetU(mons.P, id)
	g.AddEntity(mons)
	return id, mons
}

// monster generates a new monster entity using the given data.
func monster(mk monsterKind) *Entity {
	mi := MonsData[mk]
	a := NewActor(mi.Attack, mi.Defense, mi.HP, mi.Traits)
	a.Behavior = &Behavior{Kind: mk, Guard: InvalidPos, Target: InvalidPos}
	return &Entity{
		Name:   mi.Name,
		Rune:   mi.R,
		KnownP: InvalidPos,
		Role:   a,
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

// randomRevivalNearbyPoint returns a random free cell nearby to the given
// point and outside the player's field of view.
// func (g *Game) randomRevivalNearbyPoint(at gruid.Point, maxdist int) gruid.Point {
// 	q := g.randomFreeNearby(at, maxdist)
// 	for g.InFOV(q) {
// 		q = g.randomFreeNearby(at, maxdist)
// 	}
// 	return q
// }
