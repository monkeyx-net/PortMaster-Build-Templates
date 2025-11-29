//go:build !js

package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

func main() {
	optNoAnim := flag.Bool("n", false, "no animations")
	optReplay := flag.String("r", "", "path to replay file (_ means default location)")
	optGameLogs := flag.Bool("l", false, "write game logs to log file")
	optVersion := flag.Bool("version", false, "print build info")
	opt16colors := new(bool)
	opt256colors := new(bool)
	optTrueColor := new(bool)
	optFullscreen, optNoAcceleration := new(bool), new(bool)
	optWidthScale, optHeightScale := new(float64), new(float64)
	if Tiles {
		optFullscreen = flag.Bool("F", false, "fullscreen")
		optNoAcceleration = flag.Bool("noaccel", false, "disable graphic acceleration")
		optWidthScale = flag.Float64("w", 1.0, "window width scale factor (default: 1.0, examples: 0.75, 1.25)")
		optHeightScale = flag.Float64("h", 1.0, "window height scale factor (default: 1.0, examples: 0.75, 1.25)")
	} else {
		opt16colors = flag.Bool("s", false, "use standard 16-color palette (default on most systems)")
		opt256colors = flag.Bool("x", false, "use xterm 256-color palette (solarized approximation)")
		optTrueColor = flag.Bool("t", false, "use true color selenized palette (not supported by all terminals)")
	}
	flag.Parse()

	if *optVersion {
		fmt.Printf("shamogu\t%v\n", Version)
		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Print(bi)
		}
		os.Exit(0)
	}
	if *optNoAnim {
		DisableAnimations = true
	}
	if runtime.GOOS == "windows" {
		ColorMode = ColorMode8
	}
	switch {
	case *opt256colors:
		ColorMode = ColorMode256
	case *opt16colors:
		ColorMode = ColorMode16
	case *optTrueColor:
		ColorMode = ColorMode24bit
	}
	log.SetPrefix("shamogu ")
	err := InitConfig()
	if err != nil {
		log.Print(err)
	}
	initDriver(*optFullscreen, *optWidthScale, *optHeightScale, *optNoAcceleration)
	if *optReplay != "" {
		RunReplay(*optReplay)
	} else {
		RunGame(*optGameLogs)
	}
}

// RunGame starts the game using the given path for the log file (if
// non-empty).
func RunGame(gamelogs bool) {
	gd := gruid.NewGrid(UIWidth, UIHeight)
	md := &model{gd: gd, g: &Game{}, targ: &targeting{}}
	var repw io.WriteCloser
	dir, err := DataDir()
	defer func() {
		if repw != nil {
			if err := repw.Close(); err != nil {
				log.Printf("closing replay file: %v", err)
			}
		}
		if md.gameEnded && dir != "" {
			if err := RemoveSaveFile(); err != nil {
				log.Printf("removing save file: %v", err)
			}
			_, err := os.Stat(filepath.Join(dir, "replay.part"))
			if err != nil {
				log.Printf("no replay file: %v", err)
				return
			}
			if err := os.Rename(filepath.Join(dir, "replay.part"), filepath.Join(dir, "replay")); err != nil {
				log.Printf("writing replay file: %v", err)
			}
		}
	}()
	if err == nil {
		replay, err := os.OpenFile(filepath.Join(dir, "replay.part"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			repw = replay
		} else {
			log.Printf("writing to replay file: %v", err)
		}
	} else {
		log.Print(err)
	}
	if gamelogs {
		LogGame = true
	}
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      driver,
		Model:       md,
		FrameWriter: repw,
	})
	if f := setLogOutput(); f != nil {
		defer func() {
			f.Close()
		}()
	}
	err = app.Start(context.Background())
	log.SetOutput(os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
}

// RunReplay runs the given replay file.
func RunReplay(file string) {
	if file == "_" {
		dir, err := DataDir()
		if err == nil {
			file = filepath.Join(dir, "replay")
		} else {
			log.Print(err)
		}
	}
	replay, err := os.Open(file)
	if err != nil {
		log.Fatalf("loading replay file: %v", err)
	}
	defer replay.Close()
	fd, err := gruid.NewFrameDecoder(replay)
	if err != nil {
		// Attempt again assuming base64-encoded replay (from web version).
		replay.Seek(0, io.SeekStart)
		b64replay := base64.NewDecoder(base64.StdEncoding, replay)
		var errb64 error
		fd, errb64 = gruid.NewFrameDecoder(b64replay)
		if errb64 != nil {
			log.Fatalf("frame decoder: %v", err)
		}
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
	if f := setLogOutput(); f != nil {
		defer func() {
			f.Close()
		}()
	}
	if err := app.Start(context.Background()); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
}

// setLogOutput sets standard log output to the logs file in the game's data
// directory.
func setLogOutput() *os.File {
	dataDir, err := DataDir()
	if err != nil {
		log.Print(err)
		return nil
	}
	logFile := filepath.Join(dataDir, "logs.txt")
	f, err := os.Create(logFile)
	if err != nil {
		log.Print(err)
		return nil
	}
	if Tiles {
		log.SetOutput(io.MultiWriter(f, os.Stderr))
	} else {
		log.SetOutput(f)
	}
	return f
}

// subSig is a subscription that intercepts SIGTERM for closing the game
// gracefully.
func subSig(ctx context.Context, msgs chan<- gruid.Msg) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	select {
	case <-ctx.Done():
	case <-sig:
		msgs <- gruid.MsgQuit{}
	}
}
