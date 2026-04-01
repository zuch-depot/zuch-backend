package api

import (
	"log/slog"
	"net/http"
	"zuch-backend/internal/ds"

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

func acceptNewClient(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	username := r.URL.Query().Get("username")
	if username == "Server" {
		http.Error(w, "User already Connected", 400)
		gs.Logger.Info("User Tried to login as \"Server\", but that name is reserved", slog.String("Username", username))
		return
	}
	userExists := false
	for _, v := range gs.Users {
		if v.Username == username { // Username has already connected at some point
			userExists = true
			if !v.IsConnected { // reconnect user
				gs.Logger.Info("Reconnecting previously disconnected User", slog.String("Username", v.Username))
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					gs.Logger.Error("Failed to Upgrade connection")
				}
				v.Connection = conn
				v.WebSocketQueue = make(chan ds.WsEnvelope, 100)
				v.IsConnected = true
				initializeClient(v, gs)

				go checkForClientInput(v, gs)

			} else { // Deny User
				http.Error(w, "User already Connected", 400)
				gs.Logger.Info("User tried to connect but name is already in use", slog.String("Username", v.Username))
				return
			}
		}

	}
	if !userExists {

		// Create User
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			gs.Logger.Error("Failed to Upgrade connection")
		}

		gs.Logger.Info("Accepted new Client, with username "+username, slog.String("Username", username))

		user := &ds.User{Username: username, IsConnected: true, Connection: conn, WebSocketQueue: make(chan ds.WsEnvelope, 100)}
		gs.Users[username] = user

		initializeClient(user, gs)
		go checkForClientInput(user, gs)
	}
}

func initializeClient(user *ds.User, gs *ds.GameState) {
	// sonst kann man sich nicht mehr bei pausierten Spielen verbinden
	wasPaused := gs.IsPaused
	if !wasPaused {
		gs.PauseGame()
	}

	// Hier muss ich erstmal alles am stück einmal rüber senden
	// Der Kriegt quasi einmal den Savefile zugeschickt und danach nur noch die änderungen

	gs.Logger.Info("Sending Client Gamestate", slog.String("Username", user.Username))

	//  Das hier wird noch geändert werden müssen, gs hat eventuell zu viele infos, da müsste man mal schauen welche gesendet werden sollen
	stationMap := make(map[int]*ds.Station)
	for _, v := range gs.Stations {
		stationMap[v.Id] = v
	}

	envelope := ds.WsEnvelope{Type: "game.initialLoad", Msg: ds.SendAbleGamestate{Users: gs.Users, Schedules: gs.Schedules, Stations: stationMap, Tiles: gs.Tiles, Trains: gs.Trains}}
	err := user.Connection.WriteJSON(envelope)
	if err != nil {
		gs.Logger.Error("Failed parsing state to JSON", slog.String("Error", err.Error()))
	}
	go user.StartNotifiyingSingleClient(gs)
	// nur weitergehen wenn es vorher nicht pausiert war
	// sonst nur für connection pausieren
	if !wasPaused {
		gs.UnPauseGame()
	}
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
