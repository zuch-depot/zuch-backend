package main

// --------------------------------------------------
type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
}

// Fügt bei i ein gleis hinzu, wenn da keins ist
// returnt true bei erfolg und false bei error
func (t *Tile) addTrack(i int) (bool, string) {
	if !t.Tracks[i-1] {
		t.Tracks[i-1] = true
		return true, ""
	}
	return false, "There is already a Track at that Position"
}

// Entfernt bei i ein gleis, wenn da eins ist
// returnt true bei erfolg und false bei error
func (t *Tile) removeTracks(i int) (bool, string) {
	if t.Tracks[i-1] || t.IsBlocked {
		t.Tracks[i-1] = false
		return true, ""
	}
	return false, "There is no Track to Remove, or the Tile may be blocked by a Train, if so try again later"

}

// Fügt bei i ein Signal hinzu, wenn da keins ist und ein entsprechendes Gleis vorhanden ist, um bei i ein signal zu bauen muss gleis i da sein
// returnt true bei erfolg und false bei error
func (t *Tile) addSignal(i int) (bool, string) {
	if t.Tracks[i-1] || t.Signals[i-1] {
		t.Signals[i-1] = true
		return true, ""
	}
	return false, "There may be no Track to place the signal onto, or there is already a signal at that location"
}

// Fügt bei i ein Signal hinzu, wenn da keins ist
// returnt true bei erfolg und false bei error
func (t Tile) removeSignal(i int) (bool, string) {
	if t.Signals[i-1] {
		t.Signals[i-1] = false
		return true, ""
	}
	return false, "There is no Signal to remove"

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
