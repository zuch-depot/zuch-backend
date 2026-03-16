package ds

import (
	"fmt"
	"slices"
	"zuch-backend/internal/utils"
)

type Schedule struct {
	Id    int
	Name  string
	Stops []*Stop
}

// einzigartig pro Schedule, egal ob gleiche Plattform oder so
// Einzige Referenz in Schedule, deshalb werden alle mit gelöscht?
type Stop struct {
	Id                int //NOTWENDIG, muss > 1 sein, == 0 ist null
	Plattform         *Plattform
	Goal              [3]int //(?Signal als) Wegpunkt
	IsPlattform       bool
	Name              string               //Name Wegpunkt
	LoadUnloadCommand [2]LoadUnloadCommand //einmal zum Laden, einmal zum entladen (0 entladen, 1 beladen). Können auch jeweils leer sein, dann wird alles geladen/entladen
}

// braucht keine Id, weil nur an einer Stelle gespeichert mit statischer Positionierung
type LoadUnloadCommand struct {
	//wenn Loading, wenn false, dann kurz warten und auch wenn nicht voll trotzdem fahren
	//wenn Unloading, wenn false, dann "WaitTillEmpty", also warten, bis alles entladen werden kann, oder einfach weiterfahren
	WaitTillFull bool
	Loading      bool     //wenn false, dann unloading
	CargoTypes   []string //welche Güter abgeladen/aufgeladen werden dürfen
}

// returnt beide Enden der Plattform, wenn es eine ist, und sonst nur den Wegpunkt
func (s *Stop) getGoals(gs *GameState) [][3]int {
	if s.IsPlattform {
		r := s.Plattform.getFirstLast(gs)
		return [][3]int{r[0], r[1]}
	}
	return [][3]int{s.Goal}
}

// Returnt Name des Wegpunktes oder Station + Plattform
func (s *Stop) getName(gs *GameState) string {
	if s.IsPlattform {
		return s.Plattform.GetStation(gs).Name + " " + s.Plattform.Name
	}
	return s.Name
}

// validiert, ob es die Kategorien gibt und fügt als Command hinzu
func (s *Stop) SetLoadCommand(cargoTypes []string, waitTillFull bool, gs *GameState) error {
	validatedCategorys, err := gs.validateTrainCategories(cargoTypes)
	s.LoadUnloadCommand[1] = LoadUnloadCommand{Loading: true, WaitTillFull: waitTillFull, CargoTypes: validatedCategorys}
	return err
}

// validiert, ob es die Kategorien gibt und fügt als Command hinzu
func (s *Stop) SetUnloadCommand(cargoTypes []string, waitTillEmpty bool, gs *GameState) error {
	validatedCategorys, err := gs.validateTrainCategories(cargoTypes)
	s.LoadUnloadCommand[0] = LoadUnloadCommand{Loading: false, WaitTillFull: waitTillEmpty, CargoTypes: validatedCategorys}
	return err
}

// returnt nächsten Stop, wenn 0 übergeben wird der erste
func (s *Schedule) nextStop(currentStop *Stop) *Stop {
	if len(s.Stops) == 0 {
		return &Stop{}
	}
	index := slices.IndexFunc(s.Stops, func(stop *Stop) bool {
		return stop.Id == currentStop.Id
	})
	if currentStop.Id == 0 || index == len(s.Stops)-1 {
		return s.Stops[0]
	}
	return s.Stops[index+1]
}

// Entfernt den Stop mit der passenden Id. Löscht nicht den Schedule, wenn es der letzte Stop war. Die kann man nur mit der Funktion löschen
// index von 0 startend
// TODO errors, umändern in index
func (s *Schedule) RemoveStop(index int, gs *GameState) error {
	var err error

	if index < 0 || index > len(s.Stops) {
		return fmt.Errorf("Please provide a valid index.")
	}

	//Bei allen Zügen, die den Stop als nächsten Stop gerade haben, wird der nächste Stop ausgewählt und der Weg neu berechenet
	for _, train := range gs.Trains {
		if train.NextStop.Id == s.Stops[index].Id {
			train.NextStop = train.Schedule.nextStop(train.NextStop)
			train.recalculatePath(gs)
		}
	}

	s.Stops, err = utils.RemoveElementFromSlice(s.Stops, index)
	if err != nil {
		return err
	}

	return nil
}

// entfernt alle Stops innerhalb des Intervalls
func (s *Schedule) RemoveStops(indexStart int, indexEnde int, gs *GameState) error {

	if indexStart >= indexEnde {
		return fmt.Errorf("Please provide valid indices. The first one can not be grater than the second.")
	}

	for i := indexStart; i <= indexEnde; i++ {
		err := s.RemoveStop(i, gs)
		if err != nil {
			return err
		}
	}

	return nil
}

// fügt eine Station zu. Damit Waren geladen oder entladen werden sollen müssen dem Stop ein Load, bzw. Unload Befehle mit entsprechender Methode hinzugefügt werden.
func (s *Schedule) AddStopStation(plattform *Plattform, gs *GameState) (*Stop, error) {

	if plattform == nil || plattform.Id == 0 {
		return &Stop{}, fmt.Errorf("Error while adding a Stop to a schedule. Either Plattform is nil or the id is 0.")
	}

	s.Stops = append(s.Stops, &Stop{Id: int(gs.CurrentStopID.Load()), Plattform: plattform, IsPlattform: true})
	gs.CurrentStopID.Add(1)

	return s.Stops[len(s.Stops)-1], nil
}

// fügt einen Wegpunkt hinzu. Wird angefahren, es werden aber keine Load oder Unload Befehle benötigt. Id ist Standardname
func (s *Schedule) AddStopWaypoint(position [3]int, name string, gs *GameState) (*Stop, error) {

	err := utils.CheckName(name)
	if err != nil {
		return nil, err
	}

	//überprüfen koordinaten
	_, err = gs.GetTile(position[0], position[1])
	if err != nil || position[2] < 1 || position[2] > 4 {
		return &Stop{}, fmt.Errorf("Error while adding a Waypoint to a schedule. The coordinates are invalid.")
	}

	if name == "" {
		name = fmt.Sprint("Wegpunkt ", gs.CurrentStopID.Load())
	}

	s.Stops = append(s.Stops, &Stop{Id: int(gs.CurrentStopID.Load()), IsPlattform: false, Goal: position, Name: name})
	gs.CurrentStopID.Add(1)

	return s.Stops[len(s.Stops)-1], err
}

// überprüft utils.CeckName und setzt name
func (s *Schedule) Rename(name string) error {

	err := utils.CheckName(name)
	if err != nil {
		return err
	}

	s.Name = name

	return nil
}

// nimmt das element vom start index und macht, dass dieses den Zielindex hat
func (s *Schedule) ChangeSquence(indexStart int, indexGoal int) error {

	if indexStart < 0 || indexStart >= len(s.Stops) || indexGoal < 0 || indexGoal >= len(s.Stops) {
		return fmt.Errorf("Please provide valid indices.")
	}

	stop := s.Stops[indexStart]

	var err error
	s.Stops, err = utils.RemoveElementFromSlice(s.Stops, indexStart)
	if err != nil {
		return err
	}

	s.Stops = slices.Insert(s.Stops, indexGoal, stop)

	return nil
}
