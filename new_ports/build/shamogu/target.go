package main

import (
	"codeberg.org/anaseto/gruid"
)

// targeting contains information related to targeting.
type targeting struct {
	info   targInfo    // information about target position
	kb     bool        // using keyboard examine mode
	p      gruid.Point // target
	scroll bool        // scroll target description
}

// targInfo contains information about target position.
type targInfo struct {
	entities []ID      // entities at target position
	cloud    CloudKind // cloud information (if any)

	sees    bool // player sees target currently
	unknown bool // target is unknown
}

// HideCursor hides the target cursor.
func (t *targeting) HideCursor() {
	t.p = InvalidPos
}

// SetCursor sets the target cursor.
func (t *targeting) SetCursor(p gruid.Point) {
	t.p = p
}

// CancelExamine cancels current targeting.
func (t *targeting) CancelExamine() {
	*t = targeting{}
	t.HideCursor()
}

// Examine targets a given position with the cursor (if target didn't change).
func (md *model) Examine(p gruid.Point) {
	if md.targ.p == p {
		return
	}
	md.examine(p)
}

// RefreshExamineInfo refreshes information for the current target position (if
// valid).
func (md *model) RefreshExamineInfo() {
	if md.targ.p != InvalidPos {
		md.examine(md.targ.p)
	}
}

// examine targets a given position with the cursor.
func (md *model) examine(p gruid.Point) {
	if !inMap(p) {
		return
	}
	md.targ.SetCursor(p)
	md.auto.path = md.g.PlayerPath(md.g.PP(), p)
	if len(md.auto.path) == 0 {
		md.auto.path = append(md.auto.path, p)
	}
	md.updateTargInfo()
	md.targ.scroll = false
}

func (md *model) updateTargInfo() {
	g := md.g
	pi := targInfo{}
	p := md.targ.p
	pi.unknown = !IsKnown(g.Map.KnownTerrain.At(p))
	if g.Wizard.Mode.Reveal() {
		if i, a := g.ActorAt(p); a != nil {
			pi.entities = append(pi.entities, i)
		}
		if i, _ := g.ItemAt(p); i >= 0 {
			pi.entities = append(pi.entities, i)
		}
	} else {
		if i, a := g.SensedActorAt(p); a != nil {
			pi.entities = append(pi.entities, i)
		}
		if i, _ := g.KnownItemAt(p); i >= 0 {
			pi.entities = append(pi.entities, i)
		}
	}
	pi.sees = g.InFOV(p)
	if pi.sees || g.Wizard.Mode.Reveal() {
		pi.cloud = g.Map.Clouds.At(p).Kind
	}
	md.targ.info = pi
}
