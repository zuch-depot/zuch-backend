package main

import (
	"fmt"
	"slices"
)

type Train struct {
	Waggons            []TrainType //Alle müssen nebeneinander spawnen
	Schedule           Schedule
	NextStop           Stop     //nur fürs testen
	currentPath        [][3]int //neu berechnen bei laden
	currentPathSignals [][3]int
	Name               string
}

type TrainType struct {
	position [3]int //x,y,track(1,2,3,4) ->
	maxSpeed int
	//
	size  int
	cargo int
}

// aktuell wählt er automatisch den nächsten Stop aus, wenn das Pathfinding nicht funktioniert hat
// für 2 Wege Signale muss geprüft werden, ob nicht schon ein Zug zum Signal auf der anderen Seite fährt
func (t *Train) move() {
	newGenNoSignal := false

	//Auswahl des nächsten Stops wenn man am Ziel angekommen ist (oder das Pathfinding nicht funktioniert hat)
	if len(t.currentPath) == 0 {
		t.NextStop = t.Schedule.nextStop(t.NextStop)
		fmt.Println("Next Stop:", t.NextStop.Plattform.station.name, t.NextStop.Plattform.name)
		t.recalculatePath()
		newGenNoSignal = true
		//wenn das Pathfinding (immer noch) nicht funktioniert hat
		if len(t.currentPath) == 0 {
			return
		}
	}

	path := t.currentPath
	signals := t.currentPathSignals

	if newGenNoSignal && len(signals) > 1 {
		newGenNoSignal = false
	}
	//ist man bei einem Signal oder wurde neu generiert?
	// wenn es kein nächstes Signal gibt, wird bis zum Ziel geguckt
	// -----------> Ähnliche logik muss irgendwo Signale auf rot/grün/?gelb schalten
	// (vielleicht der Zug, wenn er sich merkt, bei welchem Signal er war und das umschaltet, wenn er aus block rausgefahren ist.
	// wird dann überschrieben, wenn der nächste zug nicht weiterfahren kann)
	// fmt.Println("newGenNoSignal", newGenNoSignal, "Signale:", signals, "Weg:", path)

	if newGenNoSignal || len(signals) > 1 && t.Waggons[0].position == signals[0] {
		//gucken, ob bis zum nächsten Signal alle Tiles unblocked sind, sonst fahre nicht weiter
		// (es wird immer auch das letzte Tile überprüft, da man über ein sub tile ohne signal fahren muss, um zu einem zu kommen)
		// --> wichtig für Stationen, immer letzte Subtile ansteuern
		for i := 0; (newGenNoSignal && path[i] != signals[0]) ||
			!newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			if tiles[path[i][0]][path[i][1]].IsBlocked {
				fmt.Println("Zug", t.Name, ": Blocked Tile found:", path[i], ". Waiting")
				return
			}
		}
		//da nichts geblocked war, blockt dieser Zug jetzt die Strecke zum nächsten Signal
		for i := 0; newGenNoSignal && path[i] != signals[0] || !newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			tiles[path[i][0]][path[i][1]].IsBlocked = true
		}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.currentPathSignals = t.currentPathSignals[1:]
	}

	//entblocken des letzten Tiles, wenn letzter Waggon sich rausbewegt (x oder y vom letzten unterschiedlich ist zum 2. letzten)
	if len(t.Waggons) == 1 ||
		(t.Waggons[len(t.Waggons)-1].position[0] != t.Waggons[len(t.Waggons)-2].position[0] ||
			t.Waggons[len(t.Waggons)-1].position[1] != t.Waggons[len(t.Waggons)-2].position[1]) {
		tiles[t.Waggons[len(t.Waggons)-1].position[0]][t.Waggons[len(t.Waggons)-1].position[1]].IsBlocked = false
	}

	//Bewegung der Waggons
	for i := len(t.Waggons); i > 1; i-- {
		t.Waggons[i-1].position = t.Waggons[i-2].position
	}

	//Bewegung der Lokomotive
	t.Waggons[0].position = t.currentPath[0]
	//rausschmeißen des Tiles, wo die Lok sich hinbewegt hat aus der Queue
	t.currentPath = t.currentPath[1:]
}

func (t *Train) recalculatePath() {

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
	dijkstra := func(goal [3]int, prevLength int) ([][3]int, [][3]int) {

		visited := make(map[[3]int]Visit, 1) //ggf. als *Visit
		var toVisit []ToDo

		toVisit = append(toVisit, ToDo{t.Waggons[0].position[0], t.Waggons[0].position[1], t.Waggons[0].position[2], 0, 0})
		visited[[3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}] = Visit{visited: true, gotNeighbours: true}

		fmt.Println("Train", t.Name, "Dijkstra Start ToDo:", toVisit[0], "Goal:", goal)

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
			neighbours := neighbourTracks(visitingTile[0], visitingTile[1], visitingTile[2])

			//Visit neu erstellen um gotNeighbours auf true zu stellen. Besser wenn das als []*Visit ist und nicht neu erstellt werden muss
			v := visited[visitingTile]
			visited[visitingTile] = Visit{prevX: v.prevX, prevY: v.prevY, prevSub: v.prevSub, visited: v.visited, gotNeighbours: true}

			for i := range neighbours {
				n := neighbours[i]

				//Nur durchkommen, wenn Signal richtig rum ist
				nTileSig := tiles[n[0]][n[1]].Signals
				vTileSig := tiles[visitingTile[0]][visitingTile[1]].Signals
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
					if (n[0] == t.Waggons[1].position[0] && n[1] == t.Waggons[1].position[1]) ||
						n == t.Waggons[1].position {
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
			for current := goal; current != t.Waggons[0].position; {
				//hinzufügen des aktuell betrachteten sub Tiles in Weg List
				path = append(path, current)
				//Bestimmung, ob beim aktuellen sub Tile ein Signal ist, dann füge das hinzu
				if tiles[current[0]][current[1]].Signals[current[2]-1] {
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
	goals := t.NextStop.Plattform.getFirstLast()
	var paths, pathsSignals [2][][3]int
	paths[0] = make([][3]int, 0)
	for i := range 2 {
		paths[i], pathsSignals[i] = dijkstra(goals[i], len(paths[0]))
		if len(paths[i]) == 0 {
			//testet nochmal, dieses mal wird der Zug umggedreht um zu prüfen, ob dann ein Weg zu finden ist
			fmt.Println("Teste reverse")
			t.reverseTrain()
			paths[i], pathsSignals[i] = dijkstra(goals[i], len(paths[0]))
			if len(paths[i]) == 0 {
				//wenn umgedreht auch kein Weg zu finden war, wieder zurück drehen
				t.reverseTrain()
			}
		}
	}
	if len(paths[0])+len(paths[1]) == 0 {
		fmt.Println("No Path found for", t.Name)
		return
	}
	var i int
	if len(paths[0]) >= len(paths[1]) {
		i = 1
	} else {
		i = 0
	}
	slices.Reverse(paths[i])
	//Hinzufügen der Tiles der Station ans Ende, damit der Zug bis nach Hinten einfährt
	paths[i] = append(paths[i], t.NextStop.Plattform.getPathToOpposite(goals[i])...)
	t.currentPath = paths[i]

	//Umdrehen Weg, damit der vom Start zum Ziel, war bis jetzt umgedreht
	slices.Reverse(pathsSignals[i])
	//Ziel wird als Signal hinzugefügt, da es (eigentlich) sich wie eins verhält
	pathsSignals[i] = append(pathsSignals[i], [][3]int{paths[i][len(paths[i])-1]}...)
	t.currentPathSignals = pathsSignals[i]
	fmt.Println("----------------------")

}

// func reverseTrain(train []*TrainType) {
func (t *Train) reverseTrain() {
	prevTrain := t.Waggons
	slices.Reverse(prevTrain)
	for i := range t.Waggons {
		t.Waggons[i].position = prevTrain[i].position
	}
}

/* erst auf dauerhaft blocked prüfen
*to visit (außer, da wo man hergekommen ist):
*	1 [x][y][2,3,4], [x-1][y][3]
*	2 [x][y][1,3,4], [x][y+1][4]
*	3 [x][y][1,2,4], [x+1][y][1]
*	4 [x][y][1,2,3], [x][y+1][2]
 */
// verified
func neighbourTracks(x int, y int, sub int) [][3]int {

	var r [][3]int

	appending := func(a [3]int) {
		for i := range 3 {
			o := a[i]
			if tiles[x][y].Tracks[o-1] {
				r = append(r, [3]int{x, y, o})
			}
		}
	}

	switch sub {
	case 1:
		if x > 0 {
			if tiles[x-1][y].Tracks[2] {
				r = append(r, [3]int{x - 1, y, 3})
			}
		}
		appending([3]int{2, 3, 4})
	case 2:
		if y > 0 {
			if tiles[x][y-1].Tracks[3] {
				r = append(r, [3]int{x, y - 1, 4})
			}
		}
		appending([3]int{1, 3, 4})
	case 3:
		if x != len(tiles)-1 {
			if tiles[x+1][y].Tracks[0] {
				r = append(r, [3]int{x + 1, y, 1})
			}
		}
		appending([3]int{1, 2, 4})
	case 4:
		if y != len(tiles[0])-1 {
			if tiles[x][y+1].Tracks[1] {
				r = append(r, [3]int{x, y + 1, 2})
			}
		}
		appending([3]int{1, 2, 3})
	}
	return r
}

// --------------------------------------------------
type CargoType int

const (
	Coal CargoType = iota //all following are increasing int
	Iron
	Potatos
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
