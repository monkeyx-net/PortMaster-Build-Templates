// This files defines the model structure, as well as initialization functions.

package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
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
	modeLoadGame           mode = iota // game load screen (load game)
	modeNewGame                        // primary spirit selection (new game)
	modeNewGameAdvanced                // primary spirit selection (advanced settings)
	modeNewGameMods                    // mod selection (advanced settings)
	modeNormal                         // map game mode
	modeCritical                       // hp critical warning pause
	modePager                          // pager (logs, help, lore)
	modeMenu                           // menu (game menu, inventory, settings)
	modeWizardConfirmation             // waiting for wizard mode confirmation
	modeEnd                            // game end: win or death
	modeQuitConfirmation               // waiting for no-save quit confirmation
	modeQuitting                       // wait until end message
)

// model describes the gruid.Model of the game.
type model struct {
	action         Action     // action to handle
	anims          Animations // animations
	auto           *auto      // auto-travel mode info
	desc           *ui.Label  // description label (for monsters, terrain)
	equipLabel     *ui.Label  // description label for equipping item
	gameEnded      bool       // whether the game ended
	drawGroundDesc bool       // whether to draw extra equipping desc
	g              *Game      // game state
	gd             gruid.Grid // drawing grid
	keysNormal     map[gruid.Key]Action
	keysTarget     map[gruid.Key]Action
	log            *ui.Label // game's last log messages
	menu           *menu     // menus (Menu, Settings, Inventory, Keymaps)
	menuActions    []Action  // invokable actions in last game/help/config menu
	mode           mode      // main mode
	pager          *pager    // pager (logs and the like)
	status         *gameStatus
	targ           *targeting
}

func (md *model) init() gruid.Effect {
	md.mode = modeLoadGame
	md.initStructures()
	md.initWidgets()
	md.initKeys()
	md.applyKeyConfig()

	g := md.g
	load, err := g.Load()
	md.g.md = md // handy cycle
	g.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	if !load {
		// Start a new game: go to spirit selection menu.
		// Initialize mods before starting the game, because we need
		// that information during the advanced new game settings.
		if GameConfig.AdvancedNewGame {
			g.Mods = slices.Clone(GameConfig.Mods)
			md.openSpiritSelectionMenu(modeNewGameAdvanced)
		} else {
			g.Mods = make([]bool, NMods)
			md.openSpiritSelectionMenu(modeNewGame)
		}
	}
	if err != nil {
		g.LogStyled("Warning: could not load old saved game… starting new game.", logError)
		log.Printf("Error: %v", err)
	}
	md.targ.CancelExamine()
	md.InitAnimations()
	if runtime.GOOS == "js" {
		if load {
			md.updateStatus()
			md.mode = modeNormal
		}
		return nil
	}
	return gruid.Sub(subSig)
}

func (md *model) initStructures() {
	md.auto = &auto{}
	md.auto.PRauto = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
}

func (md *model) initWidgets() {
	md.log = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.desc = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.desc.AdjustWidth = false
	md.equipLabel = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.equipLabel.AdjustWidth = false
	md.status = &gameStatus{}
	md.status.desc = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.pager = &pager{}
	md.pager.pg = ui.NewPager(ui.PagerConfig{
		Grid: gruid.NewGrid(UIWidth, UIHeight-1),
		Box:  &ui.Box{},
		Keys: ui.PagerKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.pager.markup = ui.StyledText{}.WithMarkups(Markups)
	style := ui.MenuStyle{
		Active: gruid.Style{Fg: ColorYellow},
	}
	md.menu = &menu{}
	md.menu.main = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth/2, UIHeight-1),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.menu.keys = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth, UIHeight-1),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.status.menu = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth, 1),
		Style: ui.MenuStyle{Layout: gruid.Point{0, 1}, Active: style.Active},
	})
}

func (md *model) initKeys() {
	md.keysNormal = map[gruid.Key]Action{
		gruid.KeyEscape:     ActionNone{},
		gruid.KeyArrowLeft:  ActionBump{Delta: gruid.Point{-1, 0}},
		gruid.KeyArrowDown:  ActionBump{Delta: gruid.Point{0, 1}},
		gruid.KeyArrowUp:    ActionBump{Delta: gruid.Point{0, -1}},
		gruid.KeyArrowRight: ActionBump{Delta: gruid.Point{1, 0}},
		"h":                 ActionBump{Delta: gruid.Point{-1, 0}},
		"j":                 ActionBump{Delta: gruid.Point{0, 1}},
		"k":                 ActionBump{Delta: gruid.Point{0, -1}},
		"l":                 ActionBump{Delta: gruid.Point{1, 0}},
		"H":                 ActionRun{Delta: gruid.Point{-1, 0}},
		"J":                 ActionRun{Delta: gruid.Point{0, 1}},
		"K":                 ActionRun{Delta: gruid.Point{0, -1}},
		"L":                 ActionRun{Delta: gruid.Point{1, 0}},
		".":                 ActionWait{},
		gruid.KeyEnter:      ActionWait{},
		"o":                 ActionAutoExplore{},
		"+":                 ActionNextMonster{},
		"-":                 ActionPreviousMonster{},
		"%":                 ActionNextItem{itemComestible},
		"!":                 ActionNextItem{itemTotem},
		"&":                 ActionNextItem{itemMenhir},
		">":                 ActionNextItem{itemPortal},
		"=":                 ActionNextItem{itemRune},
		"x":                 ActionExamineModeToggle{},
		"e":                 ActionInteract{},
		"i":                 ActionInventory{},
		gruid.KeySpace:      ActionMenu{},
		"?":                 ActionHelp{},
		"#":                 ActionDump{},
		"S":                 ActionSaveQuit{},
		"C":                 ActionConfig{},
		":":                 ActionSetKeys{},
		"m":                 ActionViewMessages{},
		"Q":                 ActionQuit{},
		"W":                 ActionWizard{},
		gruid.KeyPageDown:   ActionScroll{Delta: gruid.Point{0, -1}},
		gruid.KeyPageUp:     ActionScroll{Delta: gruid.Point{0, 1}},
	}
	md.keysTarget = map[gruid.Key]Action{
		gruid.KeyArrowLeft:  ActionCursorBump{Delta: gruid.Point{-1, 0}},
		gruid.KeyArrowDown:  ActionCursorBump{Delta: gruid.Point{0, 1}},
		gruid.KeyArrowUp:    ActionCursorBump{Delta: gruid.Point{0, -1}},
		gruid.KeyArrowRight: ActionCursorBump{Delta: gruid.Point{1, 0}},
		"h":                 ActionCursorBump{Delta: gruid.Point{-1, 0}},
		"j":                 ActionCursorBump{Delta: gruid.Point{0, 1}},
		"k":                 ActionCursorBump{Delta: gruid.Point{0, -1}},
		"l":                 ActionCursorBump{Delta: gruid.Point{1, 0}},
		"H":                 ActionCursorRun{Delta: gruid.Point{-1, 0}},
		"J":                 ActionCursorRun{Delta: gruid.Point{0, 1}},
		"K":                 ActionCursorRun{Delta: gruid.Point{0, -1}},
		"L":                 ActionCursorRun{Delta: gruid.Point{1, 0}},
		gruid.KeyEnter:      ActionTravel{},
		".":                 ActionTravel{},
		gruid.KeyEscape:     ActionExamineModeToggle{},
	}
	CustomKeys = false
}

func (md *model) applyKeyConfig() {
	if GameConfig.NormalModeKeys != nil {
		md.keysNormal = GameConfig.NormalModeKeys
		// For ensuring menu access and esc functionality.
		md.keysNormal[gruid.KeySpace] = ActionMenu{}
		md.keysNormal[gruid.KeyEscape] = ActionNone{}
	}
	if GameConfig.ExamineModeKeys != nil {
		md.keysTarget = GameConfig.ExamineModeKeys
		// For ensuring back to normal mode.
		md.keysTarget[gruid.KeyEscape] = ActionExamineModeToggle{}
	}
}

// InitConfig loads saved config, if any, and initializes GameConfig.
func InitConfig() error {
	GameConfig.DarkColors = true // default to dark theme
	GameConfig.Tiles = true      // default to tiles (when available)
	GameConfig.Version = Version
	load, err := LoadConfig()
	if err != nil {
		err = fmt.Errorf("error loading config: %v", err)
		saverr := SaveConfig()
		if saverr != nil {
			log.Printf("error resetting badly loaded config: %v", err)
		}
		return err
	}
	if load {
		CustomKeys = true
	}
	return err
}

var spiritKeys = []rune{'H', 'B', 'F', 'f', 'c'}
var spiritKeysAdvanced = []rune{'C', 'b'}

func (md *model) openSpiritSelectionMenu(m mode) {
	hstyle := gruid.Style{Fg: ColorCyan}
	entries := []ui.MenuEntry{}
	spirits := getPrimarySpirits(m)
	title := "New Game"
	if m == modeNewGameAdvanced {
		title = "New Game (advanced)"
	}
	keys := getPrimarySpiritKeys(m)
	for i, si := range spirits {
		r := keys[i]
		switch r {
		case 'H':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("Primary Spirit Selection").WithStyle(hstyle),
				Disabled: true,
			})
		}
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, si.Name),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.updateItemDesc(spiritEntity(md.g.Mods, spirits[0]))
	md.mode = m
	md.menu.mode = modeSelection // not really needed
}

// getPrimarySpirits returns the spirit info for primary spirits in the new
// game menu.
func getPrimarySpirits(m mode) []spiritInfo {
	if m == modeNewGameAdvanced {
		return append(slices.Clone(primarySpirits), primarySpiritsAdvanced...)
	}
	return primarySpirits
}

// getPrimarySpiritKeys returns the spirit info for primary spirits in the new
// game menu.
func getPrimarySpiritKeys(m mode) []rune {
	if m == modeNewGameAdvanced {
		return append(slices.Clone(spiritKeys), spiritKeysAdvanced...)
	}
	return spiritKeys
}
