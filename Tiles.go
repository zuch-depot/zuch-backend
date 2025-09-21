package main

// --------------------------------------------------
type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
}

func (t Tile) addTrack(i int) {

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
}

// --------------------------------------------------
type AtiveTileType int

const (
	CoalMine AtiveTileType = iota //all following are increasing int
	Coalplant
	PotatoFarm
	ChipShop
)
