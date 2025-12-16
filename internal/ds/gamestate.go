// Package df  das packet macht einfach alle datenstrukturen damit wir die nett als go doc dokumentation haben können, und damit das geht muss es auch alle funktionen die auf den datenstrukturen definiert wurde nbeinhalten
package ds

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
	"zuch-backend/internal/utils"
)

// ds = Datastructure, GS = gamestate, beides abgekürzt weil es oft vorkommen wird
type GamestateTemp struct {
	Users     []*User
	Schedules []*Schedule
	Stations  []*Station
	Tiles     [][]*Tile
	Trains    []*Train
}

type GameState struct {
	Users       []*User
	Schedules   []*Schedule
	Stations    []*Station
	Tiles       [][]*Tile
	Trains      map[int]*Train
	ActiveTiles []*ActiveTile

	LoadUnloadSpeed        int
	MinLoadUloadTicks      int
	CapacityPerStationTile int
	ConfigData             ConfigData // übergeordetes Struct, in das alles aus config.json reingeladen wird

	StationRange int
	//ALLE IDs MÜSSEN BEI 1 ANFANGEN
	CurrentTrainID      atomic.Uint64
	CurrentScheduleID   atomic.Uint64
	CurrentStopID       atomic.Uint64
	CurrentStationID    atomic.Uint64
	CurrentPlattformID  atomic.Uint64
	CurrentActiveTileID atomic.Uint64
	Ticker              *time.Ticker
	Tick                int
	IsPaused            bool

	BroadcastChannel chan WsEnvelope
	UserInputs       chan RecieveWSEnvelope
	UnPause          chan bool

	Logger *slog.Logger

	SizeX       int
	SizeY       int
	SizeSubtile int
}

type SendAbleGamestate struct {
	Users     []*User
	Schedules []*Schedule
	Stations  []*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}

func (gs *GameState) AddSchedule(name string) (*Schedule, error) {

	if name == "" {
		name = fmt.Sprint("Schedule ", gs.CurrentScheduleID.Load())
	}

	s := Schedule{Name: name, Id: int(gs.CurrentScheduleID.Load())}
	gs.CurrentScheduleID.Add(1)

	gs.Schedules = append(gs.Schedules, &s)
	return &s, nil
}

func (gs *GameState) RemoveSchedule(Id int) error {
	var err error

	for i, schedule := range gs.Schedules {
		if schedule.Id == Id {

			gs.Schedules, err = utils.RemoveElementFromSlice(gs.Schedules, i)
			return err
		}
	}

	return nil
}

// nur Erstellung einer Station, hinzufügen der Tiles muss extra gemacht werden. Id ist Standardname, standard Kapazität = 0
func (gs *GameState) AddStation(name string) (*Station, error) {

	if name == "" {
		name = fmt.Sprint("Station ", gs.CurrentStationID.Load())
	}

	gs.Stations = append(gs.Stations, &Station{Id: int(gs.CurrentStationID.Load()), Name: name, Storage: map[string]int{}, Plattforms: []*Plattform{}})
	gs.CurrentStationID.Add(1)
	return gs.Stations[len(gs.Stations)-1], nil
}

// Entfernt die Station und alle Referenzen (Stops, Plattform Tiles, etc)
func (gs *GameState) RemoveStation(Id int) error {
	var err error

	for i, station := range gs.Stations {
		if station.Id == Id {
			gs.Stations, err = utils.RemoveElementFromSlice(gs.Stations, i)
			if err != nil {
				return err
			}

			//Enfernung des Plattform Tags aller Tiles der Station, Löschung der Plattformen findet beim Löschenn des jeweils letzten Tiles statt
			for _, plattform := range station.Plattforms {
				for _, tilePos := range plattform.Tiles {
					ChangeStationTile(true, tilePos, gs)
				}
			}
			return nil
		}
	}
	return err
}
