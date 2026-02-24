package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"strings"
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

const (
	UIWidth  = 80 // UI width
	UIHeight = 24 // UI height
)

var (
	DisableAnimations bool = false       // whether to disable animations
	LogGame                = false       // write game logs to file
	ColorMode              = ColorMode16 // default 16-color palette
)

// colorMode represents various color compatibility modes.
type colorMode int

const (
	ColorMode16    colorMode = iota
	ColorMode8               // use 8-color compatibility mode (default for windows)
	ColorMode256             // use solarized 256-color approximation
	ColorMode24bit           // use true color selenized palette
)

// CustomKeys tracks whether we're using custom key bindings.
var CustomKeys bool

// GameConfig contains the current game config.
var GameConfig Config

// mode represents the main model mode
type mode int

const (
	modeNormal mode = iota
	modePager
	modeSmallPager
	modeMenu
	modeQuit
	modeQuitConfirmation
	modeJumpConfirmation
	modeWizardConfirmation
	modeDump // simplified dump visualization after end
	modeEnd  // win or death
	modeHPCritical
	modeWelcome
	modeStory
)

type pagerMode int

const (
	modeLogs pagerMode = iota
	modeHelpKeys
)

type menuMode int

const (
	modeInventory menuMode = iota
	modeSettings
	modeKeys
	modeKeysChange
	modeGameMenu
	modeHelpMenu
	modeEvocation
	modeEquip
	modeWizard
)

type model struct {
	g           *Game // game state
	gd          gruid.Grid
	mode        mode
	menuMode    menuMode
	pagerMode   pagerMode
	menu        *ui.Menu
	keysMenu    *ui.Menu
	status      *ui.Menu
	log         *ui.Label
	description *ui.Label
	equipLabel  *ui.Label
	statusDesc  *ui.Label
	pager       *ui.Pager
	smallPager  *ui.Pager
	pagerMarkup ui.StyledText
	targ        mapTargInfo
	logs        []ui.StyledText
	keysNormal  map[gruid.Key]Action
	keysTarget  map[gruid.Key]Action
	finished    bool
	statusFocus bool
	anims       Animations
	story       int // part of story sequence (0 no story, 1 first part, and so on)
	critical    bool
	auto        bool
	confirm     bool
}

type mapTargInfo struct {
	kbTargeting bool
	ex          *examination
}

func (md *model) initKeys() {
	md.keysNormal = map[gruid.Key]Action{
		gruid.KeyArrowLeft:  ActionW,
		gruid.KeyArrowDown:  ActionS,
		gruid.KeyArrowUp:    ActionN,
		gruid.KeyArrowRight: ActionE,
		"h":                 ActionW,
		"j":                 ActionS,
		"k":                 ActionN,
		"l":                 ActionE,
		"H":                 ActionRunW,
		"J":                 ActionRunS,
		"K":                 ActionRunN,
		"L":                 ActionRunE,
		".":                 ActionWaitTurn,
		gruid.KeyEnter:      ActionWaitTurn,
		"o":                 ActionExplore,
		"x":                 ActionExamineToggle,
		"u":                 ActionExamineScrollUp,
		gruid.KeyPageUp:     ActionExamineScrollUp,
		"d":                 ActionExamineScrollDown,
		gruid.KeyPageDown:   ActionExamineScrollDown,
		"v":                 ActionEvoke,
		"V":                 ActionEvoke,
		"z":                 ActionEvoke,
		"e":                 ActionInteract,
		"E":                 ActionInteract,
		"i":                 ActionInventory,
		"I":                 ActionInventory,
		"m":                 ActionLogs,
		"M":                 ActionMenu,
		gruid.KeySpace:      ActionMenu,
		"#":                 ActionDump,
		"?":                 ActionHelpMenu,
		"-":                 ActionPreviousMonster,
		"+":                 ActionNextMonster,
		">":                 ActionNextStairs,
		"n":                 ActionNextObject,
		"S":                 ActionSave,
		"Q":                 ActionQuit,
		"W":                 ActionWizard,
		"»":                 ActionWizardDescend,
		"C":                 ActionSettings,
		gruid.KeyTab:        ActionSettings,
		":":                 ActionSetKeys,
	}
	md.keysTarget = map[gruid.Key]Action{
		".":             ActionTravel,
		gruid.KeyEnter:  ActionTravel,
		gruid.KeyEscape: ActionExamineToggle,
	}
	CustomKeys = false
}

func (md *model) initWidgets() {
	stt := ui.StyledText{}.WithMarkups(Markups)
	md.log = ui.NewLabel(stt)
	md.description = ui.NewLabel(stt)
	md.description.AdjustWidth = false
	md.equipLabel = ui.NewLabel(stt)
	md.equipLabel.AdjustWidth = false
	md.statusDesc = ui.NewLabel(stt)
	md.pager = ui.NewPager(ui.PagerConfig{
		Grid: gruid.NewGrid(UIWidth, UIHeight-1),
		Box:  &ui.Box{},
		Keys: ui.PagerKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.smallPager = ui.NewPager(ui.PagerConfig{
		Grid: gruid.NewGrid(60, UIHeight-1),
		Box:  &ui.Box{},
		Keys: ui.PagerKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.pagerMarkup = stt
	style := ui.MenuStyle{
		Active: gruid.Style{}.WithFg(ColorYellow),
	}
	md.menu = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth/2, UIHeight-1),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.keysMenu = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth, UIHeight-1),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.status = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth, 1),
		Style: ui.MenuStyle{Layout: gruid.Point{0, 1}, Active: style.Active},
	})
}

func InitConfig() error {
	GameConfig.DarkColors = true
	GameConfig.Version = Version
	GameConfig.Tiles = true
	load, err := LoadConfig()
	if err != nil {
		err = fmt.Errorf("Error loading config: %v", err)
		saverr := SaveConfig()
		if saverr != nil {
			log.Printf("Error resetting badly loaded config: %v", err)
		}
		return err
	}
	if load {
		CustomKeys = true
	}
	return err
}

func (md *model) init() gruid.Effect {
	if runtime.GOOS != "js" {
		md.mode = modeWelcome
	}
	md.initKeys()
	md.initWidgets()

	g := md.g

	md.applyConfig()
	load, err := g.Load()
	md.g.md = md // handy cycle
	if !load {
		g.Init()
		g.InitLevel()
		g.checks()
		md.g.ComputeNoise()
		md.g.ComputeFOV()
		md.g.ComputeMonsterFOV()
	} else {
		g.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	if err != nil {
		g.LogStyled("Warning: could not load old saved game… starting new game.", logError)
		log.Printf("Error: %v", err)
	}
	md.updateStatusInfo()
	md.targ.ex = &examination{}
	md.CancelExamine()
	md.InitAnimations()
	if runtime.GOOS == "js" {
		return nil
	}
	return gruid.Sub(subSig)
}

func (md *model) more(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "x", "X", gruid.KeyEscape, gruid.KeySpace:
			return true
		}
	}
	return false
}

func (md *model) interrupt(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return true
	case gruid.MsgMouse:
		return msg.Action != gruid.MouseMove
	}
	return false
}

func (md *model) Update(msg gruid.Msg) gruid.Effect {
	if _, ok := msg.(gruid.MsgInit); ok {
		return md.init()
	}
	if _, ok := msg.(gruid.MsgQuit); ok {
		if md.mode != modeQuit { // in case of already quitting
			if err := md.g.Save(); err != nil {
				log.Printf("saving before quitting: %v", err)
			}
		}
		md.mode = modeQuit
		return gruid.End()
	}
	//log.Printf("msg %+v", msg)
	md.anims.draw = false
	if msg, ok := msg.(msgAnim); ok {
		if int(msg) != md.anims.idx {
			// The animation message is from a finished animation
			// sequence, so we just skip it.
			return nil
		}
		if !md.anims.Done() {
			md.anims.draw = true
			return md.AnimNext()
		}
		md.anims.Finish()
		return nil
	}
	anims := !md.anims.Done()
	if md.interrupt(msg) {
		if anims {
			md.anims.Finish()
			anims = false
		}
		md.g.StopAuto()
	}
	eff := md.update(msg)
	if cmd := md.AnimCmd(); !anims && cmd != nil {
		return gruid.Batch(eff, cmd)
	}
	return eff
}

func (md *model) update(msg gruid.Msg) gruid.Effect {
	var eff gruid.Effect
	switch md.mode {
	case modeWelcome:
		switch msg := msg.(type) {
		case gruid.MsgKeyDown:
			md.mode = modeNormal
		case gruid.MsgMouse:
			if msg.Action != gruid.MouseMove {
				md.mode = modeNormal
			}
		}
		return nil
	case modeQuit:
		return nil
	case modeEnd:
		if md.more(msg) {
			md.finished = true
			md.mode = modeDump
			md.dump(md.g.WriteDump())
		}
		return nil
	case modeDump:
		return md.updateDump(msg)
	case modeQuitConfirmation:
		eff := md.updateQuitConfirmation(msg)
		if md.mode == modeQuit {
			err := RemoveSaveFile()
			if err != nil {
				log.Printf("Error removing save file: %v", err)
			}
			RemoveReplay()
			return eff
		}
		return eff
	case modeJumpConfirmation:
		md.updateJumpConfirmation(msg)
		return nil
	case modeWizardConfirmation:
		md.updateWizardConfirmation(msg)
		return nil
	case modeHPCritical:
		if md.more(msg) {
			md.mode = modeNormal
			md.g.Log("Ok. Be careful, then.")
		}
		return nil
	case modeNormal:
		eff = md.updateNormal(msg)
	case modeStory:
		if md.more(msg) {
			md.NextStoryPart()
		}
	case modePager:
		eff = md.updatePager(msg)
	case modeSmallPager:
		eff = md.updateSmallPager(msg)
	case modeMenu:
		switch md.menuMode {
		case modeKeys:
			eff = md.updateKeysMenu(msg)
		case modeKeysChange:
			eff = md.updateKeysChange(msg)
		default:
			eff = md.updateMenu(msg)
		}
	}
	md.updateStatusInfo()
	return eff
}

func (md *model) updateQuitConfirmation(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		if msg.Key == "y" || msg.Key == "Y" {
			md.mode = modeQuit
			return gruid.End()
		}
		md.mode = modeNormal
	}
	return nil
}

func (md *model) updateJumpConfirmation(msg gruid.Msg) {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		md.mode = modeNormal
		if msg.Key == "y" || msg.Key == "Y" {
			md.g.FallAbyss(DescendJump)
		} else {
			md.g.Log("No jump, then.")
		}
	}
}

func (md *model) updateWizardConfirmation(msg gruid.Msg) {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		md.mode = modeNormal
		if msg.Key == "y" || msg.Key == "Y" {
			md.g.EnterWizardMode()
		} else {
			md.g.Log("Continuing normally, then.")
		}
	}
}

type msgAuto int

func (md *model) updateNormal(msg gruid.Msg) gruid.Effect {
	var eff gruid.Effect
	switch msg := msg.(type) {
	case msgAuto:
		if int(msg) == md.g.Turn && md.auto {
			return md.EndTurn()
		}
	case gruid.MsgKeyDown:
		eff = md.updateKeyDown(msg)
	case gruid.MsgMouse:
		eff = md.updateMouse(msg)
	}
	return eff
}

func (md *model) updateKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	if md.g.resting {
		// Ignore keyboard actions while resting.
		return nil
	}
	md.statusFocus = false
	if !md.targ.kbTargeting && InMap(md.targ.ex.p) {
		md.CancelExamine()
	}
	again, eff, err := md.normalModeKeyDown(msg.Key, msg.Mod&gruid.ModShift != 0)
	if err != nil {
		md.g.Log(err.Error())
	}
	if again {
		return eff
	}
	return md.EndTurn()
}

func (md *model) updateMouse(msg gruid.MsgMouse) gruid.Effect {
	if md.g.resting {
		// Ignore mouse actions while resting.
		return nil
	}
	if msg.P.Y == UIHeight-1 {
		return md.updateStatusMouse(msg)
	}
	md.status.SetActive(0)
	md.statusFocus = false
	p := msg.P.Add(gruid.Point{0, -2}) // relative position ignoring log
	switch msg.Action {
	case gruid.MouseWheelUp:
		if md.targ.ex.p != invalidPos {
			md.targ.ex.scroll = 0
		}
	case gruid.MouseWheelDown:
		if md.targ.ex.p != invalidPos {
			md.targ.ex.scroll++
		}
	case gruid.MouseMove:
		if md.auto || md.targ.kbTargeting {
			break
		}
		if InMap(p) {
			md.Examine(p)
		} else {
			md.CancelExamine()
		}
	case gruid.MouseSecondary:
		md.openMenu()
	case gruid.MouseMain:
		if !InMap(p) {
			return nil
		}
		var again bool
		var eff gruid.Effect
		var err error
		if md.targ.kbTargeting && p != md.targ.ex.p {
			md.Examine(p)
			return nil
		}
		if Distance(p, md.g.Player.P) == 1 {
			again, err = md.g.PlayerBump(p)
		} else {
			again, eff, err = md.DoAction(ActionTravel)
		}
		if err != nil {
			again = true
			md.g.Log(err.Error())
		}
		if again {
			return eff
		}
		return md.EndTurn()
	}
	return nil
}

type statusItem int

const (
	statusDepth statusItem = iota
	statusTurns
	statusHP
	statusMP
	statusBananas
	statusMenu
	statusInventory
	statusEvoke
	statusInteract
)

func (md *model) updateStatusMouse(msg gruid.MsgMouse) gruid.Effect {
	msg.P.Y = 0
	if !msg.P.In(md.status.Bounds()) || md.targ.kbTargeting {
		md.statusFocus = false
		return nil
	}
	md.status.Update(msg)
	update := !md.statusFocus
	switch md.status.Action() {
	case ui.MenuMove:
		update = true
	case ui.MenuInvoke:
		i := statusItem(md.status.Active())
		var action Action
		switch i {
		case statusMenu:
			action = ActionMenu
		case statusInventory:
			action = ActionInventory
		case statusEvoke:
			action = ActionEvoke
		case statusInteract:
			action = ActionInteract
			md.statusFocus = false
		}
		again, eff, err := md.DoAction(action)
		if err != nil {
			md.g.Log(err.Error())
		}
		if again {
			return eff
		}
		return md.EndTurn()
	}
	if update {
		md.CancelExamine()
		const statusIndex = statusInteract + 2
		i := statusItem(md.status.Active())
		md.statusFocus = false
		switch {
		case i == statusDepth:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Depth")}
			md.statusDesc.SetText("Dungeon depth.")
			md.statusFocus = true
		case i == statusTurns:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Turns")}
			md.statusDesc.SetText("Number of turns since the beginning.")
			md.statusFocus = true
		case i == statusHP:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Health")}
			md.statusDesc.SetText("Your hit points.")
			md.statusFocus = true
		case i == statusMP:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Magic Points")}
			md.statusDesc.SetText("Your magic points. Needed for evoking magaras.")
			md.statusFocus = true
		case i == statusBananas:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Bananas")}
			md.statusDesc.SetText("Need to eat one before sleeping in barrels.")
			md.statusFocus = true
		case i == statusMenu:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Menu (SPACE or M)")}
			md.statusDesc.SetText("Click to open menu.")
			md.statusFocus = true
		case i == statusInventory:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Inventory (i)")}
			md.statusDesc.SetText("Click to open inventory.")
			md.statusFocus = true
		case i == statusEvoke:
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Evoke magara (v)")}
			md.statusDesc.SetText("Click to open magara evocation menu.")
			md.statusFocus = true
		case i == statusInteract:
			s, ok := md.CanInteract()
			if !ok {
				break
			}
			md.statusDesc.Box = &ui.Box{Title: ui.Text("Interact (e)")}
			md.statusDesc.SetText(fmt.Sprintf("Click to %v.", s))
			md.statusFocus = true
		case i >= statusIndex:
			i := md.status.Active() - int(statusIndex)
			sts := md.sortedStatuses()
			if i > len(sts)-1 {
				if md.g.Player.HP > 0 && md.mode != modeEnd {
					md.statusDesc.Box = &ui.Box{Title: ui.Text("Player’s Tile")}
					md.statusDesc.SetText("The tile where the player is.")
					md.statusFocus = true
				}
				break
			}
			title := ui.StyledText{}.WithMarkups(Markups)
			st := sts[i]
			if !st.Flag() {
				title = title.WithTextf("%s (for %d turns)", sts[i].String(), md.g.Player.Statuses[sts[i]]/DurationStatusStep)
			} else {
				title = title.WithText(sts[i].String())
			}
			md.statusDesc.Box = &ui.Box{Title: title}
			md.statusDesc.SetText(sts[i].Desc())
			md.statusDesc.Content = md.statusDesc.Content.Format(UIWidth - 2)
			md.statusFocus = true
		}
	}
	return nil
}

func (md *model) Auto() gruid.Effect {
	md.auto = md.g.AutoPlayer()
	if md.auto {
		n := md.g.Turn
		return gruid.Cmd(func() gruid.Msg {
			t := time.NewTimer(AnimDurShort)
			<-t.C
			return msgAuto(n)
		})
	}
	return nil
}

// EndTurn finalizes player's turn and runs other events until next player
// turn.
func (md *model) EndTurn() gruid.Effect {
	md.mode = modeNormal
	md.g.EndTurn()
	eff := md.Auto()
	md.g.TurnStats()
	md.updateMapInfo()
	if md.g.Player.HP <= 0 {
		md.death()
		return nil
	}
	md.g.CheckForStorySequence()
	if md.critical {
		md.g.LogStyled("*** CRITICAL HP WARNING ***", logDamage)
		md.critical = false
		md.confirm = true
	}
	if md.confirm {
		md.g.LogStyled("[SPACE/ESC to continue]", logConfirm)
		md.confirm = false
	}
	md.g.LogNextTick = md.g.LogIndex
	return eff
}

func (md *model) updateMapInfo() {
	md.g.ComputeNoise()
	md.g.ComputeFOV()
	md.g.ComputeMonsterFOV()
	if md.g.highlight != nil {
		md.examine(md.targ.ex.p)
	}
}

func (md *model) updatePager(msg gruid.Msg) gruid.Effect {
	md.pager.Update(msg)
	if md.pager.Action() == ui.PagerQuit {
		md.mode = modeNormal
	}
	return nil
}

func (md *model) updateSmallPager(msg gruid.Msg) gruid.Effect {
	md.smallPager.Update(msg)
	if md.smallPager.Action() == ui.PagerQuit {
		md.mode = modeNormal
	}
	return nil
}

func (md *model) updateDump(msg gruid.Msg) gruid.Effect {
	md.pager.Update(msg)
	if md.pager.Action() == ui.PagerQuit {
		md.mode = modeQuit
		return gruid.End()
	}
	return nil
}

func (md *model) updateKeysChange(msg gruid.Msg) gruid.Effect {
	mkd, ok := msg.(gruid.MsgKeyDown)
	if !ok {
		return nil
	}
	key := mkd.Key
	switch key {
	case gruid.KeyEscape, gruid.KeySpace:
		md.openKeyBindings()
		return nil
	default:
		action := ConfigurableKeyActions[md.keysMenu.Active()]
		if action.ExamineModeAction() {
			md.keysTarget[key] = action
		} else {
			md.keysNormal[key] = action
		}
		GameConfig.NormalModeKeys = md.keysNormal
		GameConfig.TargetModeKeys = md.keysTarget
		err := SaveConfig()
		if err != nil {
			md.g.LogStyled("Error while saving config changes.", logError)
		}
		md.openKeyBindings()
		return nil
	}
}

func (md *model) updateKeysMenu(msg gruid.Msg) gruid.Effect {
	md.keysMenu.Update(msg)
	switch act := md.keysMenu.Action(); act {
	case ui.MenuQuit:
		md.mode = modeNormal
	case ui.MenuInvoke:
		md.menuMode = modeKeysChange
	case ui.MenuPass:
		msg, ok := msg.(gruid.MsgKeyDown)
		if !ok {
			return nil
		}
		if msg.Key == "R" {
			md.initKeys()
			GameConfig.NormalModeKeys = md.keysNormal
			GameConfig.TargetModeKeys = md.keysTarget
			err := SaveConfig()
			if err != nil {
				md.g.LogStyled("Error while resetting config changes.", logError)
			}
			md.openKeyBindings()
		}
	}
	return nil
}

func (md *model) updateMenu(msg gruid.Msg) gruid.Effect {
	md.menu.Update(msg)
	switch act := md.menu.Action(); act {
	case ui.MenuQuit:
		md.mode = modeNormal
	case ui.MenuMove, ui.MenuInvoke:
		switch md.menuMode {
		case modeInventory:
			items := []Item{md.g.Player.Inventory.Body, md.g.Player.Inventory.Neck, md.g.Player.Inventory.Misc}
			it := items[md.menu.Active()]
			stt := ui.StyledText{}.WithMarkups(Markups)
			md.description.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
			md.description.Box = &ui.Box{Title: stt.WithText(it.String())}
		case modeEvocation:
			items := md.g.Player.Magaras
			it := items[md.menu.Active()]
			stt := ui.StyledText{}.WithMarkups(Markups)
			md.description.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
			md.description.Box = &ui.Box{Title: stt.WithText(it.String())}
			if act != ui.MenuInvoke {
				break
			}
			err := md.g.UseMagara(md.menu.Active())
			if err != nil {
				md.g.Log(err.Error())
				md.mode = modeNormal
				break
			}
			return md.EndTurn()
		case modeEquip:
			items := md.g.Player.Magaras
			it := items[md.menu.Active()]
			stt := ui.StyledText{}.WithMarkups(Markups)
			md.description.Content = stt.WithText(it.Desc(md.g)).Format(UIWidth/2 - 1 - 2)
			md.description.Box = &ui.Box{Title: stt.WithTextf("%s (equipped)", it.String())}
			if act != ui.MenuInvoke {
				break
			}
			err := md.g.EquipMagara(md.menu.Active())
			if err != nil {
				md.g.Log(err.Error())
				md.mode = modeNormal
				break
			}
			return md.EndTurn()
		case modeGameMenu, modeHelpMenu, modeSettings:
			var actions []Action
			switch md.menuMode {
			case modeGameMenu:
				actions = menuActions
			case modeHelpMenu:
				actions = helpActions
			case modeSettings:
				actions = settingsActions
			}
			idx := md.menu.ActiveInvokable()
			if idx < 0 || idx >= len(actions) {
				md.description.Content = ui.StyledText{}
				break
			}
			action := actions[idx]
			md.updateActionDesc(action)
			if act != ui.MenuInvoke {
				break
			}
			_, eff, err := md.DoAction(action)
			if err != nil {
				// should not happen
				log.Printf("menu: %v", err)
			}
			return eff
		}
	}
	return nil
}

func (md *model) normalModeKeyDown(key gruid.Key, shift bool) (again bool, eff gruid.Effect, err error) {
	action := md.keysNormal[key]
	if md.targ.kbTargeting {
		if a, ok := md.keysTarget[key]; ok {
			action = a
		}
	}
	if shift && !key.IsRune() {
		switch action {
		case ActionW:
			action = ActionRunW
		case ActionS:
			action = ActionRunS
		case ActionN:
			action = ActionRunN
		case ActionE:
			action = ActionRunE
		}
	}
	again, eff, err = md.DoAction(action)
	if _, ok := err.(actionErrorUnknown); ok {
		again, err = true, nil
	}
	return again, eff, err
}

func (md *model) death() {
	g := md.g
	g.LevelStats()
	if len(g.Stats.Achievements) == 0 {
		NoAchievement.Get(g)
	}
	g.LogStyled("You die...", logSpecial)
	g.LogStyled("[SPACE/ESC to continue]", logConfirm)
	md.mode = modeEnd
}

func (md *model) win() {
	g := md.g
	if g.Wizard != WizardNone {
		g.LogStyled("You escape through the magic portal! **WIZARD**", logSpecial)
		g.LogStyled("[SPACE/ESC to continue]", logConfirm)
	} else {
		g.LogStyled("You escape through the magic portal!", logSpecial)
		g.LogStyled("[SPACE/ESC to continue]", logConfirm)
	}
	md.mode = modeEnd
	g.md.PortalEscapeAnimation()
}

func (md *model) dump(msg string, err error) {
	s := md.g.SimplifedDump(msg, err)
	lines := strings.Split(s, "\n")
	stts := []ui.StyledText{}
	for _, l := range lines {
		stts = append(stts, ui.Text(l))
	}
	md.pager.SetLines(stts)
	md.pager.SetBox(&ui.Box{Title: ui.Text("Game Summary").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.pager.SetCursor(gruid.Point{0, 0})
	//log.Printf("%v", s)
	//log.Printf("%v", stts)
}

func (md *model) applyConfig() {
	if GameConfig.NormalModeKeys != nil {
		md.keysNormal = GameConfig.NormalModeKeys
		// For ensuring menu access functionality.
		md.keysNormal[gruid.KeySpace] = ActionMenu
	}
	if GameConfig.TargetModeKeys != nil {
		md.keysTarget = GameConfig.TargetModeKeys
		// For ensuring back to normal mode.
		md.keysTarget[gruid.KeyEscape] = ActionExamineToggle
	}
}
