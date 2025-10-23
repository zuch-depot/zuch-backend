package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/telemachus/humane"
)

type gameState struct {
	Users       []*User
	Schedules   []*Schedule
	Stations    []*Station
	Tiles       [][]*Tile
	Trains      []*Train
	ActiveTiles []*ActiveTile

	loadUnloadSpeed   int
	minLoadUloadTicks int
	configData        ConfigData //übergeordetes Struct, in das alles aus config.json reingeladen wird

	stationRange int
	ticker       *time.Ticker
	tick         int
	isPaused     bool

	broadcastChannel chan wsEnvelope
	userInputs       chan recieveWSEnvelope
	unPause          chan bool

	sizeX       int
	sizeY       int
	SizeSubtile int
}

// var (
// 	users       []*User
// 	schedules   []*Schedule
// 	stations    []*Station
// 	tiles       [][]*Tile
// 	trains      []*Train
// 	activeTiles []*ActiveTile

// 	loadUnloadSpeed   int
// 	minLoadUloadTicks int
// 	configData        ConfigData //übergeordetes Struct, in das alles aus config.json reingeladen wird

// 	stationRange int

//	 isPaused bool
//		//Plattforms
//
// )
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelInfo}))

//Queue, die die UserInputs bis zum Start des nächsten Ticks speichert

func main() {

	gs := gameState{userInputs: make(chan recieveWSEnvelope, 300), broadcastChannel: make(chan wsEnvelope, 100), unPause: make(chan bool), SizeSubtile: 4}
	godotenv.Load("main.env")

	//loading global variables
	tempVar, err := strconv.ParseInt(os.Getenv("LOADUNLOADSPEED"), 10, 64)
	if err != nil {
		log.Println("Error while loading LoadUnloadSpeed", err)
	}
	gs.loadUnloadSpeed = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MINLOADUNLOADTICKS"), 10, 64)
	if err != nil {
		log.Println("Error while loading minLoadUloadTicks", err)
	}
	gs.minLoadUloadTicks = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MAXDISTANCEACTIVETILETOSTATION"), 10, 64)
	if err != nil {
		log.Println("Error while loading the radius of the station, where aktive Tiles are detected", err)
	}
	gs.stationRange = int(tempVar)

	//wichtig als initialisierung, bevor Züge verarbeitet werden
	loadConfig(&gs)

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles(&gs)
	createTrains(&gs)
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	// hier den Server starten
	go startServer(&gs)
	// Anfangen aus events an clients zu schicken
	go startListiningToBroadcast(gs.broadcastChannel, &gs)
	//Zeit pro Tick bestimmen
	ticksMilisec, err := strconv.Atoi(os.Getenv("TICKTIMEMILISEC"))
	if err != nil {
		logger.Error("Failed to convert Ticktime to Int", slog.String("Error", err.Error())) //anderes Log?
	}

	gs.ticker = time.NewTicker(time.Duration(ticksMilisec) * time.Millisecond)

	//jeder Tick
	for gs.tick = 0; ; gs.tick++ {
		// Wenn pausiert wurde, warten bis entpausiert signal kommt
		if gs.isPaused {
			confirmPause <- true
			<-gs.unPause // Hier warten bis es wieder entpausiert wird
			gs.isPaused = false
			logger.Info("continuing after Pause")
		}

		//Client Inputs
		processClientInputs(&gs)

		//Train calculate (Läd/Entläd oder bewegt) und entblocken
		if gs.tick%10 == 0 {
			// printTrains()
			calculateTrains(&gs)
		}

		//process factorys
		if gs.tick%10 == 1 {
			processActiveTiles(&gs)
		}

		//anzeigen Testing
		if gs.tick%10 == 0 {
			// printMap()
			// fmt.Println("tick", tick)
		}
		// das wartet hier bis ein tick ausgelöst wird,

		<-gs.ticker.C
	}
}

func processClientInputs(gs *gameState) {
	for len(gs.userInputs) > 0 {
		input := <-gs.userInputs
		inputCat := strings.Split(input.Type, ".")
		switch inputCat[0] {
		case "rail":
			handleRailUpdate(input, gs)
		case "signal":
			handleSignalUpdate(input, gs)
		case "train":
			handleCreateTrain(input, gs)
		default:
			input.reply(false, "Invalid Envelope Type")
		}

	}
}

func calculateTrains(gs *gameState) {
	//Speichern, welche Tiles am Ende des Threads entblocked werden muss
	var tilesToUnblock [][2]int

	for i := range gs.Trains {
		temp := gs.Trains[i].calculateTrain(gs)
		if temp[0] >= 0 {
			tilesToUnblock = append(tilesToUnblock, temp)
		}
	}

	//entblocken
	for _, i := range tilesToUnblock {
		gs.Tiles[i[0]][i[1]].IsBlocked = false
	}
}

type Produktionszyklus struct {
	Consumtion                 map[string]int `json:"Verbrauch"`
	Produktion                 map[string]int `json:"Produktion"`
	VerfuegbareLevelUndScaling []int          `json:"verfügbareLevelUndScaling"`
}

// Struktur für aktive Tiles
type ActiveTileCategory struct {
	Productioncycles []Produktionszyklus `json:"Produktionszyklen"`
}

// Root-Struktur für config.json
type ConfigData struct {
	TrainCategories      map[string][]string           `json:"Train Categories"`
	ActiveTileCategories map[string]ActiveTileCategory `json:"Aktive Tiles"`
}

func loadConfig(gs *gameState) {

	// JSON-Datei öffnen
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal in Struktur
	if err := json.Unmarshal(file, &gs.configData); err != nil {
		panic(err)
	}
}

func startListiningToBroadcast(broadcastChannel <-chan wsEnvelope, gs *gameState) {
	for {
		envelope, ok := <-broadcastChannel
		if ok {
			for _, user := range gs.Users {
				if user.isConnected {
					logger.Debug("Notifying client of Change", slog.String("User", user.username), slog.String("Type", envelope.Type))
					user.webSocketQueue <- envelope
				}
			}

		}
	}
}
