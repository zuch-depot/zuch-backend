package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"sync/atomic"
)

var currentTrainID atomic.Uint64

type Train struct {
	Name               string    // Nicht eindeutig, dafür siehe ID
	Waggons            []*Waggon //Alle müssen nebeneinander spawnen
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

type Waggon struct {
	Position     [3]int //x,y,sub
	MaxSpeed     int
	CargoStorage *CargoStorage
}

type CargoStorage struct {
	capacity        int //statisch
	filled          int
	filledCargoType string
	CargoCategory   string //statisch
}

// entfernt diesen zug, dazu wird er aus dem array genommen und sein currentPath wird auf nicht blockiert gesetzt, hoffe das passt so
// fehler sind bisher ungenutzt, irgendwas wirrd schon schiefgehen
func (t *Train) removeTrain(gs *gameState) error {

	var blockedTilesPositions [][2]int
	for _, v := range t.CurrentPath {
		gs.Tiles[v[0]][v[1]].IsBlocked = false // ich hab keine ahnung ob das so geht
		blockedTilesPositions = append(blockedTilesPositions, [2]int{v[0], v[1]})
	}
	// Das kann hier gut sein das da zeugs doppelt drinne ist aber das ist mir spontan egal, doppelt auf false setzen hält ohnehin besser
	gs.broadcastChannel <- wsEnvelope{Type: "tiles.unblock", Username: "Server", Msg: blockedTilesMSG{Tiles: blockedTilesPositions}}

	before := len(gs.Trains)
	delete(gs.Trains, t.Id)
	if !(before > len(gs.Trains)) {
		return fmt.Errorf("couldn't find train in map")
	}
	return nil
}

// Fügt einen Zug hinzu, anhand eines namens und der position sowie des typen und positionen der waggons
func addTrain(update trainCreateMSG, gs *gameState) (*Train, error) {
	// Weg muss ja frei sein, und alles müssen zusammenhängen
	err := checkIfWaggonsAreValid(update.Waggons, gs)
	if err != nil {
		return nil, err
	}

	train := &Train{Name: update.Name, Id: int(currentTrainID.Load())}
	currentTrainID.Add(1)

	for _, waggon := range update.Waggons {
		err := train.addWaggon(waggon.Position, waggon.Typ, gs.Tiles, gs)
		if err != nil {
			return nil, fmt.Errorf("this shoudln't happen; %s", err.Error())
		}
	}

	gs.Trains[train.Id] = train
	return train, nil
}

// Überprüft ob alle waggons eine Valide Position haben, also das Gleis nicht blockiert ist, das gleis existiert und die Waggons zusammenhänend sind
func checkIfWaggonsAreValid(waggons []trainCreateWaggons, gs *gameState) error {
	for i, waggon := range waggons {
		if gs.Tiles[waggon.Position[0]][waggon.Position[1]].IsBlocked {
			return fmt.Errorf("track is blocked")
		}
		// Schaut ob der n. Waggon ein Nachbar des n-1. Waggon ist, daher beim 0. Überspringen
		if i != 0 {
			prevWaggon := waggons[i-1]
			possibleTracks := neighbourTracks(prevWaggon.Position[0], prevWaggon.Position[1], prevWaggon.Position[2], gs)

			if !slices.Contains(possibleTracks, waggon.Position) {
				return fmt.Errorf("waggons are not continuos or a track is missing")
			}
		}
	}
	// Gibt einen Fehler zurück falls
	return nil
}

// Fügt einen Wagon zu einem Zug, typ gibt die art des wagongs an, daraus basierend wird capacity und maxSpeed bestimmt, bspw "Lebensmittel"
// true => Erfolgreich; false => fehler
func (t *Train) addWaggon(position [3]int, typ string, tiles [][]*Tile, gs *gameState) error {
	var capacity, maxSpeed int
	// Typ gibt kurz als string an was für einen Waggon man will
	// hier werden die passenden Attribute rausgesucht
	switch typ {
	case "Lebensmittel":
		capacity = 30
		maxSpeed = 77
	default:
		logger.Error("Invalider Typ", slog.String("Typ", typ))
		return fmt.Errorf("invalider Typ")
	}

	// Hier wird noch überorüft ob da überhaupt ein freies gleis ist
	if !tiles[position[0]][position[1]].Tracks[position[2]-1] { // Wenn dort ein gleis ist
		return fmt.Errorf("kein Gleis vorhanden")
	}

	// Waggons zu zug hinzufügen und entsprechende tiles blockieren
	t.Waggons = append(t.Waggons, &Waggon{Position: position, MaxSpeed: maxSpeed, CargoStorage: &CargoStorage{capacity: capacity, CargoCategory: typ}})
	tiles[position[0]][position[1]].IsBlocked = true
	var blockedTilesPositions [][2]int
	blockedTilesPositions = append(blockedTilesPositions, [2]int(position[:2]))
	gs.broadcastChannel <- wsEnvelope{Type: "tiles.block", Username: "Server", Msg: blockedTilesMSG{Tiles: blockedTilesPositions}}

	logger.Debug("Blockiertes", slog.Int("Pos 0", position[0]), slog.Int("Pos 1", position[1]), slog.Bool("Blocked", tiles[position[0]][position[1]].IsBlocked))
	return nil
}

// returned Tile zum entblcken
// wenn fertig mit Laden/entladen passiert ein Tick nichts und dann fährt er los
func (t *Train) calculateTrain(gs *gameState) [2]int {

	//bestimmt, ob man beim letzten Pathfining erfolgreich war. wenn ja und man den nächsten sucht, dann wird der nächste Stop ausgewählt
	//sonst wird neu versucht
	pathfindToNextAndMove := func() [2]int {
		//wie in Variablennamen beschrieben
		if t.FoundPathToNext {
			t.NextStop = t.Schedule.nextStop(t.NextStop)
		}
		t.recalculatePath(gs)
		//aktualisierung der Variable (siehe Variablenbeschreibung)
		if len(t.CurrentPath) == 0 {
			t.FoundPathToNext = false
		} else {
			t.FoundPathToNext = true
			return t.move(true, gs)
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
		if t.LoadingTime >= gs.minLoadUloadTicks && t.FinishedLoading {
			logger.Debug("Zug " + t.Name + " versucht aus " + t.CurrentStop.Plattform.station.Name + " auszufahren.")

			if len(t.CurrentPath) == 0 {
				r = pathfindToNextAndMove()
			} else {
				r = t.move(false, gs)
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
		if t.Waiting || t.LoadingTime < gs.minLoadUloadTicks || !t.FinishedLoading {
			t.FinishedLoading = t.loadUndload(gs)

			printTrains(gs)
		}
		t.LoadingTime++

		//wenn er sich nicht bewegt hat
		return [2]int{-1, -1}
	}
	if len(t.CurrentPath) == 0 {
		return pathfindToNextAndMove()
	}
	return t.move(false, gs)
}

// returnt ob der Zug voll ist oder nichts mehr zu laden ist, also abfahrtsbereit ist
func (t *Train) loadUndload(gs *gameState) bool {
	var r bool

	//station, in die der Zug steht
	sta := t.CurrentStop.Plattform.station

	//es wird durch die Reihenfolge der Commands zuerst geladen, dann entladen.
	// Dabei wird nur beladen, wenn entladen fertig ist, bzw. noch kapazität von Gütern bewegt pro Tick über gelassen hat
	avaliableLoadUnloadSpeed := gs.loadUnloadSpeed //misst, wie viel noch geladen und entladen werden darf
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
					loaded = avaliableLoadUnloadSpeed - t.loadCargo(cargo, avaliableLoadUnloadSpeed, gs) //hinzufügen in den Zug
				} else {
					loaded = sta.Storage[cargo] - t.loadCargo(cargo, sta.Storage[cargo], gs)
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
func (t *Train) loadCargo(cargoType string, toLoad int, gs *gameState) int {
	var r int

	for _, waggon := range t.Waggons {
		//wenn nichts mehr zu laden ist, breche ab
		if toLoad == 0 {
			break
		}

		//wenn Waggon richtigen CargoType hat, wenn er schon gefüllt ist, wird gefüllt, oder wenn leer ist, die passende Category hat
		if waggon.CargoStorage != nil {

			if (waggon.CargoStorage.filled == 0 && waggon.CargoStorage.CargoCategory == getCargoCategory(cargoType, gs)) ||
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
func (t *Train) move(wasRecalculated bool, gs *gameState) [2]int {
	entblocken := [2]int{-1, -1}
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
	fmt.Println("Train", t.Name, "newGenNoSignal", newGenNoSignal, "Signale:", signals, "Weg:", path)
	if newGenNoSignal || len(signals) > 1 && t.Waggons[0].Position == signals[0] {
		//gucken, ob bis zum nächsten Signal alle Tiles unblocked sind, sonst fahre nicht weiter
		// (es wird immer auch das letzte Tile überprüft, da man über ein sub tile ohne signal fahren muss, um zu einem zu kommen)
		// --> wichtig für Stationen, immer letzte Subtile ansteuern
		for i := 0; (newGenNoSignal && path[i] != signals[0]) ||
			!newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			if gs.Tiles[path[i][0]][path[i][1]].IsBlocked {
				logger.Debug("Zug " + t.Name + ": Blocked Tile found: []" + strconv.Itoa(path[i][0]) + ", " + strconv.Itoa(path[i][1]) + ", " + strconv.Itoa(path[i][2]) + ". Waiting")

				if !t.Waiting {
					//TODO
					logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
						fmt.Sprint(t.Waggons[0].Position) + " auf rot geschaltet werden soll")
				}

				t.Waiting = true
				return [2]int{-1, -1}
			}
		}
		//da nichts geblocked war, blockt dieser Zug jetzt die Strecke zum nächsten Signal
		var blockedTilesPositions [][2]int
		for i := 0; newGenNoSignal && path[i] != signals[0] || !newGenNoSignal && path[i] != signals[1] && i < len(path); i++ {
			gs.Tiles[path[i][0]][path[i][1]].IsBlocked = true
			blockedTilesPositions = append(blockedTilesPositions, [2]int{path[i][0], path[i][1]})
		}

		// dann können die clients auch nett anzeigen welche jetzt blockiert sind :D , brauche ich vielleicht auch fürs debuggen :D :D
		gs.broadcastChannel <- wsEnvelope{Type: "tiles.block", Username: "Server", Msg: blockedTilesMSG{Tiles: blockedTilesPositions}}
		//nun wird das Signal aus der Queue rausgenommen, da der Zug über das Signal fährt
		t.CurrentPathSignals = t.CurrentPathSignals[1:]

		logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
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
			logger.Debug("Hallo Jannis, hier Nachricht an client, dass das Signal an Stelle " +
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

	// Alles was bis hier gekmmen ist hat sich bewegt (laut wilken beim döner essen)
	gs.broadcastChannel <- wsEnvelope{Type: "train.move", Msg: trainMoveMSG{Id: t.Id, Waggons: t.Waggons}}

	return entblocken
}

func handleTrainUpdate(envelope recieveWSEnvelope, gs *gameState) error {
	switch envelope.Type {
	case "train.create":
		return handleCreateTrain(envelope, gs)
	case "train.remove":
		return handleRemoveTrain(envelope, gs)
	default:
		return fmt.Errorf("unknown envelope Type")
	}
}

func handleRemoveTrain(envelope recieveWSEnvelope, gs *gameState) error {
	var update trainRemoveMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	logger.Info("Tryring to remove Train", slog.String("Username", envelope.user.username), slog.Int("Train ID", update.Id))
	train, prs := gs.Trains[update.Id]
	if prs {
		gs.broadcastChannel <- wsEnvelope{
			Type:          "train.remove",
			Username:      "Server",
			TransactionID: envelope.TransactionID,
			Msg:           trainRemoveMSG{Id: update.Id},
		}
		return train.removeTrain(gs)
	} else {
		return fmt.Errorf("no matching train found, id: %d", update.Id)
	}
}

func handleCreateTrain(envelope recieveWSEnvelope, gs *gameState) error {
	var update trainCreateMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	train, err := addTrain(update, gs)
	if err != nil {
		return fmt.Errorf("error creating train; %s", err.Error())
	}

	logger.Info("Creating Train", slog.String("Username", envelope.user.username), slog.String("Zug Name", update.Name))
	gs.broadcastChannel <- wsEnvelope{
		Type:          "train.create",
		Username:      "Server",
		TransactionID: envelope.TransactionID,
		Msg:           train,
	}
	return nil
}
