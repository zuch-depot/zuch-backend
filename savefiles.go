package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log/slog"
	"os"
)

// Speichert den Spielstand
// ist aber bisher pass-by-value, dunno ob reference hier vielleicht mehr sinn macht
// Die ticks sollte man noch anhalten
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

	// Einzelnes Object das hoffentlich den ganzen status des Spiels darstellt
	state := gamestate{Users: users, Schedules: schedules, Stations: stations, Tiles: tiles, Trains: trains}
	// Die Objekte werden in einer netten JSON verpackt
	stateByte, err := json.Marshal(state)
	if err != nil {
		logger.Error("Could not save Json", slog.String("Error", err.Error()))
	}

	// Damit die nicht so riesig wird, wird noch mit gzip das ganze etwas komprimiert
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write(stateByte)

	// dann wird es auf die Festplatte geschrieben
	// Hier vielleicht noch die zeit oder das datum ans ende des dateinamen
	err = os.WriteFile(os.Getenv("SAVELOCATION"), buf.Bytes(), 04)
	if err != nil {
		logger.Error("Failure while writing File", slog.String("Error", err.Error()))
	}
}

type gamestate struct {
	Users     []User
	Schedules []Schedule
	Stations  []Station
	Tiles     [][]Tile
	Trains    []Train
}
