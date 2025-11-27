package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModCorruptedDungeon Mod = iota // mod with lots of unusual and random stuff going on
	ModAdvancedSpirits             // dedicated slot for challenge secondary spirits
	ModGluttonyRework              // mod with choice-of-two gluttony rework

	ModSmallInventory // mod limiting inventory to 3 comestible slots
	ModNoRecharges    // mod without spirit recharges
	ModHealingCombat  // combat-based healing system instead of comestible-based

	modCount // dummy mod (for NMods)
)

// Number of mods.
const NMods = int(modCount)

var gameMods = []Mod{
	// expansion mods
	ModCorruptedDungeon, ModAdvancedSpirits,
	ModGluttonyRework,
	// challenge mods
	ModSmallInventory,
	ModNoRecharges, ModHealingCombat,
}

// Number of expansion mods (for header positioning purposes in menu).
const NExpansions = 3

func (m Mod) String() string {
	switch m {
	case ModCorruptedDungeon:
		return "Corrupted Dungeon"
	case ModAdvancedSpirits:
		return "Advanced Spirits"
	case ModGluttonyRework:
		return "Gluttony Rework"
	case ModSmallInventory:
		return "Small Inventory"
	case ModNoRecharges:
		return "No Recharges"
	case ModHealingCombat:
		return "Healing Combat"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModCorruptedDungeon:
		return "The orb’s influence has scattered twisted surprises through the dungeon. Some might even be good!\n\nRecommended for experienced players that want more unpredictability and variability."
	case ModAdvancedSpirits:
		return "Enable extra advanced secondary spirits: Dazzling Zebra, Gardening Lion, Gawalt Monkey, Gluttonous Bear, Runic Chicken, Staring Owl, Stomping Elephant.\n\nAdvanced secondaries always have strong points but also serious and quirky drawbacks.\n\nBeware that regular spirits are rarer early on than advanced ones!"
	case ModGluttonyRework:
		return "Makes the Gluttonous Bear eat in pairs through a choice-of-two menu, without relying on a Gluttony status.\n\nThe Advanced Spirits mod should be enabled too for this mod to take effect."
	case ModSmallInventory:
		return "Inventory can only hold 3 comestibles. Choose them with care!"
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don’t recharge when going to the next map level. Use them wisely!"
	case ModHealingCombat:
		return "Healing happens through combat. When a monster dies, you may heal for 1 HP, with higher chance at low HP. However, comestibles don’t provide healing anymore.\n\nVampiric Bat players start with an extra “vampirism” charge and get another one at map level 5, instead of healing on monster death. Less frequent healing but better tactical control over it!\n\n@MNote.@N This experimental mod is playable but lacks polish with respect to the default experience."
	default:
		return "(unknown mod)"
	}
}

// HasVampirism reports whether the player has a vampirism ability.
func (g *Game) HasVampirism() bool {
	sp := g.Entity(PrimarySpiritID).Role.(*Spirit)
	_, ok := sp.Ability[sp.Level].(EffectVampirism)
	return ok
}
