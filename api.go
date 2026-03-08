package main

import (
	"net/http"
	"zuch-backend/internal/ds"
)

func handleSaveRequest(w http.ResponseWriter, r *http.Request, gs *ds.GameState) {
	saveGame(gs, "")
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
