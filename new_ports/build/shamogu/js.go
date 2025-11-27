//go:build js

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"syscall/js"

	"codeberg.org/anaseto/gruid"
	jsd "codeberg.org/anaseto/gruid-js"
	"codeberg.org/anaseto/gruid/ui"
)

var driver gruid.Driver

func initDriver() {
	dr := jsd.NewDriver(jsd.Config{
		TileManager: &monochromeTileManager{},
		AppCanvasId: "gamecanvas",
		AppDivId:    "gamediv",
	})
	driver = dr
}

func clearCache() {
	dr := driver.(*jsd.Driver)
	dr.ClearCache()
}

func main() {
	initDriver()
	mainMenu := newMainMenu()
	err := InitConfig()
	if err != nil {
		mainMenu.err = err
	}
	log.SetPrefix("shamogu ")
	for {
		app := gruid.NewApp(gruid.AppConfig{
			Driver: driver,
			Model:  mainMenu,
		})
		if err := app.Start(context.Background()); err != nil {
			log.Fatal(err)
		}
		switch mainMenu.action {
		case MainPlayGame:
			mainMenu.err = RunGame()
		case MainReplayGame:
			mainMenu.err = RunReplay()
		}
		mainMenu.action = MainMenuDefault
	}
}

// mainMenuAction represents the various kinds of actions in the main menu.
type mainMenuAction int

const (
	MainMenuDefault mainMenuAction = iota
	MainPlayGame
	MainReplayGame
)

// mainMenu represents the gruid model for the main menu.
type mainMenu struct {
	gd     gruid.Grid
	menu   *ui.Menu
	errs   *ui.Label
	err    error
	action mainMenuAction
}

func newMainMenu() *mainMenu {
	gd := gruid.NewGrid(UIWidth, UIHeight)
	md := &mainMenu{gd: gd}
	style := ui.MenuStyle{
		Active: gruid.Style{}.WithFg(ColorYellow),
	}
	md.menu = ui.NewMenu(ui.MenuConfig{
		Grid: gruid.NewGrid(UIWidth/2, 3),
		Entries: []ui.MenuEntry{
			{Text: ui.Text("- (P)lay"), Keys: []gruid.Key{"p", "P"}},
			{Text: ui.Text("- (R)eplay"), Keys: []gruid.Key{"r", "R"}},
		},
		Style: style,
	})
	md.errs = ui.NewLabel(ui.StyledText{}.WithStyle(gruid.Style{}.WithFg(ColorRed)))
	return md
}

func menuRange() gruid.Range {
	return gruid.NewRange(20, 18, UIWidth, UIHeight)
}

func (md *mainMenu) Update(msg gruid.Msg) gruid.Effect {
	if md.action != MainMenuDefault {
		return nil
	}
	md.menu.Update(menuRange().RelMsg(msg))
	switch md.menu.Action() {
	case ui.MenuMove:
	case ui.MenuInvoke:
		switch md.menu.ActiveInvokable() {
		case 0:
			md.action = MainPlayGame
			return gruid.End()
		case 1:
			md.action = MainReplayGame
			return gruid.End()
		}
	}
	return nil
}

func (md *mainMenu) Draw() gruid.Grid {
	md.gd.Fill(gruid.Cell{Rune: ' '})
	ui.Textf("Shamanic Mountain Guardian - Shamogu %s", Version).WithStyle(gruid.Style{}.WithFg(ColorMagenta)).
		Draw(md.gd.Slice(gruid.NewRange(-25+UIWidth/2, 5, UIWidth, UIHeight)))
	gd := md.gd.Slice(gruid.NewRange(-17+UIWidth/2, 7, UIWidth, UIHeight))
	drawGamePicture(gd)
	md.gd.Slice(menuRange()).Copy(md.menu.Draw())
	if md.err != nil {
		md.errs.SetText(md.err.Error())
		md.errs.Draw(md.gd.Slice(gruid.NewRange(10, 4, UIWidth, 6)))
	}
	return md.gd
}

const repit = "shamogureplay"
const replock = "shamogureplock"

// RunGame starts the game.
func RunGame() error {
	gd := gruid.NewGrid(UIWidth, UIHeight)
	md := &model{gd: gd, g: &Game{}, targ: &targeting{}}
	repw := &bytes.Buffer{}
	defer func() {
		if md.gameEnded {
			SetItem(repit, repw.Bytes())
			RemoveItem(replock)
			RemoveSaveFile()
		} else {
			rl, err := GetItem(replock)
			if err != nil {
				log.Printf("replock: %v", err)
			}
			if rl != nil {
				// save ongoing game replay
				SetItem(repit, repw.Bytes())
			}
		}
	}()
	rl, err := GetItem(replock)
	if rl == nil {
		if err != nil {
			log.Printf("replock: %v", err)
		}
		// new game: remove old replay
		RemoveReplay()
	}
	replay, err := GetItem(repit)
	if err != nil {
		log.Printf("get replay: %v", err)
	} else if replay != nil {
		repw.Write(replay)
	}
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      driver,
		Model:       md,
		FrameWriter: repw,
	})
	SetItem(replock, []byte("1"))
	return app.Start(context.Background())
}

// RunReplay runs a replay of a previous game saved in local storage.
func RunReplay() error {
	replay, err := GetItem(repit)
	if err != nil {
		return fmt.Errorf("replay loading: %v", err)
	}
	if replay == nil {
		return errors.New("no replay found")
	}
	repr := &bytes.Buffer{}
	repr.Write(replay)
	fd, err := gruid.NewFrameDecoder(repr)
	if err != nil {
		return fmt.Errorf("frame decoder: %v", err)
	}
	gd := gruid.NewGrid(UIWidth, UIHeight)
	rep := ui.NewReplay(ui.ReplayConfig{
		Grid:         gd,
		FrameDecoder: fd,
	})
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
		Model:  rep,
	})
	return app.Start(context.Background())
}

// DataDir returns is defined here only for build purposes, it is not used for
// the js backend.
func DataDir() (string, error) {
	return "", nil
}

// GetItem retrieves a base64 encoded item from localStorage. It returns nil if
// the item does not exist in the storage. It returns an error if localStorage
// is not available, or an item existed but could not be decoded.
func GetItem(item string) ([]byte, error) {
	storage := js.Global().Get("localStorage")
	if storage.Type() != js.TypeObject {
		return nil, errors.New("localStorage not found")
	}
	save := storage.Call("getItem", item)
	if save.Type() != js.TypeString {
		// this means the item does not exist
		return nil, nil
	}
	s, err := base64.StdEncoding.DecodeString(save.String())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SetItem sets an item to a given value in the localStorage. The value will be
// base64 encoded.
func SetItem(item string, value []byte) error {
	storage := js.Global().Get("localStorage")
	if storage.Type() != js.TypeObject {
		return errors.New("localStorage not found")
	}
	s := base64.StdEncoding.EncodeToString(value)
	storage.Call("setItem", item, s)
	return nil
}

// RemoveItem removes an item from localStorage.
func RemoveItem(item string) {
	storage := js.Global().Get("localStorage")
	if storage.Type() != js.TypeObject {
		log.Print("localStorage not found")
	}
	storage.Call("removeItem", item)
}

const shamogusave = "shamogusave"
const shamoguconfig = "shamoguconfig"

// Save saves the game to local storage.
func (g *Game) Save() error {
	save, err := g.GameSave()
	if err != nil {
		return err
	}
	err = SetItem(shamogusave, save)
	if err != nil {
		return err
	}
	return nil
}

// SaveConfig saves the game's config to local storage.
func SaveConfig() error {
	conf, err := GameConfig.ConfigSave()
	if err != nil {
		return err
	}
	err = SetItem(shamoguconfig, conf)
	if err != nil {
		return err
	}
	return nil
}

// RemoveSaveFile removes any save from local storage.
func RemoveSaveFile() error {
	RemoveItem(shamogusave)
	return nil
}

// RemoveReplay removes any replay from local storage.
func RemoveReplay() {
	RemoveItem(replock)
	RemoveItem(repit)
}

// Load loads an existing game from a saved one in local storage.
func (g *Game) Load() (bool, error) {
	s, err := GetItem(shamogusave)
	if err != nil || s == nil {
		return false, nil
	}
	lg, err := g.DecodeGameSave(s)
	if err != nil {
		return false, err
	}
	if lg.Version != Version {
		return false, nil
	}
	*g = *lg
	return true, nil
}

// LoadConfig loads game's config from local storage.
func LoadConfig() (bool, error) {
	s, err := GetItem(shamoguconfig)
	if err != nil || s == nil {
		return false, err
	}
	c, err := DecodeConfigSave(s)
	if err != nil {
		return false, err
	}
	if c.Version != GameConfig.Version {
		log.Print("ignoring incompatible old config")
		RemoveItem(shamoguconfig)
		return false, nil
	}
	GameConfig = *c
	return true, nil
}

// WriteDump writes full game statistics (to element with "dump" id) and
// returns a success log message or an error.
func (g *Game) WriteDump() (string, error) {
	pre := js.Global().Get("document").Call("getElementById", "dump")
	pre.Set("innerHTML", g.Dump())
	return "Game statistics written below.", nil
}

// subSig is define here for build compatibility purposes, it is not used for
// the js backend.
func subSig(ctx context.Context, msgs chan<- gruid.Msg) {
	// do nothing
}
