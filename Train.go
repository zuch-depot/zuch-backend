package main

import (
	"fmt"
	"slices"
)

type Train struct {
	train       []TrainType //Alle müssen nebeneinander spawnen
	schedule    Schedule
	nextStop    Stop //nur fürs testen
	currentPath [][3]int
	name        string
}

type TrainType struct {
	position [3]int //x,y,track(1,2,3,4) ->
	maxSpeed int
	//
	size  int
	cargo int
}

// aktuell wählt er automatisch den nächsten Stop aus, wenn das Pathfinding nicht funktioniert hat
func (t *Train) move() {
	if len(t.currentPath) == 0 {
		t.nextStop = t.schedule.nextStop(t.nextStop)
		fmt.Println("Next Stop:", t.nextStop.goal)
		t.recalculatePath()
	}
	if len(t.currentPath) == 0 {
		return
	}

	for i := len(t.train); i > 1; i-- {
		t.train[i-1].position = t.train[i-2].position
	}

	t.train[0].position = t.currentPath[0]
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

	dijkstra := func() bool {

		visited := make(map[[3]int]Visit, 1) //ggf. als *Visit
		var toVisit []ToDo

		toVisit = append(toVisit, ToDo{t.train[0].position[0], t.train[0].position[1], t.train[0].position[2], 0, 0})
		visited[[3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}] = Visit{visited: true, gotNeighbours: true}

		fmt.Println("Dijkstra Start ToDo:", toVisit[0])

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
			if visitingTile == t.nextStop.goal {
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
				nTileSig := tiles[n[0]][n[1]].signals
				vTileSig := tiles[visitingTile[0]][visitingTile[1]].signals
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
				if len(t.train) > 1 {
					//ist der 1. Waggon im selben Tile wie Lokomotive?
					//ist der Nachbar im Tile der Lokomotive und Wagen?, dann nicht angucken, weil kann nicht befahren werden
					if n[0] == t.train[1].position[0] && n[1] == t.train[1].position[1] {
						fmt.Println("Angeguckt", visitingTile, "Skip Nachbar:", n)
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
						newCost := visitingPathLength + 1 + Abs(t.nextStop.goal[0]-n[0]) + Abs(t.nextStop.goal[1]-n[1])
						toVisit = append(toVisit, ToDo{x: n[0], y: n[1], sub: n[2], pathLength: visitingPathLength + 1, value: newCost})
					}
				}
			}
			//rauslöschen des ersten Elementes, damit Rest nachrücken kann
			toVisit = toVisit[1:]
		}

		//hat man das Ziel gefunden?
		if succesfull {
			//rausschreibend es Weges, vom Ziel zum Start
			var path [][3]int
			for current := t.nextStop.goal; current != t.train[0].position; {
				path = append(path, current)
				v := visited[current]
				current = [3]int{v.prevX, v.prevY, v.prevSub}
			}

			//Umdrehen Weg, damit der vom Start zum Ziel
			slices.Reverse(path)
			fmt.Println(path)
			t.currentPath = path
			return true
		}
		return false
	}

	if !dijkstra() {
		//testet nochmal, dieses mal wird der Zug umggedreht um zu prüfen, ob dann ein Weg zu finden ist
		t.reverseTrain()
		if !dijkstra() {
			fmt.Println("No Path found")
		}
	}
	fmt.Println("----------------------")

}

// func reverseTrain(train []*TrainType) {
func (t *Train) reverseTrain() {
	prevTrain := t.train
	slices.Reverse(prevTrain)
	for i := range t.train {
		t.train[i].position = prevTrain[i].position
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
			if tiles[x][y].tracks[o-1] {
				r = append(r, [3]int{x, y, o})
			}
		}
	}

	switch sub {
	case 1:
		if x > 0 {
			if tiles[x-1][y].tracks[2] {
				r = append(r, [3]int{x - 1, y, 3})
			}
		}
		appending([3]int{2, 3, 4})
	case 2:
		if y > 0 {
			if tiles[x][y-1].tracks[3] {
				r = append(r, [3]int{x, y - 1, 4})
			}
		}
		appending([3]int{1, 3, 4})
	case 3:
		if x != len(tiles)-1 {
			if tiles[x+1][y].tracks[0] {
				r = append(r, [3]int{x + 1, y, 1})
			}
		}
		appending([3]int{1, 2, 4})
	case 4:
		if y != len(tiles[0])-1 {
			if tiles[x][y+1].tracks[1] {
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
