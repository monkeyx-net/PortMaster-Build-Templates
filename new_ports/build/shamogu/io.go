//go:build !js

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// DataDir returns the game's data directory location.
func DataDir() (string, error) {
	var xdg string
	if runtime.GOOS == "windows" {
		xdg = os.Getenv("LOCALAPPDATA")
	} else {
		xdg = os.Getenv("XDG_DATA_HOME")
	}
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	dataDir := filepath.Join(xdg, "shamogu")
	_, err := os.Stat(dataDir)
	if err != nil {
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			return "", fmt.Errorf("building data directory: %v", err)
		}
	}
	return dataDir, nil
}

// SaveFile writes save data to the given file.
func SaveFile(filename string, data []byte) error {
	dataDir, err := DataDir()
	if err != nil {
		return err
	}
	tempSaveFile := filepath.Join(dataDir, "temp-"+filename)
	f, err := os.OpenFile(tempSaveFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	saveFile := filepath.Join(dataDir, filename)
	if err := os.Rename(f.Name(), saveFile); err != nil {
		return err
	}
	return err
}

// Save saves the game to the save file.
func (g *Game) Save() error {
	data, err := g.GameSave()
	if err != nil {
		g.Log(err.Error())
		return err
	}
	err = SaveFile("save", data)
	if err != nil {
		g.Log(err.Error())
	}
	return err
}

// RemoveSaveFile removes the save file if it exists.
func RemoveSaveFile() error {
	return RemoveDataFile("save")
}

// RemoveReplay removes the replay file if it exists.
func RemoveReplay() {
	if err := RemoveDataFile("replay.part"); err != nil {
		log.Printf("removing replay: %v", err)
	}
}

// Load loads an existing game from a saved file.
func (g *Game) Load() (bool, error) {
	dataDir, err := DataDir()
	if err != nil {
		return false, err
	}
	saveFile := filepath.Join(dataDir, "save")
	_, err = os.Stat(saveFile)
	if err != nil {
		// no save file, new game
		return false, nil
	}
	data, err := os.ReadFile(saveFile)
	if err != nil {
		return false, err
	}
	lg, err := g.DecodeGameSave(data)
	if err != nil {
		return false, err
	}
	if lg.Version != Version {
		log.Print("ignoring incompatible old save")
		if err := RemoveDataFile("save"); err != nil {
			log.Printf("removing old save: %v", err)
		}
		return false, nil
	}
	*g = *lg
	return true, nil
}

// SaveConfig saves the game's config to the config file.
func SaveConfig() error {
	data, err := GameConfig.ConfigSave()
	if err != nil {
		return err
	}
	return SaveFile("config", data)
}

// LoadConfig loads game's config from the config file.
func LoadConfig() (bool, error) {
	dataDir, err := DataDir()
	if err != nil {
		return false, err
	}
	configFile := filepath.Join(dataDir, "config")
	_, err = os.Stat(configFile)
	if err != nil {
		// no save file, new game
		return false, nil
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return false, err
	}
	c, err := DecodeConfigSave(data)
	if err != nil {
		return false, err
	}
	if c.Version != GameConfig.Version {
		log.Print("ignoring incompatible old config")
		if err := RemoveDataFile("config"); err != nil {
			log.Printf("removing old config: %v", err)
		}
		return false, nil
	}
	GameConfig = *c
	return true, nil
}

// RemoveDataFile removes the given file in the game's data directory.
func RemoveDataFile(file string) error {
	dataDir, err := DataDir()
	if err != nil {
		return err
	}
	dataFile := filepath.Join(dataDir, file)
	_, err = os.Stat(dataFile)
	if err == nil {
		err := os.Remove(dataFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteDump writes full game statistics and returns a success log message or
// an error.
func (g *Game) WriteDump() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dataDir, "dump.txt")
	err = os.WriteFile(path, []byte(g.Dump()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing game statistics: %v", err)
	}
	return fmt.Sprintf("Game statistics written to %s.", path), nil
}
