package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func handleSaveRequest(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	saveGame(gs, "")
	w.WriteHeader(202)
}
func handlePauseGame(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	pauseGame(gs)
	w.WriteHeader(202)

}

func handleUnpauseGame(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	unPauseGame(gs)
	w.WriteHeader(202)

}

// Benutzt um Rückmeldung zu geben das der aktuelle Tick vorbei ist und vorm nächsten pausiert wurde
var confirmPause = make(chan bool)

func pauseGame(gs *ds.GameState) {
	gs.IsPaused = true
	<-confirmPause
	logger.Info("Paused Game")
}

func unPauseGame(gs *ds.GameState) {
	gs.UnPause <- true
	logger.Info("Unpaused Game")

}

func startServer(gs *ds.GameState) {

	router := chi.NewMux()
	router.Use(middleware.Logger)    // schreibt nett mit
	router.Use(middleware.Recoverer) // sollte machen das der server nicht crasht, sondern das abgefangen und geloggt wird
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: false,
	}))

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		enableCors(&w)
		acceptNewClient(w, r, gs)
	})

	config := huma.DefaultConfig("Zuch API", "0.1.0")
	config.DocsRenderer = huma.DocsRendererScalar
	api := humachi.New(router, config)

	registerGameRoutes(&api, gs)
	registerSignalRoutes(&api, gs)
	registerTrackRoutes(&api, gs)
	registerTrainRoutes(&api, gs)
	registerTileRoutes(&api, gs)

	// die werden später noch als teil des WS umgesetzt (denke ich mal), aber zum testen erstmal so
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		enableCors(&w)
		handleSaveRequest(w, r, gs)
	})

	logger.Error("error running Webserver", slog.String("Error", http.ListenAndServe("0.0.0.0:"+os.Getenv("PORT"), router).Error()))

}

// region Signals
// Hier werden die HUMA Routen für die Signale erstellt
// Heißt hier werden auch die eigentlichen methoden aufgerufen und fehler abgefangen
func registerSignalRoutes(api *huma.API, gs *ds.GameState) {
	huma.Post(*api, "/signal", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Could not find Tile; %s", err.Error())
		}
		err = tile.AddSignal(i.Body.Position[2], gs)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not create signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("created signal"), nil
	}, huma.OperationTags("signal"))

	huma.Delete(*api, "/signal", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Tile not found; %s", err.Error())
		}
		err = tile.RemoveSignal(i.Body.Position[2], gs)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not remove signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed signal"), nil
	}, huma.OperationTags("signal"))
}

// endregion signals
// region game
func registerGameRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/game/pause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		pauseGame(gs)
		return ds.CreateGenericResponse("game paused"), nil
	}, huma.OperationTags("game"))

	huma.Get(*api, "/game/unpause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		unPauseGame(gs)
		return ds.CreateGenericResponse("game unpaused"), nil
	}, huma.OperationTags("game"))

	huma.Get(*api, "/game/save", func(ctx context.Context, i *struct{}) (*ds.SaveGameResponse, error) {
		filename, err := saveGame(gs, "")
		if err != nil {
			return nil, fmt.Errorf("Error while Saving game; %s", err.Error())
		}
		msg := &ds.SaveGameResponse{}
		msg.Body.Message = "game saved"
		msg.Body.Success = true
		msg.Body.Path = filename
		return msg, nil
	}, huma.OperationTags("game"))

	huma.Put(*api, "/game/load", func(ctx context.Context, i *ds.SaveGameResponse) (*ds.GenericResponse, error) {
		err := loadGame(gs, "") // hier könnte man später noch den dateinamen mitgeben
		if err != nil {
			return nil, err
		}
		return ds.CreateGenericResponse("Loaded new Save"), nil
	}, huma.OperationTags("game"))
}

// endregion game
// region tracks
func registerTrackRoutes(api *huma.API, gs *ds.GameState) {
	// used to add tracks, if position_to is specified, the tracks are build in a straight line
	huma.Post(*api, "/track", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		var err error
		// wenn das gestetzt ist, mehrere gleise bauen
		if i.Body.Position_to != nil {
			err = gs.AddTracks(*i.Body.Position, *i.Body.Position_to)
			if err != nil {
				return nil, fmt.Errorf("could not create track(s); %s", err.Error())
			}
			// sonst nur eins
		} else {
			tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
			if err != nil {
				return nil, fmt.Errorf("Could not find Tile; %s", err.Error())
			}
			err = tile.AddTrack(i.Body.Position[2], gs)
			if err != nil {
				return nil, fmt.Errorf("Tile was found but could not create track(s); %s", err.Error())
			}
		}
		return ds.CreateGenericResponse("created track(s)"), nil
	}, huma.OperationTags("track"))

	// hier um Gleise zu entfernen, zum entfernen von mehreren Gleisen warte ich noch auf wilken
	huma.Delete(*api, "/track", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		var err error
		// für mehrere
		if i.Body.Position_to != nil {
			err = gs.RemoveTracks(*i.Body.Position, *i.Body.Position_to)
			if err != nil {
				return nil, fmt.Errorf("could not remove track(s); %s", err.Error())
			}
			// um nur eins zu entfernen
		} else {
			// erstmal holen
			tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
			if err != nil {
				return nil, fmt.Errorf("Tile not found; %s", err.Error())
			}
			// dann entfernen
			err = tile.RemoveTrack(i.Body.Position[2], gs)
			if err != nil {
				return nil, fmt.Errorf("Tile was found but could not remove track; %s", err.Error())
			}
		}
		// dann rückmelden
		return ds.CreateGenericResponse("removed track(s)"), nil

	}, huma.OperationTags("track"))
}

// endregion tracks
// region trains
func registerTrainRoutes(api *huma.API, gs *ds.GameState) {
	// Hier kriegt man infos zu einem Zug
	huma.Get(*api, "/train/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.Train, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		return train, nil
	}, huma.OperationTags("train"))

	// Hier kann man einen Zug bauen
	huma.Post(*api, "/train", func(ctx context.Context, i *struct{ Body ds.TrainCreateMSG }) (*ds.GenericResponse, error) {
		_, err := gs.AddTrain(i.Body.Name, i.Body.LocomotivePosition, "") //kein Plan was du so gemacht hast, habe das mal angepasst. Musst mal gucken, ob das so passt. Siehe Funktionsbeschreibung
		if err != nil {
			return nil, fmt.Errorf("Could not create Train; %s", err.Error())
		}
		return ds.CreateGenericResponse("created Train"), nil
	}, huma.OperationTags("train"))

	// Hier kann man einen Zug pausieren
	huma.Post(*api, "/train/{id}/pause", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.Pause(gs)
		return ds.CreateGenericResponse("paused Train"), nil
	}, huma.OperationTags("train"))

	// Hier kann man einen Zug pausieren
	huma.Post(*api, "/train/{id}/unpause", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.UnPause(gs)
		return ds.CreateGenericResponse("resumed Train"), nil
	}, huma.OperationTags("train"))

	// hier kann man Waggons anhängen
	huma.Post(*api, "/train/{id}/append", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Pos        ds.TileUpdateMSG
			Waggontype string
		}
	}) (*ds.GenericResponse, error) {
		// für mehrere in einer geraden linie
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		var err error
		if i.Body.Pos.Position_to != nil {
			err = train.AddWaggon(*i.Body.Pos.Position, i.Body.Waggontype, gs)

		} else {
			err = train.AddWaggons(*i.Body.Pos.Position, *i.Body.Pos.Position_to, i.Body.Waggontype, gs)

		}
		if err != nil {
			return nil, fmt.Errorf("Could not add waggon(s); %s", err.Error())
		}
		return ds.CreateGenericResponse("added waggons to train"), nil
	}, huma.OperationTags("train"))

	// hier kann man züge entfernen
	huma.Delete(*api, "/train/{id}/remove", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			From int
			To   *int `example:"5" required:"false" nullable:"true"`
		}
	}) (*ds.GenericResponse, error) {
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
	}, huma.OperationTags("train"))

	// hier kann man einen zug brutalst umbringen
	huma.Delete(*api, "/train/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		err := gs.RemoveTrain(train)
		if err != nil {
			return nil, fmt.Errorf("could not remove train %s", err.Error())
		}
		return ds.CreateGenericResponse("removed train"), nil
	}, huma.OperationTags("train"))

	// hier könnte man einen zug umbennen
	// aber ich finde die funktion dazu nicht
	huma.Post(*api, "/train/{id}/rename", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Name string
		}
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		err := train.Rename(i.Body.Name)
		if err != nil {
			return nil, fmt.Errorf("could not rename Train; %s", err.Error())
		}
		return ds.CreateGenericResponse("renamed train"), nil
	}, huma.OperationTags("train"))

	// hier kann man schedules zuweisen
	huma.Post(*api, "/train/{id}/schedule", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			ScheduleId int
		}
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		schedule, ok := gs.Trains[i.Body.ScheduleId]
		if !ok {
			return nil, fmt.Errorf("schedule does not exist")
		}

		train.AssignSchedule(schedule.Schedule, gs)
		return ds.CreateGenericResponse("assigned Schedule"), nil
	}, huma.OperationTags("train"))

	// hier kann man schedules zuweisen
	huma.Post(*api, "/train/{id}/schedule_unassign", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.Id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}

		train.UnassignSchedule(gs)
		return ds.CreateGenericResponse("unassigned Schedule"), nil
	}, huma.OperationTags("train"))

}

// endregion trains

// region schedules
func registerScheduleRoutes(api *huma.API, gs *ds.GameState) {
	// hier kriegt man infos zu einer schedule
	huma.Get(*api, "/schedule/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.Schedule, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		return schedule, nil
	}, huma.OperationTags("schedule"))

	// hier kann man eine Plattform anhängen
	huma.Post(*api, "/schedule/{id}/append", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			PlattformPos    [2]int
			LoadList        *[]string `required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			LoadTillFull    *bool     `required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			UnloadList      *[]string `required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
			UnloadTillEmpty *bool     `required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
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
	})

	// hier kann man waypoints hinzufügen
	huma.Post(*api, "/schedule/{id}/append-waypoint", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Pos  [3]int
			Name string
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
	})

	// hier kann man eine schedule löschen, die arme
	huma.Post(*api, "/schedule/{id}/remove-stop", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Index    *int
			Index_to *int `required:"false"`
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}

		var err error
		// wenn man von x bis y löschen will
		if i.Body.Index_to != nil {
			err = schedule.RemoveStop(*i.Body.Index, gs)
		} else {
			// wenn nur x weicht
			err = schedule.RemoveStops(*i.Body.Index, *i.Body.Index_to, gs)
		}
		if err != nil {
			return nil, fmt.Errorf("could not remove Stops; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed Stop(s)"), nil
	})

	// hier kann man die ganze schedule löschen
	huma.Delete(*api, "/schedule/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		err := gs.RemoveSchedule(i.Id)
		if err != nil {
			return nil, fmt.Errorf("could not remove schedule; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed schedule"), nil
	})

	// hier kann man eine schedule umbenennen
	huma.Post(*api, "/schedule/{id}/rename", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Name string
		}
	}) (*ds.GenericResponse, error) {
		schedule, ok := gs.Schedules[i.Id]
		if !ok {
			return nil, fmt.Errorf("Schedule does not exist")
		}
		err := schedule.Rename(i.Body.Name)
		if err != nil {
			return nil, fmt.Errorf("could not rename Station; %s", err.Error())
		}
		return ds.CreateGenericResponse("rename station"), nil
	})

	// Hier kann man die reihenfolge ändern
	huma.Post(*api, "/schedule/{id}/sequence", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
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
	})

	// Hier kann man eine Schedule updaten
	huma.Post(*api, "/schedule/{id}/change", func(ctx context.Context, i *struct {
		Id   int `path:"id"`
		Body struct {
			Stop_Index      int
			LoadList        *[]string `required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			LoadTillFull    *bool     `required:"false" doc:"both LoadList and LoadTillFull both have to be either used or omitted"`
			UnloadList      *[]string `required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
			UnloadTillEmpty *bool     `required:"false" doc:"both UnloadList and UnloadTillFull both have to be either used or omitted"`
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
	})
}

// endregion schedules

// region tiles
func registerTileRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/tiles", func(ctx context.Context, i *struct{}) (*ds.TileMessage, error) {
		mes := &ds.TileMessage{}
		mes.Body.Tiles = gs.Tiles
		return mes, nil
	}, huma.OperationTags("tiles"))
}

// endregion tiles
