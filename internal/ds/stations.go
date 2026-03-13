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
	Plattforms map[int]*Plattform
}

// alle Tiles müssen aneinander in der gleichen Richtung sein und dürfen keine Kurven haben
type Plattform struct {
	Id    int //neu, muss ggf. implementiert werden
	Name  string
	Tiles [][2]int // nur x & y, muss in Order sein. KEIN DIREKTER ZUGRIFF?
	//Station *Station
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

// muss nur noch für demo trains öffentlich sein
// returnt die Station einer Plattform. Ist notwendig, da die Kommunikation nicht mit Zirkelverweisen klar kommt
func (p *Plattform) GetStation(gs *GameState) *Station {

	for _, station := range gs.Stations {
		for _, plattform := range station.Plattforms {
			if plattform.Id == p.Id {
				return station
			}
		}
	}

	return nil
}

// returns die erste und letzte Sub-Tile Koordinate, sollte auch Länge 1 supporten
func (p *Plattform) getFirstLast(gs *GameState) [2][3]int {
	var r [2][3]int
	r[0] = [3]int{p.Tiles[0][0], p.Tiles[0][1]}
	r[1] = [3]int{p.Tiles[len(p.Tiles)-1][0], p.Tiles[len(p.Tiles)-1][1]}

	// wenn die Tracks horizontal sind, dann ist die Plattform auch horizontal
	if gs.Tiles[p.Tiles[0][0]][p.Tiles[0][1]].Tracks[0] && gs.Tiles[p.Tiles[len(p.Tiles)-1][0]][p.Tiles[len(p.Tiles)-1][1]].Tracks[2] {
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

// gerade nur für CargoPathfinding
// effizientere Methode zu überprüfen, ob die Plattform zu der Station gehört
func (p *Plattform) isPlattfromFromStation(station *Station) bool {
	for _, plattform := range station.Plattforms {
		if plattform.Id == p.Id {
			return true
		}
	}
	return false
}

// wenn nicht außen, dann splitting (noch nicht impl). Validität wird nicht geprüft
func (p *Plattform) removeTile(position [2]int, gs *GameState) error {

	ends := p.getFirstLast(gs)
	if ends[0][0] == position[0] && ends[0][1] == position[1] {
		utils.RemoveElementFromSlice(p.Tiles, 0)
	} else if ends[1][1] == position[1] && ends[1][0] == position[0] {
		utils.RemoveElementFromSlice(p.Tiles, len(p.Tiles)-1)
	} else {
		//splitting TODO, erstmal Fehler
		return nil
	}

	gs.Tiles[position[0]][position[1]].IsPlattform = false

	if len(p.Tiles) == 0 {
		//Plattform löschen
		p.GetStation(gs).deletePlattform(p.Id, gs)

	}

	return nil
}

// prüft utils.Checkname und bennent danach um
func (p *Plattform) Rename(name string) error {

	err := utils.CheckName(name)
	if err != nil {
		return err
	}
	p.Name = name

	return nil
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

// return Restwert, der keinen Platz gefunden hat
// beachtet max. Lademenge pro Tick nicht
func (s *Station) addCargo(cargoType string, quantity int) int {
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
func (s *Station) GetFillLevel() int {
	var filled int //aktueller gefüllter Wert
	for _, i := range s.Storage {
		filled += i
	}
	return filled
}

// name kann auch "", Id ist Standardname
func (s *Station) addPlattform(name string, tiles [][2]int, gs *GameState) error {

	if name == "" {
		name = fmt.Sprint("Plattform ", gs.CurrentPlattformID.Load())
	}

	s.Plattforms[int(gs.CurrentPlattformID.Load())] = &Plattform{Id: int(gs.CurrentPlattformID.Load()), Name: name, Tiles: tiles /*, Station: s*/}
	gs.CurrentPlattformID.Add(1)

	return nil
}

// entfernt nur die Plattform auch aus allen stops und wenn die Station leer ist, diese auch. Manipuliert keine Tiles
// Auch zu benutzen, wenn schon alle Tiles weg sind
func (s *Station) deletePlattform(Id int, gs *GameState) error {

	var err error

	plattform, ok := s.Plattforms[Id]
	if !ok {
		return fmt.Errorf("could not find plattform")
	}

	/*
		//Entfernt alle Tiles mit der Plattform
		//(wichtig, damit die ActiveTiles der Station aktualisiert werden)
		if len(plattform.Tiles) == 0 {
			for _, pos := range plattform.Tiles {
				_, err = ChangeStationTile(true, pos, gs)
				if err != nil {
					return err
				}
			}
			return nil
		}
	*/

	//Entfernung der Plattform aus allen Stops
	for _, schedule := range gs.Schedules {
		for _, stop := range schedule.Stops {
			if stop.IsPlattform && stop.Plattform.Id == plattform.Id {
				err = schedule.RemoveStop(stop.Id, gs)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	//Entfernung der Plattform aus der Station
	delete(s.Plattforms, Id)
	if err != nil {
		return err
	}

	//Kontrolle, ob die Station noch Plattformen hat
	if len(s.Plattforms) == 0 {
		//TODO error
		//Station löschen
		gs.deleteStation(s)
	}

	return nil
}

// entgfernt alle Tiles der Plattform
func (s *Station) RemovePlattform(Id int, gs *GameState) error {

	for _, tilePos := range s.Plattforms[Id].Tiles {
		_, err := gs.RemoveStationTile(tilePos)
		if err != nil {
			return err
		}
	}
	return nil
}

// checks the name with utils.checkname und setzt den namen
func (s *Station) Rename(name string) error {

	err := utils.CheckName(name)
	if err != nil {
		return nil
	}

	s.Name = name
	return nil
}
