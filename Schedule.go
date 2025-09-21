package main

import "slices"

type Schedule struct {
	name  string
	user  User
	stops []Stop
}

// returnt nächsten Stop
func (s *Schedule) nextStop(currentStop Stop) Stop {
	index := slices.Index(s.stops, currentStop)
	if index == len(s.stops)-1 {
		return s.stops[0]
	}
	return s.stops[index+1]
}

// --------------------------------------------------
type Stop struct {
	id   int
	goal [3]int //wird Station, bei der man sich ein Tile abholt
}
