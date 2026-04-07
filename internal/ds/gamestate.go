// Package df  das packet macht einfach alle datenstrukturen damit wir die nett als go doc dokumentation haben können, und damit das geht muss es auch alle funktionen die auf den datenstrukturen definiert wurde nbeinhalten
package ds

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Root-Struktur für config.json
type ConfigData struct {
	TrainCategories        map[string][]string           `json:"Train Categories"`
	ActiveTileCategories   map[string]ActiveTileCategory `json:"Aktive Tiles"`
	Port                   int
	TicksMilisec           int
	SizeX                  int
	SizeY                  int
	SaveLocation           string
	SaveCompressed         bool
	LOADCOMPRESSED         bool // wird das überhaupt benutzt?
	LoadUnloadSpeed        int
	MinLoadUloadTicks      int
	StationRange           int
	CapacityPerStationTile int
	Ids                    ids
	WaggonTypes            map[string]*WaggonType
}

// WICHTIG: NUR fürs speichern und laden, sonst die atmic nehmen!!
type ids struct {
	CurrentTrainID      int
	CurrentScheduleID   int
	CurrentStopID       int
	CurrentStationID    int
	CurrentPlattformID  int
	CurrentActiveTileID int
}

type GameState struct {
	Mutex sync.Mutex // sorgt dafür, dass hier nur einer gleichzeitig herumspielen kann

	Users       map[string]*User // einzigartiger name
	Schedules   map[int]*Schedule
	Stations    map[int]*Station
	Tiles       [][]*Tile
	Trains      map[int]*Train
	ActiveTiles map[int]*ActiveTile

	ConfigData ConfigData // übergeordetes Struct, in das alles aus config.json reingeladen wird

	// ALLE IDs MÜSSEN BEI 1 ANFANGEN
	CurrentTrainID      atomic.Uint64 `json:"-"`
	CurrentScheduleID   atomic.Uint64 `json:"-"`
	CurrentStopID       atomic.Uint64 `json:"-"`
	CurrentStationID    atomic.Uint64 `json:"-"`
	CurrentPlattformID  atomic.Uint64 `json:"-"`
	CurrentActiveTileID atomic.Uint64 `json:"-"`
	Ticker              *time.Ticker  `json:"-"`
	Tick                int
	IsPaused            bool `json:"-"` // nicht speichern, weil immer beim speichern pausiert ist.

	BroadcastChannel chan WsEnvelope        `json:"-"`
	UserInputs       chan RecieveWSEnvelope `json:"-"`
	UnPause          chan bool              `json:"-"`
	ConfirmPause     chan bool              `json:"-"`

	Logger *slog.Logger `json:"-"`

	SizeSubtile int // muss immer 4 sein, Jannis hatte komische Ideen

	Money int // der Kontostand

	WaggonTypes map[string]*WaggonType

	currentTrain      *Train `json:"-"` // nur für das einfache hinzufügen von Waggons, wird auch nicht im savegame gespeichert
	currentWaggonType string `json:"-"` // ebenfalls weil ich faul bin und um redundanz zu vermeiden
}

// alle Attribute, die man am anfang senden möchte
type SendAbleGamestate struct {
	Users     map[string]*User // der name ist einzigartig
	Schedules map[int]*Schedule
	Stations  map[int]*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}
