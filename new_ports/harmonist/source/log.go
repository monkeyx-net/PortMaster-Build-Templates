package main

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"

	"codeberg.org/anaseto/gruid/ui"
)

// logEntry describes a log entry.
type logEntry struct {
	Text  string
	MText string
	Index int
	Tick  bool
	Style logStyle
	Dups  int
}

// logStyle describes the various kinds of logging styles.
type logStyle int

const (
	logNormal logStyle = iota
	logNotable
	logDamage
	logSpecial
	logStatusEnd
	logError
	logConfirm
)

// Rune returns the coloring markup rune for the style.
func (st logStyle) Rune() rune {
	var r rune
	switch st {
	case logNotable:
		r = 'Y'
	case logDamage:
		r = 'O'
	case logSpecial:
		r = 'C'
	case logStatusEnd:
		r = 'V'
	case logError:
		r = 'R'
	case logConfirm:
		r = 'M'
	default:
		r = 'N'
	}
	return r
}

func (e logEntry) String() string {
	tick := ""
	if e.Tick {
		tick = "@y•@N "
	}
	s := e.Text
	if e.Dups > 0 {
		s += fmt.Sprintf(" (%d×)", e.Dups+1)
	}
	r := e.Style.Rune()
	if r != 0 {
		s = fmt.Sprintf("%s@%c%s@N", tick, r, s)
	} else {
		s = fmt.Sprintf("%s%s", tick, s)
	}
	return s
}

func (e logEntry) dumpString() string {
	tick := ""
	if e.Tick {
		tick = "• "
	}
	s := e.Text
	if e.Dups > 0 {
		s += fmt.Sprintf(" (%d×)", e.Dups+1)
	}
	s = fmt.Sprintf("%s%s", tick, s)
	return s
}

// UpperFirst returns a string with its first letter in upper case.
func UpperFirst(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || unicode.IsUpper(r) {
		return s
	}
	return string(unicode.ToUpper(r)) + s[utf8.RuneLen(r):]
}

func (g *Game) Log(s string) {
	e := logEntry{Text: UpperFirst(s), Index: g.LogIndex}
	g.LogEntry(e)
}

func (g *Game) LogStyled(s string, style logStyle) {
	e := logEntry{Text: UpperFirst(s), Index: g.LogIndex, Style: style}
	g.LogEntry(e)
}

func (g *Game) Logf(format string, a ...any) {
	e := logEntry{Text: UpperFirst(fmt.Sprintf(format, a...)), Index: g.LogIndex}
	g.LogEntry(e)
}

func (g *Game) LogfStyled(format string, style logStyle, a ...any) {
	e := logEntry{Text: UpperFirst(fmt.Sprintf(format, a...)), Index: g.LogIndex, Style: style}
	g.LogEntry(e)
}

// LogEntry adds a new entry to the game logs.
func (g *Game) LogEntry(e logEntry) {
	if e.Index == g.LogNextTick {
		e.Tick = true
	}
	if !e.Tick && len(g.Logs) > 0 {
		le := g.Logs[len(g.Logs)-1]
		if le.Text == e.Text {
			le.Dups++
			le.MText = le.String()
			g.Logs[len(g.Logs)-1] = le
			return
		}
	}
	e.MText = e.String()
	if LogGame {
		log.Printf("Depth %d:Turn %d:%v", g.Map.Depth, g.Turn, e.dumpString())
	}
	g.Logs = append(g.Logs, e)
	g.LogIndex++
	if len(g.Logs) > 100000 {
		g.Logs = g.Logs[10000:]
	}
}

func (g *Game) StoryPrint(s string) {
	g.Stats.Story = append(g.Stats.Story, fmt.Sprintf("Depth %2d|Turn %5d| %s", g.Map.Depth, g.Turn, s))
}

func (g *Game) StoryPrintf(format string, a ...any) {
	g.Stats.Story = append(g.Stats.Story, fmt.Sprintf("Depth %2d|Turn %5d| %s", g.Map.Depth, g.Turn, fmt.Sprintf(format, a...)))
}

// CrackSound produces a random crack message.
func (g *Game) CrackSound() (text string) {
	switch RandInt(4) {
	case 0:
		text = "CRACK!"
	case 1:
		text = "CRASH!"
	case 2:
		text = "CRUNCH!"
	case 3:
		text = "CREAK!"
	}
	return text
}

// ExplosionSound produces a random explosion message.
func (g *Game) ExplosionSound() (text string) {
	switch RandInt(3) {
	case 0:
		text = "BANG!"
	case 1:
		text = "POP!"
	case 2:
		text = "BOOM!"
	}
	return text
}

// DrawLog draws 2 compacted lines of log.
func (md *model) DrawLog() ui.StyledText {
	g := md.g
	stt := ui.StyledText{}.WithMarkups(Markups)
	tick := false
	for i := len(g.Logs) - 1; i >= 0; i-- {
		e := g.Logs[i]
		s := e.MText
		if stt.Text() != "" {
			if tick {
				s = s + "\n"
				tick = false
			} else {
				s = s + " "
			}
		}
		if e.Tick {
			tick = true
		}
		if stt.WithText(s+stt.Text()).Format(79).Size().Y > 2 {
			break
		}
		stt = stt.WithText(s + stt.Text()).Format(79)
	}
	return stt
}
