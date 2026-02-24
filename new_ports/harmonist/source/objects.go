package main

import (
	"errors"
	"fmt"

	"codeberg.org/anaseto/gruid"
)

// Objects gathers all map information about Objects.
type Objects struct {
	Stairs  map[gruid.Point]Stair
	Stones  map[gruid.Point]Stone
	Magaras map[gruid.Point]Magara
	Barrels map[gruid.Point]bool
	Bananas map[gruid.Point]bool
	Lights  map[gruid.Point]bool // true: on, false: off
	Scrolls map[gruid.Point]Scroll
	Story   map[gruid.Point]Story
	Lore    map[gruid.Point]int
	Items   map[gruid.Point]Item
	Potions map[gruid.Point]Potion
}

// Stair represents the various kinds of stair objects.
type Stair int

const (
	NormalStair Stair = iota
	WinStair
	SealedStair
	FakeStair
	KnownFakeStair
)

func (st Stair) String() (desc string) {
	switch st {
	case NormalStair, FakeStair:
		desc = "stairs"
	case WinStair:
		desc = "monolith portal"
	case SealedStair:
		desc = "sealed stairs"
	case KnownFakeStair:
		desc = "fake stairs"
	}
	return desc
}

const normalStairShortDesc = "stairs downwards"
const deepStairShortDesc = "deep stairs downwards"

func (st Stair) ShortString(g *Game) (desc string) {
	switch st {
	case NormalStair, FakeStair:
		if g.Map.Depth == WinDepth {
			desc = deepStairShortDesc
		} else {
			desc = normalStairShortDesc
		}
	case WinStair:
		desc = "monolith portal"
	case SealedStair:
		desc = "blocked " + NormalStair.ShortDesc(g)
	case KnownFakeStair:
		desc = "fake stairs"
	}
	return desc
}

func (st Stair) ShortDesc(g *Game) (desc string) {
	switch st {
	case NormalStair:
		if g.Map.Depth == WinDepth {
			desc = deepStairShortDesc
		} else {
			desc = normalStairShortDesc
		}
	case WinStair:
		desc = "a monolith portal"
	case SealedStair:
		desc = "blocked " + NormalStair.ShortDesc(g)
	case KnownFakeStair:
		desc = "fake stairs"
	}
	return desc
}

const normalStairDesc = "Stairs lead to the next level of Dayoriah Clan’s domain in Hareka’s Underground. You will not be able to come back, because an oric barrier seals the stairs when they are traversed by intruders. The upside of this is that enemies cannot follow you either."
const deepStairDesc = "Those very deep stairs lead to the next level of Dayoriah Clan’s domain in Hareka’s Underground. You will not be able to come back, because an oric barrier seals the stairs when they are traversed by intruders. The upside of this is that enemies cannot follow you either."

func (st Stair) Desc(g *Game) (desc string) {
	switch st {
	case WinStair:
		desc = "Going through this portal will make you escape from this place, going back to the Surface."
		if g.Map.Depth < MaxDepth {
			desc += " If you’re courageous enough, you may skip this portal and continue going deeper in the dungeon, to find Marevor’s magara, finishing Shaedra’s failed mission."
		}
	case NormalStair, FakeStair:
		desc = normalStairDesc
		if g.Map.Depth == WinDepth {
			desc = deepStairDesc
			desc += " You may want to take those after freeing Shaedra from her cell."
		}
	case SealedStair:
		desc = "Stairs lead to the next level of the Dayoriah Clan’s domain in Hareka’s Underground. These are sealed by an oric magical barrier that you have to disable by activating a corresponding seal stone. You will not be able to come back, because an oric barrier seals the stairs again when they are traversed by intruders. The upside of this is that enemies cannot follow you either."
	case KnownFakeStair:
		desc = "An illusion made by a harmonic celmist. It looks like stairs downwards, but it’s all fake."
	}
	return desc
}

func (st Stair) Style(g *Game) (r rune, fg gruid.Color) {
	r = '>'
	switch st {
	case WinStair:
		fg = ColorCyan
		r = 'Δ'
	case NormalStair, FakeStair:
		fg = ColorMagenta
		if g.Map.Depth == WinDepth {
			fg = ColorViolet
		}
	case SealedStair:
		fg = ColorCyan
	case KnownFakeStair:
		fg = ColorForeground
	}
	return r, fg
}

// Stone represents the various kinds of magical stones.
type Stone int

const (
	InertStone Stone = iota
	BarrelStone
	EarthStone
	QueenStone
	NightStone
	TeleportStone
	TreeStone
	SealStone // special
)

func (stn Stone) String() (text string) {
	switch stn {
	case InertStone:
		text = "inert stone"
	case BarrelStone:
		text = "barrel stone"
	case EarthStone:
		text = "earth stone"
	case QueenStone:
		text = "queenstone"
	case NightStone:
		text = "night stone"
	case TeleportStone:
		text = "teleport stone"
	case TreeStone:
		text = "tree stone"
	case SealStone:
		text = "seal stone"
	}
	return text
}

func (stn Stone) Desc(g *Game) (text string) {
	switch stn {
	case InertStone:
		text = "This magical stone has been depleted of magical energies but still emits a faint glow."
	case BarrelStone:
		text = "Activating this magical stone will teleport you away to a random barrel in the same level. It reveals any other barrels on the map."
	case EarthStone:
		text = "Activating this magical stone will produce an oric explosion that destroys nearby walls. It reveals interior walls."
	case QueenStone:
		text = "Activating this magical stone made of queen rock will produce an amplified harmonic sound confusing enemies in a wide area. This can also attract monsters. It reveals the current position of monsters, including the exact kind for nearby ones."
	case NightStone:
		text = "Activating this magical stone will produce hypnotic harmonic illusions inducing sleep in all the monsters in sight. It reveals any cloaks, amulets and lights on the map."
	case TeleportStone:
		text = "Activating this magical stone will teleport away monsters in sight. It reveals any magaras on the map."
	case TreeStone:
		text = "Activating this magical stone will lignify monsters in a wide area. It reveals any trees and bananas on the map as well as some foliage."
	case SealStone:
		text = "Activating this magical stone will disable a magical barrier somewhere in the same level, usually one blocking stairs. It reveals the position of the target barrier."
	}
	return text
}

func (stn Stone) ShortDesc(g *Game) string {
	return Indefinite(stn.String(), false)
}

func (stn Stone) Style(g *Game) (r rune, fg gruid.Color) {
	r = '∩'
	switch stn {
	case InertStone:
		fg = ColorForegroundEmph
	case SealStone:
		fg = ColorBlue
	case EarthStone:
		fg = ColorRed
	case QueenStone:
		fg = ColorMagenta
	case NightStone:
		fg = ColorViolet
	case BarrelStone:
		fg = ColorYellow
	case TeleportStone:
		fg = ColorCyan
	case TreeStone:
		fg = ColorGreen
	}
	return r, fg
}

// UseStone actives a stone at the given position.
func (g *Game) UseStone(p gruid.Point) {
	g.StoryPrintf("Activated %s", g.Objects.Stones[p])
	g.Objects.Stones[p] = InertStone
	g.Stats.UsedStones++
	if g.Stats.UsedStones == 10 {
		AchStoneEnthousiast.Get(g)
	}
	g.Log("The stone becomes inert.")
}

func (g *Game) ActivateStone() (err error) {
	stn, ok := g.Objects.Stones[g.Player.P]
	if !ok {
		return errors.New("No stone to activate here.")
	}
	op := g.Player.P
	switch stn {
	case InertStone:
		err = errors.New("The stone is inert.")
	case BarrelStone:
		g.ActivateBarrelStone()
	case EarthStone:
		g.ActivateEarthStone()
	case QueenStone:
		g.ActivateQueenStone()
	case NightStone:
		g.ActivateNightStone()
	case TeleportStone:
		g.ActivateTeleportStone()
	case TreeStone:
		g.ActivateTreeStone()
	case SealStone:
		g.ActivateSealStone()
	}
	if err != nil {
		return err
	}
	g.UseStone(op)
	return nil
}

const (
	FogStoneDistance   = 4
	QueenStoneDistance = 15
	MappingDistance    = 32
)

func (g *Game) ActivateEarthStone() {
	g.Log("You activate the earth stone.")
	pp := g.Player.P
	g.OricExplosion(pp)
	mp := &EarthStonePath{m: g.Map}
	draw, cost := false, 1
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, unreachable)
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw, cost = false, n.Cost
		}
		t := g.Map.At(n.P)
		if t != WallCell {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != WallCell && g.Map.hasInteriorNeighbor(n.P) {
			g.Map.KnownTerrain.Set(n.P, WallCell)
			draw = true
		}
	}
}

// Directions contains the four cardinal directions in direct order (east,
// north, west, south).
var Directions = []gruid.Point{
	{1, 0},
	{0, -1},
	{-1, 0},
	{0, 1},
}

// hasInteriorNeighbor reports whether there's a earth-menhir revelable wall at the
// given position (wall adjacent to an interior wall).
func (m *Map) hasInteriorNeighbor(p gruid.Point) bool {
	for _, dir := range Directions {
		if m.interiorWallAt(p.Add(dir)) {
			return true
		}
	}
	return false
}

// interiorWallAt reports whether there's an interior wall at the given
// position (plain wall not connected to any non-wall tile).
func (m *Map) interiorWallAt(p gruid.Point) bool {
	if m.At(p) != WallCell {
		return false
	}
	for q := range Neighbors(p) {
		if m.At(q) != WallCell {
			return false
		}
	}
	return true
}

func (g *Game) ActivateQueenStone() {
	g.MakeNoise(QueenStoneNoise, g.Player.P)
	mp := &NoisePath{g: g}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.Player.P}, QueenStoneDistance)
	g.Log("The stone releases confusing sounds.")
	g.LogStyled("RING!", logNotable)
	g.md.FOVWavesAnimation(QueenStoneDistance, WaveConfusion, g.Player.P)
	for _, m := range g.Monsters {
		if !m.Exists() {
			continue
		}
		if m.State == Resting {
			continue
		}
		m.UpdateKnowledge(g, m.P)
		c := g.PR.BreadthFirstMapAt(m.P)
		if c > QueenStoneDistance {
			continue
		}
		// Confuse and reveal nearby monster.
		m.PutConfusion(g)
		if m.Search == invalidPos {
			m.Search = m.P
		}
		if !g.Player.Sees(m.P) {
			g.SeeMonster(m, "sense")
		}
	}
}

func (g *Game) ActivateNightStone() {
	targets := []*Monster{}
	for _, mons := range g.Monsters {
		if !mons.Exists() || !g.Player.Sees(mons.P) {
			continue
		}
		if mons.State != Resting {
			targets = append(targets, mons)
		}
	}
	g.Log("The stone releases hypnotic harmonies.")
	g.md.FOVWavesAnimation(DefaultFOVRange, WaveSleeping, g.Player.P)
	for _, mons := range targets {
		g.Logf("%s falls asleep.", mons.Kind.Definite(true))
		mons.State = Resting
		mons.Dir = ZP
		mons.ExhaustTime(g, 4+RandInt(2))
	}
	for p := range g.Objects.Lights {
		g.SeeTerrainAt(p)
	}
	for p := range g.Objects.Items {
		g.SeeTerrainAt(p)
	}
}

func (g *Game) ActivateTreeStone() {
	pp := g.Player.P
	mp := &MapPath{passable: func(q gruid.Point) bool {
		return IsJumpPassable(g.Map.At(q))
	}}
	const maxdist = DefaultFOVRange
	g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, maxdist)
	targets := []*Monster{}
	for _, mons := range g.Monsters {
		if !mons.Exists() || g.PR.BreadthFirstMapAt(mons.P) > maxdist || mons.Kind.ResistsLignification() {
			continue
		}
		targets = append(targets, mons)
	}
	g.Log("The stone releases magical spores.")
	g.md.FOVWavesAnimation(DefaultFOVRange, WaveTree, pp)
	for _, mons := range targets {
		mons.EnterLignification(g)
		if mons.Search == invalidPos {
			mons.Search = mons.P
		}
	}
	for p, t := range g.Map.Terrain.All() {
		if t == TreeCell || t == BananaCell || t == FoliageCell && g.rand.IntN(3) == 0 {
			g.SeeTerrainAt(p)
		}
	}
}

func (g *Game) ActivateTeleportStone() {
	targets := g.MonstersInFOV()
	g.Log("The stone releases oric teleport energies.")
	for _, mons := range targets {
		if mons.Search == invalidPos && mons.Kind.CanOpenDoors() {
			mons.Search = mons.P
		}
		mons.TeleportAway(g)
		if mons.Kind.ReflectsTeleport() {
			g.Logf("The %s reflected back some energies.", mons.Kind)
			g.Teleportation()
			break
		}
	}
	for p := range g.Objects.Magaras {
		g.SeeTerrainAt(p)
	}
}

func (g *Game) ActivateBarrelStone() {
	g.Log("You activate the barrel stone.")
	barrels := []gruid.Point{}
	for p := range g.Objects.Barrels {
		barrels = append(barrels, p)
		g.SeeTerrainAt(p)
	}
	p := barrels[RandInt(len(barrels))]
	op := g.Player.P
	g.Log("You teleport away.")
	g.md.TeleportAnimation(op, p, true)
	g.PlacePlayerAt(p)
}

func (g *Game) ActivateSealStone() {
	if g.Map.Depth == MaxDepth {
		g.Objects.Story[g.Places.Artifact] = StoryArtifact
		g.Log("You feel oric energies dissipating.")
		g.SeeTerrainAt(g.Places.Artifact)
		return
	}
	for p, st := range g.Objects.Stairs {
		// Actually there is at most only one such stair.
		if st == SealedStair {
			g.Objects.Stairs[p] = NormalStair
			g.SeeTerrainAt(p)
		}
	}
	g.Log("You feel oric energies dissipating.")
}

// Scroll represents the various kinds of readable scrolls.
type Scroll int

const (
	ScrollStory Scroll = iota
	ScrollExtended
	ScrollDayoriahMessage
	ScrollLore
)

func (sc Scroll) String() (desc string) {
	switch sc {
	case ScrollLore:
		desc = "message"
	default:
		desc = "story message"
	}
	return desc
}

func (sc Scroll) ShortDesc() (desc string) {
	switch sc {
	case ScrollLore:
		desc = "a message"
	default:
		desc = "a story message"
	}
	return desc
}

func (sc Scroll) Text(g *Game) (desc string) {
	switch sc {
	case ScrollStory:
		desc = "Your friend Shaedra got captured by nasty people from the Dayoriah Clan while she was trying to retrieve a powerful magara artifact that was stolen from the great magara-specialist Marevor Helith.\n\nAs a gawalt monkey, you don’t understand much why people complicate so much their lives caring about artifacts and the like, but one thing is clear: you have to rescue your friend, somewhere to be found in this Underground area controlled by the Dayoriah Clan. If what you heard the guards say is true, Shaedra’s imprisoned on the eighth floor.\n\nYou are small and have good night vision, so you hope the infiltration will go smoothly..."
	case ScrollExtended:
		desc = "Now that Shaedra’s back to safety, you can either follow her advice, and get away from here too using the monolith portal, or you can finish the original mission: going deeper to find Marevor’s powerful magara, before the Dayoriah Clan does bad experiments with it. You honestly didn’t understand why it was dangerous, but Shaedra and Marevor had seemed truly concerned.\n\nMarevor said that he’ll be able to create a new portal for you when you activate the artifact upon finding it."
	case ScrollDayoriahMessage:
		desc = `“The thief that infiltrated our turf and tried to retrieve our new acquisition has been captured. However, it is possible that she has an accomplice. Please be careful and stop every suspect.”

[Order of a Dayoriah Clan’s foreman, the paper is rather recent and you can’t help but think the thief is actually Shaedra. So they really did capture her! You have to hurry and save her.]`
	case ScrollLore:
		i, ok := g.Objects.Lore[g.Player.P]
		if !ok {
			// should not happen
			desc = "Some unintelligible notes."
			break
		}
		if i < len(LoreMessages) {
			desc = LoreMessages[i]
		}
	default:
		desc = "a message"
	}
	return desc
}

func (sc Scroll) Desc(g *Game) (desc string) {
	desc = "A message. It can be read."
	return desc
}

func (sc Scroll) Style(g *Game) (r rune, fg gruid.Color) {
	r = '?'
	fg = ColorCyan
	if sc == ScrollLore {
		fg = ColorViolet
	}
	return r, fg
}

// Story represents the various kinds of "story" special positions.
type Story int

const (
	NoStory Story = iota // just a normal ground cell but free
	StoryShaedra
	StoryMarevor
	StoryArtifact
	StoryArtifactSealed
)

func (st Story) Desc(g *Game, p gruid.Point) (desc string) {
	switch st {
	case NoStory:
		desc = g.TerrainDesc(GroundCell, p)
	case StoryShaedra:
		desc = "Shaedra is the friend you came here to rescue, a human-like creature with claws, a ternian. Many other human-like creatures consider them as savages."
	case StoryMarevor:
		desc = "Marevor Helith is an ancient undead nakrus very fond of teleporting people away. He is a well-known expert in the field of magaras - items that many people simply call magical objects. His current research focus is monolith creation. Marevor, a repentant necromancer, is now searching for his old disciple Jaixel in the Underground to help him overcome the past."
	case StoryArtifact:
		desc = "This is the magara that you have to retrieve: the Gem Portal Artifact that was stolen to Marevor Helith."
	case StoryArtifactSealed:
		desc = "This is the magara that you have to retrieve: the Gem Portal Artifact that was stolen to Marevor Helith. Before taking it, you have to release the magical barrier that protects it activating the corresponding protective barrier magical stone."
	}
	return desc
}

func (st Story) String() (desc string) {
	switch st {
	case NoStory:
		desc = "paved ground"
	case StoryShaedra:
		desc = "Shaedra"
	case StoryMarevor:
		desc = "Marevor"
	case StoryArtifact:
		desc = "Gem Portal Artifact"
	case StoryArtifactSealed:
		desc = "Gem Portal Artifact (sealed)"
	}
	return desc
}

func (st Story) Style(g *Game) (r rune, fg gruid.Color) {
	fg = ColorBlue
	switch st {
	case NoStory:
		fg = ColorForeground
		r = '.'
	case StoryShaedra:
		r = 'S'
	case StoryMarevor:
		r = 'M'
	case StoryArtifact:
		r = '='
	case StoryArtifactSealed:
		r = '='
		fg = ColorCyan
	}
	return r, fg
}

type Item struct {
	Kind  ItemKind // item kind
	Depth int      // map depth where you found it
}

// ItemKind represents the various kinds of inventory items.
type ItemKind int

const (
	NoItem ItemKind = iota
	CloakHear
	CloakAcrobat
	CloakShadows
	CloakSmoke
	CloakConversion
	AmuletTeleport
	AmuletConfusion
	AmuletFog
	AmuletDazing
	MarevorMagara
)

// IsCloak reports whether the item is a cloak.
func (it Item) IsCloak() bool {
	switch it.Kind {
	case CloakHear,
		CloakAcrobat,
		CloakSmoke,
		CloakShadows,
		CloakConversion:
		return true
	}
	return false
}

// IsAmulet reports whether the item is an amulet.
func (it Item) IsAmulet() bool {
	switch it.Kind {
	case AmuletTeleport, AmuletConfusion, AmuletFog, AmuletDazing:
		return true
	}
	return false
}

func (it Item) String() (desc string) {
	switch it.Kind {
	case NoItem:
		desc = "empty slot"
	case CloakHear:
		desc = "cloak of hearing"
	case CloakAcrobat:
		desc = "cloak of acrobatics"
	case CloakShadows:
		desc = "cloak of shadows"
	case CloakSmoke:
		desc = "cloak of smoking"
	case CloakConversion:
		desc = "cloak of conversion"
	case AmuletTeleport:
		desc = "amulet of teleport"
	case AmuletConfusion:
		desc = "amulet of confusion"
	case AmuletFog:
		desc = "amulet of fog"
	case AmuletDazing:
		desc = "amulet of daze"
	case MarevorMagara:
		desc = "Moon Portal Artifact"
	}
	return desc
}

func (it Item) ShortDesc(g *Game) string {
	return it.String()
}

func (it Item) Desc(g *Game) (desc string) {
	switch it.Kind {
	case NoItem:
		return "You do not have an item equipped on this slot yet."
	case CloakHear:
		desc = "Improves hearing. Allows to sometimes hear monsters watching and turning around, as well as the breathing of nearby sleeping monsters."
	case CloakAcrobat:
		desc = "Prevents exhaustion from jumping."
	case CloakShadows:
		desc = "Reduces the range at which foes see you in the dark by one. It also improves your maximum health by one."
	case CloakSmoke:
		desc = "Leaves smoke behind as you move, making you difficult to spot."
	case CloakConversion:
		desc = "Converts lost health from wounds into magical energy."
	case AmuletTeleport:
		desc = "Teleports away foes that critically hit you."
	case AmuletConfusion:
		desc = "Confuses foes that critically hit you."
	case AmuletFog:
		desc = "Releases fog and makes you swift when critically hurt."
	case AmuletDazing:
		desc = "Dazes foes that critically hit you."
	case MarevorMagara:
		desc = "It was given to you by Marevor Helith so that he can create an escape portal when you reach Shaedra. Its sister magara, the Gem Portal Artifact, also crafted by Marevor, is the artifact that was stolen and that Shaedra was trying to retrieve before being captured.\n\nThis magara needs a lot of time to recharge, so you’ll only be able to use it once."
	}
	if it.Depth >= 1 {
		desc += fmt.Sprintf("\n\nYou found it at dungeon depth %d.", it.Depth)
	}
	return desc
}

func (it Item) Style(g *Game) (r rune, fg gruid.Color) {
	fg = ColorCyan
	if it.IsAmulet() {
		r = '='
	} else if it.IsCloak() {
		r = '['
	}
	return r, fg
}

// EquipItem equips an item at the current player position.
func (g *Game) EquipItem() error {
	it, ok := g.Objects.Items[g.Player.P]
	if !ok {
		return errors.New("Nothing to equip here.")
	}
	if it.Depth == -1 {
		it.Depth = g.Map.Depth
	}
	var oitem Item
	switch {
	case it.IsCloak():
		oitem = g.Player.Inventory.Body
		g.Player.Inventory.Body = it
	case it.IsAmulet():
		oitem = g.Player.Inventory.Neck
		g.Player.Inventory.Neck = it
	}
	if oitem.Kind != NoItem {
		g.Objects.Items[g.Player.P] = oitem
		g.Logf("You equip the %s.", it.ShortDesc(g))
		g.Logf("You leave the %s.", oitem.ShortDesc(g))
		g.StoryPrintf("Equipped %s", it.ShortDesc(g))
	} else {
		delete(g.Objects.Items, g.Player.P)
		g.Map.Set(g.Player.P, GroundCell)
		g.Logf("You equip the %s.", it.ShortDesc(g))
		g.StoryPrintf("Equipped %s", it.ShortDesc(g))
	}
	switch {
	case it.IsCloak():
		AchCloak.Get(g)
	case it.IsAmulet():
		AchAmulet.Get(g)
	}
	return nil
}

// RandomCloak returns a random cloak (not already generated).
func (g *Game) RandomCloak() Item {
	cloaks := []ItemKind{
		CloakHear,
		CloakAcrobat,
		CloakSmoke,
		CloakShadows,
		CloakConversion}
	var it ItemKind
loop:
	for {
		it = cloaks[RandInt(len(cloaks))]
		for _, cl := range g.ProcInfo.GeneratedCloaks {
			if cl == it {
				continue loop
			}
		}
		break
	}
	return Item{Kind: it, Depth: -1}
}

func (g *Game) RandomAmulet() Item {
	amulets := []ItemKind{AmuletTeleport,
		AmuletConfusion,
		AmuletFog,
		AmuletDazing}
	var it ItemKind
loop:
	for {
		it = amulets[RandInt(len(amulets))]
		for _, cl := range g.ProcInfo.GeneratedAmulets {
			if cl == it {
				continue loop
			}
		}
		break
	}
	return Item{Kind: it, Depth: -1}
}

// Potion represents a potion.
type Potion int

const (
	HealthPotion Potion = iota
	MagicPotion
)

func (ptn Potion) String() (desc string) {
	switch ptn {
	case HealthPotion:
		desc = "health potion"
	case MagicPotion:
		desc = "magic potion"
	}
	return desc
}

func (ptn Potion) ShortDesc(g *Game) (desc string) {
	return Indefinite(ptn.String(), false)
}

func (ptn Potion) Desc(g *Game) (desc string) {
	switch ptn {
	case HealthPotion:
		desc = "Drinking a health potion will cure 1 HP."
	case MagicPotion:
		desc = "Drinking a magic potion will replenish 1 MP."
	}
	desc += "\n\nYou cannot carry around potions, you drink them when you move onto them."
	return desc
}

func (ptn Potion) Style(g *Game) (r rune, fg gruid.Color) {
	r = '!'
	switch ptn {
	case HealthPotion:
		fg = ColorGreen
	case MagicPotion:
		fg = ColorBlue
	}
	return r, fg
}

// DrinkPotion performs drinking of a potion at the given position.
func (g *Game) DrinkPotion(p gruid.Point) {
	ptn, ok := g.Objects.Potions[p]
	if !ok {
		// should not happen
		g.Map.Set(p, GroundCell)
		g.LogStyled("Unexpected potion.", logError)
		return
	}
	switch ptn {
	case HealthPotion:
		if g.Player.HP >= g.Player.HPMax() {
			return
		}
		g.Player.HP++
		g.StoryPrintf("Drank %s (HP: %d)", ptn, g.Player.HP)
	case MagicPotion:
		if g.Player.MP >= g.Player.MPMax() {
			return
		}
		g.Player.MP++
		g.StoryPrintf("Drank %s (MP: %d)", ptn, g.Player.MP)
	}
	g.Logf("You drink %s.", ptn.ShortDesc(g))
	g.Stats.DrankPotions++
	g.Map.Set(p, GroundCell)
	delete(g.Objects.Potions, p)
}
