package api

import (
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

func StartServer(gs *ds.GameState) {

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

	// hier werden die einzelnen routen hinzugefügt, da steht auch genau drinne was wie welche methode nutzt
	registerGameRoutes(&api, gs)
	registerSignalRoutes(&api, gs)
	registerTrackRoutes(&api, gs)
	registerTrainRoutes(&api, gs)
	registerTileRoutes(&api, gs)
	registerScheduleRoutes(&api, gs)
	registerStationRoutes(&api, gs)

	gs.Logger.Error("error running Webserver", slog.String("Error", http.ListenAndServe("0.0.0.0:"+os.Getenv("PORT"), router).Error()))

}

func StartListiningToBroadcast(broadcastChannel <-chan ds.WsEnvelope, gs *ds.GameState) {
	for {
		envelope, ok := <-broadcastChannel
		if ok {
			for _, user := range gs.Users {
				if user.IsConnected {
					gs.Logger.Debug("Notifying client of Change", slog.String("User", user.Username), slog.String("Type", envelope.Type))
					user.WebSocketQueue <- envelope
				}
			}
		}
	}
}
