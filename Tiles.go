package main

type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
	ActiveTile  *ActiveTile

	//ob auf dem Tile gebaut werden kann. Wenn es ein Hindernis ist, muss irgendwie bestimmt werden,
	//welches Sprite das ist, es können aber auch Tracks und Signale Mapseitig unveränderlich sein.
	IsLocked bool
}

type ActiveTile struct {
	Category   *ActiveTileCategory
	Name       string
	Level      int
	Stations   []*Station //Stationen, die in der Nähe sind. wird mit changeStationTile verwaltet
	Storage    map[string]int
	maxStorage int //maximum Lager pro Gut -> sonst kann es zu unwiederruflichen auffüllen kommen

}

// Fügt bei i ein gleis hinzu, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
// TODO MUSS ANGEPASSt WERDEN, damit Stationen nicht kaputt gemacht werden
func (t *Tile) addTrack(i int) (bool, string) {
	if t.IsBlocked {
		return false, "This Tile is locked and cannot be altered"
	}
	if !t.Tracks[i-1] {
		t.Tracks[i-1] = true
		return true, ""
	}
	return false, "There is already a Track at that Position."
}

// Entfernt bei i ein gleis und Signal, wenn da eins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
// wenn kein Signal an der Stelle ist, wird kein Fehler geworfen
// TODO MUSS ANGEPASST WERDEN, Stationen müssen berücksichtigt werden
func (t *Tile) removeTrack(i int) (bool, string) {
	if t.IsLocked {
		return false, "This Tile is locked and cannot be altered"
	}
	if t.Tracks[i-1] && !t.IsBlocked {
		t.Tracks[i-1] = false
		t.Signals[i-1] = false
		return true, ""
	}
	return false, "There is no Track to Remove, or the Tile may be blocked by a Train, if so try again later."

}

// Fügt bei i ein Signal hinzu, wenn da keins ist und ein entsprechendes Gleis vorhanden ist,
// um bei i ein signal zu bauen muss gleis i da sein, und das Tile nicht locked ist;
// returnt true bei erfolg und false bei error
func (t *Tile) addSignal(i int) (bool, string) {
	if t.IsLocked {
		return false, "This Tile is locked and cannot be altered"
	}
	if t.Tracks[i-1] || t.Signals[i-1] {
		t.Signals[i-1] = true
		return true, ""
	}
	return false, "There may be no Track to place the signal onto, or there is already a signal at that location."
}

// Fügt bei i ein Signal hinzu, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
func (t *Tile) removeSignal(i int) (bool, string) {
	if t.IsLocked {
		return false, "This Tile is locked and cannot be altered"
	}
	if t.Signals[i-1] {
		t.Signals[i-1] = false
		return true, ""
	}
	return false, "There is no Signal to remove."

}

func (t Tile) removePlattform() (bool, string) {
	if t.IsLocked {
		return false, "This Tile is locked and cannot be altered"
	}
	return true, "temp"
}

// bis jetzt noch keine Level implementiert
func processActiveTiles() {

	for _, activeTile := range activeTiles {
		//wenn noch kein Map erstellt wurde, dann erstelle Map
		//maybe irgendwo anders?
		if len(activeTile.Storage) == 0 {
			activeTile.Storage = make(map[string]int)
		}

		//Waren aus umliegenden Stationen holen und max. Produktionsrate bestimmen
		for _, prodCyle := range activeTile.Category.Productioncycles {
			//ein Produktionszyklus

			//bestimmt den Prozentsatz, wie viel in diesem Durchlauf produziert werden kann, limitiert durch Ressourcenverfügbarkeit
			possibleProduction := 1.0

			//alle Zutaten durchiteriren
			for cargoTypeToConsume, neededQuantity := range prodCyle.Consumtion {

				//hole die Ware Cargotype wenn möglich aus umligenden Stationen, wenn das Lager noch nicht voll ist
				emptySpaceInActiveTile := activeTile.maxStorage - activeTile.Storage[cargoTypeToConsume]
				if emptySpaceInActiveTile > 0 {
					for _, station := range activeTile.Stations {
						//wenn nichts von dem Typ in der Station gelagert ist
						if station.Storage[cargoTypeToConsume] == 0 {
							continue
						}
						//wenn der Inhalt der Station rein passt, einfach alles reinschieben (gerade kein Limit)
						if emptySpaceInActiveTile > station.Storage[cargoTypeToConsume] {
							activeTile.Storage[cargoTypeToConsume] += station.Storage[cargoTypeToConsume]
							station.Storage[cargoTypeToConsume] = 0
						} else {
							//sonst leeren Platz füllen
							activeTile.Storage[cargoTypeToConsume] = activeTile.maxStorage
							station.Storage[cargoTypeToConsume] -= emptySpaceInActiveTile
						}
					}
				}

				//max. Produktionsrate bestimmen anhand der Vorräte
				//wenn der aktuell max. Verbrauch größer ist als die Vorräte, dann wird möglicher Verbrauch aufs maximum reduziert
				if int(float64(neededQuantity)*possibleProduction) > activeTile.Storage[cargoTypeToConsume] {
					possibleProduction = float64(activeTile.Storage[cargoTypeToConsume]) / float64(neededQuantity)
				}

			}

			//max Produktionsrate bestimmen anhand der vorhandenen Kapazität für die Produkte
			//NICHT, wenn die überschüssigen Waren gelöscht werden sollen. Dann nur anpassung bei der Produktion anpassen
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

			//alle produzierten Güter in dieser Produktionslinie durchiteriren und erhöhen um Produktionsrate
			for cargoTypeToProduce, maxProducedQuantity := range prodCyle.Produktion {
				activeTile.Storage[cargoTypeToProduce] += int(float64(maxProducedQuantity) * possibleProduction)
			}

		}

		//Verteile Ergebnis an umliegende Stationen
		//an der Stelle richtig
		//MUSS ersetzt werden mit CargoPathfinding

		for _, prodCyle := range activeTile.Category.Productioncycles {
			//macht eine Liste aller Güter, die in dem Tile produziert werden
			var typesProducing []string
			for cargoTypeToProduce := range prodCyle.Produktion {
				typesProducing = append(typesProducing, cargoTypeToProduce)
			}

			//geht die Stationen durch und versucht so so viel einzufügen von den CargoTypes,
			//die produziert werden, wie gelagert ist, durch die Anzahl der Stationen
			//das wird natürlich aus dem Lager des Tiles entfernt
			numberStations := len(activeTile.Stations)
			for _, station := range activeTile.Stations {
				for _, cargoTypeProducing := range typesProducing {
					quantityToAdd := activeTile.Storage[cargoTypeProducing] / numberStations
					activeTile.Storage[cargoTypeProducing] -= quantityToAdd - station.addCargo(cargoTypeProducing, quantityToAdd)
				}
			}
		}

	}
}

type Car struct {
}

func (A *ActiveTile) calculateCargoPaths() {

}
