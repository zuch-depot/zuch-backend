package main

import (
	"strconv"
)

type Train struct {
	Name               string
	Waggons            []*TrainType //Alle müssen nebeneinander spawnen
	Schedule           Schedule
	NextStop           Stop //nur fürs testen
	nextGoal           [3]int
	lastGoal           [3]int
	CurrentStop        Stop     //wird überschrieben, wenn der Zug von der Zielstation ausfahren möchte und nur mit StaionsStops
	currentPath        [][3]int //neu berechnen bei laden
	currentPathSignals [][3]int
	foundPathToNext    bool //ob das letzte Pathfinding nicht erfolgreich war, dann muss erneut veruscht werden, ohne neuen Stop auszuwählen

	waiting         bool //hat letzten Tick ein geblockes Tile gefunden oder keinen Weg gefunden und wartet
	loadingTime     int  //Wie lange ist der Zug schon am be-/entladen? 0 == nicht am laden. Zeiteinheit ist wie oft methode aufgerufen wurde
	finishedLoading bool //wenn nichts mehr geladen wird true. Kann auch wieder zurückgenommen werden
}

type TrainType struct {
	position [3]int //x,y,sub
	maxSpeed int

	CargoStorage *CargoStorage
}

type CargoStorage struct {
	capacity        int //statisch
	filled          int
	filledCargoType CargoType
	CargoCategory   CargoCategory //statisch
}

// returned Tile zum entblcken
// wenn fertig mit Laden/entladen passiert ein Tick nichts und dann fährt er los
func (t *Train) calculateTrain() [2]int {

	//bestimmt, ob man beim letzten Pathfining erfolgreich war. wenn ja und man den nächsten sucht, dann wird der nächste Stop ausgewählt
	//sonst wird neu versucht
	pathfindToNextAndMove := func() [2]int {
		//wie in Variablennamen beschrieben
		if t.foundPathToNext {
			t.NextStop = t.Schedule.nextStop(t.NextStop)
		}
		t.recalculatePath()
		//aktualisierung der Variable (siehe Variablenbeschreibung)
		if len(t.currentPath) == 0 {
			t.foundPathToNext = false
		} else {
			t.foundPathToNext = true
			return t.move(true)
		}
		return [2]int{-1, -1}
	}

	var r [2]int

	//ist gerade in die Staion eingefahren, also speichern der aktuellen Station. Nächste Station wird bei neuberechen überschrieben
	if t.NextStop.IsPlattform && t.loadingTime == 0 && t.nextGoal == t.Waggons[0].position {
		t.CurrentStop = t.NextStop
		t.lastGoal = t.nextGoal
		logger.Debug("Zug " + t.Name + " in " + t.CurrentStop.Plattform.station.Name + " eingefahren.")
	} else if !t.NextStop.IsPlattform && len(t.currentPath) == 0 {
		//wenn das nächste Ziel ein Wegpunkt ist und man angekommen ist, braucht man einfach den nächsten Stop aussuchen und fahren
		return pathfindToNextAndMove()
	}

	//wenn der aktuelle Stop eine Plattform ist und man an der an der Station steht
	if t.CurrentStop.IsPlattform && t.lastGoal == t.Waggons[0].position {
		//wenn min Zeit erreicht ist überprüfen und man fertig mit laden ist, ob man fahren kann
		if t.loadingTime >= minLoadUloadTicks && t.finishedLoading {
			logger.Debug("Zug " + t.Name + " versucht aus " + t.CurrentStop.Plattform.station.Name + " auszufahren.")

			if len(t.currentPath) == 0 {
				r = pathfindToNextAndMove()
			} else {
				r = t.move(false)
			}

			//Ist Zug losgefahren, also Reset der Werte fürs nächste Laden
			if !t.waiting {
				logger.Debug("Zug " + t.Name + " aus " + t.CurrentStop.getName() + "ausgefahren.")
				t.loadingTime = 0
				t.finishedLoading = false
				return r
			}
		}
		//laden/entladen, wenn er noch warten muss oder noch laden muss
		if t.waiting || t.loadingTime < minLoadUloadTicks || !t.finishedLoading {
			t.finishedLoading = t.loadUndload()

			printTrains()
		}
		t.loadingTime++

		//wenn er sich nicht bewegt hat
		return [2]int{-1, -1}
	}
	if len(t.currentPath) == 0 {
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
				if sta.capacity-sta.getFillLevel() >= avaliableLoadUnloadSpeed {
					removed = t.unloadCargo(cargo, avaliableLoadUnloadSpeed)
				} else {
					removed = t.unloadCargo(cargo, sta.capacity-sta.getFillLevel())
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
func (t *Train) loadCargo(cargoType CargoType, toLoad int) int {
	var r int

	for _, waggon := range t.Waggons {
		//wenn nichts mehr zu laden ist, breche ab
		if toLoad == 0 {
			break
		}
		//wenn Waggon richtigen CargoType hat
		if waggon.CargoStorage != nil && waggon.CargoStorage.filledCargoType == cargoType {
			emptySpace := waggon.CargoStorage.capacity - waggon.CargoStorage.filled
			//wenn Waggon voll ist, nächsten nehmen
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
		}
	}
	return r + toLoad
}

// returnt die Anzahl, die entfernt wurde. maxCargoRemoved ist dabei der Platz, der frei ist
// geht davon aus, dass maxCargoRemoved den LoadUnloadSpeed berücksichtigt und prüft es nicht selber
// -1, wenn kein passender Typ Waggon da ist -------------------> noch nicht!!!!!
func (t *Train) unloadCargo(cargoType CargoType, maxCargoRemoved int) int {
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
	newGenNoSignal := t.waiting //neu generiert und kein Signal, oder er hat letzte mal gewartet, dann gucken, ob immer noch

	//Wenn das neu generiert wurde
	if wasRecalculated {
		newGenNoSignal = true
		//wenn das Pathfinding nicht funktioniert hat
		if len(t.currentPath) == 0 {
			t.waiting = true

			logger.Debug("Zug " + t.Name + " konnte den Weg zu " + t.NextStop.getName() + " nicht finden.")
			return [2]int{-1, -1}
		}
	}

	path := t.currentPath
	signals := t.currentPathSignals

	if newGenNoSignal && len(signals) > 1 {
		newGenNoSignal = false
	}
	//ist man bei einem Signal oder wurde neu generiert?
	// wenn es kein nächstes Signal gibt, wird bis zum Ziel geguckt
	// -----------> Ähnliche logik muss irgendwo Signale auf rot/grün/?gelb schalten
	// (vielleicht der Zug, wenn er sich merkt, bei welchem Signal er war und das umschaltet, wenn er aus block rausgefahren ist.
	// wird dann überschrieben, wenn der nächste zug nicht weiterfahren kann)
	// logger.Debug("newGenNoSignal", newGenNoSignal, "Signale:", signals, "Weg:", path)
	if newGenNoSignal || len(signals) > 1 && t.Waggons[0].position == signals[0] {
		//gucken, ob bis zum nächsten Signal alle Tiles unblocked sind, sonst fahre nicht weiter
		// (es wird immer auch das letzte Tile überprüft, da man über ein sub tile ohne signal fahren muss, um zu einem zu kommen)
		// --> wichtig für Stationen, immer letzte Subtile ansteuern
		for i := 0; (newGenNoSignal && path[i] != signals[0]) ||
			!newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			if tiles[path[i][0]][path[i][1]].IsBlocked {
				logger.Debug("Zug " + t.Name + ": Blocked Tile found: []" + strconv.Itoa(path[i][0]) + ", " + strconv.Itoa(path[i][1]) + ", " + strconv.Itoa(path[i][2]) + ". Waiting")
				t.waiting = true
				return [2]int{-1, -1}
			}
		}
		//da nichts geblocked war, blockt dieser Zug jetzt die Strecke zum nächsten Signal
		for i := 0; newGenNoSignal && path[i] != signals[0] || !newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			tiles[path[i][0]][path[i][1]].IsBlocked = true
		}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.currentPathSignals = t.currentPathSignals[1:]

		t.waiting = false
	}

	//entblocken des letzten Tiles, wenn letzter Waggon sich rausbewegt (x oder y vom letzten unterschiedlich ist zum 2. letzten)
	// in die Queue schreiben, da entblocken nur am Ende des Ticks
	if len(t.Waggons) == 1 ||
		(t.Waggons[len(t.Waggons)-1].position[0] != t.Waggons[len(t.Waggons)-2].position[0] ||
			t.Waggons[len(t.Waggons)-1].position[1] != t.Waggons[len(t.Waggons)-2].position[1]) {
		entblocken = [2]int{t.Waggons[len(t.Waggons)-1].position[0], t.Waggons[len(t.Waggons)-1].position[1]}
	}

	//Bewegung der Waggons
	for i := len(t.Waggons); i > 1; i-- {
		t.Waggons[i-1].position = t.Waggons[i-2].position
	}

	//Bewegung der Lokomotive
	t.Waggons[0].position = t.currentPath[0]
	//rausschmeißen des Tiles, wo die Lok sich hinbewegt hat aus der Queue
	t.currentPath = t.currentPath[1:]

	// for alle waggons {clients.schickeNachtricht(waggong x,y, hat sich bewegt)}

	return entblocken
}

type WaggonModdel int

const (
	Dampflokomotive WaggonModdel = iota
	SchüttgutWagen
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
