package main

type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
}

type AtiveTileType int

const (
	CoalMine AtiveTileType = iota //all following are increasing int
	Coalplant
	PotatoFarm
	ChipShop
)
