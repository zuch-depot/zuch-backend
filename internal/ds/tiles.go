package ds

import (
	"fmt"
)

type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
	ActiveTile  *ActiveTile
	X           int
	Y           int

	//ob auf dem Tile gebaut werden kann. Wenn es ein Hindernis ist, muss irgendwie bestimmt werden,
	//welches Sprite das ist, es können aber auch Tracks und Signale Mapseitig unveränderlich sein.
	IsLocked bool
}

// Definiert werden muss: Id, Category, Name, MaxStorage?
type ActiveTile struct {
	Id         int
	Category   *ActiveTileCategory
	Name       string
	Level      int
	Stations   []*Station //Stationen, die in der Nähe sind. wird mit changeStationTile verwaltet
	Storage    map[string]int
	MaxStorage int //STANDARD für alle gleich? maximum Lager pro Gut -> sonst kann es zu unwiederruflichen auffüllen kommen. Nur für Verbrauchsgüter der Produktion
}

func (gs *GameState) GetTile(X int, Y int) (*Tile, error) {
	if !(0 <= X && X <= gs.SizeX) {
		return &Tile{}, fmt.Errorf("X index is not in bounds %d", gs.SizeX)
	}
	if !(0 <= Y && Y <= gs.SizeY) {
		return &Tile{}, fmt.Errorf("Y index is not in bounds %d", gs.SizeY)
	}

	return gs.Tiles[X][Y], nil
}

// returnt alle Produktionstypen mit der kumulierten maximalen Produktion pro Tick
func (a *ActiveTile) getProductionCategorys() map[string]int {
	var productionTypes = make(map[string]int)

	for _, prodCycle := range a.Category.Productioncycles {
		for prodType, prodTypeAmmount := range prodCycle.Produktion {
			productionTypes[prodType] += prodTypeAmmount //TODO Level berücksichtigen
		}
	}

	return productionTypes
}

// returnt alle Verbrauchstypen mit der kumulierten maximalen Verbrauch pro Tick
func (a *ActiveTile) getConsumptionCategorys() map[string]int {
	var consumptionTypes = make(map[string]int)

	for _, prodCycle := range a.Category.Productioncycles {
		for consType, consTypeAmmount := range prodCycle.Consumtion {
			consumptionTypes[consType] += consTypeAmmount //TODO Level berücksichtigen
		}
	}
	return consumptionTypes
}

// Fügt bei i ein gleis hinzu, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
func (t *Tile) AddTrack(subtile int, gs *GameState) (bool, error) {
	if t.ActiveTile.Category != nil {
		return false, fmt.Errorf("There is a Active Tile there, no Tracks can be build on this tile.")
	}
	if !t.Tracks[subtile-1] {
		t.Tracks[subtile-1] = true
		gs.BroadcastChannel <- WsEnvelope{Type: "rail.create", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, nil
	}
	return false, fmt.Errorf("There is already a Track at that Position.")
}

// Entfernt bei i ein gleis und Signal, wenn da eins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
// wenn kein Signal an der Stelle ist, wird kein Fehler geworfen
func (t *Tile) RemoveTrack(subtile int, gs *GameState) (bool, error) {
	if t.Tracks[subtile-1] && !t.IsBlocked {
		t.Tracks[subtile-1] = false
		t.RemoveSignal(subtile, gs) // das muss noch kommunitzier werden
		gs.BroadcastChannel <- WsEnvelope{Type: "rail.remove", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, nil
	}
	return false, fmt.Errorf("There is no Track to Remove, or the Tile may be blocked by a Train, if so try again later.")

}

// Fügt bei i ein Signal hinzu, wenn da keins ist und ein entsprechendes Gleis vorhanden ist,
// um bei i ein signal zu bauen muss gleis i da sein, und das Tile nicht locked ist;
// returnt true bei erfolg und false bei error
func (t *Tile) AddSignal(subtile int, gs *GameState) (bool, error) {
	if t.ActiveTile.Category != nil {
		return false, fmt.Errorf("There is a Active Tile there, no Signal can be build on this tile.")
	}
	if t.Tracks[subtile-1] || t.Signals[subtile-1] {
		t.Signals[subtile-1] = true
		gs.BroadcastChannel <- WsEnvelope{Type: "signal.create", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, nil
	}
	return false, fmt.Errorf("There may be no Track to place the signal onto, or there is already a signal at that location.")
}

// Entfernt bei i das Signal, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
// TODO error
func (t *Tile) RemoveSignal(subtile int, gs *GameState) (bool, error) {
	if t.Signals[subtile-1] {
		t.Signals[subtile-1] = false
		gs.BroadcastChannel <- WsEnvelope{Type: "signal.remove", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, nil
	}
	return false, fmt.Errorf("There is no Signal to remove.")

}
