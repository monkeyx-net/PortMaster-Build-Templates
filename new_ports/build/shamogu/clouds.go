package main

import (
	"codeberg.org/anaseto/gruid"
)

// Cloud represents a cloud, which can be of some specific kind (like fire),
// and has a duration.
type Cloud struct {
	Kind     CloudKind
	P        gruid.Point
	Duration int
}

// CloudKind represents the possible kinds of clouds.
type CloudKind int

// These constants represent the possible kind of clouds.
const (
	NoCloud     CloudKind = iota
	CloudNormal           // obstructs vision (steam, dust, smoke)
	CloudFire             // burns & obstructs vision
	CloudPoison           // poisons & obstructs vision
)

func (ck CloudKind) String() string {
	switch ck {
	case CloudNormal:
		return "cloud"
	case CloudFire:
		return "fire"
	case CloudPoison:
		return "poison"
	default:
		return "cloud"
	}
}

// Color returns the color of the cloud.
func (ck CloudKind) Color() gruid.Color {
	switch ck {
	case CloudFire:
		return ColorRed
	case CloudPoison:
		return ColorGreen
	default:
		return ColorForeground
	}
}

// ColorRune returns the markup color rune of the cloud.
func (ck CloudKind) ColorRune() rune {
	switch ck {
	case CloudFire:
		return 'R'
	case CloudPoison:
		return 'G'
	default:
		return 'N'
	}
}

func (ck CloudKind) Desc() string {
	switch ck {
	case CloudNormal:
		return "The steam, dust or smoke cloud obstructs vision."
	case CloudFire:
		return "The fire cloud is difficult to see through. It burns foliage and creatures."
	case CloudPoison:
		return "The poisonous cloud is difficult to see through. It poisons creatures."
	default:
		// should not happen.
		return "No description."
	}
}

const FireDamage = 1
const PoisonDamage = 1

// CloudGrid is a representation of the clouds in the map that allows both for
// linear processing of current clouds, as well as constant-time checks of a
// particular position in the map looking for any clouds.
type CloudGrid struct {
	Clouds []Cloud
	Grid   []int // index in Clouds for a map position, or -1
}

// NewCloudGrid returns a new CloudGrid properly initialized.
func NewCloudGrid() *CloudGrid {
	cg := &CloudGrid{}
	cg.Grid = make([]int, MapWidth*MapHeight)
	cg.Clouds = []Cloud{}
	for i := range cg.Grid {
		cg.Grid[i] = -1 // no cloud
	}
	return cg
}

func point2idx(p gruid.Point) int {
	return p.Y*MapWidth + p.X
}

// func idx2point(i int) gruid.Point {
// 	return gruid.Point{X: i % MapWidth, Y: i / MapWidth}
// }

// Reset removes all the clouds
func (cg *CloudGrid) Reset() {
	for _, cl := range cg.Clouds {
		p := cl.P
		i := point2idx(p)
		if i < 0 || i >= MapWidth*MapHeight {
			// invalid cloud pos (should not happen)
			continue
		}
		cg.Grid[i] = -1 // no cloud
	}
	cg.Clouds = cg.Clouds[:0]
}

// At returns a cloud at a given position, if any.
func (cg *CloudGrid) At(p gruid.Point) Cloud {
	i := point2idx(p)
	if i < 0 || i >= MapWidth*MapHeight {
		return Cloud{}
	}
	if j := cg.Grid[i]; j >= 0 {
		return cg.Clouds[j]
	}
	return Cloud{}
}

// AddCloud adds a cloud at a given position, replacing any existing one unless
// it's of the same type and of higher duration.
func (g *Game) AddCloud(cl Cloud) {
	cg := g.Map.Clouds
	p := cl.P
	if !g.Map.Passable(p) {
		return
	}
	i := point2idx(p)
	if j := cg.Grid[i]; j >= 0 {
		cj := cg.Clouds[j]
		switch {
		case cl.Kind == cj.Kind:
			// If the cloud type is the same, we use the maximum
			// duration.
			cl.Duration = max(cl.Duration, cj.Duration)
		case cj.Kind == CloudFire:
			// Fire clouds take priority over other clouds: the
			// heat expands the air and dissipates others kinds of
			// clouds. This allows for protective usage of
			// foggy-skin onion in foliage without removing the
			// fires.
			return
		}
		cg.Clouds[j] = cl
		return
	}
	switch cl.Kind {
	case CloudFire:
		g.Stats.FireClouds++
	case CloudPoison:
		g.Stats.PoisonClouds++
	}
	j := len(cg.Clouds)
	cg.Clouds = append(cg.Clouds, cl)
	cg.Grid[i] = j
}

// SwapClouds swaps any clouds at valid positions p and q.
func (g *Game) SwapClouds(p, q gruid.Point) {
	cg := g.Map.Clouds
	pidx, qidx := point2idx(p), point2idx(q)
	i, j := cg.Grid[pidx], cg.Grid[qidx]
	cg.Grid[pidx], cg.Grid[qidx] = cg.Grid[qidx], cg.Grid[pidx]
	if i >= 0 {
		cg.Clouds[i].P = q
	}
	if j >= 0 {
		cg.Clouds[j].P = p
	}
}

func (cg *CloudGrid) removeCloud(j int) {
	nl := len(cg.Clouds) - 1 // new length
	cj, cnj := cg.Clouds[j], cg.Clouds[nl]
	i, ni := point2idx(cj.P), point2idx(cnj.P)
	// We update values in Grid so that the last cloud in Clouds is now at
	// index j.
	cg.Grid[ni] = j
	cg.Grid[i] = -1
	cg.Clouds[j], cg.Clouds[nl] = cg.Clouds[nl], cg.Clouds[j]
	cg.Clouds = cg.Clouds[:nl]
}

// EndCloud ends a cloud at a given index. It leaves smoke and destroys foliage
// for fire clouds on foliage.
func (g *Game) EndCloud(j int) {
	cg := g.Map.Clouds
	if j < 0 || j >= len(cg.Clouds) {
		// Should not happen.
		return
	}
	cj := cg.Clouds[j]
	if cj.Kind == CloudFire && g.Map.Terrain.At(cj.P) == Foliage {
		// EndCloud is used for cloud expiration, so when we have a
		// cloud of fire on foliage, we replace it with a new normal
		// cloud (smoke) instead and remove the foliage.
		g.Map.Terrain.Set(cj.P, Floor)
		cj.Kind = CloudNormal
		cj.Duration = 4 + g.IntN(5)
		cg.Clouds[j] = cj
		return
	}
	cg.removeCloud(j)
}

// RemoveCloudAt removes any cloud at the given (valid) position.
func (g *Game) RemoveCloudAt(p gruid.Point) {
	cg := g.Map.Clouds
	i := point2idx(p)
	if j := cg.Grid[i]; j >= 0 {
		cg.removeCloud(j)
	}
}

// fireCloudDuration returns the typical duration of a fire cloud.
func (g *Game) fireCloudDuration() int {
	return 6 + g.IntN(7)
}

// poisonCloudDuration returns the typical duration of a poisonous cloud.
func (g *Game) poisonCloudDuration() int {
	return 7 + g.IntN(5)
}

// UpdateClouds reduces the duration of all clouds by one, and removes any
// cloud whose new duration is zero or negative.
func (g *Game) UpdateClouds() {
	cg := g.Map.Clouds
	for i := 0; i < len(cg.Clouds); {
		cl := &cg.Clouds[i]
		cl.Duration--
		if cl.Duration <= 0 {
			g.EndCloud(i)
		} else {
			i++
		}
	}
	for _, cl := range cg.Clouds {
		switch cl.Kind {
		case CloudFire:
			if j, aj := g.ActorAt(cl.P); j >= 0 {
				g.PutStatus(j, aj, StatusFire, DurationFire)
			}
			for p := range NeighborsFunc(cl.P, g.Map.Burnable) {
				if cg.At(p).Kind == NoCloud && g.IntN(3) == 0 {
					g.FireCloudAt(p)
				}
			}
		case CloudPoison:
			if j, aj := g.ActorAt(cl.P); j >= 0 {
				g.PutStatus(j, aj, StatusPoison, DurationPoisonCloud)
			}
		}
	}
}

// NormalCloudAt adds a normal cloud (steam, dust, smoke) at the given position.
func (g *Game) NormalCloudAt(p gruid.Point, d int) {
	g.AddCloud(Cloud{Kind: CloudNormal, P: p, Duration: d})
}

// FireCloudAt adds a fire cloud at the given position.
func (g *Game) FireCloudAt(p gruid.Point) {
	g.AddCloud(Cloud{Kind: CloudFire, P: p, Duration: g.fireCloudDuration()})
}

// PoisonCloudAt adds a poisonous cloud at the given position.
func (g *Game) PoisonCloudAt(p gruid.Point) {
	g.AddCloud(Cloud{Kind: CloudPoison, P: p, Duration: g.poisonCloudDuration()})
}

// FireAt reports whether there's a fire at the given position.
func (g *Game) FireAt(p gruid.Point) bool {
	return g.Map.Clouds.At(p).Kind == CloudFire
}

// FireAt reports whether there's a fire at the given position.
func (g *Game) CloudAt(p gruid.Point) bool {
	return g.Map.Clouds.At(p).Kind != NoCloud
}
