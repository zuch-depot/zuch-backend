package ds

import (
	"fmt"

	"zuch-backend/internal/utils"
)

type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool // nur für tracks
	ActiveTile  *ActiveTile
	X           int
	Y           int

	// ob auf dem Tile gebaut werden kann. Wenn es ein Hindernis ist, muss irgendwie bestimmt werden,
	// welches Sprite das ist, es können aber auch Tracks und Signale Mapseitig unveränderlich sein.
	IsLocked bool
}

// Definiert werden muss: Id, Category, Name, MaxStorage?
type ActiveTile struct {
	Id       int
	Category *ActiveTileCategory
	Name     string
	Level    int
	// Stations   []*Station
	Stations   []int // Stationen, die in der Nähe sind. wird mit changeStationTile verwaltet
	Storage    map[string]int
	MaxStorage int // STANDARD für alle gleich? maximum Lager pro Gut -> sonst kann es zu unwiederruflichen auffüllen kommen. Nur für Verbrauchsgüter der Produktion
	X          int
	Y          int // das frontend dankt :)
}

// return das Tile bei den koordinaten, kontrolliert ob die in Bounds sind
func (gs *GameState) GetTile(X int, Y int) (*Tile, error) {
	if !(0 <= X && X < gs.ConfigData.SizeX) {
		return &Tile{}, fmt.Errorf("X index is not in bounds %d", gs.ConfigData.SizeX)
	}
	if !(0 <= Y && Y < gs.ConfigData.SizeY) {
		return &Tile{}, fmt.Errorf("Y index is not in bounds %d", gs.ConfigData.SizeY)
	}

	return gs.Tiles[X][Y], nil
}

// returnt alle Produktionstypen mit der kumulierten maximalen Produktion pro Tick
func (a *ActiveTile) getProductionCategorys() map[string]int {
	productionTypes := make(map[string]int)

	for _, prodCycle := range a.Category.Productioncycles {
		for prodType, prodTypeAmmount := range prodCycle.Produktion {
			productionTypes[prodType] += prodTypeAmmount // TODO Level berücksichtigen
		}
	}

	return productionTypes
}

// returnt alle Verbrauchstypen mit der kumulierten maximalen Verbrauch pro Tick
func (a *ActiveTile) getConsumptionCategorys() map[string]int {
	consumptionTypes := make(map[string]int)

	for _, prodCycle := range a.Category.Productioncycles {
		for consType, consTypeAmmount := range prodCycle.Consumtion {
			consumptionTypes[consType] += consTypeAmmount // TODO Level berücksichtigen
		}
	}
	return consumptionTypes
}

// Überprüft utils.CheckName() und setzt namen
func (a *ActiveTile) Rename(name string) error {
	err := utils.CheckName(name)
	if err != nil {
		return err
	}

	a.Name = name
	return nil
}

// Fügt bei i ein gleis hinzu, wenn da keins ist, da keine Plattform ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
func (t *Tile) AddTrack(subtile int, gs *GameState) error {
	if t.ActiveTile.Category != nil {
		return fmt.Errorf("There is a Active Tile there, no Tracks can be build on this tile.")
	}
	if t.IsBlocked {
		return fmt.Errorf("This Tile is blocked, the track could not be build")
	}
	if t.IsPlattform {
		return fmt.Errorf("The Tile is a Plattform, no Track can be build there.")
	}

	if !t.Tracks[subtile-1] {
		t.Tracks[subtile-1] = true
		gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: *t}
		return nil
	}
	return fmt.Errorf("There is already a Track at that Position.")
}

// Entfernt bei i ein gleis und Signal, wenn da eins ist und das Tile nicht locked ist
// wenn ein Signal an der Stelle ist, wird kein Fehler geworfen und dieses entfernt
func (t *Tile) RemoveTrack(subtile int, gs *GameState) error {
	if t.IsBlocked {
		return fmt.Errorf("The Tile is blocked, try again later.")
	}

	if !t.Tracks[subtile-1] {
		return fmt.Errorf("There is no Track to Remove.")
	}

	// ggf. aktivTiles?

	// ggf. entfernen des Signals
	if t.Signals[subtile-1] {
		t.RemoveSignal(subtile, gs)
	}

	// entfernt Station, wenn das Tile eine ist
	if t.IsPlattform {
		gs.RemoveStationTile([2]int{t.X, t.Y})
	}

	// entfernen des Tracks
	t.Tracks[subtile-1] = false

	gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: *t}
	return nil
}

// Fügt bei i ein Signal hinzu, wenn da keins ist und ein entsprechendes Gleis vorhanden ist,
// um bei i ein signal zu bauen muss gleis i da sein, und das Tile nicht locked ist;
func (t *Tile) AddSignal(subtile int, gs *GameState) error {
	// ist subTile valid?
	if 1 > subtile || subtile > 4 {
		return fmt.Errorf("An error accured while adding a signal. Please provide valid subtile coordinates between 1 and 4.")
	}

	if t.ActiveTile.Category != nil {
		return fmt.Errorf("There is a Active Tile there, no Signal can be build on this tile.")
	}

	if t.IsPlattform {
		return fmt.Errorf("The Tile is a Plattform, no Signal can be build there.")
	}

	if !t.Tracks[subtile-1] {
		return fmt.Errorf("There are no Track to place the signal.")
	}

	if t.Signals[subtile-1] {
		return fmt.Errorf("There is already a signal at that location.")
	}

	if t.IsBlocked {
		return fmt.Errorf("The Tile is currently blocked, the signal could not be build.")
	}

	// hinzugügen des Signals
	t.Signals[subtile-1] = true
	gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: *t}
	return nil
}

// Entfernt bei i das Signal, wenn da eins ist und das Tile nicht locked ist
func (t *Tile) RemoveSignal(subtile int, gs *GameState) error {
	if !t.Signals[subtile-1] {
		return fmt.Errorf("There is no Signal to remove.")
	}

	if t.IsBlocked {
		return fmt.Errorf("The Tile is currently blocked, the signal could not be demolished.")
	}

	t.Signals[subtile-1] = false
	gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: *t}
	return nil
}

// bis jetzt noch keine Level implementiert
func (gs *GameState) ProcessActiveTiles() {
	for _, activeTile := range gs.ActiveTiles {
		// wenn noch kein Map erstellt wurde, dann erstelle Map
		// maybe irgendwo anders?
		if len(activeTile.Storage) == 0 {
			activeTile.Storage = make(map[string]int)
		}

		// Waren aus umliegenden Stationen holen und max. Produktionsrate bestimmen
		for _, prodCyle := range activeTile.Category.Productioncycles {
			// ein Produktionszyklus

			// bestimmt den Prozentsatz, wie viel in diesem Durchlauf produziert werden kann, limitiert durch Ressourcenverfügbarkeit
			possibleProduction := 1.0

			// alle Zutaten durchiteriren
			for cargoTypeToConsume, neededQuantity := range prodCyle.Consumtion {

				// hole die Ware Cargotype wenn möglich aus umligenden Stationen, wenn das Lager noch nicht voll ist
				emptySpaceInActiveTile := activeTile.MaxStorage - activeTile.Storage[cargoTypeToConsume]
				if emptySpaceInActiveTile > 0 {
					for _, stationId := range activeTile.Stations {
						station := gs.Stations[stationId]
						// wenn nichts von dem Typ in der Station gelagert ist
						if station.Storage[cargoTypeToConsume] == 0 {
							continue
						}
						// wenn der Inhalt der Station rein passt, einfach alles reinschieben (gerade kein Limit)
						if emptySpaceInActiveTile > station.Storage[cargoTypeToConsume] {
							activeTile.Storage[cargoTypeToConsume] += station.Storage[cargoTypeToConsume]
							station.Storage[cargoTypeToConsume] = 0
						} else {
							// sonst leeren Platz füllen
							activeTile.Storage[cargoTypeToConsume] = activeTile.MaxStorage
							station.Storage[cargoTypeToConsume] -= emptySpaceInActiveTile
						}
					}
				}

				// max. Produktionsrate bestimmen anhand der Vorräte
				// wenn der aktuell max. Verbrauch größer ist als die Vorräte, dann wird möglicher Verbrauch aufs maximum reduziert
				if int(float64(neededQuantity)*possibleProduction) > activeTile.Storage[cargoTypeToConsume] {
					possibleProduction = float64(activeTile.Storage[cargoTypeToConsume]) / float64(neededQuantity)
				}

			}

			// max Produktionsrate bestimmen anhand der vorhandenen Kapazität für die Produkte
			// NICHT, wenn die überschüssigen Waren gelöscht werden sollen. Dann nur anpassung bei der Produktion anpassen
			// for cargoTypeToProduce, maxProducedQuantity := range prodCyle.Produktion {
			// 	//wenn die zu produzierende Menge größer ist als der vorhandene Platz in der Station, reduziere, dass es passt
			// 	producingQuantity := int(float64(maxProducedQuantity) * possibleProduction)
			// 	emptySpaceInTile := activeTile.maxStorage - activeTile.Storage[cargoTypeToProduce]
			// 	if producingQuantity > emptySpaceInTile {
			// 		possibleProduction = float64(emptySpaceInTile) / float64(maxProducedQuantity)
			// 	}
			// }

			//produziere die Waren
			//alle Zutaten durchiterieren und rausnehmen
			//(es müssen alle erst durchiteriert werden, um die Produktionsrate zu bestimmen)
			for cargoTypeToConsume, neededQuantity := range prodCyle.Consumtion {
				activeTile.Storage[cargoTypeToConsume] -= int(float64(neededQuantity) * possibleProduction)
			}

			// alle produzierten Güter in dieser Produktionslinie durchiteriren und erhöhen um Produktionsrate
			for cargoTypeToProduce, maxProducedQuantity := range prodCyle.Produktion {
				activeTile.Storage[cargoTypeToProduce] += int(float64(maxProducedQuantity) * possibleProduction)
			}

		}

		// Verteile Ergebnis an umliegende Stationen, bei denen die Ware abgeholt werden kann
		prodKategorien := activeTile.getProductionCategorys() // Alle Protionskategorien, die bei dem aktivTile produziert werden
		for prodCat := range prodKategorien {
			// Erst angrenzende Stationen bestimmen, bei denen die Ware abgeholt werden kann
			stations := make([]*Station, 0)
			for _, stationId := range activeTile.Stations {
				station := gs.Stations[stationId]
				for _, schedule := range gs.Schedules {
					// Ist die Station in dem Schedule als Stop enthalten mit einem passenden LoadCommand?
					for _, stop := range schedule.Stops {

						stationFound := false
						// ist es eine Plattform der Station?
						if stop.IsPlattform && stop.Plattform.isPlattfromFromStation(station) {

							// hat die Plattform bei dem Stop einen LoadCommand für den Typen oder keinen LoadCommand?
							loadTypes := stop.LoadUnloadCommand[1].CargoTypes

							if len(loadTypes) == 0 {
								stationFound = true
							} else {
								for _, loadType := range loadTypes {
									if loadType == prodCat {
										stationFound = true
										break
									}
								}
							}
						}
						if stationFound {
							stations = append(stations, station)
						}
					}
				}
			}

			// jetzt gleichmäßig auf die Stationen aufteilen
			numberStations := len(stations)
			for _, station := range stations {
				quantityToAdd := activeTile.Storage[prodCat] / numberStations
				activeTile.Storage[prodCat] -= quantityToAdd - station.addCargo(prodCat, quantityToAdd, gs)
			}

		}
		gs.BroadcastChannel <- WsEnvelope{Type: "activeTile.update", Msg: activeTile}
	}
}
