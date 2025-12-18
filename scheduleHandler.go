package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"zuch-backend/internal/ds"
)

func handleScheduleUpdate(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	switch envelope.Type {
	case "schedule.create":
		return handleCreateSchedule(envelope, gs)
	case "schedule.remove":
		return handleRemoveSchedule(envelope, gs)
	case "schedule.assign":
		return handleCreateStation(envelope, gs)
	case "schedule.unassign":
		return handleRemoveStation(envelope, gs)
	default:
		return fmt.Errorf("unknown envelope Type")
	}
}

func handleRemoveSchedule(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.ScheduleRemoveMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	err = gs.RemoveSchedule(update.Id)
	if err != nil {
		return fmt.Errorf("error creating train; %s", err.Error())
	}

	gs.Logger.Info("Remove Schedule", slog.String("Username", envelope.User.Username))
	gs.BroadcastChannel <- ds.WsEnvelope{
		Type:          "schedule.remove",
		Username:      "Server",
		TransactionID: envelope.TransactionID,
		Msg:           &ds.ScheduleRemoveMSG{Id: update.Id},
	}
	return nil
}

func handleCreateSchedule(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	var update ds.ScheduleCreateMSG
	err := json.Unmarshal(envelope.Msg, &update)
	if err != nil {
		return fmt.Errorf("could not unpack envelope; %s", err.Error())
	}
	schedule, err := gs.AddSchedule(update.Name)
	if err != nil {
		return fmt.Errorf("error creating train; %s", err.Error())
	}
	for _, entry := range update.Entries {
		// stop, _ = schedule.AddStopStation(gs.Stations[0].Plattforms[0], gs)
		station, ok := gs.Stations[entry.StationId]
		if !ok {
			return fmt.Errorf("could not find station")
		}
		plattform, ok := station.Plattforms[entry.PlattformId]
		if !ok {
			return fmt.Errorf("could not find plattform within station")
		}

		stop, err := schedule.AddStopStation(plattform, gs)
		if err != nil {
			return err
		}

		err = stop.SetLoadCommand(entry.LoadStrings, entry.WaitTillFull)
		if err != nil {
			return err
		}
		err = stop.SetUnloadCommand(entry.UnloadString, entry.WaitTillEmpty)
		if err != nil {
			return err
		}
	}
	gs.Logger.Info("Creating Schedule", slog.String("Username", envelope.User.Username))
	gs.BroadcastChannel <- ds.WsEnvelope{
		Type:          "schedule.create",
		Username:      "Server",
		TransactionID: envelope.TransactionID,
		Msg:           schedule,
	}
	return nil

}
