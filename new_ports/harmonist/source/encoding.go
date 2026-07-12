package main

import (
	"bytes"
	"compress/zlib"
	"encoding/gob"

	"codeberg.org/anaseto/gruid"
)

func init() {
	gob.Register(AbyssFallEvent{})
	gob.Register(&StatusEvent{})
	gob.Register(&MonsterTurnEvent{})
	gob.Register(&MonsterStatusEvent{})
	gob.Register(&PosEvent{})
	gob.Register(EndTurnEvent{})
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
	w.Close()
	return buf.Bytes(), nil
}

// Config describes available configuration options.
type Config struct {
	NormalModeKeys map[gruid.Key]Action
	TargetModeKeys map[gruid.Key]Action
	DarkColors     bool
	Tiles          bool
	Version        string
	ShowNumbers    bool
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
	r.Close()
	return lg, nil
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
