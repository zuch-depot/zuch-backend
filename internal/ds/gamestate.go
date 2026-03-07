// Package df  das packet macht einfach alle datenstrukturen damit wir die nett als go doc dokumentation haben können, und damit das geht muss es auch alle funktionen die auf den datenstrukturen definiert wurde nbeinhalten
package ds

import (
	"log/slog"
	"sync/atomic"
	"time"
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
	Users       map[string]*User // einzigartiger name
	Schedules   map[int]*Schedule
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
	Users     map[string]*User // der name ist einzigartig
	Schedules map[int]*Schedule
	Stations  map[int]*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}

type SaveAbleGamestate struct {
	Users       map[string]*User // einzigartiger name
	Schedules   map[int]*Schedule
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
	CurrentTrainID      int
	CurrentScheduleID   int
	CurrentStopID       int
	CurrentStationID    int
	CurrentPlattformID  int
	CurrentActiveTileID int
	//Ticker              *time.Ticker
	Tick int
	//IsPaused            bool

	// BroadcastChannel chan WsEnvelope
	// UserInputs       chan RecieveWSEnvelope
	// UnPause          chan bool

	// Logger *slog.Logger

	SizeX       int
	SizeY       int
	SizeSubtile int
}
