package main

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"

	"codeberg.org/anaseto/gruid/ui"
)

// Logs contains the logging information.
type Logs struct {
	Story    []string   // story log (for dump)
	Entries  []logEntry // all the game log entries
	Index    int        // index of next log entry
	NextTick int        // index of first log entry in a turn
}

// logEntry describes a log entry.
type logEntry struct {
	Text  string   // text for entry
	MText string   // text for entry with markup
	Index int      // index of entry in log
	Tick  bool     // whether first entry in a turn
	Style logStyle // style
	Dups  int      // number of duplicates of current entry
}

func (e logEntry) String() string {
	tick := ""
	if e.Tick {
		tick = "• "
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

// logStyle describes various logging styles.
type logStyle int

const (
	logNormal     logStyle = iota
	logConfirm             // UI message asking for confirmation
	logError               // UI or game error
	logHurtMons            // when monsters are hurt
	logHurtPlayer          // when player is hurt
	logNotable             // when you discover or hear something notable
	logSpecial             // important special message
	logStatusEnd           // when a player's status ends
)

// Rune returns the markup @rune corresponding to each log style.
func (st logStyle) Rune() rune {
	var r rune
	switch st {
	case logConfirm:
		r = 'M'
	case logError:
		r = 'R'
	case logHurtMons:
		r = 'G'
	case logHurtPlayer:
		r = 'O'
	case logNotable:
		r = 'Y'
	case logSpecial:
		r = 'C'
	case logStatusEnd:
		r = 'B'
	default:
		r = 'N'
	}
	return r
}

// UpperFirst returns a string with its first letter in upper case.
func UpperFirst(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[utf8.RuneLen(r):]
}

func (g *Game) Log(s string) {
	e := logEntry{Text: UpperFirst(s), Index: g.Logs.Index}
	g.LogEntry(e)
}

func (g *Game) LogStyled(s string, style logStyle) {
	e := logEntry{Text: UpperFirst(s), Index: g.Logs.Index, Style: style}
	g.LogEntry(e)
}

func (g *Game) Logf(format string, a ...interface{}) {
	e := logEntry{Text: UpperFirst(fmt.Sprintf(format, a...)), Index: g.Logs.Index}
	g.LogEntry(e)
}

func (g *Game) LogfStyled(format string, style logStyle, a ...interface{}) {
	e := logEntry{Text: UpperFirst(fmt.Sprintf(format, a...)), Index: g.Logs.Index, Style: style}
	g.LogEntry(e)
}

func (g *Game) Debug(format string, a ...interface{}) {
	g.LogfStyled(format, logError, a...)
}

func (g *Game) StoryLog(s string) {
	g.Logs.Story = append(g.Logs.Story, fmt.Sprintf("Level %d|Turn %5d| %s", g.Map.Level, g.Turn, s))
}

func (g *Game) StoryLogf(format string, a ...interface{}) {
	g.Logs.Story = append(g.Logs.Story, fmt.Sprintf("Level %d|Turn %5d| %s", g.Map.Level, g.Turn, fmt.Sprintf(format, a...)))
}

// LogEntry adds a new log entry to the game logs.
func (g *Game) LogEntry(e logEntry) {
	if e.Index == g.Logs.NextTick {
		e.Tick = true
	}
	if !e.Tick && len(g.Logs.Entries) > 0 {
		le := g.Logs.Entries[len(g.Logs.Entries)-1]
		if le.Text == e.Text {
			le.Dups++
			le.MText = le.String()
			g.Logs.Entries[len(g.Logs.Entries)-1] = le
			return
		}
	}
	e.MText = e.String()
	if LogGame {
		log.Printf("gamelog:L%d:T%d: %v", g.Map.Level, g.Turn, e.dumpString())
	}
	g.Logs.Entries = append(g.Logs.Entries, e)
	g.Logs.Index++
	if len(g.Logs.Entries) > 100000 {
		g.Logs.Entries = g.Logs.Entries[10000:]
	}
}

// DrawLog draws 2 compacted lines of log.
func (md *model) DrawLog() ui.StyledText {
	g := md.g
	stt := ui.StyledText{}.WithMarkups(Markups)
	tick := false
	for i := len(g.Logs.Entries) - 1; i >= 0; i-- {
		e := g.Logs.Entries[i]
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
