package main

import (
	"fmt"
	"sort"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

type status int

const (
	StatusExhausted status = iota
	StatusSwift
	StatusLignification
	StatusConfusion
	StatusFlames // fake status
	StatusHidden
	StatusUnhidden
	StatusLight
	StatusDig
	StatusLevitation
	StatusShadows
	StatusIlluminated
	StatusTransparent
	StatusDisguised
	StatusDelay
	StatusDispersal
)

func (st status) Flag() bool {
	switch st {
	case StatusFlames, StatusHidden, StatusUnhidden, StatusLight:
		return true
	default:
		return false
	}
}

func (st status) Clean() bool {
	switch st {
	case StatusDelay:
		return true
	default:
		return false
	}
}

func (st status) Info() bool {
	switch st {
	case StatusFlames, StatusHidden, StatusUnhidden, StatusLight, StatusDelay:
		return true
	}
	return false
}

func (st status) Good() bool {
	switch st {
	case StatusSwift, StatusDig, StatusHidden, StatusLevitation, StatusShadows, StatusTransparent, StatusDisguised, StatusDispersal:
		return true
	default:
		return false
	}
}

func (st status) Bad() bool {
	switch st {
	case StatusConfusion, StatusUnhidden, StatusIlluminated:
		return true
	default:
		return false
	}
}

func (st status) String() string {
	switch st {
	case StatusExhausted:
		return "Exhaustion"
	case StatusSwift:
		return "Swiftness"
	case StatusLignification:
		return "Lignification"
	case StatusConfusion:
		return "Confusion"
	case StatusFlames:
		return "Flames"
	case StatusHidden:
		return "Hidden"
	case StatusUnhidden:
		return "Unhidden"
	case StatusDig:
		return "Dig"
	case StatusLight:
		return "Light"
	case StatusLevitation:
		return "Levitation"
	case StatusShadows:
		return "Shadows"
	case StatusIlluminated:
		return "Illumination"
	case StatusTransparent:
		return "Transparency"
	case StatusDisguised:
		return "Disguise"
	case StatusDelay:
		return "Delay"
	case StatusDispersal:
		return "Dispersal"
	default:
		// should not happen
		return "unknown"
	}
}

func (st status) Desc() string {
	switch st {
	case StatusExhausted:
		return "Forbids jumping off walls or over monsters."
	case StatusSwift:
		return "Allows moving for free without actually spending real turns for the duration, as if time had stopped for everyone but you. Prevents exhaustion from jumps."
	case StatusLignification:
		return "Forbids movement but gives extra temporary health. Prevents blinking and teleportation."
	case StatusConfusion:
		return "Forbids magara use."
	case StatusFlames:
		return "Surrounded by magical flames. The flames are harmless by themselves, but they will eventually dissolve into confusing purple mist. The purple mist will make monsters sleep, too."
	case StatusHidden:
		return "No monster can see you."
	case StatusUnhidden:
		return "Some monsters can see you."
	case StatusDig:
		return "Allows walking into walls like an earth dragon thanks to destructive oric magic."
	case StatusLight:
		return "You are in a lighted cell. Monsters may see you from a greater distance."
	case StatusLevitation:
		return "Allows to fly over chasm and oric barriers thanks to oric wind magic."
	case StatusShadows:
		return "While standing on a dark cell, only adjacent monsters will be able to see you. The harmonic shadows don’t affect your visibility on lighted cells."
	case StatusIlluminated:
		return "You are magically lighted. Monsters may see you from a greater distance unless you’re hidden in a place that covers the light."
	case StatusTransparent:
		return "Feeds surrounding light to harmonic magic to make you transparent. When standing on a lighted cell, only adjacent monsters will be able to see you. It does not affect your visibility on dark cells."
	case StatusDisguised:
		return "Surrounds you with harmonic illusions that make you look like a guard. As a result, most non-hunting monsters will ignore you. Monsters with good sense of smell may sense you from nearby if less than 3 tiles away."
	case StatusDelay:
		return "Time remaining before the trigger."
	case StatusDispersal:
		return "Any monster attempting to hit you will blink away due to an oric space-distortion."
	default:
		// should not happen
		return "unknown"
	}
}

func (st status) Short() string {
	switch st {
	case StatusExhausted:
		return "Exh"
	case StatusSwift:
		return "Sw"
	case StatusLignification:
		return "Lig"
	case StatusConfusion:
		return "Con"
	case StatusFlames:
		return "Fl"
	case StatusHidden:
		return "H+"
	case StatusUnhidden:
		return "H-"
	case StatusDig:
		return "Dig"
	case StatusLight:
		return "Lit"
	case StatusLevitation:
		return "Lev"
	case StatusShadows:
		return "Sh"
	case StatusIlluminated:
		return "Ill"
	case StatusTransparent:
		return "Tr"
	case StatusDisguised:
		return "Dgs"
	case StatusDelay:
		return "Del"
	case StatusDispersal:
		return "Dp"
	default:
		// should not happen
		return "?"
	}
}

func (md *model) statusHPColor() rune {
	g := md.g
	hpColor := 'G'
	switch g.Player.HP + g.Player.HPbonus {
	case 1:
		hpColor = 'R'
	case 2:
		hpColor = 'O'
	case 3:
		hpColor = 'Y'
	case 4:
		if g.Player.HPMax() > 4 {
			hpColor = 'Y'
		}
	}
	return hpColor
}

func (md *model) statusMPColor() rune {
	g := md.g
	mpColor := 'B'
	switch g.Player.MP {
	case 1, 2:
		mpColor = 'M'
	case 3:
		mpColor = 'V'
	case 4:
		mpColor = 'C'
	}
	return mpColor
}

func (md *model) sortedStatuses() statusSlice {
	g := md.g
	sts := statusSlice{}
	if cld, ok := g.Map.Clouds[g.Player.P]; ok && cld == CloudFire {
		g.Player.Statuses[StatusFlames] = 1
		defer func() {
			g.Player.Statuses[StatusFlames] = 0
		}()
	}
	for st, c := range g.Player.Statuses {
		if c > 0 {
			sts = append(sts, st)
		}
	}
	sort.Sort(sts)
	return sts
}

func (md *model) updateStatusInfo() {
	g := md.g
	var entries []ui.MenuEntry

	stt := ui.StyledText{}.WithMarkups(Markups)
	// depth
	var depth string
	if g.win {
		depth = " D:Out! "
	} else {
		depth = fmt.Sprintf(" D:%d ", g.Map.Depth)
	}
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(depth), Disabled: true})

	// turns
	entries = append(entries, ui.MenuEntry{Text: stt.WithTextf("T:%d ", g.Turn), Disabled: true})

	// HP
	nWounds := g.Player.HPMax() - g.Player.HP - g.Player.HPbonus
	if nWounds <= 0 {
		nWounds = 0
	}
	hpColor := md.statusHPColor()
	hps := "HP:"
	hp := g.Player.HP
	if hp < 0 {
		hp = 0
	}
	if !GameConfig.ShowNumbers {
		hps = fmt.Sprintf("%s@%c%s@B%s@N%s ",
			hps,
			hpColor,
			strings.Repeat("♥", hp),
			strings.Repeat("♥", g.Player.HPbonus),
			strings.Repeat("♥", nWounds),
		)
	} else {
		if g.Player.HPbonus > 0 {
			hps = fmt.Sprintf("%s@%c%d+%d/%d@N ", hps, hpColor, hp, g.Player.HPbonus, g.Player.HPMax())
		} else {
			hps = fmt.Sprintf("%s@%c%d/%d@N ", hps, hpColor, hp, g.Player.HPMax())
		}
	}
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(hps), Disabled: true})

	// MP
	MPspent := g.Player.MPMax() - g.Player.MP
	if MPspent <= 0 {
		MPspent = 0
	}
	mpColor := md.statusMPColor()
	mps := "MP:"
	if !GameConfig.ShowNumbers {
		mps = fmt.Sprintf("%s@%c%s@N%s ",
			mps,
			mpColor,
			strings.Repeat("♥", g.Player.MP),
			strings.Repeat("♥", MPspent),
		)
	} else {
		mps = fmt.Sprintf("%s@%c%d/%d@N ", mps, mpColor, g.Player.MP, g.Player.MPMax())
	}
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(mps), Disabled: true})

	// bananas
	bananas := fmt.Sprintf("@y)@N:%1d/%1d ", g.Player.Bananas, MaxBananas)
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(bananas), Disabled: true})

	// menus
	letter := func(r rune, active bool) string {
		if !active {
			return fmt.Sprintf("[%c]", r)
		}
		return fmt.Sprintf("@Y[%c]@N", r)
	}
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(letter('M', md.mode == modeMenu && md.menuMode == modeGameMenu))})
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(letter('I', md.mode == modeMenu && md.menuMode == modeInventory))})
	entries = append(entries, ui.MenuEntry{Text: stt.WithText(letter('V', md.mode == modeMenu && md.menuMode == modeEvocation))})
	_, interact := md.CanInteract()
	interactstt := stt.WithText("")
	if interact {
		interactstt = stt.WithText("[E]")
	}
	entries = append(entries, ui.MenuEntry{Text: interactstt, Disabled: !interact})

	// statuses
	sts := md.sortedStatuses()

	if len(sts) > 0 {
		// entries[5]
		entries = append(entries, ui.MenuEntry{Text: stt.WithText("| "), Disabled: true})
	}
	for _, st := range sts {
		r := 'Y'
		if st.Good() {
			r = 'B'
			t := DurationTurn
			if !st.Flag() && g.Player.Expire[st] >= g.Turn && g.Player.Expire[st]-g.Turn <= t {
				r = 'V'
			}
		} else if st.Bad() {
			r = 'R'
		}
		var sttext string
		if !st.Flag() {
			sttext = fmt.Sprintf("@%c%s(%d)@N ", r, st.Short(), g.Player.Statuses[st]/DurationStatusStep)
		} else {
			sttext = fmt.Sprintf("@%c%s@N ", r, st.Short())
		}
		entries = append(entries, ui.MenuEntry{Text: stt.WithText(sttext), Disabled: true})
	}

	switch {
	case g.Player.HP <= 0:
		entries = append(entries, ui.MenuEntry{
			Text:     stt.WithText("@RDead@N "),
			Disabled: true})
	case g.md.mode == modeStory:
		entries = append(entries, ui.MenuEntry{
			Text:     stt.WithText("@C[Story]@N "),
			Disabled: true})
	case g.win:
		entries = append(entries, ui.MenuEntry{
			Text:     stt.WithText("@BEscaped!@N "),
			Disabled: true})
	case md.mode == modeEnd:
	default:
		p := g.Player.P
		t := g.Map.At(p)
		if t != GroundCell {
			r, fg := g.TerrainStyle(t, p)
			if g.Illuminated(p) && IsIlluminable(t) && fg == ColorForeground {
				fg = ColorYellow
			}
			entries = append(entries, ui.MenuEntry{
				Text:     stt.WithTextf("%c ", r).WithStyle(gruid.Style{Fg: fg, Attrs: AttrInMap}),
				Disabled: true})
		}
		if md.g.resting {
			entries = append(entries, ui.MenuEntry{Text: stt.WithText("@CResting...@N"), Disabled: true})
		} else if md.targ.kbTargeting {
			entries = append(entries, ui.MenuEntry{Text: stt.WithText("@C[Examine]@N"), Disabled: true})
		}
	}
	md.status.SetEntries(entries)
}
