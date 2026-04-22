package ds

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"
)

//DAS ALLES braucht support für mehrere Instanzen auf einem Server

// Im Tick Loop NUR MIT GO aufrufen
// Speichert den Spielstand
// ist aber bisher pass-by-value, dunno ob reference hier vielleicht mehr sinn macht
// Die ticks sollte man noch anhalten
func (gs *GameState) SaveGame(saveGameName string) (string, error) {
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

	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()

	fmt.Println("Saving")

	gs.PauseGame()

	//die Ids in integern speichern
	gs.ConfigData.Ids = ids{
		CurrentTrainID:      int(gs.CurrentTrainID.Load()),
		CurrentScheduleID:   int(gs.CurrentScheduleID.Load()),
		CurrentStopID:       int(gs.CurrentStopID.Load()),
		CurrentStationID:    int(gs.CurrentStationID.Load()),
		CurrentPlattformID:  int(gs.CurrentPlattformID.Load()),
		CurrentActiveTileID: int(gs.CurrentActiveTileID.Load()),
	}

	// Die Objekte werden in einer netten JSON verpackt

	// Dateiname wird ggf. abgeändert wenn es nicht compressed wird
	// format ist yyyymmdd-hhmmss
	filename := gs.ConfigData.SaveLocation + "/savegame-" + time.Now().Format("20060102-150405")
	var bytesToWrite []byte

	stateByte, err := json.Marshal(gs)
	if err != nil {
		gs.Logger.Error("Could not convert state to JSON", slog.String("Error", err.Error()))
		return "", err
	}

	if gs.ConfigData.SaveCompressed {

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
	gs.Logger.Info("Writing gamestate to Disk", slog.String("Filename", filename))
	err = os.WriteFile(filename, bytesToWrite, 0644)

	if err != nil {
		gs.Logger.Error("Failure while writing File", slog.String("Error", err.Error()))
		return "", err
	}

	fmt.Println("Done Saving")

	gs.UnPauseGame()
	return filename, nil
}

// nur in saves
// kann nur jsons
func (gs *GameState) LoadGame(saveName string) error {

	//eigentlich pausieren
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()

	//alle Dateien aus saves rauslesen
	entries, err := os.ReadDir("saves")
	if err != nil {
		gs.Logger.Error(err.Error())
		return err
	}

	if len(entries) == 0 {
		gs.Logger.Error("No Savefile found")
		return fmt.Errorf("No Savefile found")
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
				gs.Logger.Error("That Savefile not found")
				return fmt.Errorf("Savefile not found")
			}
		}
	}

	saveFileName = "saves/" + saveFileName

	// var sgs SaveAbleGamestate

	fmt.Println(saveFileName)

	data, err := os.ReadFile(saveFileName)
	if err != nil {
		gs.Logger.Error(err.Error())
		return err
	}

	err = json.Unmarshal(data, &gs)
	if err != nil {
		gs.Logger.Error(err.Error())
		return err
	}

	gs.CurrentTrainID.Store(uint64(gs.ConfigData.Ids.CurrentTrainID))
	gs.CurrentScheduleID.Store(uint64(gs.ConfigData.Ids.CurrentScheduleID))
	gs.CurrentStopID.Store(uint64(gs.ConfigData.Ids.CurrentStopID))
	gs.CurrentStationID.Store(uint64(gs.ConfigData.Ids.CurrentStationID))
	gs.CurrentPlattformID.Store(uint64(gs.ConfigData.Ids.CurrentPlattformID))
	gs.CurrentActiveTileID.Store(uint64(gs.ConfigData.Ids.CurrentActiveTileID))

	return nil
}
