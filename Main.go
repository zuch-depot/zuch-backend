package main

import (
	"log/slog"
	"os"
	"time"

	"zuch-backend/internal/api"
	"zuch-backend/internal/ds"
	"zuch-backend/internal/utils"

	"github.com/telemachus/humane"
)

// Zum Debuggen: LevelInfo zu LevelDebug wechseln
var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelInfo}))

// var logger = slog.New(humane.NewHandler(os.Stdout, &humane.Options{AddSource: true, Level: slog.LevelDebug}))

func main() {
	utils.Logger = logger

	gs := ds.GameState{UserInputs: make(chan ds.RecieveWSEnvelope, 300),
		BroadcastChannel: make(chan ds.WsEnvelope, 100),
		UnPause:          make(chan bool),
		SizeSubtile:      4,
		ConfirmPause:     make(chan bool),
		Trains:           make(map[int]*ds.Train),
		Users:            make(map[string]*ds.User),
		Stations:         make(map[int]*ds.Station),
		Schedules:        make(map[int]*ds.Schedule),
		Logger:           logger,
	}
	// err := godotenv.Load("main.env")
	// if err != nil {
	// 	logger.Error("Oh oh ein fehler in den environment variables", slog.String("Error", err.Error()))
	// }

	// wichtig als initialisierung, bevor Züge verarbeitet werden
	gs.LoadConfig()

	// Ablauf
	// beim ersten start (eventuell probieren Dateien einzulesen) sonst defaults setzen
	// Map erstellen
	initializeTiles(&gs)
	createDemoTrains(&gs)
	// sich merken wer wer ist
	// wenn wer rausfliegt sollten die sachen noch da sein

	//lade das akutellste Savegame
	// gs.LoadGame("")

	// hier den Server starten
	go api.StartServer(&gs)
	// Anfangen aus events an clients zu schicken
	go api.StartListiningToBroadcast(gs.BroadcastChannel, &gs)

	gs.Ticker = time.NewTicker(time.Duration(gs.ConfigData.TicksMilisec) * time.Millisecond)

	// go gs.SaveGame("")

	// jeder Tick
	//for gs.Tick = 0; ; gs.Tick++ { //--> gs.tick ist standartmäßig 0, wenn nicht, dann nur, weil das rausgeladen wurde
	for ; ; gs.Tick++ {

		// Wenn pausiert wurde, warten bis entpausiert signal kommt
		if gs.IsPaused {
			gs.ConfirmPause <- true
			<-gs.UnPause // Hier warten bis es wieder entpausiert wird
			gs.IsPaused = false
			logger.Info("continuing after Pause")
		}

		//TEMP fürs testen //MAYBE Autosaves??
		// if gs.Tick%1000 == 0 {
		// go saveGame(&gs, "")
		// }

		// nicht saves in dem bereich
		gs.Mutex.Lock()

		// Züge bewegen
		gs.CalculateTrains()

		if gs.Tick&10 == 2 {
			gs.LoadUndloadTrains()
		}

		// process factorys
		if gs.Tick%20 == 1 {
			gs.ProcessActiveTiles()
		}

		gs.Mutex.Unlock()

		// if gs.Tick == 100 {
		// 	go saveGame(&gs)
		// }

		// das wartet hier bis ein tick ausgelöst wird,
		<-gs.Ticker.C
	}
}
