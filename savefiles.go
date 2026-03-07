package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
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

	fmt.Println("Saving")

	pauseGame(gs)

	// Einzelnes Object das hoffentlich den ganzen status des Spiels darstellt

	state := ds.SaveAbleGamestate{Users: gs.Users, Schedules: gs.Schedules, Stations: gs.Stations, Tiles: gs.Tiles, Trains: gs.Trains,
		ActiveTiles: gs.ActiveTiles, LoadUnloadSpeed: gs.LoadUnloadSpeed, MinLoadUloadTicks: gs.MinLoadUloadTicks, CapacityPerStationTile: gs.CapacityPerStationTile,
		CurrentTrainID: int(gs.CurrentTrainID.Load()), CurrentScheduleID: int(gs.CurrentScheduleID.Load()), ConfigData: gs.ConfigData, StationRange: gs.StationRange,
		CurrentStopID: int(gs.CurrentStopID.Load()), CurrentStationID: int(gs.CurrentStationID.Load()), CurrentPlattformID: int(gs.CurrentPlattformID.Load()),
		CurrentActiveTileID: int(gs.CurrentActiveTileID.Load()), SizeX: gs.SizeX, SizeY: gs.SizeY, SizeSubtile: gs.SizeSubtile, Tick: gs.Tick}
	// state := ds.SendAbleGamestate{Users: gs.Users, Schedules: gs.Schedules, Stations: gs.Stations, Tiles: gs.Tiles, Trains: gs.Trains}
	// Die Objekte werden in einer netten JSON verpackt

	// Dateiname wird ggf. abgeändert wenn es nicht compressed wird
	filename := os.Getenv("SAVELOCATION") + "/savegame-" + time.Now().Format("20060102-150405")
	var bytesToWrite []byte

	stateByte, err := json.Marshal(state)
	if err != nil {
		logger.Error("Could not convert state to JSON", slog.String("Error", err.Error()))
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

	fmt.Println("Done Saving")

	unPauseGame(gs)
}
