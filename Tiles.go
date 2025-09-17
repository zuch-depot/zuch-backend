package main

type TileType struct {
}

// --------------------------------------------------
type Obstacle struct {
	TileType
	removable bool
}

// --------------------------------------------------
type Tile struct {
	TileType
	tracks      []bool
	signals     []bool
	isPlattform bool
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
type AktiveTile struct {
	TileType
	storages []CargoStorage
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
