// This file defines functions for basic handling of entities.

package main

import (
	"fmt"
	"iter"

	"codeberg.org/anaseto/gruid"
)

// Entity represents various kinds of map or inventory entities. It gathers
// common components in various fields and uses the Role field for any extra
// ones.
type Entity struct {
	Name   string      // name
	Rune   rune        // default rune (for display)
	P      gruid.Point // map position (invalidPos if none)
	KnownP gruid.Point // last known position
	Seen   bool        // whether we already saw it once
	Noise  bool        // whether it made noise reaching player (monsters only currently)
	Role   any         // non-common role(s) for the entity
}

// String returns the name of the entity.
func (e *Entity) String() string {
	return e.Name
}

// Color returns the color of an entity.
func (e *Entity) Color() (fg gruid.Color) {
	switch r := e.Role.(type) {
	case *Actor:
		if r.IsPlayer() {
			return ColorBlue
		}
		fg = ColorOrange
		if r.Behavior.State == Hunting {
			fg = ColorRed
		}
		bad, good := 0, 0
		for i, turns := range r.Statuses {
			if turns >= 1 {
				if Status(i).Bad() {
					bad++
				} else {
					good++
				}
			}
		}
		switch {
		case bad == 0 && good > 0:
			fg = ColorMagenta
		case bad > 0 && good > 0:
			fg = ColorYellow
		case bad > 0 && good == 0:
			fg = ColorGreen
			if bad > 1 {
				fg = ColorCyan
			}
		}
	case *Menhir:
		if r.Used {
			fg = ColorForeground
		} else {
			fg = ColorCyan
		}
	case *Portal:
		fg = ColorCyan
		if r.Used {
			fg = ColorForeground
		}
	case *CorruptionOrb:
		fg = ColorCyan
	case *Spirit:
		fg = ColorYellow
	case *Comestible:
		fg = ColorYellow
	case *RunicTrap:
		fg = r.Color()
	}
	return fg
}

// IsActor reports whether e has an Actor component.
func (e *Entity) IsActor() bool {
	_, ok := e.Role.(*Actor)
	return ok
}

// Actor returns the underlying role actor, assuming IsActor.
func (e *Entity) Actor() *Actor {
	return e.Role.(*Actor)
}

// IsAlive reports whether e has an Actor component, and is alive.
func (e *Entity) IsAlive() bool {
	a, ok := e.Role.(*Actor)
	return ok && a.HP > 0
}

// IsItem reports whether e has an Item component.
func (e *Entity) IsItem() bool {
	_, ok := e.Role.(Item)
	return ok
}

// Item returns the underlying Item role, assuming IsItem.
func (e *Entity) Item() Item {
	return e.Role.(Item)
}

// Text returns short text for the entity, for one-line display purposes.
func (e *Entity) Text() string {
	switch r := e.Role.(type) {
	case *Spirit:
		return fmt.Sprintf("%s (%s %d/%d)", e.Name, r.Ability[r.Level].Name(), r.Charges, r.MaxCharges[r.Level])
	default:
		return e.Name
	}
}

// MoveEntity moves an entity to the given position.
func (g *Game) MoveEntity(i ID, to gruid.Point) {
	ei := g.Entity(i)
	if ai, ok := ei.Role.(*Actor); ok {
		g.MoveActor(i, ai, to)
	} else {
		ei.P = to
	}
}

// MoveActor moves an actor to the given position. It updates the actor
// position cache as necessary. The movement isn't considered as first
// bump-movement: it may be a teleport or an extra movement, so extra effects
// aren't applied.
func (g *Game) MoveActor(i ID, ai *Actor, to gruid.Point) {
	g.moveActor(i, ai, to, false)
}

// moveActor moves an actor to the given position. It updates the actor
// position cache as necessary. It may apply extra bump-movement effects if
// bumpmove is true.
func (g *Game) moveActor(i ID, ai *Actor, to gruid.Point, bumpmove bool) {
	ei := g.Entity(i)
	if ei.P == to {
		return
	}
	g.Map.ActorCache.SetU(ei.P, 0)
	if j, _ := g.ActorAt(to); j >= 0 {
		// Swap actor positions.
		ej := g.Entity(j)
		ej.P = ei.P
		g.Map.ActorCache.SetU(ei.P, j)
		if j == PlayerID {
			ej.KnownP = ej.P
		}
	}
	ei.P = to
	g.Map.ActorCache.SetU(to, i)
	switch {
	case ai.IsPlayer():
		ei.KnownP = ei.P
		if i, it := g.ItemAt(to); i >= 0 {
			ei := g.Entity(i)
			if _, ok := it.(*Spirit); ok {
				g.Logf("You step over the %s totem.", ei.Name)
			} else {
				g.Logf("You step over the %s.", ei.Name)
			}
		}
	default:
		if len(ai.Behavior.Path) >= 2 && ai.Behavior.Path[1] == to {
			// Keep using computed path.
			ai.Behavior.Path = ai.Behavior.Path[1:]
		}
	}
	if i != PlayerID && g.InFOV(ei.P) {
		g.SenseEntity(i, "see")
	}
	if bumpmove {
		g.ApplyBumpMoveEffects(i, ai)
	}
	if i != PlayerID || !ai.DoesAny(RunicChicken) {
		g.TriggerTrapAt(to)
	}
}

// BumpMoveActor is like MoveActor but with extra effects specific to
// bump-movement and some other non-teleport kind of movements.
func (g *Game) BumpMoveActor(i ID, ai *Actor, to gruid.Point) {
	if ai.IsDead() {
		return
	}
	p := g.Entity(i).P
	if ai.DoesAny(TrailingCloud) && g.Map.Terrain.At(p) == Floor {
		g.NormalCloudAt(p, 2)
	}
	if ai.DoesAny(MonsDig) {
		// Monsters that dig may attempt moving into walls, so they
		// need to dig.
		if !g.Map.Passable(to) {
			g.monsConfusionPenalty(i, ai)
		}
		g.digAt(to)
	}
	g.moveActor(i, ai, to, true)
}

// ApplyBumpMoveEffects applies effects that happen after moving to a new
// position through normal movement (but before triggering any traps).
func (g *Game) ApplyBumpMoveEffects(i ID, ai *Actor) {
	switch {
	case ai.Has(StatusPoison):
		g.InflictDamage(i, ai, PoisonDamage, AttackPoison)
	case ai.Has(StatusLignification):
		// Actors with the Walking Tree spirit can move while
		// lignified, but the status duration is then reduced.
		g.ProgressStatus(i, ai, StatusLignification)
	}
}

// Entity returns the common components for an entity with a given
// identifier.
func (g *Game) Entity(i ID) *Entity {
	return g.Entities[i]
}

// AddEntity adds a new entity at a given position and returns its index/id.
func (g *Game) AddEntity(e *Entity) ID {
	g.Entities = append(g.Entities, e)
	return ID(len(g.Entities) - 1)
}

// CleanEntities removes all the entities that are specific to the current map.
// It keeps the player and the items in the inventory.
func (g *Game) CleanEntities() {
	g.Entities = g.Entities[:FirstMapID+1]
}

// SwapEntities exchanges the identifiers of two entities.
func (g *Game) SwapEntities(i, j ID) {
	g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
}

// Player returns player's entity and actor.
func (g *Game) Player() (*Entity, *Actor) {
	pl := g.Entities[PlayerID]
	return pl, pl.Role.(*Actor)
}

// PlayerEntity returns the player entity.
func (g *Game) PlayerEntity() *Entity {
	return g.Entities[PlayerID]
}

// PlayerActor returns player's actor role.
func (g *Game) PlayerActor() *Actor {
	return g.Entities[PlayerID].Role.(*Actor)
}

// PP is a shortcut function that returns the player's position.
func (g *Game) PP() gruid.Point {
	return g.PlayerEntity().P
}

// ActorAt returns the id of the Actor alive at p, if any. It returns (-1, nil)
// if there was none.
func (g *Game) ActorAt(p gruid.Point) (ID, *Actor) {
	if i := g.Map.ActorCache.At(p); i > 0 {
		if a, ok := g.Entity(i).Role.(*Actor); ok && a.IsAlive() {
			return i, a
		}
	}
	return -1, nil
}

// IsFree returns whether the given position is both passable and not occupied
// by any actor.
func (g *Game) IsFree(p gruid.Point) bool {
	if !g.Map.Passable(p) {
		return false
	}
	i, _ := g.ActorAt(p)
	return i < 0
}

// KnownActorAt returns the id of any seen Actor at p. It returns (-1, nil) if
// there was none.
func (g *Game) KnownActorAt(p gruid.Point) (ID, *Actor) {
	for i, e := range g.Entities[FirstMapID:] {
		if e.KnownP != p {
			continue
		}
		if a, ok := e.Role.(*Actor); ok && !a.KnownDead {
			return ID(i) + FirstMapID, a
		}
	}
	return -1, nil
}

// SensedActorAt returns the id of any just seen or heard alive Actor at p. It
// returns (-1, nil) if there was none.
func (g *Game) SensedActorAt(p gruid.Point) (ID, *Actor) {
	for i, e := range g.Entities[FirstMapID:] {
		if !(e.KnownP == p || e.P == p && e.Noise && !g.InFOV(p)) {
			continue
		}
		if a, ok := e.Role.(*Actor); ok && !a.KnownDead {
			return ID(i) + FirstMapID, a
		}
	}
	return -1, nil
}

// ItemAt returns the id of the Item at p, if any. It returns (-1, nil) if
// there is none.
func (g *Game) ItemAt(p gruid.Point) (ID, Item) {
	for i, e := range g.Entities[FirstMapID+1:] {
		if e.P != p {
			continue
		}
		if it, ok := e.Role.(Item); ok {
			return ID(i) + FirstMapID + 1, it
		}
	}
	return -1, nil
}

// KnownItemAt returns the id of any seen Item at p. It returns (-1, nil) if
// there is none.
func (g *Game) KnownItemAt(p gruid.Point) (ID, Item) {
	for i, e := range g.Entities[FirstMapID+1:] {
		if e.KnownP != p {
			continue
		}
		if it, ok := e.Role.(Item); ok {
			return ID(i) + FirstMapID + 1, it
		}
	}
	return -1, nil
}

// Actors returns an iterator over alive actor map entities.
func (g *Game) Actors() iter.Seq2[ID, *Actor] {
	return func(yield func(ID, *Actor) bool) {
		for i, e := range g.Entities[FirstMapID:] {
			if a, ok := e.Role.(*Actor); ok && a.IsAlive() {
				if !yield(ID(i)+FirstMapID, a) {
					return
				}
			}
		}
	}
}

// AllActorIDs returns an iterator over actor map IDs, including dead ones.
func (g *Game) AllActorIDs() []ID {
	var ids []ID
	for id, e := range g.MapEntities() {
		if e.IsActor() {
			ids = append(ids, id)
		}
	}
	return ids
}

// AdjacentActors returns an iterator over alive actor map entities adjacent to
// the given position.
func (g *Game) AdjacentActors(p gruid.Point) iter.Seq2[ID, *Actor] {
	return func(yield func(ID, *Actor) bool) {
		for q := range Neighbors(p) {
			if i, ai := g.ActorAt(q); i >= 0 {
				if !yield(i, ai) {
					return
				}
			}
		}
	}
}

// NPMapEntities returns an iterator over all non-player map entities.
func (g *Game) NPMapEntities() iter.Seq2[ID, *Entity] {
	return func(yield func(ID, *Entity) bool) {
		for i, e := range g.Entities[FirstMapID+1:] {
			if !yield(ID(i)+FirstMapID+1, e) {
				return
			}
		}
	}
}

// MapEntities returns an iterator over all map entities, including the player.
func (g *Game) MapEntities() iter.Seq2[ID, *Entity] {
	return func(yield func(ID, *Entity) bool) {
		for i, e := range g.Entities[FirstMapID:] {
			if !yield(ID(i)+FirstMapID, e) {
				return
			}
		}
	}
}

// Monsters returns an iterator over alive non-player actor map entities.
func (g *Game) Monsters() iter.Seq2[ID, *Actor] {
	return func(yield func(ID, *Actor) bool) {
		for id, e := range g.NPMapEntities() {
			if a, ok := e.Role.(*Actor); ok && a.IsAlive() {
				if !yield(id, a) {
					return
				}
			}
		}
	}
}

// ItemEntities returns an iterator over item map entities.
func (g *Game) ItemEntities() iter.Seq2[ID, Item] {
	return func(yield func(ID, Item) bool) {
		for id, e := range g.NPMapEntities() {
			if it, ok := e.Role.(Item); ok {
				if !yield(id, it) {
					return
				}
			}
		}
	}
}

// Comestibles returns an iterator over comestible map entities.
func (g *Game) Comestibles() iter.Seq2[ID, Item] {
	return func(yield func(ID, Item) bool) {
		for id, e := range g.NPMapEntities() {
			if it, ok := e.Role.(*Comestible); ok {
				if !yield(id, it) {
					return
				}
			}
		}
	}
}

// ComestibleIDs returns a slice with all the comestible IDs that can be eaten,
// including any comestible on the ground.
func (g *Game) ComestibleIDs() []ID {
	var ids []ID
	for id := range g.PlayerComestibles() {
		ids = append(ids, id)
	}
	if id, it := g.ItemAt(g.PP()); id >= 0 {
		if IsComestible(it) {
			ids = append(ids, id)
		}
	}
	return ids
}

// TotemID returns the ID of the map's totem (if any). NOTE: there should never
// be more than one, otherwise this function returns the first one.
func (g *Game) TotemID() ID {
	for id, e := range g.NPMapEntities() {
		switch e.Role.(type) {
		case *Spirit:
			return id
		case *EmptyTotem:
			return id
		}
	}
	return -1
}

// PlayerSpirits returns an iterator over player's primary and secondary spirits.
func (g *Game) PlayerSpirits() iter.Seq2[ID, *Spirit] {
	return func(yield func(ID, *Spirit) bool) {
		for i := PrimarySpiritID; i < PrimarySpiritID+NSpirits; i++ {
			if sp, ok := g.Entity(ID(i)).Role.(*Spirit); ok {
				if !yield(ID(i), sp) {
					return
				}
			}
		}
	}
}

// PlayerComestibles returns an iterator over player's comestibles.
func (g *Game) PlayerComestibles() iter.Seq2[ID, *Comestible] {
	to := g.InventoryEnd()
	return func(yield func(ID, *Comestible) bool) {
		for i := PrimarySpiritID + NSpirits; i < to; i++ {
			if sp, ok := g.Entity(i).Role.(*Comestible); ok {
				if !yield(i, sp) {
					return
				}
			}
		}
	}
}

// InventoryEnd returns the first non-inventory ID. May be lower than
// FirstMapID with ModSmallInventory.
func (g *Game) InventoryEnd() ID {
	if g.Mod(ModSmallInventory) {
		return FirstMapID - 2
	}
	return FirstMapID
}

// InventoryItems returns an iterator over player's primary and secondary spirits.
func (g *Game) InventoryItems() iter.Seq2[ID, *Entity] {
	to := g.InventoryEnd()
	return func(yield func(ID, *Entity) bool) {
		for i, e := range g.Entities[PrimarySpiritID:to] {
			if !yield(ID(i)+PrimarySpiritID, e) {
				return
			}
		}
	}
}

// AllPortalIDs returns an iterator over portal IDs, including malfunctioning
// ones.
func (g *Game) AllPortalIDs() []ID {
	var ids []ID
	for id, e := range g.MapEntities() {
		if _, ok := e.Role.(*Portal); ok {
			ids = append(ids, id)
		}
	}
	return ids
}
