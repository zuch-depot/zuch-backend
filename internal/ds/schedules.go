package ds

import (
	"slices"
	"zuch-backend/internal/utils"
)

type Schedule struct {
	Id    int
	Name  string
	Stops []Stop
}

// einzigartig pro Schedule, egal ob gleiche Plattform oder so
// Einzige Referenz in Schedule, deshalb werden alle mit gelöscht?
type Stop struct {
	Id                int //NOTWENDIG, muss > 1 sein, == 0 ist null
	Plattform         *Plattform
	Goal              [3]int //(?Signal als) Wegpunkt
	IsPlattform       bool
	Name              string               //Name Wegpunkt
	LoadUnloadCommand [2]LoadUnloadCommand //einmal zum Laden, einmal zum entladen (0 entladen, 1 beladen)
}

// braucht keine Id, weil nur an einer Stelle gespeichert mit statischer Positionierung
type LoadUnloadCommand struct {
	//wenn Loading, wenn false, dann kurz warten und auch wenn nicht voll trotzdem fahren
	//wenn Unloading, wenn false, dann "WaitTillEmpty", also warten, bis alles entladen werden kann, oder einfach weiterfahren
	WaitTillFull bool
	Loading      bool     //wenn false, dann unloading
	CargoTypes   []string //welche Güter abgeladen/aufgeladen werden dürfen
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

// TODO errors/Validierung
func (s Stop) changeLoadCommand(cargoTypes []string, waitTillFull bool) error {
	s.LoadUnloadCommand[1] = LoadUnloadCommand{Loading: true, WaitTillFull: waitTillFull, CargoTypes: cargoTypes}
	return nil
}

// TODO errors/validierung
func (s Stop) changeUnloadCommand(cargoTypes []string, waitTillEmpty bool) error {
	s.LoadUnloadCommand[1] = LoadUnloadCommand{Loading: false, WaitTillFull: waitTillEmpty, CargoTypes: cargoTypes}
	return nil
}

// returnt nächsten Stop, wenn 0 übergeben wird der erste
func (s *Schedule) nextStop(currentStop Stop) Stop {
	index := slices.IndexFunc(s.Stops, func(stop Stop) bool {
		return stop.Id == currentStop.Id
	})
	if currentStop.Id == 0 || index == len(s.Stops)-1 {
		return s.Stops[0]
	}
	return s.Stops[index+1]
}

// Entfernt den Stop mit der passenden Id
// TODO errors
func (s *Schedule) removeStop(Id int, gs *GameState) error {
	var err error

	for i, stop := range s.Stops {
		if stop.Id == Id {
			//TODO Irgendwie die Züge aktualisieren?, dass die nicht mehr weiterfahren
			s.Stops, err = utils.RemoveElementFromSlice(s.Stops, i)
			return err
		}
	}

	return nil
}

// TODO errors
// fügt eine Station zu. Damit Waren geladen oder entladen werden sollen müssen dem Stop ein Load, bzw. Unload Befehle mit entsprechender Methode hinzugefügt werden
func (s *Schedule) AddStopStation(plattform Plattform, gs *GameState) (Stop, error) {
	var err error

	if plattform.Id == 0 {
		return Stop{}, err
	}

	s.Stops = append(s.Stops, Stop{Id: int(gs.CurrentStopID.Load()), Plattform: &plattform, IsPlattform: true})
	gs.CurrentStopID.Add(1)

	return s.Stops[len(s.Stops)-1], nil
}

// TODO errors
// fügt einen Wegpunkt hinzu. Wird angefahren, es werden aber keine Load oder Unload Befehle benötigt
func (s *Schedule) AddStopWaypoint(position [3]int, name string, gs *GameState) (Stop, error) {
	var err error

	//überprüfen koordinaten?

	s.Stops = append(s.Stops, Stop{Id: int(gs.CurrentStopID.Load()), IsPlattform: false, Goal: position, Name: name})
	gs.CurrentStopID.Add(1)

	return s.Stops[len(s.Stops)-1], err
}
