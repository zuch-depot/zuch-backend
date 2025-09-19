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
		"-.-.-.-.+.-.-.-.-.-", //0
		" . . . .|. . . . . ", //1
		" . . . .|. . . . . ", //2
		" . . . .|. . . . . ", //3
		" .-.-.-.+.-.+. . . ", //4
		" . . . .|. .|. . . ", //5
		" . . . .|. .|. . . ", //6
		" . .-.-.+.-.+.-.-. ", //7
		" . . . . . .|. . . ", //8
		" . . . . . .|. . . ", //9
	}
)

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
	tiles = make([][]Tile, sizeX)
	for i := range tiles {
		tiles[i] = make([]Tile, sizeY)
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
			tiles[o][i] = Tile{IsPlattform: false, Tracks: tracks}
		}
	}

	fmt.Println("Tiles initialised with a Map size of", sizeX, sizeY)
}

// nur fürs Testen
func createTrains() {
	//Zug eins
	trains = append(trains, Train{tilePosition: [3]int{1, 0, 3}, tileGoal: [3]int{4, 4, 4}})
}

// nur fürs Testen
func printMap() {
	//i = y
	for i := range tiles {
		fmt.Print(".")
		for o := range tiles {

			isTrain, t := isTrainAt(o, i)
			if isTrain {
				fmt.Print(t)
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
		pos := trains[i].tilePosition
		if pos[0] == x && pos[1] == y {
			return true, pos[2]
		}
	}
	return false, 0
}
