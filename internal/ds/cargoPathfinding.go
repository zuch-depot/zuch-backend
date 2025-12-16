package ds

import (
	"fmt"
	"slices"
)

// uhhh dunno was genau das hier ist
// offensichtlich nochmal Dijkstra
func (gs *GameState) CalculateCargoPaths() {

	//

	for _, startActiveTile := range gs.ActiveTiles {
		prodTypes := startActiveTile.getProductionCategorys()

		//Menge könnte man auch berücksichtigen
		for prodType := range prodTypes {
			var prodTypePaths [][]*cargoPathElement //Alle Paths, die für einen Produktionstypen gefunden wurden

			//für jeden Produktionstypen alle möglichen Ziele suchen
			for _, targetActiveTile := range gs.ActiveTiles {

				consTypes := targetActiveTile.getConsumptionCategorys() //alle Verbrauchstypen des Ziel Tiles
				for consType := range consTypes {

					if consType == prodType {
						var paths [][]*cargoPathElement

						//bestimmung aller pfade
						for _, station := range startActiveTile.Stations {
							temp, _ := gs.cargoPathfinding(station, targetActiveTile, prodType)
							paths = append(paths, temp)
						}

						//auswahl des kürzesten
						shortestPath := paths[0]
						lenShortest := len(paths[0])
						for _, path := range paths {
							tempLen := len(path)
							if tempLen < lenShortest {
								shortestPath = path
							}
						}

						//print
						// fmt.Println("pathfinding for:", startActiveTile.Name, "Prodtype:", prodType, "to:", targetActiveTile.Name, "ConsType:", consType)
						// fmt.Println("Länge", len(shortestPath))
						// for _, element := range shortestPath {
						// 	fmt.Println(element.toString())
						// }

						prodTypePaths = append(prodTypePaths, shortestPath)

					}
				}
			}

			//Verteilen auf umliegende Stationen
			numberStations := len(prodTypePaths)
			for _, path := range prodTypePaths {
				station := path[0].startStation
				quantityToAdd := startActiveTile.Storage[prodType] / numberStations
				startActiveTile.Storage[prodType] -= quantityToAdd - station.AddCargo(prodType, quantityToAdd)
			}

		}
	}
}

// Weg ist in richtiger Reihenfolge
func (gs *GameState) cargoPathfinding(start *Station, target *ActiveTile, cargoType string) ([]*cargoPathElement, error) {

	// Item der Merkliste, die die noch abgearbeitet werden müssen
	type pathfindingElement struct {
		pathLength       int
		station          *Station
		cargoPathElement cargoPathElement //Weg, der zu der Station führt (start ist prevStation, target = station)
	}

	//ToDoListe initialisieren ---> funktioniert so nicht
	var toDoStations []pathfindingElement
	toDoStations = append(toDoStations, pathfindingElement{pathLength: 0, station: start})

	visitedStations := make(map[*Station]*pathfindingElement)

	succesfull := false
	var goal *pathfindingElement
	for len(toDoStations) > 0 {

		//sortieren der ToDos
		slices.SortFunc(toDoStations, func(a, b pathfindingElement) int {
			if a.pathLength > b.pathLength {
				return 1
			}
			if a.pathLength < b.pathLength {
				return -1
			}
			return 0
		})
		currentToDo := toDoStations[0]

		//das aktuell betrachtete Element den Betrachteten hinzufügen. Man war noch ganz sicher nicht da
		visitedStations[currentToDo.station] = &currentToDo

		//ist man angekommen?
		for _, targetStation := range target.Stations {
			if currentToDo.station.Id == targetStation.Id {
				succesfull = true
				break
			}
		}
		if succesfull {
			goal = &currentToDo
			break
		}

		//Nachbarn bestimmen
		neighbours := gs.GetAvaliableStation(currentToDo.station, cargoType)

		for _, neighbour := range neighbours {
			//wenn man den Nachbarn schonmal als Nachbarn hatte, dann kann der neue Weg nichtmehr besser sein, da nur Entfernung betrachtet wird
			//mir ist egal, dass theoretisch bei der Entfernung die Plattformen betrachtet werden und es daher trotzdem zu kleinen Unterschieden kommen kann
			alreadyVisited := false
			//hat man schon vor dahin zu gehen
			for _, element := range toDoStations {
				if element.station == neighbour.targetStation {
					alreadyVisited = true
				}
			}
			//war man vielleicht auch schon da?
			if alreadyVisited && visitedStations[neighbour.targetStation] != nil {
				continue
			}

			toDoStations = append(toDoStations, pathfindingElement{
				pathLength:       neighbour.pathLength + currentToDo.pathLength,
				station:          neighbour.targetStation,
				cargoPathElement: neighbour,
			})

		}

		//entfernen des ersten elementes
		toDoStations = toDoStations[1:]
	}

	if succesfull {
		//Weg rekonstruieren
		var cargoPath []*cargoPathElement
		for current := goal.cargoPathElement; ; current = visitedStations[current.startStation].cargoPathElement {

			cargoPath = append(cargoPath, &current)
			//am start angekommen?
			if start.Id == current.startStation.Id {
				slices.Reverse(cargoPath)
				return cargoPath, nil
			}

		}
	}

	return []*cargoPathElement{}, nil
}

// eine "Kante" fürs CargoPathfinding
type cargoPathElement struct {
	startStation  *Station
	targetStation *Station
	schedule      *Schedule
	startStop     *Stop
	targetStop    *Stop
	pathLength    int
}

func (c *cargoPathElement) toString() string {
	return fmt.Sprint("From ", c.startStation.Name, " to ", c.targetStation.Name, "  schedule ", c.schedule.Name, " length: ", c.pathLength)
}

// suche in den Schedules nach allen verfügbaren erreichbaren Stationen für den Typen
// bei return: 0 ist start, 1 ist Ziel
func (gs *GameState) GetAvaliableStation(startStation *Station, cargoType string) []cargoPathElement {

	//gehe alle Stops aller Schedules durch, die entweder keinen LoadCommand haben oder einen haben, der den richtigen Typen hat,
	//  und suche nach allen Plattformen der Startstation.
	//Wenn eine Plattform der Startstation gefunden wurde,
	// dann nehme alle anderen Stationen, die entweder keinen UnloadCommand haben oder einen haben, der den richtigen Typen hat in dem Schedule in der Liste auf

	var avaliableStations []cargoPathElement

	for _, schedule := range gs.Schedules {
		var startStopFound bool // wurde ein geeigneter Stop der Startstation gefunden?
		var startStop *Stop

		// Ist die Startstation in dem Schedule als Stop enthalten mit einem passenden LoadCommand?
		for _, stop := range schedule.Stops {
			//ist es eine Plattform der Startstation?
			if stop.IsPlattform && stop.Plattform.isPlattfromFromStation(startStation) {

				//hat die Plattform bei dem Stop einen LoadCommand für den Typen oder keinen LoadCommand?
				loadTypes := stop.LoadUnloadCommand[1].CargoTypes

				if len(loadTypes) == 0 {
					startStopFound = true
				} else {
					for _, loadType := range loadTypes {
						if loadType == cargoType {
							startStopFound = true
							break
						}
					}
				}
			}
			if startStopFound {
				startStop = &stop
				break
			}
		}

		if startStopFound {
			var targetStop *Stop
			var targetStopFound bool
			//suche alle anderen Stationen in dem Schedule mit passenden UnloadCommands
			for _, stop := range schedule.Stops {

				if stop.IsPlattform && !stop.Plattform.isPlattfromFromStation(startStation) {
					//hat die Plattform bei dem Stop einen UnloadCommand für den Typen oder keinen UnloadCommand?
					//dann füge die Station der Liste hinzu
					unloadTypes := stop.LoadUnloadCommand[0].CargoTypes
					if len(unloadTypes) == 0 {
						targetStopFound = true
					} else {
						for _, unloadType := range unloadTypes {
							if unloadType == cargoType {
								targetStopFound = true
								break
							}
						}
					}
					if targetStopFound {
						targetStop = &stop
					}
				}
			}
			if targetStopFound {
				//finden der Entfernung der Plattformen mit virtuellem Zug, damit das Pathfinding benutzt werden kann
				tempTrain := Train{
					Waggons:  []*Waggon{&Waggon{Position: startStop.Plattform.GetFirstLast(gs)[0]}}, //theoretisch sollte mittleres Tile oder einmal beide genommen werden
					Schedule: schedule,
					Id:       -1,
					NextStop: *targetStop,
				}
				tempTrain.RecalculatePath(gs)
				//wenn es keinen Weg gibt, dann kann auch nichts transportiert werden
				if len(tempTrain.CurrentPath) == 0 {
					continue
				}

				//
				temp := cargoPathElement{
					startStation:  startStation,
					targetStation: targetStop.Plattform.GetStation(gs),
					startStop:     startStop,
					targetStop:    targetStop,
					schedule:      schedule,
					pathLength:    len(tempTrain.CurrentPath),
				}
				avaliableStations = append(avaliableStations, temp)
			}
		}
	}

	return avaliableStations
}
