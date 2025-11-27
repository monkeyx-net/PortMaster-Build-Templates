package main

import (
	"log"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// menu represents menus used in the game.
type menu struct {
	keys  *ui.Menu // key settings menu
	main  *ui.Menu // typical main menu (game, settings, inventory)
	bound ID       // pre-selected option for ModGluttonyRework
	mode  menuMode
}

// menuMode represents the available menu modes
type menuMode int

const (
	modeInventory menuMode = iota
	modeInventoryBound
	modeEquip
	modeConfigMenu
	modeHelpMenu
	modeKeysView
	modeKeysChange
	modeGameMenu
	modeSelection
)

func (md *model) updateMenu(msg gruid.Msg) {
	switch md.menu.mode {
	case modeKeysView:
		md.updateKeysMenu(msg)
	case modeKeysChange:
		md.updateKeysChange(msg)
	default:
		md.updateMainMenu(msg)
	}
}

func (md *model) updateMainMenu(msg gruid.Msg) {
	md.menu.main.Update(msg)
	switch act := md.menu.main.Action(); act {
	case ui.MenuQuit:
		md.mode = modeNormal
	case ui.MenuMove, ui.MenuInvoke:
		idx := md.menu.main.ActiveInvokable()
		if idx < 0 {
			break
		}
		switch md.menu.mode {
		case modeGameMenu, modeConfigMenu, modeHelpMenu:
			md.updateMenuActionDesc(idx)
			if act != ui.MenuInvoke {
				break
			}
			md.action = md.menuActions[idx]
		case modeInventory:
			g := md.g
			id := md.inventoryItemID(idx)
			e := g.Entity(id)
			md.updateItemDesc(e)
			if act != ui.MenuInvoke {
				break
			}
			md.action = ActionUseItem{ID: id}
			if !g.Mod(ModGluttonyRework) || !g.PlayerActor().DoesAny(Gluttony) {
				break
			}
			if _, ok := e.Role.(*Comestible); ok {
				if len(g.ComestibleIDs()) > 1 {
					md.action = ActionBindItem{id}
				} else {
					md.mode = modeNormal
					md.g.Log("You don’t have a second comestible!")
					md.action = ActionNone{}
				}
			}
		case modeInventoryBound:
			// Choosing second comestible (for mod ModGluttonyRework).
			g := md.g
			idx := md.menu.main.Active()
			id := md.inventoryComestibleID(idx)
			e := g.Entity(id)
			md.updateItemDesc(e)
			if act != ui.MenuInvoke || id == md.menu.bound {
				break
			}
			md.action = ActionUseTwoItems{md.menu.bound, id}
		case modeEquip:
			g := md.g
			i, it := g.ItemAt(g.PP())
			if i < 0 {
				// Should not happen.
				md.mode = modeNormal
				md.g.LogStyled("No item on the ground (BUG)", logError)
				md.action = ActionNone{}
				break
			}
			j := md.equipItemID(it, idx)
			ei, ej := g.Entity(i), g.Entity(j)
			_, ok := ej.Role.(*Spirit)
			md.drawGroundDesc = !ok
			md.updateEquipItemDesc(ei, ej)
			if act != ui.MenuInvoke {
				break
			}
			md.action = ActionEquipItemAt{ID: j}
		}
	}
}

// updateMenuActionDesc updates the description label for the menu action of
// given index in the current menu.
func (md *model) updateMenuActionDesc(idx int) {
	if idx < 0 || idx >= len(md.menuActions) {
		md.desc.Content = ui.StyledText{}
		return
	}
	a := md.menuActions[idx]
	if ad, ok := a.(ActionDesc); ok {
		md.updateActionDesc(ad)
	} else {
		md.desc.Content = ui.StyledText{}
	}
}

// equipItemID returns the entity ID corresponding to the given equipable
// inventory item equip index.
func (md *model) equipItemID(it Item, idx int) ID {
	switch it.(type) {
	case *Spirit:
		return md.equipSpiritID(idx)
	default:
		// Comestible.
		return ID(idx) + NSpirits + PrimarySpiritID
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
		return ID(j) + PrimarySpiritID
	}
	id, _ := g.ItemAt(g.PP())
	return id // should be positive
}

// inventoryComestibleID returns the entity ID corresponding to the given
// inventory raw index.
func (md *model) inventoryComestibleID(i int) ID {
	// NOTE: this function assumes there are 3 headers before the first
	// comestible in the inventory menu; if that ever changes, this should
	// be updated.
	if n := ID(i - 3); n < FirstMapID {
		return n
	}
	g := md.g
	id, _ := g.ItemAt(g.PP())
	return id // should be positive
}

func (md *model) updateKeysMenu(msg gruid.Msg) {
	md.menu.keys.Update(msg)
	switch act := md.menu.keys.Action(); act {
	case ui.MenuQuit:
		md.mode = modeNormal
	case ui.MenuInvoke:
		md.menu.mode = modeKeysChange
	case ui.MenuPass:
		msg, ok := msg.(gruid.MsgKeyDown)
		if !ok {
			return
		}
		if msg.Key == "R" {
			md.resetKeys()
			md.action = ActionSetKeys{}
		}
	}
}

func (md *model) resetKeys() {
	md.initKeys()
	GameConfig.NormalModeKeys = md.keysNormal
	GameConfig.ExamineModeKeys = md.keysTarget
	err := SaveConfig()
	if err != nil {
		log.Printf("error resetting config changes: %v", err)
		md.g.LogStyled("Error while resetting config changes.", logError)
	}
}

func (md *model) updateKeysChange(msg gruid.Msg) {
	mkd, ok := msg.(gruid.MsgKeyDown)
	if !ok {
		return
	}
	key := mkd.Key
	switch key {
	case gruid.KeyEscape, gruid.KeySpace:
		md.action = ActionSetKeys{}
	default:
		action := configurableKeyActions[md.menu.keys.ActiveInvokable()]
		if KbTargetingMode(action) {
			md.keysTarget[key] = action
		} else {
			md.keysNormal[key] = action
		}
		GameConfig.NormalModeKeys = md.keysNormal
		GameConfig.ExamineModeKeys = md.keysTarget
		err := SaveConfig()
		if err != nil {
			log.Printf("error saving config changes: %v", err)
			md.g.LogStyled("Error while saving config changes.", logError)
		}
		md.action = ActionSetKeys{}
	}
}

func (md *model) updateSpiritSelectionMenu(msg gruid.Msg) gruid.Effect {
	if msgmouse, ok := msg.(gruid.MsgMouse); ok {
		msgmouse.P.Y -= spiritSelectionMargin
		msg = msgmouse
	}
	if msgkey, ok := msg.(gruid.MsgKeyDown); ok && msgkey.Key == gruid.KeyTab {
		if md.mode == modeNewGame && len(GameConfig.Mods) == NMods {
			md.g.Mods = slices.Clone(GameConfig.Mods)
		}
		md.modSelectionMenu()
		return nil
	}
	md.menu.main.Update(msg)
	switch act := md.menu.main.Action(); act {
	case ui.MenuQuit:
		return gruid.End()
	case ui.MenuMove, ui.MenuInvoke:
		idx := md.menu.main.ActiveInvokable()
		if idx < 0 {
			break
		}
		spirits := getPrimarySpirits(md.mode)
		spe := spiritEntity(md.g.Mods, spirits[idx])
		md.updateItemDesc(spe)
		if act != ui.MenuInvoke {
			break
		}
		md.startNewGame(spe)
	}
	return nil
}

func (md *model) startNewGame(spe *Entity) {
	md.g.Init(spe)
	md.g.ComputePlayerStats()
	md.g.InitLevel()
	md.updateStatus()
	md.g.LogStyled("Press SPACE for menu and ? for help. Good luck!", logSpecial)
	md.mode = modeNormal
}

func (md *model) updateModSelectionMenu(msg gruid.Msg) gruid.Effect {
	if msgkey, ok := msg.(gruid.MsgKeyDown); ok {
		switch msgkey.Key {
		case gruid.KeyEscape:
			if GameConfig.AdvancedNewGame {
				md.openSpiritSelectionMenu(modeNewGameAdvanced)
			} else {
				md.openSpiritSelectionMenu(modeNewGame)
				clear(md.g.Mods)
			}
			return nil
		case gruid.KeySpace:
			md.openSpiritSelectionMenu(modeNewGameAdvanced)
			return nil
		}
	}
	md.menu.main.Update(msg)
	switch act := md.menu.main.Action(); act {
	case ui.MenuQuit:
		return gruid.End()
	case ui.MenuMove, ui.MenuInvoke:
		idx := md.menu.main.ActiveInvokable()
		if idx < 0 {
			break
		}
		g := md.g
		if idx == len(g.Mods) {
			l := md.desc
			l.Box = &ui.Box{Title: ui.Text("Save and continue")}
			l.Content = ui.Text("Save mod selection and continue to advanced primary spirit selection screen").
				Format(UIWidth/2 - 1 - 2)
			if act == ui.MenuInvoke {
				GameConfig.Mods = slices.Clone(g.Mods)
				if err := SaveConfig(); err != nil {
					md.g.Logf("Error saving mod config changes: %v", err)
				}
				md.openSpiritSelectionMenu(modeNewGameAdvanced)
			}
			break
		}
		if idx > len(g.Mods) {
			l := md.desc
			l.Box = &ui.Box{Title: ui.Text("Quit menu and disable mods")}
			l.Content = ui.Text("Disable any mods and return to normal spirit selection menu.").
				Format(UIWidth/2 - 1 - 2)
			if act == ui.MenuInvoke {
				md.openSpiritSelectionMenu(modeNewGame)
				clear(md.g.Mods)
			}
			break
		}
		m := gameMods[idx]
		if act != ui.MenuInvoke {
			md.updateModDesc(m)
			break
		}
		g.Mods[m] = !g.Mods[m]
		md.modSelectionMenu()
		md.menu.main.SetActiveInvokable(idx)
		md.updateModDesc(m)
	}
	return nil
}

// updateModDesc updates the description label with text for the given entity.
func (md *model) updateModDesc(m Mod) {
	l := md.desc
	stt := ui.StyledText{}.WithMarkups(Markups)
	enabled := ""
	if md.g.Mods[m] {
		enabled = "@CEnabled.@N\n"
	}
	l.Box = &ui.Box{Title: ui.Text(m.String())}
	l.Content = stt.WithText(enabled + m.Desc()).Format(UIWidth/2 - 2)
}

func (md *model) modSelectionMenu() {
	hstyle := gruid.Style{Fg: ColorCyan}
	entries := []ui.MenuEntry{}
	mods := gameMods
	title := "Mod Selection"
	r := 'a'
	for i, m := range mods {
		switch {
		case i == 0:
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Expansions").WithStyle(hstyle),
				Disabled: true,
			})
		case i == 2:
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Small Mods").WithStyle(hstyle),
				Disabled: true,
			})
		case i == NExpansions:
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Challenges").WithStyle(hstyle),
				Disabled: true,
			})
		}
		s := "[ ]"
		if md.g.Mods[Mod(i)] {
			s = "[*]"
		}
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %-30s %s", r, m.String(), s),
			Keys: []gruid.Key{gruid.Key(r)},
		})
		r++
		for {
			switch r {
			case 's', 'q', 'h', 'j', 'k', 'l':
				r++
				continue
			}
			break
		}
	}
	entries = append(entries, ui.MenuEntry{
		Text:     ui.Text("Actions").WithStyle(hstyle),
		Disabled: true,
	})
	entries = append(entries, ui.MenuEntry{
		Text: ui.Text("s - Save and continue"),
		Keys: []gruid.Key{gruid.Key('s'), gruid.Key('S')},
	})
	entries = append(entries, ui.MenuEntry{
		Text: ui.Text("q - Quit menu and disable mods"),
		Keys: []gruid.Key{gruid.Key('q'), gruid.Key('Q')},
	})
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.updateModDesc(Mod(0))
	md.mode = modeNewGameMods
	md.menu.mode = modeSelection // not really needed
}

// NOTE: Old index-computation code below, kept here in case we ever need it
// again.
//
// equipItemIdx returns the item equip index corresponding to the given
// equipped entity ID.
// func (md *model) equipItemIdx(i ID) int {
// 	g := md.g
// 	switch g.Entity(i).Role.(type) {
// 	case *Spirit:
// 		for _, e := range g.Entities[:i] {
// 			if sp, ok := e.Role.(*Spirit); !ok || sp.Level == 1 {
// 				i--
// 			}
// 		}
// 		return int(i - PrimarySpiritID)
// 	default:
// 		// Comestible.
// 		return int(i - NSpirits - PrimarySpiritID)
// 	}
// }
//
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
