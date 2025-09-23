package main

import (
	"fmt"
	"log"
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
		" .|. . .+.+. . . .|", //1
		" .|. . . .+.+. . .|", //2
		" .|. . . . .|. .+.+", //3
		" .+.-.-.+.-.+. .|.|", //4
		" . .+.-.+. .|. .|.|", //5
		" . .|. . . .|. .+.+", //6
		" . .+.-.-.-.+. . .|", //7
		" . . . . . .|. . .|", //8
		" . . . . . .|. . . ", //9
	}
	testSignals = [][3]int{
		[3]int{1, 3, 4},
	}
)

// nur fürs Testen, inkl. Schedule
func createTrains() {
	//stations
	stations = append(stations, &Station{Name: "Station Nord", capacity: 100, Storage: map[CargoType]int{Potatos: 100}})
	plattforms := []Plattform{{Name: "Gleis 1", Tiles: [][2]int{{2, 0}, {3, 0}}, station: stations[0]}}
	stations = append(stations, &Station{Name: "Station Süd", capacity: 150, Storage: map[CargoType]int{Coal: 50}})
	plattforms = append(plattforms, Plattform{Name: "Gleis 31", Tiles: [][2]int{{3, 7}, {4, 7}, {5, 7}}, station: stations[1]})

	//Zug eins mit Schedule
	Stops := []Stop{
		{Id: 1, Plattform: &plattforms[0], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			LoadUnloadCommand{Loading: true, CargoType: []CargoType{Potatos}}}},
		{Id: 2, Goal: [3]int{1, 3, 4}, Name: "Wegpunkt 1"},
		{Id: 3, Plattform: &plattforms[1], IsPlattform: true, LoadUnloadCommand: [2]LoadUnloadCommand{
			LoadUnloadCommand{CargoType: []CargoType{Potatos}}}}}
	schedules = append(schedules, &Schedule{Stops: Stops})
	temp := []*TrainType{
		{position: [3]int{4, 4, 1}},
		{position: [3]int{3, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, filledCargoType: Potatos}},
		{position: [3]int{3, 4, 1}, CargoStorage: &CargoStorage{capacity: 30, filledCargoType: Potatos}},
		{position: [3]int{2, 4, 3}, CargoStorage: &CargoStorage{capacity: 30, filledCargoType: Potatos}}}
	trains = append(trains, &Train{Waggons: temp, Schedule: *schedules[0], Name: "RE1", NextStop: Stops[0]})
	// Zug zwei
	temp = []*TrainType{
		{position: [3]int{6, 6, 2}},
		{position: [3]int{6, 5, 4}},
		{position: [3]int{6, 5, 2}}}
	trains = append(trains, &Train{Waggons: temp, Schedule: *schedules[0], Name: "RE2", NextStop: Stops[2]})
}
func initializeTiles() {
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
			var tracks [4]bool
			switch line[o] {
			case "-":
				tracks = [4]bool{true, false, true, false}
			case "|":
				tracks = [4]bool{false, true, false, true}
			case "+":
				tracks = [4]bool{true, true, true, true}
			}

			//signals testing
			var signals [4]bool
			for p := range testSignals {
				if testSignals[p][0] == int(o) && testSignals[p][1] == int(i) {
					signals[testSignals[p][2]-1] = true
				}
			}

			tiles[o][i] = &Tile{IsPlattform: false, Tracks: tracks, Signals: signals}
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

func isTrainAt(x int, y int) (bool, int) {
	for i := range trains {
		waggons := trains[i].Waggons
		for o := range waggons {
			pos := waggons[o].position
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

/* erst auf dauerhaft blocked prüfen
*to visit (außer, da wo man hergekommen ist):
*	1 [x][y][2,3,4], [x-1][y][3]
*	2 [x][y][1,3,4], [x][y+1][4]
*	3 [x][y][1,2,4], [x+1][y][1]
*	4 [x][y][1,2,3], [x][y+1][2]
 */
func neighbourTracks(x int, y int, sub int) [][3]int {
	var r [][3]int

	appending := func(a [3]int) {
		for i := range 3 {
			o := a[i]
			if tiles[x][y].Tracks[o-1] {
				r = append(r, [3]int{x, y, o})
			}
		}
	}

	switch sub {
	case 1:
		if x > 0 {
			if tiles[x-1][y].Tracks[2] {
				r = append(r, [3]int{x - 1, y, 3})
			}
		}
		appending([3]int{2, 3, 4})
	case 2:
		if y > 0 {
			if tiles[x][y-1].Tracks[3] {
				r = append(r, [3]int{x, y - 1, 4})
			}
		}
		appending([3]int{1, 3, 4})
	case 3:
		if x != len(tiles)-1 {
			if tiles[x+1][y].Tracks[0] {
				r = append(r, [3]int{x + 1, y, 1})
			}
		}
		appending([3]int{1, 2, 4})
	case 4:
		if y != len(tiles[0])-1 {
			if tiles[x][y+1].Tracks[1] {
				r = append(r, [3]int{x, y + 1, 2})
			}
		}
		appending([3]int{1, 2, 3})
	}
	return r
}
