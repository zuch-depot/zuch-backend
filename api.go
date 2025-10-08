package main

import "net/http"

func handleSaveRequest(w http.ResponseWriter, r *http.Request, gs *gameState) {
	saveGame(gs)
	w.WriteHeader(202)
}
func handlePauseGame(w http.ResponseWriter, r *http.Request, gs *gameState) {
	pauseGame(gs)
	w.WriteHeader(202)

}

func handleUnpauseGame(w http.ResponseWriter, r *http.Request, gs *gameState) {
	unPauseGame(gs)
	w.WriteHeader(202)

}

// Benutzt um rückmeldung zuu geben das der aktuelle tick vorbei ist und vorm nächsten pausiert wurde
var confirmPause = make(chan bool)

func pauseGame(gs *gameState) {
	gs.isPaused = true
	<-confirmPause
	logger.Info("Paused Game")
}

func unPauseGame(gs *gameState) {
	gs.unPause <- true
	logger.Info("Unpaused Game")

}
