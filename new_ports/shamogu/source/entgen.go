// This file implements entity generation during map initialization.

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
		Role: NewActor(2, 1, 9, Player, NoTraits),
	}
	for i := range FirstMapID {
		g.Entities[i] = emptySlot()
	}
	g.Entities[0] = spe
	// g.debugBuild()
	g.ComputePlayerStats()
}

func (g *Game) debugBuild() {
	// g.Entities[1] = spiritEntity(g.Mods, secondarySpirits[0])
	// g.Entities[2] = spiritEntity(g.Mods, secondarySpirits[3])
	// g.Entities[1] = spiritEntity(g.Mods, challengeSpirits[4])
	// g.Entities[2] = spiritEntity(g.Mods, challengeSpirits[2])
	// g.Entities[4] = comestibleEntity(BerserkingFlower)
	// g.Entities[5] = comestibleEntity(ClarityLeaves)
	// g.Entities[6] = comestibleEntity(FirebreathPepper)
	// g.Entities[7] = comestibleEntity(FoggySkinOnion)
}

func emptySlot() *Entity {
	return &Entity{Name: "(empty slot)", P: InvalidPos, Seen: true}
}

// GenEntities generates the entities for a new level, placing them on the map.
func (g *Game) GenEntities(mg *MapGen) {
	g.Map.Orb = InvalidPos
	g.Map.Portal = InvalidPos
	g.Map.Totem = InvalidPos
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
		sp := g.Entity(0).Role.(*Spirit)
		if _, ok := sp.GetAbility().(EffectVampirism); ok {
			n := 1
			if g.Mod(ModNoRecharges) {
				n *= 2
			}
			sp.MaxCharges[0] += n
			sp.MaxCharges[1] += n
			sp.Charges += n
		}
	}
	if !g.Mod(ModNoRecharges) {
		for _, sp := range g.PlayerSpirits() {
			sp.Charges = sp.GetMaxCharges()
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
	{
		Name:         "Spinning Crocodile",
		MaxCharges:   [2]int{3, 4},
		Ability:      [2]Ability{EffectTailSlap{}, EffectTailSlap{}},
		BonusAttack:  [2]int{1, 1},
		BonusDefense: [2]int{0, 1},
		BonusTraits:  [2]Traits{PatternDragging, PatternDragging}},
	{
		Name:        "Vampiric Bat",
		MaxCharges:  [2]int{1, 2},
		Ability:     [2]Ability{EffectVampirism{DurationVampirism}, EffectVampirism{DurationVampirism}},
		BonusTraits: [2]Traits{PatternSneaky, PatternSneaky}},
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
	if HasMod(mods, ModHealingCombat) || VampiricHC {
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

	// Rare comestibles (with conditional variant).
	PolymorphFungus
	RunethornRose
	TunnelingAcorn
	VanishingSnail
	WarpingBean

	// Rare conditional comestibles (exclusive to Totem Conditions).
	FortressOyster
	SeeingPotato
	TrickyChestnut

	// Miscellaneous rare comestibles.
	MoonlightLotus
)

// Name returns the name of the comestible.
func (ck comestibleKind) Name() string { return ComestibleData[ck].Name }

// Info returns the item information for the comestible.
func (ck comestibleKind) Info() itemInfo { return ComestibleData[ck] }

// ComestibleData provides information about the various kinds of comestibles.
var ComestibleData = []itemInfo{
	// Common comestibles.
	AmbrosiaBerries:    {"ambrosia berries", EffectAmbrosiaBerries{}},
	BerserkingFlower:   {"berserking flower", EffectBerserkingFlower{}},
	ClarityLeaves:      {"clarity leaves", EffectClarityLeaves{}},
	FirebreathPepper:   {"firebreath pepper", EffectFirebreathPepper{}},
	FoggySkinOnion:     {"foggy-skin onion", EffectFoggySkinOnion{}},
	LignificationFruit: {"lignification fruit", EffectLignificationFruit{}},
	TeleportMushroom:   {"teleport mushroom", EffectTeleportMushroom{}},

	// Rare comestibles with conditional variant.
	PolymorphFungus: {"polymorph fungus", EffectPolymorphFungus{}},
	RunethornRose:   {"runethorn rose", EffectRunethornRose{}},
	TunnelingAcorn:  {"tunneling acorn", EffectTunnelingAcorn{}},
	VanishingSnail:  {"vanishing snail", EffectVanishingSnail{}},
	WarpingBean:     {"warping bean", EffectWarpingBean{}},

	// Rare conditional comestibles (exclusive to Totem Conditions).
	FortressOyster: {"fortress oyster", EffectFortressOyster{}},
	SeeingPotato:   {"seeing potato", EffectSeeingPotato{}},
	TrickyChestnut: {"tricky chestnut", EffectTrickyChestnut{}},

	// Miscellaneous rare comestibles without conditional variant.
	MoonlightLotus: {"moonlight lotus", EffectMoonlightLotus{}},
}

// CommonComestibles represents the most common comestibles.
var CommonComestibles = []comestibleKind{
	AmbrosiaBerries, BerserkingFlower, ClarityLeaves,
	FirebreathPepper, FoggySkinOnion, LignificationFruit,
	TeleportMushroom}

// RareComestibles contains the list of rare comestibles available in the base
// game.
var RareComestibles = []comestibleKind{
	MoonlightLotus,
	PolymorphFungus, RunethornRose, TunnelingAcorn, VanishingSnail, WarpingBean}

// RareNCComestibles contains the list of rare comestibles that do not have a
// non-conditional variant.
var RareNCComestibles = []comestibleKind{MoonlightLotus}

// CondComestibles provides information about the various conditional
// comestibles available in Totem Conditions.
var CondComestibles = []comestibleKind{
	PolymorphFungus, RunethornRose, TunnelingAcorn, VanishingSnail, WarpingBean,
	FortressOyster, SeeingPotato,
	TrickyChestnut, // last
}

// GetExtraRareComestibles returns info for the random extra rare comestibles
// suitable for the given map level.
func (g *Game) GetExtraRareComestibles(level int) []comestibleKind {
	if !g.Mod(ModTotemConditions) {
		return RareComestibles
	}
	n := len(CondComestibles)
	if level == 9 {
		n-- // chestnuts are useless on the last level
	}
	return CondComestibles[:n]
}

// comestibleEntity generates a new comestible Entity entity using the given
// data.
func comestibleEntity(ck comestibleKind) *Entity {
	ci := ck.Info()
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
			if g.Mod(ModCorruptedDungeon) && g.IntN(3) == 0 {
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
		if g.Mod(ModCorruptedDungeon) && g.IntN(10) == 0 {
			// Small extra chance of hiding the orb location
			// (independently of whether guardians are there or
			// not).
			g.hideLocation(orb.P)
		}
	}
	// Totemic spirit.
	totemPoint := func() gruid.Point {
		if g.Mod(ModCorruptedDungeon) && g.IntN(2*MapLevels) == 0 {
			return g.randomFreeItemFloor(mg)
		}
		_, p := g.RandomVaultsPlace(mg, npvaults, PlaceItem)
		return p
	}
	var usedLevel bool // spooky corrupted level with used runes/menhirs
	if spi := g.ProcInfo.Spirits[g.Map.Level-1]; spi.Idx >= 0 {
		var sp *Entity
		if spi.Advanced {
			sp = spiritEntity(g.Mods, challengeSpirits[spi.Idx])
		} else {
			sp = spiritEntity(g.Mods, secondarySpirits[spi.Idx])
		}
		spr := sp.Role.(*Spirit)
		spr.Charges = spr.GetMaxCharges()
		if g.Mod(ModTotemConditions) {
			spr.Condition = g.ProcInfo.Conditions[g.Map.Level-1]
		}
		sp.Rune = '!'
		sp.P = totemPoint()
		g.Map.Totem = sp.P
		g.AddEntity(sp)
	} else if g.Map.Level < MapLevels {
		if g.Mod(ModCorruptedDungeon) {
			// In Corrupted Dungeon games with an empty totem,
			// empty all runes and menhirs 1/15 times on average
			// per game (2 empty totem levels).
			usedLevel = g.IntN(30) == 0
		}
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
		if mg.theme == ThemeLignification && idx == MenhirWarping {
			return MenhirEarth
		}
		return idx
	}
	switch mg.theme {
	case ThemeWarp:
		menhirIdx = func() menhirKind { return MenhirWarping }
	case ThemeFire:
		menhirIdx = func() menhirKind { return MenhirFire }
	case ThemePoison:
		menhirIdx = func() menhirKind { return MenhirPoison }
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
	addComestible := func(i int, ck comestibleKind) {
		// NOTE: first comestibles are always placed in vaults, which
		// currently means that we typically place rare comestibles in
		// vaults. Kinda distinctive in a way, to make them stand out a
		// little more, but it might be nice to sometimes have them
		// outside of vaults, too.
		co := comestibleEntity(ck)
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
		n := 0
		comestibleIdx = func() comestibleKind {
			n++
			if n <= 1 && g.IntN(4) > 0 {
				// A single ambrosia berries on berserk levels,
				// for recovering health when at 1 HP.
				return AmbrosiaBerries
			}
			return BerserkingFlower
		}
	case ThemeFire:
		comestibleIdx = func() comestibleKind { return FirebreathPepper }
	case ThemeLignification:
		n := 0
		comestibleIdx = func() comestibleKind {
			n++
			if n <= 1 && g.IntN(4) > 0 {
				// A single foggy-skin onion on lignification
				// levels to recover movement early.
				return FoggySkinOnion
			}
			return LignificationFruit
		}
	case ThemeWarp:
		n := 0
		comestibleIdx = func() comestibleKind {
			n++
			if n <= 1 && g.IntN(4) > 0 {
				// A single clarity leaves on warp levels to
				// cure daze once after teleport.
				return ClarityLeaves
			}
			return TeleportMushroom
		}
	}
	if g.Mod(ModCorruptedDungeon) && mg.theme == ThemeNone && g.IntN(4*MapLevels) == 0 {
		if g.IntN(2) == 0 {
			// Field of comestibles of same kind.
			ck := CommonComestibles[g.IntN(len(CommonComestibles))]
			if !NoChaos && g.IntN(len(CommonComestibles)) == 0 {
				ck = MoonlightLotus // Corrupted Dungeon's signature rare comestible
				if g.IntN(3) == 0 {
					// Rarely, put a random rare comestible
					// instead.
					ck = RareComestibles[g.IntN(len(RareComestibles))]
				}
			}
			for i := range nitems + 7 {
				addComestible(i, ck)
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
	case ThemePoison, ThemeFootsteps:
		// Only one extra comestible, because it can be more difficult
		// than a normal level, in particular early on, but it still
		// has varied comestibles.
		nitems++
	default:
		nitems += 4
	}
	var xrcMap []bool // where extra rare comestibles go
	if len(g.ProcInfo.NComestiblesRare) > 0 {
		k := g.ProcInfo.NComestiblesRare[g.Map.Level-1]
		xrcMap = make([]bool, nitems)
		for i := range nitems {
			xrcMap[i] = i < k
		}
		g.rand.Shuffle(nitems, func(i, j int) {
			xrcMap[i], xrcMap[j] = xrcMap[j], xrcMap[i]
		})
	}
	for i := range nitems {
		switch {
		case len(xrcMap) > 0 && xrcMap[i]:
			if g.ProcInfo.SingleRareCom > 0 {
				addComestible(i, g.ProcInfo.SingleRareCom)
			} else {
				coms := g.GetExtraRareComestibles(g.Map.Level)
				addComestible(i, coms[g.IntN(len(coms))])
			}
		case g.Map.Level == g.ProcInfo.RareComestible1:
			// Mid-game rare comestible.
			if g.Mod(ModTotemConditions) {
				// With Totem Conditions, generate a rare
				// midgame non-conditional comestible, as
				// conditional ones are generated
				// independently as extra rare ones.
				addComestible(i, RareNCComestibles[g.IntN(len(RareNCComestibles))])
				g.ProcInfo.RareComestible1 = -1
				break
			}
			addComestible(i, RareComestibles[g.IntN(len(RareComestibles))])
			g.ProcInfo.RareComestible1 = -1
		case g.Map.Level == g.ProcInfo.RareComestible2:
			// Late-game rare comestible.
			addComestible(i, RareComestibles[g.IntN(len(RareComestibles))])
			g.ProcInfo.RareComestible2 = -1
		case g.Map.Level == g.ProcInfo.ExtraLotus && i == nitems-1:
			// Rare extra lotus in Corrupted Dungeon (outside of
			// vault, for a change).
			addComestible(i, MoonlightLotus)
			g.ProcInfo.ExtraLotus = -1
		default:
			ck := comestibleIdx()
			if ck == AmbrosiaBerries && g.Mod(ModHealingCombat) && !NoChaos {
				// Replace a few ambrosia berries with
				// polymorph fungus when Healing Combat is
				// enabled (1-2 per game).
				g.ProcInfo.WaitFungus--
				switch {
				case g.ProcInfo.WaitFungus == 0:
					// First one.
					ck = PolymorphFungus
				case g.ProcInfo.WaitFungus < 0 && g.IntN(4) == 0:
					// Sometimes, extra replacement, but
					// never more than 2.
					ck = PolymorphFungus
					g.ProcInfo.WaitFungus = MapLevels
				}
			}
			addComestible(i, ck)
		}
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
		rto += 5
		rtt++
		if g.Mod(ModCorruptedDungeon) {
			// Some extra traps for more variability.
			rto += g.IntN(3)
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
	case ThemePoison:
		rto += 5 // Extra poison runes, as they're usually not as strong as others.
		trapRune = func() MagicRune { return RunePoison }
	case ThemeWarp:
		trapRune = func() MagicRune { return RuneWarp }
	}
	for range rto {
		g.genRunicTrap(mg, trapRune())
	}
	for range rtt {
		g.genRunicTrapInTunnel(mg, trapRune())
	}
	if usedLevel {
		// Rare spooky level with only used runes and menhirs.
		for _, e := range g.NPMapEntities() {
			switch it := e.Role.(type) {
			case *RunicTrap:
				it.Used = true
			case *Menhir:
				it.Used = true
			case *Portal:
				if it.Fake {
					// Mark even fake portals as already
					// used if they ever happen on such a
					// spooky level.
					it.Used = true
				}
			}
		}
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
		if g.Mod(ModCorruptedDungeon) && g.IntN(10*MapLevels) == 0 {
			// Very small chance of hiding the location.
			g.hideLocation(p)
		}
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

// ActorKind represents the various kinds of actors.
type ActorKind int

const (
	Player ActorKind = iota

	AcidMound
	BarkingHound
	BerserkingSpider
	BlazingGolem
	BlinkButterfly
	BurningPhoenix
	ChaosMegabat
	ConfusingEye
	CrazyDruid
	DraggingAlligator
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

// GetMegabatName returns the name of a megabat with current options.
func GetMegabatName() string {
	if NoChaos {
		return "screeching megabat"
	}
	return "chaos megabat"
}

// MonsData provides information about the various kinds of monsters.
var MonsData = []monsterInfo{
	// 		  {name, rune, attack, defense, hp, traits}
	AcidMound:         {"acid mound", 'a', 2, 0, 5, MonsIgnoreDefense | ImmunityImbalance | MonsCreep},
	BarkingHound:      {"barking hound", 'h', 3, 0, 4, GoodHearing | ImmunityFear},
	BerserkingSpider:  {"berserking spider", 's', 2, 1, 3, MonsBerserking | MonsSilent},
	BlazingGolem:      {"blazing golem", 'G', 2, 3, 4, MonsExplodingDeath | ImmunityFire | ImmunityPoison | ImmunityFear | ImmunityConfusion | MonsHeavyFootsteps | MonsNotable},
	BlinkButterfly:    {"blinking butterfly", 'b', 2, 3, 2, ImmunityImbalance | ImmunityLignification | MonsWingFlap},
	BurningPhoenix:    {"burning phoenix", 'P', 2, 1, 5, PatternRampage | BurningHits | ImmunityFire | ImmunityLignification | MonsWingFlap | MonsNotable},
	ChaosMegabat:      {GetMegabatName(), 'm', 2, 2, 2, PatternSneaky | GoodHearing | ImmunityLignification | MonsWingFlap | MonsNotable},
	ConfusingEye:      {"confusing eye", 'e', 2, 0, 2, PatternRanged | MonsConfusion | ImmunityConfusion | MonsLightFootsteps},
	CrazyDruid:        {"crazy druid", 'C', 3, 1, 4, PatternSwap | BurningHits | MonsBerserking | MonsNotable},
	DraggingAlligator: {"dragging alligator", 'A', 3, 1, 4, PatternDragging | MonsScales | ResistanceFire | ResistanceFear | MonsNotable},
	EarthDragon:       {"earth dragon", 'D', 4, 2, 5, MonsDig | ImmunityFear | MonsScales | MonsHeavyFootsteps | MonsNotable},
	ExplodingNadre:    {"exploding nadre", 'n', 2, 3, 1, MonsExplodingDeath | MonsScales},
	FearsomeLich:      {"fearsome lich", 'L', 2, 2, 5, PatternRanged | MonsFear | MonsIgnoreDefense | ImmunityPoison | ImmunityConfusion | MonsNotable},
	FireLlama:         {"fire llama", 'l', 2, 0, 4, PatternRanged},
	FourHeadedHydra:   {"four-headed hydra", 'H', 2, 0, 6, PatternFourDirs | MonsScales | MonsHeavyFootsteps | MonsNotable},
	HungryRat:         {"hungry rat", 'r', 2, 0, 3, MonsLightFootsteps},
	NoisyImp:          {"noisy imp", 'I', 2, 2, 4, MonsLightFootsteps | MonsNotable},
	MadOctopode:       {"mad octopode", 'O', 3, 2, 5, PatternCatch | ImmunityImbalance | VenomousMelee | MonsCreep | MonsNotable},
	LashingFrog:       {"lashing frog", 'F', 2, 1, 4, PatternCatch},
	RampagingBoar:     {"rampaging boar", 'B', 3, 0, 4, PatternRampage | PushingCharge | MonsDig},
	TemporalCat:       {"temporal cat", 'c', 2, 0, 4, PatternSwapDaze | MonsLightFootsteps},
	ThunderPorcupine:  {"thunder porcupine", 'p', 2, 0, 3, DazingSpines | ImmunityDaze | MonsLightFootsteps},
	TotemWasp:         {"totem wasp", 'w', 2, 1, 2, PatternRampage | VenomousMelee | ImmunityLignification | MonsWingFlap | MonsNotable},
	UndeadKnight:      {"undead knight", 'K', 3, 3, 4, PatternFourDirs | MonsFear | ImmunityPoison | ImmunityDaze | MonsNotable},
	VenomousViper:     {"venomous viper", 'v', 2, 1, 4, VenomousMelee | ImmunityPoison | MonsScales | MonsCreep},
	WalkingMushroom:   {"walking mushroom", 'M', 2, 1, 5, MonsConfusion | ImmunityConfusion | ImmunityLignification | VulnerabilityFire | MonsNotable},
	WalkingTree:       {"walking tree", 'T', 2, 2, 5, ImmunityImbalance | ImmunityPoison | ImmunityLignification | VulnerabilityFire | MonsHeavyFootsteps | MonsNotable},
	WarpingWraith:     {"warping wraith", 'W', 2, 3, 3, ImmunityPoison | ImmunityLignification | MonsSilent},
	WindFox:           {"wind fox", 'f', 2, 0, 4, PatternRangedRecoil | MonsLightFootsteps},
}

// monsEarly represents early level monsters.
var monsEarly = []ActorKind{
	BerserkingSpider,
	ConfusingEye,
	HungryRat,
	ThunderPorcupine,
}

// monsMid represents monsters that appear throughout the game (but less often
// early).
var monsMid = []ActorKind{
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
var monsLate = []ActorKind{
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
		nE, nM, nL = g.genGuardiansCorrupted(nE, nM, nL)
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
	if g.Map.Level == g.ProcInfo.MonsMidLate && nM > 1 {
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
		if g.genCorruptedLevel(nE, nM, nL) {
			nE, nM, nL = 0, 0, 0
		} else {
			// Occasionally spawn extra uniques or guardians
			// (possibly not unique anymore).
			nE, nM, nL = g.genMonstersCorrupted(nE, nM, nL)
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
		mks := []ActorKind{BerserkingSpider, HungryRat}
		for range 1 + nE {
			g.genMonster(g.randMonsKind(mks))
		}
		n := g.IntN(2)
		mksAnnoying := []ActorKind{CrazyDruid, ExplodingNadre}
		for range 1 + n {
			// Annoying monsters leading to poison+fire risk.
			g.genMonster(g.randMonsKind(mksAnnoying))
		}
		for range 1 + nM - n {
			g.genMonster(BarkingHound)
		}
		mksLate := []ActorKind{UndeadKnight, FearsomeLich}
		for range nL {
			g.genMonster(g.randMonsKind(mksLate))
		}
	case ThemeFire:
		for range 2 + nE {
			g.genMonster(ExplodingNadre)
		}
		mksAnnoying := []ActorKind{BerserkingSpider, VenomousViper}
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
	case ThemeFootsteps:
		mids := []ActorKind{BarkingHound, ExplodingNadre, FireLlama, LashingFrog, RampagingBoar}
		midlates := []ActorKind{CrazyDruid, CrazyDruid, DraggingAlligator, DraggingAlligator, WalkingMushroom}
		lates := []ActorKind{UndeadKnight, FearsomeLich}
		for range nE + 2*nM/3 {
			g.genMonster(g.randMonsKind(mids))
		}
		for range (1 + nM) / 3 {
			g.genMonster(g.randMonsKind(midlates))
		}
		for range max(1, nL-1) {
			g.genMonster(g.randMonsKind(lates))
		}
	case ThemeLignification:
		switch g.IntN(4) {
		case 0:
			// Rare annoying phoenix.
			g.genMonster(BurningPhoenix)
			nL--
			nM++
		case 1:
			// Rare annoying llama.
			g.genMonster(FireLlama)
		default:
			// Annoying nadre.
			g.genMonster(ExplodingNadre)
		}
		if g.IntN(3) > 0 {
			// Often, add walking mushroom.
			g.genMonster(WalkingMushroom)
			nM--
		}
		n := 1 + g.IntN(2)
		mksRanged := []ActorKind{ConfusingEye, WindFox}
		for range n {
			// Add 1-2 ranged mildly-annoying monsters.
			g.genMonster(g.randMonsKind(mksRanged))
		}
		mks := []ActorKind{RampagingBoar, BlinkButterfly, TemporalCat, LashingFrog, AcidMound, BarkingHound}
		for range nM + nE - n {
			// Many mid-monsters whose attack pattern or effect is
			// neutralized in some way by lignification.
			g.genMonster(g.randMonsKind(mks))
		}
		for range 2 {
			// A couple of walking trees, for some extra
			// involuntary lignification.
			g.genMonster(WalkingTree)
		}
		mksLate := []ActorKind{FourHeadedHydra, EarthDragon}
		for i := range nL {
			// Fill with late monsters whose attack pattern is
			// neutralized by lignification.
			if i < (1+nL)/2 {
				// Ensure some hydras, as their interaction
				// with lignification is more significant than
				// for dragons.
				g.genMonster(FourHeadedHydra)
				continue
			}
			g.genMonster(g.randMonsKind(mksLate))
		}
	case ThemePoison:
		switch g.IntN(4) {
		case 0:
			// Rare very annoying phoenix.
			g.genMonster(BurningPhoenix)
			nL--
			nE++
		case 1:
			// Rare annoying llama.
			g.genMonster(FireLlama)
		case 2:
			// Rare annoying nadre.
			g.genMonster(ExplodingNadre)
		default:
			// Rare very annoying crazy druid.
			g.genMonster(CrazyDruid)
			nE--
		}
		switch g.IntN(5) {
		case 0:
			// Rarely, no lignifying monster.
		case 1, 2:
			g.genMonster(WalkingTree)
			nL--
		default:
			g.genMonster(WalkingMushroom)
			nM--
		}
		n := 2 + g.IntN(2)
		for range n {
			// Confusion-resistant monster.
			g.genMonster(ConfusingEye)
		}
		mks := []ActorKind{BerserkingSpider, TotemWasp}
		for range 4 + nE - n {
			// Poisonous early monsters.
			g.genMonster(g.randMonsKind(mks))
		}
		n = (1 + nM) / 2
		for range n {
			// A fair amount of annoying vipers that are immune to
			// poison, but not to the confusion effects.
			g.genMonster(VenomousViper)
		}
		for range nM - n {
			// Poison-vulnerable monster that can still be annoying
			// at range.
			g.genMonster(RampagingBoar)
		}
		for range 2 {
			// A strong but not immune venomous monster.
			g.genMonster(MadOctopode)
		}
		for range nL {
			// Dangerous late monster vulnerable to poison, and a
			// digging one, like boars.
			g.genMonster(EarthDragon)
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

func (g *Game) genCorruptedLevel(nE, nM, nL int) bool {
	if g.Map.Level < 4 || g.Map.Level >= MapLevels || g.IntN(5*4) != 0 {
		return false
	}
	// Special corrupted non-thematic levels can appear in map levels 4-8.
	// Unlike thematic levels, they can combine with previous features.
	// Chance is about 1/4 per game, and each corruption has 1/16 chance.
	switch g.IntN(4) {
	case 0:
		if g.ProcInfo.ImpLevel {
			// Never two imp levels in a single run, as it's a very
			// noticeable event.
			return false
		}
		g.ProcInfo.ImpLevel = true
		impPatrol := func(mk ActorKind) {
			p := g.RandomWaypoint()
			ps := append(g.patrolPointCandidates(p), p)
			ps = ps[:min(len(ps), 2+g.IntN(2))]
			g.genMonsterPatroller(mk, ps)
			g.genMonsterPatroller(NoisyImp, ps)
			g.genMonsterPatroller(NoisyImp, ps)
		}
		// Rare level with many imps and different possible behaviors.
		if g.IntN(2) == 0 {
			// Patrolling imp+imp+(bat|butterfly|nadre|hound) trio.
			mk := g.randMonsKind([]ActorKind{ChaosMegabat, ChaosMegabat,
				BlinkButterfly, ExplodingNadre, BarkingHound})
			for range (1 + nE + nM + nL) / 2 {
				impPatrol(mk)
			}
		} else {
			// Wandering imps in great numbers.
			for range nE*2 + nM*2 + nL*3 {
				g.genMonster(NoisyImp)
			}
		}
	case 1:
		// Rare spooky level with no remaining wandering monsters with
		// no drawback!
	case 2:
		// Rare level where all remaining monsters become patrolling
		// pairs between random waypoints.
		for range (1 + nE) / 2 {
			g.placeRandomPatrollers(g.randMonsKind(monsEarly), g.RandomWaypoint(), 2)
		}
		for range nM / 2 {
			g.placeRandomPatrollers(g.randMonsKind(monsMid), g.RandomWaypoint(), 2)
		}
		for range (1 + nL) / 2 {
			g.placeRandomPatrollers(g.randMonsKind(monsLate), g.RandomWaypoint(), 2)
		}
	default:
		// Rare level where all remaining monsters become guardians
		// scattered around.
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
	return true
}

// RandomWaypoint returns a random non-player waypoint, or a random free
// position if there is no such place. It doesn't reserve the place for
// exclusive use.
func (g *Game) RandomWaypoint() gruid.Point {
	if len(g.Map.Waypoints) > 0 {
		pp := g.PP()
		p := g.Map.Waypoints[g.IntN(len(g.Map.Waypoints))]
		if paths.DistanceManhattan(pp, p) <= MaxFOVRange {
			// Try again once if we got too close to the player.
			p = g.Map.Waypoints[g.IntN(len(g.Map.Waypoints))]
			if p == pp {
				// Non-vault waypoint if we got the player's
				// waypoint.
				return g.randomNonVaultWaypoint()
			}
		}
		return g.RandomPassableWithin(p, 5)
	}
	return g.randomNonVaultWaypoint()
}

func (g *Game) randomNonVaultWaypoint() gruid.Point {
	pp := g.PP()
	p := g.RandomPassableWithoutTrap()
	if paths.DistanceManhattan(pp, p) <= MaxFOVRange {
		// Try again once if we got too close to the player.
		p = g.RandomPassableWithoutTrap()
	}
	return p
}

// randMonsKind returns a random monster kind among the given ones.
func (g *Game) randMonsKind(mks []ActorKind) ActorKind {
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
	if g.ProcInfo.GuardianTotem2 == g.Map.Level {
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
	if g.ProcInfo.GuardianTotem1 == g.Map.Level {
		nE--
		g.genMonsterGuardian(TotemWasp, g.Map.Totem)
		g.genMonsterGuardian(TotemWasp, g.Map.Totem)
	}
	if g.ProcInfo.GuardianEarly == g.Map.Level {
		nM--
		if g.ProcInfo.GuardianEarly <= 2 {
			nE-- // compensation for early bats
		}
		p := g.earlyGuardianPoint()
		g.placeRandomPatrollers(ChaosMegabat, p, g.numberOfBats())
	}
	if g.ProcInfo.GuardianPortal1 == g.Map.Level {
		nM--
		p := g.Map.Portal
		if g.IntN(4) == 0 {
			// Occasionally make golems guard a random non-player
			// vault.
			p = g.RandomWaypoint()
			if g.IntN(3) == 0 {
				// A chance of hiding the unguarded portal.
				g.hideLocation(g.Map.Portal)
			}
		}
		mk := BlazingGolem
		if g.IntN(3) == 0 {
			mk = DraggingAlligator
		}
		g.genMonsterGuardian(mk, p)
		g.genMonsterGuardian(mk, p)
	}
	return max(0, nE), max(0, nM), max(0, nL)
}

// placeRandomPatrollers places n monsters with (various) patrolling behaviors
// chosen at random. NOTE: with chaos megabat in mind only for now (and
// some special corruptions).
func (g *Game) placeRandomPatrollers(mk ActorKind, p gruid.Point, n int) {
	ps := g.patrolPointCandidates(p)
	k := g.IntN(4)
	switch {
	case len(ps) == 0:
		// Should usually not happen.
		for range n {
			g.genMonsterGuardian(mk, p)
		}
	case k == 0:
		// Random patrol point for each monster.
		for range n {
			g.genMonsterPatroller(mk, []gruid.Point{p, ps[g.IntN(len(ps))]})
		}
	case k == 1 && len(ps) > 1:
		// Sometimes, extra patrol point.
		ij := g.rand.Perm(len(ps))[:2]
		qs := []gruid.Point{ps[ij[0]], ps[ij[1]]}
		for range n {
			g.genMonsterPatroller(mk, append([]gruid.Point{p}, qs...))
		}
	default:
		// Same patrol point for all monsters.
		q := ps[g.IntN(len(ps))]
		for range n {
			g.genMonsterPatroller(mk, []gruid.Point{p, q})
		}
	}
}

func (g *Game) numberOfBats() int {
	if g.ProcInfo.GuardianEarly <= 4 || g.IntN(3) == 0 {
		return 3
	}
	// Usually extra bat when they appear late.
	return 4
}

func (g *Game) earlyGuardianPoint() gruid.Point {
	if g.IntN(6) == 0 {
		// Rarely, guard a random vault but hide at least either the
		// portal or totem.
		if g.IntN(2) == 0 {
			g.hideLocation(g.Map.Portal)
		} else {
			g.hideLocation(g.Map.Totem)
		}
		return g.RandomWaypoint()
	}
	if g.IntN(2) == 0 && (g.Map.Level > 1 || g.IntN(5) == 0) {
		// Guard the totem about half of the time, except on level 1 or
		// when totem is already guarded (much lower chance in those
		// cases).
		switch g.ProcInfo.GuardianEarly {
		case g.ProcInfo.GuardianTotem1, g.ProcInfo.GuardianTotem2:
			if g.IntN(4) == 0 {
				return g.Map.Totem
			}
		default:
			return g.Map.Totem
		}
	}
	if g.Mod(ModCorruptedDungeon) && g.IntN(3) > 0 {
		// High chance of guarding any portal.
		ids := g.AllPortalIDs()
		return g.Entity(ids[g.IntN(len(ids))]).P
	}
	return g.Map.Portal
}

// genGuardiansCorrupted generates any monster guardians and special
// wandering uniques with special adjustments to account for extra corruption.
func (g *Game) genGuardiansCorrupted(nE, nM, nL int) (int, int, int) {
	if g.Map.Orb != InvalidPos {
		p := g.corruptedPortalGuardianPoint()
		switch g.IntN(9) {
		case 0, 1:
			nE, nM, nL = g.genOrbAlternateGuardiansAt(p, nE, nM, nL)
			g.genMonster(FearsomeLich)
			g.genMonster(UndeadKnight)
		case 2, 3, 4:
			nE, nM, nL = g.genOrbComplementaryGuardiansAt(p, nE, nM, nL)
			if g.IntN(4) == 0 {
				g.genMonsterGuardian(UndeadKnight, p)
				g.genMonster(FearsomeLich)
			} else {
				g.genMonsterGuardian(FearsomeLich, p)
				g.genMonster(UndeadKnight)
			}
		default:
			if g.ProcInfo.SingleGuardKind == Player {
				g.genMonsterGuardian(FearsomeLich, p)
				g.genMonsterGuardian(UndeadKnight, p)
				break
			}
			for range 3 {
				g.genMonsterGuardian(g.ProcInfo.SingleGuardKind, p)
			}
		}
		if q := g.Map.Orb; p != q && g.IntN(2) == 0 {
			switch g.IntN(9) {
			case 0, 1:
				nL -= 2
				if g.ProcInfo.SingleGuardKind == Player {
					g.genMonsterGuardian(FearsomeLich, q)
					g.genMonsterGuardian(UndeadKnight, q)
					break
				}
				for range 3 {
					g.genMonsterGuardian(g.ProcInfo.SingleGuardKind, p)
				}
			case 2, 3, 4:
				nE, nM, nL = g.genOrbComplementaryGuardiansAt(q, nE, nM, nL)
				if g.IntN(4) == 0 {
					g.genMonsterGuardian(UndeadKnight, q)
				} else {
					g.genMonsterGuardian(FearsomeLich, q)
				}
				nL--
			default:
				nE, nM, nL = g.genOrbAlternateGuardiansAt(q, nE, nM, nL)
			}
			if g.IntN(3) > 0 {
				// Also hide p with high chance, so that it's
				// not obvious which guardians are guarding the
				// orb and which ones are guarding a random
				// vault.
				g.hideLocation(p)
			}
		}
	}
	if g.ProcInfo.GuardianPortal2 == g.Map.Level {
		p := g.corruptedPortalGuardianPoint()
		k, gmk := 1, MadOctopode
		if g.ProcInfo.SingleGuardKind != Player {
			gmk = g.ProcInfo.SingleGuardKind
			nE--
			k = 2
		}
		if g.IntN(4) > 0 {
			for range k {
				g.genMonsterGuardian(gmk, p)
			}
		} else {
			for range k {
				g.genMonster(gmk)
			}
			switch g.IntN(3) {
			case 0:
				mk := g.randMonsKind(monsMid)
				for range 2 {
					g.genMonsterGuardian(mk, p)
				}
				g.genMonsterGuardian(g.randMonsKind(monsEarly), p)
				nM -= 2
				nE--
			case 1:
				g.genMonsterGuardian(g.randMonsKind(monsLate), p)
				g.genMonsterGuardian(g.randMonsKind(monsEarly), p)
				nL--
				nE--
			}
		}
	}
	getGuardKind := func(mk ActorKind) ActorKind {
		if g.ProcInfo.SingleGuardKind == Player {
			return mk
		}
		// Replaces guard kind (corrupted dungeon feature).
		return g.ProcInfo.SingleGuardKind
	}
	if g.ProcInfo.GuardianTotem2 == g.Map.Level {
		p := g.corruptedTotemGuardianPoint()
		if g.IntN(4) > 0 {
			g.genMonsterGuardian(getGuardKind(CrazyDruid), p)
		} else {
			g.genMonster(getGuardKind(CrazyDruid))
			if g.IntN(3) > 0 {
				mk := g.randMonsKind(monsEarly)
				for range 3 {
					g.genMonsterGuardian(mk, p)
				}
				nE -= 3
			}
		}
	}
	if g.ProcInfo.WanderingUnique1 == g.Map.Level {
		nM--
		if g.IntN(3) > 0 {
			g.genMonster(getGuardKind(WalkingMushroom))
		} else {
			p := g.corruptedPortalGuardianPoint()
			g.genMonsterGuardian(getGuardKind(WalkingMushroom), p)
		}
	}
	if g.ProcInfo.WanderingUnique2 == g.Map.Level {
		nM--
		if g.IntN(3) > 0 {
			g.genMonster(getGuardKind(NoisyImp))
		} else {
			p := g.corruptedPortalGuardianPoint()
			g.genMonsterGuardian(getGuardKind(NoisyImp), p)
		}
	}
	if g.ProcInfo.GuardianTotem1 == g.Map.Level {
		nE--
		p := g.corruptedTotemGuardianPoint()
		if g.IntN(4) > 0 {
			mk := TotemWasp
			if g.IntN(15) == 0 {
				mk = BurningPhoenix
				nM -= 2
			} else if g.ProcInfo.SingleGuardKind != Player {
				mk = g.ProcInfo.SingleGuardKind
				nM--
			}
			g.genMonsterGuardian(mk, p)
			if g.IntN(30) > 0 {
				g.genMonsterGuardian(mk, p)
			} else {
				nE++
			}
		} else {
			g.genMonster(TotemWasp)
			g.genMonster(TotemWasp)
			if g.IntN(3) > 0 {
				nmk := g.randMonsKind(monsEarly)
				g.genMonsterGuardian(nmk, p)
				g.genMonsterGuardian(nmk, p)
				nE -= 2
			}
		}
	}
	if g.ProcInfo.GuardianEarly == g.Map.Level {
		nM--
		if g.ProcInfo.GuardianEarly <= 2 {
			nE-- // compensation for early bats
		}
		p := g.earlyGuardianPoint()
		mk := getGuardKind(ChaosMegabat)
		g.placeRandomPatrollers(mk, p, g.numberOfBats())
		if mk != ChaosMegabat {
			// The alternate ones are stronger, so compensate.
			nM--
			nE--
		}
		if g.Map.Level >= 3 && g.IntN(3) == 0 {
			// Rarely, add an imp guardian friend.
			g.genMonsterGuardian(NoisyImp, p)
			nE--
		}
	}
	if g.ProcInfo.GuardianPortal1 == g.Map.Level {
		nM--
		mk := BlazingGolem
		if g.IntN(3) == 0 {
			mk = DraggingAlligator
		}
		if g.IntN(9) == 0 {
			mk = WarpingWraith
		}
		mk = getGuardKind(mk)
		p := g.corruptedPortalGuardianPoint()
		if g.IntN(4) > 0 {
			g.genMonsterGuardian(mk, p)
			if g.IntN(30) > 0 {
				g.genMonsterGuardian(mk, p)
			} else {
				nM++
			}
		} else {
			g.genMonster(mk)
			g.genMonster(mk)
			if g.IntN(3) > 0 {
				nmk := g.randMonsKind(monsMid)
				g.genMonsterGuardian(nmk, p)
				g.genMonsterGuardian(nmk, p)
				nM -= 2
			}
		}
	}
	if g.IntN(3*MapLevels) == 0 && nE >= 2 && nM >= 2 {
		switch g.IntN(2) {
		case 0:
			p := g.RandomWaypoint()
			var mk ActorKind
			switch g.IntN(4) {
			case 0:
				mk = BerserkingSpider
			case 1:
				mk = ConfusingEye
			default:
				mk = HungryRat
			}
			n := 5 + g.IntN(g.Map.Level/2)
			if g.Map.Level == 1 {
				n -= g.IntN(2) + g.IntN(2)
			}
			for range n {
				g.genMonsterGuardian(mk, p)
			}
			nE -= 1 + n/5
			nM--
		default:
			p := g.RandomWaypoint()
			var mk ActorKind
			switch g.IntN(4) {
			case 0:
				mk = BlinkButterfly
			case 1, 2:
				mk = ExplodingNadre
			default:
				mk = g.randMonsKind(monsMid)
			}
			n := 3 + g.IntN(1+g.Map.Level/4)
			if g.Map.Level == 1 {
				n -= g.IntN(2)
			}
			for range n {
				g.genMonsterGuardian(mk, p)
			}
			nE--
			nM -= 1 + n/4
		}
	}
	return max(0, nE), max(0, nM), max(0, nL)
}

func (g *Game) genOrbAlternateGuardiansAt(p gruid.Point, nE, nM, nL int) (int, int, int) {
	switch g.IntN(7) {
	case 0:
		for range 2 {
			g.genMonsterGuardian(g.randMonsKind(monsLate), p)
		}
		g.genMonsterGuardian(g.randMonsKind(monsEarly), p)
		nL -= 2
		nE--
		if g.IntN(3) > 0 {
			// High chance of hiding the orb with
			// normal-looking guardians.
			g.hideLocation(p)
		}
	case 1:
		for range 2 {
			g.genMonsterGuardian(CrazyDruid, p)
		}
		mk := BlazingGolem
		if g.IntN(2) == 0 {
			mk = DraggingAlligator
		}
		g.genMonsterGuardian(mk, p)
		nL -= 2
		nM--
	case 2:
		for range 2 {
			g.genMonsterGuardian(MadOctopode, p)
		}
		nL -= 2
		nE--
	case 3:
		// Rare chance of annoying guardians.
		if g.IntN(2) == 0 {
			for range 2 {
				g.genMonsterGuardian(WarpingWraith, p)
			}
		} else {
			g.genMonsterGuardian(WalkingMushroom, p)
			g.genMonsterGuardian(BurningPhoenix, p)
		}
		nL -= 2
	default:
		mk := g.randMonsKind(monsMid)
		n := 4
		if g.IntN(1+len(monsMid)) == 0 {
			// Sneaky megabat nest as rare orb guardians, too.
			mk = ChaosMegabat
			n += 2
		}
		if mk == ExplodingNadre {
			// Nadres are frail, so add an extra one to compensate.
			n++
			nE++
		}
		for range n {
			g.genMonsterGuardian(mk, p)
			nM--
		}
	}
	return nE, nM, nL
}

func (g *Game) genOrbComplementaryGuardiansAt(p gruid.Point, nE, nM, nL int) (int, int, int) {
	switch g.IntN(4) {
	case 0:
		g.genMonsterGuardian(CrazyDruid, p)
		if g.IntN(2) == 0 {
			g.genMonsterGuardian(BlazingGolem, p)
		} else {
			g.genMonsterGuardian(DraggingAlligator, p)
		}
	case 1:
		g.genMonsterGuardian(MadOctopode, p)
	case 2:
		g.genMonsterGuardian(WarpingWraith, p)
	default:
		g.genMonsterGuardian(g.randMonsKind(monsLate), p)
		g.genMonsterGuardian(g.randMonsKind(monsEarly), p)
	}
	nL--
	nE--
	return nE, nM, nL
}

func (g *Game) corruptedPortalGuardianPoint() gruid.Point {
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
	if g.IntN(5) == 0 && (g.Map.Orb == InvalidPos || g.IntN(3) > 0) {
		// Occasionally guard a random non-player vault.
		p = g.RandomWaypoint()
	}
	switch {
	case g.Map.Orb != InvalidPos && g.Map.Orb != p && g.IntN(3) > 0:
		// Hide the orb with high chance if guardians are elsewhere.
		g.hideLocation(g.Map.Orb)
	case g.Map.Portal != InvalidPos && g.Map.Portal != p && g.IntN(3) == 0:
		// Sometimes hide the portal when guardians are elsewhere.
		g.hideLocation(g.Map.Portal)
	}
	return p
}

func (g *Game) corruptedTotemGuardianPoint() gruid.Point {
	if g.IntN(5) > 0 {
		return g.Map.Totem
	}
	if g.IntN(3) > 0 {
		g.hideLocation(g.Map.Totem)
	}
	return g.RandomWaypoint()
}

// hideLocation attempts to hide the given point by adding foliage or rubble
// around in a parsimonious way.
func (g *Game) hideLocation(at gruid.Point) {
	Diags := []gruid.Point{
		{1, 1},
		{1, -1},
		{-1, 1},
		{-1, -1},
	}
	// Shuffle to avoid introducing any weird bias when placing foliage and
	// rubble.
	g.rand.Shuffle(len(Diags), func(i, j int) { Diags[i], Diags[j] = Diags[j], Diags[i] })
	lt := &lighter{g: g}
	for _, dir := range Diags {
		g.hideFrom(lt, at, dir)
	}
}

func (g *Game) hideFrom(lt *lighter, from, dir gruid.Point) {
	to := from.Add(dir)
	p, q := from.Add(gruid.Point{dir.X, 0}), from.Add(gruid.Point{0, dir.Y})
	fill := func(p gruid.Point) {
		if id, _ := g.ItemAt(p); id >= 0 {
			return
		}
		switch g.IntN(10) {
		case 0:
			// Rarely: do not hide from that direction.
		case 1, 2:
			g.Map.Terrain.Set(p, Rubble)
		default:
			g.Map.Terrain.Set(p, Foliage)
		}
	}
	if g.Map.Terrain.At(p) == Floor {
		fill(p)
	}
	if g.Map.Terrain.At(q) == Floor {
		fill(q)
	}
	vis := lt.diagonalVisibility(from, to)
	switch vis {
	case Opaque:
		// Already hidden diagonally.
	case Fuzzy:
		// Already fuzzy diagonally: use normal fill favoring foliage,
		// as it's quite enough.
		if g.Map.Terrain.At(to) == Floor {
			fill(to)
		}
	default:
		// Clear: either one or both laterals are not walls, foliage
		// nor rubble (so floor or translucent walls). Attempt to cover
		// a maximum by putting some rubble.
		if id, _ := g.ItemAt(to); id < 0 && g.Map.Terrain.At(to) == Floor {
			g.Map.Terrain.Set(to, Rubble)
		}
	}
}

// genMonstersCorrupted may generate extra special wandering monsters.
func (g *Game) genMonstersCorrupted(nE, nM, nL int) (int, int, int) {
	const rare = 18
	// Special early monsters.
	if g.IntN(rare) == 0 && nE >= 0 {
		nEs := min(nE, 1+g.IntN(3))
		for range nEs {
			g.genMonster(TotemWasp)
		}
		nE -= nEs
	}
	// Special mid monsters.
	if g.IntN(rare) == 0 && nM >= 0 {
		mks := []ActorKind{WalkingMushroom, BlazingGolem, CrazyDruid, NoisyImp}
		if g.IntN(3) == 0 {
			// Sometimes replace the wandering golem by an alligator.
			mks[1] = DraggingAlligator
		}
		if g.IntN(3) == 0 {
			// Sometimes replace the crazy druid by a bat.
			mks[2] = ChaosMegabat
			g.genMonster(ChaosMegabat) // extra one (because weaker)
		}
		nMs := min(nM, 1+g.IntN(2), len(mks))
		monsSpecialMid := g.monsterSelection(mks, len(mks))
		for _, mk := range monsSpecialMid[:nMs] {
			g.genMonster(mk)
		}
		nM -= nMs
	}
	// Special late monsters.
	if g.IntN(rare) == 0 && nL >= 0 {
		nLs := min(nL, 1+g.IntN(2))
		if g.IntN(3) == 0 && g.Map.Level < MapLevels {
			// Rare chance of special late monsters as totem
			// guardians.
			mks := []ActorKind{FearsomeLich, FearsomeLich, MadOctopode}
			for range nLs {
				p := g.corruptedTotemGuardianPoint()
				g.genMonsterGuardian(g.randMonsKind(mks), p)
			}
		} else {
			mks := []ActorKind{UndeadKnight, UndeadKnight, MadOctopode}
			for range nLs {
				g.genMonster(g.randMonsKind(mks))
			}
		}
		nL -= nLs
		if nLs > 1 {
			// Difficulty adjustment in case of two late special
			// monsters.
			nE--
		}
	}
	return nE, nM, nL
}

// monsterSelection returns a random monster kind subset of size n.
func (g *Game) monsterSelection(mks []ActorKind, n int) []ActorKind {
	md := slices.Clone(mks)
	g.rand.Shuffle(len(md), func(i, j int) {
		md[i], md[j] = md[j], md[i]
	})
	md = md[:min(n, len(md))]
	return md
}

// genMonster spawns a monster of the given kind on a random free floor.
func (g *Game) genMonster(mk ActorKind) (ID, *Entity) {
	p := g.PP()
	q := g.randomFreeFloor()
	// XXX: use walkable-path distance?
	for paths.DistanceManhattan(p, q) <= MaxFOVRange+1 {
		q = g.randomFreeFloor()
	}
	return g.genMonsterAt(mk, q)
}

// genMonsterGuardian spawns a guardian monster of the given kind around a
// certain position.
func (g *Game) genMonsterGuardian(mk ActorKind, at gruid.Point) (ID, *Actor) {
	p := g.randomFreeNearby(at, 5)
	id, mons := g.genMonsterAt(mk, p)
	a := mons.Actor()
	a.Behavior.Guard = []gruid.Point{at}
	return id, a
}

// genMonsterPatroller spawns a patrolling monster between certain positions.
// Assumes a non-empty list of points.
func (g *Game) genMonsterPatroller(mk ActorKind, ps []gruid.Point) (ID, *Actor) {
	p := g.randomFreeNearby(ps[g.IntN(len(ps))], 5)
	id, mons := g.genMonsterAt(mk, p)
	a := mons.Actor()
	a.Behavior.Guard = ps
	return id, a
}

// patrolPointCandidates returns a list of suitable patrol destinations sorted
// by distance to from (excluding from).
func (g *Game) patrolPointCandidates(from gruid.Point) []gruid.Point {
	var ps []gruid.Point
	// Collect item positions that aren't too close nor too far.
	pp := g.PP()
	maxdist := 3 * MaxFOVRange
	dij := &MapPath{passable: g.Map.PassableWithoutTraps}
	g.PR.BreadthFirstMap(dij, []gruid.Point{from}, maxdist)
	for _, e := range g.NPMapEntities() {
		d := g.PR.BreadthFirstMapAt(e.P)
		if e.IsItem() && d >= MaxFOVRange && d <= maxdist && paths.DistanceManhattan(e.P, pp) > MaxFOVRange {
			ps = append(ps, e.P)
		}
	}
	if len(ps) > 0 {
		return ps
	}
	// When there are no nearby items, try choosing a random
	// waypoint at an appropriate distance.
	var qs []gruid.Point
	for _, p := range g.Map.Waypoints {
		d := g.PR.BreadthFirstMapAt(p)
		if d >= MaxFOVRange && d <= 2*MaxFOVRange && paths.DistanceManhattan(p, pp) > MaxFOVRange {
			qs = append(qs, p)
		}
	}
	if len(qs) > 0 {
		ps = []gruid.Point{g.RandomPassableWithin(qs[g.IntN(len(qs))], 5)}
	}
	return ps
}

// genMonsterAt spawns a monster of the given kind on the given position
// (assumed to be free).
func (g *Game) genMonsterAt(mk ActorKind, p gruid.Point) (ID, *Entity) {
	mons := monster(mk)
	mons.P = p
	id := ID(len(g.Entities))
	g.Map.ActorCache.SetU(mons.P, id)
	g.AddEntity(mons)
	return id, mons
}

// monster generates a new monster entity using the given data.
func monster(mk ActorKind) *Entity {
	mi := MonsData[mk]
	a := NewActor(mi.Attack, mi.Defense, mi.HP, mk, mi.Traits)
	a.Behavior = &Behavior{Target: InvalidPos}
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
