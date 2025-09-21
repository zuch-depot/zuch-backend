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
		"-.+.-.-.+.+.-.-.-.-", //0
		" .|. . .+.+. . . . ", //1
		" .|. . . .+.+. . . ", //2
		" .|. . . . .|. . . ", //3
		" .+.-.-.+.-.+. . . ", //4
		" . .+.-.+. .|. . . ", //5
		" . .|. . . .|. . . ", //6
		" . .+.-.-.-.+.-.-. ", //7
		" . . . . . .|. . . ", //8
		" . . . . . .|. . . ", //9
	}
	testSignals = [][3]int{
		[3]int{3, 4, 3},
		[3]int{4, 5, 1},
		//[3]int{5, 7, 3},
		[3]int{6, 6, 2},
		[3]int{5, 2, 2},
	}
)

// nur fürs Testen, inkl. Schedule
func createTrains() {
	//Zug eins mit Schedule
	stops := []Stop{
		Stop{id: 1, goal: [3]int{6, 7, 2}},
		Stop{id: 2, goal: [3]int{4, 0, 1}},
		Stop{id: 3, goal: [3]int{1, 3, 4}}}
	schedules = append(schedules, Schedule{stops: stops})
	temp := []TrainType{
		TrainType{position: [3]int{4, 4, 1}},
		TrainType{position: [3]int{3, 4, 3}},
		TrainType{position: [3]int{3, 4, 1}},
		TrainType{position: [3]int{2, 4, 3}}}
	trains = append(trains, Train{train: temp, schedule: schedules[0], name: "1"})
	stops = []Stop{
		Stop{id: 1, goal: [3]int{6, 7, 2}},
		Stop{id: 2, goal: [3]int{5, 4, 1}}}
	schedules = append(schedules, Schedule{stops: stops})
	temp = []TrainType{
		TrainType{position: [3]int{3, 7, 3}},
		TrainType{position: [3]int{3, 7, 1}},
		TrainType{position: [3]int{2, 7, 3}}}
	trains = append(trains, Train{train: temp, schedule: schedules[1], name: "2"})
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

			tiles[o][i] = &Tile{isPlattform: false, tracks: tracks, signals: signals}
		}
	}

	fmt.Println("Tiles initialised with a Map size of", sizeX, sizeY)
}

// testing
func printTrains() {
	for i := range trains {
		fmt.Println("Train", trains[i].name, trains[i].train)
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
				switch tiles[o][i].tracks {
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
		waggons := trains[i].train
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
	signals := tiles[x][y].signals
	for i := range signals {
		if signals[i] {
			return true, i + 1
		}
	}
	return false, 0
}
