package main

import "slices"

type Schedule struct {
	Name  string
	User  User
	Stops []Stop
}

// returnt nächsten Stop
func (s *Schedule) nextStop(currentStop Stop) Stop {
	index := slices.Index(s.Stops, currentStop)
	if index == len(s.Stops)-1 {
		return s.Stops[0]
	}
	return s.Stops[index+1]
}

// --------------------------------------------------
type Stop struct {
	id   int
	goal [3]int //wird Station, bei der man sich ein Tile abholt
}
