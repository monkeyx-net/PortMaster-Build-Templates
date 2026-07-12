package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// actionErrorUnknown represents an unknown action (mostly for incorrect
// keypresses).
type actionErrorUnknown struct{}

func (e actionErrorUnknown) Error() string {
	return "unknown action"
}

// Action represents the various kinds of UI actions.
type Action int

const (
	ActionNone Action = iota
	ActionW
	ActionS
	ActionN
	ActionE
	ActionRunW
	ActionRunS
	ActionRunN
	ActionRunE
	ActionWaitTurn
	ActionDescend
	ActionExplore
	ActionExamineToggle
	ActionExamineScrollUp
	ActionExamineScrollDown
	ActionEvoke
	ActionInteract
	ActionInventory
	ActionLogs
	ActionDump
	ActionHelpMenu
	ActionHelpKeybindings
	ActionHelpBasics
	ActionHelpItems
	ActionHelpStatuses
	ActionHelpTips
	ActionSave
	ActionQuit
	ActionWizard
	ActionWizardDescend

	ActionPreviousMonster
	ActionNextMonster
	ActionNextObject
	ActionTravel

	ActionSettings
	ActionMenu
	ActionNextStairs

	ActionSetKeys
	ActionInvertTheme
	ActionToggleTiles
	ActionToggleShowNumbers

	ActionWizardToggleMode
)

// ConfigurableKeyActions contains configurable UI actions.
var ConfigurableKeyActions = [...]Action{
	ActionW,
	ActionS,
	ActionN,
	ActionE,
	ActionRunW,
	ActionRunS,
	ActionRunN,
	ActionRunE,
	ActionWaitTurn,
	ActionEvoke,
	ActionInteract,
	ActionInventory,
	ActionExamineToggle,
	ActionExamineScrollUp,
	ActionExamineScrollDown,
	ActionLogs,
	ActionPreviousMonster,
	ActionNextMonster,
	ActionNextObject,
	ActionNextStairs,
	ActionTravel,
	ActionExplore,
	ActionDump,
	ActionSave,
	ActionQuit,
	ActionMenu,
	ActionWizard,
}

func (k Action) String() (text string) {
	switch k {
	case ActionW:
		text = "Move west"
	case ActionS:
		text = "Move south"
	case ActionN:
		text = "Move north"
	case ActionE:
		text = "Move east"
	case ActionRunW:
		text = "Travel west"
	case ActionRunS:
		text = "Travel south"
	case ActionRunN:
		text = "Travel north"
	case ActionRunE:
		text = "Travel east"
	case ActionWaitTurn:
		text = "Wait a turn"
	case ActionDescend:
		text = "Descend stairs"
	case ActionExamineToggle:
		text = "Examine"
	case ActionExamineScrollUp:
		text = "Scroll description boxes up"
	case ActionExamineScrollDown:
		text = "Scroll description boxes down"
	case ActionEvoke:
		text = "Evoke card"
	case ActionInteract:
		text = "Interact"
	case ActionInventory:
		text = "Inventory"
	case ActionLogs:
		text = "View messages"
	case ActionPreviousMonster:
		text = "Target previous monster"
	case ActionNextMonster:
		text = "Target next monster"
	case ActionNextObject:
		text = "Target next stone/magara/item"
	case ActionNextStairs:
		text = "Target next stairs"
	case ActionTravel:
		text = "Go to"
	case ActionExplore:
		text = "Autoexplore"
	case ActionDump:
		text = "Dump game statistics"
	case ActionSave:
		text = "Save and quit"
	case ActionQuit:
		text = "Quit (without saving)"
	case ActionHelpMenu:
		text = "Help Menu"
	case ActionHelpKeybindings:
		text = "Keybindings (default values)"
	case ActionHelpBasics:
		text = "Basics"
	case ActionHelpItems:
		text = "Items"
	case ActionHelpStatuses:
		text = "Statuses"
	case ActionHelpTips:
		text = "Tips"
	case ActionSettings:
		text = "Configure settings"
	case ActionWizard:
		text = "Wizard mode (irreversible)"
	case ActionMenu:
		text = "Action Menu"
	case ActionSetKeys:
		text = "Change key bindings"
	case ActionInvertTheme:
		text = "Toggle dark/light color theme"
	case ActionToggleTiles:
		text = "Toggle tiles/ASCII display"
	case ActionToggleShowNumbers:
		text = "Toggle hearts/numbers"
	case ActionWizardToggleMode:
		text = "toggle normal/map/all wizard mode"
	}
	return text
}

// Desc returns a detailed description for some actions.
func (k Action) Desc() (text string) {
	switch k {
	case ActionLogs:
		text = "Opens a pager with previous message logs. The pager supports page up/down, mouse scrolling, and other basic less-like keybindings."
	case ActionDump:
		if runtime.GOOS == "js" {
			text = "Writes game statistics below."
		} else {
			text = "Writes game statistics to a dump.txt file in the game’s data directory."
		}
	case ActionSave:
		text = "Saves current progress and quits the game. The next time you start the game, it will directly resume from here."
	case ActionQuit:
		text = "Deletes any saved progress for current playthrough and quits the game."
	case ActionHelpMenu:
		text = "Opens a menu leading to various help topics."
	case ActionHelpKeybindings:
		text = "Shows a short one-page summary with most default keybindings."
	case ActionHelpBasics:
		text = "Explains various core aspects of the game, in particular stealth."
	case ActionHelpItems:
		text = "Gives an overview about most fundamental items found in the game."
	case ActionHelpStatuses:
		text = "Explains in detail how the various status effects work, as well as the color conventions."
	case ActionHelpTips:
		text = "Provides various tips for new players."
	case ActionSettings:
		text = "Opens a configuration menu with various options."
	}
	return text
}

// updateActionDesc updates the description label for the given described action.
func (md *model) updateActionDesc(a Action) {
	l := md.description
	stt := ui.StyledText{}.WithMarkups(Markups)
	desc := a.Desc()
	l.Box = &ui.Box{Title: ui.Text(a.String())}
	if desc != "" {
		l.Content = stt.WithText(desc).Format(UIWidth/2 - 2)
	} else {
		l.Content = stt.WithText("")
	}
}

// ExamineModeAction reports whether the action is specific to examine mode.
func (k Action) ExamineModeAction() bool {
	switch k {
	case ActionTravel:
		return true
	default:
		return false
	}
}

// CanInteract reports whether the player can interact with current cell, along
// with a description of the interaction.
func (md *model) CanInteract() (string, bool) {
	g := md.g
	t := g.Map.At(g.Player.P)
	switch t {
	case StairCell:
		return "descend", true
	case BarrelCell:
		return "rest", true
	case MagaraCell:
		return "equip magara", true
	case StoneCell:
		return "activate stone", true
	case ScrollCell:
		return "read scroll", true
	case ItemCell:
		return "equip item", true
	case LightCell:
		return "extinguish light", true
	case StoryCell:
		if g.Objects.Story[g.Player.P] == StoryArtifact && !g.LiberatedArtifact ||
			g.Objects.Story[g.Player.P] == StoryArtifactSealed {
			return "take artifact", true
		}
		return "", false
	}
	return "", false
}

// DoAction performs the given action. It reports whether the turn wasn't spent
// (again), and may return an in-game error that needs to be logged.
//
// TODO: We probably would want to just return a string message instead of an
// error and adjust again, as in-game errors aren't really errors. Maybe not
// worth the effort, though.
func (md *model) DoAction(action Action) (again bool, eff gruid.Effect, err error) {
	g := md.g
	switch action {
	case ActionNone:
		again = true
		err = actionErrorUnknown{}
	case ActionW, ActionS, ActionN, ActionE:
		if !md.targ.kbTargeting {
			again, err = g.PlayerBump(g.Player.P.Add(KeyToDir(action)))
		} else {
			p := md.targ.ex.p.Add(KeyToDir(action))
			if InMap(p) {
				md.Examine(p)
			}
			again = true
		}
	case ActionRunW, ActionRunS, ActionRunN, ActionRunE:
		if !md.targ.kbTargeting {
			again, err = g.GoToDir(KeyToDir(action))
		} else {
			q := invalidPos
			p := md.targ.ex.p
			for range 5 {
				p = p.Add(KeyToDir(action))
				if !InMap(p) {
					break
				}
				q = p
			}
			if q != invalidPos {
				md.Examine(q)
			}
			again = true
		}
	case ActionPreviousMonster:
		again = true
		md.NextMonster("-")
	case ActionNextMonster:
		again = true
		md.NextMonster("+")
	case ActionNextObject:
		again = true
		md.NextObject(md.targ.ex.p)
	case ActionNextStairs:
		again = true
		md.NextStair()
	case ActionTravel:
		again = true
		err = md.UpdateTarget()
		if err != nil {
			break
		}
		if g.MoveToTarget() {
			again = false
		}
	case ActionWaitTurn:
		g.WaitTurn()
	case ActionInteract:
		t := g.Map.At(g.Player.P)
		switch t {
		case StairCell:
			switch g.Objects.Stairs[g.Player.P] {
			case NormalStair, WinStair:
				strt := g.Objects.Stairs[g.Player.P]
				err = md.checkShaedra(strt)
				if err != nil {
					break
				}
				again = true
				if g.Descend(DescendNormal) {
					md.win()
				}
			case SealedStair:
				err = errors.New("The stairs are blocked by an oric magical barrier.")
			default:
				err = errors.New("You cannot use fake stairs made of harmonic illusions!")
			}
		case BarrelCell:
			err = g.Rest()
		case MagaraCell:
			again = true
			md.equipMagaraMenu()
		case StoneCell:
			err = g.ActivateStone()
		case ScrollCell:
			again = true
			md.readScroll()
		case ItemCell:
			err = md.g.EquipItem()
		case LightCell:
			err = g.ExtinguishFire()
		case StoryCell:
			if g.Objects.Story[g.Player.P] == StoryArtifact && !g.LiberatedArtifact {
				// We liberate the artifact and proceed to story
				// sequence.
				g.LogStyled("You quickly take and activate the artifact!", logSpecial)
				g.LogStyled("[SPACE/ESC to continue]", logConfirm)
				g.LiberatedArtifact = true
				g.md.mode = modeStory
				g.md.story = 1
				// We make the action instant to avoid any
				// surprises and to make the logs easier to
				// follow, even if it is a little bit less
				// realistic. There's still a chance of dying
				// in the next turn, so it doesn't completely
				// guarantee a win.
				again = true
			} else if g.Objects.Story[g.Player.P] == StoryArtifactSealed {
				err = errors.New("The artifact is protected by a magical stone barrier.")
			} else {
				err = errors.New("You cannot interact with anything here.")
			}
		default:
			err = errors.New("You cannot interact with anything here.")
		}
	case ActionEvoke:
		again = true
		md.evokeMagaraMenu()
	case ActionInventory:
		again = true
		md.openIventory()
	case ActionExplore:
		again, err = g.Autoexplore()
	case ActionExamineToggle:
		again = true
		if md.targ.kbTargeting {
			md.CancelExamine()
		} else {
			md.KeyboardExamine()
		}
	case ActionExamineScrollUp:
		again = true
		md.targ.ex.scroll = 0
	case ActionExamineScrollDown:
		again = true
		md.targ.ex.scroll++
	case ActionHelpMenu:
		again = true
		md.openHelp()
	case ActionMenu:
		again = true
		md.openMenu()
	case ActionHelpKeybindings:
		again = true
		md.keysHelp()
	case ActionHelpBasics:
		again = true
		md.helpBasics()
	case ActionHelpItems:
		again = true
		md.helpItems()
	case ActionHelpStatuses:
		again = true
		md.helpStatuses()
	case ActionHelpTips:
		again = true
		md.helpTips()
	case ActionLogs:
		again = true
		if len(md.logs) > 0 {
			md.logs = md.logs[:len(md.logs)-1]
		}
		for _, e := range g.Logs[len(md.logs):] {
			md.logs = append(md.logs, md.pagerMarkup.WithText(e.MText))
		}
		md.pager.SetLines(md.logs)
		md.pager.SetCursor(gruid.Point{0, len(md.logs)})
		md.pager.SetBox(&ui.Box{Title: ui.Text("Messages").WithStyle(gruid.Style{Fg: ColorYellow})})
		md.mode = modePager
		md.pagerMode = modeLogs
	case ActionSave:
		again = true
		g.checks()
		errsave := g.Save()
		if errsave != nil {
			log.Printf("error saving game: %v", errsave)
			md.mode = modeNormal
		} else {
			md.mode = modeQuit
			eff = gruid.End()
		}
	case ActionDump:
		again = true
		msg, errdump := g.WriteDump()
		if errdump != nil {
			log.Printf("error writing dump: %v", errdump)
			g.LogStyled("Error writing statistics.", logError)
		} else {
			g.Log(msg)
		}
		md.mode = modeNormal
	case ActionWizardDescend:
		if g.Wizard != WizardNone && g.Map.Depth == WinDepth && !g.LiberatedShaedra {
			g.LiberatedShaedra = true
			g.RescuedShaedra()
		}
		if g.Wizard != WizardNone && g.Map.Depth < MaxDepth {
			g.StoryPrint("Descended wizardly")
			again = true
			if g.Descend(DescendNormal) {
				md.win()
			}
		} else {
			err = actionErrorUnknown{}
		}
	case ActionWizard:
		again = true
		md.g.md.CancelExamine()
		switch md.g.Wizard {
		case WizardNone:
			md.g.LogStyled("Do you really want to enter Wizard Mode ? (debug/cheat mode) [y/N]", logConfirm)
			md.mode = modeWizardConfirmation
		case WizardNormal:
			md.g.Wizard = WizardSeeAll
		case WizardSeeAll:
			md.g.Wizard = WizardMap
		default:
			md.g.Wizard = WizardNormal
		}
	case ActionQuit:
		again = true
		md.Quit()
	case ActionSettings:
		again = true
		md.openSettings()
	case ActionSetKeys:
		again = true
		md.openKeyBindings()
	case ActionInvertTheme:
		again = true
		GameConfig.DarkColors = !GameConfig.DarkColors
		err := SaveConfig()
		if err != nil {
			log.Print(err)
		}
		clearCache()
		eff = gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
		md.mode = modeNormal
	case ActionToggleTiles:
		again = true
		md.ApplyToggleTiles()
		eff = gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
		err := SaveConfig()
		if err != nil {
			log.Print(err)
		}
		md.mode = modeNormal
	case ActionToggleShowNumbers:
		again = true
		GameConfig.ShowNumbers = !GameConfig.ShowNumbers
		err := SaveConfig()
		if err != nil {
			log.Print(err)
		}
		md.updateStatusInfo()
		md.mode = modeNormal
	case ActionWizardToggleMode:
		switch g.Wizard {
		case WizardNormal:
			g.Wizard = WizardMap
		case WizardMap:
			g.Wizard = WizardSeeAll
		case WizardSeeAll:
			g.Wizard = WizardNormal
		}
		md.mode = modeNormal
	default:
		err = actionErrorUnknown{}
	}
	if err != nil {
		again = true
	}
	return again, eff, err
}

// readScroll perfroms scroll reading.
func (md *model) readScroll() {
	sc, ok := md.g.Objects.Scrolls[md.g.Player.P]
	if !ok {
		// Should not happen.
		md.g.LogStyled("Error while reading message.", logError)
		return
	}
	md.g.Log("You read the message.")
	md.mode = modeSmallPager
	st := gruid.Style{}
	switch sc {
	case ScrollLore:
		md.smallPager.SetBox(&ui.Box{Title: ui.Text("Lore Message").WithStyle(st.WithFg(ColorViolet))})
		stts := []ui.StyledText{}
		text := ui.Text(sc.Text(md.g)).Format(56)
		for s := range strings.SplitSeq(text.Text(), "\n") {
			stts = append(stts, ui.Text(s))
		}
		md.smallPager.SetLines(stts)
		md.smallPager.SetCursor(gruid.Point{0, 0})
		if !md.g.Stats.Lore[md.g.Map.Depth] {
			md.g.StoryPrint("Read lore message")
		}
		md.g.Stats.Lore[md.g.Map.Depth] = true
		if len(md.g.Stats.Lore) == 4 {
			AchLoreStudent.Get(md.g)
		}
		if len(md.g.Stats.Lore) == len(md.g.Params.Lore) {
			AchLoremaster.Get(md.g)
		}
	default:
		md.smallPager.SetBox(&ui.Box{Title: ui.Text("Story Message").WithStyle(st.WithFg(ColorCyan))})
		stts := []ui.StyledText{}
		text := ui.Text(sc.Text(md.g)).Format(56)
		for s := range strings.SplitSeq(text.Text(), "\n") {
			stts = append(stts, ui.Text(s))
		}
		md.smallPager.SetLines(stts)
		md.smallPager.SetCursor(gruid.Point{0, 0})
	}
}

// altBgEntries alternates background color for a list of menu entries.
func altBgEntries(entries []ui.MenuEntry) {
	for i := range entries {
		if i%2 == 1 {
			st := entries[i].Text.Style()
			entries[i].Text = entries[i].Text.WithStyle(st.WithBg(ColorBackgroundSecondary))
		}
	}
}

// settingsActions represents the various configuration actions.
var settingsActions = []Action{
	ActionSetKeys,
	ActionInvertTheme,
	ActionToggleShowNumbers,
}

var settingsKeys = []rune{':', 'c', 'n'}

func (md *model) openSettings() {
	entries := []ui.MenuEntry{}
	for i, it := range settingsActions {
		r := settingsKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Settings").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.updateActionDesc(settingsActions[0])
	md.menuMode = modeSettings
}

func (md *model) keysForAction(a Action) string {
	keys := []gruid.Key{}
	for k, action := range md.keysNormal {
		if k == " " {
			k = "SPACE"
		}
		if a == action && !k.In(keys) {
			keys = append(keys, k)
		}
	}
	for k, action := range md.keysTarget {
		if k == " " {
			k = "SPACE"
		}
		if a == action && !k.In(keys) {
			keys = append(keys, k)
		}
	}
	ks := make([]string, len(keys))
	for i := range ks {
		ks[i] = string(keys[i])
	}
	sort.Strings(ks)
	return strings.Join(ks, " or ")
}

func (md *model) openKeyBindings() {
	entries := []ui.MenuEntry{}
	for _, it := range ConfigurableKeyActions {
		desc := it.String()
		desc = fmt.Sprintf(" %-36s %s", desc, md.keysForAction(it))
		entries = append(entries, ui.MenuEntry{
			Text: ui.Text(desc),
		})
	}
	altBgEntries(entries)
	md.keysMenu.SetBox(&ui.Box{Title: ui.Text("Key Bindings").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.keysMenu.SetEntries(entries)
	md.keysMenu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menuMode = modeKeys
}

// menuActions represents main game menu action.
var menuActions = []Action{
	ActionLogs,
	ActionHelpMenu,
	ActionDump,
	ActionSettings,
	ActionSave,
	ActionQuit,
}

var menuKeys = []rune{'m', '?', '#', 'C', 'S', 'Q'}

func (md *model) openMenu() {
	hstyle := gruid.Style{Fg: ColorCyan}
	entries := []ui.MenuEntry{}
	for i, it := range menuActions {
		r := menuKeys[i]
		switch r {
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
		}
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Menu").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.updateActionDesc(menuActions[0])
	md.menuMode = modeGameMenu
}

func (md *model) openIventory() {
	entries := []ui.MenuEntry{}
	items := []Item{md.g.Player.Inventory.Body, md.g.Player.Inventory.Neck, md.g.Player.Inventory.Misc}
	parts := []string{"body", "neck", "backpack"}
	r := 'a'
	for i, it := range items {
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s (%s)", r, it.ShortDesc(md.g), parts[i]),
			Keys: []gruid.Key{gruid.Key(r)},
		})
		r++
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Inventory").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menuMode = modeInventory
	it := items[0]
	stt := ui.StyledText{}.WithMarkups(Markups)
	md.description.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
	md.description.Box = &ui.Box{Title: stt.WithText(it.String())}
}

func (md *model) evokeMagaraMenu() {
	entries := []ui.MenuEntry{}
	items := md.g.Player.Magaras
	r := 'a'
	stt := ui.StyledText{}.WithMarkups(Markups)
	for _, it := range items {
		entries = append(entries, ui.MenuEntry{
			Text: stt.WithTextf("%c - %s ", r, it.ShortDesc()),
			Keys: []gruid.Key{gruid.Key(r)},
		})
		r++
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Evoke").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menuMode = modeEvocation
	it := items[0]
	md.description.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
	md.description.Box = &ui.Box{Title: stt.WithText(it.String())}
}

func (md *model) equipMagaraMenu() {
	entries := []ui.MenuEntry{}
	items := md.g.Player.Magaras
	r := 'a'
	for _, it := range items {
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s ", r, it.ShortDesc()),
			Keys: []gruid.Key{gruid.Key(r)},
		})
		r++
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Replace which magara?").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menuMode = modeEquip
	it := items[0]
	md.description.Content = ui.Text(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
	md.description.Box = &ui.Box{Title: ui.Textf("%s (equipped)", it.String())}
}

// helpActions represents main game menu action.
var helpActions = []Action{
	ActionHelpKeybindings,
	ActionHelpBasics,
	ActionHelpItems,
	ActionHelpStatuses,
	ActionHelpTips,
}

var helpKeys = []rune{'?', 'b', 'i', 's', 't'}

func (md *model) openHelp() {
	entries := []ui.MenuEntry{}
	for i, it := range helpActions {
		r := helpKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.SetBox(&ui.Box{Title: ui.Text("Help Menu").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.SetEntries(entries)
	md.menu.SetActiveInvokable(0)
	md.mode = modeMenu
	md.updateActionDesc(helpActions[0])
	md.menuMode = modeHelpMenu
}

func (md *model) keysHelp() {
	entries := []string{
		"Basic Game Actions", "",
		"Move/Jump", "arrows or hjkl or mouse left",
		"Wait a turn", "“.” or ENTER or mouse left on player",
		"Interact (Equip/Descend/Rest...)", "e",
		"Evoke/Zap magara", "v or z",
		"Inventory", "i",
		"Toggle keyboard examine mode", "x",
		"Open menu", "SPACE or mouse right",
		"Close menu, inventory...", "SPACE or ESC or mouse left outside",
		"Advanced Game Actions", "",
		"View previous messages", "m",
		"Scroll description boxes up/down", "u/d or page up/down or mouse scroll",
		"Examine previous/next monster", "- +",
		"Examine next stair/magical item", "> n",
		"Run (auto-move in direction)", "shift+arrows or HJKL",
		"Travel (auto-move to destination)", "“.” or ENTER in keyboard examine mode",
		"Autoexplore (use with caution)", "o",
		"Other Common Actions", "",
		"Save and quit", "S",
		"Quit (without saving)", "Q",
	}
	md.updateKeysDescription("Keybindings (default values)", entries)
}

func (md *model) updateKeysDescription(title string, actions []string) {
	md.pagerMode = modeHelpKeys
	md.mode = modePager
	if CustomKeys {
		title = fmt.Sprintf(" Default %s ", title)
	} else {
		title = fmt.Sprintf(" %s ", title)
	}
	md.pager.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{Fg: ColorYellow})})
	lines := []ui.StyledText{}
	for i := 0; i < len(actions)-1; i += 2 {
		stt := ui.StyledText{}
		if actions[i+1] != "" {
			stt = stt.WithTextf(" %-36s %s", actions[i], actions[i+1])
		} else {
			stt = stt.WithTextf(" %s ", actions[i]).WithStyle(gruid.Style{Fg: ColorCyan})
		}
		if i%4 == 2 {
			stt = stt.WithStyle(stt.Style().WithBg(ColorBackgroundSecondary))
		}
		lines = append(lines, stt)
	}
	md.pager.SetLines(lines)
	md.pager.SetCursor(gruid.Point{0, 0})
}

// Markups contains the styling markup-characters we use for StyledText.
var Markups = map[rune]gruid.Style{
	'B': {Fg: ColorBlue},
	'C': {Fg: ColorCyan},
	'G': {Fg: ColorGreen},
	'M': {Fg: ColorMagenta},
	'O': {Fg: ColorOrange},
	'R': {Fg: ColorRed},
	'V': {Fg: ColorViolet}, // unused
	'Y': {Fg: ColorYellow},
	'b': {Fg: ColorBlue, Attrs: AttrInMap},
	'c': {Fg: ColorCyan, Attrs: AttrInMap},
	'g': {Fg: ColorGreen, Attrs: AttrInMap},
	'm': {Fg: ColorMagenta, Attrs: AttrInMap},
	'o': {Fg: ColorOrange, Attrs: AttrInMap},
	'r': {Fg: ColorRed, Attrs: AttrInMap},
	'v': {Fg: ColorViolet, Attrs: AttrInMap}, // unused
	'y': {Fg: ColorYellow, Attrs: AttrInMap},
}

// helpTopic opens the pager with the given title and text content.
func (md *model) helpTopic(title, content string) {
	text := ui.Text(content).WithMarkups(Markups).Format(78)
	stts := text.Lines()
	md.pager.SetLines(stts)
	md.pager.SetCursor(gruid.Point{0, 0})
	md.mode = modePager
	md.pagerMode = modeHelpKeys
	md.mode = modePager
	md.pager.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{Fg: ColorYellow})})
}

func (md *model) helpBasics() {
	s := strings.TrimSpace(`
@YTURN-BASED SYSTEM@N

Harmonist is a turn-based game. Each actor acts once per turn: the player acts first, then the monsters do. Monsters act one after another in a non-predictable order, but except during some animations, you’ll only see the final state and often feel they acted all at once.

@YFIELD OF VIEW@N

In Harmonist, the player and the monsters do not share the same field of view.

The player’s field of view is multi-directional and has longer range.

All monsters, except for spiders, have a @Cdirectional@N conic field of view: they only look in a given direction. The game shows that area using an @Oorange@N color. If you examine a specific monster, only its field of view will be shown.

In both cases, visibility range is impacted by @Clighting@N: lighted @Yyellow@N tiles are visible from a greater range. Lighting may come from static campfires as well as kerejats: be careful when stepping into lighted terrain! Note that in various kinds of terrains, like when hiding under a table or at the top of a tree, the player can avoid the lighting. Also, tiles with a magical stone are always illuminated.

@YSTEALTH@N

Monsters will notice the player if you enter their field of view. In such cases, monsters start @Rhunting@N you. If they lose sight of you, they will give up after searching around for a bit.

Beware that, while most monsters can only hurt you when adjacent, some monsters have special ranged attacks.

@YNOISE@N

Some effects produce noise that can be heard by monsters, from more or less far away depending on the noise’s nature. Things like nadre explosions are heard from quite far away, for example.

Harmonist also shows visually the sound made by moving monsters out of view: @Rfootsteps@N, @Olight footsteps@N, @Gcreep noise@N, @Cflapping of wings@N, and @Mheavy footsteps@N. Note that monsters cannot hear your footsteps or, at the very least, they don’t care about such small noises!
`)
	md.helpTopic("Basics (Help)", s)
}

func (md *model) helpItems() {
	s := strings.TrimSpace(`
@CBananas@N are found on most levels and collected by moving onto them.

@CBarrels@N are found in every level. You can hide within them as long as nobody can see you while doing so. You can eat a banana and rest inside using the @MInteract Key@N(e) to recover both HP and MP.

@CPotions@N are found on most levels and drunk by moving onto them.

@CAmulets@N are magical items found on some levels. They can be equipped using the @MInteract Key@N(e) and offer special protection when being critically hit (left with 1 HP).

@CCloaks@N are magical items found on some levels. They can be equipped using the @MInteract Key@N(e) and offer some kind of passive benefit.

@CMagaras@N are found once on most levels. They are magical rods that you can equip by moving onto them and then @MEvoke@N(v) to use them. They have a finite number of charges and also require MP to use. When all magara slots are filled, you’ll have to equip any new one with the @MInteract Key@N(e) and select another magara to leave on the ground. That action will spend one turn.

@CMagical Stones@N are static items found in every level. Pressing the @MInteract Key@N(e) will activate them.

@CScrolls@N are story and flavour items that can be read using the @MInteract Key@N(e).

@CStairs@N are found once on every level and lead to the next one when used with the @MInteract Key@N(e). They may sometimes be blocked and require the activation of a seal stone to unblock them.
`)
	md.helpTopic("Items (Help)", s)
}

func (md *model) helpStatuses() {
	var sb strings.Builder
	sb.WriteString(`@YCOLORS@N

In the status bar and descriptions, positive status effects are shown as @BBlue@N, and negative ones are @RRed@N, while statuses that are neither are shown as @YYellow@N.

On the map, monster color may change depending on active status effects.  By default, status-free wandering monsters are @OOrange@N and hunting monsters are @RRed@N. Peaceful monsters are @BBlue@N and sleeping ones are @VViolet@N. @YYellow@N indicates lignification, @GGreen@N confusion, @CCyan@N daze.`)
	sb.WriteString("\n\n")
	sb.WriteString(`@YPLAYER STATUSES@N`)
	sb.WriteString("\n\n")
	for _, st := range playerStatuses {
		color := 'Y'
		switch {
		case st.Good():
			color = 'B'
		case st.Bad():
			color = 'R'
		}
		fmt.Fprintf(&sb, "@%c%s@N (@%c%s@N). ", color, st.String(), color, st.Short())
		sb.WriteString(st.Desc())
		sb.WriteString("\n\n")
	}
	sb.WriteString(`@YMONSTER STATUSES@N`)
	sb.WriteString("\n\n")
	sb.WriteString(`Monster statuses are less varied than player ones in Harmonist. Most happen as the result of using some magaras.

@YLignified@N monsters cannot move, but may attack you if adjacent, while @Cdazed@N monsters cannot act at all, and @Gconfused@N monsters can move as usual but cannot attack you. Some monsters may become exhausted or @Bsatiated@N after performing some special techniques and need to wait a few turns before performing them again.`)
	sb.WriteString("\n\n")
	md.helpTopic("Statuses (Help)", sb.String())
}

var playerStatuses = []status{
	StatusConfusion,
	StatusDelay,
	StatusDig,
	StatusDisguised,
	StatusDispersal,
	StatusExhausted,
	StatusFlames,
	StatusHidden,
	StatusIlluminated,
	StatusLevitation,
	StatusLight,
	StatusLignification,
	StatusShadows,
	StatusSwift,
	StatusTransparent,
	StatusUnhidden,
}

func (md *model) helpTips() {
	s := strings.TrimSpace(`
@YGENERAL TIPS@N

Harmonist is a @Mturn-based@N game: take your time when in perilous situations!

Always check your magaras to see if there’s one that can help in your situation. The sooner is often the better!

@YSTEALTH TIPS@N

Try to take advantage of the various different kinds of terrains, like foliage, tables, trees, holes in the walls, or barrels.

Magical stones can be useful for facilitating escape or incapacitating monsters in various ways, but they are also interesting for their map-revealing effects, so you might sometimes want to use some of them even if there are no monsters around. It may help exploration.

@CJumping off a wall@N is a useful escape tactic that can make you avoid harm even if cornered in a dead-end. However, it’s of course usually better to avoid dead-ends if you can! Otherwise, without a wall for propulsion, you can still perform a shorter jump over a monster, though that will give the monster time to hit you once.

Monsters don’t look behind diagonally, so performing lateral evasion tactics in the open is a simple way to escape an adjacent monster’s field of view by getting into one of its blind corners.

Watch out for @Rfootstep noise@N: while going through dense foliage or waiting in some hidden place, it can be very helpful to avoid unwanted encounters.
`)
	md.helpTopic("Tips (Help)", s)
}

// checkShaedra checks whether we already rescued Shaedra or not, to prevent
// going to optional levels before.
func (md *model) checkShaedra(st Stair) (err error) {
	g := md.g
	if g.Map.Depth == WinDepth && st == NormalStair && g.Map.At(g.Places.Shaedra) == StoryCell {
		err = errors.New("You have to rescue Shaedra first!")
	}
	return err

}

// Quit enters confirmation mode for quit without saving.
func (md *model) Quit() {
	md.g.LogStyled("Do you really want to quit without saving? [y/N]", logConfirm)
	md.mode = modeQuitConfirmation
}
