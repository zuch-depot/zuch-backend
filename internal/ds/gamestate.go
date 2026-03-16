// Package df  das packet macht einfach alle datenstrukturen damit wir die nett als go doc dokumentation haben können, und damit das geht muss es auch alle funktionen die auf den datenstrukturen definiert wurde nbeinhalten
package ds

import (
	"log/slog"
	"sync/atomic"
	"time"
)

// ds = Datastructure, GS = gamestate, beides abgekürzt weil es oft vorkommen wird

// type GamestateTemp struct {
// 	Users     []*User
// 	Schedules []*Schedule
// 	Stations  []*Station
// 	Tiles     [][]*Tile
// 	Trains    []*Train
// }

type GameState struct {
	Users       map[string]*User // einzigartiger name
	Schedules   map[int]*Schedule
	Stations    map[int]*Station
	Tiles       [][]*Tile
	Trains      map[int]*Train
	ActiveTiles []*ActiveTile
	Plattforms  map[int]*Plattform

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

	currentTrain      *Train //nur für das einfache hinzufügen von Waggons, wird auch nicht im savegame gespeichert
	currentWaggonType string //ebenfalls weil ich faul bin und um redundanz zu vermeiden
}

// alle Attribute, die man am anfang senden möchte
type SendAbleGamestate struct {
	Users     map[string]*User // der name ist einzigartig
	Schedules map[int]*Schedule
	Stations  map[int]*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}

// alle attribute, die man speichern möchte
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
	Tick                int

	SizeX       int
	SizeY       int
	SizeSubtile int
}
