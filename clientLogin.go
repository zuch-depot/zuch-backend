package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"zuch-backend/internal/ds"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

// Wird genutzt um HTTP anfragen zu Websockets zu upgraden
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // erstmal bei cors alles erlauben
}

func startServer(gs *ds.GameState) {

	router := chi.NewMux()
	router.Use(middleware.Logger)    // schreibt nett mit
	router.Use(middleware.Recoverer) // sollte machen das der server nicht crasht, sondern das abgefangen und geloggt wird
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
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

func acceptNewClient(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	username := r.URL.Query().Get("username")
	if username == "Server" {
		http.Error(w, "User already Connected", 400)
		logger.Info("User Tried to login as \"Server\", but that name is reserved", slog.String("Username", username))
		return
	}
	userExists := false
	for _, v := range gs.Users {
		if v.Username == username { // Username has already connected at some point
			userExists = true
			if !v.IsConnected { // reconnect user
				logger.Info("Reconnecting previously disconnected User", slog.String("Username", v.Username))
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					logger.Error("Failed to Upgrade connection")
				}
				v.Connection = conn
				v.WebSocketQueue = make(chan ds.WsEnvelope, 100)
				v.IsConnected = true
				initializeClient(v, gs)

				go checkForClientInput(v, gs)

			} else { // Deny User
				http.Error(w, "User already Connected", 400)
				logger.Info("User tried to connect but name is already in use", slog.String("Username", v.Username))
				return
			}
		}

	}
	if !userExists {

		// Create User
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("Failed to Upgrade connection")
		}

		logger.Info("Accepted new Client, with username "+username, slog.String("Username", username))

		user := &ds.User{Username: username, IsConnected: true, Connection: conn, WebSocketQueue: make(chan ds.WsEnvelope, 100)}
		gs.Users[username] = user

		initializeClient(user, gs)
		go checkForClientInput(user, gs)
	}
}

func initializeClient(user *ds.User, gs *ds.GameState) {
	pauseGame(gs)

	// Hier muss ich erstmal alles am stück einmal rüber senden
	// Der Kriegt quasi einmal den Savefile zugeschickt und danach nur noch die änderungen

	logger.Info("Sending Client Gamestate", slog.String("Username", user.Username))

	//  Das hier wird noch geändert werden müssen, gs hat eventuell zu viele infos, da müsste man mal schauen welche gesendet werden sollen
	stationMap := make(map[int]*ds.Station)
	for _, v := range gs.Stations {
		stationMap[v.Id] = v
	}

	envelope := ds.WsEnvelope{Type: "game.initialLoad", Username: "Server", TransactionID: "", Msg: ds.SendAbleGamestate{Users: gs.Users, Schedules: gs.Schedules, Stations: stationMap, Tiles: gs.Tiles, Trains: gs.Trains}}
	err := user.Connection.WriteJSON(envelope)
	if err != nil {
		logger.Error("Failed parsing state to JSON", slog.String("Error", err.Error()))
	}
	go user.StartNotifiyingSingleClient(gs)
	unPauseGame(gs)
}

func checkForClientInput(user *ds.User, gs *ds.GameState) {
	// Für jeden User der Connected ist läuft die funktion hier die ganze Zeit in einer Goroutine
	// Wird in einput Empfangen wird er als recieveWSEnvelope in den userInputs Channel / Queue gelegt
	// Von dort aus wird er dann abgearbeitet
	for {
		var v ds.RecieveWSEnvelope
		err := user.Connection.ReadJSON(&v)
		if err != nil {
			gs.Logger.Warn(user.Username+": Error while checking for input, Closing Connection", slog.String("Error", err.Error())) //logger or log?
			// Bei fehlern werden die Clients mit gewalt disconnected, müssen se sich halt wieder neu verbinden (oder einfach keine fehler verursachen :D)
			user.IsConnected = false
			user.Connection.Close()
			user.Connection = nil
			return
		}
		// Manche werte kommen nicht direkt aus dem WS sondern werden hier ergänzt
		// User um nachher zugriff auf den Username und auf die connection zu haben, damit man rückmeldung geben kann
		v.User = user
		gs.UserInputs <- v
	}
}
