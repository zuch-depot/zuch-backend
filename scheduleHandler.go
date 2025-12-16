package main

import (
	"fmt"
	"zuch-backend/internal/ds"
)

func handleScheduleUpdate(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
	switch envelope.Type {
	case "schedule.create":
		return handleCreateStation(envelope, gs)
	case "schedule.remove":
		return handleRemoveStation(envelope, gs)
	case "schedule.assign":
		return handleCreateStation(envelope, gs)
	case "schedule.unassign":
		return handleRemoveStation(envelope, gs)
	default:
		return fmt.Errorf("unknown envelope Type")
	}
}

// func handleCreateSchedule(envelope ds.RecieveWSEnvelope, gs *ds.GameState) error {
// 	var update ds.ScheduleCreateMsg
// 	err := json.Unmarshal(envelope.Msg, &update)
// 	if err != nil {
// 		return fmt.Errorf("could not unpack envelope; %s", err.Error())
// 	}
// 	schedule, err := gs.AddSchedule(update.Name)
// 	if err != nil {
// 		return fmt.Errorf("error creating train; %s", err.Error())
// 	}
// 	for _, entry := range update.Entries {
// 		// stop, _ = schedule.AddStopStation(gs.Stations[0].Plattforms[0], gs)

// 	}

// }
