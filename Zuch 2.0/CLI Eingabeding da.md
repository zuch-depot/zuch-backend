# wie soll das aussehen
- man tippert in eine textzeile was ein
- geht an den server
- der antwortet nett als text  
- Zeugs das für die anzeige relevant ist, kommt über websocket auch noch eine antwort
	- bpsw. wie vorher train.create oder so 
	- 
create train \[1 2 3] \[4 5 6] Holzwagon
=> das ist aber keine linie 
create train \[1 2 3] \[1 5 1] Holzwagon

case incencitiv?

**show**
- train \<name\>
- schedule \<name\>
- activeTile \<name\>
- station \<name\>
- ?user \<name\>

**list**
- trains
- schedules
- activeTiles
- stations
- users

**create/delete**
- train \<name\> \<x, y, z\> \<x, y, z\> ([\<Anzahl\>] \<Waggonart\>)*
	- wenn man Anzahl an Waggons angibt, kann man beliebig viele verschiedene Waggonarten aneinanderreihen
- schedule \<name\>
- track \<x, y[, \<z\>]\>[ \<x, y[, \<z\>]\>]
	- entweder mit z oder ohne, dann aber bei beiden
		- ohne z werden alle aus dem Tile gelöscht
- signal \<x, y[, \<z\>]\>
	- entweder mit z oder ohne, dann aber bei beiden
		- ohne z werden alle aus dem Tile gelöscht
- stationTile \<x, y\> [\<x, y\>]
	- versucht aus dem Tile ein StationTile zu machen oder zu entfernen
- bei allen
	- -f
		- alles richtige ausführen so weit es geht
		- ansonsten wird abgebrochen, wenn der Befehl einen Fehler wirft

**edit**
- train \<name\>
	- rename \<name\>
	- remove waggon \<index\>
	- remove waggons \<index\> \<index\>
	- append waggon
- schedule \<name\>
	- rename \<name\>
	- add stop \<PlattformName\>
- station \<name\>
	- rename plattform \<name\>

**pause**
- train \<name\>

**unpause**
- train \<name\>