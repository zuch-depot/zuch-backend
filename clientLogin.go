package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type User struct {
	username    string
	isConnected bool
	connection  *websocket.Conn
}

type UserInput struct {
	action    string
	username  string
	parameter any
}

// Wird genutzt um HTTP anfragen zu Websockets zu upgraden
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func startServer() {

	http.HandleFunc("/ws", acceptNewClient)
	// die werden später noch als teil des WS umgesetzt (denke ich mal), aber zum testen erstmal so
	http.HandleFunc("/save", handleSaveRequest)
	http.HandleFunc("/pause", handlePauseGame)
	http.HandleFunc("/unpause", handleUnpauseGame)

	logger.Error("error running Webserver", slog.String("Error", http.ListenAndServe("localhost:"+os.Getenv("PORT"), nil).Error()))

}

func acceptNewClient(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	userExists := false
	for _, v := range users {
		if v.username == username { // Username has already connected at some point
			userExists = true
			if !v.isConnected { // reconnect user
				logger.Info("Reconnecting previously disconnected User", slog.String("Username", v.username))
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					logger.Error("Failed to Upgrade connection")
				}
				v.connection = conn
				v.isConnected = true
				initializeClient(v, &gamestateTemp{Users: users, Schedules: []*Schedule{}, Stations: []*Station{}, Tiles: tiles, Trains: []*Train{}})

				go checkForClientInput(v)

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

		user := User{username: username, isConnected: true, connection: conn}
		users = append(users, &user)

		initializeClient(&user, &gamestateTemp{Users: users, Schedules: schedules, Stations: stations, Tiles: tiles, Trains: trains})
		go checkForClientInput(&user)
	}
}

type gamestateTemp struct {
	Users     []*User
	Schedules []*Schedule
	Stations  []*Station
	Tiles     [][]*Tile
	Trains    []*Train
}

func initializeClient(user *User, state *gamestateTemp) {
	pauseGame()

	// Hier muss ich erstmal alles am stück einmal rüber senden
	// Der Kriegt quasi einmal den Savefile zugeschickt und danach nur noch die änderungen

	logger.Info("Sending Client Gamestate", slog.String("Username", user.username))

	envelope := wsEnvelope{Type: "game.initialLoad", Msg: state}
	err := user.connection.WriteJSON(envelope)
	if err != nil {
		logger.Error("Failes parsing state to JSON", slog.String("Error", err.Error()))
	}
	unPauseGame()
}

func checkForClientInput(user *User) {
	for {
		var v recieveWSEnvelope
		err := user.connection.ReadJSON(&v)
		if err != nil {
			logger.Warn(user.username+": Error while checking for input, Closing Connection", slog.String("Error", err.Error())) //logger or log?
			user.isConnected = false
			user.connection.Close()
			user.connection = nil
			return
		}

		v.Username = user.username
		userInputs <- v
	}
}
