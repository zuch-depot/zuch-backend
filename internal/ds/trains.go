package ds

import (
	"fmt"
	"log/slog"
	"strconv"
)

// gesetzt werden müssen: Name, Waggons, Id
type Train struct {
	Name               string    // Nicht eindeutig, dafür siehe ID
	Waggons            []*Waggon //Alle müssen nebeneinander spawnen
	Schedule           Schedule
	NextStop           Stop     //der Stop, in dem der Zug gerade hinfährt oder plant hinzufahren
	LastStop           Stop     //der Stop, in dem der Zug gerade ist oder gerade war || Überflüssig??
	CurrentPath        [][3]int //neu berechnen bei laden
	CurrentPathSignals [][3]int
	//FoundPathToNext    bool //ob das letzte Pathfinding nicht erfolgreich war, dann muss erneut veruscht werden, ohne neuen Stop auszuwählen, NEIN??

	Waiting         bool //hat letzten Tick ein geblockes Tile gefunden oder keinen Weg gefunden und wartet
	LoadingTime     int  //Wie lange ist der Zug schon am be-/entladen? 0 == nicht am laden. Zeiteinheit ist wie oft methode aufgerufen wurde
	FinishedLoading bool //wenn nichts mehr geladen wird true. Kann auch wieder zurückgenommen werden
	Id              int
	//User            *User nur ggf., falls wird das implementieren
}

type Waggon struct {
	Position     [3]int //x,y,sub
	MaxSpeed     int
	CargoStorage *CargoStorage
}

type CargoStorage struct {
	Capacity        int //statisch
	Filled          int
	FilledCargoType string
	CargoCategory   string //statisch
}

// entfernt diesen zug, dazu wird er aus dem array genommen und sein currentPath wird auf nicht blockiert gesetzt, hoffe das passt so
// fehler sind bisher ungenutzt, irgendwas wirrd schon schiefgehen
func (t *Train) RemoveTrain(gs *GameState) error {

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

// testet selber, ob es einen Weg gibt und berechnet bei Bedarf neu
// für 2 Wege Signale muss geprüft werden, ob nicht schon ein Zug zum Signal auf der anderen Seite fährt
// returnt Tile zum unblocken
func (t *Train) Move(gs *GameState) [2]int {
	entblocken := [2]int{-1, -1}

	// wenn ein Zug an einem Ziel angekommen ist und er nicht am ein/ausladen ist oder fertig damit ist, dann neu berechnen
	if len(t.CurrentPath) == 0 && (t.LoadingTime == 0 || t.LoadingTime > 0 && t.FinishedLoading) {
		//nächstes Ziel auswählen und den WEg dorthin berechnen
		t.LastStop = t.NextStop
		t.NextStop = t.Schedule.nextStop(t.NextStop)
		t.RecalculatePath(gs)

		//war das erfolgreich?
		if len(t.CurrentPath) == 0 {
			//TODO Nachricht: konnte das Ziel t.NextStop.Name() nicht erreichen
			gs.Logger.Debug(fmt.Sprintln("Zug", t.Name, "konnte das Ziel", t.NextStop, "nicht finden."))
			foundValidStop := false
			for i := 0; i < len(t.Schedule.Stops)-1; i++ {
				t.NextStop = t.Schedule.nextStop(t.NextStop)
				//t.RecalculatePath(gs)
				if len(t.CurrentPath) > 0 {
					foundValidStop = true
					break
				}
				//TODO Nachricht: ?Konnte Ziel xy nicht erreichen?
			}
			if !foundValidStop {
				//TODO Nachricht: Konnte kein Ziel erreichen, Zug ist stuck
				gs.Logger.Debug(fmt.Sprintln("Zug", t.Name, "kann kein Ziel erreichen."))
				return [2]int{}
			}
		}
	}

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

				if !t.Waiting {
					//TODO
					gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
						fmt.Sprint(t.Waggons[0].Position) + " auf rot geschaltet werden soll")
				}

				t.Waiting = true
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
		gs.BroadcastChannel <- WsEnvelope{Type: "tiles.block", Username: "Server", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.CurrentPathSignals = t.CurrentPathSignals[1:]

		gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
			fmt.Sprint(t.Waggons[0].Position) + " auf grün geschaltet werden soll. Irgnoriere das beim laden bitte xD")

		t.Waiting = false
	}

	//entblocken des letzten Tiles, wenn letzter Waggon sich rausbewegt (x oder y vom letzten unterschiedlich ist zum 2. letzten)
	// in die Queue schreiben, da entblocken nur am Ende des Ticks
	letzterWaggon := t.Waggons[len(t.Waggons)-1]
	if len(t.Waggons) == 1 || (letzterWaggon.Position[0] != t.Waggons[len(t.Waggons)-2].Position[0] ||
		letzterWaggon.Position[1] != t.Waggons[len(t.Waggons)-2].Position[1]) {

		// tiles[letzterWaggon.Position[0]][letzterWaggon.Position[1]].IsBlocked = false
		entblocken = [2]int{letzterWaggon.Position[0], letzterWaggon.Position[1]}

		//wenn das subTile, aus dem man sich herausbewegt ein Signal hat, hat man ein Signal passiert
		//das passt, da alle Signal außen an den Tiles stehen
		if gs.Tiles[letzterWaggon.Position[0]][letzterWaggon.Position[1]].Signals[letzterWaggon.Position[2]-1] {
			gs.Logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
				fmt.Sprint(letzterWaggon.Position) + " auf blau geschaltet werden soll")
		}
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

	// Alles was bis hier gekommen ist hat sich bewegt (laut wilken beim döner essen) || xD
	gs.BroadcastChannel <- WsEnvelope{Type: "train.move", Msg: TrainMoveMSG{Id: t.Id, Waggons: t.Waggons}}

	return entblocken
}

// returned Tile zum entblcken
// wenn fertig mit Laden/entladen passiert ein Tick nichts und dann fährt er los
func (t *Train) CalculateTrain(gs *GameState) [2]int {
	var r [2]int

	//ist gerade in die Staion eingefahren, also speichern der aktuellen Station. Nächste Station wird bei neuberechen überschrieben
	if t.NextStop.IsPlattform && t.LoadingTime == 0 && t.NextStop.Goal == t.Waggons[0].Position {
		t.LastStop = t.NextStop
		gs.Logger.Debug("Zug " + t.Name + " in " + t.LastStop.Plattform.Station.Name + " eingefahren.")
	} else if !t.NextStop.IsPlattform && len(t.CurrentPath) == 0 {
		//wenn das nächste Ziel ein Wegpunkt ist und man angekommen ist, braucht man einfach den nächsten Stop aussuchen und fahren
		return t.Move(gs)
	}

	//wenn der aktuelle Stop eine Plattform ist und man an der an der Station steht
	if t.LastStop.IsPlattform && t.LastStop.Goal == t.Waggons[0].Position {
		//wenn min Zeit erreicht ist überprüfen und man fertig mit laden ist, ob man fahren kann
		if t.LoadingTime >= gs.MinLoadUloadTicks && t.FinishedLoading {
			gs.Logger.Debug("Zug " + t.Name + " versucht aus " + t.LastStop.Plattform.Station.Name + " auszufahren.")

			r = t.Move(gs)

			//Ist Zug losgefahren, also Reset der Werte fürs nächste Laden
			if !t.Waiting {
				gs.Logger.Debug("Zug " + t.Name + " aus " + t.LastStop.getName() + "ausgefahren.")
				t.LoadingTime = 0
				t.FinishedLoading = false
				return r
			}
		}
		//laden/entladen, wenn er noch warten muss oder noch laden muss
		if t.Waiting || t.LoadingTime < gs.MinLoadUloadTicks || !t.FinishedLoading {
			t.FinishedLoading = t.LoadUndload(gs)

			printTrains(gs)
		}
		t.LoadingTime++

		//wenn er sich nicht bewegt hat
		return [2]int{-1, -1}
	}

	return t.Move(gs)
}

// returnt ob der Zug voll ist oder nichts mehr zu laden ist, also abfahrtsbereit ist
func (t *Train) LoadUndload(gs *GameState) bool {
	var r bool

	//station, in die der Zug steht
	sta := t.LastStop.Plattform.Station

	//es wird durch die Reihenfolge der Commands zuerst geladen, dann entladen.
	// Dabei wird nur beladen, wenn entladen fertig ist, bzw. noch kapazität von Gütern bewegt pro Tick über gelassen hat
	avaliableLoadUnloadSpeed := gs.LoadUnloadSpeed //misst, wie viel noch geladen und entladen werden darf
	for _, command := range t.LastStop.LoadUnloadCommand {
		//wenn man nichts mehr verladen darf, dann kann man noch nicht fertig sein
		if avaliableLoadUnloadSpeed == 0 {
			return false
		}
		if command.Loading {
			//loading the train
			for _, cargo := range command.CargoTypes {
				var loaded int
				//Berücksichtigung, dass max LoadUnloadSpeed pro Vorgang verladen wird
				if sta.Storage[cargo] >= avaliableLoadUnloadSpeed {
					loaded = avaliableLoadUnloadSpeed - t.LoadCargo(cargo, avaliableLoadUnloadSpeed, gs) //hinzufügen in den Zug
				} else {
					loaded = sta.Storage[cargo] - t.LoadCargo(cargo, sta.Storage[cargo], gs)
				}
				sta.Storage[cargo] -= loaded //Entfernen aus der Station
				avaliableLoadUnloadSpeed -= loaded

				if loaded > 0 {
					gs.Logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(loaded) + " Tonnen " + string(cargo) + " geladen")
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
			for _, cargo := range command.CargoTypes {
				var removed int
				//ausladen was geht aus den Züge, max LoadUnloadSpeed
				if sta.Capacity-sta.GetFillLevel() >= avaliableLoadUnloadSpeed {
					removed = t.UnloadCargo(cargo, avaliableLoadUnloadSpeed)
				} else {
					removed = t.UnloadCargo(cargo, sta.Capacity-sta.GetFillLevel())
				}
				avaliableLoadUnloadSpeed -= removed

				sta.AddCargo(cargo, removed)

				if removed > 0 {
					gs.Logger.Debug("Zug: " + t.Name + " hat " + strconv.Itoa(removed) + " Tonnen " + string(cargo) + " entladen")
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
func (t *Train) LoadCargo(cargoType string, toLoad int, gs *GameState) int {
	var r int

	for _, waggon := range t.Waggons {
		//wenn nichts mehr zu laden ist, breche ab
		if toLoad == 0 {
			break
		}

		//wenn Waggon richtigen CargoType hat, wenn er schon gefüllt ist, wird gefüllt, oder wenn leer ist, die passende Category hat
		if waggon.CargoStorage != nil {

			if (waggon.CargoStorage.Filled == 0 && waggon.CargoStorage.CargoCategory == getCargoCategory(cargoType, gs)) ||
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
func (t *Train) UnloadCargo(cargoType string, maxCargoRemoved int) int {
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
// true => Erfolgreich; false => fehler
func (t *Train) AddWaggon(position [3]int, typ string, gs *GameState) error {
	var capacity, maxSpeed int
	// Typ gibt kurz als string an was für einen Waggon man will
	// hier werden die passenden Attribute rausgesucht
	switch typ {
	case "Lebensmittel":
		capacity = 30
		maxSpeed = 77
	default:
		gs.Logger.Error("Invalider Typ", slog.String("Typ", typ))
		return fmt.Errorf("invalider Typ")
	}

	// Hier wird noch überorüft ob da überhaupt ein freies gleis ist
	if !gs.Tiles[position[0]][position[1]].Tracks[position[2]-1] { // Wenn dort ein gleis ist
		return fmt.Errorf("kein Gleis vorhanden")
	}

	// Waggons zu zug hinzufügen und entsprechende tiles blockieren
	t.Waggons = append(t.Waggons, &Waggon{Position: position, MaxSpeed: maxSpeed, CargoStorage: &CargoStorage{Capacity: capacity, CargoCategory: typ}})
	var blockedTilesPositions [][2]int
	blockedTilesPositions = append(blockedTilesPositions, [2]int(position[:2]))
	gs.Tiles[position[0]][position[1]].IsBlocked = true
	gs.BroadcastChannel <- WsEnvelope{Type: "tiles.block", Username: "Server", Msg: BlockedTilesMSG{Tiles: blockedTilesPositions}}

	gs.Logger.Debug("Blockiertes", slog.Int("Pos 0", position[0]), slog.Int("Pos 1", position[1]), slog.Bool("Blocked", gs.Tiles[position[0]][position[1]].IsBlocked))
	return nil
}
