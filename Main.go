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
	schedules []Schedule
	stations  []Station
	tiles     [][]*Tile
	trains    []Train
)
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true}))
var userInputs = make(chan UserInput, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert
var unPause = make(chan bool)

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

		//Speichern, welche Tiles am Ende des Threads entblocked werden muss
		var tilesToUnblock []*Tile

		//Client Inputs
		processClientInputs()

		//Train move
		if tick%10 == 0 {
			moveTrains()
			printTrains()
		}

		//process factorys

		//load/unload

		//entblocken
		for i := range tilesToUnblock {
			tilesToUnblock[i].IsBlocked = false
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

		switch input.action {
		case "":

		}
	}
}

func moveTrains() {
	for i := range trains {
		trains[i].move()
	}
}
