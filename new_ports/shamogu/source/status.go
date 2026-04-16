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
	level := fmt.Sprintf("L:%d ", g.Map.Level)
	if md.g.win {
		level = "L:@GOut!@N "
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
			Text:     stt.WithTextf("@%c%s(%d)@N ", statusColor(st, turns), g.StatusAbbr(pa, st), turns),
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
				Text:     stt.WithTextf("%c ", TerrainRune(t)).WithStyle(gruid.Style{Fg: ColorForeground, Attrs: AttrInMap}),
				Disabled: true})
		}
		clk := g.Map.Clouds.At(pp).Kind
		switch clk {
		case NoCloud:
		case CloudPoison:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@g≡@N "),
				Disabled: true})
		case CloudFire:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@r≡@N "),
				Disabled: true})
		case CloudNormal:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("≡ ").WithStyle(gruid.Style{Attrs: AttrInMap}),
				Disabled: true})
		}
		switch {
		case md.mode == modeCritical:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@O[Warning]@N "),
				Disabled: true})
		case md.mode == modeQuitConfirmation || md.mode == modeWizardConfirmation:
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithText("@C[Confirm]@N "),
				Disabled: true})
		case md.targ.kb:
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
		switch st {
		case StatusDisorient:
			// Disorient is special in that it only has an effect
			// during the monster's turn, so it is completely
			// useless on the last turn.
			return 'S'
		case StatusBerserk, StatusLignification, StatusShadow, StatusFoggySkin, StatusClarity:
			// Some statuses have both an effect during the
			// player's turn and the monsters' turn, so only
			// the effect during the former remains relevant.
			return 'C'
		default:
			// Other statuses don't have an effect during the
			// monsters' turn (except in Confusion+Imbalance that
			// has minimal impact with respect to last turn
			// considerations).
			return 'V'
		}
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
	hp := max(0, a.HP)
	var base, debt int // base HP without bonus, temporary hp bonus debt
	switch {
	case a.Has(StatusLignification) && a.Has(StatusBerserk):
		hpbonus := a.HPBonus()
		debt = 2 * hpbonus
		// When both statuses are active at the same time, we use as
		// base the HP the actor would have after both statuses expire.
		// Result depends on the (expected) expiration order.
		// Note that when both expire on the same turn, Berserk always
		// expires first.
		if a.Statuses[StatusBerserk] <= a.Statuses[StatusLignification] {
			nhp := min(hp, max(1, hp-hpbonus))
			base = min(nhp, max(3, nhp-hpbonus))
		} else {
			nhp := min(hp, max(3, hp-hpbonus))
			base = min(nhp, max(1, nhp-hpbonus))
		}
	case a.Has(StatusLignification):
		debt = a.HPBonus()
		base = min(hp, max(3, hp-debt))
	case a.Has(StatusBerserk):
		debt = a.HPBonus()
		base = min(hp, max(1, hp-debt))
	default:
		base = hp
	}
	hpmax := a.MaxHP + debt
	bonus := hp - base // HP bonus
	if base == 0 || bonus == 0 {
		hpColor := statusHPColor(hp, hpmax)
		return fmt.Sprintf("@%c%d/%d@N", hpColor, base, hpmax)
	}
	returnHPColor := statusHPColor(base, a.MaxHP)
	hpColor := statusHPColor(bonus, debt-base)
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
