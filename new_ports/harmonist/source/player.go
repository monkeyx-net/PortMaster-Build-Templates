package main

import (
	"errors"
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// Player represents the player.
type Player struct {
	HP        int                  // current HP
	HPbonus   int                  // temporary HP bonus (lignification)
	MP        int                  // current MP
	Bananas   int                  // number of carried bananas
	Magaras   []Magara             // equipped magaras
	Statuses  map[status]int       // active status effect durations
	Expire    map[status]int       // turn of expiration for active status effects
	P         gruid.Point          // current position
	AFOV      map[gruid.Point]bool // actual field of view (accounting for darkness)
	FOV       *rl.FOV              // field of view (full range)
	Inventory Inventory            // inventory (cloak, amulet, artifact)
}

// Inventory represents the inventory.
type Inventory struct {
	Body Item // cloak
	Neck Item // amulet
	Misc Item // artifact
}

// DefaultHealth represents the starting max HP.
const DefaultHealth = 4

func (pl *Player) HPMax() int {
	hpmax := DefaultHealth
	if pl.Inventory.Body.Kind == CloakShadows {
		hpmax++
	}
	if hpmax < 2 {
		hpmax = 2
	}
	return hpmax
}

// MPMax returns the maximum MP. It's the number of magaras increased by one,
// so that it increases by one once you liberate Shaedra.
func (pl *Player) MPMax() int {
	return len(pl.Magaras) + 1
}

// HasStatus reports whether the player has the given status effect.
func (pl *Player) HasStatus(st status) bool {
	return pl.Statuses[st] > 0
}

// AutoTodir updates auto-movement in direction and reports whether it
// continues.
func (g *Game) AutoToDir() bool {
	if g.MonsterInFOV() != nil {
		g.autoDir = ZP
		return false
	}
	p := g.Player.P.Add(g.autoDir)
	nbs := g.dirNeighbors(p, g.autoDir)
	blocked := g.dirBlocked(p, g.autoDir)
	if g.PlayerCanPass(p) && (g.autoDirNeighbors == nbs || nbs != dirFreeLaterals && blocked ||
		nbs == dirBlockedLaterals && (g.autoDirChanged || g.dirNeighbors(g.Player.P, g.autoDir) == dirBlockedLaterals)) {
		again, err := g.PlayerBump(p)
		if err != nil {
			g.Log(err.Error())
			g.autoDir = ZP
			return false
		}
		if again {
			g.autoDir = ZP
			return false
		}
		// player moved in the direction
		g.autoDirChanged = false
		if blocked && nbs > 0 {
			if g.PlayerCanPass(Left(p, g.autoDir)) {
				g.autoDir = Left(p, g.autoDir).Sub(p)
				g.autoDirChanged = true
			} else if g.PlayerCanPass(Right(p, g.autoDir)) {
				g.autoDir = Right(p, g.autoDir).Sub(p)
				g.autoDirChanged = true
			}
			g.autoDirNeighbors = g.dirNeighbors(p, g.autoDir)
		}
		return true
	}
	g.autoDir = ZP
	return false
}

// dirNeighbors represents the various kinds of neighbor-configurations (for
// auto-running purposes).
type dirNeighbors int

const (
	dirFreeLaterals dirNeighbors = iota
	dirBlockedLeft
	dirBlockedRight
	dirBlockedLaterals
)

func (g *Game) dirNeighbors(p, dir gruid.Point) (dn dirNeighbors) {
	if !g.PlayerCanPass(Left(p, dir)) {
		dn += dirBlockedLeft
	}
	if !g.PlayerCanPass(Right(p, dir)) {
		dn += dirBlockedRight
	}
	return dn
}

func (g *Game) dirBlocked(p, dir gruid.Point) bool {
	return !g.PlayerCanPass(p.Add(dir))
}

// GoToDir implements auto-move in direction.
func (g *Game) GoToDir(dir gruid.Point) (again bool, err error) {
	if g.MonsterInFOV() != nil {
		g.autoDir = ZP
		return again, errors.New("You cannot travel while there are monsters in view.")
	}
	p := g.Player.P.Add(dir)
	if !g.PlayerCanPass(p) {
		return again, errors.New("You cannot move in that direction.")
	}
	again, err = g.PlayerBump(p)
	if err != nil || again {
		return again, err
	}
	g.autoDir = dir
	nbs := g.dirNeighbors(p, dir)
	g.autoDirChanged = false
	blocked := g.dirBlocked(p, dir)
	if blocked && nbs > 0 {
		if g.PlayerCanPass(Left(p, dir)) {
			g.autoDir = Left(p, dir).Sub(p)
			g.autoDirChanged = true
		} else if g.PlayerCanPass(Right(p, dir)) {
			g.autoDir = Right(p, dir).Sub(p)
			g.autoDirChanged = true
		}
		g.autoDirNeighbors = g.dirNeighbors(p, g.autoDir)
	} else {
		g.autoDirNeighbors = nbs
	}
	return again, err
}

// MoveToTarget makes the player continue moving to the next point in the path
// toward the target.
func (g *Game) MoveToTarget() bool {
	if !InMap(g.autoTarget) {
		return false
	}
	path := g.PlayerPath(g.Player.P, g.autoTarget)
	if g.MonsterInFOV() != nil {
		g.autoTarget = invalidPos
	}
	if len(path) < 1 {
		g.autoTarget = invalidPos
		return false
	}
	var err error
	var again bool
	if len(path) > 1 {
		again, err = g.PlayerBump(path[1])
	} else {
		g.WaitTurn()
	}
	if err != nil {
		g.Log(err.Error())
		g.autoTarget = invalidPos
		return false
	}
	if InMap(g.autoTarget) && g.Player.P == g.autoTarget {
		g.autoTarget = invalidPos
	}
	return !again
}

// WaitTurn is called when the player waits a turn.
func (g *Game) WaitTurn() {
	g.Stats.Waits++
	if t := g.Map.At(g.Player.P); t == PotionCell {
		g.DrinkPotion(g.Player.P)
	}
}

// Monstercount returns the number of (alive) monsters.
func (g *Game) MonsterCount() (count int) {
	for _, mons := range g.Monsters {
		if mons.Exists() {
			count++
		}
	}
	return count
}

// Rest makes the player eat a banana and rest.
func (g *Game) Rest() error {
	if g.Map.At(g.Player.P) != BarrelCell {
		return fmt.Errorf("This place is not safe for sleeping.")
	}
	if cld, ok := g.Map.Clouds[g.Player.P]; ok && cld == CloudFire {
		return errors.New("You cannot rest on flames.")
	}
	if !g.NeedsRegenRest() && !g.StatusRest() {
		return errors.New("You do not need to rest.")
	}
	if g.Player.Bananas <= 0 {
		return errors.New("You cannot sleep without eating for dinner a banana first.")
	}
	g.resting = true
	g.restingTurns = RandInt(5) // you do not wake up when you want
	g.Player.Bananas--
	return nil
}

// StatusRest reports whether some non-info status effect is still active.
func (g *Game) StatusRest() bool {
	for st, q := range g.Player.Statuses {
		if st.Info() {
			continue
		}
		if q > 0 {
			return true
		}
	}
	return false
}

// NeedsRegenRest reports whether the player's HP or MP are not maxed out.
func (g *Game) NeedsRegenRest() bool {
	return g.Player.HP < g.Player.HPMax() || g.Player.MP < g.Player.MPMax()
}

// Teleportation makes the player teleport away.
func (g *Game) Teleportation() bool {
	if g.Player.HasStatus(StatusLignification) {
		g.LogStyled("You resist teleportation while lignified.", logNotable)
		return false
	}
	p := g.TeleportPos()
	if InMap(p) {
		// should always happen
		opos := g.Player.P
		g.Log("You teleport away.")
		g.md.TeleportAnimation(opos, p, true)
		g.PlacePlayerAt(p)
	} else {
		// should not happen
		g.Log("Something went wrong with the teleportation.")
	}
	return true
}

// TeleportPos returns a position suitable for teleportation.
func (g *Game) TeleportPos() gruid.Point {
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("Teleportation")
		}
		p := g.FreePassableCell()
		if Distance(p, g.Player.P) < 15 && count < maxIterations/2 {
			continue
		}
		return p

	}
}

// MaxBananas represents the maximum number of bananas the player can carry.
const MaxBananas = 4

// Collectground makes the player automatically collect some items from the
// ground without spending an extra turn, and handles terrain-related
// achievements.
func (g *Game) CollectGround() {
	p := g.Player.P
	t := g.Map.At(p)
	if IsNotable(t) {
		g.autoexploreMapRebuild = true
	switchcell:
		switch t {
		case BarrelCell:
			g.Log("You hide yourself inside the barrel.")
		case BananaCell:
			if g.Player.Bananas >= MaxBananas {
				g.Log("There is a banana, but your pack is already full.")
			} else {
				g.Log("You take a banana.")
				g.Player.Bananas++
				g.StoryPrintf("Found banana (bananas: %d)", g.Player.Bananas)
				g.Map.Set(p, GroundCell)
				delete(g.Objects.Bananas, p)
				if g.Player.Bananas == MaxBananas {
					AchBananaCollector.Get(g)
				}
			}
		case MagaraCell:
			for i, mag := range g.Player.Magaras {
				if mag.Kind != NoMagara {
					continue
				}
				g.Player.Magaras[i] = g.Objects.Magaras[p]
				g.Player.Magaras[i].Depth = g.Map.Depth
				delete(g.Objects.Magaras, p)
				g.Map.Set(p, GroundCell)
				g.Logf("You take the %s.", g.Player.Magaras[i])
				g.StoryPrintf("Took %s", g.Player.Magaras[i])
				break switchcell
			}
			g.Logf("You stand over %s.", Indefinite(g.Objects.Magaras[p].String(), false))
		case StairCell:
			if g.Objects.Stairs[p] == FakeStair {
				g.Objects.Stairs[p] = KnownFakeStair
				g.LogStyled("You stand over fake stairs: harmonic illusions!", logSpecial)
				g.StoryPrint("Found harmonic fake stairs!")
				g.md.FoundFakeStairsAnimation()
			}
		case PotionCell:
			g.DrinkPotion(p)
		default:
			g.Logf("You are standing over %s.", g.TerrainShortDesc(t, p))
		}
	} else if t == DoorCell {
		g.Log("You stand at the door.")
	}
	if AchivementTerrain(t) {
		g.Reach(p)
	}
}

// FallAbyss makes the player fall into the abyss (either due to voluntary jump
// or unvoluntary fall).
func (g *Game) FallAbyss(style descendstyle) {
	if g.Player.HasStatus(StatusLevitation) {
		return
	}
	g.Player.HP -= 2
	if g.Player.HP <= 0 {
		g.Player.HP = 1
	}
	g.Player.MP -= 2
	if g.Player.MP < 0 {
		g.Player.MP = 0
	}
	if g.Player.Bananas > 0 {
		g.Player.Bananas--
	}
	if style == DescendFall && (g.Map.Depth == MaxDepth || g.Map.Depth == WinDepth) {
		g.Player.HP = 0
		return
	}
	g.Descend(style)
}

// AbyssJump asks for confirmation before jumping into the abyss (if it's not
// deep chasm).
func (g *Game) AbyssJump() error {
	if g.DeepChasmDepth() {
		return errors.New("You cannot jump into deep chasm.")
	}
	g.LogStyled("Do you really want to jump into the abyss? (DANGEROUS) [y/N]", logConfirm)
	g.md.mode = modeJumpConfirmation
	return nil
}

// DeepChasmDepth reports whether we're on Shaedra's or the Artifact's level
// (which may feature deep deadly chasm instead of regular chasm).
func (g *Game) DeepChasmDepth() bool {
	return g.Map.Depth == WinDepth || g.Map.Depth == MaxDepth
}

// PlayerBump handles player "bump" movement.
func (g *Game) PlayerBump(p gruid.Point) (again bool, err error) {
	if !InMap(p) {
		return g.WallJump(p)
	}
	t := g.Map.At(p)
	switch t {
	case BarrierCell:
		if !g.Player.HasStatus(StatusLevitation) {
			return g.WallJump(p)
		}
	case WallCell, WindowCell:
		if !g.Player.HasStatus(StatusDig) {
			return g.WallJump(p)
		}
	case BarrelCell:
		if g.MonsterFOV[g.Player.P] {
			return again, errors.New("You cannot enter a barrel while unhidden.")
		}
	}
	mons := g.MonsterAt(p)
	if !mons.Exists() {
		if g.Player.HasStatus(StatusLignification) {
			return again, errors.New("You cannot move while lignified.")
		}
		if t == ChasmCell && !g.Player.HasStatus(StatusLevitation) {
			again = true
			return again, g.AbyssJump()
		}
		switch t {
		case TableCell:
			g.Log("You hide yourself under the table.")
		case TreeCell:
			g.Log("You climb to the top.")
		case HoledWallCell:
			g.Log("You crawl under the wall.")
		}
		if IsDiggable(t) && t != HoledWallCell {
			g.Map.Set(p, RubbleCell)
			g.MakeNoise(WallNoise, p)
			g.LogStyled(g.CrackSound(), logNotable)
			g.Fog(p, 1)
			g.NewDigStats()
		}
		if g.Player.Inventory.Body.Kind == CloakSmoke {
			_, ok := g.Map.Clouds[g.Player.P]
			if !ok && AllowsFog(g.Map.At(g.Player.P)) {
				g.Map.Clouds[g.Player.P] = CloudPlain
				g.PushEventD(&PosEvent{P: g.Player.P, Action: CloudEnd}, DurationFogSmokingCloak)
			}
		}
		g.Stats.Moves++
		g.PlacePlayerAt(p)
	} else if again, err = g.Jump(mons); err != nil {
		return again, err
	}
	if g.Player.HasStatus(StatusSwift) {
		again = true
		g.Player.Statuses[StatusSwift]--
		if !g.Player.HasStatus(StatusSwift) {
			g.Log("You no longer feel swift.")
		}
		g.md.updateMapInfo()
		// We need to check for Shaedra's liberation at each swift
		// step.
		g.CheckForStorySequence()
		return again, nil
	}
	return again, nil
}

// ActivateAmuletFog implements the amulet of fog's effect.
func (g *Game) ActivateAmuletFog() {
	mp := &NoisePath{g: g}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.Player.P}, 2)
	for _, n := range nodes {
		_, ok := g.Map.Clouds[n.P]
		if !ok && AllowsFog(g.Map.At(n.P)) {
			g.Map.Clouds[n.P] = CloudPlain
			g.PushEvent(&PosEvent{P: n.P, Action: CloudEnd}, g.Turn+DurationFog+RandInt(DurationFog/2))
		}
	}
	g.PutStatus(StatusSwift, DurationSwiftnessAmulet)
	g.ComputeFOV()
	g.Log("You feel an energy burst and smoke comes out from you.")
}

// PutConfusion makes the player confused.
func (g *Game) PutConfusion() {
	if g.PutStatus(StatusConfusion, DurationConfusionPlayer) {
		g.Log("You feel confused.")
	}
}

// PlacePlayerAt moves the player to a given position, swapping positions with
// a monster if necessary, and handles FOV and monsters awareness update,
// ground collecting, and footsteps noise.
func (g *Game) PlacePlayerAt(p gruid.Point) {
	if p == g.Player.P {
		return
	}
	m := g.MonsterAt(p)
	ppos := g.Player.P
	g.Player.P = p
	if m.Exists() {
		if !m.MoveTo(g, ppos) {
			// Ensure swapping even in unlikely edge-cases.
			m.PlaceAt(g, ppos)
		}
		m.Swapped = true
	}
	if g.Map.At(g.Player.P) == QueenRockCell && !g.Player.HasStatus(StatusLevitation) {
		g.MakeNoise(QueenRockFootstepNoise, g.Player.P)
		g.LogStyled("Tap-Tap.", logNotable)
	}
	g.CollectGround()
	g.ComputeFOV()
}

// LignificationHPbonus represents the amount of extra temporary HP granted by
// lignification for the player.
const LignificationHPbonus = 4

// EnterLignification lignifies the player.
func (g *Game) EnterLignification() {
	if g.PutStatus(StatusLignification, DurationLignificationPlayer) {
		g.Log("You feel rooted to the ground.")
		g.Player.HPbonus += LignificationHPbonus
	}
}

// ExtinguishFire extinguishes the campfire at the player's position (assuming
// there's one).
func (g *Game) ExtinguishFire() error {
	g.Map.Set(g.Player.P, ExtinguishedLightCell)
	g.Objects.Lights[g.Player.P] = false
	g.Stats.Extinguishments++
	if g.Stats.Extinguishments >= 15 {
		AchExtinguisher.Get(g)
	}
	g.Log("You extinguish the fire.")
	return nil
}

// PutStatus gives a status effect to the player with given duration if it
// isn't already active.
func (g *Game) PutStatus(st status, duration int) bool {
	if g.Player.Statuses[st] != 0 {
		return false
	}
	g.Player.Statuses[st] += duration
	g.PushEventD(&StatusEvent{Status: st}, DurationStatusStep)
	g.Stats.Statuses[st]++
	if st.Good() {
		g.Player.Expire[st] = g.Turn + duration
	}
	return true
}

// PutFakeStatus is like PutStatus but doesn't produce a StatusEvent.
func (g *Game) PutFakeStatus(st status, duration int) bool {
	if g.Player.Statuses[st] != 0 {
		return false
	}
	g.Player.Statuses[st] += duration
	g.Stats.Statuses[st]++
	if st.Good() {
		g.Player.Expire[st] = g.Turn + duration
	}
	return true
}

// PlayerCanPass reports whether the player can pass through the given map
// point, possibly levitating or digging.
func (g *Game) PlayerCanPass(p gruid.Point) bool {
	if !InMap(p) {
		return false
	}
	t := g.Map.At(p)
	return IsPlayerPassable(t) ||
		g.Player.HasStatus(StatusLevitation) && (t == BarrierCell || IsLevitatePassable(t)) ||
		g.Player.HasStatus(StatusDig) && IsDiggable(t)
}

// PlayerCanJumpTo reports whether the player can jump the given map point,
// possibly levitating or digging.
func (g *Game) PlayerCanJumpTo(p gruid.Point) bool {
	if g.Map.At(p) == BarrelCell {
		return false
	}
	return g.PlayerCanPass(p)
}

// PlayerCanJumpPass reports whether the player can pass through the given map
// point, possibly jumping or levitating.
func (g *Game) PlayerCanJumpPass(p gruid.Point) bool {
	if !InMap(p) {
		return false
	}
	t := g.Map.At(p)
	return IsJumpPassable(t) ||
		t == BarrierCell && g.Player.HasStatus(StatusLevitation) ||
		t == DoorCell && (g.MonsterAt(p).Exists() || g.Player.P == p)
}
