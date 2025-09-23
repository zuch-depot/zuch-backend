package main

// --------------------------------------------------
type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
}

func (t *Tile) addTrack(i int) {
	t.Tracks[i-1] = true
}

func (t Tile) removeTracks() {

}

func (t Tile) addSignal(i int) {

}

func (t Tile) removeSignals() {

}

func (t Tile) removePlattform() {

}

// --------------------------------------------------

type CargoStorage struct {
	capacity  int
	filled    int
	CargoType CargoType
}

type AtiveTileType int

const (
	CoalMine AtiveTileType = iota //all following are increasing int
	Coalplant
	PotatoFarm
	ChipShop
)

type CargoType int

const (
	Coal CargoType = iota //all following are increasing int
	Iron
	Potatos
)

type TestType string

const (
	Morgen TestType = "Morgen"
	Abend  TestType = "Abend"
)

type TestTypeType []TestType
