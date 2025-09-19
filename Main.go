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
	tiles     [][]Tile
	trains    []Train
)
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true}))
var userInputs = make(chan UserInput, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert

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

		//Speichern, welche Tiles am Ende des Threads entblocked werden muss
		var tilesToUnblock []*Tile

		//Client Inputs
		processClientInputs()

		//Train move
		if tick%30 == 0 {
			moveTrains()
		}

		//process factorys

		//load/unload

		//entblocken
		for i := range tilesToUnblock {
			tilesToUnblock[i].isBlocked = false
		}

		//anzeigen Testing
		if tick%30 == 0 {
			printMap()
		}
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
