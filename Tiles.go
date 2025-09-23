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

// konkrete Güter, wie z.B. Kohle, Eisen, etc...
type CargoType string

const (
	Coal    CargoType = "Kohle"
	Iron    CargoType = "Eisen"
	Potatos CargoType = "Kartoffeln"
)

// Güterarten zusammengefasst, wie z.B. Schüttgut oder Flüssigkeiten
type CargoCategory string

const (
	Food      CargoCategory = "Lebensmittel"
	BulkGoods CargoCategory = "Schüttgut"
	Liquid    CargoCategory = "Flüssigkeiten"
)
