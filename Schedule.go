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

type Stop struct {
	Id          int
	Plattform   *Plattform
	Goal        [3]int //(?Signal als) Wegpunkt
	IsPlattform bool
	Name        string
}

func (s Stop) getGoals() [][3]int {
	if s.IsPlattform {
		r := s.Plattform.getFirstLast()
		return [][3]int{r[0], r[1]}
	}
	return [][3]int{s.Goal}
}

//Returnt Name des Wegpunktes oder Station + Plattform
func (s Stop) getName() string {
	if s.IsPlattform {
		return s.Plattform.station.name + s.Plattform.name
	}
	return s.Name
}
