package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// vault represents a special vault region.
type vault struct {
	p       gruid.Point
	w       int
	h       int
	entries []ventry
	places  []place
	vault   *rl.Vault
	special bool
	tunnels int
}

// places represents the position of various special story places.
type places struct {
	Shaedra  gruid.Point
	Monolith gruid.Point
	Marevor  gruid.Point
	Artifact gruid.Point
}

// ventry represents a vault entry.
type ventry struct {
	p      gruid.Point // position
	used   bool        // whether the entry is actually used by a tunnel
	nodoor bool        // whether the entry doesn't want a door
}

// placeKind represents the various kinds of special places within a vault.
type placeKind int

const (
	PlaceDoor placeKind = iota
	PlacePatrol
	PlaceStatic
	PlaceSpecialStatic
	PlaceItem
	PlaceStory
	PlacePatrolSpecial
)

// place represents a special place within a vault.
type place struct {
	p    gruid.Point // position
	kind placeKind   // kind of place
	used bool        // whether the place is already used
}

// vaultDistance returns the distance between the center of two vaults.
func vaultDistance(r1, r2 *vault) int {
	return paths.DistanceManhattan(r1.p.Add(gruid.Point{r1.w / 2, r1.h / 2}), r2.p.Add(gruid.Point{r2.w / 2, r2.h / 2}))
}

// HasSpace reports whether there's enough space for the vault.
func (r *vault) HasSpace(dg *MapGen) bool {
	if MapWidth-r.p.X < r.w || MapHeight-r.p.Y < r.h {
		return false
	}
	for i := r.p.X - 1; i <= r.p.X+r.w; i++ {
		for j := r.p.Y - 1; j <= r.p.Y+r.h; j++ {
			p := gruid.Point{i, j}
			if InMap(p) && dg.vault[p] {
				return false
			}
		}
	}
	return true
}

// Dig draws the vault and gathers information.
func (r *vault) Dig(dg *MapGen) {
	runedraw := func(p gruid.Point, c rune) {
		q := gruid.Point{X: r.p.X + p.X, Y: r.p.Y + p.Y}
		if InMap(q) && c != '?' {
			dg.vault[q] = true
		}
		// Draw terrain.
		switch c {
		case '.', '>', '!', 'P', '_', '|', 'G', '-':
			if InMap(q) {
				dg.m.Set(q, GroundCell)
			}
		case 'B':
			// obstacle
			t := WallCell
			switch RandInt(9) {
			case 0, 6:
				t = TreeCell
			case 1:
				if RandInt(2) == 0 {
					t = QueenRockCell
				} else {
					t = LightCell
				}
			case 2:
				if RandInt(2) == 0 {
					t = ChasmCell
				} else {
					t = TableCell
				}
			case 3:
				t = TableCell
			case 4, 5:
				t = GroundCell
			}
			if InMap(q) {
				dg.m.Set(q, t)
			}
		case '#', '+':
			if InMap(q) {
				dg.m.Set(q, WallCell)
			}
		case ':':
			if InMap(q) {
				switch dg.rand.IntN(4) {
				case 0:
					dg.m.Set(q, FoliageCell)
				case 1:
					dg.m.Set(q, RubbleCell)
				default:
					dg.m.Set(q, GroundCell)
				}
			}
		case '"':
			if InMap(q) {
				dg.m.Set(q, FoliageCell)
			}
			if dg.rand.IntN(10) == 0 {
				dg.m.Set(q, GroundCell)
			}
		case ',':
			if InMap(q) {
				dg.m.Set(q, CavernCell)
			}
		case '~':
			if InMap(q) {
				dg.m.Set(q, WaterCell)
			}
		case 'c':
			if InMap(q) {
				dg.m.Set(q, ChasmCell)
			}
		case 'q':
			if InMap(q) {
				dg.m.Set(q, QueenRockCell)
			}
		case 'T':
			if InMap(q) {
				dg.m.Set(q, TreeCell)
			}
		case 'π':
			if InMap(q) {
				dg.m.Set(q, TableCell)
			}
			if dg.rand.IntN(7) == 0 {
				dg.m.Set(q, GroundCell)
			}
		case 'l':
			if InMap(q) {
				dg.m.Set(q, LightCell)
			}
			if dg.rand.IntN(7) == 0 {
				dg.m.Set(q, GroundCell)
			}
		case '%':
			if InMap(q) {
				switch dg.rand.IntN(10) {
				case 0:
					dg.m.Set(q, WindowCell)
				case 1:
					dg.m.Set(q, HoledWallCell)
				case 2:
					dg.m.Set(q, RubbleCell)
				default:
					dg.m.Set(q, WallCell)
				}
			}
		case 'W':
			if InMap(q) {
				switch dg.rand.IntN(10) {
				case 0:
					dg.m.Set(q, WallCell)
				case 1:
					dg.m.Set(q, HoledWallCell)
				case 2:
					dg.m.Set(q, RubbleCell)
				default:
					dg.m.Set(q, WindowCell)
				}
			}
		case '?':
		case 'S', 'M', 'Δ', 'A':
			dg.m.Set(q, StoryCell)
		default:
			log.Fatalf("Invalid terrain: %c for vault w:%d h:%d pos:%+v\n%s", c, r.w, r.h, r.p, r.vault.Content())
		}
		// Record special places.
		switch c {
		case '>':
			r.places = append(r.places, place{p: q, kind: PlaceSpecialStatic})
		case '!':
			r.places = append(r.places, place{p: q, kind: PlaceItem})
		case 'P':
			r.places = append(r.places, place{p: q, kind: PlacePatrol})
		case 'G':
			r.places = append(r.places, place{p: q, kind: PlacePatrolSpecial})
		case '_':
			r.places = append(r.places, place{p: q, kind: PlaceStatic})
		case '|':
			r.places = append(r.places, place{p: q, kind: PlaceDoor})
		case '+', '-':
			if q.X == 0 || q.X == MapWidth-1 || q.Y == 0 || q.Y == MapHeight-1 {
				break
			}
			e := ventry{}
			e.p = q
			if c == '-' {
				e.nodoor = true
			}
			r.entries = append(r.entries, e)
		case 'S':
			r.places = append(r.places, place{p: q, kind: PlaceStory})
			dg.spl.Shaedra = q
		case 'M':
			r.places = append(r.places, place{p: q, kind: PlaceStory})
			dg.spl.Marevor = q
		case 'Δ':
			r.places = append(r.places, place{p: q, kind: PlaceStory})
			dg.spl.Monolith = q
		case 'A':
			r.places = append(r.places, place{p: q, kind: PlaceStory})
			dg.spl.Artifact = q
		}
	}
	r.vault.Iter(runedraw)
}

// UnusedEntry returns an unused entry, if possible, or a random entry
// otherwise.
func (r *vault) UnusedEntry() int {
	ens := []int{}
	for i, e := range r.entries {
		if !e.used {
			ens = append(ens, i)
		}
	}
	if len(ens) == 0 {
		return RandInt(len(r.entries))
	}
	return ens[RandInt(len(ens))]
}

// RandomPlace returns a random place of the given kind in the vault.
func (r *vault) RandomPlace(kind placeKind) gruid.Point {
	var p []int
	for i, pl := range r.places {
		if pl.kind == kind && !pl.used {
			p = append(p, i)
		}
	}
	if len(p) == 0 {
		return invalidPos
	}
	j := p[RandInt(len(p))]
	r.places[j].used = true
	return r.places[j].p
}

var PlaceSpecialOrStatic = []placeKind{PlaceSpecialStatic, PlaceStatic}

// RandomPlaces returns a random place of any of the given kinds.
func (r *vault) RandomPlaces(kinds []placeKind) gruid.Point {
	p := invalidPos
	for _, kind := range kinds {
		p = r.RandomPlace(kind)
		if p != invalidPos {
			break
		}
	}
	return p
}

const (
	VaultAlmostSquare = `
?###+##??
#_....!#?
+..PBP.!#
#!....._#
?###+###?`
	VaultSquareBis = `
?##?#+##?
#_!#!.._#
#..|.P.P+
#_!#!.._#
?##?#+##?`
	VaultRoundSimple = `
??#+#??
?#!.!#?
#_.P._#
+.PBP.+
#_.P._#
?#!.!#?
??#+#??`
	VaultLittle = `
?#+#?
#_._#
+.P.+
#_.!#
?#+#?`
	VaultLittleDiamond = `
??#+#??
##!._##
+..P..+
##_._##
??#+#??`
	VaultLittleColumnDiamond = `
?##+#??
#_..!##
+.PBP.+
##!.._#
??#+##?`
	VaultLittleTreeDiamond = `
??#+##?
##!.B_#
+.P.P.#
#_B.!##
?##+#??`
	VaultRound = `
???##+##???
??#".P."#??
##".!B_."##
+.P.B#B.P.+
##"._B_."##
??#".P."#??
???##+##???`
	VaultRoundTree = `
???##+##???
??#".P."#??
##"._._."##
+.P..B..P.+
##"._.!."##
??#".P."#??
???##+##???`
)

var vaultNormalTemplates = []string{VaultAlmostSquare, VaultSquareBis, VaultRoundSimple, VaultLittle, VaultLittleDiamond, VaultLittleColumnDiamond, VaultRound, VaultLittleTreeDiamond, VaultRoundTree}

const (
	VaultBigColumns = `
?####?#++#?####?
#!.._##..##!..>#
##.P........P.##
+...._##%#._...+
##.P........P.##
#!..>##..##>..!#
?####?#++#?####?`
	VaultBigGarden = `
?####?#++#?####?
#""""##..##...!#
#"T""".!P...~~.#
#""""">P_>.~~~.+
#"T""".!P...~~"#
#""""##..##..""#
?####?#++#?####?`
	VaultBigVaults = `
?####?#++#?###??
#_..!##..##!..#?
#"""..#..#..π._#
#"""P.|.P|.P..>#
#"""..#..#..π.!#
#>..!##..##...#?
?####?#++#?###??`
	VaultColumns = `
###+##+###
+..P..P..+
#.#>B#!#.#
#.#!##>B.#
+..P..P..+
###+##+###`
	VaultTriangles = `
?####?######?
#...>#_.....#
+.P.#...#.P.+
#..#_...!#..#
#!....P...#>#
?#..!#|#..>#?
??###.-.###??`
	VaultHome1 = `
?########+#?
#>.P..|..P!#
#...B.#....#
#!...!#_...#
#%#|#%####|#
#...P...|..#
#>.....!#P.#
?########+#?`
	VaultHome2 = `
##########??
+...#....>#?
#.P.|..P...#
##|###..._.#
#...>%!...!#
#....##|####
#!P..|..P..+
?########+##`
	VaultHome3 = `
?#############?
#>P.|.........#
#..!##|##!.P..+
##%#!...!#_...#
+..|.P>._###|##
##%#!...!#!...#
#!.>##|##..P..+
#.P.....|...B.#
?#......#....#?
??###########??`
	VaultHome4 = `
?############?
#..._.#.....!#
#.P...|..B...#
##|####....P.+
#...._#>.##|##
+..P..####>..#
#!....|....P!#
?###+########?`
	VaultHome5 = `
????##########?
??###!...>#..!#
?###>.....!#..#
#..%...P.B.%.P+
+P.#_......#..+
#..####|####..#
#!.....P....._#
?######+######?`
	VaultHome6 = `
?###+#???
#>.#q##??
#.P|P_##?
#.!#.#>.#
#_##.##|#
##!.....#
+.PBPB..#
#......_#
?###+###?`
	VaultCaban = `
???????-??????
?????""""?????
???"":"""""???
??"":###."""??
?"":#>!|PP"""?
-""":###.""""-
??""":""""""??
????"""""?????
???????-??????`
	VaultDolmen = `
???????-?????
????.....????
????.!#!.????
???..#>#..???
??....P....??
?...B_._B...?
-.P...P...P.-
??...#>#...??
????.!#!.????
????.....????
???????-?????`
	VaultSmallTemple = `
?????????????
?????...?????
????..#..????
???.:#>#..???
??..#>._#..??
??..%...%..??
?...#!P!#...?
-.P..#|#..P.-
?..#.....#..?
??..#.P_#..??
???.......???
??????-??????`
	VaultTemple = `
????.....????
??...###...??
?..##!>!##..?
..#_....._#..
.#....B....#.
.%.##.P.##.%.
.#.>#...#>.#.
-#..P...P..#.
..###|||###..
?.....P.....-
??."T...T".??
???"".-.""???`
	VaultSchool = `
????#######???
???#l..!..l#??
??#!..Pπ..!##?
?#.........|.#
.W.π.π..π..#>#
.#.........l##
.W.π.π..π.##>#
.#........#..#
-W.π.π..π.#P.#
.#l...P...|._#
..##%#||#####-
?.....--P.....`
	VaultShop = `
??????###???
?????#>P>#??
??#####|##??
?#l...P..l#?
.W.π.π!.π.#?
.#.π.π..π.#?
-W.π.π..π.#?
.#...π.P..l#
.#l!..qq_!#?
..##%#||##.-
?.....--P..?`
	VaultTavern = `
??##########??
?#l.π...|..P#?
.#..!...###>#?
.W.π.π...!###?
.#....P.π.#>.#
.W.π.π..π.#P.#
.#l...qqπ.|._#
-.####||###|#-
?.....--P.....`
	VaultDoctor = `
????#####??
???#l...!#?
??##.Pππ.#?
?#.|.....#?
.W>#_...l#?
.####|###??
.W...P#>.#?
.#.π..##|#.
.W.!.....W-
.#l..Pq._#.
-.##%#|##..
?.....-.P.?`
	VaultRuins = `
????????-????
????....P????
???..###.????
???.##>#..???
-.....P....??
?..##_"!##..?
??.#""""#.P.-
??.""B###..??
????"#>#.????
????..P..????
??????-??????`
	VaultPillars = `
???????-?????
???:......???
?...#!#P#...?
?.#!......#.?
-.P.B>B>B.P.-
?.#......!#.?
?...#P#!#...?
???.......???
?????-???????`
	VaultRoundColumns = `
???##+##???
??#!...!#??
##_.BP#._##
+...P>P...+
##_.#P#._##
??#!...!#??
???##+##???`
	VaultTriangle = `
?????#?????
????#>#????
???#!.!#???
??#_..._#??
##!.#.B.!##
+..P...P..+
##_..P.._##
??###+###??`
	VaultSpiraling = `
?#####+#
#!.P...+
#.>##%##
#!.P..!#
##%##>.#
+...P.!#
#+#####?`
	VaultSpiralingCircle = `
??##+#?##??
##.P.!#:>#?
+.P.##....#
#._#...#_.#
#....##!P.+
?#.>#.P.:##
??###+###??`
	VaultCircleDouble = `
???####+#???
??#""..P#???
?#":..#|#???
#"""!#!P!#??
#.._#.....#?
#..#..>#>..#
+.P|P.###.P#
#..#.._#>..#
#"._#.....##
#"":!#!P!#!#
?#""..#|#..#
??#"...P..#?
???####+##??`
	VaultGardenHome = `
???#######???
??#.#>#>#.#??
?#!...P...!#?
#...#._.#...#
#_........._#
##%###|###%##
+.....P..#_>#
+.......#!..#
#######|#...#
#":""..P|.P.#
?#"T"...#...+
??#""".!#!..#
???#######+#?`
	VaultAltar = `
#+#??#######??#+#
+P_##>..!..>##_P+
#...#..#>#..#...#
?##.#!:.P..!#.##?
??#..P.....P..#??
???#####+#####???`
	VaultRoundGarden = `
???##+##???
??#_.P.>#??
##!.""".!##
+.P""B""P.+
##!.""".!##
??#>.P.>#??
???##+##???`
	VaultLongHall = `
#################
+.P...........P.+
#..!#!>.B.>!#!..#
+.P...........P.+
#################`
	VaultGardenHall = `
?#############?
#"""".>!>.""""#
+..P...P...P..+
#""""._!>.""""#
?#############?`
	VaultRoundHall = `
????###?###+#?
?###!_##>#.P.#
#....###|#...#
+.P_..P....P.+
#....#!#.....#
?###.>#?#.π.##
????##???###??`
	VaultToilets = `
??######
?#!.P..+
-Wπ..###
.#!..|>#
.WπP.##?
.#!..|>#
-|.P.###
?####???`
	VaultPicnic = `
????""""????
??""""T"""??
?"T.!P..T""?
?"".πlπ!.."?
-..P.>.P...-
?""!π.π..."?
?"T..P!.T""?
??""".."""??
??????-?????`
	VaultSnake = `
?????#####???###
?#?##!..._#?##.+
#B#...:....#.|.#
#...###%##.P.###
#..##....###..#?
-P.#...#.!#>#.#?
-..#..##>P..#.#?
#π#":..!##%#..#?
?##"""T.P:...#??
???#"""....!#???
????########????`
)

var vaultBigTemplates = []string{VaultBigColumns, VaultBigGarden, VaultColumns,
	VaultRoundColumns, VaultRoundGarden, VaultLongHall, VaultGardenHall,
	VaultTriangles, VaultHome1, VaultHome2, VaultHome3, VaultHome4, VaultHome5, VaultHome6,
	VaultTriangle, VaultSpiraling, VaultSpiralingCircle, VaultAltar,
	VaultCircleDouble, VaultGardenHome, VaultBigVaults, VaultCaban, VaultDolmen,
	VaultSmallTemple, VaultTemple, VaultSchool, VaultTavern, VaultShop,
	VaultDoctor, VaultRuins, VaultPillars, VaultRoundHall, VaultToilets, VaultPicnic, VaultSnake}

const (
	CellShaedra = `
?#?#?#?#?
#########
#SMΔ#_!_#
##|###|##
+.G.P.G.+
##|###|##
#_!_#_!_#
#########
?#?#?#?#?
`
	CellShaedra2 = `
ccccccc-?
c#####c..
c#SMΔ#ccP
c##|####.
cWG..G.|-
c##|##.##
c#_!_#!π#
c#####|#?
cccccc-P?
`
	CellShaedra3 = `
ccccccc-?
c#####c..
c#_!_#ccP
c##|####.
cWG..G.|-
c##|##.##
c#SMΔ#!π#
c#####|#?
cccccc-P?
`
	CellShaedra4 = `
?????c?????
????ccc????
??ccccccc??
?cc##W##cc?
cc##!G!##cc
c#_#...#S#c
c#!|...|M#c
c#_#.G.#Δ#c
c###_._###c
-.###|###.-
?.P..-..P.?
`
)

var vaultCellTemplates = []string{CellShaedra, CellShaedra2, CellShaedra3, CellShaedra4}

const (
	VaultArtifact = `
????#????
???#c#???
??#cAc#??
?##.MΔ##?
####|####
#!_#.#_!#
+P.G.G.P+
#>..P..>#
?###+###?`
	VaultArtifact2 = `
??cccc??
?cc#####
cc#cc!P+
c##M#G.#
c#ΔA|..#
c##.#G.#
c###π.!#
-|q..P>#
P######?`
	VaultArtifact3 = `
?#####+#?
#B..Gcq.#
#.Gc.cP.#
#.ccMccq#
#.c.AΔcP#
#.>ccc!.#
#...P...#
#!.#q_.B#
?###+###?`
)

var vaultArtifactTemplates = []string{VaultArtifact, VaultArtifact2, VaultArtifact3}

const (
	VaultSpecialNixes = `
?#########???
##.#~>~##>##?
#>.G~.~G...!#
##.......B..#
#!.#qqq#..._#
##.G...G.#..#
#...._....P.+
+P.......#!##
#....P...!#??
?####+####???`
	VaultSpecialVampires = `
?#####+#####?
#>.G.#.#...>#
?#.......G.#?
??#!_.P._.#!#
???##....#>G#
??#~_#P._!#|#
?#...|......#
#>.G.#.#..._#
?#####+#####?`
	VaultSpecialVampires2 = `
?~~~~~~~~~~-.
~~##W##W##~~.
~#>._#.π.##~P
~W.G.|...|.~-
~|q.!#.G.|.G-
~####>...|.~-
~#>G|....##~P
~~###.π.!%~~.
?~~~##W###~~.
???~~~~~~~~-.`
	VaultSpecialFrogs = `
?~~-~~?~??
~~.P~~~~~?
~..._~B.~?
?~.G.>~.~~
~~..!.G.B~
-P..B>~.~~
~~.G>~~.~?
?~!..BG.~~
?~~~P..~~?
???~-~~???`
	VaultSpecialMilfids = `
?????????-???
???......P.??
??.:!?G._?..?
?.?.?#>c.?.??
-P.G.>#>G...?
?.?._c>c.??.?
??.......?..-
????!.G...P??
????????-????`
	VaultSpecialMilfids2 = `
?????????-??
-P.ccc?c.Gc?
?c.G..cc..cc
??ccc._G..>c
?ccc>....!cc
??c!...:>ccc
??cc.G.cccc?
???c.....???
?????P-.????`
	VaultSpecialTreeMushrooms = `
????"--.???????
??"""..G."""???
??""?....!:""??
??....T>T..G..?
?.G.T..!..T..P-
-..!..T>T._"".-
-P???..G.."":"?
.-??????.-?????`
	VaultSpecialTreeMushrooms2 = `
???""-???????
??""?.."""???
??"":.G..""??
?""?....!..??
?".T.T.T.T..?
?.G..>!>.G.P-
-.!π.T.T....-
-.?......_""?
?P??".TG.""??
.-?????.-????`
	VaultSpecialHarpies = `
?-????##??????
?P???#..#?????
?.???#G.>####?
?.??#...##.>#?
?.?#.._:...G.#
-G........!..#
??.#.G._._#>#?
??P?#.>###?#??
??-??##???????`
	VaultSpecialHarpies2 = `
?-#????##..P
-P.####...#-
?#.G.#..G.!#
?##..:_....#
?#.G.....!##
#>.#_.G.##??
?##?###.>#??
???????##???`
	VaultSpecialCelmists = `
?############+##?
#>#_......>#.P._#
#...G!#!G.##....#
#....###...|...P+
#_...###...|...P+
##..G!#!G.##....#
#>........>%.P._#
?############+##?`
	VaultSpecialCelmists2 = `
?##?##########
#_.#....G....+
#..#>#W#|#W#!#
#..###>.._.###
+P.|.G.....G>#
#..###._.!.###
#..#!#W#|#W#>#
#_#.....G....#
?##+#########?`
	VaultSpecialCelmists3 = `
?###???.-.??????
#...#..P...?????
+.P.|....P.....-
##|##W#|||#W#%##
?#....G...G....#
??#!._#._.#_.!#?
???#..G...G..#??
????#!.....!#???
?????#>#>#>#????
??????#?#?#?????`
	VaultSpecialMirrorSpecters = `
########-########
-P.....W.W..q..P-
##W##_.#q#._##W##
-P......G......P-
##W##..##W##.#W##
#>.!W..W>!>W.W!>#
#.G.#..#.G.#.#G.#
#...............#
#################`
	VaultSpecialMirrorSpecters2 = `
######--#####
-P...W.P...P-
#.#####B.##.#
#.W>.G...!W.#
#.###..#.##G#
#.....#..q..#
#W#.B#>.!##W#
#>#!.....#.>#
#..#..B.#..#?
?#G..#!.G.#??
??###?####???`
)

type specialVault int

const (
	noSpecialVault specialVault = iota
	vaultMilfids
	vaultFrogs
	vaultNixes
	vaultVampires
	vaultCelmists
	vaultHarpies
	vaultTreeMushrooms
	vaultMirrorSpecters
	vaultShaedra
	vaultArtifact
)

// Templates returns the template for a given special vault.
func (sr specialVault) Templates() (tpl []string) {
	switch sr {
	case vaultMilfids:
		tpl = append(tpl, VaultSpecialMilfids, VaultSpecialMilfids2)
	case vaultFrogs:
		tpl = append(tpl, VaultSpecialFrogs)
	case vaultVampires:
		tpl = append(tpl, VaultSpecialVampires, VaultSpecialVampires2)
	case vaultCelmists:
		tpl = append(tpl, VaultSpecialCelmists, VaultSpecialCelmists2, VaultSpecialCelmists3)
	case vaultNixes:
		tpl = append(tpl, VaultSpecialNixes)
	case vaultHarpies:
		tpl = append(tpl, VaultSpecialHarpies, VaultSpecialHarpies2)
	case vaultTreeMushrooms:
		tpl = append(tpl, VaultSpecialTreeMushrooms, VaultSpecialTreeMushrooms2)
	case vaultMirrorSpecters:
		tpl = append(tpl, VaultSpecialMirrorSpecters, VaultSpecialMirrorSpecters2)
	case vaultShaedra:
		tpl = vaultCellTemplates
	case vaultArtifact:
		tpl = vaultArtifactTemplates
	}
	return tpl
}
