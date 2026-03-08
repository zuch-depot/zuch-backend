package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"
	"zuch-backend/internal/ds"
)

//DAS ALLES braucht support für mehrere Instanzen auf einem Server

// Im Tick Loop NUR MIT GO aufrufen
// Speichert den Spielstand
// ist aber bisher pass-by-value, dunno ob reference hier vielleicht mehr sinn macht
// Die ticks sollte man noch anhalten
func saveGame(gs *ds.GameState, saveGameName string) string {
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
	// format ist yyyymmdd-hhmmss
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
	return filename
}

// nur in saves
// kann nur jsons
func loadGame(gs *ds.GameState, saveName string) {

	//eigentlich pausieren

	//alle Dateien aus saves rauslesen
	entries, error := os.ReadDir("saves")
	if error != nil {
		logger.Error(error.Error())
		return
	}

	if len(entries) == 0 {
		logger.Error("No Savefile found")
		return
	}

	//neuste zuerst, also größtes Datum zuerst
	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		if a.Name() < b.Name() {
			return 1
		}
		if a.Name() > b.Name() {
			return -1
		}
		return 0
	})

	var saveFileName string

	if saveName == "" {
		saveFileName = entries[0].Name()
	} else {
		for _, entry := range entries {
			if strings.Split(entry.Name(), "-")[3] == saveName {
				saveFileName = entry.Name()
				break
			} else {
				//TODO konnte das FIle nicht finden
				logger.Error("That Savefile not found")
				return
			}
		}
	}

	saveFileName = "saves/" + saveFileName

	var sgs ds.SaveAbleGamestate

	fmt.Println(saveFileName)

	data, error := os.ReadFile(saveFileName)
	if error != nil {
		logger.Error(error.Error())
		return
	}

	error = json.Unmarshal(data, &sgs)
	if error != nil {
		logger.Error(error.Error())
		return
	}

	gs.Users = sgs.Users
	for _, user := range gs.Users {
		user.IsConnected = false
	}
	gs.Schedules = sgs.Schedules
	gs.Stations = sgs.Stations
	gs.Tiles = sgs.Tiles
	gs.Trains = sgs.Trains
	gs.ActiveTiles = sgs.ActiveTiles

	gs.LoadUnloadSpeed = sgs.LoadUnloadSpeed
	gs.MinLoadUloadTicks = sgs.MinLoadUloadTicks
	gs.CapacityPerStationTile = sgs.CapacityPerStationTile
	gs.ConfigData = sgs.ConfigData

	gs.StationRange = sgs.StationRange
	setAtomic(&gs.CurrentTrainID, sgs.CurrentTrainID)
	setAtomic(&gs.CurrentScheduleID, sgs.CurrentScheduleID)
	setAtomic(&gs.CurrentStopID, sgs.CurrentStopID)
	setAtomic(&gs.CurrentStationID, sgs.CurrentStationID)
	setAtomic(&gs.CurrentPlattformID, sgs.CurrentPlattformID)
	setAtomic(&gs.CurrentActiveTileID, sgs.CurrentActiveTileID)

	gs.Tick = sgs.Tick

	gs.SizeX = sgs.SizeX
	gs.SizeY = sgs.SizeY
	gs.SizeSubtile = sgs.SizeSubtile
}

func setAtomic(atomic *atomic.Uint64, add int) {
	for range add {
		atomic.Add(1)
	}
}
