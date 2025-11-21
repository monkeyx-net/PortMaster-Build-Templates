package main

import (
	"bytes"
	"compress/zlib"
	"encoding/gob"

	"codeberg.org/anaseto/gruid"
)

func init() {
	// Effects.
	gob.Register(EffectAmbrosiaBerries{})
	gob.Register(EffectBerserkingFlower{})
	gob.Register(EffectLignificationFruit{})
	gob.Register(EffectClarityLeaves{})
	gob.Register(EffectFoggySkinOnion{})
	gob.Register(EffectFirebreathPepper{})
	gob.Register(EffectTeleportMushroom{})

	gob.Register(EffectFocus{})
	gob.Register(EffectDig{})
	gob.Register(EffectJump{})
	gob.Register(EffectPushingGale{})
	gob.Register(EffectTimeStop{})
	gob.Register(EffectLightning{})
	gob.Register(EffectBark{})
	gob.Register(EffectNoxiousSmell{})
	gob.Register(EffectLignify{})
	gob.Register(EffectPoisonCloud{})
	gob.Register(EffectSprint{})
	gob.Register(EffectFireRetreat{})

	gob.Register(EffectEarthMenhir{})
	gob.Register(EffectWarpingMenhir{})
	gob.Register(EffectPoisonMenhir{})
	gob.Register(EffectFireMenhir{})

	// Actions in key-config.
	gob.Register(ActionAutoExplore{})
	gob.Register(ActionBump{})
	gob.Register(ActionConfig{})
	gob.Register(ActionCursorBump{})
	gob.Register(ActionCursorRun{})
	gob.Register(ActionDump{})
	gob.Register(ActionExamineModeToggle{})
	gob.Register(ActionHelp{})
	gob.Register(ActionInteract{})
	gob.Register(ActionInventory{})
	gob.Register(ActionMenu{})
	gob.Register(ActionNextItem{})
	gob.Register(ActionNextMonster{})
	gob.Register(ActionNone{})
	gob.Register(ActionPreviousMonster{})
	gob.Register(ActionQuit{})
	gob.Register(ActionRun{})
	gob.Register(ActionSaveQuit{})
	gob.Register(ActionSetKeys{})
	gob.Register(ActionScroll{})
	gob.Register(ActionTravel{})
	gob.Register(ActionViewMessages{})
	gob.Register(ActionWait{})
	gob.Register(ActionWizard{})

	// Items.
	gob.Register(&Spirit{})
	gob.Register(&Comestible{})
	gob.Register(&Menhir{})
	gob.Register(&Portal{})
	gob.Register(&CorruptionOrb{})
	gob.Register(&RunicTrap{})
	gob.Register(&EmptyTotem{})

	// Other roles.
	gob.Register(&Actor{})
}

// GameSave returns encoded game data for saving.
func (g *Game) GameSave() ([]byte, error) {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)
	err := enc.Encode(g)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(data.Bytes())
	err = w.Close()
	return buf.Bytes(), err
}

// Config describes available configuration options.
type Config struct {
	DarkColors      bool                 // whether to use a dark color theme
	ExamineModeKeys map[gruid.Key]Action // custom examine mode keys
	NormalModeKeys  map[gruid.Key]Action // custom normal mode keys
	Tiles           bool                 // whether to use Tiles or Unicode
	Version         string               // config's game version
	WarnFirePoison  bool                 // whether to stop and warn about fire/poison
}

// ConfigSave returns encoded config data for saving.
func (c *Config) ConfigSave() ([]byte, error) {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)
	err := enc.Encode(c)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

// DecodeGameSave retrieves a *Game object from game data encoded with
// GameSave.
func (g *Game) DecodeGameSave(data []byte) (*Game, error) {
	buf := bytes.NewReader(data)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(r)
	lg := &Game{}
	err = dec.Decode(lg)
	if err != nil {
		return nil, err
	}
	err = r.Close()
	return lg, err
}

// DecodeConfigSave retrieves a *config object from config data encoded with
// ConfigSave.
func DecodeConfigSave(data []byte) (*Config, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	c := &Config{}
	err := dec.Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
