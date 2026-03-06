package main

/*
	Nur zum Verwalten von Zügen, nicht für funktionsweise
*/

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"zuch-backend/internal/ds"
)

/*
// Fügt einen Zug hinzu, anhand eines namens und der position sowie des typen und positionen der waggons
func addTrain(update ds.TrainCreateMSG, gs *ds.GameState) (*ds.Train, error) {

	// Weg muss ja frei sein, und alles müssen zusammenhängen
	err := ds.checkIfWaggonsAreValid(update.Waggons, gs)
	if err != nil {
		return nil, err
	}

	name := update.Name
	if name == "" {
		name = fmt.Sprint(gs.CurrentTrainID.Load())
	}

	train := &ds.Train{Name: name, Id: int(gs.CurrentTrainID.Load())}
	gs.CurrentTrainID.Add(1)

	for _, waggon := range update.Waggons {
		err := train.AddWaggon(waggon.Position, waggon.Typ, gs)
		if err != nil {
			return nil, fmt.Errorf("this shoudln't happen; %s", err.Error())
		}
	}

	gs.Trains[train.Id] = train
	return train, nil
}*/

/*
// Überprüft ob alle waggons eine Valide Position haben, also das Gleis nicht blockiert ist, das gleis existiert und die Waggons zusammenhängend sind
func checkIfWaggonsAreValid(waggons []ds.TrainCreateWaggons, gs *ds.GameState) error {
	for i, waggon := range waggons {
		if gs.Tiles[waggon.Position[0]][waggon.Position[1]].IsBlocked {
			return fmt.Errorf("track is blocked")
		}
		// Schaut ob der n. Waggon ein Nachbar des n-1. Waggon ist, daher beim 0. Überspringen
		if i != 0 {
			prevWaggon := waggons[i-1]
			possibleTracks := ds.NeighbourTracks(prevWaggon.Position[0], prevWaggon.Position[1], prevWaggon.Position[2], gs)

			if !slices.Contains(possibleTracks, waggon.Position) {
				return fmt.Errorf("waggons are not continuos or a track is missing")
			}
		}
	}
	// Gibt einen Fehler zurück falls
	return nil
}*/

func handleTrainUpdate(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	switch envelope.Type {
	case "train.create":
		return handleCreateTrain(envelope, gs)
	case "train.remove":
		return handleRemoveTrain(envelope, gs)
	default:
		return fmt.Errorf("unknown envelope Type")
	}
}

func handleRemoveTrain(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.TrainRemoveMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	logger.Info("Tryring to remove Train", slog.String("Username", envelope.User.Username), slog.Int("Train ID", update.Id))
	train, prs := gs.Trains[update.Id]
	if prs {
		gs.BroadcastChannel <- ds.WsEnvelope{
			Type:          "train.remove",
			Username:      "Server",
			TransactionID: envelope.TransactionID,
			Msg:           ds.TrainRemoveMSG{Id: update.Id},
		}
		return gs.RemoveTrain(train)
	} else {
		return fmt.Errorf("no matching train found, id: %d", update.Id)
	}
}

func handleCreateTrain(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.TrainCreateMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	train, err := gs.AddTrain(update.Name, update.Waggons)
	if err != nil {
		return fmt.Errorf("error creating train; %s", err.Error())
	}

	logger.Info("Creating Train", slog.String("Username", envelope.User.Username), slog.String("Zug Name", update.Name))
	gs.BroadcastChannel <- ds.WsEnvelope{
		Type:          "train.create",
		Username:      "Server",
		TransactionID: envelope.TransactionID,
		Msg:           train,
	}
	return nil
}
