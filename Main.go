package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/telemachus/humane"
)

var (
	users     []*User
	schedules []*Schedule
	stations  []*Station
	tiles     [][]*Tile
	trains    []*Train
	//Plattform
)
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelInfo}))
var userInputs = make(chan recieveWSEnvelope, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert
var unPause = make(chan bool)
var broadcastChannel = make(chan wsEnvelope, 100)

var isPaused = false

func main() {
	godotenv.Load("main.env")

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

	ticker := time.NewTicker(time.Duration(ticksMilisec) * time.Millisecond)
	if err != nil {
		logger.Error("Failed to convert Ticktime to Int", slog.String("Error", err.Error())) //anderes Log?// Panic beendet das programm :(
	}

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

		//Train move
		if tick%10 == 0 {
			moveTrains()
			//printTrains()
		}

		//process factorys
		//load/unload

		//anzeigen Testing
		if tick%100 == 0 {
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

func moveTrains() {
	//Speichern, welche Tiles am Ende des Threads entblocked werden muss
	var tilesToUnblock [][2]int

	for i := range trains {
		temp := trains[i].move()
		if temp[0] >= 0 {
			tilesToUnblock = append(tilesToUnblock, temp)
		}
	}

	//entblocken
	for _, i := range tilesToUnblock {
		tiles[i[0]][i[1]].IsBlocked = false
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
