package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"zuch-backend/internal/ds"
	"zuch-backend/internal/utils"

	"github.com/joho/godotenv"
	"github.com/telemachus/humane"
)

// Zum Debuggen: LevelInfo zu LevelDebug wechseln
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelInfo}))

func main() {
	utils.Logger = logger

	gs := ds.GameState{UserInputs: make(chan ds.RecieveWSEnvelope, 300), BroadcastChannel: make(chan ds.WsEnvelope, 100), UnPause: make(chan bool), SizeSubtile: 4, Trains: make(map[int]*ds.Train), Logger: logger}
	err := godotenv.Load("main.env")
	if err != nil {
		logger.Error("Oh oh ein fehler in den environment variables", slog.String("Error", err.Error()))
	}
	// loading global variables
	tempVar, err := strconv.ParseInt(os.Getenv("LOADUNLOADSPEED"), 10, 64)
	if err != nil {
		logger.Error("Error while loading LoadUnloadSpeed", slog.String("Error", err.Error()))
	}
	gs.LoadUnloadSpeed = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MINLOADUNLOADTICKS"), 10, 64)
	if err != nil {
		logger.Error("Error while loading MinLoadUloadTicks", slog.String("Error", err.Error()))
	}
	gs.MinLoadUloadTicks = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("MAXDISTANCEACTIVETILETOSTATION"), 10, 64)
	if err != nil {
		logger.Error("Error while loading the radius of the station, where aktive Tiles are detected", slog.String("Error", err.Error()))
	}
	gs.StationRange = int(tempVar)

	tempVar, err = strconv.ParseInt(os.Getenv("CAPACITYPERSTATIONTILE"), 10, 64)
	if err != nil {
		logger.Error("Error while loading the capacity per station tile", slog.String("Error", err.Error()))
	}
	gs.CapacityPerStationTile = int(tempVar)

	// wichtig als initialisierung, bevor Züge verarbeitet werden
	loadConfig(&gs)

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles(&gs)
	createDemoTrains(&gs)
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	// hier den Server starten
	go startServer(&gs)
	// Anfangen aus events an clients zu schicken
	go startListiningToBroadcast(gs.BroadcastChannel, &gs)
	// Zeit pro Tick bestimmen
	ticksMilisec, err := strconv.Atoi(os.Getenv("TICKTIMEMILISEC"))
	if err != nil {
		logger.Error("Failed to convert Ticktime to Int", slog.String("Error", err.Error())) // anderes Log?
	}

	gs.Ticker = time.NewTicker(time.Duration(ticksMilisec) * time.Millisecond)

	// jeder Tick
	for gs.Tick = 0; ; gs.Tick++ {
		// Wenn pausiert wurde, warten bis entpausiert signal kommt
		if gs.IsPaused {
			confirmPause <- true
			<-gs.UnPause // Hier warten bis es wieder entpausiert wird
			gs.IsPaused = false
			logger.Info("continuing after Pause")
		}

		// Client Inputs
		processClientInputs(&gs)

		// Train calculate (Läd/Entläd oder bewegt) und entblocken
		if gs.Tick%10 == 0 {
			calculateTrains(&gs)
			// for _, train := range gs.Trains {
			// 	fmt.Print(train.Name, " ")
			// 	for _, waggon := range train.Waggons {
			// 		fmt.Print(waggon.CargoStorage.FilledCargoType, " ", waggon.CargoStorage.Filled, " ")
			// 	}
			// 	fmt.Println()
			// }
		}

		// process factorys
		if gs.Tick%10 == 1 {
			processActiveTiles(&gs)
			// for _, station := range gs.Stations {
			// 	fmt.Println(station.Name, station.Storage)
			// }
			// for _, active := range gs.ActiveTiles {
			// 	fmt.Println(active.Name, " ", active.Storage)
			// }
		}

		if gs.Tick%10 == 2 {
			gs.CalculateCargoPaths()
		}

		// anzeigen Testing
		if gs.Tick%10 == 0 {
			// fmt.Println("tick", tick)
		}
		// das wartet hier bis ein tick ausgelöst wird,

		<-gs.Ticker.C
	}
}

func processClientInputs(gs *ds.GameState) {
	for len(gs.UserInputs) > 0 {
		input := <-gs.UserInputs
		inputCat := strings.Split(input.Type, ".")
		var err error
		switch inputCat[0] {
		case "rail":
			err = handleTileUpdate(input, gs)
		case "signal":
			err = handleTileUpdate(input, gs)
		case "train":
			err = handleTrainUpdate(input, gs)
		default:
			input.Reply(false, "Invalid Envelope Type", gs)
		}
		if err != nil {
			logger.Debug(err.Error())
			input.Reply(false, err.Error(), gs)
		} else {
			input.Reply(true, "", gs)
		}

	}
}

func calculateTrains(gs *ds.GameState) {
	// Speichern, welche Tiles am Ende des Threads entblocked werden muss
	var tilesToUnblock [][2]int

	for i := range gs.Trains {
		temp := gs.Trains[i].CalculateTrain(gs)
		if temp[0] >= 0 {
			tilesToUnblock = append(tilesToUnblock, temp)
		}
	}

	// entblocken
	for _, i := range tilesToUnblock {
		gs.Tiles[i[0]][i[1]].IsBlocked = false
	}
	if len(tilesToUnblock) > 0 {
		gs.BroadcastChannel <- ds.WsEnvelope{Type: "tiles.unblock", Username: "kaputt", Msg: ds.BlockedTilesMSG{Tiles: tilesToUnblock}}
	}
}

func loadConfig(gs *ds.GameState) {
	// JSON-Datei öffnen
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal in Struktur
	if err := json.Unmarshal(file, &gs.ConfigData); err != nil {
		panic(err)
	}
}

func startListiningToBroadcast(broadcastChannel <-chan ds.WsEnvelope, gs *ds.GameState) {
	for {
		envelope, ok := <-broadcastChannel
		if ok {
			for _, user := range gs.Users {
				if user.IsConnected {
					logger.Debug("Notifying client of Change", slog.String("User", user.Username), slog.String("Type", envelope.Type))
					user.WebSocketQueue <- envelope
				}
			}
		}
	}
}
