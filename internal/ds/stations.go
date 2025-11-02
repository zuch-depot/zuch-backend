package ds

import (
	"slices"
	"zuch-backend/internal/utils"
)

type Station struct {
	Id       int //muss unique sein
	Name     string
	Storage  map[string]int //von was ist wieviel gelagert: CargoType? Hinzufügen nur mit Methode, rausnehmen direkt. MUSS mit make komplett initiiert werden
	Capacity int            //was ist die maximale Kapazität
	User     *User
}

// ist einzigartig durch station & name zusammen (ggf. id?)
type Plattform struct {
	Name    string
	Tiles   [][2]int // nur x & y, muss in Order sein. KEIN DIREKTER ZUGRIFF?
	Station *Station
}

type Produktionszyklus struct {
	Consumtion                 map[string]int `json:"Verbrauch"`
	Produktion                 map[string]int `json:"Produktion"`
	VerfuegbareLevelUndScaling []int          `json:"verfügbareLevelUndScaling"`
}

// Struktur für aktive Tiles
type ActiveTileCategory struct {
	Productioncycles []Produktionszyklus `json:"Produktionszyklen"`
}

// Root-Struktur für config.json
type ConfigData struct {
	TrainCategories      map[string][]string           `json:"Train Categories"`
	ActiveTileCategories map[string]ActiveTileCategory `json:"Aktive Tiles"`
}

// returns die erste und letzte Sub-Tile Koordinate
func (p *Plattform) GetFirstLast() [2][3]int {
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
func (p *Plattform) GetPathToOpposite(initial [3]int) [][3]int {
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

// remove = true -> referenz wird entfernt, = false -> wird hinzugefügt;
// findet alle Aktiven Tiles im validen Radius um das neue Tile der Station und verknüpft/entfernt entsprechend die Station, falls noch nicht geschehen
// egal ob neue Station oder nicht
func (s Station) ChangeStationTile(remove bool, position [2]int, gs *GameState) {
	//minimum bestimmen, maximum wird in schleife geprüft, dass nicht out of bounds
	xMin := position[0] - gs.StationRange
	if xMin < 0 {
		xMin = 0
	}
	yMin := position[1] - gs.StationRange
	if yMin < 0 {
		yMin = 0
	}

	// Ich war mal so frech - Jannis
	if remove {
		gs.Tiles[position[0]][position[1]].IsPlattform = false
	} else {
		gs.Tiles[position[0]][position[1]].IsPlattform = true

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
					if tile.ActiveTile.Stations[i].Id == s.Id {
						stationFound = true
						if remove {
							tile.ActiveTile.Stations = utils.RemoveElementFromSlice(tile.ActiveTile.Stations, i)
						}
						break
					}
				}
				// wenn die Station nicht gefunden wurde und hinzugefügt werden soll, hinzufügen der Station
				if !stationFound && !remove {
					gs.Tiles[x][y].ActiveTile.Stations = append(gs.Tiles[x][y].ActiveTile.Stations, &s)
					// continue
				}
			}
		}
	}
}

// return Restwert, der keinen Platz gefunden hat
func (s *Station) AddCargo(cargoType string, quantity int) int {
	filled := s.GetFillLevel()
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
func (s Station) GetFillLevel() int {
	var filled int //aktueller gefüllter Wert
	for _, i := range s.Storage {
		filled += i
	}
	return filled
}

//-------------------------------------------------------------------------------------
