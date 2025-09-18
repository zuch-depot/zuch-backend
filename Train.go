package main

type Train struct {
	//irgendwie noch zusmamenfassung in einen Zug, selbstreferenz funktioniert nicht
	tilePosition [3]int //x,y,track
	tileGoal     [3]int //nur fürs testen
	currentPath  chan [3]int
	maxSpeed     int

	//
	size  int
	cargo int
}

func (t Train) move() {

}

func (t Train) recalculatePath() {
	t.currentPath = make(chan [3]int, 300) //

	// var visitedTiles map[[3]int][4]int //[4] == vorherhigX, vorherig Y, vorherigSub, Strecke

	/* erst auf dauerhaft blocked prüfen
	*to visit (außer, da wo man hergekommen ist):
	*	1 [x][y][2,3,4], [x-1][y][3]
	*	2 [x][y][1,3,4], [x][y+1][4]
	*	3 [x][y][1,2,4], [x+1][y][1]
	*	4 [x][y][1,2,3], [x][y+1][2]
	 */

}

// --------------------------------------------------
type CargoType int

const (
	Coal CargoType = iota //all following are increasing int
	Iron
	Potatos
)
