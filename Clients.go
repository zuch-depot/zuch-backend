package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type UserInput struct {
	action    string
	username  string
	parameter interface{}
}

// Wird genutzt um HTTP anfragen zu Websockets zu upgraden
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var playerConnections = make(map[string]*websocket.Conn)

func startServer() {

	http.HandleFunc("/ws", acceptNewClient)

	log.Fatal(http.ListenAndServe("localhost:"+os.Getenv("PORT"), nil))

}

func acceptNewClient(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	for k, _ := range playerConnections {
		if k == username {
			w.WriteHeader(423)
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println("Failed to Upgrade connection")
	}
	logger.Println("Accepted new Client, with username " + username)

	playerConnections[username] = conn

	go checkForClientInput(username, conn)

}

func checkForClientInput(username string, conn *websocket.Conn) {
	for {
		var v UserInput
		err := conn.ReadJSON(&v)
		if err != nil {
			log.Println(username+"; Error while checking for input", err)
		}

		v.username = username
		userInputs <- v
	}
}
