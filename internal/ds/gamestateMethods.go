package ds

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"zuch-backend/internal/utils"
)

//Außer Cargopathfinding

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

// nur intern
// Entfernt die Station
// nicht mehr alle Referenzen (Stops, Plattform Tiles, etc)
func (gs *GameState) removeStation(s *Station) error {
	var err error
	before := len(gs.Stations)
	delete(gs.Stations, s.Id)
	if !(before > len(gs.Stations)) {
		return fmt.Errorf("couldn't find station in map")
	}
	//Enfernung des Plattform Tags aller Tiles der Station, Löschung der Plattformen findet beim Löschenn des jeweils letzten Tiles statt
	/*for _, plattform := range s.Plattforms {
		for _, tilePos := range plattform.Tiles {
			ChangeStationTile(true, tilePos, gs)
		}
	}*/
	return err
}

func (gs *GameState) AddStationTile(position [2]int) (*Station, error) {
	return gs.changeStationTile(false, position)
}
func (gs *GameState) RemoveStationTile(position [2]int) (*Station, error) {
	return gs.changeStationTile(true, position)
}

// remove = true -> referenz wird entfernt, = false -> wird hinzugefügt;
// findet alle Aktiven Tiles im validen Radius um das neue Tile der Station und verknüpft/entfernt entsprechend die Station, falls noch nicht geschehen
// egal ob neue Station oder nicht
// baut nur neue Staion, baut keine Schienen
// TODO dynamische Zuweisung zu Gleisen anhand Ausrichtung und nachbarn
func (gs *GameState) changeStationTile(remove bool, position [2]int) (*Station, error) {
	var err error
	tile := gs.Tiles[position[0]][position[1]]
	var station *Station //die Station des Tiles, wird bestimmt

	//überprüft, ob das Tile gerade ist
	var horizontal bool
	if tile.Tracks[0] && tile.Tracks[2] && !tile.Tracks[1] && !tile.Tracks[3] {
		horizontal = true
	} else if tile.Tracks[1] && tile.Tracks[3] && !tile.Tracks[0] && !tile.Tracks[2] {
		horizontal = false
	} else {
		//TODO error
		// ich glaube der fehler hier ist wenn da noch keine gleise liegen?
		return nil, fmt.Errorf("Keine Gleise vorhanden oder nicht nur horizonale oder vertikale Gleise")
	}

	// Ich war mal so frech - Jannis
	// Ich habe das mal für dich ausformuliert - Wilken
	if remove {
		if !tile.IsPlattform {
			//TODO Error, kann nicht von nichts entfernen
			return nil, nil
		}

		//Bestimmung der Plattform
		var plattform *Plattform
		plattform, err = gs.GetPlattform(position)
		if err != nil {
			return nil, err
		}

		station = plattform.GetStation(gs)

		//verkleinerung der max. Kapazität
		station.Capacity -= gs.CapacityPerStationTile

		//Entfernung des Tags und weitere Berechnungen
		err = plattform.removeTile(position, gs)
		if err != nil {
			return nil, err
		}

		//gucken
		if len(plattform.Tiles) == 0 {
			//Plattform löschen
			station.removePlattform(plattform.Id, gs)
		}

	} else {

		//berührt das Tile eine Station? Wenn nein, dann Id = 0
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
				//wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					//TODO Error, grenzt an 2 Stationen an
					return nil, nil
				}
				stationBordering = temp.GetStation(gs)
			}
		}
		if position[0] < gs.SizeX-1 {
			right = gs.Tiles[position[0]+1][position[1]]
			if right.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0] + 1, position[1]})
				if err != nil {
					return nil, err
				}
				//wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					//TODO Error, grenzt an 2 Stationen an
					return nil, nil
				}
				stationBordering = temp.GetStation(gs)
			}
		}
		if position[1] < gs.SizeY-1 {
			under = gs.Tiles[position[0]][position[1]+1]
			if under.IsPlattform {
				var temp *Plattform
				temp, err = gs.GetPlattform([2]int{position[0], position[1] + 1})
				if err != nil {
					return nil, err
				}
				//wenn eine andere Station schon angegrenzt, dann Fehler
				if stationBordering != nil && stationBordering.Id != temp.GetStation(gs).Id {
					//TODO Error, grenzt an 2 Stationen an
					return nil, nil
				}
				stationBordering = temp.GetStation(gs)
			}
		}

		if stationBordering == nil {
			//neue Station mit Standartwerten
			station, err = gs.addStation("")
			if err != nil {
				return nil, err
			}
			gs.Tiles[position[0]][position[1]].IsPlattform = true
		} else {
			station = stationBordering
		}
		//es gibt eine Station. Da ?sicher ein Tile hinzugefügt wird, vergrößere den Platz
		station.Capacity += gs.CapacityPerStationTile

		// Nun wird das Gleis bestimmt
		gs.Tiles[position[0]][position[1]].IsPlattform = true

		//Gleis bestimmen und hinzufügen
		if horizontal {
			var first bool //ist linker eine Station? -> wenn beide Seiten, dann Fehler
			var last bool

			if position[0] > 0 {
				//wenn es Plattform ist, muss ausschließlich bei 0,2 Treacks sein
				if left.Tracks[0] && left.IsPlattform {
					first = true
				}
			}
			var plattform *Plattform //die Plattform, wenn eine angrenzt

			//nach rechts gucken
			if position[0] < gs.SizeX-1 {

				//ist rechter eine Plattfrom und richtig ausgerichtet?
				if right.Tracks[0] && right.IsPlattform {
					if first {
						//TODO Error, grenzt an 2 Gleise an
						return nil, nil
					}
					last = true
				}
			}

			if first {
				//grenzt nur links an
				plattform, err = gs.GetPlattform([2]int{position[0] - 1, position[1]})
				if err != nil {
					return nil, err
				}
				//hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)

			} else if last {
				//grenzt nur rechts an
				plattform, err = gs.GetPlattform([2]int{position[0] + 1, position[1]})
				if err != nil {
					return nil, err
				}
				//hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				// TODO neues Gleis
				err = station.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil, nil
				}
				gs.Tiles[position[0]][position[1]].IsPlattform = true
			}

		} else {
			//vertikal

			var first bool //ist oben eine Station? -> wenn beide Seiten, dann Fehler
			var last bool

			if position[1] > 0 {
				//wenn es Plattform ist, muss ausschließlich bei 1,3 Treacks sein
				if above.Tracks[1] && above.IsPlattform {
					first = true
				}
			}
			var plattform *Plattform //die Plattform, wenn eine angrenzt

			//nach unten gucken
			if position[1] < gs.SizeY-1 {

				//ist unten eine Plattfrom und richtig ausgerichtet?
				if under.Tracks[1] && under.IsPlattform {
					if first {
						//TODO Error, grenzt an 2 Gleise an
						return nil, nil
					}
					last = true
				}
			}
			if first {
				//grenzt nur links an
				plattform, err = gs.GetPlattform([2]int{position[0], position[1] - 1})
				if err != nil {
					return nil, err
				}
				//hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)
			} else if last {
				//grenzt nur rechts an
				plattform, err = gs.GetPlattform([2]int{position[0], position[1] + 1})
				if err != nil {
					return nil, err
				}
				//hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				// TODO neues Gleis
				err = station.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil, nil
				}
				gs.Tiles[position[0]][position[1]].IsPlattform = true
			}
		}

	}

	//Referenzen zu activeTiles aktualisieren

	//minimum bestimmen, maximum wird in schleife geprüft, dass nicht out of bounds
	xMin := position[0] - gs.StationRange
	if xMin < 0 {
		xMin = 0
	}
	yMin := position[1] - gs.StationRange
	if yMin < 0 {
		yMin = 0
	}

	//durchiteriren durch alle Tiles in Reichweite und innerhalb der Bounds
	for y := yMin; y <= yMin+(gs.StationRange*2) && y < len(gs.Tiles); y++ {
		for x := xMin; x <= xMin+(gs.StationRange*2) && x < len(gs.Tiles[0]); x++ {
			//jedes Aktive Tile braucht eine Category, muss also nicht existent sein, wenn keine Referenz darauf auf eine gibt
			//daran erkenne ich jetzt, dass in dem Tile ein aktives Tile ist
			if gs.Tiles[x][y].ActiveTile.Category != nil {
				tile := gs.Tiles[x][y]

				//gucken, ob die Station schon hinterlegt ist
				stationFound := false
				for i := 0; i < len(tile.ActiveTile.Stations); i++ {
					//wenn aktuelle Station schon referenziert ist
					// muss über Id, weil sonst Pointer verglichen werden und die Objekte kann man nciht vergleichen wegen map
					if tile.ActiveTile.Stations[i].Id == station.Id {
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
					gs.Tiles[x][y].ActiveTile.Stations = append(gs.Tiles[x][y].ActiveTile.Stations, station)
					// continue
				}
			}
		}
	}

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

// -------------------------------------- ZÜGE ---------------------------------------------

//vielleicht noch woanders hin

// nur für Beladen von Zügen
// finde die CargoCategory des CargoTypes
func (gs *GameState) getCargoCategory(cargoType string) string {
	//iteriere die CargoCategorys
	for key, value := range gs.ConfigData.TrainCategories {
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

// Fügt einen Zug hinzu, anhand eines namens und der position sowie des typen und positionen der waggons
func (gs *GameState) AddTrain(name string, Waggons []TrainCreateWaggons) (*Train, error) {

	// Weg muss ja frei sein, und alles müssen zusammenhängen
	err := gs.checkIfWaggonsAreValid(Waggons)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = fmt.Sprint(gs.CurrentTrainID.Load())
	}

	train := &Train{Name: name, Id: int(gs.CurrentTrainID.Load())}
	gs.CurrentTrainID.Add(1)

	for _, waggon := range Waggons {
		err := train.AddWaggon(waggon.Position, waggon.Typ, gs)
		if err != nil {
			return nil, fmt.Errorf("this shoudln't happen; %s", err.Error())
		}
	}

	gs.Trains[train.Id] = train
	return train, nil
}

// Überprüft ob alle waggons eine Valide Position haben, also das Gleis nicht blockiert ist, das gleis existiert und die Waggons zusammenhängend sind
func (gs *GameState) checkIfWaggonsAreValid(waggons []TrainCreateWaggons) error {
	for i, waggon := range waggons {
		if gs.Tiles[waggon.Position[0]][waggon.Position[1]].IsBlocked {
			return fmt.Errorf("track is blocked")
		}
		// Schaut ob der n. Waggon ein Nachbar des n-1. Waggon ist, daher beim 0. Überspringen
		if i != 0 {
			prevWaggon := waggons[i-1]
			possibleTracks := gs.neighbourTracks(prevWaggon.Position[0], prevWaggon.Position[1], prevWaggon.Position[2])

			if !slices.Contains(possibleTracks, waggon.Position) {
				return fmt.Errorf("waggons are not continuos or a track is missing")
			}
		}
	}
	// Gibt einen Fehler zurück falls
	return nil
}

// nur für Main, hier eigentlich gut, oder?
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
		gs.BroadcastChannel <- WsEnvelope{Type: "tiles.unblock", Username: "kaputt", Msg: BlockedTilesMSG{Tiles: tilesToUnblock}}
	}
}

// entfernt diesen zug, dazu wird er aus dem array genommen und sein currentPath wird auf nicht blockiert gesetzt, hoffe das passt so
// fehler sind bisher ungenutzt, irgendwas wirrd schon schiefgehen
func (gs *GameState) RemoveTrain(t *Train) error {

	var blockedTilesPositions [][2]int
	for _, v := range t.CurrentPath {
		gs.Tiles[v[0]][v[1]].IsBlocked = false // ich hab keine ahnung ob das so geht
		blockedTilesPositions = append(blockedTilesPositions, [2]int{v[0], v[1]})
	}
	// Das kann hier gut sein das da zeugs doppelt drinne ist aber das ist mir spontan egal, doppelt auf false setzen hält ohnehin besser
	gs.BroadcastChannel <- WsEnvelope{Type: "tiles.unblock", Username: "Server", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}

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
	return &s, nil
}

// Löscht alle Stops und danach den Schedule aus der Map
func (gs *GameState) RemoveSchedule(Id int) error {
	schedule := gs.Schedules[Id]

	if schedule == nil {
		return fmt.Errorf("couldn't find schedule in map")
	}

	for _, stop := range schedule.Stops {
		schedule.RemoveStop(stop.Id, gs)
	}

	delete(gs.Schedules, Id)

	return nil
}

//---------------------------------------------- Tiles -----------------------------------------

// fügt bei allen SubTiles zwischen den beiden SubTiles, die inklusive, wenn möglich eine Schiene ein
func (gs *GameState) AddTracks(startSubTile [3]int, endSubTile [3]int) error {

	gs.editTracks(startSubTile, endSubTile, "Error while building tracks.", func(gs *GameState, coordinate [3]int) error {
		tile, error := gs.GetTile(coordinate[0], coordinate[1])
		if error != nil {
			gs.Logger.Error("Out of Bound, voll Scheiße, sollte nicht gehen.")
			return fmt.Errorf("Scheiße, läuft gar nicht gut")
		}
		return tile.AddTrack(coordinate[2], gs)
	})

	return nil
}

func (gs *GameState) RemoveTracks(startSubTile [3]int, endSubTile [3]int) error {

	gs.editTracks(startSubTile, endSubTile, "Error while removing tracks.", func(gs *GameState, coordinate [3]int) error {
		tile, error := gs.GetTile(coordinate[0], coordinate[1])
		if error != nil {
			gs.Logger.Error("Out of Bound, voll Scheiße, sollte nicht gehen.")
			return fmt.Errorf("Scheiße, läuft gar nicht gut")
		}
		//return tile.AddTrack(coordinate[2], gs)
		return tile.RemoveTrack(coordinate[2], gs)
	})

	return nil
}

// führt für alle Subtiles zwischen den beiden angebenen die Methode aus, inklusive derer.
// Kontrolliert, ob die in einer Reihe sind und ob die in Bonds sind.
// ob die Aktion durchgeführt werden kann sollte in der Methode überprüft werden, die in der übergebenen Methode aufgerufen wird.
func (gs *GameState) editTracks(startSubTile [3]int, endSubTile [3]int, defaultError string, methodForEach func(gs *GameState, coordinate [3]int) error) error {

	sst := startSubTile
	est := endSubTile

	//sind die y dann in range?
	_, error := gs.GetTile(sst[0], sst[1])
	if error != nil {
		return fmt.Errorf("%s", defaultError+" The start koordinate have the problem: "+error.Error())
	}
	_, error = gs.GetTile(est[0], est[1])
	if error != nil {
		return fmt.Errorf("%s", defaultError+" The end koordinate have the problem: "+error.Error())
	}

	//liste der, die nicht hinzugefügt werden konnten
	notEdit := []string{"Could not edit the following subtiles because of the following reasons: "}

	//sind sie in einer linie (ggf. als einzelne Methode)
	// sind die Subtiles beide horizontal?
	if (sst[2] == 1 || sst[2] == 3) && (est[2] == 1 || est[2] == 3) {
		//sind sie auf einer y?
		if sst[1] != est[1] {
			return fmt.Errorf("%s", defaultError+" The Subtiles are horizontal, but the y koordinates differs.")
		}

		//edititeren der Tiles
		curTile := sst
		countUp := (sst[0] == est[0] && sst[2] < est[2]) || sst[0] < est[0] // ob sst links von est ist
		for curTile[0] != est[0] || curTile[1] != est[1] || curTile[2] != est[2] {

			//Methode auf Tile anwenden, signal,
			error := methodForEach(gs, curTile)

			if error != nil {
				notEdit = append(notEdit, "("+strconv.Itoa(curTile[0])+", "+strconv.Itoa(curTile[1])+", "+strconv.Itoa(curTile[2])+") "+error.Error())
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

		//sind sie vertikal?
	} else if (sst[2] == 2 || sst[2] == 4) && (est[2] == 2 || est[2] == 4) {
		//sind sie auf einer x?
		if sst[0] != est[0] {
			return fmt.Errorf("%s", defaultError+" The Subtiles are vertikal, but the x koordinates differ")
		}

		//jedes Sub-tile durchiterieren
		curTile := sst
		countUp := (sst[1] == est[1] && sst[2] < est[2]) || sst[1] < est[1] // ob sst unter est ist
		for curTile[0] != est[0] || curTile[1] != est[1] || curTile[2] != est[2] {

			error := methodForEach(gs, curTile)

			if error != nil {
				notEdit = append(notEdit, "("+strconv.Itoa(curTile[0])+", "+strconv.Itoa(curTile[1])+", "+strconv.Itoa(curTile[2])+") "+error.Error())
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
		//sind beides nicht
		return fmt.Errorf("%s", defaultError+" Sub-Tiles don't allign. They both have to be horizntal or vertikal")
	}

	if len(notEdit) > 0 {
		return fmt.Errorf("%s", strings.Join(notEdit, "\n"))
	}

	return nil
}
