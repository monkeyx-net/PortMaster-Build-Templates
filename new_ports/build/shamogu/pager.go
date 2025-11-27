package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// pager gathers the structures used in the various kinds of pagers.
type pager struct {
	pg     *ui.Pager       // pager widget
	mode   pagerMode       // pager mode
	markup ui.StyledText   // styled text with default markup for pager
	lines  []ui.StyledText // log lines cache (unused in dump & help modes)
}

// pagerMode represents the available pager modes
type pagerMode int

const (
	modeLogs pagerMode = iota
	modeDump
	modeHelp
)

func (md *model) updatePager(msg gruid.Msg) gruid.Effect {
	switch md.pager.mode {
	case modeDump:
		md.pager.pg.Update(msg)
		if md.pager.pg.Action() == ui.PagerQuit {
			md.mode = modeQuitting
			return gruid.End()
		}
	default:
		md.pager.pg.Update(msg)
		if md.pager.pg.Action() == ui.PagerQuit {
			md.mode = modeNormal
		}
	}
	return nil
}
