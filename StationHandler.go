package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"zuch-backend/internal/ds"
)

func handleStationUpdate(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	switch envelope.Type {
	case "station.create":
		return handleCreateStation(envelope, gs)
	case "station.remove":
		return handleRemoveStation(envelope, gs)
	default:
		return fmt.Errorf("unknown envelope Type")
	}
}

func handleRemoveStation(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.StationUpdateMsg
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	station, err := gs.RemoveStationTile(update.Position)
	if err != nil {
		return fmt.Errorf("error removing station; %s", err.Error())
	}

	if err == nil {
		gs.Logger.Info("Removing Station", slog.String("Username", envelope.User.Username), slog.Int("x", update.Position[0]), slog.Int("y", update.Position[1]))
		gs.BroadcastChannel <- ds.WsEnvelope{
			Type:          "station.remove",
			Username:      "Server",
			TransactionID: envelope.TransactionID,
			Msg:           station,
		}
	} else {
		return fmt.Errorf("error removing station; %s", err.Error())
	}
	return err
}

func handleCreateStation(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.StationUpdateMsg
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	station, err := gs.AddStationTile(update.Position)
	if err != nil {
		return fmt.Errorf("error Creating Station; %s", err.Error())
	}

	gs.Logger.Info("Creating Station", slog.String("Username", envelope.User.Username), slog.Int("x", update.Position[0]), slog.Int("y", update.Position[1]))
	gs.BroadcastChannel <- ds.WsEnvelope{
		Type:          "station.create",
		Username:      "Server",
		TransactionID: envelope.TransactionID,
		Msg:           station,
	}

	return err
}
