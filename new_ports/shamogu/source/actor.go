package main

import (
	"fmt"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Actor holds data relevant to any kind of actor (player or monster).
type Actor struct {
	HP        int         // health points
	MaxHP     int         // maximum health points
	Defense   int         // defense (defensive potential)
	Attack    int         // attack (offensive potential)
	FirePos   gruid.Point // position at start of turn while on Fire
	KnownDead bool        // whether it's a monster known to be dead
	Statuses  Statuses    // statuses (confused, etc.)
	Kind      ActorKind   // actor kind (player or some kind of monster)
	Traits    Traits      // passives traits
	Behavior  *Behavior   // behavior component (monsters only)
}

// NewActor returns a new actor with the given combat stats and traits.
func NewActor(attack, defense, hp int, ak ActorKind, t Traits) *Actor {
	a := &Actor{
		Attack:   attack,
		Defense:  defense,
		HP:       hp,
		MaxHP:    hp,
		Kind:     ak,
		Statuses: make(Statuses, NStatuses),
		Traits:   t,
	}
	return a
}

// DoesAny reports whether the player has all traits in t.
func (a *Actor) DoesAny(t Traits) bool {
	return a.Traits.Any(t)
}

// DoesAll reports whether the player has all traits in t.
func (a *Actor) DoesAll(t Traits) bool {
	return a.Traits.All(t)
}

// Has reports whether the player has status st.
func (a *Actor) Has(st Status) bool {
	return a.Statuses.Has(st)
}

// Is reports whether the actor is of the given kind.
func (a *Actor) Is(ak ActorKind) bool {
	return a.Kind == ak
}

// CanMove reports whether the actor isn't under any status that prevents
// movement.
func (a *Actor) CanMove() bool {
	return !a.Has(StatusLignification) || a.DoesAny(WoodyLegs)
}

// ResistsMove reports whether the actor resists against forced movement. This
// includes lignified actors that can still move due to woodylegs as well as
// walking trees.
func (a *Actor) ResistsMove() bool {
	return a.Has(StatusLignification) || a.Is(WalkingTree)
}

// IsAlive reports whether the actor is still alive.
func (a *Actor) IsAlive() bool {
	return a.HP > 0
}

// IsDead reports whether the actor is dead.
func (a *Actor) IsDead() bool {
	return a.HP <= 0
}

// AdjustHP heals an actor by a given amount (or hurts if negative). The result
// is guaranteed to be in [1,a.MaxHP].
func (g *Game) AdjustHP(i ID, ai *Actor, amount int) {
	g.AdjustHPWithMin(i, ai, amount, 1)
}

// AdjustHPWithMin heals an actor by a given amount (or hurts if negative). The
// result is guaranteed to be in [min(m,ai.HP,maxHP), maxHP] where maxHP is
// ai.GetMaxHP().
func (g *Game) AdjustHPWithMin(i ID, ai *Actor, amount int, m int) {
	if ai.IsDead() {
		// Should not happen: dead monsters remain dead.
		return
	}
	hp, maxHP := ai.HP, ai.GetMaxHP()
	ai.HP = min(maxHP, max(min(m, hp), hp+amount))
	if i == PlayerID {
		if ai.HP > hp {
			g.Logf("You feel better (+%d HP).", ai.HP-hp)
		} else if ai.HP < hp {
			g.Logf("You feel weaker (-%d HP).", hp-ai.HP)
		} else if amount >= 0 {
			g.Log("You feel no change (+0 HP).")
		}
	}
}

// UpdateFirePos updates the FirePos field of the given actor.
func (g *Game) UpdateFirePos(i ID, ai *Actor) {
	if ai.Has(StatusFire) {
		ai.FirePos = g.Entity(i).P
	} else {
		ai.FirePos = InvalidPos
	}
}

// Traits represents special caracteristics of an actor as a bitset.
type Traits uint64

// Any reports whether any of the traits is in the set.
func (t Traits) Any(of Traits) bool {
	return t&of != 0
}

// All reports whether all of the traits are in the set.
func (t Traits) All(of Traits) bool {
	return t&of == of
}

// Extra Traits. NOTE: there shouldn't be more than 64 (remaining free: 12).
//
// Mostly to be used for player traits and traits that may be shared by more
// than a single actor kind (but currently various traits are specific to
// single monsters).
const (
	PatternCatch = 1 << iota // unbalancing and catching attack (frog, octopode)

	PatternDragging     // dragging attack pattern (crocodile players, alligator)
	PatternFourDirs     // four-directional attack in 4 dirs (hydra, undead knight)
	PatternRampage      // rampaging (boar, phoenix)
	PatternRanged       // ranged attack (eye, llama, lich)
	PatternRangedRecoil // ranged attack with recoil (wind fox)
	PatternSneaky       // sneaky attack pattern (vampiric bat, chaos megabat)
	PatternSwap         // swapping-attack (crazy druid)
	PatternSwapDaze     // dazing swapping-attack (temporal cat)

	BadHearing      // worse hearing of noise (footsteps)
	BurningHits     // hits may burn
	DazingSpines    // body covered in dazing spines (porcupine)
	Dazzling        // zebra's camouflage
	Elephanty       // rat phobia, difficult rotation when facing a wall, immune to berserk after-effects
	Gawalt          // wall-jumping, weakened hits, menhir sensing and hiding
	Gluttony        // gluttony status after eating
	GoodHearing     // better hearing of noise (footsteps)
	GoodSmell       // smell food from a distance
	NocturnalFlying // flying a bit above the ground like an owl
	PushingCharge   // unbalancing pushing when charging (boar)
	RunicChicken    // doesn't trigger traps, cackles loudly on sighting a portal or totem
	ScaryRoar       // intimidating roar on first sight
	TrailingCloud   // leave dust behind after moving (on floor tiles)
	VenomousMelee   // (non-ranged) attacks may poison foes
	WoodyLegs       // can move while lignified

	VulnerabilityFire // increases duration of fire status

	ResistanceConfusion // halves duration of confusion status
	ResistanceDaze      // allows spirit invocation while dazed
	ResistanceFear      // halves duration of fear status
	ResistanceFire      // delays burning
	ResistanceImbalance // halves duration of imbalance status
	ResistancePoison    // halves duration of poison status

	ImmunityConfusion     // immune to confusion
	ImmunityDaze          // immune to daze status effect
	ImmunityFear          // immune to fear
	ImmunityFire          // immune to fire
	ImmunityImbalance     // immune to imbalance
	ImmunityLignification // immune to lignification
	ImmunityPoison        // immune to poison

	// Monster-only traits shared by various monsters.

	MonsBerserking     // hits may make you berserk (spiders, crazy druid)
	MonsConfusion      // confusing attack (floating eyes, walking mushroom)
	MonsDig            // dig walls when hunting (boars, dragons)
	MonsExplodingDeath // explode on death (nadres, golems)
	MonsFear           // attack may frighten (lich, knight)
	MonsIgnoreDefense  // attack ignores defense (acid mound, lich)
	MonsScales         // scales provide extra protection from ranged attacks (viper, hydra, nadre, dragon, alligator)

	// NOTE: we don't do descriptions for the traits below (beyond the log
	// messages about noise), so they could be moved to another struct
	// field to make place for more traits if we ever need it.

	MonsCreep          // creep noise (instead of footsteps)
	MonsHeavyFootsteps // heavy footsteps
	MonsLightFootsteps // light footsteps
	MonsSilent         // silent movement (spiders, wraith)
	MonsWingFlap       // wing flapping noise (instead of footsteps)

	MonsNotable // notable monster (mention in timeline)

	NoTraits = 0
)

// TraitDesc returns a suitable trait description for the given actor kind and
// extra traits. For monsters, it includes the description of some
// monster-specific traits.
func TraitDesc(ak ActorKind, t Traits) string {
	desc := []string{}

	// Attack patterns.
	desc = traitDesc(desc, t, PatternCatch, "ranged catching attack that @Ounbalances@N foes")
	desc = traitDesc(desc, t, PatternDragging, "@Ounbalancing@N dragging attack")
	desc = traitDesc(desc, t, PatternFourDirs, "four-directional attack")
	desc = traitDesc(desc, t, PatternRampage, "rampaging charge")
	desc = traitDesc(desc, t, PatternRanged, "ranged attack")
	desc = traitDesc(desc, t, PatternRangedRecoil, "ranged attack with recoil")
	desc = traitDesc(desc, t, PatternSneaky, "sneaky attack pattern")
	desc = traitDesc(desc, t, PatternSwap, "ranged space-distorting silent attack")
	desc = traitDesc(desc, t, PatternSwapDaze, "ranged space-distorting silent attack that may @Odaze@N foes")
	switch ak {
	case Player:
		// Player crocodile gets slow rotation, too.
		switch {
		case t.Any(PatternDragging):
			desc = append(desc, "slow rotation (cannot turn and attack)")
		case t.Any(PatternFourDirs):
			desc = append(desc, "attack bonus per extra foe")
		case t.Any(PatternSneaky):
			desc = append(desc, "confusing charge")
		}
	case DraggingAlligator:
		desc = append(desc, "confusing bite")
	case ChaosMegabat:
		desc = append(desc, "chaos bite (random effect)")
	}

	// Extra traits.
	desc = traitDesc(desc, t, BurningHits, "hits may @Oburn@N foes")
	desc = traitDesc(desc, t, DazingSpines, "spines may @Odaze@N attacking foes in melee")
	desc = traitDesc(desc, t, Dazzling, "dazzling (redirect attacks just behind you, easily noticed)")
	desc = traitDesc(desc, t, Elephanty, "rat phobia, difficult rotation when facing a wall, unnoticed in dead-ends, immune to berserk after-effects")
	desc = traitDesc(desc, t, Gawalt, "wall-jumping movement and attack, weakened hits, menhir sensing and hiding")
	desc = traitDesc(desc, t, Gluttony, "gluttony (needs to eat in pairs)")
	desc = traitDesc(desc, t, GoodSmell, "smells food from afar")
	desc = traitDesc(desc, t, NocturnalFlying, "nocturnal (reduced view range), flies over foliage")
	desc = traitDesc(desc, t, PushingCharge, "charging pushes and @Ounbalances@N foes")
	desc = traitDesc(desc, t, RunicChicken, "senses and doesn’t trigger runes, cackles loudly on sighting a portal or totem, confusion vulnerability (longer duration)")
	desc = traitDesc(desc, t, ScaryRoar, "scary roar on first sight")
	desc = traitDesc(desc, t, TrailingCloud, "movement leaves dust behind")
	desc = traitDesc(desc, t, VenomousMelee, "@Ovenomous@N melee attack")
	desc = traitDesc(desc, t, WoodyLegs, "woody legs (can move while lignified)")
	switch {
	case t.Any(GoodHearing) && !t.Any(BadHearing):
		desc = append(desc, "good hearing")
	case !t.Any(GoodHearing) && t.Any(BadHearing):
		desc = append(desc, "bad hearing")
	}

	// Extra monster-specific traits.
	switch ak {
	case BarkingHound:
		desc = append(desc, "barks on sight and @Ofrightens@N you")
	case BlinkButterfly:
		desc = append(desc, "hits makes you blink")
	case BurningPhoenix:
		desc = append(desc, "imbalance vulnerability (decreased burning chance)")
	case ChaosMegabat:
		// Megabat hunts by sound (when out-of-view for noises
		// originating on the player's tile).
		desc = append(desc, "hunts by sound")
	case EarthDragon:
		desc = append(desc, "hits may push and @Ounbalance@N foes")
	case FireLlama:
		desc = append(desc, "spites @Ofire@N from afar")
	case FourHeadedHydra:
		desc = append(desc, "attacks with all four heads")
	case HungryRat:
		desc = append(desc, "hunts you by smell, becomes berserk if you eat in view")
	case NoisyImp:
		desc = append(desc, "plays guitar for you, avoids fights")
	case WalkingMushroom:
		desc = append(desc, "releases @Blignifying@N spores on sight")
	case WalkingTree:
		desc = append(desc, "hits may @Blignify@N you, hits ignore lignification’s defense bonus, woody legs, permanently lignified")
	case WarpingWraith:
		desc = append(desc, "hits teleport you away")
	}
	desc = traitDesc(desc, t, MonsBerserking, "hits may @Bberserk@N you briefly")
	desc = traitDesc(desc, t, MonsConfusion, "hits may @Oconfuse@N you")
	desc = traitDesc(desc, t, MonsDig, "digs walls when hunting")
	desc = traitDesc(desc, t, MonsExplodingDeath, "explodes on death")
	desc = traitDesc(desc, t, MonsFear, "hits may @Ofrighten@N you")
	desc = traitDesc(desc, t, MonsIgnoreDefense, "attack ignores defense")
	desc = traitDesc(desc, t, MonsScales, "scales (+1 Defense against ranged attacks)")

	// Vulnerabilities, Resistances, Immunities.
	desc = traitDesc(desc, t, VulnerabilityFire, "@OFire@N vulnerability (longer duration)")
	desc = traitDesc(desc, t, ResistanceConfusion, "@OConfusion@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistanceDaze, "@ODaze@N resistance (can invoke spirits while dazed)")
	desc = traitDesc(desc, t, ResistanceFear, "@OFear@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistanceFire, "@OFire@N resistance (delayed burning)")
	desc = traitDesc(desc, t, ResistanceImbalance, "@OImbalance@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistancePoison, "@OPoison@N resistance (half duration)")
	desc = traitDesc(desc, t, ImmunityConfusion, "immune to @OConfusion@N")
	desc = traitDesc(desc, t, ImmunityDaze, "@ODaze@N immunity")
	desc = traitDesc(desc, t, ImmunityFear, "@OFear@N immunity")
	desc = traitDesc(desc, t, ImmunityFire, "@OFire@N immunity")
	desc = traitDesc(desc, t, ImmunityImbalance, "@OImbalance@N immunity")
	desc = traitDesc(desc, t, ImmunityLignification, "@BLignification@N immunity")
	desc = traitDesc(desc, t, ImmunityPoison, "@OPoison@N immunity")

	return strings.Join(desc, ", ")
}

func traitDesc(desc []string, t1, t2 Traits, s string) []string {
	if t1.All(t2) {
		return append(desc, s)
	}
	return desc
}

// Status describes different kind of statuses.
type Status int

const (
	StatusBerserk Status = iota
	StatusClarity
	StatusConfusion
	StatusDaze
	StatusDig
	StatusDisorient
	StatusFear
	StatusFire
	StatusFocus
	StatusFoggySkin
	StatusGardener
	StatusGluttony
	StatusImbalance
	StatusLignification
	StatusPoison
	StatusShadow
	StatusSprint
	StatusTimeStop
	StatusVampirism
)

// Bad reports whether the status is inherently bad.
func (st Status) Bad() bool {
	switch st {
	case StatusConfusion, StatusDaze, StatusFear, StatusFire,
		StatusImbalance, StatusPoison:
		return true
	default:
		return false
	}
}

// NStatuses is the number of different kind of statuses.
const NStatuses = StatusVampirism + 1

var statusAbbrevation = []string{
	StatusBerserk:       "Ber",
	StatusClarity:       "Cla",
	StatusConfusion:     "Con",
	StatusDaze:          "Daz",
	StatusDig:           "Dig",
	StatusDisorient:     "Dis",
	StatusFear:          "Fea",
	StatusFire:          "Fir",
	StatusFocus:         "Foc",
	StatusGardener:      "Gar",
	StatusGluttony:      "Glu",
	StatusImbalance:     "Imb",
	StatusFoggySkin:     "Fog",
	StatusLignification: "Lig",
	StatusPoison:        "Poi",
	StatusShadow:        "Sha",
	StatusSprint:        "Spr",
	StatusTimeStop:      "Tim",
	StatusVampirism:     "Vam",
}

func (st Status) Abbr() string {
	return statusAbbrevation[st]
}

// StatusAbbr is like Abbr but accounts for actor-specific resistances.
func (g *Game) StatusAbbr(ai *Actor, st Status) string {
	switch {
	case st == StatusLignification && ai.DoesAny(WoodyLegs) ||
		st == StatusDaze && ai.DoesAny(ResistanceDaze):
		return statusAbbrevation[st] + "*"
	default:
		return statusAbbrevation[st]
	}
}

var statusName = []string{
	StatusBerserk:       "Berserk",
	StatusClarity:       "Clarity",
	StatusConfusion:     "Confusion",
	StatusDaze:          "Daze",
	StatusDig:           "Dig",
	StatusDisorient:     "Disorient",
	StatusFear:          "Fear",
	StatusFire:          "Fire",
	StatusFocus:         "Focus",
	StatusFoggySkin:     "Foggy Skin",
	StatusGardener:      "Gardener",
	StatusGluttony:      "Gluttony",
	StatusImbalance:     "Imbalance",
	StatusLignification: "Lignification",
	StatusPoison:        "Poison",
	StatusShadow:        "Shadow",
	StatusSprint:        "Sprint",
	StatusTimeStop:      "Time Stop",
	StatusVampirism:     "Vampirism",
}

func (st Status) Name() string {
	return statusName[st]
}

var statusAdjective = []string{
	StatusBerserk:       "berserk",
	StatusClarity:       "clear-minded",
	StatusConfusion:     "confused",
	StatusDaze:          "dazed",
	StatusDig:           "digging",
	StatusDisorient:     "disorienting",
	StatusFear:          "afraid",
	StatusFire:          "on fire",
	StatusFocus:         "focusing",
	StatusFoggySkin:     "foggy",
	StatusGardener:      "gardening",
	StatusGluttony:      "gluttonous",
	StatusImbalance:     "imbalanced",
	StatusLignification: "lignified",
	StatusPoison:        "poisoned",
	StatusShadow:        "shadowy",
	StatusSprint:        "sprinting",
	StatusTimeStop:      "freezing time",
	StatusVampirism:     "vampiric",
}

func (st Status) String() string {
	return statusAdjective[st]
}

var statusDesc = []string{
	StatusBerserk: fmt.Sprintf(
		"Gives %+d Attack and +1 effective damage. Provides temporary HP bonus. Protects from @OFear@N. Followed by @OPoison@N on expiration for %d turns.",
		BerserkAttackBonus, DurationPoisonBerserk),
	StatusClarity:   "Protects from @OConfusion@N, @ODaze@N, and @BBerserk@N. Makes you sense nearby monsters (2× view range).",
	StatusConfusion: "Using spirit abilities will hurt you for 1 damage. Doubles duration of new @OImbalance@N, @ODaze@N and @OFear@N effects.",
	StatusDaze:      "Forbids all actions except waiting or eating until you’re hurt or it expires.",
	StatusDig:       "Allows walking into walls. Gives +1 effective damage when charging and guarantees pushing hits (even in melee).",
	StatusDisorient: "Makes any hunting monsters in view move in the opposite direction with respect to your current direction. Monsters may mistakenly attack any monster on the way while doing so. Disoriented monsters do not perform four-directional charges.",
	StatusFear:      "Forbids attacking and walking toward foes.",
	StatusFire:      fmt.Sprintf("Remaining in place or moving into a cloud of fire will burn you for %d damage. When first catching @OFire@N, you also get burned once immediately.", FireDamage),
	StatusFocus:     "Attack frontally with all four heads.",
	StatusFoggySkin: "Exudes fog on the sides and behind. Protects from @OFire@N and @BLignification@N.",
	StatusGardener:  "Grows foliage in a 2-tile radius each turn.",
	StatusGluttony:  fmt.Sprintf("Unless you eat before expiration, you’ll eat a random comestible without taking a turn. If there is no food, you’ll become confused for %d turns.", DurationConfusionGluttony),
	StatusImbalance: fmt.Sprintf("Halves attack. When also confused, gives %+d Defense.", ConfusedImbalanceDefenseBonus),
	StatusLignification: fmt.Sprintf(
		"Gives %+d Defense. Caps incoming attack damage at 1. Provides temporary HP bonus. Protects from @OPoison@N and @OImbalance@N. Delays burning. Prevents movement. Followed by @OImbalance@N for you and any adjacent actors on expiration for %d turns.",
		LignificationDefenseBonus, DurationImbalanceLignification),
	StatusPoison: fmt.Sprintf("Walking will hurt you for %d damage. Stinky toxins are exuded on expiration, confusing other creatures for %d turns.",
		PoisonDamage, DurationConfusionToxins),
	StatusShadow:   "Harmonic shadows make non-hunting monsters not notice you unless you attack them. You can move instantly through translucent walls. It makes combat silent and hits make hunting monsters lose track of you.",
	StatusSprint:   "You can move to a visible destination three times as fast on the current direction and twice as fast laterally. You can jump over foes, unbalancing them. Canceled by waiting.",
	StatusTimeStop: "Time is frozen for everyone but you until expiration. Status effects affecting you will still progress.",
	StatusVampirism: fmt.Sprintf("Guarantees successful melee bite that restores HP for the amount of damage you deal. Gives %+d Attack.",
		VampirismAttackBonus),
}

// Desc returns a description of the status, without extra details, as shown in
// ability and comestible descriptions.
func (st Status) Desc() string {
	return statusDesc[st]
}

// Details returns extra description details about the status that appear in
// help and status mouse hover.
func (st Status) Details() string {
	var details string
	switch st {
	case StatusBerserk:
		details = "HP never falls below 1 after losing bonus HP. Duration and HP bonus may be renewed by eating a berserking flower."
	case StatusClarity:
		details = "Player-only effect obtainable by eating clarity leaves. May be removed by eating a berserking flower for half the @BBerserk@N duration."
	case StatusDaze:
		details = fmt.Sprintf("Players with “daze resistance” can still invoke spirits; status shows as @O%s*@N for them.", StatusDaze.Abbr())
	case StatusConfusion:
		details = "Confused monsters will attack other monsters too when adjacent. Some monsters may even hurt or affect themselves when performing special actions like digging, spitting fire, barking, or pushing."
	case StatusDisorient:
		details = "Player-only effect obtainable with the Dazzling Zebra spirit."
	case StatusFear:
		details = "Monsters will flee when facing you. Ambushing charges against unseen foes can still happen. At the end of a turn, a cornered and afraid actor with no possible escape will become berserk. When lignified and afraid, getting hit makes you berserk, too."
	case StatusFire:
		details = "Players with “fire resistance” from the upgraded Fire Salamander don’t take that extra immediate damage."
	case StatusFocus:
		details = "Cancels @BSprint@N. Player-only effect obtainable with the Four-Headed Hydra spirit."
	case StatusFoggySkin:
		details = "Player-only effect obtainable by eating a foggy-skin onion. May be removed by eating a lignification fruit for half the @BLignification@N duration."
	case StatusGardener:
		details = "Player-only effect obtainable with the Gardening Lion spirit."
	case StatusGluttony:
		details = "Player-only effect obtainable with the Gluttonous Bear spirit."
	case StatusImbalance:
		details = "Sprinting or jumping while imbalanced may sometimes lead to a @Odazing@N fall."
	case StatusLignification:
		details = fmt.Sprintf("The player’s HP never falls below 3 after losing bonus HP. Monsters’ HP floor is only 1, like for @BBerserk@N. Duration and HP bonus may be renewed by eating a lignification fruit. Cancels @BSprint@N. Actors with “woody legs” aren’t rooted, so they lack delayed burning protection but can move while lignified at the cost of reducing the duration by an extra turn; status shows as @B%s*@N for them, and there is no @OImbalance@N on expiration, because no root withdrawing happens.", StatusLignification.Abbr())
	case StatusPoison:
		details = "Watch out for confusion if any monsters survive poisoning!"
	case StatusShadow:
		details = "Note that hitting a wandering monster will make it hunt you as usual. Player-only effect obtainable with the Gawalt Monkey spirit."
	case StatusSprint:
		details = "Disables normal attack. Player-only effect obtainable with the Sprinting Gazelle spirit."
	case StatusTimeStop:
		details = "Player-only effect obtainable with the Temporal Cat spirit."
	case StatusVampirism:
		details = "Player-only effect obtainable with the Vampiric Bat spirit."
	}
	return details
}

// Statuses maps ongoing statuses to their remaining turns. Slices of that type should be created with NStatuses elements.
type Statuses []int

// Put puts on a particular status for a given number of turns (if greater than current one).
func (sts Statuses) Put(st Status, turns int) bool {
	if turns <= sts[st] {
		return false
	}
	sts[st] = turns
	return true
}

// Has reports whether the given status is ongoing.
func (sts Statuses) Has(st Status) bool {
	return sts[st] > 0
}

// PutStatus1 is PutStatusN(..., 1).  It increases final duration of turns by
// one: meant for statuses that are acquired during the actor's current turn
// and don't have any effect until after the end of the turn.
func (g *Game) PutStatus1(i ID, ai *Actor, st Status, turns int) bool {
	return g.PutStatusN(i, ai, st, turns, 1)
}

// PutStatus is PutStatusN(..., false).
func (g *Game) PutStatus(i ID, ai *Actor, st Status, turns int) bool {
	return g.PutStatusN(i, ai, st, turns, 0)
}

// PutStatusN adds a status on an actor for a given number of turns. It
// adjusts final duration of turns by n.
func (g *Game) PutStatusN(i ID, ai *Actor, st Status, turns int, n int) bool {
	if ai.IsDead() {
		return false
	}
	active := ai.Has(st) // whether the status is already active
	switch st {
	case StatusBerserk:
		if active || ai.Has(StatusClarity) {
			return false
		}
	case StatusFire:
		if ai.Has(StatusFoggySkin) || ai.DoesAny(ImmunityFire) {
			return false
		}
		if ai.DoesAny(VulnerabilityFire) {
			turns += 2
		}
		if active || ai.DoesAny(ResistanceFire) || ai.Has(StatusLignification) && !ai.DoesAny(WoodyLegs) {
			// If Fire is already active or we have fire resistance
			// or Lignification without woody legs, we skip extra
			// damage at onset.
			break
		}
		g.InflictDamage(i, ai, FireDamage, AttackFireCatch)
		if ai.IsDead() {
			return false
		}
	case StatusFear:
		if ai.Has(StatusBerserk) || ai.DoesAny(ImmunityFear) {
			return false
		}
		if ai.DoesAny(ResistanceFear) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusDaze:
		if ai.Has(StatusClarity) || ai.DoesAny(ImmunityDaze) {
			return false
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusConfusion:
		if ai.Has(StatusClarity) || ai.DoesAny(ImmunityConfusion) {
			return false
		}
		if ai.DoesAny(RunicChicken) {
			turns += (turns + 1) / 2
		}
		if ai.DoesAny(ResistanceConfusion) {
			turns = (turns + 1) / 2
		}
	case StatusImbalance:
		if ai.Has(StatusLignification) || ai.DoesAny(ImmunityImbalance) {
			return false
		}
		if ai.DoesAny(ResistanceImbalance) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusLignification:
		if ai.DoesAny(ImmunityLignification) || active || ai.Has(StatusFoggySkin) {
			return false
		}
	case StatusPoison:
		if ai.Has(StatusLignification) || ai.DoesAny(ImmunityPoison) {
			return false
		}
		if ai.DoesAny(ResistancePoison) {
			turns = (turns + 1) / 2
		}
	}
	if turns <= 0 {
		return false
	}
	turns += n
	animateStatus := func() {
		if i == PlayerID && !active {
			// Alert the player with a short animation when a new status
			// becomes active.
			if st.Bad() {
				g.md.BadStatusAnimation()
			} else {
				g.md.GoodStatusAnimation()
			}
		}
	}
	if n >= 1 {
		// We animate before getting the status, as n >= 1 currently
		// means the status is gained during the actor's turn (n == 1),
		// so showing it after would briefly show duration+1, which is
		// a bit unexpected.
		animateStatus()
	}
	if !ai.Statuses.Put(st, turns) {
		return false
	}
	if i == PlayerID {
		g.Stats.Statuses[st]++
		style := logNotable
		if st.Bad() {
			g.StoryLogf("Got %s (%s)", st, fmtTurns(turns, n))
			style = logHurtPlayer
			if !active && GameConfig.WarningsExtra && (st == StatusFire || st == StatusPoison) {
				g.WarningPrompt()
			}
		} else if st == StatusLignification || st == StatusBerserk || st == StatusFoggySkin {
			g.StoryLogf("Became %s (%s)", st, fmtTurns(turns, n))
		}
		g.LogfStyled("You’re %s (%s).", style, st, fmtTurns(turns, n))
		g.md.StopAuto()
	} else if ei := g.Entity(i); g.InFOV(ei.P) {
		g.Logf("The %s is %s (%s).", ei.Name, st, fmtTurns(turns, n))
	}
	switch st {
	case StatusBerserk:
		g.AdjustHP(i, ai, ai.HPBonus())
		g.RemoveStatus(i, ai, StatusFear)
	case StatusClarity:
		g.RemoveStatus(i, ai, StatusConfusion)
		g.RemoveStatus(i, ai, StatusDaze)
		g.RemoveStatus(i, ai, StatusBerserk)
		g.senseMonsters()
	case StatusLignification:
		g.AdjustHP(i, ai, ai.HPBonus())
		g.RemoveStatus(i, ai, StatusPoison)
		g.RemoveStatus(i, ai, StatusFoggySkin)
		g.RemoveStatus(i, ai, StatusImbalance)
		g.RemoveStatus(i, ai, StatusSprint)
	case StatusFoggySkin:
		g.RemoveStatus(i, ai, StatusLignification)
		g.RemoveStatus(i, ai, StatusFire)
	case StatusSprint:
		g.RemoveStatus(i, ai, StatusFocus)
	case StatusFocus:
		g.RemoveStatus(i, ai, StatusSprint)
	}
	if n == 0 {
		animateStatus()
	}
	return true
}

func fmtTurns(turns int, n int) string {
	turns -= n
	return CountNoun("turn", turns)
}

// HPBonus returns the amount of HP for berserk or lignification.
func (a *Actor) HPBonus() int {
	bonus := (2 * a.MaxHP) / 3
	return max(1, bonus)
}

// RemoveStatus removes a status on an actor.
func (g *Game) RemoveStatus(i ID, ai *Actor, st Status) {
	if ai.IsDead() {
		return
	}
	if ai.Statuses[st] <= 0 {
		return
	}
	ai.Statuses[st] = 0
	if i == PlayerID {
		g.LogfStyled("You’re no longer %s.", logStatusEnd, Status(st))
		g.md.StopAuto()
	}
	switch st {
	case StatusPoison:
		dist := MaxFOVRange
		if i == PlayerID {
			g.Log("You expulse stinky and confusing toxins.")
		} else if ei := g.Entity(i); g.InFOV(ei.P) {
			g.Logf("The %s expulses stinky and confusing toxins.", ei.Name)
		}
		g.confusingSmell(i, dist, DurationConfusionToxins)
	case StatusBerserk:
		g.AdjustHP(i, ai, -ai.HPBonus())
		if !ai.DoesAny(Elephanty) {
			mode := g.md.mode
			g.PutStatus(i, ai, StatusPoison, DurationPoisonBerserk)
			// No warning from Berserk's Poison (got warned on
			// previous turn already).
			g.md.mode = mode
		}
	case StatusLignification:
		floor := 3
		if i != PlayerID {
			floor = 1
		}
		g.AdjustHPWithMin(i, ai, -ai.HPBonus(), floor)
		if !ai.DoesAny(WoodyLegs) {
			g.PutStatus(i, ai, StatusImbalance, DurationImbalanceLignification)
		}
		p := g.Entity(i).P
		for j, aj := range g.AdjacentActors(p) {
			if !ai.DoesAny(WoodyLegs) {
				g.PutStatus(j, aj, StatusImbalance, DurationImbalanceLignification)
			}
		}
	}
}

// randomComestible returns a random comestible among inventory and ground
// comestibles.
func (g *Game) randomComestible() (ID, *Comestible) {
	ids := g.ComestibleIDs()
	if len(ids) == 0 {
		return -1, nil
	}
	id := ids[g.IntN(len(ids))]
	return id, g.Entity(id).Role.(*Comestible)
}

// RememberStatusTurns updates the last turn the player spent with each status
// at the start of turn.
func (g *Game) RememberStatusTurns() {
	pa := g.PlayerActor()
	for st, turns := range pa.Statuses {
		if turns <= 0 {
			continue
		}
		g.StatusTurn[st] = g.Turn
	}
}

// ResistStatusRoll reports whether the actor got lucky and was granted
// resistance to the given status. The intention is similar as with lucky rolls
// when at low HP: make it so the player feels unlucky less often. It's only
// used for close re-applications due to some chance-based on-hit effects;
// deterministic applications shouldn't use such a resist roll.
func (g *Game) ResistStatusRoll(i ID, st Status) bool {
	if i != PlayerID {
		return false
	}
	switch g.Turn - g.StatusTurn[st] {
	case 0:
		// A decent chance to resist re-application of
		// statuses if they're already active or just expired
		// on the same turn.
		// In particular, this reduces the likeliness of an
		// annoying immediate re-application of Daze or a
		// helpful Berserk or Lignification.
		return g.IntN(2) == 0
	case 1:
		switch st {
		case StatusDaze, StatusFire, StatusImbalance:
			// A decent chance to resist re-application of some
			// statuses that can feel annoying if repeated on the
			// turn following expiration without a pause. In
			// particular, this make consecutive fire or imbalance
			// dances less likely.
			return g.IntN(2) == 0
		}
	}
	return false
}

// ProgressStatus reduces the number of remaining turns for a status by one.
func (g *Game) ProgressStatus(i ID, ai *Actor, st Status) {
	turns := ai.Statuses[st]
	if turns <= 1 {
		g.RemoveStatus(i, ai, st)
		return
	}
	ai.Statuses[st]--
	if turns > 2 || i != PlayerID || !GameConfig.WarningsExtra {
		return
	}
	switch st {
	case StatusBerserk, StatusLignification, StatusDisorient, StatusGluttony, StatusShadow:
		// For statuses where that matters, warn when there's
		// only one turn remaining after progression.
		g.WarningPrompt()
	}
}

// Behavior holds simple Behavior data for a monster.
type Behavior struct {
	Path     []gruid.Point // path to destination
	Guard    gruid.Point   // position to guard (InvalidPos if not)
	Target   gruid.Point   // Current goal (if not static)
	State    Mindstate     // Wandering, Hunting, ...
	SkipTurn bool          // skip turn (like after swapped positions with another monster)
}

// Mindstate represents the kind of monster behavioral states.
type Mindstate int

// Those constants represent all the possible monster mindstates.
const (
	Wandering Mindstate = iota
	Hunting
)

// MindStateString returns a text describing a monster's main state.
func (g *Game) MindStateString(a *Actor) string {
	beh := a.Behavior
	if beh == nil {
		return ""
	}
	switch beh.State {
	case Hunting:
		return "hunting"
	default:
		if g.Mod(ModCorruptedDungeon) {
			return ""
		}
		if beh.Guard != InvalidPos {
			return "guarding"
		}
		return "wandering"
	}
}

// HandleActorTurn handles turn status updates for an actor with given ID, as
// well as monster behavior for non-player actors.
func (g *Game) HandleActorTurn(i ID, ai *Actor) {
	ei := g.Entity(i)
	if ai.IsDead() {
		return
	}
	if ai.Behavior != nil {
		g.UpdateFirePos(i, ai)
		g.HandleMonsterTurn(i, ai)
	}
	if ai.Has(StatusFire) && (ei.P == ai.FirePos || ai.FirePos != InvalidPos && g.FireAt(ei.P)) {
		g.InflictDamage(i, ai, FireDamage, AttackFire)
	}
	if ai.IsDead() {
		// The actor died due to fire or poison (the latter only for a
		// monster, as it would have happened earlier for the player).
		return
	}
	g.handleCornered(i, ai)
	if i == PlayerID {
		// Gluttony is handled separately from other statuses, because
		// it may lead to eating and may hence mess with other
		// statuses, so it should happen before the progression of
		// other statuses, as well as before some special effects like
		// exuding fog.
		g.handleGluttony()
		if ai.IsDead() {
			return
		}
		if ai.Has(StatusFoggySkin) {
			g.exudeFog()
		}
		if ai.Has(StatusGardener) {
			g.garden()
		}
	}
	// NOTE: We make statuses progress at the end of the turn. That means
	// bonus HP from expiring Berserk or Lignification stills protects
	// against Fire or Poison damage happening during the player's turn but
	// not against any damage inflicted by monsters acting later.
	// While not perfect, the alternative is problematic in other ways
	// (many monster statuses would only be partially relevant on their
	// last turn, making it confusing for the player from an UI point of
	// view).
	for _, st := range sortedStatuses {
		g.ProgressStatus(i, ai, st)
	}
}

// handleGluttony handles progression of the Gluttony status.
func (g *Game) handleGluttony() {
	if g.Mod(ModGluttonyRework) {
		return
	}
	pa := g.PlayerActor()
	switch {
	case pa.Statuses[StatusGluttony] <= 0:
	case pa.Statuses[StatusGluttony] == 1:
		// On Gluttony expiration, the Gluttonous Bear cannot control
		// itself anymore and eats a random comestible, or becomes
		// confused if it cannot find one.
		if id, co := g.randomComestible(); id >= 0 {
			g.LogStyled("You cannot hold your appetite any longer!", logStatusEnd)
			co.Use(g, id)
		} else {
			pa.Statuses[StatusGluttony] = 0
			g.LogStyled("You search but cannot find anything to eat!", logStatusEnd)
			g.PutStatus1(PlayerID, pa, StatusConfusion, DurationConfusionGluttony)
		}
	default:
		pa.Statuses[StatusGluttony]--
	}
}

// sortedStatuses contains statuses sorted for progression dependencies,
// so that it's easier to get durations right in a consistent way. NOTE: it
// allows Poison to expire before being applied again too (which may or may not
// make more sense than just prolonging it, though it's a quite rare
// edge-case).
var sortedStatuses = []Status{
	StatusClarity,
	StatusConfusion,
	StatusDaze,
	StatusDig,
	StatusDisorient,
	StatusFear,
	StatusFire,
	StatusFocus,
	StatusFoggySkin,
	StatusGardener,
	StatusImbalance,
	StatusPoison,
	StatusShadow,
	StatusSprint,
	StatusTimeStop,
	StatusVampirism,

	StatusBerserk,       // after Poison
	StatusLignification, // after Imbalance & Berserk
}

// exudeFog implements the fog for the player-only status effect Foggy-Skin.
func (g *Game) exudeFog() {
	pp := g.PP()
	for q := range g.Map.PassableNeighbors(pp) {
		if q == pp.Add(g.Dir) {
			// NOTE: this assumes exudeFog is only called for the
			// player for now (currently foggy skin status is
			// player-only).
			continue
		}
		g.NormalCloudAt(q, 2+g.IntN(3))
	}
}

// garden implements gardening for the player-only status effect Gardener.
func (g *Game) garden() {
	pp := g.PP()
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, 2)
	for _, n := range nodes {
		if i, _ := g.ItemAt(n.P); i >= 0 {
			continue
		}
		if g.Map.Terrain.At(n.P) == Floor {
			g.Map.Terrain.Set(n.P, Foliage)
		}
	}
}

// handleCornered makes cornered and afraid actors become berserk.
func (g *Game) handleCornered(i ID, ai *Actor) {
	switch {
	case ai.Has(StatusClarity):
		// Players under clarity cannot become Berserk.
		return
	case ai.Has(StatusFear) || ai.Is(NoisyImp) && !ai.Has(StatusBerserk):
		// Actor is afraid or is playing music while not berserk: it
		// may become berserk if cornered.
	case i == PlayerID && ai.DoesAny(Elephanty):
		g.handleCorneredWithRatPhobia()
		return
	default:
		// The actor isn't afraid of anything, so it cannot feel
		// cornered.
		return
	}
	ei := g.Entity(i)
	from := ei.P
	if !g.InFOV(from) {
		return
	}
	for _, dir := range Directions {
		p := from.Add(dir)
		if !inMap(p) {
			continue
		}
		if !g.Map.Passable(p) && !ai.Has(StatusDig) && !ai.DoesAny(MonsDig) {
			continue
		}
		if j, _ := g.ActorAt(p); j >= 0 {
			continue
		}
		j, _ := g.RangedTargetInDir(from, dir)
		if j < 0 {
			// One direction is free of any actors: the afraid
			// actor is not cornered yet.
			return
		}
		if i != PlayerID && j != PlayerID {
			// The afraid monster is not cornered yet: one of the
			// directions has free space and no player.
			return
		}
	}
	// The actor is afraid and cornered: become berserk for a short while.
	if i == PlayerID {
		g.Log("You feel cornered.")
		g.Stats.Cornered++
	} else if g.InFOV(from) {
		g.LogfStyled("The afraid %s feels cornered.", logSpecial, ei.Name)
	}
	g.PutStatus1(i, ai, StatusBerserk, DurationBerserkAfraid)
}

func (g *Game) handleCorneredWithRatPhobia() {
	from, pa := g.PP(), g.PlayerActor()
	for _, dir := range Directions {
		p := from.Add(dir)
		if !inMap(p) {
			continue
		}
		if !g.Map.Passable(p) && !pa.Has(StatusDig) && !pa.DoesAny(MonsDig) {
			continue
		}
		if j, aj := g.ActorAt(p); j >= 0 && g.PlayerFears(j, aj) {
			continue
		}
		if j, aj := g.RangedTargetInDir(from, dir); j < 0 || !g.PlayerFears(j, aj) {
			// One direction is free of any fearsome actors: the
			// player is not cornered yet.
			return
		}
	}
	// You feel cornered: become berserk for a short while.
	g.Log("You feel cornered.")
	g.Stats.Cornered++
	g.PutStatus1(PlayerID, pa, StatusBerserk, DurationBerserkAfraid)
}

// TeleportActor teleports the given actor, adjusting final duration of debuff
// statuses by n.
func (g *Game) TeleportActor(i ID, ai *Actor, n int) bool {
	if ai.ResistsMove() {
		return false
	}
	from := g.Entity(i).P
	to := g.TeleportPoint(from)
	if i == PlayerID && g.Map.Orb != InvalidPos && g.IntN(2) == 0 {
		// On last level, give the orb a chance to drive the player
		// away from it, to make teleport-based strategies less
		// reliable (but still viable enough!). Greater chance with
		// ModCorruptedDungeon, because even more unpredictability
		// sounds thematic there.
		//
		// Result: teleport to furthest point from the orb among 3
		// random possible teleport points.
		for range 2 {
			if nto := g.TeleportPoint(from); paths.DistanceManhattan(nto, g.Map.Orb) > paths.DistanceManhattan(to, g.Map.Orb) {
				to = nto
			}
		}
		g.LogStyled("A strange energy disturbs your teleportation!", logSpecial)
	}
	g.MoveActor(i, ai, to, MovTeleport)
	g.teleportStatuses(i, ai, n)
	return true
}

// TeleportPoint returns a suitable teleport destination for the current given
// position. In particular, it ensures that the destination is floor, rubble or
// foliage, unoccupied by any other actor or trap, and not too close to the
// origin.
func (g *Game) TeleportPoint(from gruid.Point) gruid.Point {
	const mindist = 2 * MaxFOVRange
	for {
		p := g.RandomPassableWithoutTrap()
		if paths.DistanceManhattan(from, p) < mindist {
			continue
		}
		if i, _ := g.ActorAt(p); i >= 0 {
			continue
		}
		return p
	}
}
