package ds

import "slices"

type Schedule struct {
	Name  string
	User  *User
	Stops []Stop
}

type Stop struct {
	Id                int //NOTWENDIG
	Plattform         *Plattform
	Goal              [3]int //(?Signal als) Wegpunkt
	IsPlattform       bool
	Name              string               //Name Wegpunkt
	LoadUnloadCommand [2]LoadUnloadCommand //einmal zum Laden, einmal zum entladen (0 entladen, 1 beladen)
}

type LoadUnloadCommand struct {
	//wenn Loading, wenn false, dann kurz warten und auch wenn nicht voll trotzdem fahren
	//wenn Unloading, wenn false, dann "WaitTillEmpty", also warten, bis alles entladen werden kann, oder einfach weiterfahren
	WaitTillFull bool
	Loading      bool     //wenn false, dann unloading
	CargoType    []string //welche Güter abgeladen/aufgeladen werden dürfen
}

func (s Stop) getGoals() [][3]int {
	if s.IsPlattform {
		r := s.Plattform.GetFirstLast()
		return [][3]int{r[0], r[1]}
	}
	return [][3]int{s.Goal}
}

// Returnt Name des Wegpunktes oder Station + Plattform
func (s Stop) getName() string {
	if s.IsPlattform {
		return s.Plattform.Station.Name + " " + s.Plattform.Name
	}
	return s.Name
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
