package main

// --------------------------------------------------
type Tile struct {
	tracks      [4]bool
	signals     [4]bool
	isPlattform bool
	isBlocked   bool
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
