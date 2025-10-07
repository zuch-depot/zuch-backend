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
func createTrains() {
	//stations inkl. Initialisieren
	stations = append(stations, &Station{Name: "Station Nord", Id: 1, Capacity: 100, Storage: map[string]int{}})
	plattforms := []Plattform{{Name: "Gleis 1", Tiles: [][2]int{{2, 0}, {3, 0}}, station: stations[0]}}
	stations[0].changeStationTile(false, [2]int{2, 0})
	stations[0].changeStationTile(false, [2]int{3, 0})

	stations = append(stations, &Station{Name: "Station Süd", Id: 2, Capacity: 150, Storage: map[string]int{}})
	plattforms = append(plattforms, Plattform{Name: "Gleis 31", Tiles: [][2]int{{3, 7}, {4, 7}, {5, 7}}, station: stations[1]})
	stations[1].changeStationTile(false, [2]int{3, 7})
	stations[1].changeStationTile(false, [2]int{4, 7})
	stations[1].changeStationTile(false, [2]int{5, 7})

	//Zug eins mit Schedule
	Stops := []Stop{
		{Id: 1, Plattform: &plattforms[0], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			{CargoType: []string{"Pommes"}},
			{Loading: true, CargoType: []string{"Kartoffeln", "Sonnenblumenöl"}}}},
		// {Id: 2, Goal: [3]int{1, 3, 4}, Name: "Wegpunkt 1"},
		{Id: 3, Plattform: &plattforms[1], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			{CargoType: []string{"Kartoffeln", "Sonnenblumenöl"}},
			{Loading: true, CargoType: []string{"Pommes"}}}}}
	schedules = append(schedules, &Schedule{Stops: Stops})
	temp := []*Waggon{
		{Position: [3]int{4, 4, 1}},
		{Position: [3]int{3, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{3, 4, 1}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{2, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}}}
	trains = append(trains, &Train{Waggons: temp, Schedule: *schedules[0], Name: "RE1", NextStop: Stops[0], Id: int(currentTrainID.Load())})
	currentTrainID.Add(1)

	// Zug zwei
	temp = []*Waggon{
		{Position: [3]int{6, 6, 2}},
		{Position: [3]int{6, 5, 4}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}},
		{Position: [3]int{6, 5, 2}, CargoStorage: &CargoStorage{capacity: 30, CargoCategory: "Lebensmittel"}}}
	trains = append(trains, &Train{Waggons: temp, Schedule: *schedules[0], Name: "RE2", NextStop: Stops[1], Id: int(currentTrainID.Load())})
	currentTrainID.Add(1)

}

func initializeTiles() {
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
	tiles = make([][]*Tile, sizeX)
	for i := range tiles {
		tiles[i] = make([]*Tile, sizeY)
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
				temp := configData.ActiveTileCategories["Bauernhof"]
				aktiveTile = ActiveTile{Name: "Bauernhof Nord", Category: &temp, maxStorage: 150}
				activeTiles = append(activeTiles, &aktiveTile)
			case "L":
				temp := configData.ActiveTileCategories["Lebensmittelfabrik"]
				aktiveTile = ActiveTile{Name: "Lebensmittelfabrik Süd", Category: &temp, maxStorage: 50, Storage: map[string]int{"Kartoffeln": 100}}
				activeTiles = append(activeTiles, &aktiveTile)
			}

			//signals testing
			var signals [4]bool
			for p := range testSignals {
				if testSignals[p][0] == int(o) && testSignals[p][1] == int(i) {
					signals[testSignals[p][2]-1] = true
				}
			}

			tiles[o][i] = &Tile{IsPlattform: false, Tracks: tracks, Signals: signals, ActiveTile: &aktiveTile}
		}
	}

	fmt.Println("Tiles initialised with a Map size of", sizeX, sizeY)
}

// testing
func printTrains() {
	for _, i := range trains {
		fmt.Print("Train", i.Name)
		for _, waggon := range i.Waggons {
			fmt.Print(waggon.CargoStorage, "")
		}
		fmt.Println("")
	}
	fmt.Println("-----------------------")
}

// nur fürs Testen
func printMap() {
	//i = y
	for i := range tiles {
		fmt.Print(".")
		for o := range tiles {

			isSignal, s := isSignalAt(o, i)

			isTrain, t := isTrainAt(o, i)
			if isTrain {
				fmt.Print(t)
			} else if isSignal {
				fmt.Print(s)
			} else {
				switch tiles[o][i].Tracks {
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
func isTrainAt(x int, y int) (bool, int) {
	for i := range trains {
		waggons := trains[i].Waggons
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
func isSignalAt(x int, y int) (bool, int) {
	signals := tiles[x][y].Signals
	for i := range signals {
		if signals[i] {
			return true, i + 1
		}
	}
	return false, 0
}

func handleTileUpdate(envelope recieveWSEnvelope, tiles [][]*Tile) {
	var update tileUpdateMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		logger.Error("EROOR", slog.String("error", err.Error()))
	}

	sizeX, err := strconv.ParseInt(os.Getenv("XSIZE"), 10, 64)
	if err != nil {
		logger.Error("Error while loading Size of Map in the x dimension", slog.String("error", err.Error()))
	}

	sizeY, err := strconv.ParseInt(os.Getenv("YSIZE"), 10, 64)
	if err != nil {
		logger.Error("Error while loading Size of Map in the y dimension", slog.String("error", err.Error()))
	}

	if !((0 <= update.position[0] && update.position[0] < int(sizeX)) && (0 <= update.position[1] && update.position[1] < int(sizeY)) && (0 < update.position[2] && update.position[2] <= 4)) {
		logger.Error("Invalid coordinates in wsEnvolope, ignoring this envolpe and continuing", slog.String("username", envelope.user.username))
		return
	}

	switch update.Action {
	case "build":
		switch update.Subject {
		case "rail":
			executeAndReply(tiles[update.position[0]][update.position[1]].addTrack, &envelope, &update)
		case "signal":
			executeAndReply(tiles[update.position[0]][update.position[1]].addSignal, &envelope, &update)

		}
	case "remove":
		switch update.Subject {
		case "rail":
			executeAndReply(tiles[update.position[0]][update.position[1]].removeTrack, &envelope, &update)
		case "signal":
			executeAndReply(tiles[update.position[0]][update.position[1]].removeSignal, &envelope, &update)

		}
	}

}

func executeAndReply(callback func(int) (bool, string), envelope *recieveWSEnvelope, update *tileUpdateMSG) {
	success, msg := callback(update.position[2])
	if success {
		broadcastChannel <- wsEnvelope{Type: envelope.Type, Username: "Server", Msg: &tileUpdateMSG{position: update.position, Subject: update.Subject, Action: update.Action}}
		envelope.reply(success, "")
	} else {
		envelope.reply(success, msg)
	}
}
