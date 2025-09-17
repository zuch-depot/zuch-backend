package main

type Train struct {
}

func (t Train) move() {

}

// --------------------------------------------------
type TrainPart struct {
	//followingTrainPart TrainPart -> illegal
	maxSpeed int
}

// --------------------------------------------------
type Wagon struct {
	TrainPart
	size  int
	cargo int
}

// --------------------------------------------------
type Lokomotive struct {
	TrainPart
	performance int
}

// --------------------------------------------------
type CargoType int

const (
	Coal CargoType = iota //all following are increasing int
	Iron
	Potatos
)
