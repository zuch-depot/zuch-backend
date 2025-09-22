package main

type Station struct {
	Plattforms []Plattform
}

// --------------------------------------------------
type Plattform struct {
	Tiles [][3]int
}

//immer zum nähesten Pathfinden und dann doch bis ans Ende fahren
