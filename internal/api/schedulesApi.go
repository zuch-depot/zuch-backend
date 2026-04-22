package api

import (
	"context"
	"fmt"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region schedules
func registerScheduleRoutes(api *huma.API, gs *ds.GameState) {
	// hier kriegt man alle schedules
	huma.Get(*api, "/schedules", func(ctx context.Context, i *struct{}) (*struct {
		Body struct {
			Schedules map[int]*ds.Schedule
		}
	}, error) {
		return &struct {
			Body struct{ Schedules map[int]*ds.Schedule }
		}{Body: struct{ Schedules map[int]*ds.Schedule }{Schedules: gs.Schedules}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Schedules auslisten"
	})

	// Schedule erstellen
	huma.Post(*api, "/schedule", func(ctx context.Context, i *struct {
		Body struct {
			Name string `example:"fred"`
		}
	}) (*ds.GenericResponse, error) {
		_, err := gs.AddSchedule(i.Body.Name)
		if err != nil {
			return nil, fmt.Errorf("Could not create Schedule; %s", err.Error())
		}
		return ds.CreateGenericResponse("created Schedule"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Schedule erstellen"
	})

	// hier kriegt man infos zu einer schedule
	huma.Get(*api, "/schedule/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id" example:"1"`
	}) (*struct {
		Body struct {
			Schedule ds.Schedule
		}
	}, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		return &struct {
			Body struct{ Schedule ds.Schedule }
		}{Body: struct{ Schedule ds.Schedule }{Schedule: *schedule}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Details zu einer Schedule"
	})

	// hier kann man eine Plattform anhängen
	huma.Post(*api, "/schedule/{id}/append", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			PlattformPos    [2]int    `example:"[2,0]"`
			LoadList        *[]string `example:"Sonnenblumenöl" required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			LoadTillFull    *bool     `example:"false" required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			UnloadList      *[]string `example:"Sonnenblumenöl" required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
			UnloadTillEmpty *bool     `example:"false" required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
		}
	}) (*ds.GenericResponse, error) {

		// erstmal muss ich den Stop erstellen
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}

		plattform, err := gs.GetPlattform(i.Body.PlattformPos)
		if err != nil {
			return nil, fmt.Errorf("Plattform does not exist")
		}
		stop, err := schedule.AddStopStation(plattform, gs)
		if err != nil {
			return nil, fmt.Errorf("could not add Plattform; %s", err.Error())
		}
		// dann sagen was rauf
		if i.Body.LoadList != nil && i.Body.LoadTillFull != nil {
			err = stop.SetLoadCommand(*i.Body.LoadList, *i.Body.LoadTillFull, gs)
			if err != nil {
				return nil, fmt.Errorf("could not set Load Command; %s", err.Error())
			}
		}

		// dann was weg
		if i.Body.UnloadList != nil && i.Body.UnloadTillEmpty != nil {
			err = stop.SetUnloadCommand(*i.Body.UnloadList, *i.Body.UnloadTillEmpty, gs)
			if err != nil {
				return nil, fmt.Errorf("could not set Unload Command; %s", err.Error())
			}
		}

		return ds.CreateGenericResponse("added Stop"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Stop hinzufügen"
	})

	// hier kann man waypoints hinzufügen
	huma.Post(*api, "/schedule/{id}/append-waypoint", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Pos  [3]int `example:"[1,2,3]"`
			Name string `example:"Fred"`
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		_, err := schedule.AddStopWaypoint(i.Body.Pos, i.Body.Name, gs)
		if err != nil {
			return nil, fmt.Errorf("could not add waypoint; %s", err.Error())
		}
		return ds.CreateGenericResponse("added Waypoint"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Waypoint hinzufügen"
	})

	// hier kann man Stops aus einer  schedule löschen, die arme
	huma.Post(*api, "/schedule/{id}/remove-stop", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Index    *int `example:"1"`
			Index_to *int `required:"false" example:"2"`
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}

		var err error
		// wenn man von x bis y löschen will
		if i.Body.Index_to == nil {
			err = schedule.RemoveStop(*i.Body.Index, gs)
		} else {
			// wenn nur x weicht
			err = schedule.RemoveStops(*i.Body.Index, *i.Body.Index_to, gs)
		}
		if err != nil {
			return nil, fmt.Errorf("could not remove Stops; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed Stop(s)"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Stops entfernen"
	})

	// hier kann man die ganze schedule löschen
	huma.Delete(*api, "/schedule/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id" example:"1"`
	}) (*ds.GenericResponse, error) {
		err := gs.RemoveSchedule(i.Id)
		if err != nil {
			return nil, fmt.Errorf("could not remove schedule; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed schedule"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Schedule zerstören"
	})

	// hier kann man eine schedule umbenennen
	huma.Post(*api, "/schedule/{id}/rename", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Name string `example:"Fred"`
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		err := schedule.Rename(i.Body.Name, gs)
		if err != nil {
			return nil, fmt.Errorf("could not rename Schedule; %s", err.Error())
		}
		return ds.CreateGenericResponse("rename Schedule"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Schedule umbennenen"
	})

	// Hier kann man die reihenfolge ändern
	huma.Post(*api, "/schedule/{id}/sequence", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Index     int
			Index_Two int
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		err := schedule.ChangeSquence(i.Body.Index, i.Body.Index_Two)
		if err != nil {
			return nil, fmt.Errorf("could not change sequence; %s", err.Error())
		}
		return ds.CreateGenericResponse("changed sequence"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Reihenfolge einer schedule ändern"
	})

	// Hier kann man eine Schedule updaten
	huma.Post(*api, "/schedule/{id}/change", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Stop_Index      int
			LoadList        *[]string `example:"Sonnenblumenöl" required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			LoadTillFull    *bool     `example:"false" required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			UnloadList      *[]string `example:"Sonnenblumenöl" required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
			UnloadTillEmpty *bool     `example:"false" required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		stop := schedule.Stops[i.Body.Stop_Index]
		if stop == nil {
			return nil, fmt.Errorf("could not find Stop")
		}

		var err error
		// dann sagen was rauf
		if i.Body.LoadList != nil && i.Body.LoadTillFull != nil {
			err = stop.SetLoadCommand(*i.Body.LoadList, *i.Body.LoadTillFull, gs)
			if err != nil {
				return nil, fmt.Errorf("could not set Load Command; %s", err.Error())
			}
		}

		// dann was weg
		if i.Body.UnloadList != nil && i.Body.UnloadTillEmpty != nil {
			err = stop.SetUnloadCommand(*i.Body.UnloadList, *i.Body.UnloadTillEmpty, gs)
			if err != nil {
				return nil, fmt.Errorf("could not set Unload Command; %s", err.Error())
			}
		}
		return ds.CreateGenericResponse("updated Stop"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"schedule"}
		o.Summary = "Stop aktualisieren"
	})
}

// endregion schedules
