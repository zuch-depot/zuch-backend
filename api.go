package main

import (
	"context"
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
	saveGame(gs)
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

	api := humachi.New(router, huma.DefaultConfig("Zuch API", "0.1.0"))

	registerGameRoutes(&api, gs)
	registerSignalRoutes(&api, gs)

	// die werden später noch als teil des WS umgesetzt (denke ich mal), aber zum testen erstmal so
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		enableCors(&w)
		handleSaveRequest(w, r, gs)
	})

	logger.Error("error running Webserver", slog.String("Error", http.ListenAndServe("0.0.0.0:"+os.Getenv("PORT"), router).Error()))

}

// Hier werden die HUMA Routen für die Signale erstellt
// Heißt hier werden auch die eigentlichen methoden aufgerufen und fehler abgefangen
func registerSignalRoutes(api *huma.API, gs *ds.GameState) {
	huma.Post(*api, "/signal", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return ds.CreateGenericResponse("Tile not found", false), nil
		}
		success, err := tile.AddSignal(i.Body.Position[2]-1, gs)
		if err != nil {
			return ds.CreateGenericResponse("There was an error while creating the signal", success), nil
		}
		return ds.CreateGenericResponse("created signal", success), err
	}, huma.OperationTags("signal"))

	huma.Delete(*api, "/signal", func(ctx context.Context, i *struct{ Body ds.TileUpdateMSG }) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return ds.CreateGenericResponse("Tile not found", false), nil
		}
		success, err := tile.AddSignal(i.Body.Position[2]-1, gs)
		if err != nil {
			return ds.CreateGenericResponse("There was an error while removing the signal", success), nil
		}
		return ds.CreateGenericResponse("removed signal", success), err
	}, huma.OperationTags("signal"))
}

func registerGameRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/game/pause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		pauseGame(gs)
		return &ds.GenericResponse{Body: struct {
			Message string
			Success bool
		}{Message: "game paused", Success: true}}, nil
	}, huma.OperationTags("game"))

	huma.Get(*api, "/game/unpause", func(ctx context.Context, i *struct{}) (*ds.GenericResponse, error) {
		unPauseGame(gs)
		return &ds.GenericResponse{Body: struct {
			Message string
			Success bool
		}{Message: "game unpaused", Success: true}}, nil
	}, huma.OperationTags("game"))
}
