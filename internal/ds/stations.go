package ds

import (
	"fmt"
	"slices"
	"zuch-backend/internal/utils"
)

type Station struct {
	Id         int //muss unique sein
	Name       string
	Storage    map[string]int //von was ist wieviel gelagert: CargoType? Hinzufügen nur mit Methode, rausnehmen direkt. MUSS mit make komplett initiiert werden
	Capacity   int            //was ist die maximale Kapazität
	Plattforms []*Plattform
}

// alle Tiles müssen aneinander in der gleichen Richtung sein und dürfen keine Kurven haben
type Plattform struct {
	Id      int //neu, muss ggf. implementiert werden
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

// Validität wird hier nicht geprüft, nur am richtigen Ende hinzugefügt. (eig. nur für ChangeStationTile)
// smaller: links oder oben
func (p *Plattform) addTile(position [2]int, smaller bool, gs *GameState) error {

	if smaller {
		p.Tiles = append([][2]int{position}, p.Tiles...)
	} else {
		p.Tiles = append(p.Tiles, position)
	}

	gs.Tiles[position[0]][position[1]].IsPlattform = true
	return nil
}

// wenn nicht außen, dann splitting (noch nicht impl). Validität wird nicht geprüft
func (p *Plattform) removeTile(position [2]int, gs *GameState) error {

	ends := p.GetFirstLast()
	if ends[0][0] == position[0] {
		utils.RemoveElementFromSlice(p.Tiles, 0)
	} else if ends[1][1] == position[1] {
		utils.RemoveElementFromSlice(p.Tiles, len(p.Tiles)-1)
	} else {
		//splitting TODO, erstmal Fehler
		return nil
	}

	gs.Tiles[position[0]][position[1]].IsPlattform = false

	return nil
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
// baut nur neue Staion, baut keine Schienen
// TODO dynamische Zuweisung zu Gleisen anhand Ausrichtung und nachbarn
func (s Station) ChangeStationTile(remove bool, position [2]int, gs *GameState) error {
	var err error
	tile := gs.Tiles[position[0]][position[1]]

	//überprüft, ob das Tile gerade ist
	var horizontal bool
	if tile.Tracks[0] && tile.Tracks[2] && !tile.Tracks[1] && !tile.Tracks[3] {
		horizontal = true
	} else if tile.Tracks[1] && tile.Tracks[3] && !tile.Tracks[0] && !tile.Tracks[2] {
		horizontal = false
	} else {
		//TODO error
		return nil
	}

	// Ich war mal so frech - Jannis
	// Ich habe das mal für dich ausformuliert - Wilken
	if remove {
		if !tile.IsPlattform {
			//TODO Error, kann nicht von nichts entfernen
			return nil
		}

		//Bestimmung der Plattform
		var plattform Plattform
		plattform, err = getPlattform(position, gs)
		if err != nil {
			return err
		}

		//Entfernung des Tags und weitere Berechnungen
		err = plattform.removeTile(position, gs)
		if err != nil {
			return err
		}

	} else {

		//berührt das Tile eine Station? Wenn nein, dann Id = 0
		var stationBordering *Station
		var above, under, left, right *Tile
		if position[0] > 0 {
			left = gs.Tiles[position[0]-1][position[1]]
			if left.IsPlattform {
				var temp Plattform
				temp, err = getPlattform([2]int{position[0] - 1, position[1]}, gs)
				if err != nil {
					return err
				}
				stationBordering = temp.Station
			}
		}
		if position[1] > 0 {
			above = gs.Tiles[position[0]][position[1]-1]
			if above.IsPlattform {
				var temp Plattform
				temp, err = getPlattform([2]int{position[0], position[1] - 1}, gs)
				if err != nil {
					return err
				}
				stationBordering = temp.Station
			}
		}
		if position[0] < gs.SizeX-1 {
			right = gs.Tiles[position[0]+1][position[1]]
			if right.IsPlattform {
				var temp Plattform
				temp, err = getPlattform([2]int{position[0] + 1, position[1]}, gs)
				if err != nil {
					return err
				}
				stationBordering = temp.Station
			}
		}
		if position[0] < gs.SizeY-1 {
			under = gs.Tiles[position[0]][position[1]+1]
			if under.IsPlattform {
				var temp Plattform
				temp, err = getPlattform([2]int{position[0], position[1] + 1}, gs)
				if err != nil {
					return err
				}
				stationBordering = temp.Station
			}
		}

		if stationBordering.Id == 0 {
			//TODO neue Station initalisieren mit Standartwerten -------------------------------------------------------
			return nil
		}
		//es gibt eine Station. Nun wird das Gleis bestimmt
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
			var plattform Plattform //die Plattform, wenn eine angrenzt

			//nach rechts gucken
			if position[0] < gs.SizeX-1 {

				//ist rechter eine Plattfrom und richtig ausgerichtet?
				if right.Tracks[0] && right.IsPlattform {
					if first {
						//TODO Error, grenzt an 2 Gleise an
						return nil
					}
					last = true
				}
			}
			if first {
				//grenzt nur links an
				plattform, err = getPlattform([2]int{position[0] - 1, position[1]}, gs)
				if err != nil {
					return err
				}
				//hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)
			} else if last {
				//grenzt nur rechts an
				plattform, err = getPlattform([2]int{position[0] + 1, position[1]}, gs)
				if err != nil {
					return err
				}
				//hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				// TODO neues Gleis
				err = stationBordering.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil
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
			var plattform Plattform //die Plattform, wenn eine angrenzt

			//nach unten gucken
			if position[1] < gs.SizeY-1 {

				//ist unten eine Plattfrom und richtig ausgerichtet?
				if under.Tracks[1] && under.IsPlattform {
					if first {
						//TODO Error, grenzt an 2 Gleise an
						return nil
					}
					last = true
				}
			}
			if first {
				//grenzt nur links an
				plattform, err = getPlattform([2]int{position[0], position[1] - 1}, gs)
				if err != nil {
					return err
				}
				//hinzufügen zur Plattform am Anfang
				plattform.addTile(position, true, gs)
			} else if last {
				//grenzt nur rechts an
				plattform, err = getPlattform([2]int{position[0], position[1] + 1}, gs)
				if err != nil {
					return err
				}
				//hinzufügen zur Plattform am Ende
				plattform.addTile(position, false, gs)
			} else {
				// TODO neues Gleis
				err = stationBordering.addPlattform("", [][2]int{position}, gs)
				if err != nil {
					return nil
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
					if tile.ActiveTile.Stations[i].Id == s.Id {
						stationFound = true
						if remove {
							tile.ActiveTile.Stations, err = utils.RemoveElementFromSlice(tile.ActiveTile.Stations, i)
							if err != nil {
								return err
							}
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
	return nil
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

// name kann auch ""
func (s Station) addPlattform(name string, tiles [][2]int, gs *GameState) error {

	if name == "" {
		name = fmt.Sprint(gs.CurrentPlattformID.Load())
	}

	s.Plattforms = append(s.Plattforms, &Plattform{Id: int(gs.CurrentPlattformID.Load()), Name: name, Tiles: tiles, Station: &s})
	gs.CurrentPlattformID.Add(1)

	return nil
}

// gibt die Plattform zurück, zu der ein Tile gehört
func getPlattform(position [2]int, gs *GameState) (Plattform, error) {

	for _, station := range gs.Stations {
		for _, plattform := range station.Plattforms {
			for _, tile := range plattform.Tiles {
				if tile[0] == position[0] && tile[1] == position[1] {
					return *plattform, nil
				}
			}
		}
	}

	return Plattform{}, nil
}
