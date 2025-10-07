package main

import (
	"slices"
)

type Station struct {
	Id       int //muss unique sein
	Name     string
	Storage  map[string]int //von was ist wieviel gelagert: CargoType? Hinzufügen nur mit Methode, rausnehmen direkt. MUSS mit make komplett initiiert werden
	Capacity int            //was ist die maximale Kapazität
	User     *User
}

// remove = true -> referenz wird entfernt, = false -> wird hinzugefügt;
// findet alle Aktiven Tiles im validen Radius um das neue Tile der Station und verknüpft/entfernt entsprechend die Station, falls noch nicht geschehen
// egal ob neue Station oder nicht
func (s Station) changeStationTile(remove bool, position [2]int) {
	//minimum bestimmen, maximum wird in schleife geprüft, dass nicht out of bounds
	xMin := position[0] - stationRange
	if xMin < 0 {
		xMin = 0
	}
	yMin := position[1] - stationRange
	if yMin < 0 {
		yMin = 0
	}

	// Ich war mal so frech - Jannis
	if remove {
		tiles[position[0]][position[1]].IsPlattform = false
	} else {
		tiles[position[0]][position[1]].IsPlattform = true

	}

	//durchiteriren durch alle Tiles in Reichweite und innerhalb der Bounds
	for y := yMin; y <= yMin+(stationRange*2) && y < len(tiles); y++ {
		for x := xMin; x <= xMin+(stationRange*2) && x < len(tiles[0]); x++ {
			//jedes Aktive Tile braucht eine Category, muss also nicht existent sein, wenn keine Referenz darauf auf eine gibt
			//daran erkenne ich jetzt, dass in dem Tile ein aktives Tile ist
			if tiles[x][y].ActiveTile.Category != nil {
				tile := *tiles[x][y]

				//gucken, ob die Station schon hinterlegt ist
				stationFound := false
				for i := 0; i < len(tile.ActiveTile.Stations); i++ {
					//wenn aktuelle Station schon referenziert ist
					// muss über Id, weil sonst Pointer verglichen werden und die Objekte kann man nciht vergleichen wegen map
					if tile.ActiveTile.Stations[i].Id == s.Id {
						stationFound = true
						if remove {
							tile.ActiveTile.Stations = removeElementFromSlice(tile.ActiveTile.Stations, i)
						}
						break
					}
				}
				// wenn die Station nicht gefunden wurde und hinzugefügt werden soll, hinzufügen der Station
				if !stationFound && !remove {
					tiles[x][y].ActiveTile.Stations = append(tiles[x][y].ActiveTile.Stations, &s)
					// continue
				}
			}
		}
	}
}

// return Restwert, der keinen Platz gefunden hat
func (s *Station) addCargo(cargoType string, quantity int) int {
	filled := s.getFillLevel()
	//ist noch platz?
	if filled < s.Capacity {
		//wie viel Platz?
		emtySpace := s.Capacity - filled
		overflow := quantity - emtySpace
		//reicht der Platz?, dann gibt es keinen Overflow
		if overflow < 0 {
			overflow = 0
		}
		//fülle auf
		s.Storage[cargoType] += quantity - overflow
		return overflow
	}
	return quantity
}

// returnt den Füllstand der Station
func (s Station) getFillLevel() int {
	var filled int //aktueller gefüllter Wert
	for _, i := range s.Storage {
		filled += i
	}
	return filled
}

//-------------------------------------------------------------------------------------

// ist einzigartig durch station & name zusammen (ggf. id?)
type Plattform struct {
	Name    string
	Tiles   [][2]int // nur x & y, muss in Order sein. KEIN DIREKTER ZUGRIFF?
	station *Station
}

// returns die erste und letzte Sub-Tile Koordinate
func (p *Plattform) getFirstLast() [2][3]int {
	var r [2][3]int
	r[0] = [3]int{p.Tiles[0][0], p.Tiles[0][1]}
	r[1] = [3]int{p.Tiles[len(p.Tiles)-1][0], p.Tiles[len(p.Tiles)-1][1]}

	//wenn die bei y gleich sind, dann ist die Plattform horizontal
	if p.Tiles[0][1] == p.Tiles[len(p.Tiles)-1][1] {
		//wenn erstes Tile weiter links als das zweite ist, dann ist der Äußere Sub-Tile 1, sont 3
		//fürs letzte Tile entsprechend andersherum
		if p.Tiles[0][0] < p.Tiles[len(p.Tiles)-1][0] {
			r[0][2] = 1
			r[1][2] = 3
		} else {
			r[0][2] = 3
			r[1][2] = 1
		}
	} else {
		//hier das ganze für oben und unten
		if p.Tiles[0][1] < p.Tiles[len(p.Tiles)-1][1] {
			r[0][2] = 2
			r[1][2] = 4
		} else {
			r[0][2] = 4
			r[1][2] = 2
		}
	}

	return r
}

// Weg zum anderen Ende ohne initial
// damit die Züge in der Mitte halten, muss das richtige Tile anhand von Zuglänge bestimmt werden, wenn gewollt
func (p *Plattform) getPathToOpposite(initial [3]int) [][3]int {
	var r [][3]int

	//wenn die bei y gleich sind, dann ist die Plattform horizontal
	if p.Tiles[0][1] == p.Tiles[len(p.Tiles)-1][1] {
		//wenn erstes Tile weiter links als das zweite ist, dann ist der Äußere Sub-Tile 1, sont 3
		//fürs letzte Tile entsprechend andersherum
		if p.Tiles[0][0] < p.Tiles[len(p.Tiles)-1][0] {
			for i := range (p.Tiles[len(p.Tiles)-1][0] - p.Tiles[0][0] + 1) * 2 {
				if i%2 == 0 {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 1})
				} else {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 3})
				}
			}
		} else {
			for i := range (p.Tiles[0][0] - p.Tiles[len(p.Tiles)-1][0] + 1) * 2 {
				if i%2 == 0 {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 3})
				} else {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 1})
				}
			}
		}
	} else {
		if p.Tiles[0][1] < p.Tiles[len(p.Tiles)-1][1] {
			for i := range (p.Tiles[len(p.Tiles)-1][1] - p.Tiles[0][1] + 1) * 2 {
				if i%2 == 0 {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 2})
				} else {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 4})
				}
			}
		} else {
			for i := range (p.Tiles[0][1] - p.Tiles[len(p.Tiles)-1][1] + 1) * 2 {
				if i%2 == 0 {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 4})
				} else {
					r = append(r, [3]int{p.Tiles[i/2][0], p.Tiles[i/2][1], 2})
				}
			}
		}
	}

	//reverse if neccececary
	if r[0] != initial {
		slices.Reverse(r)
	}
	//erste rausnehmen, da das schon im Pfad ist
	r = r[1:]

	return r
}
