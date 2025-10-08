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
func getCargoCategory(cargoType string, gs *gameState) string {
	//iteriere die CargoCategorys
	for key, value := range gs.configData.TrainCategories {
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
