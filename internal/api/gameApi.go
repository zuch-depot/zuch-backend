package api

import (
	"context"
	"fmt"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region game
func registerGameRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/game/pause", func(ctx context.Context, i *struct{ CostsQuery }) (*ds.GenericResponse, error) {
		gs.PauseGame()
		return ds.CreateGenericResponse("game paused"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"game"}
		o.Summary = "Spiel pausieren"
	})

	huma.Get(*api, "/game/unpause", func(ctx context.Context, i *struct{ CostsQuery }) (*ds.GenericResponse, error) {
		gs.UnPauseGame()
		return ds.CreateGenericResponse("game unpaused"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"game"}
		o.Summary = "Spiel fortsetzen"
	})

	huma.Get(*api, "/game/save", func(ctx context.Context, i *struct{CostsQuery  }) (*ds.SaveGameResponse, error) {
		filename, err := gs.SaveGame("")
		if err != nil {
			return nil, fmt.Errorf("Error while Saving game; %s", err.Error())
		}
		msg := &ds.SaveGameResponse{}
		msg.Body.Message = "game saved"
		msg.Body.Success = true
		msg.Body.Path = filename
		return msg, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"game"}
		o.Summary = "Spielstand speichern"
	})

	huma.Put(*api, "/game/load", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.SaveGameResponse
	}) (*ds.GenericResponse, error) {
		err := gs.LoadGame("") // hier könnte man später noch den dateinamen mitgeben
		if err != nil {
			return nil, err
		}
		return ds.CreateGenericResponse("Loaded new Save"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"game"}
		o.Summary = "Spielstand laden"
	})
}

// endregion game
