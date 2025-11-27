package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// gameStatus represents the game's status bar and relevant data.
type gameStatus struct {
	menu  *ui.Menu
	desc  *ui.Label
	focus bool
}

// statusEntry represents the various kinds of entries in the status bar.
type statusEntry int

const (
	statusLevel statusEntry = iota
	statusTurns
	statusMenu
	statusDirection
	statusHP
	statusAttack
	statusDefense
)

func (md *model) updateStatus() {
	g := md.g
	var entries []ui.MenuEntry

	stt := ui.StyledText{}.WithMarkups(Markups)

	// Map level.
	level := fmt.Sprintf(" L:%d ", g.Map.Level)
	if md.g.win {
		level = " L:@GOut!@N "
	}
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(level), Disabled: true})

	// Turns.
	entries = append(entries, ui.MenuEntry{Text: stt.WithTextf("T:%d ", g.Turn), Disabled: true})

	// Menu button.
	if md.mode == modeMenu && md.menu.mode == modeGameMenu {
		entries = append(entries, ui.MenuEntry{Text: stt.WithTextf("@Y[M]@N")})
	} else {
		entries = append(entries, ui.MenuEntry{Text: stt.WithTextf("[M]")})
	}

	// Direction
	dirs := " @B.@N "
	switch md.g.Dir {
	case gruid.Point{1, 0}:
		dirs = " @B→@N "
	case gruid.Point{-1, 0}:
		dirs = " @B←@N "
	case gruid.Point{0, 1}:
		dirs = " @B↓@N "
	case gruid.Point{0, -1}:
		dirs = " @B↑@N "
	}
	entries = append(entries, ui.MenuEntry{
		Text:     stt.WithText(dirs),
		Disabled: true})

	// HP
	pa := g.PlayerActor()
	hps := pa.fmtHP()
	entries = append(entries, ui.MenuEntry{
		Text:     stt.WithTextf("HP:%s ", hps),
		Disabled: true})

	// attack
	entries = append(entries, ui.MenuEntry{
		Text:     stt.WithTextf("A:%s ", fmtAD(pa.GetAttack(), pa.Attack)),
		Disabled: true})

	// defense
	entries = append(entries, ui.MenuEntry{
		Text:     stt.WithTextf("D:%s ", fmtAD(pa.GetDefense(), pa.Defense)),
		Disabled: true})

	// statuses
	for i, turns := range pa.Statuses {
		if turns <= 0 {
			continue
		}
		st := Status(i)
		entries = append(entries, ui.MenuEntry{
			Text:     stt.WithTextf("@%c%s(%d)@N ", statusColor(st, turns), st.Abbr(), turns),
			Disabled: true})
	}
	switch {
	case pa.IsDead():
		entries = append(entries, ui.MenuEntry{
			Text:     stt.WithText("@RDead@N "),
			Disabled: true})
	case md.g.win:
	default:
		pp := g.PP()
		if i, _ := g.ItemAt(pp); i >= 0 {
			ei := g.Entity(i)
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithTextf("%c ", ei.Rune).WithStyle(gruid.Style{Fg: ei.Color(), Attrs: AttrInMap}),
				Disabled: true})
		} else if t := g.Map.Terrain.At(pp); t != Floor {
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithTextf("%c ", MapRune(t)).WithStyle(gruid.Style{Fg: ColorForeground, Attrs: AttrInMap}),
				Disabled: true})
		}
		clk := g.Map.Clouds.At(pp).Kind
		switch clk {
		case NoCloud:
		case CloudPoison:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@g§@N "),
				Disabled: true})
		case CloudFire:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@r§@N "),
				Disabled: true})
		case CloudNormal:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("§ ").WithStyle(gruid.Style{Attrs: AttrInMap}),
				Disabled: true})
		}
		if md.targ.kb {
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@C[Examine]@N "),
				Disabled: true})
		}
	}

	md.status.menu.SetEntries(entries)
}

func statusColor(st Status, turns int) rune {
	if st.Bad() {
		return 'O'
	}
	if turns <= 1 {
		return 'C'
	}
	return 'B'
}

func statusHPColor(hp, hpmax int) rune {
	switch {
	case hp <= 1 || hp <= hpmax/3:
		return 'O'
	case hp <= (3*hpmax)/4:
		return 'Y'
	default:
		return 'G'
	}
}

func (a *Actor) fmtHP() string {
	hp := a.HP
	if hp < 0 {
		hp = 0
	}
	debt := 0
	m := 1 // minimum return HP after bonus expiration (if above)
	if a.Has(StatusLignification) {
		debt += a.HPBonus()
		m = 3
	}
	if a.Has(StatusBerserk) {
		debt += a.HPBonus()
	}
	hpmax := a.MaxHP + debt
	base := hp - debt // HP without bonus
	if base < m {
		base = min(hp, m)
	}
	bonus := hp - base // HP bonus
	if base == 0 || bonus == 0 {
		hpColor := statusHPColor(hp, hpmax)
		return fmt.Sprintf("@%c%d/%d@N", hpColor, base, hpmax)
	}
	returnHPColor := statusHPColor(base, a.MaxHP)
	hpColor := statusHPColor(bonus, debt)
	return fmt.Sprintf("@%c%d@%c+%d/%d@N", returnHPColor, base, hpColor, bonus, hpmax)
}

func fmtAD(val, def int) string {
	color := 'N'
	switch {
	case val < def:
		color = 'O'
	case val > def:
		color = 'B'
	}
	return fmt.Sprintf("@%c%d@N", color, val)
}
