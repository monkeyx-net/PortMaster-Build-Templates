package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
)

// bandInfo represents monster band information for generation.
type bandInfo struct {
	Path []gruid.Point
	I    int
	Kind monsterBand
	Beh  mbehaviour
}

type monsterBand int

const (
	LoneGuard monsterBand = iota
	LoneHighGuard
	LoneYack
	LoneOricCelmist
	LoneHarmonicCelmist
	LoneSatowalgaPlant
	LoneBlinkingFrog
	LoneWorm
	LoneMirrorSpecter
	LoneDog
	LoneExplosiveNadre
	LoneWingedMilfid
	LoneMadNixe
	LoneTreeMushroom
	LoneEarthDragon
	LoneButterfly
	LoneVampire
	LoneHarpy
	LoneHazeCat
	LoneAcidMound
	LoneSpider
	PairGuard
	PairYack
	PairFrog
	PairDog
	PairTreeMushroom
	PairSpider
	PairHazeCat
	PairSatowalga
	PairWorm
	PairOricCelmist
	PairHarmonicCelmist
	PairVampire
	PairNixe
	PairExplosiveNadre
	PairWingedMilfid
	SpecialLoneVampire
	SpecialLoneNixe
	SpecialLoneMilfid
	SpecialLoneOricCelmist
	SpecialLoneHarmonicCelmist
	SpecialLoneHighGuard
	SpecialLoneHarpy
	SpecialLoneTreeMushroom
	SpecialLoneMirrorSpecter
	SpecialLoneAcidMound
	SpecialLoneHazeCat
	SpecialLoneSpider
	SpecialLoneBlinkingFrog
	SpecialLoneExplosiveNadre
	SpecialLoneYack
	SpecialLoneDog
	UniqueCrazyImp
)

type mbehaviour int

const (
	BehPatrol mbehaviour = iota
	BehGuard
	BehWander
	BehExplore
	BehCrazyImp
)

// monsterBandData holds generation info for a given kind of monster band.
type monsterBandData struct {
	Distribution map[monsterKind]int // distribution of several monster kinds
	Band         bool                // whether there are several monsters
	Monster      monsterKind         // single kind of monster band
}

// monsBands lists all the possible monster bands.
var monsBands = []monsterBandData{
	LoneGuard:                  {Monster: MonsGuard},
	LoneHighGuard:              {Monster: MonsHighGuard},
	LoneYack:                   {Monster: MonsYack},
	LoneOricCelmist:            {Monster: MonsOricCelmist},
	LoneHarmonicCelmist:        {Monster: MonsHarmonicCelmist},
	LoneSatowalgaPlant:         {Monster: MonsSatowalgaPlant},
	LoneBlinkingFrog:           {Monster: MonsBlinkingFrog},
	LoneWorm:                   {Monster: MonsWorm},
	LoneMirrorSpecter:          {Monster: MonsMirrorSpecter},
	LoneDog:                    {Monster: MonsDog},
	LoneExplosiveNadre:         {Monster: MonsExplosiveNadre},
	LoneWingedMilfid:           {Monster: MonsWingedMilfid},
	LoneMadNixe:                {Monster: MonsMadNixe},
	LoneTreeMushroom:           {Monster: MonsTreeMushroom},
	LoneEarthDragon:            {Monster: MonsEarthDragon},
	LoneButterfly:              {Monster: MonsButterfly},
	LoneVampire:                {Monster: MonsVampire},
	LoneHarpy:                  {Monster: MonsTinyHarpy},
	LoneHazeCat:                {Monster: MonsHazeCat},
	LoneAcidMound:              {Monster: MonsAcidMound},
	LoneSpider:                 {Monster: MonsSpider},
	PairGuard:                  {Band: true, Distribution: map[monsterKind]int{MonsGuard: 2}},
	PairYack:                   {Band: true, Distribution: map[monsterKind]int{MonsYack: 2}},
	PairFrog:                   {Band: true, Distribution: map[monsterKind]int{MonsBlinkingFrog: 2}},
	PairDog:                    {Band: true, Distribution: map[monsterKind]int{MonsDog: 2}},
	PairTreeMushroom:           {Band: true, Distribution: map[monsterKind]int{MonsTreeMushroom: 2}},
	PairSpider:                 {Band: true, Distribution: map[monsterKind]int{MonsSpider: 2}},
	PairHazeCat:                {Band: true, Distribution: map[monsterKind]int{MonsHazeCat: 2}},
	PairSatowalga:              {Band: true, Distribution: map[monsterKind]int{MonsSatowalgaPlant: 2}},
	PairWorm:                   {Band: true, Distribution: map[monsterKind]int{MonsWorm: 2}},
	PairVampire:                {Band: true, Distribution: map[monsterKind]int{MonsVampire: 2}},
	PairOricCelmist:            {Band: true, Distribution: map[monsterKind]int{MonsOricCelmist: 2}},
	PairHarmonicCelmist:        {Band: true, Distribution: map[monsterKind]int{MonsHarmonicCelmist: 2}},
	PairNixe:                   {Band: true, Distribution: map[monsterKind]int{MonsMadNixe: 2}},
	PairExplosiveNadre:         {Band: true, Distribution: map[monsterKind]int{MonsExplosiveNadre: 2}},
	PairWingedMilfid:           {Band: true, Distribution: map[monsterKind]int{MonsWingedMilfid: 2}},
	SpecialLoneVampire:         {Monster: MonsVampire},
	SpecialLoneNixe:            {Monster: MonsMadNixe},
	SpecialLoneMilfid:          {Monster: MonsWingedMilfid},
	SpecialLoneOricCelmist:     {Monster: MonsOricCelmist},
	SpecialLoneHarmonicCelmist: {Monster: MonsHarmonicCelmist},
	SpecialLoneHighGuard:       {Monster: MonsHighGuard},
	SpecialLoneHarpy:           {Monster: MonsTinyHarpy},
	SpecialLoneTreeMushroom:    {Monster: MonsTreeMushroom},
	SpecialLoneMirrorSpecter:   {Monster: MonsMirrorSpecter},
	SpecialLoneAcidMound:       {Monster: MonsAcidMound},
	SpecialLoneHazeCat:         {Monster: MonsHazeCat},
	SpecialLoneSpider:          {Monster: MonsSpider},
	SpecialLoneBlinkingFrog:    {Monster: MonsBlinkingFrog},
	SpecialLoneDog:             {Monster: MonsDog},
	SpecialLoneExplosiveNadre:  {Monster: MonsExplosiveNadre},
	SpecialLoneYack:            {Monster: MonsYack},
	UniqueCrazyImp:             {Monster: MonsCrazyImp},
}

// GenMonsters generates and places monsters.
func (mg *MapGen) GenMonsters(g *Game) {
	g.Monsters = []*Monster{}
	g.MonsterBands = []bandInfo{}
	// common bands
	bandsGuard := []monsterBand{LoneGuard}
	bandsButterfly := []monsterBand{LoneButterfly}
	bandsHighGuard := []monsterBand{LoneHighGuard}
	bandsAnimals := []monsterBand{LoneYack, LoneWorm, LoneDog, LoneBlinkingFrog, LoneExplosiveNadre, LoneHarpy, LoneAcidMound}
	bandsPlants := []monsterBand{LoneSatowalgaPlant}
	bandsBipeds := []monsterBand{LoneOricCelmist, LoneMirrorSpecter, LoneWingedMilfid, LoneMadNixe, LoneVampire, LoneHarmonicCelmist}
	bandsRare := []monsterBand{LoneTreeMushroom, LoneEarthDragon, LoneHazeCat, LoneSpider}
	// monster specific bands
	bandNadre := []monsterBand{LoneExplosiveNadre}
	bandFrog := []monsterBand{LoneBlinkingFrog}
	bandDog := []monsterBand{LoneDog}
	bandYack := []monsterBand{LoneYack}
	bandVampire := []monsterBand{LoneVampire}
	bandOricCelmist := []monsterBand{LoneOricCelmist}
	bandHarmonicCelmist := []monsterBand{LoneHarmonicCelmist}
	bandMadNixe := []monsterBand{LoneMadNixe}
	bandMirrorSpecter := []monsterBand{LoneMirrorSpecter}
	bandTreeMushroom := []monsterBand{LoneTreeMushroom}
	bandHazeCat := []monsterBand{LoneHazeCat}
	bandSpider := []monsterBand{LoneSpider}
	bandDragon := []monsterBand{LoneEarthDragon}
	bandGuardPair := []monsterBand{PairGuard}
	bandYackPair := []monsterBand{PairYack}
	bandExplosiveNadrePair := []monsterBand{PairExplosiveNadre}
	bandWingedMilfidPair := []monsterBand{PairWingedMilfid}
	bandNixePair := []monsterBand{PairNixe}
	bandVampirePair := []monsterBand{PairVampire}
	bandOricCelmistPair := []monsterBand{PairOricCelmist}
	bandHarmonicCelmistPair := []monsterBand{PairHarmonicCelmist}
	// special bands
	if g.Params.Special[g.Map.Depth] != noSpecialVault {
		switch mg.special {
		case vaultVampires:
			mg.putRandomBandN(g, []monsterBand{SpecialLoneVampire}, 2)
		case vaultNixes:
			mg.putRandomBandN(g, []monsterBand{SpecialLoneNixe}, 2)
		case vaultFrogs:
			mg.putRandomBandN(g, []monsterBand{SpecialLoneBlinkingFrog}, 2)
		case vaultMilfids:
			switch mg.rand.IntN(6) {
			case 0:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneYack}, 2)
			case 1:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneDog}, 2)
			default:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneMilfid}, 2)
			}
		case vaultCelmists:
			switch mg.rand.IntN(3) {
			case 0:
				bandOricCelmists := []monsterBand{SpecialLoneOricCelmist}
				mg.putRandomBandN(g, bandOricCelmists, 2)
			case 1:
				bandHarmonicCelmists := []monsterBand{SpecialLoneHarmonicCelmist}
				mg.putRandomBandN(g, bandHarmonicCelmists, 2)
			case 2:
				bandOricCelmists := []monsterBand{SpecialLoneOricCelmist}
				bandHarmonicCelmists := []monsterBand{SpecialLoneHarmonicCelmist}
				mg.putRandomBandN(g, bandHarmonicCelmists, 1)
				mg.putRandomBandN(g, bandOricCelmists, 1)
			}
		case vaultHarpies:
			if mg.rand.IntN(3) > 0 {
				mg.putRandomBandN(g, []monsterBand{SpecialLoneHarpy}, 2)
			} else {
				mg.putRandomBandN(g, []monsterBand{SpecialLoneSpider}, 2)
			}
		case vaultTreeMushrooms:
			if mg.rand.IntN(3) > 0 {
				mg.putRandomBandN(g, []monsterBand{SpecialLoneTreeMushroom}, 2)
			} else {
				mg.putRandomBandN(g, []monsterBand{SpecialLoneHazeCat}, 2)
			}
		case vaultMirrorSpecters:
			switch mg.rand.IntN(6) {
			case 0:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneAcidMound}, 2)
			case 1:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneExplosiveNadre}, 2)
			default:
				mg.putRandomBandN(g, []monsterBand{SpecialLoneMirrorSpecter}, 2)
			}
		case vaultShaedra:
			if mg.rand.IntN(3) > 0 {
				mg.putRandomBand(g, []monsterBand{SpecialLoneHighGuard})
			} else {
				mg.putRandomBand(g, []monsterBand{SpecialLoneOricCelmist})
			}
		case vaultArtifact:
			switch mg.rand.IntN(3) {
			case 0:
				mg.putRandomBand(g, []monsterBand{SpecialLoneHarmonicCelmist})
			case 1:
				mg.putRandomBand(g, []monsterBand{SpecialLoneOricCelmist})
			default:
				mg.putRandomBand(g, []monsterBand{SpecialLoneHighGuard})
			}
		default:
			// XXX not used now
			bandOricCelmists := []monsterBand{SpecialLoneOricCelmist}
			mg.putRandomBandN(g, bandOricCelmists, 2)
		}
	}
	if g.Map.Depth == g.Params.CrazyImp {
		mg.putRandomBand(g, []monsterBand{UniqueCrazyImp})
	}
	mg.putRandomBandN(g, bandsButterfly, 2)
	if mg.layout == RandomSmallWalkCaveUrbanised {
		mg.putRandomBandN(g, bandsGuard, 1+(g.Map.Depth+1)/4)
	}
	switch g.Map.Depth {
	case 1:
		// 8-9
		if mg.rand.IntN(2) == 0 {
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandsGuard, 5)
			} else {
				mg.putRandomBandN(g, bandsGuard, 4)
				mg.putRandomBandN(g, bandsBipeds, 1)
			}
			mg.putRandomBandN(g, bandsAnimals, 3)
		} else {
			mg.putRandomBandN(g, bandsGuard, 4)
			if mg.rand.IntN(5) > 0 {
				mg.putRandomBandN(g, bandsAnimals, 5)
			} else {
				mg.putRandomBandN(g, bandsAnimals, 3)
				mg.putRandomBandN(g, bandsRare, 1)
			}
		}
	case 2:
		// 10-11
		mg.putRandomBandN(g, bandsGuard, 3)
		switch mg.rand.IntN(5) {
		case 0, 1:
			// 7
			mg.putRandomBandN(g, bandsBipeds, 1)
			mg.putRandomBandN(g, bandsAnimals, 4)
			mg.putRandomBandN(g, bandsPlants, 2)
		case 2, 3:
			// 8
			mg.putRandomBandN(g, bandsAnimals, 3)
			mg.putRandomBandN(g, bandsRare, 1)
			mg.putRandomBandN(g, bandsButterfly, 2)
			mg.putRandomBandN(g, bandsPlants, 2)
		case 4:
			// 8
			mg.putRandomBandN(g, bandsPlants, 3)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandFrog, 5)
			} else {
				mg.putRandomBandN(g, bandYack, 5)
			}
		}
	case 3:
		// 12-13
		mg.putRandomBandN(g, bandsHighGuard, 2)
		mg.putRandomBandN(g, bandsGuard, 4)
		switch mg.rand.IntN(5) {
		case 0, 1:
			// 6
			if mg.rand.IntN(3) == 0 {
				mg.putRandomBandN(g, bandDog, 4)
			} else {
				mg.putRandomBandN(g, bandsAnimals, 4)
			}
			mg.putRandomBandN(g, bandsPlants, 2)
		case 2, 3:
			// 6
			mg.putRandomBandN(g, bandsAnimals, 2)
			mg.putRandomBandN(g, bandsPlants, 1)
			mg.putRandomBandN(g, bandsBipeds, 3)
		case 4:
			// 7
			mg.putRandomBandN(g, bandsPlants, 1)
			mg.putRandomBandN(g, bandNadre, 6)
		}
	case 4:
		// 13-14
		mg.putRandomBandN(g, bandsHighGuard, 2)
		switch mg.rand.IntN(5) {
		case 0, 1:
			// 10
			mg.putRandomBandN(g, bandsGuard, 5)
			mg.putRandomBandN(g, bandsRare, 2)
			mg.putRandomBandN(g, bandGuardPair, 1)
			mg.putRandomBandN(g, bandsBipeds, 1)
			mg.putRandomBandN(g, bandsPlants, 1)
		case 2, 3:
			// 11
			mg.putRandomBandN(g, bandsGuard, 8)
			mg.putRandomBandN(g, bandGuardPair, 1)
			mg.putRandomBandN(g, bandsAnimals, 1)
			mg.putRandomBandN(g, bandsPlants, 1)
		case 4:
			mg.putRandomBandN(g, bandsGuard, 5)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandOricCelmist, 4)
			} else {
				mg.putRandomBandN(g, bandHarmonicCelmist, 4)
			}
			mg.putRandomBandN(g, bandsPlants, 1)
		}
	case 5:
		// 14-15
		mg.putRandomBandN(g, bandsHighGuard, 2)
		if mg.rand.IntN(2) == 0 {
			// 12
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandsGuard, 2)
				mg.putRandomBandN(g, bandGuardPair, 1)
			} else {
				mg.putRandomBandN(g, bandsGuard, 4)
			}
			mg.putRandomBandN(g, bandsAnimals, 1)
			mg.putRandomBandN(g, bandsRare, 3)
			mg.putRandomBandN(g, bandsBipeds, 3)
			mg.putRandomBandN(g, bandsPlants, 1)
		} else {
			// 13
			mg.putRandomBandN(g, bandsGuard, 2)
			mg.putRandomBandN(g, bandsAnimals, 4)
			mg.putRandomBandN(g, bandsBipeds, 2)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandOricCelmistPair, 1)
			} else {
				mg.putRandomBandN(g, bandHarmonicCelmistPair, 1)
			}
			mg.putRandomBandN(g, bandsRare, 1)
			mg.putRandomBandN(g, bandsPlants, 2)
		}
	case 6:
		// 16-17
		mg.putRandomBandN(g, bandsHighGuard, 1)
		if mg.rand.IntN(2) == 0 {
			// 15
			mg.putRandomBandN(g, bandsGuard, 3)
			mg.putRandomBandN(g, bandsAnimals, 3)
			mg.putRandomBandN(g, bandsRare, 3)
			mg.putRandomBandN(g, bandsBipeds, 1)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandYackPair, 1)
			} else {
				mg.putRandomBandN(g, bandWingedMilfidPair, 1)
			}
			mg.putRandomBandN(g, bandsPlants, 3)
		} else {
			// 16
			mg.putRandomBandN(g, bandsGuard, 2)
			if mg.rand.IntN(2) == 0 {
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, bandYack, 8)
				} else {
					mg.putRandomBandN(g, bandFrog, 8)
				}
			} else {
				mg.putRandomBandN(g, bandsRare, 2)
				if mg.rand.IntN(3) == 0 {
					mg.putRandomBandN(g, bandsAnimals, 4)
					mg.putRandomBandN(g, []monsterBand{PairWorm}, 1)
				} else {
					mg.putRandomBandN(g, bandsAnimals, 6)
				}
			}
			mg.putRandomBandN(g, bandsButterfly, 1)
			mg.putRandomBandN(g, bandsPlants, 5)
		}
	case 7:
		// 19
		mg.putRandomBandN(g, bandsHighGuard, 1)
		if mg.rand.IntN(2) == 0 {
			// 18
			mg.putRandomBandN(g, bandsGuard, 4)
			if mg.rand.IntN(3) == 0 {
				mg.putRandomBandN(g, bandDog, 4)
				mg.putRandomBandN(g, bandsAnimals, 2)
			} else {
				mg.putRandomBandN(g, bandsAnimals, 6)
			}
			mg.putRandomBandN(g, bandsButterfly, 1)
			mg.putRandomBandN(g, bandsRare, 2)
			mg.putRandomBandN(g, bandsBipeds, 3)
			mg.putRandomBandN(g, bandsPlants, 2)
		} else {
			// 18
			mg.putRandomBandN(g, bandsGuard, 1)
			mg.putRandomBandN(g, bandsRare, 4)
			mg.putRandomBandN(g, bandsButterfly, 1)
			if mg.rand.IntN(3) == 0 {
				mg.putRandomBandN(g, bandNadre, 7)
			} else {
				mg.putRandomBandN(g, bandsAnimals, 5)
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, []monsterBand{PairFrog}, 1)
				} else {
					mg.putRandomBandN(g, []monsterBand{PairDog}, 1)
				}
			}
			mg.putRandomBandN(g, bandsPlants, 5)
		}
	case 8:
		// 18-19
		mg.putRandomBandN(g, bandsHighGuard, 4)
		if mg.rand.IntN(2) == 0 {
			// 14
			mg.putRandomBandN(g, bandsGuard, 5)
			mg.putRandomBandN(g, bandsRare, 1)
			if mg.rand.IntN(3) == 0 {
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, bandOricCelmist, 6)
				} else {
					mg.putRandomBandN(g, bandMadNixe, 6)
				}
				mg.putRandomBandN(g, bandsBipeds, 2)
			} else {
				mg.putRandomBandN(g, bandsBipeds, 8)
			}
		} else {
			// 15
			mg.putRandomBandN(g, bandsGuard, 5)
			mg.putRandomBandN(g, bandsAnimals, 2)
			mg.putRandomBandN(g, bandsRare, 1)
			mg.putRandomBandN(g, bandHarmonicCelmistPair, 1)
			mg.putRandomBandN(g, bandsBipeds, 4)
			mg.putRandomBandN(g, bandsPlants, 1)
		}
	case 9:
		// 20-24
		mg.putRandomBandN(g, bandsHighGuard, 2)
		if mg.rand.IntN(2) == 0 {
			// 18
			mg.putRandomBandN(g, bandsGuard, 3)
			if mg.rand.IntN(2) == 0 {
				switch mg.rand.IntN(4) {
				case 0:
					mg.putRandomBandN(g, bandTreeMushroom, 4)
					mg.putRandomBandN(g, []monsterBand{PairTreeMushroom}, 1)
				case 1:
					mg.putRandomBandN(g, bandDragon, 6)
					mg.putRandomBandN(g, bandsAnimals, 2)
				case 2:
					mg.putRandomBandN(g, bandHazeCat, 4)
					mg.putRandomBandN(g, []monsterBand{PairHazeCat}, 1)
				case 3:
					mg.putRandomBandN(g, bandSpider, 4)
					mg.putRandomBandN(g, []monsterBand{PairSpider}, 1)
				}
			} else {
				mg.putRandomBandN(g, bandsRare, 4)
				mg.putRandomBandN(g, []monsterBand{PairTreeMushroom}, 1)
			}
			mg.putRandomBandN(g, bandsRare, 4)
			mg.putRandomBandN(g, bandsAnimals, 1)
			mg.putRandomBandN(g, bandsBipeds, 2)
			mg.putRandomBandN(g, bandsPlants, 2)
		} else {
			// 22+2
			mg.putRandomBandN(g, bandsButterfly, 2)
			mg.putRandomBandN(g, bandsGuard, 2)
			mg.putRandomBandN(g, bandsAnimals, 8)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandExplosiveNadrePair, 2)
			} else {
				mg.putRandomBandN(g, bandYackPair, 2)
			}
			mg.putRandomBandN(g, bandsRare, 3)
			mg.putRandomBandN(g, bandsPlants, 3)
			mg.putRandomBandN(g, []monsterBand{PairSatowalga}, 1)
		}
	case 10:
		// 22
		mg.putRandomBandN(g, bandsHighGuard, 3)
		if mg.rand.IntN(2) == 0 {
			// 19
			mg.putRandomBandN(g, bandsGuard, 7)
			mg.putRandomBandN(g, bandGuardPair, 1)
			mg.putRandomBandN(g, bandsRare, 2)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandsBipeds, 8)
			} else {
				mg.putRandomBandN(g, bandsBipeds, 4)
				mg.putRandomBandN(g, bandMirrorSpecter, 4)
			}
		} else {
			// 19
			mg.putRandomBandN(g, bandGuardPair, 1)
			if mg.rand.IntN(3) == 0 {
				mg.putRandomBandN(g, bandsGuard, 4)
				mg.putRandomBandN(g, bandVampire, 4)
				mg.putRandomBandN(g, []monsterBand{PairVampire}, 1)
				mg.putRandomBandN(g, bandsRare, 2)
			} else {
				mg.putRandomBandN(g, bandsGuard, 6)
				mg.putRandomBandN(g, bandsBipeds, 3)
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, []monsterBand{PairNixe}, 1)
				} else {
					mg.putRandomBandN(g, []monsterBand{PairOricCelmist}, 1)
				}
				mg.putRandomBandN(g, bandsAnimals, 2)
			}
			mg.putRandomBandN(g, bandsAnimals, 2)
			mg.putRandomBandN(g, bandsPlants, 1)
			mg.putRandomBandN(g, bandsRare, 1)
		}
	case 11:
		// 26
		mg.putRandomBandN(g, bandsHighGuard, 5)
		if mg.rand.IntN(2) == 0 {
			// 21
			mg.putRandomBandN(g, bandsGuard, 5)
			mg.putRandomBandN(g, bandsRare, 2)
			if mg.rand.IntN(3) == 0 {
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, bandOricCelmist, 5)
				} else {
					mg.putRandomBandN(g, bandHarmonicCelmist, 5)
				}
				mg.putRandomBandN(g, bandsBipeds, 5)
			} else {
				mg.putRandomBandN(g, bandsBipeds, 10)
			}
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandVampirePair, 1)
			} else {
				if mg.rand.IntN(2) == 0 {
					mg.putRandomBandN(g, bandOricCelmistPair, 1)
				} else {
					mg.putRandomBandN(g, bandHarmonicCelmistPair, 1)
				}
			}
			mg.putRandomBandN(g, bandsAnimals, 2)
		} else {
			// 21
			mg.putRandomBandN(g, bandsGuard, 5)
			mg.putRandomBandN(g, bandOricCelmistPair, 1)
			mg.putRandomBandN(g, []monsterBand{PairGuard}, 1)
			mg.putRandomBandN(g, bandsRare, 1)
			mg.putRandomBandN(g, bandsBipeds, 7)
			if mg.rand.IntN(2) == 0 {
				mg.putRandomBandN(g, bandHarmonicCelmistPair, 1)
			} else {
				mg.putRandomBandN(g, bandNixePair, 1)
			}
			mg.putRandomBandN(g, bandsAnimals, 1)
			mg.putRandomBandN(g, bandsPlants, 1)
		}
	}
}

func (g *Game) genBand(band monsterBand) []monsterKind {
	mbd := monsBands[band]
	if !mbd.Band {
		return []monsterKind{mbd.Monster}
	}
	bandMonsters := []monsterKind{}
	for m, n := range mbd.Distribution {
		for range n {
			bandMonsters = append(bandMonsters, m)
		}
	}
	return bandMonsters
}

func (mg *MapGen) bandInfoGuard(g *Game, band monsterBand, pl placeKind) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	p := invalidPos
	count := 0
loop:
	for p == invalidPos {
		count++
		if count > maxIterations {
			p = mg.insideCell(g)
			break
		}
		for range 20 {
			r := mg.vaults[RandInt(len(mg.vaults)-1)]
			for _, e := range r.places {
				if e.kind == PlaceSpecialStatic {
					p = r.RandomPlace(pl)
					break
				}
			}
			if p != invalidPos && !g.MonsterAt(p).Exists() {
				break loop
			}
		}
		r := mg.vaults[RandInt(len(mg.vaults)-1)]
		p = r.RandomPlace(pl)
	}
	bandinfo.Path = append(bandinfo.Path, p)
	bandinfo.Beh = BehGuard
	return bandinfo
}

func (mg *MapGen) bandInfoGuardSpecial(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	p := invalidPos
	count := 0
	for _, r := range mg.vaults {
		count++
		if count > 1 {
			log.Print("unavailable special guard position")
			p = mg.insideCell(g)
			break
		}
		p = r.RandomPlace(PlacePatrolSpecial)
		if p != invalidPos && !g.MonsterAt(p).Exists() {
			break
		}
	}
	bandinfo.Path = append(bandinfo.Path, p)
	bandinfo.Beh = BehGuard
	return bandinfo
}

func (mg *MapGen) bandInfoPatrol(g *Game, band monsterBand, pl placeKind) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	p := invalidPos
	count := 0
	for p == invalidPos || g.MonsterAt(p).Exists() {
		count++
		if count > 4000 {
			p = mg.insideCell(g)
			break
		}
		p = mg.vaults[RandInt(len(mg.vaults)-1)].RandomPlace(pl)
	}
	target := invalidPos
	count = 0
	for target == invalidPos {
		// TODO: only find place in other room?
		count++
		if count > 4000 {
			target = mg.insideCell(g)
			break
		}
		target = mg.vaults[RandInt(len(mg.vaults)-1)].RandomPlace(pl)
	}
	bandinfo.Path = append(bandinfo.Path, p)
	bandinfo.Path = append(bandinfo.Path, target)
	bandinfo.Beh = BehPatrol
	return bandinfo
}

func (mg *MapGen) bandInfoPatrolSpecial(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	p := invalidPos
	count := 0
	for _, r := range mg.vaults {
		count++
		if count > 1 {
			log.Print("unavailable special patrol position")
			p = mg.insideCell(g)
			break
		}
		p = r.RandomPlace(PlacePatrolSpecial)
		if p != invalidPos && !g.MonsterAt(p).Exists() {
			break
		}
	}
	target := invalidPos
	count = 0
	for _, r := range mg.vaults {
		count++
		if count > 1 {
			log.Print("unavailable special second patrol position")
			p = mg.insideCell(g)
			break
		}
		target = r.RandomPlace(PlacePatrolSpecial)
		if target != invalidPos {
			break
		}
	}
	bandinfo.Path = append(bandinfo.Path, p)
	bandinfo.Path = append(bandinfo.Path, target)
	bandinfo.Beh = BehPatrol
	return bandinfo
}

// insideCell returns a free room cell.
func (mg *MapGen) insideCell(g *Game) gruid.Point {
	count := 0
	for {
		count++
		if count > 1500 {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		if Distance(p, g.Player.P) < DefaultFOVRange {
			continue
		}
		t := mg.m.At(p)
		if mg.vault[p] && (t == FoliageCell || t == GroundCell) {
			return p
		}
	}
}

func (mg *MapGen) bandInfoOutsideGround(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.outsideGroundCell(g))
	bandinfo.Beh = BehWander
	return bandinfo
}

func (mg *MapGen) bandInfoSatowalga(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.satowalgaCell(g))
	bandinfo.Beh = BehWander
	return bandinfo
}

func (mg *MapGen) satowalgaCell(g *Game) gruid.Point {
	// Try to find a suitable cell outside of vaults.
	for _, p := range mg.rndPos {
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if (t == GroundCell || t == CavernCell) && !mg.vault[p] && !mg.m.HasTooManyWallNeighbors(p) {
			return p
		}
	}
	// Otherwise, try to find a suitable cell inside of vaults or in some lateral wall.
	for _, p := range mg.rndPos {
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if (t == GroundCell || t == CavernCell) && mg.vault[p] && !mg.m.HasTooManyWallNeighbors(p) {
			return p
		}
		if t == WallCell && g.HasNonWallNeighbors(p) {
			g.Map.Set(p, GroundCell)
			return p
		}
	}
	// Should not happen.
	log.Print("mapgen warning: no suitable position for satowalga")
	return g.freeCellForMonster()
}

func (mg *MapGen) bandInfoOutside(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.outsideCell(g))
	bandinfo.Beh = BehWander
	return bandinfo
}

func (mg *MapGen) bandInfoOutsideExplore(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.outsideCell(g))
	for range 4 {
		for range 100 {
			p := mg.outsideCell(g)
			if mg.PR.CCMapAt(p) == mg.PR.CCMapAt(bandinfo.Path[0]) {
				bandinfo.Path = append(bandinfo.Path, p)
				break
			}
		}
	}
	bandinfo.Beh = BehExplore
	return bandinfo
}

func (mg *MapGen) bandInfoOutsideExploreButterfly(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.outsideCell(g))
	for range 9 {
		for range 100 {
			p := mg.outsideCell(g)
			if mg.PR.CCMapAt(p) == mg.PR.CCMapAt(bandinfo.Path[0]) {
				bandinfo.Path = append(bandinfo.Path, p)
				break
			}
		}
	}
	bandinfo.Beh = BehExplore
	return bandinfo
}

// outsideCell returns a free non-vault position.
func (mg *MapGen) outsideCell(g *Game) gruid.Point {
	count := 0
	for {
		count++
		if count > 1500 {
			panic("infinite loop")
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if !mg.vault[p] && (t == FoliageCell || t == GroundCell) {
			return p
		}
	}
}

func (mg *MapGen) bandInfoFoliage(g *Game, band monsterBand) bandInfo {
	bandinfo := bandInfo{Kind: monsterBand(band)}
	bandinfo.Path = append(bandinfo.Path, mg.foliageCell(g))
	bandinfo.Beh = BehWander
	return bandinfo
}

// foliageCell returns a free foliage position.
func (mg *MapGen) foliageCell(g *Game) gruid.Point {
	count := 0
	for {
		count++
		if count > 1500 {
			return mg.outsideCell(g)
		}
		x := mg.rand.IntN(MapWidth)
		y := mg.rand.IntN(MapHeight)
		p := gruid.Point{x, y}
		if Distance(p, g.Player.P) < DefaultFOVRange {
			continue
		}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		t := mg.m.At(p)
		if t == FoliageCell {
			return p
		}
	}
}

// freeCellForMonster returns a free position for monster placement.
func (g *Game) freeCellForMonster() gruid.Point {
	d := g.Map
	count := 0
	for {
		count++
		if count > maxIterations {
			panic("infinite loop")
		}
		x := RandInt(MapWidth)
		y := RandInt(MapHeight)
		p := gruid.Point{x, y}
		t := d.At(p)
		switch t {
		case CavernCell, GroundCell, QueenRockCell, FoliageCell:
		default:
			continue
		}
		if g.Player != nil && Distance(g.Player.P, p) < 8 {
			continue
		}
		mons := g.MonsterAt(p)
		if mons.Exists() {
			continue
		}
		return p
	}
}

func (g *Game) freeCellForBandMonster(p gruid.Point) gruid.Point {
	count := 0
	for {
		count++
		if count > maxIterations {
			return g.freeCellForMonster()
		}
		neighbors := g.PlayerPassableNeighbors(p)
		if len(neighbors) == 0 {
			// should not happen
			log.Printf("no neighbors for %v", p)
			count = maxIterations + 1
			continue
		}
		r := RandInt(len(neighbors))
		p = neighbors[r]
		if g.Player != nil && Distance(g.Player.P, p) < 8 {
			continue
		}
		mons := g.MonsterAt(p)
		t := g.Map.At(p)
		if mons.Exists() {
			continue
		}
		switch t {
		case CavernCell, GroundCell, QueenRockCell, FoliageCell:
		default:
			continue
		}
		return p
	}
}

// putMonsterBand generates the given monster band.
func (mg *MapGen) putMonsterBand(g *Game, band monsterBand) bool {
	monsters := g.genBand(band)
	if monsters == nil {
		return false
	}
	awake := mg.rand.IntN(5) > 0
	var bdinf bandInfo
	switch band {
	case LoneYack, LoneWorm, PairYack:
		bdinf = mg.bandInfoFoliage(g, band)
	case LoneDog, LoneHarpy:
		bdinf = mg.bandInfoOutsideGround(g, band)
	case LoneBlinkingFrog, LoneExplosiveNadre, PairExplosiveNadre:
		bdinf = mg.bandInfoOutside(g, band)
	case LoneMirrorSpecter, LoneWingedMilfid, LoneVampire, PairWingedMilfid, LoneEarthDragon, LoneHazeCat, LoneSpider:
		bdinf = mg.bandInfoOutsideExplore(g, band)
	case LoneButterfly:
		bdinf = mg.bandInfoOutsideExploreButterfly(g, band)
	case LoneTreeMushroom, LoneAcidMound:
		bdinf = mg.bandInfoOutside(g, band)
	case LoneHighGuard:
		bdinf = mg.bandInfoGuard(g, band, PlacePatrol)
	case LoneSatowalgaPlant, PairSatowalga:
		bdinf = mg.bandInfoSatowalga(g, band)
	case SpecialLoneVampire, SpecialLoneNixe, SpecialLoneMilfid, SpecialLoneOricCelmist, SpecialLoneHarmonicCelmist, SpecialLoneHighGuard,
		SpecialLoneHarpy, SpecialLoneTreeMushroom, SpecialLoneMirrorSpecter, SpecialLoneHazeCat, SpecialLoneSpider, SpecialLoneAcidMound, SpecialLoneDog, SpecialLoneExplosiveNadre, SpecialLoneYack, SpecialLoneBlinkingFrog:
		if mg.rand.IntN(5) > 0 {
			bdinf = mg.bandInfoPatrolSpecial(g, band)
		} else {
			bdinf = mg.bandInfoGuardSpecial(g, band)
		}
		if !awake && mg.rand.IntN(2) > 0 {
			awake = true
		}
	case UniqueCrazyImp:
		bdinf = mg.bandInfoOutside(g, band)
		bdinf.Beh = BehCrazyImp
	default:
		bdinf = mg.bandInfoPatrol(g, band, PlacePatrol)
	}
	g.MonsterBands = append(g.MonsterBands, bdinf)
	var p gruid.Point
	if len(bdinf.Path) == 0 {
		// should not happen now
		p = g.freeCellForMonster()
	} else {
		p = bdinf.Path[0]
		if g.MonsterAt(p).Exists() {
			log.Printf("already %s at %v: no place for %v", g.MonsterAt(p).Kind, p, monsters)
			p = g.freeCellForMonster()
		}
	}
	for i, mk := range monsters {
		mons := &Monster{Kind: mk}
		if awake {
			mons.State = Wandering
		}
		g.Monsters = append(g.Monsters, mons)
		mons.Init()
		mons.Index = len(g.Monsters) - 1
		mons.Band = len(g.MonsterBands) - 1
		mons.PlaceAtStart(g, p)
		mons.Target = mons.NextTarget(g)
		if i < len(monsters)-1 {
			p = g.freeCellForBandMonster(p)
		}
	}
	return true
}

func (mg *MapGen) putRandomBand(g *Game, bands []monsterBand) bool {
	return mg.putMonsterBand(g, bands[RandInt(len(bands))])
}

func (mg *MapGen) putRandomBandN(g *Game, bands []monsterBand, n int) {
	for range n {
		mg.putMonsterBand(g, bands[RandInt(len(bands))])
	}
}
