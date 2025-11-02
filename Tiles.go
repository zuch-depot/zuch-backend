package main

import "zuch-backend/internal/ds"

// bis jetzt noch keine Level implementiert
func processActiveTiles(gs *ds.GameState) {

	for _, activeTile := range gs.ActiveTiles {
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
				emptySpaceInActiveTile := activeTile.MaxStorage - activeTile.Storage[cargoTypeToConsume]
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
							activeTile.Storage[cargoTypeToConsume] = activeTile.MaxStorage
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
					activeTile.Storage[cargoTypeProducing] -= quantityToAdd - station.AddCargo(cargoTypeProducing, quantityToAdd)
				}
			}
		}

	}
}

type Car struct {
}
