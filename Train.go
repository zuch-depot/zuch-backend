package main

import "fmt"

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
		x     int
		y     int
		sub   int
		value int
	}

	type Visit struct {
		prevX         int
		prevY         int
		prevSub       int
		value         int
		visited       bool
		gotNeighbours bool //false is only looked at at least once
	}

	visited := make(map[[3]int]Visit, 1) //ggf. als *Visit
	var toVisit []ToDo                   //[7] == x, y, sub, vorherigeX, vorherigeY, vorherigeSub, Wert

	/* erst auf dauerhaft blocked prüfen
	*to visit (außer, da wo man hergekommen ist):
	*	1 [x][y][2,3,4], [x-1][y][3]
	*	2 [x][y][1,3,4], [x][y+1][4]
	*	3 [x][y][1,2,4], [x+1][y][1]
	*	4 [x][y][1,2,3], [x][y+1][2]
	 */
	toVisit = append(toVisit, ToDo{t.position[0], t.position[1], t.position[2], 0})
	visited[[3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}] = Visit{visited: true, gotNeighbours: true}

	//max := 5
	succesfull := false
	for len(toVisit) > 0 { //&& max > 0 {

		// fmt.Println(toVisit)
		// fmt.Println("")
		// fmt.Println(visited)
		// fmt.Println("---------------------")

		visitingTile := [3]int{toVisit[0].x, toVisit[0].y, toVisit[0].sub}

		visitingValue := toVisit[0].value

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
			// fmt.Println("Nachbar", i, ":", n, "visited:", visited[n].visited, "gotNeighbours", visited[n].gotNeighbours)
			if !visited[n].visited {
				visited[n] = Visit{prevX: visitingTile[0], prevY: visitingTile[1], prevSub: visitingTile[2], value: visitingValue + 1, visited: true, gotNeighbours: false}
			}
			if !visited[n].gotNeighbours {
				alreadyToDo := false
				for o := range toVisit {
					if [3]int{toVisit[o].x, toVisit[o].y, toVisit[o].sub} == n {
						alreadyToDo = true
					}
				}
				if !alreadyToDo {
					toVisit = append(toVisit, ToDo{x: n[0], y: n[1], sub: n[2], value: visitingValue + 1})
				}
			}

			/*
				if visited[n][3] > visitingValue+1 || visited[n][3] != 0 {
					visited[n] = [4]int{visitingTile[0], visitingTile[1], visitingTile[2], visitingValue + 1}
				}
				if visited[n][3] != 0 {
					toVisit = append(toVisit, [7]int{n[0], n[1], n[2], visitingTile[0], visitingTile[1], visitingTile[2], visitingValue + 1})
				}*/
		}

		toVisit = toVisit[1:]
		//max--
	}
	if succesfull {
		//path := make(chan [3]int, 300) // wird falschrum reingeladen. Braucht first in, first out
		var path [][3]int
		for current := t.goal; current != t.position; {
			//path <- current
			path = append(path, current)
			v := visited[current]
			current = [3]int{v.prevX, v.prevY, v.prevSub}
		}
		//t.currentPath = path
		fmt.Print(path)
	}

	fmt.Println("Ende Dijkstra")

}

//verified
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
