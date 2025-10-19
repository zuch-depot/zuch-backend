package main

// entfernt ntes Element aus slice
func removeElementFromSlice[T any](slice []T, n int) []T {
	if n >= len(slice) {
		logger.Debug("Index out of Bounds. Tried to remove the not existing Element " + string(rune(n)))
	}
	if n < 0 {
		logger.Debug("Index out of Bounds. Tried to remove a negative Element" + string(rune(n)))
	}
	tempSlice := slice
	first := tempSlice[:n]
	var last []T
	if n < len(slice)-1 {
		last = (slice)[n+1:]
	}
	tempSlice = first
	tempSlice = append(tempSlice, last...)
	return tempSlice
}

// bestimmt den absoluten Wert eines Ints
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// finde die CargoCategory des CargoTypes
func getCargoCategory(cargoType string) string {
	//iteriere die CargoCategorys
	for key, value := range configData.TrainCategories {
		//suche in der aktuellen Category nach dem Type
		for _, value2 := range value {
			//wenn gefunden, kann der zurückgegeben werden
			if value2 == cargoType {
				return key
			}
		}
	}
	return ""
}

/* erst auf dauerhaft blocked prüfen
*to visit (außer, da wo man hergekommen ist):
*	1 [x][y][2,3,4], [x-1][y][3]
*	2 [x][y][1,3,4], [x][y+1][4]
*	3 [x][y][1,2,4], [x+1][y][1]
*	4 [x][y][1,2,3], [x][y+1][2]
 */
// verified
func neighbourTracks(x int, y int, sub int) [][3]int {
	var connectedNeigbours [][3]int // [3]int identifiziert mit x y und subtile ein exaktes Subtile,
	// Alle Koordinaten von Subtiles zurückgeben die angrenzen die können im gleichem subtile oder im angrenzendem

	appending := func(subtilesToCheck [3]int) {
		for i := range 3 {
			o := subtilesToCheck[i]
			if tiles[x][y].Tracks[o-1] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y, o})
			}
		}
	}

	switch sub {
	case 1:
		if x > 0 {
			if tiles[x-1][y].Tracks[2] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x - 1, y, 3})
			}
		}
		appending([3]int{2, 3, 4})
	case 2:
		if y > 0 {
			if tiles[x][y-1].Tracks[3] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y - 1, 4})
			}
		}
		appending([3]int{1, 3, 4})
	case 3:
		if x != len(tiles)-1 {
			if tiles[x+1][y].Tracks[0] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x + 1, y, 1})
			}
		}
		appending([3]int{1, 2, 4})
	case 4:
		if y != len(tiles[0])-1 {
			if tiles[x][y+1].Tracks[1] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y + 1, 2})
			}
		}
		appending([3]int{1, 2, 3})
	}
	return connectedNeigbours
}
