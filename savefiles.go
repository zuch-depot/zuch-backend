package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log/slog"
	"os"
	"time"
	"zuch-backend/internal/ds"
)

// Speichert den Spielstand
// ist aber bisher pass-by-value, dunno ob reference hier vielleicht mehr sinn macht
// Die ticks sollte man noch anhalten
func saveGame(gs *ds.GameState) {
	// ich will den ganzen bums hier eigentlich ja nur speichern
	// Alles soll gerne in eine Json datei
	// wenn die uns um die ohren fliegt kann man ja immernoch komprimieren
	// grobe Struktur
	// {
	//  users []*User,
	// 	schedules []Schedule,
	// 	stations []Station,
	// 	tiles [][]*Tile,
	// 	trains []Train,
	// }

	pauseGame(gs)

	// Einzelnes Object das hoffentlich den ganzen status des Spiels darstellt

	state := ds.SendAbleGamestate{Users: gs.Users, Schedules: gs.Schedules, Stations: gs.Stations, Tiles: gs.Tiles, Trains: gs.Trains}
	// Die Objekte werden in einer netten JSON verpackt

	// Dateiname wird ggf. abgeändert wenn es nicht compressed wird
	filename := os.Getenv("SAVELOCATION") + "/savegame-" + time.Now().Format("20060102-150405")
	var bytesToWrite []byte

	stateByte, err := json.Marshal(state)
	if err != nil {
		logger.Error("COuld not convert state to JSON", slog.String("Error", err.Error()))
	}

	if os.Getenv("SAVECOMPRESSED") == "True" {

		// Damit die nicht so riesig wird, wird noch mit gzip das ganze etwas komprimiert
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		zw.Write(stateByte)
		bytesToWrite = stateByte
	} else {
		// falls man die sich selber mal anschauen will
		filename += ".json"
		bytesToWrite = stateByte
	}
	// dann wird es auf die Festplatte geschrieben
	// Hier vielleicht noch die zeit oder das datum ans ende des dateinamen
	logger.Info("Writing gamestate to Disk", slog.String("Filename", filename))
	err = os.WriteFile(filename, bytesToWrite, 0644)

	if err != nil {
		logger.Error("Failure while writing File", slog.String("Error", err.Error()))
	}

	unPauseGame(gs)
}
