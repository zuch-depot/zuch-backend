package main

import (
	"log"
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

	loadUnloadSpeed   int
	minLoadUloadTicks int
	//Plattforms
)
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelDebug}))
var userInputs = make(chan UserInput, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert
var unPause = make(chan bool)

var isPaused = false

type wsEnvelope struct {
	Type string
	Msg  any
}

func main() {
	godotenv.Load("main.env")

	temp, err := strconv.ParseInt(os.Getenv("LOADUNLOADSPEED"), 10, 64)
	if err != nil {
		log.Println("Error while loading LoadUnloadSpeed", err)
	}
	loadUnloadSpeed = int(temp)

	temp, err = strconv.ParseInt(os.Getenv("MINLOADUNLOADTICKS"), 10, 64)
	if err != nil {
		log.Println("Error while loading minLoadUloadTicks", err)
	}
	minLoadUloadTicks = int(temp)

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles()
	createTrains()
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	// hier den Server starten
	go startServer()

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

		//Train calculate (Läd/Entläd oder bewegt) und entblocken
		if tick%10 == 0 {
			// printTrains()
			calculateTrains()
		}

		//process factorys

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

		switch input.action {
		case "":

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
