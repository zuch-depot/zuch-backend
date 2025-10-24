package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// nur testing
var (
	testMap = []string{
		/*
		 0.1.2.3.4.5.6.7.8.9*/
		" .+.-.-.+.+.-. . .|", //0
		" .|.B. .+.+. . . .|", //1
		" .|. . . .+.+. . .|", //2
		" .|. . . . .|. .+.+", //3
		" .+.-.-.+.-.+. .|.|", //4
		" . .+.-.+. .|. .|.|", //5
		" . .|. . . .|. .+.+", //6
		" . .+.-.-.-.+. . .|", //7
		" .+.+. . . .|. . .|", //8
		" .+.+. . .L.|. . . ", //9
	}
	testSignals = [][3]int{
		{1, 3, 4},
		{4, 5, 2},
	}
)

// nur fürs Testen, inkl. Schedule
func createTrains(gs *gameState) {
	//stations inkl. Initialisieren
	gs.Stations = append(gs.Stations, &Station{Name: "Station Nord", Id: 1, Capacity: 100, Storage: map[string]int{}})
	plattforms := []Plattform{{Name: "Gleis 1", Tiles: [][2]int{{2, 0}, {3, 0}}, station: gs.Stations[0]}}
	gs.Stations[0].changeStationTile(false, [2]int{2, 0}, gs)
	gs.Stations[0].changeStationTile(false, [2]int{3, 0}, gs)

	gs.Stations = append(gs.Stations, &Station{Name: "Station Süd", Id: 2, Capacity: 150, Storage: map[string]int{}})
	plattforms = append(plattforms, Plattform{Name: "Gleis 31", Tiles: [][2]int{{3, 7}, {4, 7}, {5, 7}}, station: gs.Stations[1]})
	gs.Stations[1].changeStationTile(false, [2]int{3, 7}, gs)
	gs.Stations[1].changeStationTile(false, [2]int{4, 7}, gs)
	gs.Stations[1].changeStationTile(false, [2]int{5, 7}, gs)

	//Zug eins mit Schedule
	Stops := []Stop{
		{Id: 1, Plattform: &plattforms[0], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			{CargoType: []string{"Pommes"}},
			{Loading: true, CargoType: []string{"Kartoffeln", "Sonnenblumenöl"}}}},
		// {Id: 2, Goal: [3]int{1, 3, 4}, Name: "Wegpunkt 1"},
		{Id: 3, Plattform: &plattforms[1], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			{CargoType: []string{"Kartoffeln", "Sonnenblumenöl"}},
			{Loading: true, CargoType: []string{"Pommes"}}}}}
	gs.Schedules = append(gs.Schedules, &Schedule{Stops: Stops})
	temp := []*Waggon{
		{Position: [3]int{4, 4, 1}},
		{Position: [3]int{3, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{3, 4, 1}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{2, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}}}
	gs.Trains[int(currentTrainID.Load())] = &Train{Waggons: temp, Schedule: *gs.Schedules[0], Name: "RE1", NextStop: Stops[0], Id: int(currentTrainID.Load())}
	currentTrainID.Add(1)

	// Zug zwei
	temp = []*Waggon{
		{Position: [3]int{6, 6, 2}},
		{Position: [3]int{6, 5, 4}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{6, 5, 2}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}}}
	gs.Trains[int(currentTrainID.Load())] = &Train{Waggons: temp, Schedule: *gs.Schedules[0], Name: "RE2", NextStop: Stops[1], Id: int(currentTrainID.Load())}
	currentTrainID.Add(1)

}

func initializeTiles(gs *gameState) {
	// Setzt die erste Zug ID, pass hier halbwegs zum initialisieren
	currentTrainID.Store(0)

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
	gs.Tiles = make([][]*Tile, sizeX)
	for i := range gs.Tiles {
		gs.Tiles[i] = make([]*Tile, sizeY)
	}

	//Erstellung der Tiles
	for i := range sizeY {
		line := strings.Split(testMap[i], ".") //testing
		for o := range sizeX {
			//hier die Infos für das Tile laden

			//testing
			var aktiveTile ActiveTile
			var tracks [4]bool
			switch line[o] {
			case "-":
				tracks = [4]bool{true, false, true, false}
			case "|":
				tracks = [4]bool{false, true, false, true}
			case "+":
				tracks = [4]bool{true, true, true, true}
			case "B":
				temp := gs.configData.ActiveTileCategories["Bauernhof"]
				aktiveTile = ActiveTile{Name: "Bauernhof Nord", Category: &temp, maxStorage: 150}
				gs.ActiveTiles = append(gs.ActiveTiles, &aktiveTile)
			case "L":
				temp := gs.configData.ActiveTileCategories["Lebensmittelfabrik"]
				aktiveTile = ActiveTile{Name: "Lebensmittelfabrik Süd", Category: &temp, maxStorage: 50, Storage: map[string]int{"Kartoffeln": 100}}
				gs.ActiveTiles = append(gs.ActiveTiles, &aktiveTile)
			}

			//signals testing
			var signals [4]bool
			for p := range testSignals {
				if testSignals[p][0] == int(o) && testSignals[p][1] == int(i) {
					signals[testSignals[p][2]-1] = true
				}
			}

			gs.Tiles[o][i] = &Tile{IsPlattform: false, Tracks: tracks, Signals: signals, ActiveTile: &aktiveTile}
		}
	}
	gs.sizeX = int(sizeX)
	gs.sizeY = int(sizeY)
	fmt.Println("Tiles initialised with a Map size of", sizeX, sizeY)
}

// testing
func printTrains(gs *gameState) {
	for _, i := range gs.Trains {
		fmt.Print("Train", i.Name)
		for _, waggon := range i.Waggons {
			fmt.Print(waggon.CargoStorage, "")
		}
		fmt.Println("")
	}
	fmt.Println("-----------------------")
}

// nur fürs Testen
func printMap(gs *gameState) {
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
func isTrainAt(x int, y int, gs *gameState) (bool, int) {
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
func isSignalAt(x int, y int, gs *gameState) (bool, int) {
	signals := gs.Tiles[x][y].Signals
	for i := range signals {
		if signals[i] {
			return true, i + 1
		}
	}
	return false, 0
}

func unpackEnvelope[T any](envelope recieveWSEnvelope, typ T) (T, error) {
	var dest T
	err := json.Unmarshal(envelope.Msg, &dest)
	if err != nil {
		logger.Error("error", slog.String("error", err.Error()))
		return dest, fmt.Errorf("%s", err.Error())
	}
	return dest, nil
}

func checkIfCoordinatesAreValid(position [3]int, gs *gameState) error {
	if !((0 <= position[0] && position[0] < int(gs.sizeX)) && (0 <= position[1] && position[1] < int(gs.sizeY)) && (0 < position[2] && position[2] <= gs.SizeSubtile)) {
		return fmt.Errorf("coordinates are invalid")
	} else {
		return nil
	}
}

func handleTileUpdate(envelope recieveWSEnvelope, gs *gameState) error {
	update, err := unpackEnvelope(envelope, tileUpdateMSG{})
	if err != nil {
		return fmt.Errorf("Could not unpack Envelope", slog.String("error", err.Error()))

	}
	err = checkIfCoordinatesAreValid(update.Position, gs)
	if err != nil {
		return fmt.Errorf("Envelope contains invalid Coordinates", slog.String("error", err.Error()))
	}

	switch envelope.Type {
	case "rail.create":
		return executeAndReply(gs.Tiles[update.Position[0]][update.Position[1]].addTrack, &envelope, &update, gs)
	case "rail.remove":
		return executeAndReply(gs.Tiles[update.Position[0]][update.Position[1]].removeTrack, &envelope, &update, gs)
	case "signal.create":
		return executeAndReply(gs.Tiles[update.Position[0]][update.Position[1]].addSignal, &envelope, &update, gs)
	case "signal.remove":
		return executeAndReply(gs.Tiles[update.Position[0]][update.Position[1]].removeSignal, &envelope, &update, gs)
	default:
		return fmt.Errorf("Unknown envelope Tyoe")
	}
}

// Führt das callback mit den daten des envelopes aus, tritt ein fehler aus wird der zurück gegeben, andererseits wird die nachricht an alle geschickt
func executeAndReply(callback func(int) (bool, string), envelope *recieveWSEnvelope, update *tileUpdateMSG, gs *gameState) error {
	success, msg := callback(update.Position[2])
	if success {
		gs.broadcastChannel <- wsEnvelope{Type: envelope.Type, Username: "Server", Msg: &tileUpdateMSG{Position: update.Position}}
		envelope.reply(success, "")
		return nil
	} else {
		return fmt.Errorf(msg)
	}
}
