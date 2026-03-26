package ds

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
)

// gesetzt werden müssen: Name, Waggons, Id
type Train struct {
	Id                 int
	Name               string    // Nicht eindeutig, dafür siehe ID
	Waggons            []*Waggon //Alle müssen nebeneinander spawnen
	Schedule           *Schedule
	NextStop           *Stop    //der Stop, in dem der Zug gerade hinfährt oder plant hinzufahren. Wenn 0, dann nächster Stop aus Schedule
	CurrentPath        [][3]int //ggf. nur bis zum nächsten Signal notwendig -> dann andere Bedingung finden, wann man angekommen ist
	CurrentPathSignals [][3]int //dementsprechend ggf. unnötig -> auch unabhängig davon ggf. unnötig

	// Waiting         bool //hat letzten Tick ein geblockes Tile gefunden oder keinen Weg gefunden und wartet

	LoadingTime     int  //Wie lange ist der Zug schon am be-/entladen? 0 == nicht am laden, > 0 ist an nextStop.Plattform . Zeiteinheit ist wie oft methode aufgerufen wurde
	FinishedLoading bool //wenn nichts mehr geladen wird true. Wird nur beim losfahren zurückgesetzt

	paused bool //wurde er von Spieler blockiert?

	tickTillNextMove int // wie oft move aufgerufen werden muss, damit er sich wieder bewegen darf
}

type Waggon struct {
	Position     [3]int //x,y,sub
	MaxSpeed     int
	CargoStorage *CargoStorage

	Power       int //wie viele Tonnen kann er ziehen
	EmptyWeight int //wie viel wiegt der Waggon leer
}

type CargoStorage struct {
	Capacity        int //statisch
	Filled          int
	FilledCargoType string
	CargoCategory   string //statisch
}

// Todo, ebenfalls Gewicht von Ladung berücksichtigen
// return die Gesamtmasse des Zuges
func (t *Train) GetWeight() int {
	weight := 0

	for _, waggon := range t.Waggons {
		weight += waggon.EmptyWeight
		//HIER LADUNGSGEWICHT BERÜCKSICHTIGEN, gerade wird die m^3 als Gewicht genommen, muss noch mit dem gewicht pro m^3 multipliziert werden
		// weight += waggon.CargoStorage.Filled // * gewichtProM^3
	}

	return weight
}

// return die gesamte Zugkraft des Zuges
func (t *Train) GetPower() int {
	power := 0
	for _, waggon := range t.Waggons {
		power += waggon.Power
	}
	return power
}

// Bestimmung der kleinsten maximalen Geschwindigkeit in kilometer pro stunde
func (t *Train) GetMaxSpeed() int {

	r := t.Waggons[0].MaxSpeed
	for _, waggon := range t.Waggons[1:] {
		if r > waggon.MaxSpeed {
			r = waggon.MaxSpeed
		}
	}

	return r
}

// Überprüft, ob sich der letzte Wagen rausbewegt, wenn er sich einen weiter bewegt
// entblocken des letzten Tiles, wenn letzter Waggon sich rausbewegt (x oder y vom letzten unterschiedlich ist zum 2. letzten)
// oder wenn der Zug nur 1 lang ist, wenn sich die x oder y vom nächsten Tile des Weges unterscheiden
func (t *Train) checkUnblockOnMove(gs *GameState) [2]int {

	entblocken := [2]int{-1, -1}

	letzterWaggon := t.Waggons[len(t.Waggons)-1]
	if (len(t.Waggons) == 1 && (t.CurrentPath[0][0] != t.Waggons[0].Position[0] ||
		t.CurrentPath[0][1] != t.Waggons[0].Position[1])) ||
		(len(t.Waggons) > 1 && (letzterWaggon.Position[0] != t.Waggons[len(t.Waggons)-2].Position[0] ||
			letzterWaggon.Position[1] != t.Waggons[len(t.Waggons)-2].Position[1])) {

		entblocken = [2]int{letzterWaggon.Position[0], letzterWaggon.Position[1]}

		//wenn das subTile, aus dem man sich herausbewegt ein Signal hat, hat man ein Signal passiert
		//das passt, da alle Signal außen an den Tiles stehen
		if gs.Tiles[letzterWaggon.Position[0]][letzterWaggon.Position[1]].Signals[letzterWaggon.Position[2]-1] {
			gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
				fmt.Sprint(letzterWaggon.Position) + " auf blau geschaltet werden soll")
		}
	}

	return entblocken
}

// WARENLOGIK UND NEU BERECHNEN NEU machen, da auch doppelungen mit Recalculate Path bestehen -------------------------------------------------------------------------
// testet selber, ob es einen Weg gibt und berechnet bei Bedarf neu
// für 2 Wege Signale muss geprüft werden, ob nicht schon ein Zug zum Signal auf der anderen Seite fährt
// fährt nur, wenn nicht beim laden. Wenn fertig laden, geht selber aus Lademodus raus
// returnt Tile zum unblocken
func (t *Train) move(gs *GameState) [2]int {
	entblocken := [2]int{-1, -1}

	if t.LoadingTime != 0 {
		if !t.FinishedLoading {
			// noch nicht fertig mit entladen, also nicht losfahren
			return entblocken
		} else {
			// fertig mit entladen, also damit aufhören und wieder fahren
			t.LoadingTime = 0
			t.FinishedLoading = false
		}
	}

	// erstmal Warenladen ignorieren
	// wenn man am Ende des Weges angekommen ist oder bei einem Signal ist, neuberechnen
	pos := t.Waggons[0].Position
	if len(t.CurrentPath) == 0 || gs.Tiles[pos[0]][pos[1]].Signals[pos[2]-1] {
		//wenn am Ziel angekommen, nächstes Ziel auswählen, anosten bleibt das Ziel gleich
		if len(t.CurrentPath) == 0 {
			t.NextStop = t.Schedule.nextStop(t.NextStop)
			if t.NextStop.Id == 0 {
				return entblocken
			}
		}

		// den Weg berechnen
		t.recalculatePath(gs)

		//war das erfolgreich?
		if len(t.CurrentPath) == 0 {

			gs.Logger.Debug(fmt.Sprintln("Zug", t.Name, "konnte das Ziel", t.NextStop.getName(gs), "nicht finden."))
			foundValidStop := false
			for i := 0; i < len(t.Schedule.Stops)-1; i++ {
				t.NextStop = t.Schedule.nextStop(t.NextStop)
				t.recalculatePath(gs) //Hoffentlich kein Fehler
				if len(t.CurrentPath) > 0 {
					foundValidStop = true
					break
				}
				//TODO Nachricht: ?Konnte Ziel xy nicht erreichen?
			}
			if !foundValidStop {
				//TODO Nachricht: Konnte kein Ziel erreichen, Zug ist stuck
				gs.Logger.Debug(fmt.Sprintln("Zug", t.Name, "kann kein Ziel erreichen."))
				return [2]int{-1, -1}
			}
		}
	}

	// Weg ist da, jetzt nur fahren, wenn es der richtige Tick ist
	if t.tickTillNextMove != 0 {
		//ein tick weniger
		t.tickTillNextMove -= 1
		return entblocken
	}

	// bestimmung der geschwindigkeit und dementsprechend setzten der anzahl der Ticks dazwischen
	// ist hardcoded, sollte vielleicht anders gelöst werden
	percentMaxSpeed := 1.0
	percent := float64(t.GetPower()) / float64(t.GetWeight())
	if percent >= 1 {

	} else if percent < 1 && percent >= 0.9 {
		percentMaxSpeed = 0.8
	} else if percent < 0.9 && percent >= 0.7 {
		percentMaxSpeed = 0.5
	} else if percent < 0.7 && percent >= 0.5 {
		percentMaxSpeed = 0.1
	} else {
		percentMaxSpeed = 0
	}

	maxSpeed := float64(t.GetMaxSpeed()) * percentMaxSpeed // in kilometer pro stunde. Max. 720 km/h
	// Subiltes pro Minute, pro tick 50 Millisekungen, 0,05 sekunden also max. 1200 Subtiles pro Minute

	// km/h -> m/s -> subTiles/s -> subTiles/tick -> (1/..) ticks/subTile
	speed := 1.0 / (maxSpeed / 3.6 / 10 / 20) // in ticks/subtile
	t.tickTillNextMove = int(math.Round(speed))

	path := t.CurrentPath
	signals := t.CurrentPathSignals

	//testen, ob an einem Signal angekommen. Wenn das nicht das letzte war, versuche weiter zu fahren
	// -----------> Ähnliche logik muss irgendwo Signale auf rot/grün/?gelb schalten
	// (vielleicht der Zug, wenn er sich merkt, bei welchem Signal er war und das umschaltet, wenn er aus block rausgefahren ist.
	// wird dann überschrieben, wenn der nächste zug nicht weiterfahren kann)
	// logger.Debug("newGenNoSignal", newGenNoSignal, "Signale:", signals, "Weg:", path)
	gs.Logger.Debug(fmt.Sprintln("Train", t.Name, "Signale:", signals, "Weg:", path))
	if len(t.CurrentPathSignals) > 1 && t.Waggons[0].Position == signals[0] {

		//gucken, ob bis zum nächsten Signal alle Tiles unblocked sind, sonst fahre nicht weiter
		// (es wird immer auch das letzte Tile überprüft, da man über ein sub tile ohne signal fahren muss, um zu einem zu kommen)
		// --> wichtig für Stationen, immer letzte Subtile ansteuern
		for i := 0; path[i] != signals[1]; i++ {
			if i == 0 && t.Waggons[0].Position[0] == path[0][0] && t.Waggons[0].Position[1] == path[0][1] {
				continue
			}
			//man könnte hier auch eigentlich nur jeden 2. Pathtile testen, da immer 2 nebeneinander gleiches Tile sind Optimierung
			if gs.Tiles[path[i][0]][path[i][1]].IsBlocked {
				gs.Logger.Debug("Zug " + t.Name + ": Blocked Tile found: []" + strconv.Itoa(path[i][0]) + ", " + strconv.Itoa(path[i][1]) + ", " + strconv.Itoa(path[i][2]) + ". Waiting")

				// if !t.Waiting {
				// 	//TODO
				// 	gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
				// 		fmt.Sprint(t.Waggons[0].Position) + " auf rot geschaltet werden soll")
				// }

				// t.Waiting = true
				return [2]int{-1, -1}
			}
		}
		//da nichts geblocked war, blockt dieser Zug jetzt die Strecke zum nächsten Signal
		var blockedTilesPositions [][2]int
		for i := 0; path[i] != signals[1]; i++ {
			gs.Tiles[path[i][0]][path[i][1]].IsBlocked = true
			blockedTilesPositions = append(blockedTilesPositions, [2]int{path[i][0], path[i][1]})
		}

		// dann können die clients auch nett anzeigen welche jetzt blockiert sind :D , brauche ich vielleicht auch fürs debuggen :D :D
		gs.BroadcastChannel <- WsEnvelope{Type: "tiles.block", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.CurrentPathSignals = t.CurrentPathSignals[1:]

		gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
			fmt.Sprint(t.Waggons[0].Position) + " auf grün geschaltet werden soll. Irgnoriere das beim laden bitte xD")

		// t.Waiting = false
	}

	entblocken = t.checkUnblockOnMove(gs) // in die Queue schreiben, da entblocken nur am Ende des Ticks

	//Bewegung der Waggons
	for i := len(t.Waggons); i > 1; i-- {
		t.Waggons[i-1].Position = t.Waggons[i-2].Position
	}

	//Bewegung der Lokomotive
	t.Waggons[0].Position = t.CurrentPath[0]
	//rausschmeißen des Tiles, wo die Lok sich hinbewegt hat aus der Queue
	t.CurrentPath = t.CurrentPath[1:]

	// for alle waggons {clients.schickeNachtricht(waggong x,y, hat sich bewegt)}

	// Alles was bis hier gekommen ist hat sich bewegt (laut wilken beim döner essen) || xD
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}
	return entblocken
}

// returned Tile zum entblcken
// wenn fertig mit Laden/entladen passiert ein Tick nichts und dann fährt er los
func (t *Train) calculateTrain(gs *GameState) [2]int {
	//var r [2]int

	if t.paused {
		return [2]int{-1, -1}
	}

	if t.Schedule == nil {
		return [2]int{-1, -1}
	}

	//ist gerade in die Staion eingefahren. wird in loadTime gespeichert als. die station ist die, des aktuellen Stoppes. Muss sich nicht bewegen, wenn gerade in station eingelaufen
	// if t.NextStop.IsPlattform && t.LoadingTime == 0 && len(t.CurrentPath) == 0 {
	goals := t.NextStop.getGoals(gs)
	if t.NextStop.IsPlattform && slices.Contains(goals, t.Waggons[0].Position) && !t.FinishedLoading && len(t.CurrentPath) == 0 {
		gs.Logger.Debug("Zug " + t.Name + " in " + t.NextStop.Plattform.GetStation(gs).Name + " eingefahren.")

		t.LoadingTime++

		return [2]int{-1, -1}
	} else if !t.NextStop.IsPlattform && len(t.CurrentPath) == 0 {
		//wenn das aktuelle Ziel ein Wegpunkt ist und man angekommen ist, braucht man einfach den nächsten Stop aussuchen und fahren
		return t.move(gs)
	}

	return t.move(gs)
}

// läd und entläd den Zug, wenn er gerade in einem Bahnhof ist und der Zug nicht pausiert ist oder schon losfahren möchte
func (t *Train) loadUnload(gs *GameState) error {

	if t.paused || t.LoadingTime == 0 || t.FinishedLoading {
		// Ist nicht wirklich ein Fehler, wird halt einfach nicht beladen
		return nil
	}

	t.LoadingTime++

	var r bool

	//station, in die der Zug steht
	sta := t.NextStop.Plattform.GetStation(gs)

	//es wird durch die Reihenfolge der Commands zuerst geladen, dann entladen.
	// Dabei wird nur beladen, wenn entladen fertig ist, bzw. noch kapazität von Gütern bewegt pro Tick über gelassen hat
	avaliableLoadUnloadSpeed := gs.ConfigData.LoadUnloadSpeed //misst, wie viel noch geladen und entladen werden darf
	for _, command := range t.NextStop.LoadUnloadCommand {
		//wenn man nichts mehr verladen darf, dann kann man noch nicht fertig sein
		if avaliableLoadUnloadSpeed == 0 {
			r = false
			break
			// return false
		}
		if command.Loading {
			//loading the train
			for _, cargo := range command.CargoTypes {
				var loaded int
				//Berücksichtigung, dass max LoadUnloadSpeed pro Vorgang verladen wird
				if sta.Storage[cargo] >= avaliableLoadUnloadSpeed {
					loaded = avaliableLoadUnloadSpeed - t.loadCargo(cargo, avaliableLoadUnloadSpeed, gs) //hinzufügen in den Zug
				} else {
					loaded = sta.Storage[cargo] - t.loadCargo(cargo, sta.Storage[cargo], gs)
				}
				sta.Storage[cargo] -= loaded //Entfernen aus der Station
				avaliableLoadUnloadSpeed -= loaded

				if loaded > 0 {
					gs.Logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(loaded) + " Tonnen " + string(cargo) + " geladen")
				}

				//wenn man nicht bis Voll wartet und nicht max. aufgeladen wurde, ist man fertig
				if !command.WaitTillFull && loaded <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				//ist fertig, wenn warten auf voll sein und zug voll ist ------------------> TODO EINGÜGEN!
				if loaded <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				r = false
			}
		} else {
			//unloading the train
			for _, cargo := range command.CargoTypes {
				var removed int
				//ausladen was geht aus den Züge, max LoadUnloadSpeed
				if sta.Capacity-sta.GetFillLevel() >= avaliableLoadUnloadSpeed {
					removed = t.unloadCargo(cargo, avaliableLoadUnloadSpeed)
				} else {
					removed = t.unloadCargo(cargo, sta.Capacity-sta.GetFillLevel())
				}
				avaliableLoadUnloadSpeed -= removed

				sta.addCargo(cargo, removed)

				if removed > 0 {
					gs.Logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(removed) + " Tonnen " + string(cargo) + " entladen")
					// fmt.Println("Zug: " + t.Name + " hat " + strconv.Itoa(removed) + " Tonnen " + string(cargo) + " entladen")
				}
				//wenn nichts bewegt wurde und man nicht bis leer sein wartet, ist der Ladevorgang beendet
				//(damit ist immer einmal überprüfen, ohne, dass was passiert -> eig. nicht schlimm)
				if !command.WaitTillFull && removed <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				//wenn man bis leer sein wartet, muss der Zug leer sein ---------------> bestimmung, ob der Leer ist nach CargoCategory EINFÜGEN!
				if removed <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				r = false
			}
		}
	}

	// if avaliableLoadUnloadSpeed < gs.LoadUnloadSpeed && t.LoadingTime >= gs.MinLoadUloadTicks {
	if avaliableLoadUnloadSpeed > 0 && t.LoadingTime >= gs.ConfigData.MinLoadUloadTicks {
		r = true
	}

	// der user kriegt einfach den neuen zuch
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

	//
	t.FinishedLoading = r

	return nil
}

// return nicht geladenen Cargo. Geht davon aus, dass toLoad in Grenzen des LoadUnloadSpeedes ist
func (t *Train) loadCargo(cargoType string, toLoad int, gs *GameState) int {
	var r int

	for _, waggon := range t.Waggons {
		//wenn nichts mehr zu laden ist, breche ab
		if toLoad == 0 {
			break
		}

		//wenn Waggon richtigen CargoType hat, wenn er schon gefüllt ist, wird gefüllt, oder wenn leer ist, die passende Category hat
		if waggon.CargoStorage != nil {

			if (waggon.CargoStorage.Filled == 0 && waggon.CargoStorage.CargoCategory == gs.getCargoCategory(cargoType)) ||
				(cargoType == waggon.CargoStorage.FilledCargoType) {
				emptySpace := waggon.CargoStorage.Capacity - waggon.CargoStorage.Filled
				//wenn Waggon voll ist oder gefüllter wert, wenn was gefüllt ist, nächsten nehmen
				if emptySpace == 0 {
					continue
				}
				if emptySpace >= toLoad {
					waggon.CargoStorage.Filled += toLoad //auffüllen mit Rest zum Laden
					toLoad = 0                           //alles ist verladen
				} else {
					waggon.CargoStorage.Filled += emptySpace //auffüllen, bis voll
					toLoad -= emptySpace                     //aufgefüllte Menge aus der, die Aufzufüllen ist, entfernen
				}
				waggon.CargoStorage.FilledCargoType = cargoType
			}
		}
	}
	return r + toLoad
}

// returnt die Anzahl, die entfernt wurde. maxCargoRemoved ist dabei der Platz, der frei ist
// geht davon aus, dass maxCargoRemoved den LoadUnloadSpeed berücksichtigt und prüft es nicht selber
// -1, wenn kein passender Typ Waggon da ist -------------------> noch nicht!!!!!
func (t *Train) unloadCargo(cargoType string, maxCargoRemoved int) int {
	cargoRemovedSoFar := 0

	for _, waggon := range t.Waggons {
		if cargoRemovedSoFar == maxCargoRemoved {
			return cargoRemovedSoFar
		}
		//wenn richtiger CargoType gefunden wurde
		if waggon.CargoStorage != nil && waggon.CargoStorage.FilledCargoType == cargoType {
			cargoInWaggon := waggon.CargoStorage.Filled
			if cargoInWaggon > 0 {
				//wenn der noch zu entnehmende Platz größer oder gleich groß ist, als die Menge, die im Wagen ist, nehme einfach alles
				if maxCargoRemoved-cargoRemovedSoFar >= cargoInWaggon {
					cargoRemovedSoFar += waggon.CargoStorage.Filled
					waggon.CargoStorage.Filled = 0
				} else {
					//wenn nicht mehr alles rauszunehmen ist, nehme den Rest Platz aus Waggon raus und dann ist maxRemoved die Menge, die entfernt wurde
					waggon.CargoStorage.Filled -= maxCargoRemoved - cargoRemovedSoFar
					return maxCargoRemoved
				}
			}
		}
	}

	return cargoRemovedSoFar
}

// Fügt einen Wagon zu einem Zug, typ gibt die art des wagongs an, daraus basierend wird capacity und maxSpeed bestimmt, bspw "Lebensmittel"
// TODO waggonType umstellen <----------------------------------------------
func (t *Train) AddWaggon(position [3]int, typ string, gs *GameState) error {
	var capacity, maxSpeed, emptyWeight, power int

	// Typ gibt kurz als string an was für einen Waggon man will
	// hier werden die passenden Attribute rausgesucht
	switch typ {
	case "Lebensmittel":
		capacity = 30
		maxSpeed = 77
		emptyWeight = 10
	case "":
		capacity = 30
		maxSpeed = 180
		power = 100
		emptyWeight = 20
	default:
		gs.Logger.Error("Invalider Typ", slog.String("Typ", typ))
		return fmt.Errorf("invalider Typ")
	}

	//kontrollieren, dass das SubTile valide ist
	err := gs.iterateSubTiles(position, position, "An error accured while adding a waggon.", func(gs *GameState, coordinate [3]int) error { return nil })
	if err != nil {
		return err
	}

	// prüft, ob der waggon anhängt, wenn es schon welche gibt
	if len(t.Waggons) > 0 {
		err = t.isWaggonValid(position, gs)
		if err != nil {
			return err
		}
	} else {
		//wenn das die Lokomotive ist, muss man manuell überprüfen, ob da ein Track ist und ob das blockiert ist, sonst in isWaggonValid
		if !gs.Tiles[position[0]][position[1]].Tracks[position[2]-1] {
			return fmt.Errorf("Could not build a waggon there, there is no track on that position.")
		}
		if gs.Tiles[position[0]][position[1]].IsBlocked {
			return fmt.Errorf("Could not build a waggon there, the tile is blocked.")
		}
	}

	// Waggons zu zug hinzufügen und entsprechende tiles blockieren
	waggon := &Waggon{Position: position, MaxSpeed: maxSpeed, EmptyWeight: emptyWeight, Power: power, CargoStorage: &CargoStorage{Capacity: capacity, CargoCategory: typ}}
	if len(t.Waggons) == 0 {
		t.Waggons = []*Waggon{waggon}
	} else {
		t.Waggons = append(t.Waggons, waggon)
	}
	var blockedTilesPositions [][2]int
	blockedTilesPositions = append(blockedTilesPositions, [2]int(position[:2]))
	gs.Tiles[position[0]][position[1]].IsBlocked = true
	gs.BroadcastChannel <- WsEnvelope{Type: "tiles.block", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

	gs.Logger.Debug("Blockiertes", slog.Int("Pos 0", position[0]), slog.Int("Pos 1", position[1]), slog.Bool("Blocked", gs.Tiles[position[0]][position[1]].IsBlocked))
	return nil
}

// Überprüft ob die position valide wäre für einen waggon
func (t *Train) isWaggonValid(position [3]int, gs *GameState) error {

	prevWaggon := t.Waggons[len(t.Waggons)-1]

	//wenn das im selben tile wie der vorgänger ist, dann wir das Tile von diesem blockiert, daher der Bau trotzdem erlaubt
	if gs.Tiles[position[0]][position[1]].IsBlocked && (prevWaggon.Position[0] != position[0] || prevWaggon.Position[1] != position[1]) {
		return fmt.Errorf("track is blocked")
	}

	// man kann nicht auf dem gleichen SubTile wie der vorher letzte Waggon bauen
	if t.Waggons[len(t.Waggons)-1].Position == position {
		return fmt.Errorf("There is already a waggon there. Please provide a valid coordinate at the end of the train.")
	}

	// wenn es mehr als 1 waggon gibt, gucken, dass er nicht zwischendurch bauen möchte
	if len(t.Waggons) > 1 {
		secondLastPos := t.Waggons[len(t.Waggons)-2].Position
		if secondLastPos[0] == position[0] && secondLastPos[1] == position[1] {
			return fmt.Errorf("Please only try to add waggons to the end of the train.")
		}
	}

	possibleTracks := gs.neighbourTracks(prevWaggon.Position[0], prevWaggon.Position[1], prevWaggon.Position[2])

	if !slices.Contains(possibleTracks, position) {
		return fmt.Errorf("waggons are not continuous or a track is missing")
	}

	// Gibt einen Fehler zurück falls
	return nil
}

// Fügt Waggons im Bereich zu
func (t *Train) AddWaggons(startSubTile [3]int, endSubTile [3]int, waggonType string, gs *GameState) error {

	// zum verwenden in Methode
	gs.currentWaggonType = waggonType
	gs.currentTrain = t

	//prüfen der Parameter und hinzufügen der Waggons
	return gs.iterateSubTiles(startSubTile, endSubTile, "An error accured while adding waggons.", func(gs *GameState, coordinate [3]int) error {

		//hinzufügen des Waggons
		return gs.currentTrain.AddWaggon(coordinate, gs.currentWaggonType, gs)

	})
}

// kontrolliert nur, dass der index in range ist. Entblockt auch. Löscht den Zug, wenn dieser Leer ist
func (t *Train) RemoveWaggon(index int, gs *GameState) error {

	//validieren des Indexes
	if index > len(t.Waggons)-1 || index < 0 {
		return fmt.Errorf("%s", "Waggon out of bounds. "+strconv.Itoa(index)+" not a valid index for "+strconv.Itoa(len(t.Waggons))+" Waggons from train "+t.Name)
	}

	//ggf. enblocken des Tiles, da die Waggons aufrücken
	entblocken := t.checkUnblockOnMove(gs)

	//entblocken durchführen. Muss nicht in Queue, da nicht ausgeführt wird, während Bewegungen stattfinden
	gs.Tiles[entblocken[0]][entblocken[1]].IsBlocked = false
	gs.BroadcastChannel <- WsEnvelope{Type: "tiles.unblock", Msg: BlockedTilesMSG{Tiles: []([2]int){entblocken}}}

	//entfernt den Zug, wenn des keine Waggons mehr gibt
	if len(t.Waggons) == 0 {
		gs.RemoveTrain(t)
	}

	//entfernen des Waggons
	t.Waggons = append(t.Waggons[:index], t.Waggons[index:]...)
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

	return nil
}

// validiert die Indizes, entblockt, löscht den Zug, wenn dieser leer ist. Löscht alle innerhalb der indizes, die valide sind
func (t *Train) RemoveWaggons(indexStart int, indexEnd int, gs *GameState) error {

	//validieren der Indizes
	if (indexStart > len(t.Waggons)-1 || indexStart < 0) && (indexEnd > len(t.Waggons)-1 || indexEnd < 0) && indexStart <= indexEnd {
		return fmt.Errorf("%s", "Waggon out of bounds. "+strconv.Itoa(indexStart)+" and/or "+strconv.Itoa(indexEnd)+" not valid indizes for "+strconv.Itoa(len(t.Waggons))+" Waggons from train "+t.Name)
	}

	//entfernen der Waggons
	for i := indexStart; i < indexEnd; i++ {
		err := t.RemoveWaggon(i, gs)
		if err != nil {
			return err
		}
	}

	return nil
}

// weist dem Zug einen Fahrplan zu, noch kein Fehler wird geworfen
func (t *Train) AssignSchedule(schedule *Schedule, gs *GameState) {
	t.Schedule = schedule
	t.NextStop = schedule.nextStop(&Stop{Id: 0})
	t.CurrentPath = [][3]int{}
	t.CurrentPathSignals = [][3]int{}
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

}

// entfernt den Fahrplan von dem Zug, noch kein Fehler wird geworfen
func (t *Train) UnassignSchedule(gs *GameState) {
	t.Schedule = nil
	t.NextStop = nil
	t.CurrentPath = [][3]int{}
	t.CurrentPathSignals = [][3]int{}
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

}

// keine Überprüfung
func (t *Train) Pause(gs *GameState) {
	t.paused = true
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

}

// keine Überprüfung
func (t *Train) UnPause(gs *GameState) {
	t.paused = false
	gs.BroadcastChannel <- WsEnvelope{Type: "train.update", Msg: t}

}

// auch sonderzeigen erlaubt, ggf. anpassen
func (t *Train) Rename(name string) error {
	if name == "" {
		return fmt.Errorf("Please provide a name.")
	}
	t.Name = name
	return nil
}
