package main

import (
	"fmt"
	"log"
	"os"
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

	//every Tick
	for {
		//Speichern, welche Tiles am Ende des Threads entblocked werden muss

		start := time.Now()

		time.Sleep(10 * time.Millisecond)
		fmt.Println("ich bin in main", <-userInputs)
		remainingTime := time.Since(start) //zielzeit muss noch eingerechnet werden -> sollte in config

		if remainingTime.Milliseconds() < 1 {
			//logging, dass der Server hin
		} else {
			time.Sleep(time.Duration(remainingTime.Nanoseconds()))
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
