package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"zuch-backend/internal/ds"
)

// nur testing
var (
	testMap = []string{
		/* 	für ecken↘↙↗↖, zeigt jeweils in die ecke wo die schiene ist
			für Ts ↓→↑←, zeigen aud das mittlere der drei subtiles
		0.1.2.3.4.5.6.7.8.9*/
		" .↖.-.-.↓.↓.-. . .|", //0
		" .|.B. .↙.←. . . .|", //1
		" .|. . . .↙.↗. . .|", //2
		" .|. . . . .|. .↖.←", //3
		" .↙.-.-.↓.-.+.↓.↘.|", //4
		" . .↖.-.↘. .|.|.S.|", //5
		" . .|. . . .|.↙.-.←", //6
		" . .→.-.-.-.←. . .|", //7
		" .↖.←. . . .|. . .|", //8
		" .↙.↘. . .L.|. . . ", //9
	}
	testSignals = [][3]int{
		{1, 3, 4},
		{4, 5, 2},
		{4, 5, 2},
		{7, 5, 2},
	}
)

// nur fürs Testen, inkl. Schedule
func createDemoTrains(gs *ds.GameState) {
	//TEMP
	gs.CurrentPlattformID.Add(1)
	gs.CurrentStationID.Add(1)
	gs.CurrentScheduleID.Add(1)
	gs.CurrentStopID.Add(1)
	gs.CurrentTrainID.Add(1)
	gs.CurrentActiveTileID.Add(1)

	//stations inkl. Initialisieren
	pos := [2]int{2, 0}
	gs.AddStationTile(pos) //hier wird auch die Station und Plattform erstellt
	//Bestimmung der Station zum umbenennen
	plattform, _ := gs.GetPlattform(pos)
	plattform.GetStation(gs).Name = "Station Nord"
	plattform.Name = "Gleis 1"
	gs.AddStationTile([2]int{3, 0})

	pos = [2]int{3, 7}
	gs.AddStationTile(pos)
	plattform2, _ := gs.GetPlattform(pos)
	plattform2.GetStation(gs).Name = "Station Süd"
	plattform2.Name = "Gleis 31"
	gs.AddStationTile([2]int{4, 7})
	gs.AddStationTile([2]int{5, 7})

	pos = [2]int{9, 4}
	gs.AddStationTile(pos)
	plattform, _ = gs.GetPlattform(pos)
	plattform.GetStation(gs).Name = "Station Ost"
	plattform.Name = "Gleis 2"
	gs.AddStationTile([2]int{9, 5})

	// fmt.Println(gs.Stations)

	//Zug eins mit Schedule
	var schedule *ds.Schedule
	schedule, _ = gs.AddSchedule("Schdule Nord")
	var stop *ds.Stop
	stop, _ = schedule.AddStopStation(gs.Stations[1].Plattforms[1], gs)
	stop.SetLoadCommand([]string{"Kartoffeln", "Sonnenblumenöl"}, false)
	stop.SetUnloadCommand([]string{"Pommes"}, false)

	stop, _ = schedule.AddStopStation(gs.Stations[3].Plattforms[3], gs)
	stop.SetLoadCommand([]string{"Pommes"}, false)
	stop.SetUnloadCommand([]string{"Kartoffeln", "Sonnenblumenöl"}, false)

	train, err := gs.AddTrain("RE1",
		[]ds.TrainCreateWaggons{
			{Position: [3]int{3, 4, 3}, Typ: "Lebensmittel"},
			{Position: [3]int{3, 4, 1}, Typ: "Lebensmittel"},
			{Position: [3]int{2, 4, 3}, Typ: "Lebensmittel"},
			{Position: [3]int{2, 4, 1}, Typ: "Lebensmittel"}},
	)
	train.Schedule = schedule
	if err != nil {
		gs.Logger.Error("Fehler, aber ist im demo ding egal")
	}

	// Zug zwei mit eigenem Schedule
	schedule, _ = gs.AddSchedule("Schedule Süd")
	stop, _ = schedule.AddStopStation(gs.Stations[3].Plattforms[3], gs)
	stop.SetLoadCommand([]string{"Kartoffeln", "Sonnenblumenöl"}, false)
	stop.SetUnloadCommand([]string{"Pommes"}, false)

	stop, _ = schedule.AddStopStation(gs.Stations[2].Plattforms[2], gs)
	stop.SetLoadCommand([]string{"Pommes"}, false)
	stop.SetUnloadCommand([]string{"Kartoffeln", "Sonnenblumenöl"}, false)

	train, err = gs.AddTrain("RE2",
		[]ds.TrainCreateWaggons{
			{Position: [3]int{6, 6, 2}, Typ: "Lebensmittel"},
			{Position: [3]int{6, 5, 4}, Typ: "Lebensmittel"},
			{Position: [3]int{6, 5, 2}, Typ: "Lebensmittel"},
			{Position: [3]int{6, 4, 4}, Typ: "Lebensmittel"}})
	train.Schedule = schedule

	if err != nil {
		gs.Logger.Error("Fehler, aber ist im demo ding egal")
	}
}

func initializeTiles(gs *ds.GameState) {
	// Setzt die erste Zug ID, pass hier halbwegs zum initialisieren
	gs.CurrentTrainID.Store(0)

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
	gs.Tiles = make([][]*ds.Tile, sizeX)
	for i := range gs.Tiles {
		gs.Tiles[i] = make([]*ds.Tile, sizeY)
	}

	//Erstellung der Tiles
	for y := range sizeY {
		line := strings.Split(testMap[y], ".") //testing
		for x := range sizeX {
			//hier die Infos für das Tile laden

			//testing
			var aktiveTile ds.ActiveTile
			var tracks [4]bool
			switch line[x] {
			case "-":
				tracks = [4]bool{true, false, true, false}
			case "|":
				tracks = [4]bool{false, true, false, true}
			case "+":
				tracks = [4]bool{true, true, true, true}
			case "↖":
				tracks = [4]bool{false, false, true, true}
			case "↗":
				tracks = [4]bool{true, false, false, true}
			case "↙":
				tracks = [4]bool{false, true, true, false}
			case "↘":
				tracks = [4]bool{true, true, false, false}
			case "→":
				tracks = [4]bool{false, true, true, true}
			case "↓":
				tracks = [4]bool{true, false, true, true}
			case "←":
				tracks = [4]bool{true, true, false, true}
			case "↑":
				tracks = [4]bool{true, true, true, false}
			case "B":
				temp := gs.ConfigData.ActiveTileCategories["Bauernhof"]
				aktiveTile = ds.ActiveTile{Id: 3, Name: "Bauernhof Nord", Category: &temp, MaxStorage: 150}
				gs.ActiveTiles = append(gs.ActiveTiles, &aktiveTile)
			case "L":
				temp := gs.ConfigData.ActiveTileCategories["Lebensmittelfabrik"]
				aktiveTile = ds.ActiveTile{Id: 1, Name: "Lebensmittelfabrik Süd", Category: &temp, MaxStorage: 50, Storage: map[string]int{"Kartoffeln": 100, "Sonnenblumenöl": 50}}
				gs.ActiveTiles = append(gs.ActiveTiles, &aktiveTile)
			case "S":
				temp := gs.ConfigData.ActiveTileCategories["Stadt"]
				aktiveTile = ds.ActiveTile{Id: 2, Name: "Wuppertal", Category: &temp, MaxStorage: 50}
				gs.ActiveTiles = append(gs.ActiveTiles, &aktiveTile)
			}

			//signals testing
			var signals [4]bool
			for p := range testSignals {
				if testSignals[p][0] == int(x) && testSignals[p][1] == int(y) {
					signals[testSignals[p][2]-1] = true
				}
			}

			gs.Tiles[x][y] = &ds.Tile{IsPlattform: false, Tracks: tracks, Signals: signals, ActiveTile: &aktiveTile, IsLocked: aktiveTile.Stations == nil, X: int(x), Y: int(y)}
		}
	}
	gs.SizeX = int(sizeX)
	gs.SizeY = int(sizeY)
	logger.Info("Tiles initialised with a Map size of", slog.Int64("SizeX", sizeX), slog.Int64("SizeY", sizeY))
}

// nur fürs Testen
func printMap(gs *ds.GameState) {
	//i = y
	for i := range gs.Tiles {
		fmt.Print(".")
		for o := range gs.Tiles {

			isSignal, s := isSignalAt(o, i, gs)

			isTrain, t := isTrainAt(o, i, gs)
			if isTrain {
				fmt.Print(t)
			} else if isSignal {
				fmt.Print(s)
			} else {
				switch gs.Tiles[o][i].Tracks {
				case [4]bool{true, true, true, true}:
					fmt.Print("+")
				case [4]bool{true, false, true, false}:
					fmt.Print("-")
				case [4]bool{false, true, false, true}:
					fmt.Print("|")
				default:
					fmt.Print(" ")
				}
			}
			fmt.Print(".")
		}
		fmt.Println("")
	}
	fmt.Println("----------------------")
}

// fürs testen (printMap)
func isTrainAt(x int, y int, gs *ds.GameState) (bool, int) {
	for i := range gs.Trains {
		waggons := gs.Trains[i].Waggons
		for o := range waggons {
			pos := waggons[o].Position
			if pos[0] == x && pos[1] == y {
				return true, pos[2]
			}
		}
	}
	return false, 0
}

// nur 1 Signal pro Tile anzeigen
func isSignalAt(x int, y int, gs *ds.GameState) (bool, int) {
	signals := gs.Tiles[x][y].Signals
	for i := range signals {
		if signals[i] {
			return true, i + 1
		}
	}
	return false, 0
}

func unpackEnvelope[T any](envelope ds.RecieveWSEnvelope, typ T) (T, error) {
	var dest T
	err := json.Unmarshal(envelope.Msg, &dest)
	if err != nil {
		logger.Error("error", slog.String("error", err.Error()))
		return dest, fmt.Errorf("%s", err.Error())
	}
	return dest, nil
}

func checkIfCoordinatesAreValid(position [3]int, gs *ds.GameState) error {
	if !((0 <= position[0] && position[0] < int(gs.SizeX)) && (0 <= position[1] && position[1] < int(gs.SizeY)) && (0 < position[2] && position[2] <= gs.SizeSubtile)) {
		return fmt.Errorf("coordinates are invalid")
	} else {
		return nil
	}
}

// warum hast du das so komisch gemacht Jannis?
// -> wenn du die funktionen direkt aufrufst, sieht das in der Übersicht besser aus
// Führt das callback mit den daten des envelopes aus, tritt ein fehler aus wird der zurück gegeben, andererseits wird die nachricht an alle geschickt
func executeAndReply(callback func(int, *ds.GameState) (bool, string), envelope *ds.RecieveWSEnvelope, update *ds.TileUpdateMSG, gs *ds.GameState) error {
	success, msg := callback(update.Position[2], gs)
	if success {
		envelope.Reply(success, "", gs)
		return nil
	} else {
		return fmt.Errorf("%s", msg)
	}
}
