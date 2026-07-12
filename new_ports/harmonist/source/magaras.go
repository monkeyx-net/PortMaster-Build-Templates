package main

import (
	"errors"
	"fmt"
	"slices"

	"codeberg.org/anaseto/gruid"
)

// Magara represents a magara.
type Magara struct {
	Kind    magaraKind // kind of magara
	Charges int        // number of charges
	Depth   int        // map depth where you found it
}

type magaraKind int

const (
	NoMagara magaraKind = iota
	BlinkMagara
	DigMagara
	TeleportMagara
	SwiftnessMagara
	LevitationMagara
	FireMagara
	FogMagara
	ShadowsMagara
	NoisesMagara
	ConfusionMagara
	SleepingMagara
	TeleportOtherMagara
	SwappingMagara
	DazingMagara
	ObstructionMagara
	LignificationMagara
	EnergyMagara
	TransparencyMagara
	DisguiseMagara
	DelayedNoiseMagara
	DispersalMagara
	OricExplosionMagara
)

// Harmonist reports whether the magara is of harmonic nature (for achievement
// purposes).
func (mag Magara) Harmonic() bool {
	switch mag.Kind {
	case FogMagara, ShadowsMagara, NoisesMagara, ConfusionMagara,
		SleepingMagara, DazingMagara, TransparencyMagara, DisguiseMagara,
		DelayedNoiseMagara:
		return true
	default:
		return false
	}
}

// Oric reports whether the magara is of oric nature (for achievement
// purposes).
func (mag Magara) Oric() bool {
	switch mag.Kind {
	case BlinkMagara, DigMagara, TeleportMagara, LevitationMagara,
		TeleportOtherMagara, SwappingMagara, ObstructionMagara,
		DispersalMagara, OricExplosionMagara:
		return true
	default:
		return false
	}
}

// DefaultCharges returns the default number of charges for a magara kind.
func (m magaraKind) DefaultCharges() int {
	switch m {
	case LevitationMagara, ShadowsMagara, FogMagara, NoisesMagara, FireMagara, DelayedNoiseMagara:
		return 6
	case ObstructionMagara, TransparencyMagara, BlinkMagara, LignificationMagara,
		OricExplosionMagara, DispersalMagara:
		return 5
	case EnergyMagara:
		return 1
	default:
		// ConfusionMagara, DigMagara, TeleportMagara, SwiftnessMagara,
		// TeleportOtherMagara, SleepingMagara, SwappingMagara,
		// DazingMagara, DisguiseMagara:
		return 4
	}
}

// RandomStartingMagara returns a random starting magara.
func (g *Game) RandomStartingMagara() Magara {
	mags := []magaraKind{BlinkMagara, DigMagara, TeleportMagara,
		SwiftnessMagara, LevitationMagara, FireMagara, FogMagara,
		ShadowsMagara, NoisesMagara, ConfusionMagara, SleepingMagara,
		TeleportOtherMagara, SwappingMagara, DazingMagara,
		ObstructionMagara, LignificationMagara, TransparencyMagara,
		DelayedNoiseMagara, DisguiseMagara}
	var mag magaraKind
loop:
	for {
		mag = mags[RandInt(len(mags))]
		for _, m := range g.ProcInfo.GeneratedMagaras {
			if m == mag {
				continue loop
			}
		}
		break
	}
	return Magara{Kind: mag, Charges: mag.DefaultCharges()}
}

// RandomMagara returns a random magara that hasn't been generated yet.
func (g *Game) RandomMagara() Magara {
	mags := []magaraKind{BlinkMagara, DigMagara, TeleportMagara,
		SwiftnessMagara, LevitationMagara, FireMagara, FogMagara,
		ShadowsMagara, NoisesMagara, ConfusionMagara, SleepingMagara,
		TeleportOtherMagara, SwappingMagara, DazingMagara,
		ObstructionMagara, LignificationMagara, TransparencyMagara,
		DelayedNoiseMagara, DisguiseMagara, EnergyMagara}
	var mag magaraKind
loop:
	for {
		mag = mags[RandInt(len(mags))]
		if slices.Contains(g.ProcInfo.GeneratedMagaras, mag) {
			continue loop
		}
		break
	}
	return Magara{Kind: mag, Charges: mag.DefaultCharges(), Depth: -1}
}

// EquipMagara equips the the magara on the ground on the given slot index.
func (g *Game) EquipMagara(i int) (err error) {
	omagara := g.Player.Magaras[i]
	g.Player.Magaras[i] = g.Objects.Magaras[g.Player.P]
	if g.Player.Magaras[i].Depth == -1 {
		g.Player.Magaras[i].Depth = g.Map.Depth
	}
	g.Objects.Magaras[g.Player.P] = omagara
	g.Logf("You take the %s.", g.Player.Magaras[i])
	g.Logf("You leave the %s.", omagara)
	g.StoryPrintf("Took %s (%d), left %s (%d)", g.Player.Magaras[i], g.Player.Magaras[i].Charges, omagara, omagara.Charges)
	return nil
}

// UseMagara uses the given magara by slot index.
func (g *Game) UseMagara(i int) (err error) {
	if g.Player.HasStatus(StatusConfusion) {
		return errors.New("You cannot use magaras while confused.")
	}
	mag := g.Player.Magaras[i]
	if mag.Kind == NoMagara {
		return errors.New("You cannot evoke an empty slot!")
	}
	if mag.MPCost(g) > g.Player.MP {
		return errors.New("Not enough magic points for using this magara.")
	}
	if mag.Charges <= 0 {
		return errors.New("Not enough charges for using this magara.")
	}
	switch mag.Kind {
	case BlinkMagara:
		err = g.EvokeBlink()
	case DigMagara:
		err = g.EvokeDig()
	case TeleportMagara:
		err = g.EvokeTeleport()
	case SwiftnessMagara:
		err = g.EvokeSwiftness()
	case LevitationMagara:
		err = g.EvokeLevitation()
	case FireMagara:
		err = g.EvokeFire()
	case FogMagara:
		err = g.EvokeFog()
	case ShadowsMagara:
		err = g.EvokeShadows()
	case NoisesMagara:
		err = g.EvokeNoise()
	case ConfusionMagara:
		err = g.EvokeConfusion()
	case DazingMagara:
		err = g.EvokeDazing()
	case SleepingMagara:
		err = g.EvokeSleeping()
	case TeleportOtherMagara:
		err = g.EvokeTeleportOther()
	case SwappingMagara:
		err = g.EvokeSwapping()
	case ObstructionMagara:
		err = g.EvokeObstruction()
	case LignificationMagara:
		err = g.EvokeLignification()
	case EnergyMagara:
		err = g.EvokeEnergyMagara()
	case TransparencyMagara:
		err = g.EvokeTransparencyMagara()
	case DisguiseMagara:
		err = g.EvokeDisguiseMagara()
	case DelayedNoiseMagara:
		err = g.EvokeDelayedNoiseMagara()
	case DispersalMagara:
		err = g.EvokeDispersalMagara()
	case OricExplosionMagara:
		err = g.EvokeOricExplosionMagara()
	}
	if err != nil {
		return err
	}
	g.Stats.MagarasUsed++
	g.Stats.UsedMagaras[mag.Kind]++
	g.Stats.DMagaraUses[g.Map.Depth]++
	g.Player.MP -= mag.MPCost(g)
	g.Player.Magaras[i].Charges--
	g.StoryPrintf("Evoked %s (MP: %d, Charges: %d)", mag, g.Player.MP, g.Player.Magaras[i].Charges)
	if mag.Harmonic() {
		g.Stats.HarmonicMagUse++
		if g.Stats.HarmonicMagUse == 6 {
			AchHarmonistNovice.Get(g)
		}
		if g.Stats.HarmonicMagUse == 11 {
			AchHarmonistInitiate.Get(g)
		}
		if g.Stats.HarmonicMagUse == 16 {
			AchHarmonistMaster.Get(g)
		}
	} else if mag.Oric() {
		g.Stats.OricMagUse++
		if g.Stats.OricMagUse == 6 {
			AchNoviceOricCelmist.Get(g)
		}
		if g.Stats.OricMagUse == 11 {
			AchInitiateOricCelmist.Get(g)
		}
		if g.Stats.OricMagUse == 16 {
			AchMasterOricCelmist.Get(g)
		}
	} else if mag.Kind == FireMagara {
		g.Stats.FireUse++
		if g.Stats.FireUse == 2 {
			AchPyromancerNovice.Get(g)
		}
		if g.Stats.FireUse == 4 {
			AchPyromancerInitiate.Get(g)
		}
		if g.Stats.FireUse == 6 {
			AchPyromancerMaster.Get(g)
		}
	}
	switch mag.Kind {
	case TeleportMagara, TeleportOtherMagara, BlinkMagara, SwappingMagara, DispersalMagara:
		g.Stats.OricTelUse++
		if g.Stats.OricTelUse == 14 {
			AchTeleport.Get(g)
		}
	}
	return nil
}

func (mag Magara) String() (desc string) {
	return mag.Kind.String()
}

func (magk magaraKind) String() (desc string) {
	switch magk {
	case NoMagara:
		desc = "empty slot"
	case BlinkMagara:
		desc = "magara of blinking"
	case DigMagara:
		desc = "magara of digging"
	case TeleportMagara:
		desc = "magara of teleportation"
	case SwiftnessMagara:
		desc = "magara of swiftness"
	case LevitationMagara:
		desc = "magara of levitation"
	case FireMagara:
		desc = "magara of fire"
	case FogMagara:
		desc = "magara of fog"
	case ShadowsMagara:
		desc = "magara of shadows"
	case NoisesMagara:
		desc = "magara of noises"
	case ConfusionMagara:
		desc = "magara of confusion"
	case SleepingMagara:
		desc = "magara of sleeping"
	case TeleportOtherMagara:
		desc = "magara of teleport other"
	case SwappingMagara:
		desc = "magara of swapping"
	case DazingMagara:
		desc = "magara of dazing"
	case ObstructionMagara:
		desc = "magara of obstruction"
	case LignificationMagara:
		desc = "magara of lignification"
	case EnergyMagara:
		desc = "magara of energy"
	case TransparencyMagara:
		desc = "magara of transparency"
	case DisguiseMagara:
		desc = "magara of disguise"
	case DelayedNoiseMagara:
		desc = "magara of delayed noise"
	case DispersalMagara:
		desc = "magara of dispersal"
	case OricExplosionMagara:
		desc = "magara of oric explosion"
	}
	return desc
}

// ShortDesc returns a short description of the magara for use in the evocation
// menu.
func (mag Magara) ShortDesc() string {
	if mag.Kind == NoMagara {
		return mag.String()
	}
	return fmt.Sprintf("%s (%d)", mag.String(), mag.Charges)
}

// Desc provides a long description for the magara.
func (mag Magara) Desc(g *Game) (desc string) {
	switch mag.Kind {
	case NoMagara:
		return "You do not have a magara equipped on this slot yet."
	case BlinkMagara:
		desc = "Makes you blink away within your line of sight by using an oric energy disturbance. The magara is more susceptible to send you to the cells that are most far from you."
	case DigMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusDig.String(), DurationDigging, StatusDig.Desc())
	case TeleportMagara:
		desc = "Creates an oric energy disturbance, making you teleport far away on the same level."
	case SwiftnessMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusSwift.String(), DurationSwiftness, StatusSwift.Desc())
	case LevitationMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusLevitation.String(), DurationLevitation, StatusLevitation.Desc())
	case FireMagara:
		desc = "Throws small magical sparks at flammable terrain adjacent to you. Flammable terrain is first consumed by magical flames that are by themselves harmless to creatures, and then smoke will produce purple mist around, inducing sleep and confusion in monsters. As a gawalt monkey, you resist sleepiness, but you will still feel confused. The fire does often expand to other adjacent flammable terrain."
	case FogMagara:
		desc = "Creates a dense fog cloud in a 3-range radius using harmonic energies. The cloud will dissipate with time."
	case ShadowsMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusShadows.String(), DurationShadows, StatusShadows.Desc())
	case NoisesMagara:
		desc = "Tricks monsters in a 12-range area with harmonic magical sounds, making them go away from you for a few turns. The possible monster destinations will be marked with noise symbols in the map. It only works on monsters that are not already seeing you."
	case ConfusionMagara:
		desc = fmt.Sprintf("Confuses monsters in sight with harmonic light, leaving them unable to attack you for %d turns.", DurationConfusionMonster)
	case DazingMagara:
		desc = fmt.Sprintf("Makes monsters in sight unable to act by disturbing their senses with harmonic illusions for %d turns.", DurationDazeMonster)
	case SleepingMagara:
		desc = "Induces deep sleeping and exhaustion for monsters you see in cardinal directions using hypnotic harmonic illusions."
	case TeleportOtherMagara:
		desc = "Creates oric energy disturbances, teleporting monsters you see in cardinal directions."
	case SwappingMagara:
		desc = "Creates an oric distortion that makes you swap positions with the farthest monster in sight. If there is more than one at the same distance, it will be chosen randomly."
	case ObstructionMagara:
		desc = fmt.Sprintf("Creates temporal barriers with oric energy between you and monsters in sight. They last at least %d turns.", DurationMagicalBarrier)
	case LignificationMagara:
		desc = fmt.Sprintf("Liberates magical beams that lignify monsters in view, so that they cannot move for %d turns. The monsters can still fight.", DurationLignificationMonster)
	case EnergyMagara:
		desc = "Replenishes your MP and HP. Unlike other magaras, it doesn’t require MP to use."
	case TransparencyMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusTransparent.String(), DurationTransparency, StatusTransparent.Desc())
	case DisguiseMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusDisguised.String(), DurationDisguise, StatusDisguised.Desc())
	case DelayedNoiseMagara:
		desc = fmt.Sprintf("Will produce a thunderous harmonic noise in your current cell after a @YDelay@N of %d turns.", DurationHarmonicNoiseDelay)
	case DispersalMagara:
		desc = fmt.Sprintf("Gives @B%s@N for %d turns. %s", StatusDispersal.String(), DurationDispersal, StatusDispersal.Desc())
	case OricExplosionMagara:
		desc = fmt.Sprintf("Will produce a big rock-destroying oric explosion at your current cell after a @YDelay@N of %d turns. It destroys only walls.", DurationHarmonicNoiseDelay)
	}
	desc += "\n"
	if mag.Depth >= 1 {
		desc += fmt.Sprintf("\nYou found it at dungeon depth %d.", mag.Depth)
	}
	desc += fmt.Sprintf("\nIt currently has %d charges.", mag.Charges)
	return desc
}

// MPCost returns the MP cost of the magara (usually 1).
func (mag Magara) MPCost(g *Game) int {
	if mag.Kind == NoMagara || mag.Kind == EnergyMagara {
		return 0
	}
	cost := 1
	return cost
}

func (g *Game) EvokeBlink() error {
	if g.Player.HasStatus(StatusLignification) {
		return errors.New("You cannot blink while lignified.")
	}
	g.Blink()
	return nil
}

// Blink makes the player blink away.
func (g *Game) Blink() bool {
	if g.Player.HasStatus(StatusLignification) {
		g.Log("You resist blinking while lignified.")
		return false
	}
	npos := g.BlinkPos(false)
	if !InMap(npos) {
		// should not happen
		g.Log("You could not blink.")
		return false
	}
	opos := g.Player.P
	if npos == opos {
		g.Log("You blink in-place.")
	} else {
		g.Log("You blink away.")
	}
	g.md.TeleportAnimation(opos, npos, true)
	g.PlacePlayerAt(npos)
	return true
}

// BlinkPos returns a suitable blinking position (for monster or player).
func (g *Game) BlinkPos(mpassable bool) gruid.Point {
	losPos := []gruid.Point{}
	for p, b := range g.Player.AFOV {
		if !b {
			continue
		}
		t := g.Map.At(p)
		if !IsPlayerPassable(t) || mpassable && !Passable(t) || t == StoryCell {
			continue
		}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		losPos = append(losPos, p)
	}
	if len(losPos) == 0 {
		return invalidPos
	}
	q := losPos[RandInt(len(losPos))]
	for range 4 {
		p := losPos[RandInt(len(losPos))]
		if Distance(q, g.Player.P) < Distance(p, g.Player.P) {
			q = p
		}
	}
	return q
}

func (g *Game) EvokeTeleport() error {
	if g.Player.HasStatus(StatusLignification) {
		return errors.New("You cannot teleport while lignified.")
	}
	g.Teleportation()
	return nil
}

func (g *Game) EvokeDig() error {
	if !g.PutStatus(StatusDig, DurationDigging) {
		return errors.New("You are already digging.")
	}
	g.Log("You feel like an earth dragon.")
	g.md.PlayerGoodEffectAnimation()
	return nil
}

func (g *Game) EvokeTeleportOther() error {
	ms := g.MonstersInCardinalFOV()
	if len(ms) == 0 {
		return errors.New("There are no targetable monsters.")
	}
	done := false
	for _, mons := range ms {
		if mons.Search == invalidPos {
			mons.Search = mons.P
		}
		done = mons.TeleportAway(g) || done
		if mons.Kind.ReflectsTeleport() {
			g.Logf("The %s reflected back some energies.", mons.Kind)
			done = g.Teleportation() || done
			break
		}
	}
	if !done {
		return errors.New("There are no targetable monsters.")
	}
	return nil
}

// MonstersInCardinalFOV returns monsters in view in cardinal directions.
func (g *Game) MonstersInCardinalFOV() []*Monster {
	ms := []*Monster{}
	for _, mons := range g.Monsters {
		if mons.Exists() && g.Player.Sees(mons.P) && (mons.P.X == g.Player.P.X || mons.P.Y == g.Player.P.Y) {
			ms = append(ms, mons)
		}
	}
	// shuffle before, because the order could be unnaturally predicted
	for i := 0; i < len(ms); i++ {
		j := i + RandInt(len(ms)-i)
		ms[i], ms[j] = ms[j], ms[i]
	}
	return ms
}

func (g *Game) EvokeHealWounds() error {
	g.Player.HP = g.Player.HPMax()
	g.Log("Your feel healthy again.")
	g.md.PlayerGoodEffectAnimation()
	return nil
}

func (g *Game) EvokeRefillMagic() error {
	g.Player.MP = g.Player.MPMax()
	g.Log("Your magic forces return.")
	g.md.PlayerGoodEffectAnimation()
	return nil
}

func (g *Game) EvokeSwiftness() error {
	if g.Player.HasStatus(StatusSwift) {
		return errors.New("You are already swift.")
	}
	g.Player.Statuses[StatusSwift] += DurationSwiftness
	g.Logf("You feel swift.")
	g.md.PlayerGoodEffectAnimation()
	return nil
}

func (g *Game) EvokeLevitation() error {
	if !g.PutStatus(StatusLevitation, DurationLevitation) {
		return errors.New("You are already levitating.")
	}
	g.Logf("You feel light.")
	g.md.PlayerGoodEffectAnimation()
	return nil
}

func (g *Game) EvokeSwapping() error {
	if g.Player.HasStatus(StatusLignification) {
		return errors.New("You cannot use this magara while lignified.")
	}
	switch t := g.Map.At(g.Player.P); t {
	case BarrelCell, TableCell, HoledWallCell:
		return errors.New("You cannot use this magara in a cramped place.")
	}
	ms := g.MonstersInFOV()
	var mons *Monster
	best := 0
	for _, m := range ms {
		if m.Status(MonsLignified) {
			continue
		}
		if Distance(m.P, g.Player.P) > best {
			best = Distance(m.P, g.Player.P)
			mons = m
		}
	}
	if !mons.Exists() {
		return errors.New("No monsters suitable for swapping in view.")
	}
	if mons.Kind.CanOpenDoors() {
		// only intelligent monsters understand swapping
		mons.Search = mons.P
	}
	g.swapWithMonster(mons)
	return nil
}

func (g *Game) swapWithMonster(mons *Monster) {
	g.Logf("You swap positions with the %s.", mons.Kind)
	g.md.SwappingAnimation(mons.P, g.Player.P)
	g.PlacePlayerAt(mons.P)
	if g.Map.At(g.Player.P) == ChasmCell {
		g.PushEventFirst(AbyssFallEvent{}, g.Turn)
	}
}

// MonstersInFOV returns a slice with all visible monsters.
func (g *Game) MonstersInFOV() []*Monster {
	ms := []*Monster{}
	for _, mons := range g.Monsters {
		if mons.Exists() && g.Player.Sees(mons.P) {
			ms = append(ms, mons)
		}
	}
	// shuffle before, because the order could be unnaturally predicted
	for i := 0; i < len(ms); i++ {
		j := i + RandInt(len(ms)-i)
		ms[i], ms[j] = ms[j], ms[i]
	}
	return ms
}

// Cloud represents a cloud kind.
type Cloud int

const (
	CloudPlain Cloud = iota
	CloudFire
	CloudPurpleMist
)

func (cl Cloud) String() (s string) {
	switch cl {
	case CloudPlain:
		s = "cloud" // dense fog, dust
	case CloudFire:
		s = "fire"
	case CloudPurpleMist:
		s = "purple mist"
	}
	return s
}

func (g *Game) EvokeFog() error {
	g.Fog(g.Player.P, 3)
	g.Log("You are surrounded by a dense fog.")
	g.md.EffectAtPPAnimation(ColorForeground)
	return nil
}

// Fog generates fog around the given position.
func (g *Game) Fog(at gruid.Point, radius int) {
	mp := &NoisePath{g: g}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{at}, radius)
	for _, n := range nodes {
		_, ok := g.Map.Clouds[n.P]
		if !ok && AllowsFog(g.Map.At(n.P)) {
			g.Map.Clouds[n.P] = CloudPlain
			g.PushEvent(&PosEvent{P: n.P, Action: CloudEnd}, g.Turn+DurationFog+RandInt(DurationFog/2))
		}
	}
	g.ComputeFOV()
}

func (g *Game) EvokeShadows() error {
	if g.Player.HasStatus(StatusIlluminated) {
		return errors.New("You cannot surround yourself by shadows while illuminated.")
	}
	if !g.PutStatus(StatusShadows, DurationShadows) {
		return errors.New("You are already surrounded by shadows.")
	}
	g.Log("You are surrounded by shadows.")
	g.md.EffectAtPPAnimation(ColorForeground)
	return nil
}

func (g *Game) EvokeDazing() error {
	ms := g.MonstersInFOV()
	count := 0
	for _, mons := range ms {
		if mons.PutStatus(g, MonsDazed, DurationDazeMonster) {
			count++
			if mons.Search == invalidPos {
				mons.Search = mons.P
			}
		}
	}
	if count == 0 {
		return errors.New("No suitable targets in view.")
	}
	g.Log("Whoosh! A slowing luminous wave emerges.")
	g.md.FOVWavesAnimation(DefaultFOVRange, WaveDaze, g.Player.P)
	return nil
}

func (g *Game) EvokeSleeping() error {
	ms := g.MonstersInCardinalFOV()
	if len(ms) == 0 {
		return errors.New("There are no targetable monsters.")
	}
	targets := []gruid.Point{}
	for _, mons := range ms {
		if mons.State != Resting {
			g.Logf("%s falls asleep.", mons.Kind.Definite(true))
		} else {
			continue
		}
		mons.State = Resting
		mons.Dir = ZP
		mons.ExhaustTime(g, 4+RandInt(2))
		targets = append(targets, g.Ray(mons.P)...)
	}
	if len(targets) == 0 {
		return errors.New("There are no suitable targets.")
	}
	g.Log("Beams of sleeping emerge.")
	g.md.BeamsAnimation(targets, BeamSleeping)
	return nil
}

func (g *Game) EvokeLignification() error {
	ms := g.MonstersInFOV()
	if len(ms) == 0 {
		return errors.New("There are no monsters in view.")
	}
	targets := []gruid.Point{}
	for _, mons := range ms {
		if mons.Status(MonsLignified) || mons.Kind.ResistsLignification() {
			continue
		}
		mons.EnterLignification(g)
		if mons.Search == invalidPos {
			mons.Search = mons.P
		}
		targets = append(targets, g.Ray(mons.P)...)
	}
	if len(targets) == 0 {
		return errors.New("There are no suitable targets.")
	}
	g.Log("Beams of lignification emerge.")
	g.md.BeamsAnimation(targets, BeamLignification)
	return nil
}

func (g *Game) EvokeNoise() error {
	mp := &NoisePath{g: g}
	const noiseDist = 23
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.Player.P}, noiseDist)
	noises := []gruid.Point{}
	clear(g.Map.NoiseIllusion)
	for _, mons := range g.Monsters {
		if !mons.Exists() {
			continue
		}
		c := g.PR.BreadthFirstMapAt(mons.P)
		if c > DefaultFOVRange {
			continue
		}
		if mons.SeesPlayer(g) {
			continue
		}
		mp := &MonPath{g: g, monster: mons}
		target := mons.P
		best := c
		for {
			ncost := best
			for _, p := range mp.Neighbors(target) {
				c := g.PR.BreadthFirstMapAt(p)
				if c > noiseDist {
					continue
				}
				ncost := c
				if ncost > best {
					target = p
					best = ncost
				}
			}
			if ncost == best {
				break
			}
		}
		if mons.Kind == MonsSatowalgaPlant {
			mons.State = Hunting
		} else if mons.State != Hunting {
			mons.State = Wandering
		}
		mons.Target = target
		noises = append(noises, target)
		g.Map.NoiseIllusion[target] = 3
	}
	g.Log("Monsters are tricked by magical sounds.")
	g.md.NoiseAnimation(noises)
	return nil
}

func (g *Game) EvokeConfusion() error {
	ms := g.MonstersInFOV()
	count := 0
	for _, mons := range ms {
		if mons.PutConfusion(g) {
			count++
			if mons.Search == invalidPos {
				mons.Search = mons.P
			}
		}
	}
	if count == 0 {
		return errors.New("No suitable targets in view.")
	}
	g.Log("Whoosh! A confusing luminous wave emerges.")
	g.md.FOVWavesAnimation(DefaultFOVRange, WaveConfusion, g.Player.P)
	return nil
}

func (g *Game) EvokeFire() error {
	burnpos := g.FlammableNeighbors(g.Player.P)
	if len(burnpos) == 0 {
		return errors.New("You are not surrounded by any flammable terrain.")
	}
	g.Log("Sparks emanate from the magara.")
	g.md.EffectAtPPAnimation(ColorViolet)
	for _, p := range burnpos {
		g.Burn(p)
	}
	return nil
}

func (g *Game) EvokeObstruction() error {
	targets := []gruid.Point{}
	for _, mons := range g.Monsters {
		if !mons.Exists() || !g.Player.Sees(mons.P) {
			continue
		}
		ray := g.Ray(mons.P)
		for i, p := range ray[1:] {
			if p == g.Player.P {
				break
			}
			mons := g.MonsterAt(p)
			if mons.Exists() {
				continue
			}
			g.MagicalBarrierAt(p)
			if len(ray) == 0 {
				break
			}
			ray = ray[i+1:]
			targets = append(targets, ray...)
			break
		}
	}
	if len(targets) == 0 {
		return errors.New("No targetable monsters in view.")
	}
	g.Log("Magical barriers emerged.")
	g.md.BeamsAnimation(targets, BeamObstruction)
	return nil
}

// MagicalBarrierAt puts a temporary magical barrier at the given position (if
// possible).
func (g *Game) MagicalBarrierAt(p gruid.Point) {
	t := g.Map.At(p)
	switch t {
	case WallCell, WindowCell, BarrierCell:
		return
	}
	g.Map.Set(p, BarrierCell)
	delete(g.Map.Clouds, p)
	g.Map.MagicalBarriers[p] = t
	g.PushEvent(&PosEvent{P: p, Action: ObstructionEnd}, g.Turn+DurationMagicalBarrier+RandInt(DurationMagicalBarrier/2))
	g.ComputeFOV()
}

func (g *Game) EvokeEnergyMagara() error {
	if g.Player.MP == g.Player.MPMax() && g.Player.HP == g.Player.HPMax() {
		return errors.New("You are already full of energy.")
	}
	g.Log("The magara glows.")
	g.md.PlayerGoodEffectAnimation()
	g.Player.MP = g.Player.MPMax()
	g.Player.HP = g.Player.HPMax()
	return nil
}

func (g *Game) EvokeTransparencyMagara() error {
	if !g.PutStatus(StatusTransparent, DurationTransparency) {
		return errors.New("You are already transparent.")
	}
	g.Log("Light makes you diaphanous.")
	g.md.EffectAtPPAnimation(ColorYellow)
	return nil
}

func (g *Game) EvokeDisguiseMagara() error {
	if !g.PutStatus(StatusDisguised, DurationDisguise) {
		return errors.New("You are already disguised.")
	}
	g.Log("You look now like a normal guard.")
	g.md.EffectAtPPAnimation(ColorViolet)
	return nil
}

func (g *Game) EvokeDelayedNoiseMagara() error {
	if !g.PutFakeStatus(StatusDelay, DurationHarmonicNoiseDelay) {
		return errors.New("You are already using delayed magic.")
	}
	g.PushEventD(&PosEvent{P: g.Player.P, Action: DelayedHarmonicNoiseEvent,
		Timer: DurationHarmonicNoiseDelay}, DurationTurn)
	g.Log("Timer activated.")
	g.md.EffectAtPPAnimation(ColorMagenta)
	return nil
}

func (g *Game) EvokeDispersalMagara() error {
	if !g.PutStatus(StatusDispersal, DurationDispersal) {
		return errors.New("You are already dispersing.")
	}
	g.Log("You feel unstable.")
	g.md.EffectAtPPAnimation(ColorCyan)
	return nil
}

func (g *Game) EvokeOricExplosionMagara() error {
	if !g.PutFakeStatus(StatusDelay, DurationHarmonicNoiseDelay) {
		return errors.New("You are already using delayed magic.")
	}
	g.PushEventD(&PosEvent{P: g.Player.P, Action: DelayedOricExplosionEvent,
		Timer: DurationOricExplosionDelay}, DurationTurn)
	g.Log("Timer activated.")
	g.md.EffectAtPPAnimation(ColorMagenta)
	return nil
}
