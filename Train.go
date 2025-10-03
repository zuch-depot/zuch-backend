package main

import (
	"strconv"
)

type Train struct {
	Name               string
	Waggons            []*TrainType //Alle müssen nebeneinander spawnen
	Schedule           Schedule
	NextStop           Stop //nur fürs testen
	NextGoal           [3]int
	LastGoal           [3]int
	CurrentStop        Stop     //wird überschrieben, wenn der Zug von der Zielstation ausfahren möchte und nur mit StaionsStops
	CurrentPath        [][3]int //neu berechnen bei laden
	CurrentPathSignals [][3]int
	FoundPathToNext    bool //ob das letzte Pathfinding nicht erfolgreich war, dann muss erneut veruscht werden, ohne neuen Stop auszuwählen

	Waiting         bool //hat letzten Tick ein geblockes Tile gefunden oder keinen Weg gefunden und wartet
	LoadingTime     int  //Wie lange ist der Zug schon am be-/entladen? 0 == nicht am laden. Zeiteinheit ist wie oft methode aufgerufen wurde
	FinishedLoading bool //wenn nichts mehr geladen wird true. Kann auch wieder zurückgenommen werden
	Id              int
	User            *User
}

type TrainType struct {
	Position     [3]int //x,y,sub
	MaxSpeed     int
	Id           int
	Size         int
	Cargo        int //was ist das?
	CargoStorage *CargoStorage
}

type CargoStorage struct {
	capacity        int //statisch
	filled          int
	filledCargoType string
	CargoCategory   string //statisch
}

// returned Tile zum entblcken
// wenn fertig mit Laden/entladen passiert ein Tick nichts und dann fährt er los
func (t *Train) calculateTrain() [2]int {

	//bestimmt, ob man beim letzten Pathfining erfolgreich war. wenn ja und man den nächsten sucht, dann wird der nächste Stop ausgewählt
	//sonst wird neu versucht
	pathfindToNextAndMove := func() [2]int {
		//wie in Variablennamen beschrieben
		if t.FoundPathToNext {
			t.NextStop = t.Schedule.nextStop(t.NextStop)
		}
		t.recalculatePath()
		//aktualisierung der Variable (siehe Variablenbeschreibung)
		if len(t.CurrentPath) == 0 {
			t.FoundPathToNext = false
		} else {
			t.FoundPathToNext = true
			return t.move(true)
		}
		return [2]int{-1, -1}
	}

	var r [2]int

	//ist gerade in die Staion eingefahren, also speichern der aktuellen Station. Nächste Station wird bei neuberechen überschrieben
	if t.NextStop.IsPlattform && t.LoadingTime == 0 && t.NextGoal == t.Waggons[0].Position {
		t.CurrentStop = t.NextStop
		t.LastGoal = t.NextGoal
		logger.Debug("Zug " + t.Name + " in " + t.CurrentStop.Plattform.station.Name + " eingefahren.")
	} else if !t.NextStop.IsPlattform && len(t.CurrentPath) == 0 {
		//wenn das nächste Ziel ein Wegpunkt ist und man angekommen ist, braucht man einfach den nächsten Stop aussuchen und fahren
		return pathfindToNextAndMove()
	}

	//wenn der aktuelle Stop eine Plattform ist und man an der an der Station steht
	if t.CurrentStop.IsPlattform && t.LastGoal == t.Waggons[0].Position {
		//wenn min Zeit erreicht ist überprüfen und man fertig mit laden ist, ob man fahren kann
		if t.LoadingTime >= minLoadUloadTicks && t.FinishedLoading {
			logger.Debug("Zug " + t.Name + " versucht aus " + t.CurrentStop.Plattform.station.Name + " auszufahren.")

			if len(t.CurrentPath) == 0 {
				r = pathfindToNextAndMove()
			} else {
				r = t.move(false)
			}

			//Ist Zug losgefahren, also Reset der Werte fürs nächste Laden
			if !t.Waiting {
				logger.Debug("Zug " + t.Name + " aus " + t.CurrentStop.getName() + "ausgefahren.")
				t.LoadingTime = 0
				t.FinishedLoading = false
				return r
			}
		}
		//laden/entladen, wenn er noch warten muss oder noch laden muss
		if t.Waiting || t.LoadingTime < minLoadUloadTicks || !t.FinishedLoading {
			t.FinishedLoading = t.loadUndload()

			printTrains()
		}
		t.LoadingTime++

		//wenn er sich nicht bewegt hat
		return [2]int{-1, -1}
	}
	if len(t.CurrentPath) == 0 {
		return pathfindToNextAndMove()
	}
	return t.move(false)
}

// returnt ob der Zug voll ist oder nichts mehr zu laden ist, also abfahrtsbereit ist
func (t *Train) loadUndload() bool {
	var r bool

	//station, in die der Zug steht
	sta := t.CurrentStop.Plattform.station

	//es wird durch die Reihenfolge der Commands zuerst geladen, dann entladen.
	// Dabei wird nur beladen, wenn entladen fertig ist, bzw. noch kapazität von Gütern bewegt pro Tick über gelassen hat
	avaliableLoadUnloadSpeed := loadUnloadSpeed //misst, wie viel noch geladen und entladen werden darf
	for _, command := range t.CurrentStop.LoadUnloadCommand {
		//wenn man nichts mehr verladen darf, dann kann man noch nicht fertig sein
		if avaliableLoadUnloadSpeed == 0 {
			return false
		}
		if command.Loading {
			//loading the train
			for _, cargo := range command.CargoType {
				var loaded int
				//Berücksichtigung, dass max LoadUnloadSpeed pro Vorgang verladen wird
				if sta.Storage[cargo] >= avaliableLoadUnloadSpeed {
					loaded = avaliableLoadUnloadSpeed - t.loadCargo(cargo, avaliableLoadUnloadSpeed) //hinzufügen in den Zug
				} else {
					loaded = sta.Storage[cargo] - t.loadCargo(cargo, sta.Storage[cargo])
				}
				sta.Storage[cargo] -= loaded //Entfernen aus der Station
				avaliableLoadUnloadSpeed -= loaded

				if loaded > 0 {
					logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(loaded) + " Tonnen " + string(cargo) + " geladen")
				}

				//wenn man nicht bis Voll wartet und nichts verladen wurde, ist man fertig
				if !command.WaitTillFull && loaded <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				//ist fertig, wenn warten auf voll sein und zug voll ist ------------------> EINGÜGEN!
				if loaded <= 0 && avaliableLoadUnloadSpeed != 0 {
					r = true
					continue
				}
				r = false
			}
		} else {
			//unloading the train
			for _, cargo := range command.CargoType {
				var removed int
				//ausladen was geht aus den Züge, max LoadUnloadSpeed
				if sta.Capacity-sta.getFillLevel() >= avaliableLoadUnloadSpeed {
					removed = t.unloadCargo(cargo, avaliableLoadUnloadSpeed)
				} else {
					removed = t.unloadCargo(cargo, sta.Capacity-sta.getFillLevel())
				}
				avaliableLoadUnloadSpeed -= removed

				sta.addCargo(cargo, removed)

				if removed > 0 {
					logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(removed) + " Tonnen " + string(cargo) + " entladen")
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

	return r
}

// return nicht geladenen Cargo. Geht davon aus, dass toLoad in Grenzen des LoadUnloadSpeedes ist
func (t *Train) loadCargo(cargoType string, toLoad int) int {
	var r int

	for _, waggon := range t.Waggons {
		//wenn nichts mehr zu laden ist, breche ab
		if toLoad == 0 {
			break
		}

		//wenn Waggon richtigen CargoType hat, wenn er schon gefüllt ist, wird gefüllt, oder wenn leer ist, die passende Category hat
		if waggon.CargoStorage != nil {

			if (waggon.CargoStorage.filled == 0 && waggon.CargoStorage.CargoCategory == getCargoCategory(cargoType)) ||
				(cargoType == waggon.CargoStorage.filledCargoType) {
				emptySpace := waggon.CargoStorage.capacity - waggon.CargoStorage.filled
				//wenn Waggon voll ist oder gefüllter wert, wenn was gefüllt ist, nächsten nehmen
				if emptySpace == 0 {
					continue
				}
				if emptySpace >= toLoad {
					waggon.CargoStorage.filled += toLoad //auffüllen mit Rest zum Laden
					toLoad = 0                           //alles ist verladen
				} else {
					waggon.CargoStorage.filled += emptySpace //auffüllen, bis voll
					toLoad -= emptySpace                     //aufgefüllte Menge aus der, die Aufzufüllen ist, entfernen
				}
				waggon.CargoStorage.filledCargoType = cargoType
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
		if waggon.CargoStorage != nil && waggon.CargoStorage.filledCargoType == cargoType {
			cargoInWaggon := waggon.CargoStorage.filled
			if cargoInWaggon > 0 {
				//wenn der noch zu entnehmende Platz größer oder gleich groß ist, als die Menge, die im Wagen ist, nehme einfach alles
				if maxCargoRemoved-cargoRemovedSoFar >= cargoInWaggon {
					cargoRemovedSoFar += waggon.CargoStorage.filled
					waggon.CargoStorage.filled = 0
				} else {
					//wenn nicht mehr alles rauszunehmen ist, nehme den Rest Platz aus Waggon raus und dann ist maxRemoved die Menge, die entfernt wurde
					waggon.CargoStorage.filled -= maxCargoRemoved - cargoRemovedSoFar
					return maxCargoRemoved
				}
			}
		}
	}

	return cargoRemovedSoFar
}

// für 2 Wege Signale muss geprüft werden, ob nicht schon ein Zug zum Signal auf der anderen Seite fährt
// returnt Tile zum unblocken
func (t *Train) move(wasRecalculated bool) [2]int {
	var entblocken [2]int
	newGenNoSignal := t.Waiting //neu generiert und kein Signal, oder er hat letzte mal gewartet, dann gucken, ob immer noch

	//Wenn das neu generiert wurde
	if wasRecalculated {
		newGenNoSignal = true
		//wenn das Pathfinding nicht funktioniert hat
		if len(t.CurrentPath) == 0 {
			t.Waiting = true

			logger.Debug("Zug " + t.Name + " konnte den Weg zu " + t.NextStop.getName() + " nicht finden.")
			return [2]int{-1, -1}
		}
	}

	path := t.CurrentPath
	signals := t.CurrentPathSignals

	if newGenNoSignal && len(signals) > 1 {
		newGenNoSignal = false
	}
	//ist man bei einem Signal oder wurde neu generiert?
	// wenn es kein nächstes Signal gibt, wird bis zum Ziel geguckt
	// -----------> Ähnliche logik muss irgendwo Signale auf rot/grün/?gelb schalten
	// (vielleicht der Zug, wenn er sich merkt, bei welchem Signal er war und das umschaltet, wenn er aus block rausgefahren ist.
	// wird dann überschrieben, wenn der nächste zug nicht weiterfahren kann)
	// logger.Debug("newGenNoSignal", newGenNoSignal, "Signale:", signals, "Weg:", path)
	if newGenNoSignal || len(signals) > 1 && t.Waggons[0].Position == signals[0] {
		//gucken, ob bis zum nächsten Signal alle Tiles unblocked sind, sonst fahre nicht weiter
		// (es wird immer auch das letzte Tile überprüft, da man über ein sub tile ohne signal fahren muss, um zu einem zu kommen)
		// --> wichtig für Stationen, immer letzte Subtile ansteuern
		for i := 0; (newGenNoSignal && path[i] != signals[0]) ||
			!newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			if tiles[path[i][0]][path[i][1]].IsBlocked {
				logger.Debug("Zug " + t.Name + ": Blocked Tile found: []" + strconv.Itoa(path[i][0]) + ", " + strconv.Itoa(path[i][1]) + ", " + strconv.Itoa(path[i][2]) + ". Waiting")
				t.Waiting = true
				return [2]int{-1, -1}
			}
		}
		//da nichts geblocked war, blockt dieser Zug jetzt die Strecke zum nächsten Signal
		for i := 0; newGenNoSignal && path[i] != signals[0] || !newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			tiles[path[i][0]][path[i][1]].IsBlocked = true
		}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.CurrentPathSignals = t.CurrentPathSignals[1:]

		t.Waiting = false
	}

	//entblocken des letzten Tiles, wenn letzter Waggon sich rausbewegt (x oder y vom letzten unterschiedlich ist zum 2. letzten)
	// in die Queue schreiben, da entblocken nur am Ende des Ticks
	if len(t.Waggons) == 1 ||
		(t.Waggons[len(t.Waggons)-1].Position[0] != t.Waggons[len(t.Waggons)-2].Position[0] ||
			t.Waggons[len(t.Waggons)-1].Position[1] != t.Waggons[len(t.Waggons)-2].Position[1]) {
		letzterWagON := t.Waggons[len(t.Waggons)-1]
		tiles[letzterWagON.Position[0]][letzterWagON.Position[1]].IsBlocked = false
	}

	//Bewegung der Waggons
	for i := len(t.Waggons); i > 1; i-- {
		t.Waggons[i-1].Position = t.Waggons[i-2].Position
	}

	//Bewegung der Lokomotive
	t.Waggons[0].Position = t.CurrentPath[0]
	//rausschmeißen des Tiles, wo die Lok sich hinbewegt hat aus der Queue
	t.CurrentPath = t.CurrentPath[1:]

	// for alle waggons {clients.schickeNachtricht(waggong x,y, hat sich bewegt)}

	// Alles was bis hier gekmmen ist hat sich bewegt (laut wilken beim döner essen)
	broadcastChannel <- wsEnvelope{Type: "train.move", Msg: trainMoveMSG{Id: t.Id, Waggons: t.Waggons}}

	return entblocken
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
