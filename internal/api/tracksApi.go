package api

import (
	"context"
	"fmt"

	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region tracks
func registerTrackRoutes(api *huma.API, gs *ds.GameState) {
	// used to add tracks, if position_to is specified, the tracks are build in a straight line
	huma.Post(*api, "/track", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.MultitileUpdateMSG
	},
	) (*ds.GenericResponse, error) {
		var err error
		// wenn das gestetzt ist, mehrere gleise bauen
		if i.Body.Position_to != nil {
			err = gs.AddTracks(*i.Body.Position, *i.Body.Position_to)
			if err != nil {
				return nil, fmt.Errorf("could not create track(s); %s", err.Error())
			}
			// sonst nur eins
		} else {
			tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
			if err != nil {
				return nil, fmt.Errorf("Could not find Tile; %s", err.Error())
			}
			err = tile.AddTrack(i.Body.Position[2], gs)
			if err != nil {
				return nil, fmt.Errorf("Tile was found but could not create track(s); %s", err.Error())
			}
		}
		return ds.CreateGenericResponse("created track(s)"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"track"}
		o.Summary = "Gleis(e) bauen"
	})

	// hier um Gleise zu entfernen, zum entfernen von mehreren Gleisen warte ich noch auf wilken
	huma.Delete(*api, "/track", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.MultitileUpdateMSG
	},
	) (*ds.GenericResponse, error) {
		var err error
		// für mehrere
		if i.Body.Position_to != nil {
			err = gs.RemoveTracks(*i.Body.Position, *i.Body.Position_to)
			if err != nil {
				return nil, fmt.Errorf("could not remove track(s); %s", err.Error())
			}
			// um nur eins zu entfernen
		} else {
			// erstmal holen
			tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
			if err != nil {
				return nil, fmt.Errorf("Tile not found; %s", err.Error())
			}
			// dann entfernen
			err = tile.RemoveTrack(i.Body.Position[2], gs)
			if err != nil {
				return nil, fmt.Errorf("Tile was found but could not remove track; %s", err.Error())
			}
		}
		// dann rückmelden
		return ds.CreateGenericResponse("removed track(s)"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"track"}
		o.Summary = "Gleis(e) zerstören"
	})
}

// endregion tracks
