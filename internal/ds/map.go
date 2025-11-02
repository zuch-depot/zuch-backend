package ds

type Stop struct {
	Id                int //NOTWENDIG
	Plattform         *Plattform
	Goal              [3]int //(?Signal als) Wegpunkt
	IsPlattform       bool
	Name              string               //Name Wegpunkt
	LoadUnloadCommand [2]LoadUnloadCommand //einmal zum Laden, einmal zum entladen (0 entladen, 1 beladen)
}

type LoadUnloadCommand struct {
	//wenn Loading, wenn false, dann kurz warten und auch wenn nicht voll trotzdem fahren
	//wenn Unloading, wenn false, dann "WaitTillEmpty", also warten, bis alles entladen werden kann, oder einfach weiterfahren
	WaitTillFull bool
	Loading      bool     //wenn false, dann unloading
	CargoType    []string //welche Güter abgeladen/aufgeladen werden dürfen
}
