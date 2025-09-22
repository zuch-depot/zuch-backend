package main

import "slices"

type Schedule struct {
	Name  string
	User  User
	Stops []Stop
}

// returnt nächsten Stop
func (s *Schedule) nextStop(currentStop Stop) Stop {
	index := slices.IndexFunc(s.Stops, func(stop Stop) bool {
		return stop.Id == currentStop.Id
	})
	if index == len(s.Stops)-1 {
		return s.Stops[0]
	}
	return s.Stops[index+1]
}

// --------------------------------------------------
type Stop struct {
	Id        int
	Plattform *Plattform
}
