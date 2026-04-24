package api

import (
	"context"
	"fmt"

	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region trains
func registerTrainRoutes(api *huma.API, gs *ds.GameState) {
	// hier kriegt man alle Züge
	huma.Get(*api, "/trains", func(ctx context.Context, i *struct{ CostsQuery }) (*struct {
		Body struct {
			Trains map[int]*ds.Train
		}
	}, error,
	) {
		return &struct {
			Body struct{ Trains map[int]*ds.Train }
		}{Body: struct{ Trains map[int]*ds.Train }{Trains: gs.Trains}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Züge auflisten"
	})

	// Hier kriegt man infos zu einem Zug
	huma.Get(*api, "/train/{id}", func(ctx context.Context, i *struct {
		CostsQuery
		Id int `path:"id"`
	}) (*struct {
		Body struct {
			Train *ds.Train
		}
	}, error,
	) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		return &struct{ Body struct{ Train *ds.Train } }{Body: struct{ Train *ds.Train }{Train: train}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Details zu Zug"
	})

	// Hier kann man einen Zug bauen
	huma.Post(*api, "/train", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.TrainCreateMSG
	}) (*ds.GenericResponse, error) {
		// TODO: level und lokomotive kann man angeben, auch beim anhängen. Habe erstmal default werte genommen
		cost, _, err := gs.AddTrain(i.Body.Name, i.Body.LocomotivePosition, "Dampflock", 1, true) // TODO: Waggontype und ggf. level mit in die API nehmen, ersetzt Dampflock
		gs.Logger.Debug("Temp, damit cost nicht Fehler wirft", cost)                              // TODO: cost handeln und bool richtig setzten
		if err != nil {
			return nil, fmt.Errorf("Could not create Train; %s", err.Error())
		}
		return ds.CreateGenericResponse("created Train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Zug hinzufügen"
	})

	// Hier kann man einen Zug pausieren
	huma.Post(*api, "/train/{id}/pause", func(ctx context.Context, i *struct {
		CostsQuery
		Id int `path:"id"`
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.Pause(gs)
		return ds.CreateGenericResponse("paused Train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Zug pausieren"
	})

	// Hier kann man einen Zug fortsetzen
	huma.Post(*api, "/train/{id}/unpause", func(ctx context.Context, i *struct {
		CostsQuery
		Id int `path:"id"`
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.UnPause(gs)
		return ds.CreateGenericResponse("resumed Train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Zug entpausieren?"
	})

	// hier kann man Waggons anhängen
	huma.Post(*api, "/train/{id}/append", func(ctx context.Context, i *struct {
		CostsQuery
		Id   int `path:"id"`
		Body struct {
			Pos        ds.MultitileUpdateMSG
			Waggontype string
		}
	},
	) (*ds.GenericResponse, error) {
		// für mehrere in einer geraden linie
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		var err error
		if i.Body.Pos.Position_to == nil {
			var cost int
			cost, err = train.AddWaggon(*i.Body.Pos.Position, i.Body.Waggontype, 1, gs, true) // TODO: cost handeln und bool richtig setzten
			gs.Logger.Debug("Temp, damit cost nicht Fehler wirft", cost)
		} else {
			var cost int
			cost, err = train.AddWaggons(*i.Body.Pos.Position, *i.Body.Pos.Position_to, i.Body.Waggontype, 1, gs, true) // TODO: cost handeln und bool richtig setzten
			gs.Logger.Debug("Temp, damit cost nicht Fehler wirft", cost)
		}
		if err != nil {
			return nil, fmt.Errorf("Could not add waggon(s); %s", err.Error())
		}
		return ds.CreateGenericResponse("added waggons to train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Waggons anhängen"
	})

	// hier kann man züge entfernen
	huma.Delete(*api, "/train/{id}/remove", func(ctx context.Context, i *struct {
		CostsQuery
		Id   int `path:"id"`
		Body struct {
			From int
			To   *int `example:"5" required:"false" nullable:"true"`
		}
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		var err error
		if i.Body.To == nil {
			err = train.RemoveWaggon(i.Body.From, gs)
		} else {
			err = train.RemoveWaggons(i.Body.From, *i.Body.To, gs)
		}
		if err != nil {
			return nil, fmt.Errorf("could not remove waggon(s) %s", err.Error())
		}
		return ds.CreateGenericResponse("removed waggon(s)"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Waggons zerstören"
	})

	// hier kann man einen zug brutalst umbringen
	huma.Delete(*api, "/train/{id}", func(ctx context.Context, i *struct {
		CostsQuery
		Id int `path:"id"`
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		err := gs.RemoveTrain(train)
		if err != nil {
			return nil, fmt.Errorf("could not remove train %s", err.Error())
		}
		return ds.CreateGenericResponse("removed train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Zug ermorden"
	})

	// hier könnte man einen zug umbennen
	huma.Post(*api, "/train/{id}/rename", func(ctx context.Context, i *struct {
		CostsQuery
		Id   int `path:"id"`
		Body struct {
			Name string
		}
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		// das hier meckert irgendwie immer das es gs gar nicht braucht aber das stimmt ja gar nicht :(((
		err := train.Rename(i.Body.Name, gs)
		if err != nil {
			return nil, fmt.Errorf("could not rename Train; %s", err.Error())
		}
		return ds.CreateGenericResponse("renamed train"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Zug unmbennen"
	})

	// hier kann man schedules zuweisen
	huma.Post(*api, "/train/{id}/schedule", func(ctx context.Context, i *struct {
		CostsQuery
		Id   int `path:"id"`
		Body struct {
			ScheduleId int
		}
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		schedule, ok := gs.Schedules[i.Body.ScheduleId]
		if !ok {
			return nil, fmt.Errorf("schedule does not exist")
		}

		train.AssignSchedule(schedule, gs)
		return ds.CreateGenericResponse("assigned Schedule"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Schedule zuweisen"
	})

	// hier kann man schedules von zügen entfernen
	huma.Post(*api, "/train/{id}/schedule_unassign", func(ctx context.Context, i *struct {
		CostsQuery
		Id int `path:"id"`
	},
	) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}

		train.UnassignSchedule(gs)
		return ds.CreateGenericResponse("unassigned Schedule"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"train"}
		o.Summary = "Schedule entfernen"
	})
}

// endregion trains
