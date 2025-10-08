package main

import (
	"slices"
	"strconv"
)

func (t *Train) recalculatePath(gs *gameState) {

	type ToDo struct {
		x          int
		y          int
		sub        int
		value      int //to prioritize closness
		pathLength int
	}

	type Visit struct {
		prevX         int
		prevY         int
		prevSub       int
		value         int //path length
		visited       bool
		gotNeighbours bool //false is only looked at at least once
	}

	//Paths sind in falscher Reihenfolge
	dijkstra := func(goal [3]int) ([][3]int, [][3]int) {

		visited := make(map[[3]int]Visit, 1) //ggf. als *Visit
		var toVisit []ToDo

		toVisit = append(toVisit, ToDo{t.Waggons[0].Position[0], t.Waggons[0].Position[1], t.Waggons[0].Position[2], 0, 0})
		visited[[3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}] = Visit{visited: true, gotNeighbours: true}

		logger.Debug("Train " + t.Name + " Dijkstra Start ToDo: [" + strconv.Itoa(toVisit[0].x) + ", " +
			strconv.Itoa(toVisit[0].y) + ", " + strconv.Itoa(toVisit[0].sub) + "] Goal: [" + strconv.Itoa(goal[0]) + ", " +
			strconv.Itoa(goal[1]) + ", " + strconv.Itoa(goal[2]) + "]")

		succesfull := false
		for len(toVisit) > 0 {

			//sortieren der ToDos
			slices.SortFunc(toVisit, func(a, b ToDo) int {
				if a.value > b.value {
					return 1
				}
				if a.value < b.value {
					return -1
				}
				return 0
			})
			//Auswählen des aktuellen tiles
			visitingTile := [3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}

			//Läge des Weges zu dem Tile
			visitingPathLength := toVisit[0].pathLength

			//ist man angekommen?
			if visitingTile == goal {
				succesfull = true
				break
			}

			//Nachbarn bestimmen
			neighbours := neighbourTracks(visitingTile[0], visitingTile[1], visitingTile[2], gs)

			//Visit neu erstellen um gotNeighbours auf true zu stellen. Besser wenn das als []*Visit ist und nicht neu erstellt werden muss
			v := visited[visitingTile]
			visited[visitingTile] = Visit{prevX: v.prevX, prevY: v.prevY, prevSub: v.prevSub, visited: v.visited, gotNeighbours: true}

			for i := range neighbours {
				n := neighbours[i]

				//Nur durchkommen, wenn Signal richtig rum ist
				nTileSig := gs.Tiles[n[0]][n[1]].Signals
				vTileSig := gs.Tiles[visitingTile[0]][visitingTile[1]].Signals
				//wenn man sich sub3 anguckt und auf dem rechten Tile ein Signal steht und nicht auf sub 3 auch eins steht, dann nicht den Nachbarn wählen
				if (visitingTile[2] == 3 && n[0] > visitingTile[0] && n[2] == 1 && nTileSig[0] && !vTileSig[2]) ||
					(visitingTile[2] == 4 && n[1] > visitingTile[1] && n[2] == 2 && nTileSig[1] && !vTileSig[3]) ||
					(visitingTile[2] == 1 && n[0] < visitingTile[0] && n[2] == 3 && nTileSig[2] && !vTileSig[0]) ||
					(visitingTile[2] == 2 && n[1] < visitingTile[1] && n[2] == 4 && nTileSig[3] && !vTileSig[1]) {
					continue
				}
				//wenn 1. Waggon im selben Tile, nicht im selben Tile weiter gucken
				//wenn 1. Waggon nicht im selben Tile, nicht im anderen Tile gucken
				// gibt es einen 1. Waggon, sonst fahre frei
				if len(t.Waggons) > 1 {
					//ist der 1. Waggon im selben Tile wie Lokomotive?
					//ist der Nachbar im Tile der Lokomotive und Wagen?, dann nicht angucken, weil kann nicht befahren werden
					if (n[0] == t.Waggons[1].Position[0] && n[1] == t.Waggons[1].Position[1]) ||
						n == t.Waggons[1].Position {
						continue
					}
				}

				//war man schonmal da?
				if !visited[n].visited {
					//wenn nicht, Erstellung eines neuen Visit
					visited[n] = Visit{prevX: visitingTile[0], prevY: visitingTile[1], prevSub: visitingTile[2], value: visitingPathLength + 1, visited: true, gotNeighbours: false}
				}
				//hat man sich schonmal die Nachbarn angeguckt?, sonst
				if !visited[n].gotNeighbours {
					// gibt es das SubTile schon in der Todo?
					alreadyToDo := false
					for o := range toVisit {
						if [3]int{toVisit[o].x, toVisit[o].y, toVisit[o].sub} == n {
							alreadyToDo = true
						}
					}
					//sonst füge in ToDo ein, dass man sich den mal angucken sollte
					if !alreadyToDo {
						//optimierung nach A*
						newCost := visitingPathLength + 1 + Abs(goal[0]-n[0]) + Abs(goal[1]-n[1])
						toVisit = append(toVisit, ToDo{x: n[0], y: n[1], sub: n[2], pathLength: visitingPathLength + 1, value: newCost})
					}
				}
			}
			//rauslöschen des ersten Elementes, damit Rest nachrücken kann
			toVisit = toVisit[1:]
		}

		//hat man das Ziel gefunden?
		if succesfull {
			//rausschreiben des Weges, vom Ziel zum Start. Ebenfalls speichern, wo es Signale gibt
			var path [][3]int
			var pathSignals [][3]int
			for current := goal; current != t.Waggons[0].Position; {
				//hinzufügen des aktuell betrachteten sub Tiles in Weg List
				path = append(path, current)
				//Bestimmung, ob beim aktuellen sub Tile ein Signal ist, dann füge das hinzu
				if gs.Tiles[current[0]][current[1]].Signals[current[2]-1] {
					pathSignals = append(pathSignals, current)
				}
				//Bestimmung des nächsten zu betrachtenen sub Tile
				v := visited[current]
				current = [3]int{v.prevX, v.prevY, v.prevSub}
			}

			return path, pathSignals
		}
		return make([][3]int, 0), make([][3]int, 0)
	}

	//sucht einen Weg zu beiden Enden der Zielplattform und nimmt den kürzeren
	// (Optimierung: brich ab, wenn der Weg sicher länger als der andere ist
	// ODER paralelles Pathfinding)
	goals := t.NextStop.getGoals()
	var paths, pathsSignals [2][][3]int
	paths[0] = make([][3]int, 0)

	channelPath := [2]chan [][3]int{
		make(chan [][3]int, 1),
		make(chan [][3]int, 1),
	}
	channelPathSignals := [2]chan [][3]int{
		make(chan [][3]int, 1),
		make(chan [][3]int, 1),
	}

	sub := func(i int, goal [3]int, outPath chan<- [][3]int, outPathSignals chan<- [][3]int) {

		path, pathSignals := dijkstra(goal)
		if len(paths) == 0 {
			//testet nochmal, dieses mal wird der Zug umggedreht um zu prüfen, ob dann ein Weg zu finden ist
			logger.Debug("Teste reverse")
			t.reverseTrain()
			path, pathSignals = dijkstra(goal)
			if len(path) == 0 {
				//wenn umgedreht auch kein Weg zu finden war, wieder zurück drehen
				t.reverseTrain()
			}
		}
		outPath <- path
		outPathSignals <- pathSignals
	}

	//Start der go routinen
	for i := range goals {
		go sub(i, goals[i], channelPath[i], channelPathSignals[i])
	}
	//auslesen aus dem Buffer
	for i := range goals {
		paths[i] = <-channelPath[i]
		pathsSignals[i] = <-channelPathSignals[i]
	}

	//gibt es einen Weg?
	if len(paths[0])+len(paths[1]) == 0 {
		logger.Debug("No Path found for" + t.Name)
		return
	}
	//wenn nur ein Weg, dann der, sonst der bessere
	var i int
	if len(paths[0]) >= len(paths[1]) && len(paths[1]) > 0 {
		i = 1
	} else {
		i = 0
	}
	slices.Reverse(paths[i])
	//Hinzufügen der Tiles der Station ans Ende, damit der Zug bis nach Hinten einfährt, wenn das Ziel eine Plattform ist
	if t.NextStop.IsPlattform {
		paths[i] = append(paths[i], t.NextStop.Plattform.getPathToOpposite(goals[i])...)
	}
	t.CurrentPath = paths[i]

	t.NextGoal = paths[i][len(paths[i])-1]

	//Umdrehen Weg, damit der vom Start zum Ziel, war bis jetzt umgedreht
	slices.Reverse(pathsSignals[i])
	//Ziel wird als Signal hinzugefügt, da es (eigentlich) sich wie eins verhält
	pathsSignals[i] = append(pathsSignals[i], [][3]int{paths[i][len(paths[i])-1]}...)
	t.CurrentPathSignals = pathsSignals[i]
	logger.Debug("----------------------")

}

// func reverseTrain(train []*TrainType) {
func (t *Train) reverseTrain() {
	prevTrain := t.Waggons
	slices.Reverse(prevTrain)
	for i := range t.Waggons {
		t.Waggons[i].Position = prevTrain[i].Position
	}
}
