package main

import (
	_ "embed"
	"log"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

//go:embed data/small-vaults.txt
var smallVaults string

//go:embed data/big-vaults.txt
var bigVaults string

// SmallVaultTemplates contains the raw small vaults.
var SmallVaultTemplates []string

// BigVaultTemplates contains the raw big vaults.
var BigVaultTemplates []string

func init() {
	SmallVaultTemplates = splitVaults(smallVaults)
	BigVaultTemplates = splitVaults(bigVaults)
}

func splitVaults(s string) []string {
	vaults := strings.Split(strings.TrimSpace(strings.ReplaceAll(s, " ", "")), "\n\n")
	for _, v := range vaults {
		// We ensure all vaults contains essential places.
		if !strings.ContainsRune(v, 'W') {
			log.Fatalf("No W in vault:\n%s", v)
		}
		if !strings.ContainsRune(v, '>') {
			log.Fatalf("No > in vault:\n%s", v)
		}
		if !strings.ContainsRune(v, '!') {
			log.Fatalf("No ! in vault:\n%s", v)
		}
		if !strings.ContainsAny(v, "+-") {
			log.Fatalf("No + nor - in vault:\n%s", v)
		}
	}
	return vaults
}

// vault represents a special vault region.
type vault struct {
	p       gruid.Point // position
	w       int         // width of the vault
	h       int         // height of the vault
	entries []ventry    // available vault entries (for tunnels)
	places  []place     // special places in the vault (like for items)
	vault   *rl.Vault   // parsed vault string representation
	tunnels int         // number of tunnels
}

// ventry represents a vault entry.
type ventry struct {
	p    gruid.Point // position
	used bool        // whether the entry is actually used by a tunnel
}

// place represents a special place within a vault.
type place struct {
	p    gruid.Point // position
	kind placeKind   // kind of place
	used bool        // whether the place is already used
}

// placeKind represents the various kinds of places within a vault.
type placeKind int

const (
	PlaceEntry placeKind = iota
	PlaceItem
	PlaceStatic
	PlaceWaypoint
)

func vaultDistance(v1, v2 *vault) int {
	return paths.DistanceManhattan(v1.p.Add(gruid.Point{v1.w / 2, v1.h / 2}), v2.p.Add(gruid.Point{v2.w / 2, v2.h / 2}))
}

// DigVault actually draws the vault onto the map and gathers place information.
func (mg *MapGen) DigVault(v *vault) {
	runedraw := func(p gruid.Point, c rune) {
		q := gruid.Point{X: v.p.X + p.X, Y: v.p.Y + p.Y}
		if inMap(q) && c != '?' {
			mg.vault.SetU(q, true)
		}
		switch c {
		case '.', '!', '-', '>', 'W':
			if inMap(q) {
				mg.terrain.Set(q, Floor)
			}
		case '#', '+':
			if inMap(q) {
				mg.terrain.Set(q, Wall)
				// mg.terrain.Set(q, TranslucentWall)
			}
		case '$':
			if inMap(q) {
				mg.terrain.Set(q, TranslucentWall)
			}
		case '%':
			if inMap(q) {
				if mg.rand.IntN(2) == 0 {
					mg.terrain.Set(q, Wall)
				} else {
					mg.terrain.Set(q, TranslucentWall)
				}
			}
		case '&':
			if inMap(q) {
				switch mg.rand.IntN(5) {
				case 0:
					mg.terrain.Set(q, Wall)
				case 1:
					mg.terrain.Set(q, TranslucentWall)
				case 2:
					mg.terrain.Set(q, Foliage)
				case 3:
					mg.terrain.Set(q, Rubble)
				default:
					mg.terrain.Set(q, Floor)
				}
			}
		case '"':
			if inMap(q) {
				mg.terrain.Set(q, Foliage)
			}
		case '^':
			if inMap(q) {
				mg.terrain.Set(q, Rubble)
			}
		case ':':
			if inMap(q) {
				switch mg.rand.IntN(3) {
				case 0:
					mg.terrain.Set(q, Floor)
				case 1:
					mg.terrain.Set(q, Foliage)
				default:
					mg.terrain.Set(q, Rubble)
				}
			}
		case '?':
		default:
			log.Printf("mapgen: invalid terrain: %c for room w:%d h:%d pos:%+v\n%s", c, v.w, v.h, v.p, v.vault.Content())
		}
		switch c {
		case 'W':
			v.places = append(v.places, place{p: q, kind: PlaceWaypoint})
		case '!':
			v.places = append(v.places, place{p: q, kind: PlaceItem})
		case '>':
			v.places = append(v.places, place{p: q, kind: PlaceStatic})
		case '+', '-':
			if q.X == 0 || q.X == MapWidth-1 || q.Y == 0 || q.Y == MapHeight-1 {
				break
			}
			v.entries = append(v.entries, ventry{p: q})
		case '"':
			if inMap(q) {
				mg.terrain.Set(q, Foliage)
			}
		case '^':
			if inMap(q) {
				mg.terrain.Set(q, Rubble)
			}
		}
	}
	v.vault.Iter(runedraw)
}

// UnusedEntry returns an unused entry, if possible, or a random entry
// otherwise.
func (mg *MapGen) UnusedEntry(v *vault) int {
	ens := []int{}
	for i, e := range v.entries {
		if !e.used {
			ens = append(ens, i)
		}
	}
	if len(ens) == 0 {
		return mg.rand.IntN(len(v.entries))
	}
	return ens[mg.rand.IntN(len(ens))]
}

// RandomPlace returns a random free place of the given kind (among all vaults).
func (mg *MapGen) RandomPlace(kind placeKind) gruid.Point {
	_, v := mg.RandomVaultsPlace(mg.vaults, kind)
	return v
}

// RandomVaultsPlace returns a random free place of the given kind among given
// vaults.
func (mg *MapGen) RandomVaultsPlace(vs []*vault, kind placeKind) (int, gruid.Point) {
	return mg.RandomVaultsPlaceWithFunc(vs, kind, mg.rand.IntN)
}

// RandomVaultsPlaceWithFunc returns a random free place of the given kind
// among given vaults using the given custom IntN-like rand function.
func (mg *MapGen) RandomVaultsPlaceWithFunc(vs []*vault, kind placeKind, rf func(n int) int) (int, gruid.Point) {
	// We assume len(vs) > 0 and that there are remaining free places of
	// that kind.
	const maxIterations = 10000
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		i := rf(len(vs))
		p := mg.RandomVaultPlace(vs[i], kind)
		if p != InvalidPos {
			return i, p
		}
	}
}

// RandomVaultPlace returns a random free place of the given kind in the given
// vault, or returns an invalid position if there is no such place.
func (mg *MapGen) RandomVaultPlace(v *vault, kind placeKind) gruid.Point {
	var p []int
	for i, pl := range v.places {
		if pl.kind == kind && !pl.used {
			p = append(p, i)
		}
	}
	if len(p) == 0 {
		return InvalidPos
	}
	j := p[mg.rand.IntN(len(p))]
	v.places[j].used = true
	return v.places[j].p
}
