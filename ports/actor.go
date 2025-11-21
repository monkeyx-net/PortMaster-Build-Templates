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
	Traits    Traits      // passives traits
	Behavior  *Behavior   // behavior component (monsters only)
}

// NewActor returns a new actor with the given combat stats and traits.
func NewActor(attack, defense, hp int, t Traits) *Actor {
	a := &Actor{
		Attack:   attack,
		Defense:  defense,
		HP:       hp,
		MaxHP:    hp,
		Statuses: make(Statuses, NStatuses),
		Traits:   t,
	}
	if !a.IsPlayer() {
		a.Behavior = &Behavior{Guard: InvalidPos, Target: InvalidPos}
	}
	return a
}

// IsPlayer reports whether the actor has the Player trait.
func (a *Actor) IsPlayer() bool {
	return a.DoesAny(Player)
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

// CanMove reports whether the actor isn't under any status that prevents
// movement.
func (a *Actor) CanMove() bool {
	return !a.Has(StatusLignification) || a.DoesAny(WoodyLegs)
}

// ResistsMove reports whether the actor resists against forced movement. This
// includes lignified actors that can still move due to woodylegs as well as
// walking trees.
func (a *Actor) ResistsMove() bool {
	return a.Has(StatusLignification) || a.DoesAny(MonsLignify)
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
	if ai.IsPlayer() {
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

// Traits. NOTE: there shouldn't be more than 64.
const (
	Player Traits = 1 << iota

	PatternCatch        // unbalancing and catching attack (frog, octopode)
	PatternFourDirs     // four-directional attack in 4 dirs (hydra players)
	PatternFourDirsMons // four-directional attack (four-headed hydra, undead knight)
	PatternRampage      // rampaging (boar, phoenix)
	PatternRanged       // ranged attack (eye, llama, lich)
	PatternRangedRecoil // ranged attack with recoil (wind fox)
	PatternSwap         // swapping-attack (crazy druid)
	PatternSwapDaze     // dazing swapping-attack (temporal cat)

	BadHearing    // worse hearing of noise (footsteps)
	BurningHits   // hits may burn
	DazingSpines  // body covered in dazing spines (porcupine)
	GoodHearing   // better hearing of noise (footsteps)
	GoodSmell     // smell food from a distance
	Pushing       // unbalancing pushing (earth-dragon)
	PushingCharge // unbalancing pushing when charging (boar)
	TrailingCloud // leave dust behind after moving (on floor tiles)
	VenomousMelee // (non-ranged) attacks may poison foes
	WoodyLegs     // can move while lignified

	VulnerabilityFire // increases duration of fire status

	ResistanceConfusion // halves duration of confusion status
	ResistanceDaze      // allows spirit invocation while dazed
	ResistanceFear      // halves duration of fear status
	ResistanceFire      // halves duration of fire status
	ResistanceImbalance // halves duration of imbalance status
	ResistancePoison    // halves duration of poison status

	// Monster-only traits, often similar to player's spirit passive or
	// active abilities but not exactly the same and with different
	// activation conditions.

	MonsBarking        // bark with a fear inducing chance on noticing the player
	MonsBerserking     // berserking spiders may make you berserk
	MonsBlink          // attack makes you blink
	MonsConfusion      // confusing attack (like floating eyes)
	MonsDig            // dig walls when hunting
	MonsExplodingDeath // explode on death
	MonsFear           // attack may frighten
	MonsFourHeaded     // perform 4 attacks per attack
	MonsHungry         // hunts you with smell, becomes berserk if you eat in view
	MonsIgnoreDefense  // attack ignores defense
	MonsLignify        // attack may lignify you (also makes unmovable by force)
	MonsSpores         // lignifying spores
	MonsMusic          // makes music for you (noisy imp)
	MonsScales         // scales provide extra protection from ranged attacks
	MonsSpitFire       // spit fire when ranged (requires PatternRanged)
	MonsTeleport       // attack may teleport you

	MonsImmunityConfusion     // immune to confusion
	MonsImmunityDaze          // immune to daze status effect
	MonsImmunityFear          // immune to fear
	MonsImmunityFire          // immune to fire
	MonsImmunityImbalance     // immune to imbalance
	MonsImmunityLignification // immune to lignification
	MonsImmunityPoison        // immune to poison

	// We don't do descriptions for those (beyond the log messages about noise).
	MonsCreep          // creep noise (instead of footsteps)
	MonsHeavyFootsteps // heavy footsteps
	MonsLightFootsteps // light footsteps
	MonsSilent         // silent movement
	MonsWingFlap       // wing flapping noise (instead of footsteps)
	MonsNotable        // dangerous monster (mention in timeline)

	NoTraits = 0
)

func (t Traits) String() string {
	desc := []string{}
	desc = traitDesc(desc, t, PatternFourDirs, "four-directional attack with attack bonus per extra foe")
	desc = traitDesc(desc, t, PatternFourDirsMons, "four-directional attack")
	desc = traitDesc(desc, t, PatternCatch, "ranged catching attack that @Ounbalances@N foes")
	desc = traitDesc(desc, t, PatternRampage, "rampaging charge")
	desc = traitDesc(desc, t, PatternRanged, "ranged attack")
	desc = traitDesc(desc, t, PatternRangedRecoil, "ranged attack with recoil")
	desc = traitDesc(desc, t, PatternSwap, "ranged space-distorting silent attack")
	desc = traitDesc(desc, t, PatternSwapDaze, "ranged space-distorting silent attack that may @Odaze@N foes")

	desc = traitDesc(desc, t, Pushing, "hits may push and @Ounbalance@N foes")
	desc = traitDesc(desc, t, PushingCharge, "charging pushes and @Ounbalances@N foes")
	desc = traitDesc(desc, t, BurningHits, "hits may @Oburn@N foes")
	desc = traitDesc(desc, t, VenomousMelee, "@Ovenomous@N melee attack")
	desc = traitDesc(desc, t, DazingSpines, "spines may @Odaze@N foes in melee")
	switch {
	case t.All(GoodHearing) && !t.All(BadHearing):
		desc = append(desc, "good hearing")
	case !t.All(GoodHearing) && t.All(BadHearing):
		desc = append(desc, "bad hearing")
	}
	desc = traitDesc(desc, t, GoodSmell, "smells food from afar")
	desc = traitDesc(desc, t, TrailingCloud, "movement leaves dust behind")
	desc = traitDesc(desc, t, WoodyLegs, "woody legs (can move while lignified)")

	desc = traitDesc(desc, t, VulnerabilityFire, "@OFire@N vulnerability (longer duration)")

	desc = traitDesc(desc, t, ResistanceConfusion, "@OConfusion@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistanceDaze, "@ODaze@N resistance (can invoke spirits while dazed)")
	desc = traitDesc(desc, t, ResistanceFear, "@OFear@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistanceFire, "@OFire@N resistance (delayed burning)")
	desc = traitDesc(desc, t, ResistanceImbalance, "@OImbalance@N resistance (half duration)")
	desc = traitDesc(desc, t, ResistancePoison, "@OPoison@N resistance (half duration)")

	desc = traitDesc(desc, t, MonsBarking, "barks on sight and @Ofrightens@N you")
	desc = traitDesc(desc, t, MonsBerserking, "hits may @Bberserk@N you briefly")
	desc = traitDesc(desc, t, MonsBlink, "hits makes you blink")
	desc = traitDesc(desc, t, MonsConfusion, "@Oconfusing@N hits")
	desc = traitDesc(desc, t, MonsDig, "digs walls when hunting")
	desc = traitDesc(desc, t, MonsExplodingDeath, "explodes on death")
	desc = traitDesc(desc, t, MonsFear, "@Ofrightening@N hits")
	desc = traitDesc(desc, t, MonsFourHeaded, "attacks with all four heads")
	desc = traitDesc(desc, t, MonsHungry, "hunts you with smell, becomes berserk if you eat in view")
	desc = traitDesc(desc, t, MonsIgnoreDefense, "attack ignores defense")
	desc = traitDesc(desc, t, MonsLignify, "hits may @Blignify@N you briefly, woody legs, permanently @Blignified@N")
	desc = traitDesc(desc, t, MonsMusic, "plays guitar for you, avoids fights")
	desc = traitDesc(desc, t, MonsScales, "scales (+1 Defense against ranged attacks)")
	desc = traitDesc(desc, t, MonsSpitFire, "spites @Ofire@N from afar")
	desc = traitDesc(desc, t, MonsSpores, "releases @Blignifying@N spores on sight")

	desc = traitDesc(desc, t, MonsImmunityConfusion, "immune to @OConfusion@N")
	desc = traitDesc(desc, t, MonsImmunityDaze, "@ODaze@N immunity")
	desc = traitDesc(desc, t, MonsImmunityFear, "@OFear@N immunity")
	desc = traitDesc(desc, t, MonsImmunityFire, "@OFire@N immunity")
	desc = traitDesc(desc, t, MonsImmunityImbalance, "@OImbalance@N immunity")
	desc = traitDesc(desc, t, MonsImmunityLignification, "@BLignification@N immunity")
	desc = traitDesc(desc, t, MonsImmunityPoison, "@OPoison@N immunity")

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
	StatusFear
	StatusFire
	StatusFocus
	StatusFoggySkin
	StatusImbalance
	StatusLignification
	StatusPoison
	StatusSprint
	StatusTimeStop
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
const NStatuses = StatusTimeStop + 1

var statusAbbrevation = []string{
	StatusBerserk:       "Ber",
	StatusClarity:       "Cla",
	StatusConfusion:     "Con",
	StatusDaze:          "Daz",
	StatusDig:           "Dig",
	StatusFear:          "Fea",
	StatusFire:          "Fir",
	StatusFocus:         "Foc",
	StatusImbalance:     "Imb",
	StatusFoggySkin:     "Fog",
	StatusLignification: "Lig",
	StatusPoison:        "Poi",
	StatusSprint:        "Spr",
	StatusTimeStop:      "Tim",
}

func (st Status) Abbr() string {
	return statusAbbrevation[st]
}

var statusName = []string{
	StatusBerserk:       "Berserk",
	StatusClarity:       "Clarity",
	StatusConfusion:     "Confusion",
	StatusDaze:          "Daze",
	StatusDig:           "Dig",
	StatusFear:          "Fear",
	StatusFire:          "Fire",
	StatusFocus:         "Focus",
	StatusFoggySkin:     "Foggy Skin",
	StatusImbalance:     "Imbalance",
	StatusLignification: "Lignification",
	StatusPoison:        "Poison",
	StatusSprint:        "Sprint",
	StatusTimeStop:      "Time Stop",
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
	StatusFear:          "afraid",
	StatusFire:          "on fire",
	StatusFocus:         "focusing",
	StatusFoggySkin:     "foggy",
	StatusImbalance:     "imbalanced",
	StatusLignification: "lignified",
	StatusPoison:        "poisoned",
	StatusSprint:        "sprinting",
	StatusTimeStop:      "freezing time",
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
	StatusFear:      "Forbids attacking and walking toward foes.",
	StatusFire:      fmt.Sprintf("Remaining in place or moving into a cloud of fire will burn you for %d damage. When first catching @OFire@N, you also get burned once immediately.", FireDamage),
	StatusFocus:     "Attack frontally with all four heads.",
	StatusFoggySkin: "Exudes fog on the sides and behind. Protects from @OFire@N and @BLignification@N.",
	StatusImbalance: fmt.Sprintf("Halves attack. When also confused, gives %+d Defense.", ConfusedImbalanceDefenseBonus),
	StatusLignification: fmt.Sprintf(
		"Gives %+d Defense with an extra +1 against ranged attacks. Caps incoming attack damage at 1. Provides temporary HP bonus. Protects from @OPoison@N and @OImbalance@N. Prevents movement. Increases @OFire@N duration by 2. Followed by @OImbalance@N for you and any adjacent actors on expiration for %d turns.",
		LignificationDefenseBonus, DurationImbalanceLignification),
	StatusPoison: fmt.Sprintf("Walking will hurt you for %d damage. Stinky toxins are exuded on expiration, confusing other creatures for %d turns.",
		PoisonDamage, DurationConfusionToxins),
	StatusSprint:   "You can move to a visible destination three times as fast on the current direction and twice as fast laterally. You can jump over foes, unbalancing them. Canceled by waiting.",
	StatusTimeStop: "Time is frozen for everyone but you until expiration. Status effects affecting you will still progress.",
}

func (st Status) Desc() string {
	return statusDesc[st]
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
	switch {
	case st == StatusBerserk:
		if active || ai.Has(StatusClarity) {
			return false
		}
	case st == StatusFire:
		if ai.Has(StatusFoggySkin) || ai.DoesAny(MonsImmunityFire) {
			return false
		}
		if ai.DoesAny(VulnerabilityFire) || ai.Has(StatusLignification) {
			turns += 2
		}
		if active || ai.DoesAny(ResistanceFire) {
			// If Fire is already active or we have fire
			// resistance, we skip extra damage at onset.
			break
		}
		g.InflictDamage(i, ai, FireDamage, AttackFireCatch)
		if ai.IsDead() {
			return false
		}
	case st == StatusFear:
		if ai.Has(StatusBerserk) || ai.DoesAny(MonsImmunityFear) {
			return false
		}
		if ai.DoesAny(ResistanceFear) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case st == StatusDaze:
		if ai.Has(StatusClarity) || ai.DoesAny(MonsImmunityDaze) {
			return false
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case st == StatusConfusion:
		if ai.Has(StatusClarity) || ai.DoesAny(MonsImmunityConfusion) {
			return false
		}
		if ai.DoesAny(ResistanceConfusion) {
			turns = (turns + 1) / 2
		}
	case st == StatusImbalance:
		if ai.Has(StatusLignification) || ai.DoesAny(MonsImmunityImbalance) {
			return false
		}
		if ai.DoesAny(ResistanceImbalance) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case st == StatusLignification:
		if ai.DoesAny(MonsImmunityLignification) || active || ai.Has(StatusFoggySkin) {
			return false
		}
	case st == StatusPoison:
		if ai.Has(StatusLignification) || ai.DoesAny(MonsImmunityPoison) {
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
	if !ai.Statuses.Put(st, turns) {
		return false
	}
	if i == PlayerID {
		g.Stats.Statuses[st]++
		style := logNotable
		if st.Bad() {
			g.StoryLogf("Got %s (%s)", st, fmtTurns(turns, n))
			style = logHurtPlayer
			if GameConfig.WarnFirePoison && (st == StatusFire || st == StatusPoison) {
				g.md.mode = modeCritical
			}
		} else if st == StatusLignification || st == StatusBerserk || st == StatusFoggySkin {
			g.StoryLogf("Became %s (%s)", st, fmtTurns(turns, n))
		}
		g.LogfStyled("You’re %s (%s).", style, st, fmtTurns(turns, n))
		g.md.auto.mode = noAuto
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
	if i == PlayerID && !active {
		// Alert the player with a short animation when a new status
		// becomes active.
		if st.Bad() {
			g.md.BadStatusAnimation()
		} else {
			g.md.GoodStatusAnimation()
		}
	}
	return true
}

func fmtTurns(turns int, n int) string {
	turns -= n
	return fmt.Sprintf("%d %s", turns, Countable("turn", turns))
}

// HPBonus returns the amount of HP for berserk or lignification.
func (a *Actor) HPBonus() int {
	bonus := (2 * a.MaxHP) / 3
	if bonus == 0 {
		bonus = 1
	}
	return bonus
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
		g.md.auto.mode = noAuto
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
		g.PutStatus(i, ai, StatusPoison, DurationPoisonBerserk)
	case StatusLignification:
		g.AdjustHPWithMin(i, ai, -ai.HPBonus(), 3)
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

// ProgressStatus reduces the number of remaining turns for a status by one.
func (g *Game) ProgressStatus(i ID, ai *Actor, st Status) {
	if ai.Statuses[st] <= 1 {
		g.RemoveStatus(i, ai, st)
	} else {
		ai.Statuses[st]--
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
func (a *Actor) MindStateString() string {
	beh := a.Behavior
	if beh == nil {
		return ""
	}
	switch beh.State {
	case Hunting:
		return "hunting"
	default:
		if beh.Guard != InvalidPos {
			return "guarding"
		}
		return "wandering"
	}
}

// HandleActorTurn handles turn status updates for an actor with given ID, as
// well as monster behavior for non-player actors.
func (g *Game) HandleActorTurn(i ID, ai *Actor) {
	if ai.IsDead() {
		return
	}
	ei := g.Entity(i)
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
	if ai.Has(StatusFoggySkin) {
		g.exudeFog()
	}
	g.handleCornered(i, ai)
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
	StatusFear,
	StatusFire,
	StatusFocus,
	StatusFoggySkin,
	StatusImbalance,
	StatusPoison,
	StatusSprint,
	StatusTimeStop,

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

// handleCornered makes cornered and afraid actors become berserk.
func (g *Game) handleCornered(i ID, ai *Actor) {
	if !ai.Has(StatusFear) && !ai.DoesAny(MonsMusic) || ai.Has(StatusClarity) {
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
	} else if g.InFOV(from) {
		g.LogfStyled("The afraid %s feels cornered.", logSpecial, ei.Name)
	}
	g.PutStatus1(i, ai, StatusBerserk, DurationBerserkAfraid)
}

// TeleportActor teleports the given actor, adjusting final duration of debuff
// statuses by n.
func (g *Game) TeleportActor(i ID, ai *Actor, n int) bool {
	if ai.ResistsMove() {
		return false
	}
	from := g.Entity(i).P
	to := g.TeleportPoint(from)
	if i == PlayerID || g.InFOV(from) {
		g.md.TeleportAnimation(from, to, i == PlayerID || g.InFOV(to))
	}
	g.MoveActor(i, ai, to)
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
