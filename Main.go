package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/telemachus/humane"
)

var (
	users       []*User
	schedules   []*Schedule
	stations    []*Station
	tiles       [][]*Tile
	trains      []*Train
	activeTiles []*ActiveTile

	loadUnloadSpeed   int
	minLoadUloadTicks int
	configData        ConfigData //übergeordetes Struct, in das alles aus config.json reingeladen wird

	stationRange int

	//Plattforms
)
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelInfo}))
var userInputs = make(chan recieveWSEnvelope, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert
var unPause = make(chan bool)
var broadcastChannel = make(chan wsEnvelope, 100)

var isPaused = false

func main() {
	godotenv.Load("main.env")

	//loading global variables
	tempVar, err := strconv.ParseInt(os.Getenv("LOADUNLOADSPEED"), 10, 64)
	if err != nil {
		log.Println("Error while loading LoadUnloadSpeed", err)
	}
	loadUnloadSpeed = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MINLOADUNLOADTICKS"), 10, 64)
	if err != nil {
		log.Println("Error while loading minLoadUloadTicks", err)
	}
	minLoadUloadTicks = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MAXDISTANCEACTIVETILETOSTATION"), 10, 64)
	if err != nil {
		log.Println("Error while loading the radius of the station, where aktive Tiles are detected", err)
	}
	stationRange = int(tempVar)

	//wichtig als initialisierung, bevor Züge verarbeitet werden
	loadConfig()

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles()
	createTrains()
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	// hier den Server starten
	go startServer()
	// Anfangen aus events an clients zu schicken
	go startListiningToBroadcast(broadcastChannel)
	//Zeit pro Tick bestimmen
	ticksMilisec, err := strconv.Atoi(os.Getenv("TICKTIMEMILISEC"))
	if err != nil {
		logger.Error("Failed to convert Ticktime to Int", slog.String("Error", err.Error())) //anderes Log?// Panic beendet das programm :(
	}

	ticker := time.NewTicker(time.Duration(ticksMilisec) * time.Millisecond)

	//jeder Tick
	for tick := 0; ; tick++ {
		// Wenn pausiert wurde, warten bis entpausiert signal kommt
		if isPaused {
			confirmPause <- true
			<-unPause
			isPaused = false
			logger.Info("continuing after Pause")
		}

		//Client Inputs
		processClientInputs()

		//Train calculate (Läd/Entläd oder bewegt) und entblocken
		if tick%10 == 0 {
			// printTrains()
			calculateTrains()
		}

		//process factorys
		if (tick+1)%10 == 0 {
			processActiveTiles()
		}

		//anzeigen Testing
		if tick%10 == 0 {
			printMap()
			// fmt.Println("tick", tick)
		}
		// das wartet hier bis ein tick ausgelöst wird,

		<-ticker.C
	}
}

func processClientInputs() {
	for len(userInputs) > 0 {
		input := <-userInputs

		switch input.Type {
		case "tile.update":
			handleTileUpdate(input, tiles)

		}
	}
}

func calculateTrains() {
	//Speichern, welche Tiles am Ende des Threads entblocked werden muss
	var tilesToUnblock [][2]int

	for i := range trains {
		temp := trains[i].calculateTrain()
		if temp[0] >= 0 {
			tilesToUnblock = append(tilesToUnblock, temp)
		}
	}

	//entblocken
	for _, i := range tilesToUnblock {
		tiles[i[0]][i[1]].IsBlocked = false
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

func loadConfig() {

	// JSON-Datei öffnen
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal in Struktur
	if err := json.Unmarshal(file, &configData); err != nil {
		panic(err)
	}
}

func startListiningToBroadcast(broadcastChannel <-chan wsEnvelope) {
	for {
		envelope, ok := <-broadcastChannel
		if ok {
			for _, user := range users {
				if user.isConnected {
					logger.Debug("Notifying client of Change", slog.String("User", user.username), slog.String("Type", envelope.Type))
					user.webSocketQueue <- envelope
				}
			}

		}
	}
}
