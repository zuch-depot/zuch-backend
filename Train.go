package main

import (
	"fmt"
	"slices"
)

type Train struct {
	//irgendwie noch zusmamenfassung in einen Zug, selbstreferenz funktioniert nicht
	position    [3]int //x,y,track(1,2,3,4) ->
	goal        [3]int //nur fürs testen
	currentPath chan [3]int
	maxSpeed    int

	//
	size  int
	cargo int
}

func (t Train) move() {

}

func (t Train) recalculatePath() {

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

	visited := make(map[[3]int]Visit, 1) //ggf. als *Visit
	var toVisit []ToDo

	toVisit = append(toVisit, ToDo{t.position[0], t.position[1], t.position[2], 0, 0})
	visited[[3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}] = Visit{visited: true, gotNeighbours: true}

	succesfull := false
	steps := 0
	for len(toVisit) > 0 {

		// fmt.Println(toVisit)
		// fmt.Println("")
		// fmt.Println(visited)
		// fmt.Println("---------------------")

		//sortieren der ToDos
		// fmt.Println(toVisit)
		slices.SortFunc(toVisit, func(a, b ToDo) int {
			if a.value > b.value {
				return 1
			}
			if a.value < b.value {
				return -1
			}
			return 0
		})
		// fmt.Println("")
		// fmt.Println(toVisit)
		//Auswählen des aktuellen tiles
		visitingTile := [3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}

		visitingPathLength := toVisit[0].pathLength

		if visitingTile == t.goal {
			succesfull = true
			break
		}

		//Nachbarn bestimmen
		neighbours := neighbourTracks(visitingTile[0], visitingTile[1], visitingTile[2])
		v := visited[visitingTile]
		visited[visitingTile] = Visit{prevX: v.prevX, prevY: v.prevY, prevSub: v.prevSub, visited: v.visited, gotNeighbours: true}
		//visited[visitingTile].gotNeighbours = true

		for i := range neighbours {

			n := neighbours[i]
			if !visited[n].visited {
				visited[n] = Visit{prevX: visitingTile[0], prevY: visitingTile[1], prevSub: visitingTile[2], value: visitingPathLength + 1, visited: true, gotNeighbours: false}
			}
			if !visited[n].gotNeighbours {
				alreadyToDo := false
				for o := range toVisit {
					if [3]int{toVisit[o].x, toVisit[o].y, toVisit[o].sub} == n {
						alreadyToDo = true
					}
				}
				if !alreadyToDo {
					newCost := visitingPathLength + 1 + Abs(t.goal[0]-n[0]) + Abs(t.goal[1]-n[1])
					//fmt.Println(visitingTile[0], visitingTile[1], visitingTile[2], "|", n, Abs(t.goal[0]-n[0])-Abs(t.goal[1]-n[1]))
					toVisit = append(toVisit, ToDo{x: n[0], y: n[1], sub: n[2], pathLength: visitingPathLength + 1, value: newCost})
				}
			}
		}
		toVisit = toVisit[1:]
		steps++
	}
	if succesfull {

		var path [][3]int
		for current := t.goal; current != t.position; {
			//path <- current
			path = append(path, current)
			v := visited[current]
			current = [3]int{v.prevX, v.prevY, v.prevSub}
		}
		fmt.Println(path)

		path2 := make(chan [3]int, 300) // wird falschrum reingeladen. Braucht first in, first out
		for i := len(path); i > 0; i-- {
			path2 <- path[i-1]
		}
		t.currentPath = path2
	}

	fmt.Println("Ende Dijkstra, Anzahl Schritte:", steps)

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
