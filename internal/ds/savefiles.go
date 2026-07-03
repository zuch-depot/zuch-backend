package ds

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

//DAS ALLES braucht support für mehrere Instanzen auf einem Server
// vielleicht auch erstmal nicht

// Im Tick Loop NUR MIT GO aufrufen, der schließt aber richtig, ich schwöre
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
	if saveGameName == "" {
		// format ist yyyymmdd-hhmmss
		saveGameName = gs.ConfigData.SaveLocation + "/savegame-" + time.Now().Format("20060102-150405")
	} else {
		saveGameName = gs.ConfigData.SaveLocation + "/savegame-" + saveGameName + "-" + time.Now().Format("20060102-150405")
	}
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
		saveGameName += ".json"
		bytesToWrite = stateByte
	}
	// dann wird es auf die Festplatte geschrieben
	// Hier vielleicht noch die zeit oder das datum ans ende des dateinamen
	gs.Logger.Info("Writing gamestate to Disk", slog.String("Filename", saveGameName))
	err = os.WriteFile(saveGameName, bytesToWrite, 0644)

	if err != nil {
		gs.Logger.Error("Failure while writing File", slog.String("Error", err.Error()))
		return "", err
	}

	// max. 5 savefiles
	files, _ := os.ReadDir(gs.ConfigData.SaveLocation)
	autosaves := make([]os.DirEntry, 0)
	// zählen der autosaves
	for _, b := range files {
		pieces := strings.Split(b.Name(), "-")
		if _, err := strconv.Atoi(pieces[1]); err != nil {
			continue
		}
		pieces[2] = strings.TrimSuffix(pieces[2], ".json")
		if _, err := strconv.Atoi(pieces[2]); err != nil {
			continue
		}
		if pieces[0] != "savegame" {
			continue
		}
		autosaves = append(autosaves, b)
	}
	if len(autosaves) > gs.ConfigData.NumberSaveFiles {
		// sortieren

		slices.SortFunc(autosaves, func(a os.DirEntry, b os.DirEntry) int {
			if a.Name() == b.Name() {
				return 0
			}
			if a.Name() < b.Name() {
				return -1
			}
			return 1
		})
		// macht desc
		slices.Reverse(autosaves)
		// löschen der nicht letzten 5
		for _, b := range autosaves[5:] {
			os.Remove(gs.ConfigData.SaveLocation + "//" + b.Name())
		}
	}

	fmt.Println("Done Saving")

	gs.UnPauseGame()
	return saveGameName, nil
}

// wenn es autosaves gibt, dann nimm den aktuellsten. Ansonsten den ersten aus map nehmen. Wenn es einen savePath gibt, dann den nehmen
// kann nur jsons, Path muss mit json sein
func (gs *GameState) LoadGame(savePath string) error {

	//eigentlich pausieren
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()

	// wen kein Path angegeben wurde
	if savePath == "" {
		//alle Dateien aus saves rauslesen
		entries, err := os.ReadDir("saves")
		if err != nil {
			gs.Logger.Error(err.Error())
			return err
		}

		// gibt keine autosaves, also einen aus maps nehmen
		if len(entries) == 0 {

			entries, err = os.ReadDir("maps")
			if err != nil {
				gs.Logger.Error(err.Error())
				return err
			}

			// es gibt auch keine Map
			if len(entries) == 0 {
				return fmt.Errorf("No savefile in saves or maps found")
			}

			savePath = "maps/" + entries[0].Name()

		} else {
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
			savePath = "saves/" + entries[0].Name()
		}
	}

	data, err := os.ReadFile(savePath)
	if err != nil {
		gs.Logger.Error(err.Error())
		return err
	}

	err = json.Unmarshal(data, &gs)
	if err != nil {
		gs.Logger.Error(err.Error())
		return err
	}

	fmt.Println("Loaded Game " + savePath)

	gs.CurrentTrainID.Store(uint64(gs.ConfigData.Ids.CurrentTrainID))
	gs.CurrentScheduleID.Store(uint64(gs.ConfigData.Ids.CurrentScheduleID))
	gs.CurrentStopID.Store(uint64(gs.ConfigData.Ids.CurrentStopID))
	gs.CurrentStationID.Store(uint64(gs.ConfigData.Ids.CurrentStationID))
	gs.CurrentPlattformID.Store(uint64(gs.ConfigData.Ids.CurrentPlattformID))
	gs.CurrentActiveTileID.Store(uint64(gs.ConfigData.Ids.CurrentActiveTileID))

	return nil
}
