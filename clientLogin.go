package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type User struct {
	username       string
	isConnected    bool
	connection     *websocket.Conn
	webSocketQueue chan wsEnvelope
}

// Folgendes Problem, ich hab 2 Channel aus denen nachrichten für diesen Client kommen können, einmal der Broadcast Channel auf dem alle clients sind und der channel der nur für den einzelnen Client ist
// hier werden beide zusammengefügt und dann abgearbeitet

func (user *User) startNotifiyingSingleClient() {
	// Kombiniert den Broadcast und den einzelnen Channel, damit der client über die gleiche verbindung beides erhält
	for {
		if user.isConnected {

			envelope := <-user.webSocketQueue

			logger.Debug("Notifying client of event", slog.String("user", user.username), slog.String("Event Type", envelope.Type))
			err := user.connection.WriteJSON(envelope)
			if err != nil {
				logger.Error("Could not Write to client", slog.String("Error", err.Error()), slog.String("Username", user.username))
			}
		} else {
			break
		}

	}
}

// Wird genutzt um HTTP anfragen zu Websockets zu upgraden
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func startServer(gs *gameState) {

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		acceptNewClient(w, r, gs)
	})
	// die werden später noch als teil des WS umgesetzt (denke ich mal), aber zum testen erstmal so
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		handleSaveRequest(w, r, gs)
	})
	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		handlePauseGame(w, r, gs)
	})
	http.HandleFunc("/unpause", func(w http.ResponseWriter, r *http.Request) { // Muss so gelöst werden damit ich noch die referenz zum Gamestate übertragen kann
		handleUnpauseGame(w, r, gs)
	})

	logger.Error("error running Webserver", slog.String("Error", http.ListenAndServe("localhost:"+os.Getenv("PORT"), nil).Error()))

}

func acceptNewClient(w http.ResponseWriter, r *http.Request, gs *gameState) {
	username := r.URL.Query().Get("username")
	if username == "Server" {
		http.Error(w, "User already Connected", 400)
		logger.Info("User Tried to login as \"Server\", but that name is reserved", slog.String("Username", username))
		return
	}
	userExists := false
	for _, v := range gs.Users {
		if v.username == username { // Username has already connected at some point
			userExists = true
			if !v.isConnected { // reconnect user
				logger.Info("Reconnecting previously disconnected User", slog.String("Username", v.username))
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					logger.Error("Failed to Upgrade connection")
				}
				v.connection = conn
				v.webSocketQueue = make(chan wsEnvelope, 100)
				v.isConnected = true
				initializeClient(v, gs)

				go checkForClientInput(v, gs)

			} else { // Deny User
				http.Error(w, "User already Connected", 400)
				logger.Info("User tried to connect but name is already in use", slog.String("Username", v.username))
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

		user := User{username: username, isConnected: true, connection: conn, webSocketQueue: make(chan wsEnvelope, 100)}
		gs.Users = append(gs.Users, &user)

		initializeClient(&user, gs)
		go checkForClientInput(&user, gs)
	}
}

type gamestateTemp struct {
	Users     []*User
	Schedules []*Schedule
	Stations  []*Station
	Tiles     [][]*Tile
	Trains    []*Train
}

func initializeClient(user *User, gs *gameState) {
	pauseGame(gs)

	// Hier muss ich erstmal alles am stück einmal rüber senden
	// Der Kriegt quasi einmal den Savefile zugeschickt und danach nur noch die änderungen

	logger.Info("Sending Client Gamestate", slog.String("Username", user.username))

	//  Das hier wird noch geändert werden müssen, gs hat eventuell zu viele infos, da müsste man mal schauen welche gesendet werden sollen
	envelope := wsEnvelope{Type: "game.initialLoad", Msg: gs}
	err := user.connection.WriteJSON(envelope)
	if err != nil {
		logger.Error("Failes parsing state to JSON", slog.String("Error", err.Error()))
	}
	go user.startNotifiyingSingleClient()
	unPauseGame(gs)
}

func checkForClientInput(user *User, gs *gameState) {
	// Für jeden User der Connected ist läuft die funktion hier die ganze Zeit in einer Goroutine
	// Wird in einput Empfangen wird er als recieveWSEnvelope in den userInputs Channel / Queue gelegt
	// Von dort aus wird er dann abgearbeitet
	for {
		var v recieveWSEnvelope
		err := user.connection.ReadJSON(&v)
		if err != nil {
			logger.Warn(user.username+": Error while checking for input, Closing Connection", slog.String("Error", err.Error())) //logger or log?
			// Bei fehlern werden die Clients mit gewalt disconnected, müssen se sich halt wieder neu verbinden (oder einfach keine fehler verursachen :D)
			user.isConnected = false
			user.connection.Close()
			user.connection = nil
			return
		}
		// Manche werte kommen nicht direkt aus dem WS sondern werden hier ergänzt
		// User um nachher zugriff auf den Username und auf die connection zu haben, damit man rückmeldung geben kann
		v.user = user
		gs.userInputs <- v
	}
}
