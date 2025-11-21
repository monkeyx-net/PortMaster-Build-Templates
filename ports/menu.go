package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// menu represents menus used in the game.
type menu struct {
	keys *ui.Menu // key settings menu
	main *ui.Menu // typical main menu (game, settings, inventory)
	mode menuMode
}

// menuMode represents the available menu modes
type menuMode int

const (
	modeInventory menuMode = iota
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
	g := md.g
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
			if act != ui.MenuInvoke {
				break
			}
			md.action = md.menuActions[idx]
		case modeInventory:
			id := md.inventoryItemID(idx)
			e := g.Entity(id)
			md.updateItemDesc(e)
			if act != ui.MenuInvoke {
				break
			}
			md.action = ActionUseItem{ID: id}
		case modeEquip:
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
	md.menu.main.Update(msg)
	g := md.g
	switch act := md.menu.main.Action(); act {
	case ui.MenuQuit:
		return gruid.End()
	case ui.MenuMove, ui.MenuInvoke:
		idx := md.menu.main.ActiveInvokable()
		if idx < 0 {
			break
		}
		e := spiritEntity(primarySpirits[idx])
		md.updateItemDesc(e)
		if act != ui.MenuInvoke {
			break
		}
		g.Entities[PrimarySpiritID] = e
		g.ComputePlayerStats()
		md.g.InitLevel()
		md.updateStatus()
		g.LogStyled("Press SPACE for menu and ? for help. Good luck!", logSpecial)
		md.mode = modeNormal
	}
	return nil
}
