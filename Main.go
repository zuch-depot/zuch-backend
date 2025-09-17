package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	users []User
	//acceptor  ClientAcceptor
	schedules []Schedule
	stations  []Station
	tiles     []Tile
	trains    []Train
)
var logger = log.New(os.Stdout, "Server:", log.LstdFlags)
var userInputs = make(chan UserInput, 300) //Queue, die die UserInputs bis zum Start des nächsten Ticks speichert

func main() {
	godotenv.Load("main.env")

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
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

		//Client Inputs

		//Train move

		//process factorys

		//load/unload

		//syncen, dass jeder Tick nur 1 mal
		remainingTime := tickTime - time.Since(start).Abs().Milliseconds()
		if remainingTime < 1 {
			//logging, dass der Server hinterher hinkt
		} else {
			time.Sleep(time.Duration(remainingTime * 1000))
		}
	}
}

type User struct {
	id       int
	username string
}

func processClientInputs() {
	for len(userInputs) > 0 {

	}
}
