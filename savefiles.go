package main

import (
	"encoding/json"
	"os"
)

func saveGame(users []*User,
	schedules []Schedule,
	stations []Station,
	tiles [][]Tile,
	trains []Train,
) {
	// ich will den ganzen bums hier eigentlich ja nur speichern
	// Alles soll gerne in eine Json datei
	// wenn die uns um die ohren fliegt kann man ja immernoch komprimieren
	// grobe Struktur
	// {
	//  users []*User,
	// 	schedules []Schedule,
	// 	stations []Station,
	// 	tiles [][]Tile,
	// 	trains []Train,
	// }

	file, err := os.Create("./saves/savegame.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(gamestate{Users: users, Schedules: schedules, Stations: stations, Tiles: tiles, Trains: trains})

}

type gamestate struct {
	Users     []*User
	Schedules []Schedule
	Stations  []Station
	Tiles     [][]Tile
	Trains    []Train
}
