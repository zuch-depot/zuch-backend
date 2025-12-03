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
	Users       []*User
	Schedules   []*Schedule
	Stations    []*Station
	Tiles       [][]*Tile
	Trains      map[int]*Train
	ActiveTiles []*ActiveTile

	LoadUnloadSpeed   int
	MinLoadUloadTicks int
	ConfigData        ConfigData // übergeordetes Struct, in das alles aus config.json reingeladen wird

	StationRange   int
	CurrentTrainID atomic.Uint64
	Ticker         *time.Ticker
	Tick           int
	IsPaused       bool

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
