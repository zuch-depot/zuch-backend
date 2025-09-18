package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	users     []User
	schedules []Schedule
	stations  []Station
	tiles     [][]Tile
	trains    []Train
)
var logger = log.New(os.Stdout, "Server:", log.LstdFlags)
var userInputs = make(chan UserInput, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert

func main() {
	godotenv.Load("main.env")

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles()
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	// hier den Server starten
	go startServer()

	//Zeit pro Tick bestimmen
	tickTime, err := strconv.ParseInt(os.Getenv("TICKTIMEMILISEC"), 10, 64)
	if err != nil {
		panic(err) //anderes Log?
	}

	//jeder Tick
	for {
		start := time.Now()

		//Speichern, welche Tiles am Ende des Threads entblocked werden muss
		var tilesToUnblock []Tile

		//Client Inputs
		processClientInputs()

		//Train move

		//process factorys

		//load/unload

		//entblocken
		for i := range tilesToUnblock {
			tilesToUnblock[i].isBlocked = false
		}

		//syncen, dass jeder Tick nur 1 mal
		remainingTime := tickTime - time.Since(start).Abs().Milliseconds()
		if remainingTime < 1 {
			//logging, dass der Server hinterher hinkt
		} else {
			time.Sleep(time.Duration(remainingTime * 1000))
		}
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

func initializeTiles() {
	//Map Größe aus config laden
	sizeX, err := strconv.ParseInt(os.Getenv("XSIZE"), 10, 64)
	if err != nil {
		log.Println("Error while loading Size of Map in the x dimension", err)
	}

	sizeY, err := strconv.ParseInt(os.Getenv("YSIZE"), 10, 64)
	if err != nil {
		log.Println("Error while loading Size of Map in the y dimension", err)
	}

	//initalising 2d slice
	tiles := make([][]Tile, sizeX)
	for i := range tiles {
		tiles[i] = make([]Tile, sizeY)
	}

	//Erstellung der Tiles
	for i := range sizeX {
		for o := range sizeY {
			//hier die Infos für das Tile laden
			tiles[i][o] = Tile{isPlattform: false}
		}
	}

	fmt.Println("Tiles initialised with a Map size of", sizeX, sizeY)
}
