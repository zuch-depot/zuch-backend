// Package df  das packet macht einfach alle datenstrukturen damit wir die nett als go doc dokumentation haben können, und damit das geht muss es auch alle funktionen die auf den datenstrukturen definiert wurde nbeinhalten
package ds

import (
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
	Stations    map[int]*Station
	Tiles       [][]*Tile
	Trains      map[int]*Train
	ActiveTiles []*ActiveTile

	LoadUnloadSpeed        int
	MinLoadUloadTicks      int
	CapacityPerStationTile int
	ConfigData             ConfigData // übergeordetes Struct, in das alles aus config.json reingeladen wird

	StationRange int
	//ALLE IDs MÜSSEN BEI 1 ANFANGEN
	CurrentTrainID     atomic.Uint64
	CurrentScheduleID  atomic.Uint64
	CurrentStopID      atomic.Uint64
	CurrentStationID   atomic.Uint64
	CurrentPlattformID atomic.Uint64
	Ticker             *time.Ticker
	Tick               int
	IsPaused           bool

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
	Stations  map[int]*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}

func (gs *GameState) AddSchedule(name string) (*Schedule, error) {
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
