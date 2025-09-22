package main

import "net/http"

func handleSaveRequest(w http.ResponseWriter, r *http.Request) {
	saveGame(users, schedules, stations, tiles, trains)
	w.WriteHeader(202)
}
func handlePauseGame(w http.ResponseWriter, r *http.Request) {
	pauseGame()
	w.WriteHeader(202)

}

func handleUnpauseGame(w http.ResponseWriter, r *http.Request) {
	unPauseGame()
	w.WriteHeader(202)

}

// Benutzt um rückmeldung zuu geben das der aktuelle tick vorbei ist und vorm nächsten pausiert wurde
var confirmPause = make(chan bool)

func pauseGame() {
	isPaused = true
	<-confirmPause
	logger.Info("Paused Game")
}

func unPauseGame() {
	unPause <- true
	logger.Info("Unpaused Game")

}
