// This file defines the Update method for the model.

package main

import (
	"fmt"
	"log"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/ui"
)

// Update implements Update() for gruid.Model.
func (md *model) Update(msg gruid.Msg) gruid.Effect {
	if _, ok := msg.(gruid.MsgInit); ok {
		return md.init()
	}
	if _, ok := msg.(gruid.MsgQuit); ok {
		if md.mode != modeQuitting {
			// Save game before quitting.
			if err := md.g.Save(); err != nil {
				log.Printf("saving before quitting: %v", err)
			}
		}
		md.mode = modeQuitting
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
		md.auto.mode = noAuto
	}
	eff := md.update(msg)
	if cmd := md.AnimCmd(); !anims && cmd != nil {
		return gruid.Batch(eff, cmd)
	}
	return eff
}

func (md *model) interrupt(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return true
	case gruid.MsgMouse:
		return msg.Action != gruid.MouseMove && msg.Action != gruid.MouseRelease
	}
	return false
}

func (md *model) update(msg gruid.Msg) gruid.Effect {
	md.action = ActionNone{}
	md.g.instant = false
	switch md.mode {
	case modeLoadGame:
		switch msg := msg.(type) {
		case gruid.MsgKeyDown:
			md.updateStatus()
			md.mode = modeNormal
		case gruid.MsgMouse:
			if msg.Action != gruid.MouseMove && msg.Action != gruid.MouseRelease {
				md.updateStatus()
				md.mode = modeNormal
			}
		}
	case modeNewGame:
		return md.updateSpiritSelectionMenu(msg)
	case modeNormal:
		md.updateNormal(msg)
	case modeCritical:
		if md.more(msg) {
			md.mode = modeNormal
			md.g.Log("Ok. Be careful, then.")
		}
	case modePager:
		return md.updatePager(msg)
	case modeMenu:
		md.updateMenu(msg)
	case modeWizardConfirmation:
		st := md.updateConfirmation(msg)
		md.action = ActionWizardConfirm{State: st}
	case modeQuitConfirmation:
		st := md.updateConfirmation(msg)
		md.action = ActionQuitConfirm{State: st}
	case modeQuitting:
		// Do nothing.
		return nil
	case modeEnd:
		if md.more(msg) {
			md.gameEnded = true
			md.mode = modePager
			md.pager.mode = modeDump
			md.dump(md.g.WriteDump())
		}
	}
	md.g.UpdateFirePos(PlayerID, md.g.PlayerActor())
	eff, done := md.action.Handle(md)
	if done {
		md.endTurn()
		md.RefreshExamineInfo()
	}
	md.updateStatus()
	return eff
}

func (md *model) more(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case gruid.KeyEscape, gruid.KeySpace:
			return true
		}
	case gruid.MsgMouse:
		if msg.Action == gruid.MouseSecondary {
			return true
		}
	}
	return false
}

// confirm describes the possible results a confirmation update on message.
type confirm int

const (
	confirmTrue confirm = iota
	confirmFalse
	confirmPass
)

func (md *model) updateConfirmation(msg gruid.Msg) confirm {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		if msg.Key == "y" || msg.Key == "Y" {
			return confirmTrue
		}
		return confirmFalse
	case gruid.MsgMouse:
		if msg.Action == gruid.MouseSecondary {
			return confirmFalse
		}
	}
	return confirmPass
}

func (md *model) updateNormal(msg gruid.Msg) {
	switch msg := msg.(type) {
	case msgAuto:
		md.action = actionAuto{msg: msg}
	case gruid.MsgKeyDown:
		md.updateKeyDown(msg)
	case gruid.MsgMouse:
		md.updateMouse(msg)
	}
}

func (md *model) updateKeyDown(msg gruid.MsgKeyDown) {
	md.status.focus = false
	if !md.targ.kb && inMap(md.targ.p) && msg.Key != gruid.KeyPageDown && msg.Key != gruid.KeyPageUp {
		md.targ.CancelExamine()
	}
	a := md.keysNormal[msg.Key]
	if md.targ.kb {
		b := md.keysTarget[msg.Key]
		if b != nil {
			a = b
		}
	}
	shift := msg.Mod&gruid.ModShift != 0
	if shift && !msg.Key.IsRune() {
		switch b := a.(type) {
		case ActionBump:
			a = ActionRun(b)
		case ActionCursorBump:
			a = ActionCursorRun(b)
		}
	}
	if a != nil {
		md.action = a
	}
}

func (md *model) updateMouse(msg gruid.MsgMouse) {
	if msg.P.Y == UIHeight-1 {
		md.updateStatusMouse(msg)
		return
	}
	md.status.focus = false
	md.status.menu.SetActive(0)
	p := msg.P.Add(gruid.Point{0, -2}) // relative position ignoring log
	switch msg.Action {
	case gruid.MouseWheelUp:
		md.action = ActionScroll{Delta: gruid.Point{0, 1}}
	case gruid.MouseWheelDown:
		md.action = ActionScroll{Delta: gruid.Point{0, -1}}
	case gruid.MouseMove:
		if md.auto.mode == noAuto && !md.targ.kb {
			md.action = ActionExamine{Target: p}
		}
	case gruid.MouseSecondary:
		md.action = ActionMenu{}
	case gruid.MouseMain:
		if !inMap(p) {
			if p.Y >= -2 && p.Y <= 0 {
				md.action = ActionViewMessages{}
			}
			return
		}
		pp := md.g.PP()
		if md.targ.p != p {
			md.Examine(p)
			return
		}
		if paths.DistanceManhattan(p, pp) == 1 {
			md.action = ActionBump{Delta: p.Sub(pp)}
		} else {
			md.action = ActionTravel{}
		}
	}
}

func (md *model) updateStatusMouse(msg gruid.MsgMouse) {
	msg.P.Y = 0
	if !msg.P.In(md.status.menu.Bounds()) || md.targ.kb {
		md.status.focus = false
		return
	}
	md.status.menu.Update(msg)
	update := !md.status.focus
	switch md.status.menu.Action() {
	case ui.MenuMove:
		update = true
	case ui.MenuInvoke:
		i := statusEntry(md.status.menu.Active())
		switch i {
		case statusMenu:
			md.status.focus = false
			md.action = ActionMenu{}
			return
		default:
			update = true
		}
	}
	if update {
		md.targ.CancelExamine()
		const statusIndex = statusDefense + 1
		i := statusEntry(md.status.menu.Active())
		md.status.focus = false
		switch {
		case i == statusLevel:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Level")}
			md.status.desc.SetText("Current map level.")
			md.status.focus = true
		case i == statusTurns:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Turns")}
			md.status.desc.SetText("Number of elapsed turns since the start.")
			md.status.focus = true
		case i == statusMenu:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Menu")}
			md.status.desc.SetText("Click to open menu or press either the Space or M key.")
			md.status.focus = true
		case i == statusDirection:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Direction")}
			md.status.desc.SetText("The direction of the last attack or movement.")
			md.status.focus = true
		case i == statusHP:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Health")}
			md.status.desc.SetText("Your hit points.")
			md.status.focus = true
		case i == statusAttack:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Attack")}
			md.status.desc.SetText("Your attack attribute.")
			md.status.focus = true
		case i == statusDefense:
			md.status.desc.Box = &ui.Box{Title: ui.Text("Defense")}
			md.status.desc.SetText("Your defense attribute.")
			md.status.focus = true
		case i >= statusIndex:
			i := md.status.menu.Active() - int(statusIndex)
			var st Status = -1
			var n int
			pa := md.g.PlayerActor()
			for j, turns := range pa.Statuses {
				if turns <= 0 {
					continue
				}
				if n == i {
					st = Status(j)
					break
				}
				n++
			}
			if st < 0 {
				if !pa.IsDead() && !md.g.win {
					md.status.desc.Box = &ui.Box{Title: ui.Text("Player’s Tile")}
					md.status.desc.SetText("The tile where the player is.")
					md.status.focus = true
				}
				break
			}
			turns := pa.Statuses[st]
			title := ui.StyledText{}.WithMarkups(Markups).
				WithTextf("@%c%s@N (for %d %s)",
					statusColor(st, turns), st.Name(), turns, Countable("turns", turns))
			md.status.desc.Box = &ui.Box{Title: title}
			md.status.desc.SetText(st.Desc())
			md.status.desc.Content = md.status.desc.Content.Format(UIWidth - 2)
			md.status.focus = true
		}
	}
}

// endTurn finalizes player's turn and runs other events until next player
// turn.
func (md *model) endTurn() {
	md.mode = modeNormal
	g := md.g
	g.EndTurn()
	pa := g.PlayerActor()
	if pa.DoesAny(GoodSmell) {
		g.SmellFood()
	}
	//g.TurnStats()
	hp := pa.HP
	if hp <= 0 {
		md.death()
		return
	}
	if md.mode == modeCritical {
		if hp <= HPCritical {
			g.LogStyled("*** CRITICAL HP WARNING ***", logConfirm)
		}
		md.logConfirmContinue()
	}
	g.Logs.NextTick = g.Logs.Index
}

func (md *model) death() {
	g := md.g
	g.StoryLog("Died")
	g.LogStyled("You die…", logSpecial)
	g.LevelStats()
	md.logConfirmContinue()
	md.mode = modeEnd
}

func (md *model) logConfirmContinue() {
	md.g.LogStyled("[SPACE/ESC to continue]", logConfirm)
	md.targ.CancelExamine()
}

func (md *model) dump(msg string, err error) {
	s := md.g.DumpSummary()
	lines := strings.Split(s, "\n")
	stts := []ui.StyledText{}
	for _, l := range lines {
		stts = append(stts, ui.Text(l))
	}
	var details string
	if err != nil {
		details = fmt.Sprintf("Error writing dump: %v.", err)
	} else {
		details = msg
	}
	stts = append(stts, ui.Text(details))
	md.pager.pg.SetLines(stts)
	md.pager.pg.SetCursor(gruid.Point{0, 0})
}
