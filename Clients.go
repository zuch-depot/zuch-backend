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

func handleSaveRequest(w http.ResponseWriter, r *http.Request) {
	saveGame(users, schedules, stations, tiles, trains)
	w.WriteHeader(202)
}
func handlePauseGame(w http.ResponseWriter, r *http.Request) {
	isPaused = true
	logger.Info("Paused Game")
}

func handleUnpauseGame(w http.ResponseWriter, r *http.Request) {
	unpause <- true
	logger.Info("Unpaused Game")
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
				initializeClient(v)
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

		initializeClient(&user)
		go checkForClientInput(&user)
	}
}

func initializeClient(user *User) {
	user.connection.WriteJSON("Hier kommt sicherlich bald eine nette funktion hin die einem alles für den anfang schickt :)")
}

func checkForClientInput(user *User) {
	for {
		var v UserInput
		err := user.connection.ReadJSON(&v)
		if err != nil {
			logger.Warn(user.username+": Error while checking for input, Closing Connection", slog.String("Error", err.Error())) //logger or log?
			user.isConnected = false
			user.connection = nil
			return
		}

		v.username = user.username
		userInputs <- v
	}
}
