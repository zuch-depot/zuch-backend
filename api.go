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
		return ds.CreateGenericResponse("created signal", err == nil), nil
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
		return ds.CreateGenericResponse("removed signal", err == nil), nil
	}, huma.OperationTags("signal"))
}

// endregion signals
// region game
func registerGameRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/game/pause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		pauseGame(gs)
		return ds.CreateGenericResponse("game paused", true), nil
	}, huma.OperationTags("game"))

	huma.Get(*api, "/game/unpause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		unPauseGame(gs)
		return ds.CreateGenericResponse("game unpaused", true), nil
	}, huma.OperationTags("game"))

	huma.Get(*api, "/game/save", func(ctx context.Context, i *struct{}) (*ds.SaveGameMessage, error) {
		filename, err := saveGame(gs, "")
		if err != nil {
			return nil, fmt.Errorf("Error while Saving game; %s", err.Error())
		}
		msg := &ds.SaveGameMessage{}
		msg.Body.Message = "game saved"
		msg.Body.Success = true
		msg.Body.Path = filename
		return msg, nil
	}, huma.OperationTags("game"))

	huma.Put(*api, "/game/load", func(ctx context.Context, i *ds.SaveGameMessage) (*ds.GenericResponse, error) {
		err := loadGame(gs, "") // hier könnte man später noch den dateinamen mitgeben
		if err != nil {
			return nil, err
		}
		return ds.CreateGenericResponse("Loaded new Save", true), nil
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
		return ds.CreateGenericResponse("created track(s)", true), nil
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
		return ds.CreateGenericResponse("removed track(s)", true), nil

	}, huma.OperationTags("track"))
}

// endregion tracks
// region trains
func registerTrainRoutes(api *huma.API, gs *ds.GameState) {
	// Hier kriegt man infos zu einem Zug
	huma.Get(*api, "/train/{id}", func(ctx context.Context, i *struct {
		id int `path:"id"`
	}) (*ds.Train, error) {
		train, ok := gs.Trains[i.id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		return train, nil
	})
	// Hier kann man einen Zug bauen
	huma.Post(*api, "/train", func(ctx context.Context, i *struct{ Body ds.TrainCreateMSG }) (*ds.GenericResponse, error) {
		_, err := gs.AddTrain(i.Body.Name, i.Body.LocomotivePosition, "") //kein Plan was du so gemacht hast, habe das mal angepasst. Musst mal gucken, ob das so passt. Siehe Funktionsbeschreibung
		if err != nil {
			return nil, fmt.Errorf("Could not create Train; %s", err.Error())
		}
		return ds.CreateGenericResponse("created Train", true), nil
	})
	// Hier kann man einen Zug pausieren
	huma.Post(*api, "/train/{id}/pause", func(ctx context.Context, i *struct {
		id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.Pause()
		return ds.CreateGenericResponse("paused Train", true), nil
	})
	// Hier kann man einen Zug pausieren
	huma.Post(*api, "/train/{id}/unpause", func(ctx context.Context, i *struct {
		id int `path:"id"`
	}) (*ds.GenericResponse, error) {
		train, ok := gs.Trains[i.id]
		if !ok {
			return nil, fmt.Errorf("Train does not exist")
		}
		train.UnPause()
		return ds.CreateGenericResponse("resumed Train", true), nil
	})
}

// endregion trains
// region tiles
func registerTileRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/tiles", func(ctx context.Context, i *struct{}) (*ds.TileMessage, error) {
		mes := &ds.TileMessage{}
		mes.Body.Tiles = gs.Tiles
		return mes, nil
	}, huma.OperationTags("tiles"))
}

// endregion tiles
