// This file defines most actions available in the game.

package main

import (
	"cmp"
	"fmt"
	"log"
	"runtime"
	"slices"
	"strings"
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/ui"
)

// Action represents types that describe and handle a game action, often the
// last UI Action performed.
type Action interface {
	// Handle processes an action and returns possibly an effect along with
	// a boolean that reports whether the action ends the player's game
	// turn.
	Handle(*model) (gruid.Effect, bool)
}

// ActionNone does nothing.
type ActionNone struct{}

func (a ActionNone) Handle(md *model) (gruid.Effect, bool) {
	return nil, false
}

// ActionWait waits for a turn.
type ActionWait struct{}

func (a ActionWait) Handle(md *model) (gruid.Effect, bool) {
	md.g.stopSprint()
	return nil, true
}

func (g *Game) stopSprint() {
	if pa := g.PlayerActor(); pa.Has(StatusSprint) {
		pa.Statuses[StatusSprint] = 0
		g.LogStyled("You stop sprinting.", logStatusEnd)
	}
}

func (a ActionWait) String() string {
	return "Wait a turn"
}

// ActionBump moves the player to a given position and updates FOV information,
// or attacks if there is a monster.
type ActionBump struct {
	Delta gruid.Point
}

func (a ActionBump) Handle(md *model) (eff gruid.Effect, done bool) {
	return nil, md.g.PlayerBump(a.Delta)
}

func (a ActionBump) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "Move west"
	case gruid.Point{1, 0}:
		text = "Move east"
	case gruid.Point{0, -1}:
		text = "Move north"
	case gruid.Point{0, 1}:
		text = "Move south"
	}
	return text
}

// ActionRun performs automatic movement in a given direction.
type ActionRun struct {
	Delta gruid.Point
}

func (a ActionRun) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.DangerInFOV() {
		g.Log("You cannot run: danger in view!")
		return nil, false
	}
	if pa := g.PlayerActor(); pa.Has(StatusPoison) {
		g.Log("You cannot run while poisoned!")
		return nil, false
	}
	p := g.PP().Add(a.Delta)
	if !g.PlayerPassableNoTraps(p) {
		g.Log("You cannot run in that direction.")
		return nil, false
	}
	bump := ActionBump(a)
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	md.UpdateRun(a.Delta)
	md.targ.CancelExamine()
	return gruid.Batch(eff, md.autoCmd()), done
}

func (a ActionRun) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "Run west"
	case gruid.Point{1, 0}:
		text = "Run east"
	case gruid.Point{0, -1}:
		text = "Run north"
	case gruid.Point{0, 1}:
		text = "Run south"
	}
	return text
}

// ActionTravel starts travelling to the currently examined target.
type ActionTravel struct{}

func (a ActionTravel) Handle(md *model) (gruid.Effect, bool) {
	tp := md.targ.p
	if tp == InvalidPos {
		return nil, false
	}
	g := md.g
	pp := g.PP()
	if tp == pp {
		// wait a turn
		return nil, true
	}
	path := md.auto.path
	if len(path) <= 1 || path[len(path)-1] != tp || path[0] != pp {
		path = g.PlayerPath(pp, tp)
	}
	if len(path) <= 1 {
		return nil, false
	}
	fif := g.DangerInFOV() || g.PlayerActor().Has(StatusPoison)
	bump := ActionBump{Delta: path[1].Sub(pp)}
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	md.auto.path = path[1:]
	if fif {
		// do not start auto-travel if there was a foe in view.
		return eff, true
	}
	md.targ.CancelExamine()
	md.auto.mode = autoTravel
	return gruid.Batch(eff, md.autoCmd()), true
}

func (a ActionTravel) String() string {
	return "Travel to target"
}

// ActionAutoExplore starts automatic exploration.
type ActionAutoExplore struct{}

func (a ActionAutoExplore) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.DangerInFOV() {
		g.Log("You cannot auto-explore: danger in view!")
		return nil, false
	}
	if pa := g.PlayerActor(); pa.Has(StatusPoison) {
		g.Log("You cannot auto-explore while poisoned!")
		return nil, false
	}
	md.auto.aemRebuild = true
	n, ok := md.UpdateAutoExplore()
	if !ok {
		return nil, false
	}
	bump := ActionBump{Delta: n.Sub(g.PP())}
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	md.targ.CancelExamine()
	md.auto.mode = autoExplore
	return gruid.Batch(eff, md.autoCmd()), true
}

func (a ActionAutoExplore) String() string {
	return "Auto-explore"
}

// ActionCursorBump moves the cursor.
type ActionCursorBump struct {
	Delta gruid.Point
}

func (a ActionCursorBump) Handle(md *model) (gruid.Effect, bool) {
	np := md.targ.p.Add(a.Delta)
	md.Examine(np)
	return nil, false
}

func (a ActionCursorBump) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "Move cursor west"
	case gruid.Point{1, 0}:
		text = "Move cursor east"
	case gruid.Point{0, -1}:
		text = "Move cursor north"
	case gruid.Point{0, 1}:
		text = "Move cursor south"
	}
	return text
}

// ActionCursorRun moves the cursor (fast).
type ActionCursorRun struct {
	Delta gruid.Point
}

func (a ActionCursorRun) Handle(md *model) (gruid.Effect, bool) {
	np := md.targ.p.Add(a.Delta.Mul(4))
	if np.X < 0 {
		np.X = 0
	}
	if np.X >= MapWidth {
		np.X = MapWidth - 1
	}
	if np.Y < 0 {
		np.Y = 0
	}
	if np.Y >= MapHeight {
		np.Y = MapHeight - 1
	}
	md.Examine(np)
	return nil, false
}

func (a ActionCursorRun) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "Move cursor fast west"
	case gruid.Point{1, 0}:
		text = "Move cursor fast east"
	case gruid.Point{0, -1}:
		text = "Move cursor fast north"
	case gruid.Point{0, 1}:
		text = "Move cursor fast south"
	}
	return text
}

// ActionNextMonster examines next monster (sorted by distance to player).
type ActionNextMonster struct{}

func (a ActionNextMonster) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.Wizard.Mode == WizardRevealTerrain {
		return nil, false
	}
	me := g.getMonsterTargIDs()
	if len(me) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(me)
	i := g.getMonsterTargIDIndex(me)
	i = (i + 1) % len(me)
	var p gruid.Point
	if g.Wizard.Mode == WizardReveal {
		p = g.Entity(me[i]).P
	} else {
		p = g.Entity(me[i]).KnownP
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

// getMonsterTargIDs returns the ids of monsters that can be examined.
func (g *Game) getMonsterTargIDs() []ID {
	var me []ID
	for i, ei := range g.NPMapEntities() {
		ai, ok := ei.Role.(*Actor)
		if !ok {
			continue
		}
		var b bool // whether monster is examinable
		switch {
		case g.Wizard.Mode == WizardReveal:
			b = ai.IsAlive()
		default:
			b = ei.KnownP != InvalidPos && !ai.KnownDead
		}
		if b {
			me = append(me, i)
		}
	}
	return me
}

// sortEntityTargIDs sorts the ids of monsters that can be examined by
// distance to the player.
func (g *Game) sortEntityTargIDs(me []ID) {
	pp := g.PP()
	slices.SortFunc(me, func(i, j ID) int {
		pi, pj := g.Entity(i).KnownP, g.Entity(j).KnownP
		if g.Wizard.Mode.Reveal() {
			pi, pj = g.Entity(i).P, g.Entity(j).P
		}
		return cmp.Compare(paths.DistanceManhattan(pp, pi), paths.DistanceManhattan(pp, pj))
	})
}

// getMonsterTargIDIndex returns the id of any monster currently being
// examined, or -1 if none.
func (g *Game) getMonsterTargIDIndex(me []ID) int {
	if g.Wizard.Mode == WizardReveal {
		if id, _ := g.ActorAt(g.md.targ.p); id >= 0 {
			return slices.Index(me, id)
		}
		return -1
	}
	if id, _ := g.KnownActorAt(g.md.targ.p); id >= 0 {
		return slices.Index(me, id)
	}
	return -1
}

func (a ActionNextMonster) String() string {
	return "Examine next monster"
}

// ActionPreviousMonster examines previous monster (sorted by distance to
// player).
type ActionPreviousMonster struct{}

func (a ActionPreviousMonster) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.Wizard.Mode == WizardRevealTerrain {
		return nil, false
	}
	me := g.getMonsterTargIDs()
	if len(me) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(me)
	i := g.getMonsterTargIDIndex(me)
	if i == -1 {
		i = 0
	}
	i--
	if i < 0 {
		i += len(me)
	}
	var p gruid.Point
	if g.Wizard.Mode == WizardReveal {
		p = g.Entity(me[i]).P
	} else {
		p = g.Entity(me[i]).KnownP
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

func (a ActionPreviousMonster) String() string {
	return "Examine previous monster"
}

// ActionNextItem examines next food item (sorted by distance to player).
type ActionNextItem struct{ Type itemType }

// itemType represents various kinds of targetable item kinds.
type itemType int

const (
	itemComestible itemType = iota
	itemTotem
	itemMenhir
	itemPortal // includes CorruptionOrb too
	itemRune
)

func item2type(it Item) itemType {
	switch it.(type) {
	case *Comestible:
		return itemComestible
	case *Spirit:
		return itemTotem
	case *EmptyTotem:
		return itemTotem
	case *Menhir:
		return itemMenhir
	case *Portal:
		return itemPortal
	case *CorruptionOrb:
		return itemPortal
	case *RunicTrap:
		return itemRune
	default:
		panic("unexpected item")
	}
}

func (a ActionNextItem) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	items := g.getItemTargIDs(a.Type)
	if len(items) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(items)
	i := g.getItemTargIDIndex(items, a.Type)
	i = (i + 1) % len(items)
	var p gruid.Point
	if g.Wizard.Mode.Reveal() {
		p = g.Entity(items[i]).P
	} else {
		p = g.Entity(items[i]).KnownP
	}
	if p == InvalidPos {
		return nil, false
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

// getItemTargIDs returns the ids of items that can be examined.
func (g *Game) getItemTargIDs(ty itemType) []ID {
	var items []ID
	for i, it := range g.ItemEntities() {
		if item2type(it) != ty {
			continue
		}
		if g.Entity(i).KnownP != InvalidPos || g.Wizard.Mode.Reveal() {
			items = append(items, i)
		}
	}
	return items
}

// getItemTargIDIndex returns the id of any monster currently being
// examined, or -1 if none.
func (g *Game) getItemTargIDIndex(items []ID, ty itemType) int {
	if g.Wizard.Mode.Reveal() {
		if id, it := g.ItemAt(g.md.targ.p); id >= 0 {
			if item2type(it) == ty {
				return slices.Index(items, id)
			}
		}
		return -1
	}
	if id, it := g.KnownItemAt(g.md.targ.p); id >= 0 {
		if item2type(it) == ty {
			return slices.Index(items, id)
		}
	}
	return -1
}

func (a ActionNextItem) String() string {
	switch a.Type {
	case itemComestible:
		return "Examine next known comestible"
	case itemTotem:
		return "Examine next known spirit"
	case itemMenhir:
		return "Examine next known menhir"
	case itemPortal:
		// Also works for the orb of corruption.
		return "Examine next known portal"
	case itemRune:
		return "Examine next known runic trap"
	default:
		// Should not happen.
		return "Examine next known item"
	}
}

// ActionExamine examines a target screen position (relative position ignoring
// log at the top, so starting at the map).
type ActionExamine struct {
	Target gruid.Point
}

func (a ActionExamine) Handle(md *model) (gruid.Effect, bool) {
	p := a.Target
	if inMap(p) {
		md.Examine(p)
	} else {
		md.targ.CancelExamine()
	}
	return nil, false
}

// ActionExamineModeToggle enters keyboard targeting mode.
type ActionExamineModeToggle struct{}

func (a ActionExamineModeToggle) Handle(md *model) (gruid.Effect, bool) {
	if !md.targ.kb {
		md.targ.kb = true
		md.Examine(md.g.PP())
		return nil, false
	}
	md.targ.CancelExamine()
	return nil, false
}

func (a ActionExamineModeToggle) String() string {
	return "Examine (keyboard mode)"
}

// ActionScroll represents a vertical scrolling action (only usable for target
// examination currently).
type ActionScroll struct {
	Delta gruid.Point
}

func (a ActionScroll) Handle(md *model) (gruid.Effect, bool) {
	switch a.Delta {
	case gruid.Point{0, 1}:
		if md.targ.p != InvalidPos {
			md.targ.scroll = false
		}
	case gruid.Point{0, -1}:
		if md.targ.p != InvalidPos {
			md.targ.scroll = true
		}
	}
	return nil, false
}

// ActionInteract interacts with items on current cell.
type ActionInteract struct{}

func (a ActionInteract) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	md.mode = modeNormal
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.logDaze()
		return nil, false
	}
	pp := g.PP()
	for i, ei := range g.NPMapEntities() {
		if ei.P != pp {
			continue
		}
		switch it := ei.Role.(type) {
		case *Comestible:
			return nil, equipItem(md, i)
		case *Spirit:
			return nil, equipItem(md, i)
		case Item:
			// Interacting with non-equippable means using them.
			return nil, it.Use(g, i)
		}
	}
	g.Log("Nothing to interact with.")
	return nil, false
}

func (a ActionInteract) String() string {
	return "Interact (equip/activate)"
}

// ActionInventory opens the inventory menu.
type ActionInventory struct{}

func (a ActionInventory) Handle(md *model) (gruid.Effect, bool) {
	entries := []ui.MenuEntry{}
	r := 'a'
	for i, ei := range md.g.InventoryItems() {
		entries = appendMenuHeader(entries, i)
		var entry ui.MenuEntry
		if ei.IsItem() {
			entry.Text = ui.Textf("%c - %s", r, ei.Text())
			entry.Keys = []gruid.Key{gruid.Key(r)}
		} else {
			entry.Text = ui.Textf("  - %s", ei.Name)
			entry.Disabled = true
		}
		entries = append(entries, entry)
		r = nextRuneKey(r)
	}
	if id, it := md.g.ItemAt(md.g.PP()); id >= 0 {
		if _, ok := it.(*Comestible); ok {
			entries = appendMenuHeader(entries, InventorySize)
			var entry ui.MenuEntry
			entry.Text = ui.Textf("%c - %s", r, md.g.Entity(id).Name)
			entry.Keys = []gruid.Key{gruid.Key(r)}
			entries = append(entries, entry)
		}
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("Inventory (use)").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.mode = modeMenu
	md.menu.mode = modeInventory
	md.menu.main.SetActiveInvokable(0)
	e := md.g.Entity(PrimarySpiritID)
	md.updateItemDesc(e)
	return nil, false
}

func nextRuneKey(r rune) rune {
	for {
		r++
		switch r {
		case 'h', 'l', 'j', 'k':
		default:
			return r
		}
	}
}

func appendMenuHeader(entries []ui.MenuEntry, i ID) []ui.MenuEntry {
	hstyle := gruid.Style{}.WithFg(ColorCyan)
	switch i {
	case PrimarySpiritID:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("Primary Spirit").WithStyle(hstyle),
			Disabled: true,
		})
	case PrimarySpiritID + 1:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("Secondary Spirits").WithStyle(hstyle),
			Disabled: true,
		})
	case PrimarySpiritID + NSpirits:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("Comestibles").WithStyle(hstyle),
			Disabled: true,
		})
	case PrimarySpiritID + NSpirits + NComestibles:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("Comestible on the floor").WithStyle(hstyle),
			Disabled: true,
		})
	}
	return entries
}

// inventoryItemID returns the entity ID corresponding to the given invokable
// inventory index.
func (md *model) inventoryItemID(i int) ID {
	g := md.g
	n := -1
	for j, e := range g.Entities[:InventorySize] {
		if e.Role != nil {
			n++
		}
		if n < i {
			continue
		}
		return ID(j + PrimarySpiritID)
	}
	id, _ := md.g.ItemAt(md.g.PP())
	return id // should be positive
}

// equipItemID returns the entity ID corresponding to the given equipable
// inventory item equip index.
func (md *model) equipItemID(it Item, idx int) ID {
	switch it.(type) {
	case *Spirit:
		return md.equipSpiritID(idx)
	default:
		// Comestible.
		return ID(idx + NSpirits + PrimarySpiritID)
	}
}

// equipSpiritID returns the entity ID corresponding to the given equipable
// inventory spirit equip index.
func (md *model) equipSpiritID(idx int) ID {
	n := -1
	for i, e := range md.g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); !ok || sp.Level == 0 {
			n++
		}
		if n == idx {
			return ID(i)
		}
	}
	return ID(-1)
}

// equipItemIdx returns the item equip index corresponding to the given
// equipped entity ID.
func (md *model) equipItemIdx(i ID) int {
	g := md.g
	switch g.Entity(i).Role.(type) {
	case *Spirit:
		for _, e := range g.Entities[:i] {
			if sp, ok := e.Role.(*Spirit); !ok || sp.Level == 1 {
				i--
			}
		}
		return int(i - PrimarySpiritID)
	default:
		// Comestible.
		return int(i - NSpirits - PrimarySpiritID)
	}
}

// inventoryItemIdx returns the inventory item index corresponding to the given
// equipped entity ID.
// func (md *model) inventoryItemIdx(i ID) int {
// 	g := md.g
// 	switch g.Entity(i).Role.(type) {
// 	case *Spirit:
// 		for _, e := range g.Entities[:i] {
// 			if e.Role == nil {
// 				i--
// 			}
// 		}
// 		return int(i - PrimarySpiritID)
// 	default:
// 		// Comestible.
// 		return int(i - NSpirits - PrimarySpiritID)
// 	}
// }

// altBgEntries updates entries to use alternate background color for entries
// of odd index.
func altBgEntries(entries []ui.MenuEntry) {
	for i := range entries {
		if i%2 == 1 {
			st := entries[i].Text.Style()
			entries[i].Text = entries[i].Text.WithStyle(st.WithBg(ColorBackgroundSecondary))
		}
	}
}

// updateItemDesc updates the description label with text for the given entity.
func (md *model) updateItemDesc(e *Entity) {
	l := md.desc
	stt := ui.StyledText{}.WithMarkups(Markups)
	l.Box = &ui.Box{Title: ui.Text(e.Name)}
	switch it := e.Role.(type) {
	case *Spirit:
		if it.Level == 1 {
			l.Box.Title = ui.Textf("%s (upgraded)", e.Name)
		}
		l.Content = stt.WithText(it.Desc()).Format(UIWidth/2 - 1 - 2)
	case Item:
		l.Content = stt.WithText(it.Desc()).Format(UIWidth/2 - 1 - 2)
	default:
		// Should not happen.
		l.Content = stt.WithText(e.Name).Format(UIWidth/2 - 1 - 2)
	}
}

func (a ActionInventory) String() string {
	return "Open inventory"
}

// ActionUseItem uses an item.
type ActionUseItem struct {
	ID ID
}

func (a ActionUseItem) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if a.ID >= InventorySize {
		// Comestible on the floor.
		if a.ID, _ = g.ItemAt(g.PP()); a.ID < 0 {
			// Should not happen.
			return nil, false
		}
	}
	e := g.Entity(a.ID)
	ev, ok := e.Role.(Item)
	if ok {
		ok = ev.Use(g, a.ID)
	}
	if !ok {
		md.mode = modeNormal
	}
	return nil, ok
}

func equipItem(md *model, i ID) bool {
	g := md.g
	ei := g.Entity(i)
	spirit := false
	var from, to ID
	switch ei.Role.(type) {
	case *Spirit:
		spirit = true
		from, to = PrimarySpiritID, PrimarySpiritID+NSpirits
	case *Comestible:
		for i, e := range g.Entities[PrimarySpiritID+NSpirits : FirstMapID] {
			if !e.IsItem() {
				// Equip on first empty slot.
				return g.EquipItemAt(ID(i + NSpirits))
			}
		}
		from, to = PrimarySpiritID+NSpirits, FirstMapID
	}
	entries := []ui.MenuEntry{}
	r := 'a'
	for j, ej := range g.InventoryItems() {
		entries = appendMenuHeader(entries, j)
		var entry ui.MenuEntry
		entry.Disabled = j < from || j >= to
		if !entry.Disabled {
			if rj, ok := ej.Role.(*Spirit); ok && rj.Level >= 1 {
				entry.Disabled = true
				if j == from {
					from++
				}
			}
		}
		if entry.Disabled {
			entry.Text = ui.Textf("  - %s", ej.Name)
		} else {
			entry.Text = ui.Textf("%c - %s", r, ej.Text())
			entry.Keys = []gruid.Key{gruid.Key(r)}
		}
		entries = append(entries, entry)
		r = nextRuneKey(r)
	}
	if spirit && from == to {
		// NOTE: At some point I considered providing another kind of
		// benefit when all spirits were already upgraded, but I
		// decided against it to not encourage perfectionism.
		g.Log("No spirit upgrades or new slots available.")
		return false
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("Inventory (equip)").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.mode = modeMenu
	md.menu.mode = modeEquip
	ej := g.Entity(from)
	md.menu.main.SetActiveInvokable(md.equipItemIdx(from))
	_, ok := ej.Role.(*Spirit)
	md.drawGroundDesc = !ok
	md.updateEquipItemDesc(ei, ej)
	return false
}

// updateEquipItemDesc updates the label for equipping ei on ej.
func (md *model) updateEquipItemDesc(ei, ej *Entity) {
	l := md.desc
	stt := ui.StyledText{}.WithMarkups(Markups)
	l.Box = &ui.Box{Title: ui.Text(ej.Name)}
	switch rj := ej.Role.(type) {
	case *Spirit:
		l.Box.Title = ui.Textf("%s (upgrade)", ej.Name)
		l.Content = stt.WithText(rj.UpgradeDesc()).Format(UIWidth/2 - 1 - 2)
	case *Comestible:
		l.Box.Title = ui.Textf("%s (replace)", ej.Name)
		l.Content = stt.WithText(rj.Desc()).Format(UIWidth/2 - 1 - 2)
		if ri, ok := ei.Role.(*Comestible); ok {
			md.equipLabel.Box = &ui.Box{Title: ui.Textf("%s (ground)", ei.Name)}
			md.equipLabel.Content = stt.WithText(ri.Desc()).Format(UIWidth/2 - 1 - 2)
		}
	default:
		// Empty slot: should only happen with a spirit.
		if ri, ok := ei.Role.(*Spirit); ok {
			l.Content = stt.WithText("@MThe spirit on the ground can be equipped on this empty slot.@N").Format(UIWidth/2 - 1 - 2)
			md.equipLabel.Box = &ui.Box{Title: ui.Textf("%s (ground)", ei.Name)}
			md.equipLabel.Content = stt.WithText(ri.Desc()).Format(UIWidth/2 - 1 - 2)
		}
	}
}

// ActionEquipItemAt equips an item.
type ActionEquipItemAt struct {
	ID ID
}

func (a ActionEquipItemAt) Handle(md *model) (gruid.Effect, bool) {
	done := md.g.EquipItemAt(a.ID)
	if !done {
		md.mode = modeNormal
	}
	return nil, done
}

// ActionMenu open main game menu.
type ActionMenu struct{}

// menuActions represents the various entries in the main menu: they should
// have a corresponding entry in menuKeys.
var menuActions = []Action{
	ActionInteract{},
	ActionInventory{},
	ActionViewMessages{},
	ActionHelp{},
	ActionDump{},
	ActionConfig{},
	ActionSaveQuit{},
	ActionQuit{},
	ActionWizardNextLevel{},
}

var menuKeys = []rune{'e', 'i', 'm', '?', '#', 'C', 'S', 'Q', '>'}

func init() {
	if len(menuActions) != len(menuKeys) {
		panic("length mismatch between menuActions and menuKeys")
	}
}

func (a ActionMenu) Handle(md *model) (gruid.Effect, bool) {
	const nwizard = 1 // number of wizard actions
	md.menuActions = menuActions
	if !md.g.Wizard.Extra {
		md.menuActions = md.menuActions[:len(md.menuActions)-nwizard]
	}
	hstyle := gruid.Style{}.WithFg(ColorCyan)
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := menuKeys[i]
		switch r {
		case 'e':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Gameplay Actions").WithStyle(hstyle),
				Disabled: true,
			})
		case 'm':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Gameplay Info").WithStyle(hstyle),
				Disabled: true,
			})
		case 'C':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Other Actions").WithStyle(hstyle),
				Disabled: true,
			})
		case '>':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Wizard Actions").WithStyle(hstyle),
				Disabled: true,
			})
		}
		disabled := false
		if r == 'e' {
			// The first action (interact) isn't always available.
			if j, _ := md.g.ItemAt(md.g.PP()); j < 0 {
				disabled = true
				md.menuActions = md.menuActions[1:]
			}
		}
		if !disabled {
			entries = append(entries, ui.MenuEntry{
				Text: ui.Textf("%c - %s", r, it),
				Keys: []gruid.Key{gruid.Key(r)},
			})
		}
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("Menu").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menu.mode = modeGameMenu
	return nil, false
}

func (a ActionMenu) String() string {
	return "Menu"
}

// ActionConfig opens settings menu.
type ActionConfig struct{}

var configActions = []Action{
	ActionSetKeys{},
	ActionToggleDarkLight{},
	ActionToggleWarningFirePois{},
}

var configKeys = []rune{':', 'c', 'w', 't'}

func (a ActionConfig) Handle(md *model) (gruid.Effect, bool) {
	md.menuActions = configActions
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := configKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("Config").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menu.mode = modeConfigMenu
	return nil, false
}

func (a ActionConfig) String() string {
	return "Configure settings"
}

// ActionToggleDarkLight enters keyboard targeting mode.
type ActionToggleDarkLight struct{}

func (a ActionToggleDarkLight) Handle(md *model) (gruid.Effect, bool) {
	GameConfig.DarkColors = !GameConfig.DarkColors
	err := SaveConfig()
	if err != nil {
		log.Printf("saving config: %v", err)
		md.g.Log(err.Error())
	}
	clearCache()
	eff := gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
	md.mode = modeNormal
	return eff, false
}

func (a ActionToggleDarkLight) String() string {
	if GameConfig.DarkColors {
		return "Switch to light color theme"
	}
	return "Switch to dark color theme"
}

// ActionToggleWarningFirePois enters keyboard targeting mode.
type ActionToggleWarningFirePois struct{}

func (a ActionToggleWarningFirePois) Handle(md *model) (gruid.Effect, bool) {
	GameConfig.WarnFirePoison = !GameConfig.WarnFirePoison
	err := SaveConfig()
	if err != nil {
		log.Printf("saving config: %v", err)
		md.g.Log(err.Error())
	}
	clearCache()
	eff := gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
	md.mode = modeNormal
	return eff, false
}

func (a ActionToggleWarningFirePois) String() string {
	if GameConfig.WarnFirePoison {
		return "Disable warning on Fire/Poison"
	}
	return "Enable warning on Fire/Poison"
}

// ActionSetKeys opens the keymap settings.
type ActionSetKeys struct{}

// configurableAction represents actions that may be configurable so need a
// textual representation in the UI.
type configurableAction interface {
	fmt.Stringer
	Action
}

var configurableKeyActions = [...]configurableAction{
	ActionBump{Delta: gruid.Point{-1, 0}},
	ActionBump{Delta: gruid.Point{1, 0}},
	ActionBump{Delta: gruid.Point{0, -1}},
	ActionBump{Delta: gruid.Point{0, 1}},
	ActionRun{Delta: gruid.Point{-1, 0}},
	ActionRun{Delta: gruid.Point{1, 0}},
	ActionRun{Delta: gruid.Point{0, -1}},
	ActionRun{Delta: gruid.Point{0, 1}},
	ActionWait{},
	ActionCursorBump{Delta: gruid.Point{-1, 0}},
	ActionCursorBump{Delta: gruid.Point{1, 0}},
	ActionCursorBump{Delta: gruid.Point{0, -1}},
	ActionCursorBump{Delta: gruid.Point{0, 1}},
	ActionCursorRun{Delta: gruid.Point{-1, 0}},
	ActionCursorRun{Delta: gruid.Point{1, 0}},
	ActionCursorRun{Delta: gruid.Point{0, -1}},
	ActionCursorRun{Delta: gruid.Point{0, 1}},
	ActionInteract{},
	ActionInventory{},
	ActionViewMessages{},
	ActionExamineModeToggle{},
	ActionNextItem{itemComestible},
	ActionNextItem{itemTotem},
	ActionNextItem{itemMenhir},
	ActionNextItem{itemPortal},
	ActionNextItem{itemRune},
	ActionNextMonster{},
	ActionPreviousMonster{},
	ActionTravel{},
	ActionAutoExplore{},
	ActionMenu{},
	ActionHelp{},
	ActionConfig{},
	ActionSaveQuit{},
	ActionQuit{},
	ActionWizard{},
}

func (a ActionSetKeys) Handle(md *model) (gruid.Effect, bool) {
	entries := []ui.MenuEntry{}
	r := 'a'
	for _, it := range configurableKeyActions {
		desc := it.String()
		desc = fmt.Sprintf(" %-36s %s", desc, md.keysForAction(it))
		entries = append(entries, ui.MenuEntry{
			Text: ui.Text(desc),
		})
		r = nextRuneKey(r)
	}
	altBgEntries(entries)
	md.menu.keys.SetBox(&ui.Box{Title: ui.Text("Key Bindings").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.keys.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menu.mode = modeKeysView
	return nil, false
}

func (md *model) keysForAction(a Action) string {
	keys := []gruid.Key{}
	for k, action := range md.keysNormal {
		if a == action && !k.In(keys) {
			keys = append(keys, k)
		}
	}
	if KbTargetingMode(a) {
		for k, action := range md.keysTarget {
			if a == action && !k.In(keys) {
				keys = append(keys, k)
			}
		}
	}
	ks := make([]string, len(keys))
	for i := range ks {
		ki := string(keys[i])
		if ki == " " {
			ki = "SPACE"
		}
		ks[i] = ki
	}
	slices.Sort(ks)
	return strings.Join(ks, " or ")
}

func (a ActionSetKeys) String() string {
	return "View/Customize keybindings"
}

// ActionViewMessages opens the log message viewer.
type ActionViewMessages struct{}

func (a ActionViewMessages) Handle(md *model) (gruid.Effect, bool) {
	if len(md.pager.lines) > 0 {
		md.pager.lines = md.pager.lines[:len(md.pager.lines)-1]
	}
	for _, e := range md.g.Logs.Entries[len(md.pager.lines):] {
		md.pager.lines = append(md.pager.lines, md.pager.markup.WithText(e.MText))
	}
	md.pager.pg.SetLines(md.pager.lines)
	md.pager.pg.SetCursor(gruid.Point{0, len(md.pager.lines)})
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text("Messages").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.mode = modePager
	md.pager.mode = modeLogs
	return nil, false
}

func (a ActionViewMessages) String() string {
	return "View messages"
}

// ActionDump writes a dump with game statistics.
type ActionDump struct{}

func (a ActionDump) Handle(md *model) (gruid.Effect, bool) {
	if msg, err := md.g.WriteDump(); err != nil {
		md.g.LogfStyled("Error: %v.", logError, err)
	} else {
		md.g.Log(msg)
	}
	return nil, false
}

func (a ActionDump) String() string {
	return "Dump game statistics"
}

// ActionSaveQuit asks for quitting the game after saving.
type ActionSaveQuit struct{}

func (a ActionSaveQuit) Handle(md *model) (gruid.Effect, bool) {
	if err := md.g.Save(); err != nil {
		md.g.LogStyled("Error while saving game.", logError)
		return nil, false
	}
	md.mode = modeQuitting
	return gruid.End(), false
}

func (a ActionSaveQuit) String() string {
	return "Save and Quit"
}

// ActionQuit asks for quitting the game, without saving.
type ActionQuit struct{}

func (a ActionQuit) Handle(md *model) (gruid.Effect, bool) {
	md.mode = modeQuitConfirmation
	md.g.LogStyled("Do you really want to quit without saving? [y/N]", logConfirm)
	return nil, false
}

func (a ActionQuit) String() string {
	return "Quit (without saving)"
}

// ActionQuitConfirm quits the game.
type ActionQuitConfirm struct {
	State confirm
}

func (a ActionQuitConfirm) Handle(md *model) (gruid.Effect, bool) {
	switch a.State {
	case confirmTrue:
		md.mode = modeQuitting
		err := RemoveSaveFile()
		if err != nil {
			log.Printf("Error removing save file: %v", err)
		}
		RemoveReplay()
		return gruid.End(), false
	case confirmFalse:
		md.g.Log("Keep playing, then.")
		md.mode = modeNormal
	}
	return nil, false
}

// Wizard collects information relative to wizard mode.
type Wizard struct {
	Mode          WizardMode // Currently active wizard mode
	Resurrections int        // Number of resurrections
	Extra         bool       // Whether to enable extra wizard cheats
}

// WizardMode describes the possible wizard modes.
type WizardMode int

const (
	WizardNone          WizardMode = iota
	WizardNormal                   // non-permadeath mode with normal visibility
	WizardReveal                   // reveal map and monsters
	WizardRevealTerrain            // reveal map (without monsters)
)

func (wm WizardMode) Reveal() bool {
	return wm == WizardReveal || wm == WizardRevealTerrain
}

// ActionWizard enters wizard-mode confirmation. When already in wizard mode,
// cycle to next wizard visibility mode. The first wizard mode is just a
// non-permadeath mode, without extra cheats (other than resurrection).
type ActionWizard struct{}

func (a ActionWizard) Handle(md *model) (gruid.Effect, bool) {
	switch md.g.Wizard.Mode {
	case WizardNone:
		md.g.LogStyled("Do you really want to enter Wizard Mode? (non-permadeath mode) [y/N]", logConfirm)
		md.mode = modeWizardConfirmation
	case WizardNormal:
		if !md.g.Wizard.Extra {
			md.g.LogStyled("Do you really want extra cheats? (reveal map) [y/N]", logConfirm)
			md.mode = modeWizardConfirmation
		} else {
			md.g.Wizard.Mode = WizardReveal
			md.targ.CancelExamine()
		}
	case WizardReveal:
		md.g.Wizard.Mode = WizardRevealTerrain
		md.targ.CancelExamine()
	case WizardRevealTerrain:
		md.g.Wizard.Mode = WizardNormal
		md.targ.CancelExamine()
	}
	return nil, false
}

func (a ActionWizard) String() string {
	return "Enter Wizard Mode (irreversible)"
}

// ActionWizardConfirm enters wizard-mode.
type ActionWizardConfirm struct {
	State confirm
}

func (a ActionWizardConfirm) Handle(md *model) (gruid.Effect, bool) {
	switch a.State {
	case confirmTrue:
		if md.g.Wizard.Mode == WizardNone {
			md.g.Wizard.Mode = WizardNormal
			md.g.LogStyled("Wizard Mode enabled.", logSpecial)
			md.g.StoryLog("Entered Wizard Mode")
		} else {
			md.g.Wizard.Extra = true
			md.g.Wizard.Mode = WizardReveal
			md.targ.CancelExamine()
			md.g.LogStyled("Extra cheats enabled.", logSpecial)
			md.g.StoryLog("Enabled extra Wizard Mode cheats!")
		}
		md.mode = modeNormal
		return nil, false
	case confirmFalse:
		md.mode = modeNormal
		if md.g.Wizard.Mode == WizardNone {
			md.g.Log("Keep playing normally, then.")
		} else {
			md.g.Log("Don’t reveal map, then.")
		}
	}
	return nil, false
}

// ActionWizardNextLevel is a debug cheat to quickly go to next level. Only
// accessible after revealing map.
type ActionWizardNextLevel struct{}

func (a ActionWizardNextLevel) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.Map.Level == MapLevels {
		g.Log("You’re already at the last level.")
		md.mode = modeNormal
		return nil, false
	}
	md.g.StoryLog("Wizard Mode: used “go to next level” cheat")
	g.LogStyled("You use the “go to next level” cheat.", logSpecial)
	g.NextLevel()
	md.mode = modeNormal
	return nil, false
}

func (a ActionWizardNextLevel) String() string {
	return "Go to next level"
}

// actionAuto handles automatic movement of all kinds. It is triggered when
// receiving msgAuto messages. The value of thoses messages is used to ensure
// that we only use it on a specific turn, or discard it, which should be
// normally the case, but better be safe than sorry.
type actionAuto struct {
	msg msgAuto
}

type msgAuto int

func (a actionAuto) Handle(md *model) (gruid.Effect, bool) {
	if int(a.msg) != md.g.Turn {
		return nil, false
	}
	switch md.auto.mode {
	case autoRun:
		if !md.KeepRunning() {
			md.auto.mode = noAuto
			return nil, false
		}
		bump := ActionBump{Delta: md.auto.delta}
		eff, done := bump.Handle(md)
		if !done {
			md.auto.mode = noAuto
			return eff, false
		}
		md.UpdateRun(md.auto.delta)
		return gruid.Batch(eff, md.autoCmd()), true
	case autoTravel:
		next, ok := md.UpdateTravel()
		if !ok {
			md.auto.mode = noAuto
			return nil, false
		}
		bump := ActionBump{Delta: next.Sub(md.g.PP())}
		eff, done := bump.Handle(md)
		if !done {
			md.auto.mode = noAuto
			return eff, false
		}
		return gruid.Batch(eff, md.autoCmd()), true
	case autoExplore:
		g := md.g
		n, ok := md.UpdateAutoExplore()
		if !ok {
			md.auto.mode = noAuto
			return nil, false
		}
		bump := ActionBump{Delta: n.Sub(g.PP())}
		eff, done := bump.Handle(md)
		if !done {
			md.auto.mode = noAuto
			return eff, false
		}
		return gruid.Batch(eff, md.autoCmd()), true
	}
	return nil, false
}

func (md *model) autoCmd() gruid.Effect {
	n := md.g.Turn
	return gruid.Cmd(func() gruid.Msg {
		t := time.NewTimer(AnimDurShort)
		<-t.C
		return msgAuto(n + 1)
	})
}

func (a actionAuto) String() string {
	return "action on msgAuto"
}

// KbTargetingMode reports whether the action is used in keyboard targeting mode.
func KbTargetingMode(a Action) bool {
	switch a.(type) {
	case ActionCursorBump, ActionCursorRun, ActionTravel:
		return true
	}
	return false
}

// ActionHelp enters a help menu.
type ActionHelp struct{}

// helpActions represents the various entries in the main help: they should
// have a corresponding entry in helpKeys.
var helpActions = []Action{
	ActionHelpDefaultKeys{},
	ActionHelpCombat{},
	ActionHelpItems{},
	ActionHelpStatuses{},
	ActionHelpTips{},
}

var helpKeys = []rune{'?', 'c', 'i', 's', 't'}

func init() {
	if len(helpActions) != len(helpKeys) {
		panic("length mismatch between helpActions and helpKeys")
	}
}

func (a ActionHelp) Handle(md *model) (gruid.Effect, bool) {
	md.menuActions = helpActions
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := helpKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("Help Menu").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menu.mode = modeHelpMenu
	return nil, false
}

func (a ActionHelp) String() string {
	return "Help"
}

// ActionHelpDefaultKeys enters a help menu.
type ActionHelpDefaultKeys struct{}

func (a ActionHelpDefaultKeys) Handle(md *model) (gruid.Effect, bool) {
	md.KeysHelp()
	return nil, false
}

func (a ActionHelpDefaultKeys) String() string {
	return "Keybindings (default values)"
}

func (md *model) updateKeysDescription(title string, actions []string) {
	md.pager.mode = modeHelp
	md.mode = modePager
	if CustomKeys {
		title = fmt.Sprintf(" Default %s ", title)
	} else {
		title = fmt.Sprintf(" %s ", title)
	}
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	lines := []ui.StyledText{}
	for i := 0; i < len(actions)-1; i += 2 {
		stt := ui.StyledText{}
		if actions[i+1] != "" {
			stt = stt.WithTextf(" %-36s %s", actions[i], actions[i+1])
		} else {
			stt = stt.WithTextf(" %s ", actions[i]).WithStyle(gruid.Style{}.WithFg(ColorCyan))
		}
		if i%4 == 2 {
			stt = stt.WithStyle(stt.Style().WithBg(ColorBackgroundSecondary))
		}
		lines = append(lines, stt)
	}
	md.pager.pg.SetLines(lines)
	md.pager.pg.SetCursor(gruid.Point{0, 0})
}

func (md *model) KeysHelp() {
	runKeys := "shift+arrows or HJKL"
	if runtime.GOOS == "js" {
		runKeys = "HJKL"
	}
	entries := []string{
		"Basic Game Actions", "",
		"Move or Attack", "arrows or hjkl or mouse left",
		"Wait a turn", "“.” or ENTER or mouse left on @",
		"Interact (Equip/Activate/…)", "e",
		"Use inventory item", "i",
		"Toggle keyboard examine mode", "x",
		"Open menu", "SPACE or mouse right",
		"Close menu, inventory…", "SPACE or ESC or mouse left outside",
		"Advanced Game Actions", "",
		"View previous messages", "m",
		"Examine next portal/menhir/rune", "> & =",
		"Examine next totem/comestible", "! %",
		"Examine previous/next monster", "- +",
		"Run (auto-move in direction)", runKeys,
		"Travel (auto-move to destination)", "“.” or ENTER in keyboard examine mode",
		"Autoexplore (use with caution)", "o",
		"Other Common Actions", "",
		"Save and Quit", "S",
		"Quit (without saving)", "Q",
	}
	md.updateKeysDescription("Keybindings (default values)", entries)
}

// helpTopic opens the pager with the given title and text content.
func (md *model) helpTopic(title, content string) {
	text := ui.Text(content).WithMarkups(Markups).Format(78)
	stts := text.Lines()
	md.pager.pg.SetLines(stts)
	md.pager.pg.SetCursor(gruid.Point{0, 0})
	md.mode = modePager
	md.pager.mode = modeHelp
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{}.WithFg(ColorYellow))})
}

// ActionHelpStatuses opens the log message viewer.
type ActionHelpStatuses struct{}

func (a ActionHelpStatuses) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("Statuses (Help)", statusesHelpText())
	return nil, false
}

func (a ActionHelpStatuses) String() string {
	return "Statuses"
}

func statusesHelpText() string {
	var sb strings.Builder
	sb.WriteString(`@YCOLORS@N

In the status bar and descriptions, positive status effects are shown as @BBlue@N and negative ones are @OOrange@N.

On the map, monster color may change depending on active status effects.  By default, status-free wandering monsters are @OOrange@N and hunting monsters are @ORed@N. @MMagenta@N indicates a positive status. @YYellow@N indicates both positive and negative statuses. @GGreen@N indicates a negative status. @CCyan@N indicates two or more negative statuses.`)
	sb.WriteString("\n\n")
	sb.WriteString(`@YSTATUSES@N`)
	sb.WriteString("\n\n")
	for i, desc := range statusDesc {
		st := Status(i)
		color := statusColor(st, 42)
		fmt.Fprintf(&sb, "@%c%s@N (@%c%s@N). ", color, st.Name(), color, st.Abbr())
		sb.WriteString(desc)
		var details string
		switch st {
		case StatusBerserk:
			details = "HP never falls below 1 after losing bonus HP."
		case StatusClarity:
			details = "Player-only effect obtainable by eating clarity leaves. May be removed by eating a berserking flower for half the @BBerserk@N duration."
		case StatusLignification:
			details = "HP never falls below 3 after losing bonus HP. Cancels @BSprint@N. Players with “woody legs” can move while @Blignified@N at the cost of reducing the duration by an extra turn."
		case StatusConfusion:
			details = "Confused monsters will attack other monsters too when adjacent. Some monsters may even hurt or affect themselves when performing special actions like digging, spitting fire, barking, or pushing."
		case StatusFear:
			details = "Monsters will flee when facing you. Ambushing charges against unseen foes can still happen. At the end of a turn, a cornered and afraid actor with no possible escape will become berserk. "
		case StatusFire:
			details = "Players with “fire resistance” from the upgraded Fire Salamander don’t take that extra immediate damage."
		case StatusImbalance:
			details = "Sprinting or jumping while imbalanced may sometimes lead to a @Odazing@N fall."
		case StatusPoison:
			details = "Watch out for confusion if any monsters survive poisoning!"
		case StatusFoggySkin:
			details = "Player-only effect obtainable by eating a foggy-skin onion. May be removed by eating a lignification fruit for half the @BLignification@N duration."
		case StatusSprint:
			details = "Disables normal attack. Player-only effect obtainable with the Sprinting Gazelle spirit."
		case StatusTimeStop:
			details = "Player-only effect obtainable with the Temporal Cat spirit."
		case StatusFocus:
			details = "Cancels @BSprint@N. Player-only effect obtainable with the Four-Headed Hydra spirit."
		}
		if details != "" {
			sb.WriteRune(' ')
			sb.WriteString(details)
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// ActionHelpCombat opens the log message viewer.
type ActionHelpCombat struct{}

func (a ActionHelpCombat) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("Combat (Help)", combatHelpText())
	return nil, false
}

func (a ActionHelpCombat) String() string {
	return "Combat"
}

func combatHelpText() string {
	return strings.TrimSpace(`
@YTURN-BASED SYSTEM@N

Shamogu is a turn-based game. Each actor acts once per turn: the player acts first, then the monsters do. Monsters act one after another in a non-predictable order, but except during some animations, you’ll only see the final state and often feel they acted all at once.

@YCOMBAT DAMAGE@N

Combat damage is affected by attack (A) and defense (D). Maximum damage is the minimum of attack A and 3. Higher values of A increase chances of higher damage when attacking, while higher values of D decrease that chance when defending. Zero damage represents a miss.

Players under the @BFocus@N status effect and four-headed hydras perform four damage-rolls per attack, doing four times as much damage.

Special effects like @BBerserk@N or @BDig@N may then increase effective total attack damage by one, bypassing defense and any maximum damage caps.

@YCURRENT DIRECTION@N

In the status bar, an arrow shows you the direction of your last movement or attack. That direction is used in various contexts by some effects and abilities, like when eating firebreath pepper, sprinting or jumping.

@YATTACK PATTERNS@N

In Shamogu, both the player and the monsters may use one of several attack patterns.

@CPlain@N attacks are the simplest kind of attack performed by some monsters. Such monsters attack when they are adjacent to you or when they perform a @Mfrontal charge@N from an adjacent position.

@CFour-directional@N attacks affect all four adjacent foes. When charging, it affects foes adjacent to the new position. That means walking laterally does not save you from the charge of an adjacent monster with four-directional attacks! When fighting more than one foe at once, Four-Headed Hydra players’ attack attribute is increased by the number of foes.

@CRampaging@N attacks allow to charge from as far as you can see. When charging, there is a bonus of 1 to attack, except against lignified foes that resist forced movement.

@CPushing@N attacks push foes one tile away, unbalancing them. When there’s another foe just behind, you perform an extra attack on them instead. Rampaging boars, as well as Rampaging Boar spirit players, perform pushing when charging from beyond a 2-tile range, and distances beyond a 4-tile range result in double @OImbalance@N duration. Earth dragons and players with @BDig@N may push with melee attacks, too. @OPoison@N prevents pushing in melee.

@CRanged@N attacks are performed without moving from as far as you can see.

@CCatching@N attacks are ranged attacks that catch foes @Mon successful hits@N and move them next to you, unbalancing them. When catching a foe from afar, there is a bonus of 1 to attack. Catching from beyond a 4-tile range results in double @OImbalance@N duration.

@CRecoil@N is an extra attack effect for wind fox monsters, as well as Wind Fox spirit players. After an attack from within a 4-tile range, they may move backwards one tile when ranged, or two cells with extra wind noise when in melee.

@CSpace-Distorting@N attacks are ranged attacks that @Mswap positions@N on successful hits, but result in normal movement on misses. Any foes adjacent to the landing hit position are blinked away. Also, note that such attacks reverse your current direction.

@YNOISE@N

Combat and some other actions produce noise, as shown with onomatopoeias in the game logs. Some noises can be heard from a few tiles away only, while others may reach farther, up to two times the view range in the worst case. Possible noises include a gentle “Smash!”, a more serious “ROAR!” or “CRACK!”, a problematic “POP-BOOM!”, or an even more alarming “RING-RING!”. Beware that monsters will investigate the source of those noises.

Shamogu also shows visually the sound made by monsters out of view: @Rfootsteps@N, @Olight footsteps@N, @Gcreep noise@N, @Cflapping of wings@N, and @Mheavy footsteps@N. Note that monsters cannot hear your footsteps or, at the very least, they don’t care about such small noises!

Players and monsters with the “good hearing” trait may hear sounds farther than usual. Players with “bad hearing” cannot hear @Gcreep noises@N nor @Cflapping of wings@N well, but they hear others sounds normally.
`)
}

// ActionHelpItems opens the log message viewer.
type ActionHelpItems struct{}

func (a ActionHelpItems) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("Items (Help)", itemHelpText())
	return nil, false
}

func (a ActionHelpItems) String() string {
	return "Items"
}

func itemHelpText() string {
	return strings.TrimSpace(`
@CComestibles@N are found in every level. When the player is over them, pressing the @MInteract Key@N(e) allows to pick them. When your inventory is full, you’ll have to select a comestible to leave on the ground. Opening the @MInventory@N(i) allows to eat an equipped comestible or any comestible on the ground.

@CTotems@N are found once per level. By pressing the @MInteract Key@N(e), their spirit may either be chosen as a new secondary spirit or used to upgrade one of your existing spirits. Opening the @MInventory@N(i) allows then to use your chosen spirits. Beware that monsters that see you chosing a spirit will go berserk!

@CMenhirs@N are found in every level. Pressing the @MInteract Key@N(e) will activate them.

@CPortals@N are found once per level. Pressing the @MInteract Key@N(e) will make you reach the next map level. When doing so, spirit charges and HP are restored. Occasionally, you’ll come across an extra malfunctioning portal whose activation will frighten nearby monsters.

@CRunes@N are static traps found in every level. They are triggered when the player or a monster steps on them. Wandering monsters avoid them, but hunting ones do not.
`)
}

// ActionHelpTips opens the log message viewer.
type ActionHelpTips struct{}

func (a ActionHelpTips) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("Tips (Help)", tipsHelpText())
	return nil, false
}

func (a ActionHelpTips) String() string {
	return "Tips"
}

func tipsHelpText() string {
	return strings.TrimSpace(`
@YGENERAL TIPS@N

Shamogu is a @Mturn-based@N game: take your time when in perilous situations!

Always check your inventory to see if some spirit ability or comestible can help in your situation. The sooner is often the better!

Wasting spirit charges and comestibles is bad, but dying without using them is worse!

Most items and abilities have several effects and can be used for several purposes. For example, barking or freezing time may be useful both when attacking and fleeing, but freezing time can also help wait for some problematic status effect to end.

@YCOMBAT TIPS@N

Try to always get the @Mfirst hit@N against monsters. In particular, avoid moving yourself in front of ranged or rampaging monsters. Zigzag if necessary. Don’t forget that even plain melee monsters can charge frontally from one tile away!

Try to learn the characteristics of each kind of monster.

Some ranged monsters, like fire llamas or lashing frogs, are more dangerous ranged than in melee. You may sometimes want to force melee even if you’re ranged too!

Some monsters have special attacks, like acid mounds ignoring defense, or venomous vipers poisoning you.

@YSTEALTH TIPS@N

Killing all the monsters does not provide any benefit. So, once you have enough food items and don’t need more totems or already found one, you can skip exploring the rest of the map and go through a portal to the next level. Also, missing a totem or two is acceptable.

Monsters usually give up quickly their search once they don’t see you: use foliage and rubble to lose them.

Watch out for @Mfootstep noise@N: while going through dense foliage, it can be very helpful to avoid unwanted encounters.

@YWIZARD MODE@N

While the recommended way to discover the game is to progressively improve your skills by analyzing and enjoying your previous deaths, frustrated casual players should consider enabling @MWizard Mode@N(W) before giving up. Without further enabling extra cheats, it works as a kind of “adventure mode” that simply resurrects you each time you die. Resurrections are recorded in the game’s timeline, so you can check later how many lives you spent and try to do better next time until you can win in the default permadeath mode.
`)
}
