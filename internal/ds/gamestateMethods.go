package ds

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"zuch-backend/internal/utils"
)

// Außer Cargopathfinding

// region Stations
// ------------------------------------- stations -------------------------------------------

// nur intern
// nur Erstellung einer Station, hinzufügen der Tiles muss extra gemacht werden. Id ist Standardname
func (gs *GameState) addStation(name string) (*Station, error) {
	if name == "" {
		name = fmt.Sprint(gs.CurrentStationID.Load())
	}

	station := &Station{Id: int(gs.CurrentStationID.Load()), Name: name, Storage: make(map[string]int), Plattforms: make(map[int]*Plattform)}
	gs.Stations[int(gs.CurrentStationID.Load())] = station
	gs.CurrentStationID.Add(1)
	return station, nil // ich glaube bisher kann hier kein fehler kommen? surelly jaja
}

// nur intern, entfernt nur die Stationen, die keine Plattformen mehr haben
func (gs *GameState) deleteStation(s *Station) error {
	var err error
	before := len(gs.Stations)
	delete(gs.Stations, s.Id)
	if !(before > len(gs.Stations)) {
		return fmt.Errorf("couldn't find station in map")
	}
	return err
}

// löscht alle Tiles der Station und damit auch alle Plattformen und die Station selber
func (gs *GameState) RemoveStation(s *Station) error {
	// Enfernung des Plattform Tags aller Tiles der Station, Löschung der Plattformen findet beim Löschen des jeweils letzten Tiles statt
	for _, plattform := range s.Plattforms {
		err := s.RemovePlattform(plattform.Id, gs)
		if err != nil {
			return err
		}

	}

	gs.BroadcastChannel <- WsEnvelope{Type: "station.remove", Msg: s.Id}
	// ich probiere das einfach mal
	return nil
}

// Fügt an der Position ein Station Tile ein, kümmert sich um Erstellen Plattform und Station. Gibt die Station zurück, der das Tile hinzugefügt wurde.
func (gs *GameState) AddStationTile(position [2]int) (*Station, error) {
	return gs.changeStationTile(false, position)
}

func (gs *GameState) AddStationTiles(positionStart [2]int, positionEnd [2]int) error {
	return gs.changeStationTiles(false, positionStart, positionEnd)
}

// Löscht an der Position das Station Tile, kümmert sich um Löschen der  Plattform und Station.
func (gs *GameState) RemoveStationTile(position [2]int) (*Station, error) {
	return gs.changeStationTile(true, position)
}

func (gs *GameState) RemoveStationTiles(positionStart [2]int, positionEnd [2]int) error {
	return gs.changeStationTiles(true, positionStart, positionEnd)
}

func (gs *GameState) changeStationTiles(remove bool, positionStart [2]int, positionEnd [2]int) error {
	if positionStart[1] == positionEnd[1] {
		// horizontal oder gleich

		countUp := positionStart[0] <= positionEnd[0]

		for {
			_, err := gs.changeStationTile(remove, positionStart)
			if err != nil {
				return err
			}
			if positionStart == positionEnd {
				break
			}
			if countUp {
				positionStart[0]++
			} else {
				positionStart[0]--
			}
		}
	} else if positionStart[0] == positionEnd[0] {
		// vertikal

		countUp := positionStart[1] <= positionEnd[1]

		for {
			_, err := gs.changeStationTile(remove, positionStart)
			if err != nil {
				return err
			}
			if positionStart == positionEnd {
				break
			}
			if countUp {
				positionStart[1]++
			} else {
				positionStart[1]--
			}
		}
	} else {
		return fmt.Errorf("Please provide coordinates that are in one line")
	}
	return nil
}

// remove = true -> referenz wird entfernt, = false -> wird hinzugefügt;
// findet alle Aktiven Tiles im validen Radius um das neue Tile der Station und verknüpft/entfernt entsprechend die Station, falls noch nicht geschehen
// egal ob neue Station oder nicht
// baut nur neue Staion, baut keine Schienen
func (gs *GameState) changeStationTile(remove bool, position [2]int) (*Station, error) {
	var err error

	// sind die y dann in range?
	_, error := gs.GetTile(position[0], position[1])
	if error != nil {
		return nil, fmt.Errorf("%s", " The koordinate has the problem: "+error.Error())
	}

	tile := gs.Tiles[position[0]][position[1]]
	var station *Station // die Station des Tiles, wird bestimmt

	// überprüft, ob das Tile gerade ist
	var horizontal bool
	if tile.Tracks[0] && tile.Tracks[2] && !tile.Tracks[1] && !tile.Tracks[3] {
		horizontal = true
	} else if tile.Tracks[1] && tile.Tracks[3] && !tile.Tracks[0] && !tile.Tracks[2] {
		horizontal = false
	} else {
		// ich glaube der fehler hier ist wenn da noch keine gleise liegen?
		return nil, fmt.Errorf("Keine Gleise vorhanden oder nicht nur horizonale oder vertikale Gleise")
	}

	// Ich war mal so frech - Jannis
	// Ich habe das mal für dich ausformuliert - Wilken
	if remove {
		if !tile.IsPlattform {
			return nil, fmt.Errorf("Tile gehört nicht zu einer Station, kann nicht entfernt werden.")
		}

		// Bestimmung der Plattform
		var plattform *Plattform
		plattform, err = gs.GetPlattform(position)
		if err != nil {
			return nil, err
		}

		station = plattform.GetStation(gs)

		// verkleinerung der max. Kapazität
		station.Capacity -= gs.ConfigData.CapacityPerStationTile

		// Entfernung des Tags und weitere Berechnungen
		err = plattform.removeTile(position, gs)
		if err != nil {
			return nil, err
		}

		// gucken
		if len(plattform.Tiles) == 0 {
			// Plattform löschen
			station.deletePlattform(plattform.Id, gs)
		}

	} else {

		tile := gs.Tiles[position[0]][position[1]]

		if tile.IsPlattform {
			return nil, fmt.Errorf("Could not build a station there since there is one already.")
		}

		for _, signal := range tile.Signals {
			if signal {
				return nil, fmt.Errorf("Could not build a station there because there is a signal on that tile already.")
			}
		}

		// berührt das Tile eine Station? Wenn nein, dann Id = 0
		var stationBordering *Station
		var above, under, left, right *Tile
		if position[0] > 0 {
			left = gs.Tiles[position[0]-1][position[1]]
			if left.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0] - 1, position[1]})
				if err != nil {
					return nil, err
				}

				stationBordering = temp.GetStation(gs)
			}
		}
		if position[1] > 0 {
			above = gs.Tiles[position[0]][position[1]-1]
			if above.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0], position[1] - 1})
				if err != nil {
					return nil, err
				}
				// wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					return nil, fmt.Errorf("Could not build a Station there, the tile is bordering at least two stations.")
				}
				stationBordering = temp.GetStation(gs)
			}
		}
		if position[0] < gs.ConfigData.SizeX-1 {
			right = gs.Tiles[position[0]+1][position[1]]
			if right.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0] + 1, position[1]})
				if err != nil {
					return nil, err
				}
				// wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					return nil, fmt.Errorf("Could not build a Station there, the tile is bordering two stations.")
				}
				stationBordering = temp.GetStation(gs)
			}
		}
		if position[1] < gs.ConfigData.SizeY-1 {
			under = gs.Tiles[position[0]][position[1]+1]
			if under.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0], position[1] + 1})
				if err != nil {
					return nil, err
				}
				// wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					return nil, fmt.Errorf("Could not build a Station there, the tile is bordering two stations.")
				}
				stationBordering = temp.GetStation(gs)
			}
		}

		if stationBordering == nil {
			// neue Station mit Standartwerten
			station, err = gs.addStation("")
			if err != nil {
				return nil, err
			}
			gs.Tiles[position[0]][position[1]].IsPlattform = true
		} else {
			station = stationBordering
		}
		// es gibt eine Station. Da ?sicher ein Tile hinzugefügt wird, vergrößere den Platz
		station.Capacity += gs.ConfigData.CapacityPerStationTile

		// Nun wird das Gleis bestimmt
		gs.Tiles[position[0]][position[1]].IsPlattform = true

		// Gleis bestimmen und hinzufügen
		if horizontal {
			var first bool // ist linker eine Station? -> wenn beide Seiten, dann Fehler
			var last bool

			if position[0] > 0 {
				// wenn es Plattform ist, muss ausschließlich bei 0,2 Treacks sein
				if left.Tracks[0] && left.IsPlattform {
					first = true
				}
			}
			var plattform *Plattform // die Plattform, wenn eine angrenzt

			// nach rechts gucken
			if position[0] < gs.ConfigData.SizeX-1 {
				// ist rechter eine Plattfrom und richtig ausgerichtet?
				if right.Tracks[0] && right.IsPlattform {
					if first {
						return nil, fmt.Errorf("Could not build a Station there, the tile is bordering two stations.")
					}
					last = true
				}
			}

			if first {
				// grenzt nur links an
				plattform, err = gs.GetPlattform([2]int{position[0] - 1, position[1]})
				if err != nil {
					return nil, err
				}
				// hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)

			} else if last {
				// grenzt nur rechts an
				plattform, err = gs.GetPlattform([2]int{position[0] + 1, position[1]})
				if err != nil {
					return nil, err
				}
				// hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				err = station.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil, nil
				}
				gs.Tiles[position[0]][position[1]].IsPlattform = true
			}

		} else {
			// vertikal

			var first bool // ist oben eine Station? -> wenn beide Seiten, dann Fehler
			var last bool

			if position[1] > 0 {
				// wenn es Plattform ist, muss ausschließlich bei 1,3 Treacks sein
				if above.Tracks[1] && above.IsPlattform {
					first = true
				}
			}
			var plattform *Plattform // die Plattform, wenn eine angrenzt

			// nach unten gucken
			if position[1] < gs.ConfigData.SizeY-1 {
				// ist unten eine Plattfrom und richtig ausgerichtet?
				if under.Tracks[1] && under.IsPlattform {
					if first {
						return nil, fmt.Errorf("Could not build a Station there, the tile is bordering two stations.")
					}
					last = true
				}
			}
			if first {
				// grenzt nur links an
				plattform, err = gs.GetPlattform([2]int{position[0], position[1] - 1})
				if err != nil {
					return nil, err
				}
				// hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)
			} else if last {
				// grenzt nur rechts an
				plattform, err = gs.GetPlattform([2]int{position[0], position[1] + 1})
				if err != nil {
					return nil, err
				}
				// hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				err = station.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil, nil
				}
				gs.Tiles[position[0]][position[1]].IsPlattform = true
			}
		}

	}

	// Referenzen zu activeTiles aktualisieren

	// minimum bestimmen, maximum wird in schleife geprüft, dass nicht out of bounds
	xMin := position[0] - gs.ConfigData.StationRange
	if xMin < 0 {
		xMin = 0
	}
	yMin := position[1] - gs.ConfigData.StationRange
	if yMin < 0 {
		yMin = 0
	}

	// durchiteriren durch alle Tiles in Reichweite und innerhalb der Bounds
	for y := yMin; y <= yMin+(gs.ConfigData.StationRange*2) && y < len(gs.Tiles); y++ {
		for x := xMin; x <= xMin+(gs.ConfigData.StationRange*2) && x < len(gs.Tiles[0]); x++ {
			// jedes Aktive Tile braucht eine Category, muss also nicht existent sein, wenn keine Referenz darauf auf eine gibt
			// daran erkenne ich jetzt, dass in dem Tile ein aktives Tile ist
			if gs.Tiles[x][y].ActiveTile.Category != nil {
				tile := gs.Tiles[x][y]

				// gucken, ob die Station schon hinterlegt ist
				stationFound := false
				for i := 0; i < len(tile.ActiveTile.Stations); i++ {
					// wenn aktuelle Station schon referenziert ist
					// muss über Id, weil sonst Pointer verglichen werden und die Objekte kann man nciht vergleichen wegen map
					if tile.ActiveTile.Stations[i] == station.Id {
						stationFound = true
						if remove {
							tile.ActiveTile.Stations, err = utils.RemoveElementFromSlice(tile.ActiveTile.Stations, i)
							if err != nil {
								return nil, err
							}
						}
						break
					}
				}
				// wenn die Station nicht gefunden wurde und hinzugefügt werden soll, hinzufügen der Station
				if !stationFound && !remove {
					gs.Tiles[x][y].ActiveTile.Stations = append(gs.Tiles[x][y].ActiveTile.Stations, station.Id)
					// continue
				}
			}
		}
	}

	gs.BroadcastChannel <- WsEnvelope{Type: "station.update", Msg: station}
	gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: gs.Tiles[position[0]][position[1]]}
	return station, nil
}

// nur für changeStationTile (und demo)
// gibt die Plattform zurück, zu der ein Tile gehört
func (gs *GameState) GetPlattform(position [2]int) (*Plattform, error) {
	for _, station := range gs.Stations {
		for _, plattform := range station.Plattforms {
			for _, tile := range plattform.Tiles {
				if tile[0] == position[0] && tile[1] == position[1] {
					return plattform, nil
				}
			}
		}
	}

	return nil, nil
}

// GetPlattformByID Gibt einem die Plattform mit der jeweiligen ID dA ich zu faul bin die in eine map zu wandeln
func (gs *GameState) GetPlattformByID(id int) (*Plattform, error) {
	var res *Plattform = nil
	for _, station := range gs.Stations {
		for _, plattform := range station.Plattforms {
			if plattform.Id == id {
				res = plattform
			}
		}
	}
	if res != nil {
		return res, nil
	} else {
		return nil, fmt.Errorf("could not find Plattform")
	}
}

// region Züge
// -------------------------------------- ZÜGE ---------------------------------------------

// vielleicht noch woanders hin

// nur für Beladen von Zügen
// finde die CargoCategory des CargoTypes
func (gs *GameState) getCargoCategory(cargoType string) string {
	// iteriere die CargoCategorys
	for key, value := range gs.ConfigData.TrainCategories {
		// suche in der aktuellen Category nach dem Type
		for _, value2 := range value {
			// wenn gefunden, kann der zurückgegeben werden
			if value2 == cargoType {
				return key
			}
		}
	}
	return ""
}

// fügt Zug hinzu, wenn name leer ist, wird Id genommen.
func (gs *GameState) AddTrain(name string, position [3]int, lokmotive string, level int) (*Train, error) {
	if name == "" {
		name = fmt.Sprint(gs.CurrentTrainID.Load())
	}

	train := &Train{Name: name, Id: int(gs.CurrentTrainID.Load()), Waggons: make([]*Waggon, 0)}

	// Überprüft, ob das Subtile korrekt ist
	err := gs.iterateSubTiles(position, position, "An Error accured while adding a Train.", func(gs *GameState, coordinate [3]int) error { return nil })
	if err != nil {
		return nil, err
	}

	// Fügt die Lock hinzu
	err = train.AddWaggon(position, lokmotive, level, gs)
	if err != nil {
		return nil, err
	}

	gs.Trains[train.Id] = train
	gs.CurrentTrainID.Add(1)

	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: train}

	return train, nil
}

// nur für Main, hier eigentlich gut, oder?
// berechnet alle Bewegungen und Pathfinding
func (gs *GameState) CalculateTrains() {
	// Speichern, welche Tiles am Ende des Threads entblocked werden muss
	var tilesToUnblock [][2]int

	for i := range gs.Trains {
		temp := gs.Trains[i].calculateTrain(gs)
		if temp[0] >= 0 {
			tilesToUnblock = append(tilesToUnblock, temp)
		}
	}

	// entblocken
	for _, i := range tilesToUnblock {
		gs.Tiles[i[0]][i[1]].IsBlocked = false
	}
	if len(tilesToUnblock) > 0 {
		gs.BroadcastChannel <- WsEnvelope{Type: "tiles.unblock", Msg: BlockedTilesMSG{Tiles: tilesToUnblock}}
	}
}

// Funktion, damit die Züge den Belade und Entladeprozess machen.
// ist einzeln, damit es unabhängig von den Fahrtgeschwindigkeiten ist
func (gs *GameState) LoadUndloadTrains() error {
	for _, t := range gs.Trains {
		err := t.loadUnload(gs)
		if err != nil {
			return err
		}
	}

	return nil
}

// entfernt diesen zug, dazu wird er aus dem array genommen und sein currentPath wird auf nicht blockiert gesetzt, hoffe das passt so
// hoffentlich reicht das so, mal gucken
func (gs *GameState) RemoveTrain(t *Train) error {
	var blockedTilesPositions [][2]int
	for _, v := range t.CurrentPath {
		gs.Tiles[v[0]][v[1]].IsBlocked = false // ich hab keine ahnung ob das so geht
		blockedTilesPositions = append(blockedTilesPositions, [2]int{v[0], v[1]})
	}

	// auch die Tiles des Zuges entblocken
	for _, w := range t.Waggons {
		gs.Tiles[w.Position[0]][w.Position[1]].IsBlocked = false
		blockedTilesPositions = append(blockedTilesPositions, [2]int{w.Position[0], w.Position[1]})
	}
	// Das kann hier gut sein das da zeugs doppelt drinne ist aber das ist mir spontan egal, doppelt auf false setzen hält ohnehin besser
	gs.BroadcastChannel <- WsEnvelope{Type: "tiles.unblock", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}
	gs.BroadcastChannel <- WsEnvelope{Type: "train.remove", Msg: TrainRemoveMSG{Id: t.Id}}

	before := len(gs.Trains)
	delete(gs.Trains, t.Id)
	if !(before > len(gs.Trains)) {
		return fmt.Errorf("couldn't find train in map")
	}
	return nil
}

/* fürs Pathfinding und ob Waggons valide sind oder nicht
* erst auf dauerhaft blocked prüfen
* to visit (außer, da wo man hergekommen ist):
*	1 [x][y][2,3,4], [x-1][y][3]
*	2 [x][y][1,3,4], [x][y+1][4]
*	3 [x][y][1,2,4], [x+1][y][1]
*	4 [x][y][1,2,3], [x][y+1][2]
 */
// verified
func (gs *GameState) neighbourTracks(x int, y int, sub int) [][3]int {
	var connectedNeigbours [][3]int // [3]int identifiziert mit x y und subtile ein exaktes Subtile,
	// Alle Koordinaten von Subtiles zurückgeben die angrenzen die können im gleichem subtile oder im angrenzendem

	appending := func(subtilesToCheck [3]int) {
		for i := range 3 {
			o := subtilesToCheck[i]
			if gs.Tiles[x][y].Tracks[o-1] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y, o})
			}
		}
	}

	switch sub {
	case 1:
		if x > 0 {
			if gs.Tiles[x-1][y].Tracks[2] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x - 1, y, 3})
			}
		}
		appending([3]int{2, 3, 4})
	case 2:
		if y > 0 {
			if gs.Tiles[x][y-1].Tracks[3] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y - 1, 4})
			}
		}
		appending([3]int{1, 3, 4})
	case 3:
		if x != len(gs.Tiles)-1 {
			if gs.Tiles[x+1][y].Tracks[0] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x + 1, y, 1})
			}
		}
		appending([3]int{1, 2, 4})
	case 4:
		if y != len(gs.Tiles[0])-1 {
			if gs.Tiles[x][y+1].Tracks[1] {
				connectedNeigbours = append(connectedNeigbours, [3]int{x, y + 1, 2})
			}
		}
		appending([3]int{1, 2, 3})
	}
	return connectedNeigbours
}

// region schedules
// --------------------------------------- Schedules ------------------------------------------

// wenn Name leer ist, wird die Id als Name genommen
func (gs *GameState) AddSchedule(name string) (*Schedule, error) {
	id := int(gs.CurrentScheduleID.Load())
	if name == "" {
		name = fmt.Sprint("Schedule", id)
	}
	s := Schedule{Name: name, Id: id}
	gs.Schedules[int(gs.CurrentScheduleID.Load())] = &s
	gs.CurrentScheduleID.Add(1)
	gs.BroadcastChannel <- WsEnvelope{Type: "schedule.update", Msg: s}
	return &s, nil
}

// Löscht alle Stops und danach den Schedule aus der Map
func (gs *GameState) RemoveSchedule(Id int) error {
	schedule := gs.Schedules[Id]

	if schedule == nil {
		return fmt.Errorf("couldn't find schedule")
	}

	for index := range schedule.Stops {
		schedule.RemoveStop(index+1, gs)
	}

	delete(gs.Schedules, Id)

	gs.BroadcastChannel <- WsEnvelope{Type: "schedule.remove", Msg: Id}
	return nil
}

// gibt die train categories zurück, die aus der Liste korrekt sind. Im Fehler stehen alle, die invalide sind, doppelte werden einfach mitgenommen
// keine Validierung, ob richtige Caracter oder so verwendet wurden, nur ob es die gibt
func (gs *GameState) validateTrainCategories(categories []string) ([]string, error) {
	validCategories := make(map[string]bool)
	error := "The following Train Categories are invalid: "

	for _, category := range categories {
		// suchen der übergebenen Kategorie in den Kategorien der Config, wenn gefunden, dann ist sie valide, wenn nicht, dann Fehler
		for _, trainCategory := range gs.ConfigData.TrainCategories {
			if slices.Contains(trainCategory, category) {
				validCategories[category] = true
			}
		}
		// wenn nicht gefunden, dann Fehler
		if !validCategories[category] {
			error += category + ", "
		}
	}

	r := make([]string, 0, len(validCategories))
	for key := range validCategories {
		r = append(r, key)
	}

	// remove last ", " and return error if there are invalid categories
	if strings.Contains(error, ", ") {
		error = strings.TrimSuffix(error, ", ")
		return r, fmt.Errorf("%s", error)
	}

	return r, nil
}

//---------------------------------------------- Tiles -----------------------------------------

// fügt bei allen SubTiles zwischen den beiden SubTiles, die inklusive, wenn möglich eine Schiene ein
func (gs *GameState) AddTracks(startSubTile [3]int, endSubTile [3]int) error {
	return gs.iterateSubTiles(startSubTile, endSubTile, "Error while building tracks.", func(gs *GameState, coordinate [3]int) error {
		tile, error := gs.GetTile(coordinate[0], coordinate[1])
		if error != nil {
			gs.Logger.Error("Out of Bound, voll Scheiße, sollte nicht gehen.")
			return fmt.Errorf("Scheiße, läuft gar nicht gut")
		}
		return tile.AddTrack(coordinate[2], gs)
	})
}

// entfernt bei allen Sub-Tiles die Schienen
func (gs *GameState) RemoveTracks(startSubTile [3]int, endSubTile [3]int) error {
	return gs.iterateSubTiles(startSubTile, endSubTile, "Error while removing tracks.", func(gs *GameState, coordinate [3]int) error {
		tile, error := gs.GetTile(coordinate[0], coordinate[1])
		if error != nil {
			gs.Logger.Error("Out of Bound, voll Scheiße, sollte nicht gehen.")
			return fmt.Errorf("Scheiße, läuft gar nicht gut")
		}
		// return tile.AddTrack(coordinate[2], gs)
		return tile.RemoveTrack(coordinate[2], gs)
	})
}

// region allgemein
// -------------------------------------------------------------------- allgemein -------------------------------------------------------------------------

// führt für alle Subtiles zwischen den beiden angebenen die Methode aus, inklusive derer.
// Kontrolliert, ob die in einer Reihe sind und ob die in Bonds sind. Kann hoffentlich für Verifizierung einzelner SubTiles verwendet werden.
// ob die Aktion durchgeführt werden kann sollte in der Methode überprüft werden, die in der übergebenen Methode aufgerufen wird.
func (gs *GameState) iterateSubTiles(startSubTile [3]int, endSubTile [3]int, defaultError string, methodForEach func(gs *GameState, coordinate [3]int) error) error {
	sst := startSubTile
	est := endSubTile

	// sind die y und x in range?
	_, error := gs.GetTile(sst[0], sst[1])
	if error != nil {
		return fmt.Errorf("%s", defaultError+" The start koordinate have the problem: "+error.Error())
	}
	_, error = gs.GetTile(est[0], est[1])
	if error != nil {
		return fmt.Errorf("%s", defaultError+" The end koordinate have the problem: "+error.Error())
	}

	// liste der, die nicht hinzugefügt werden konnten
	notEdit := []string{"Could not edit the following subtiles because of the following reasons: "}

	// sind sie in einer linie (ggf. als einzelne Methode)
	// sind die Subtiles beide horizontal?

	// wenn gleich ist, muss die Methode aus der Schleife trotzdem einmal laufen
	if sst == est {
		// Methode auf Tile anwenden
		error := methodForEach(gs, sst)
		if error != nil {
			notEdit = append(notEdit, "("+strconv.Itoa(sst[0])+", "+strconv.Itoa(sst[1])+", "+strconv.Itoa(sst[2])+") "+error.Error())
		}
	} else if (sst[2] == 1 || sst[2] == 3) && (est[2] == 1 || est[2] == 3) {
		// sind sie auf einer y?
		if sst[1] != est[1] {
			return fmt.Errorf("%s", defaultError+" The Subtiles are horizontal, but the y koordinates differs.")
		}

		// edititeren der Tiles
		curTile := sst
		countUp := (sst[0] == est[0] && sst[2] < est[2]) || sst[0] < est[0] // ob sst links von est ist
		for {

			// Methode auf Tile anwenden, signal,
			error := methodForEach(gs, curTile)

			if error != nil {
				notEdit = append(notEdit, "("+strconv.Itoa(curTile[0])+", "+strconv.Itoa(curTile[1])+", "+strconv.Itoa(curTile[2])+") "+error.Error())
			}

			// abbruchbedingung, wenn das betrachtete Subtile das letzte war
			if curTile[0] == est[0] && curTile[2] == est[2] {
				break
			}

			if countUp {
				if curTile[2] == 3 {
					curTile[0] += 1
					curTile[2] = 1
				} else {
					curTile[2] = 3
				}
			} else {
				if curTile[2] == 1 {
					curTile[0] -= 1
					curTile[2] = 3
				} else {
					curTile[2] = 1
				}
			}
		}

		// sind sie vertikal?
	} else if (sst[2] == 2 || sst[2] == 4) && (est[2] == 2 || est[2] == 4) {
		// sind sie auf einer x?
		if sst[0] != est[0] {
			return fmt.Errorf("%s", defaultError+" The Subtiles are vertikal, but the x koordinates differ")
		}

		// jedes Sub-tile durchiterieren
		curTile := sst
		countUp := (sst[1] == est[1] && sst[2] < est[2]) || sst[1] < est[1] // ob sst unter est ist
		for true {

			error := methodForEach(gs, curTile)

			if error != nil {
				notEdit = append(notEdit, "("+strconv.Itoa(curTile[0])+", "+strconv.Itoa(curTile[1])+", "+strconv.Itoa(curTile[2])+") "+error.Error())
			}

			// abbruchbedingung, wenn das betrachtete Subtile das letzt war
			if curTile[1] == est[1] && curTile[2] == est[2] {
				break
			}

			if countUp {
				if curTile[2] == 4 {
					curTile[1] += 1
					curTile[2] = 2
				} else {
					curTile[2] = 4
				}
			} else {
				if curTile[2] == 2 {
					curTile[2] = 4
					curTile[1] -= 1
				} else {
					curTile[2] = 2
				}
			}
		}

	} else {
		// sind beides nicht
		return fmt.Errorf("%s", defaultError+" Sub-Tiles don't allign. They both have to be horizntal or vertikal")
	}

	if len(notEdit) > 1 {
		return fmt.Errorf("%s", strings.Join(notEdit, "\n"))
	}

	return nil
}

// Fügt Geld hinzu
func (gs *GameState) AddMoney(moneyToAdd int) {
	gs.Money += moneyToAdd
	gs.BroadcastChannel <- WsEnvelope{Type: "money.update", Msg: gs.Money}
}

// Überprüft, ob genug Geld da ist und zieht das in dem Fall ab
func (gs *GameState) SubtractMoney(moneyToSubtract int) error {
	err := gs.EnoughMoney(moneyToSubtract)
	if err != nil {
		return err
	}

	gs.Money -= moneyToSubtract

	gs.BroadcastChannel <- WsEnvelope{Type: "money.update", Msg: gs.Money}
	return nil
}

// Überprüft, ob genug Geld da ists
func (gs *GameState) EnoughMoney(moneyToSubtract int) error {
	if gs.Money-moneyToSubtract >= 0 {
		return nil
	}

	return fmt.Errorf("Not enough money.")
}

func (gs *GameState) PauseGame() {
	gs.IsPaused = true

	gs.BroadcastChannel <- WsEnvelope{Type: "game.pause", Msg: true}
	<-gs.ConfirmPause
	gs.Logger.Info("Paused Game")
}

func (gs *GameState) UnPauseGame() {
	gs.BroadcastChannel <- WsEnvelope{Type: "game.unpause", Msg: false}
	gs.UnPause <- true
	gs.Logger.Info("Unpaused Game")
}

func (gs *GameState) LoadConfig() {
	// JSON-Datei öffnen
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal in Struktur
	if err := json.Unmarshal(file, &gs.ConfigData); err != nil {
		panic(err)
	}
}

// kann auch Rechteck und nicht nur Linie, löscht alles, wenn was gelöscht werden kann
func (gs *GameState) ClearTiles(positionStart [2]int, positionEnd [2]int) error {
	// sind die y und x in range?
	_, error := gs.GetTile(positionStart[0], positionStart[1])
	if error != nil {
		return fmt.Errorf("%s", "The start koordinate has the problem: "+error.Error())
	}
	_, error = gs.GetTile(positionEnd[0], positionEnd[1])
	if error != nil {
		return fmt.Errorf("%s", "The end koordinate has the problem: "+error.Error())
	}

	countUpX := positionStart[0] <= positionEnd[0]
	countUpY := positionStart[1] <= positionEnd[1]

	positionCur := positionStart

	for {
		for {

			// Irgnoriert Fehler, weil nur out of bounds und blocked sein kann und out of bound schon getestet wird und blocked ignoeriert werden soll
			gs.ClearTile(positionStart)

			// Abbruchbedingung
			if positionCur[0] == positionEnd[0] {
				positionCur[0] = positionStart[0]
				break
			}

			if countUpX {
				positionCur[0]++
			} else {
				positionCur[0]--
			}
		}

		// Abbruchbedingung
		if positionCur == positionEnd {
			break
		}

		if countUpY {
			positionCur[1]++
		} else {
			positionCur[1]--
		}
	}

	return nil
}

// macht fast alles selber, löscht alles, wenn was gelöscht werden kann
func (gs *GameState) ClearTile(position [2]int) error {
	// sind die y und x in range?
	tile, error := gs.GetTile(position[0], position[1])
	if error != nil {
		return fmt.Errorf("%s", "The koordinate has the problem: "+error.Error())
	}

	if tile.IsBlocked || tile.IsLocked {
		return fmt.Errorf("%s", "Could not clear the Tile since it's locked.")
	}

	_, error = gs.RemoveStationTile(position)
	if error != nil {
		return error
	}

	tile.Tracks = [4]bool{false, false, false, false}
	tile.Signals = [4]bool{false, false, false, false}

	gs.BroadcastChannel <- WsEnvelope{Type: "tile.update", Msg: tile}
	return nil
}
