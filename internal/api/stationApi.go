package api

import (
	"context"
	"fmt"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region station
func registerStationRoutes(api *huma.API, gs *ds.GameState) {
	// hier kriegt man alle Stationen
	huma.Get(*api, "/stations", func(ctx context.Context, i *struct{}) (*struct {
		Body struct {
			Stations map[int]*ds.Station
		}
	}, error) {
		return &struct {
			Body struct{ Stations map[int]*ds.Station }
		}{Body: struct{ Stations map[int]*ds.Station }{Stations: gs.Stations}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Stationen auflisten"
	})

	// hier kann man Inofs zu einer Station bekommen
	huma.Get(*api, "/station/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id" example:"1"`
	}) (*struct {
		Body struct {
			Station ds.Station
		}
	}, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		return &struct{ Body struct{ Station ds.Station } }{Body: struct{ Station ds.Station }{Station: *station}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Details zu einer Station"
	})

	// hier kann man eine Station umbenennen
	huma.Post(*api, "/station/{id}/rename", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Name string `example:"Fred"`
		}
	}) (*ds.GenericResponse, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		err := station.Rename(i.Body.Name)
		if err != nil {
			return nil, fmt.Errorf("could not rename Station; %s", err.Error())
		}
		return ds.CreateGenericResponse("renamed station"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Station umbennenen"
	})

	// hier kann man die ganze Station löschen
	huma.Delete(*api, "/station/{id}", func(ctx context.Context, i *struct {
		Id int `path:"id" example:"1"`
	}) (*ds.GenericResponse, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		// das ding lässt einen auch zeugs löschen das es nicht gibt
		err := gs.RemoveStation(station)
		if err != nil {
			return nil, fmt.Errorf("could not remove Station; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed station"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Station zerstören"
	})

	// hier kriegt man einfos über eine Plattform
	huma.Get(*api, "/station/{id}/plattform/{idPlat}", func(ctx context.Context, i *struct {
		Id     int `path:"id" example:"1"`
		IdPlat int `path:"idPlat" example:"1"`
	}) (*struct {
		Body struct {
			Plattform ds.Plattform
		}
	}, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		platform, ok := station.Plattforms[i.IdPlat]
		if !ok {
			return nil, fmt.Errorf("could not find Plattform")
		}

		return &struct {
			Body struct{ Plattform ds.Plattform }
		}{Body: struct{ Plattform ds.Plattform }{Plattform: *platform}}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Details zu einer Plattform"
	})

	// hier nennt man eine Plattform um
	huma.Post(*api, "/station/{id}/plattform/{idPlat}/rename", func(ctx context.Context, i *struct {
		Id     int `path:"id" example:"1"`
		IdPlat int `path:"idPlat" example:"1"`
		Body   struct {
			Name string `example:"Fred"`
		}
	}) (*ds.GenericResponse, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		platform, ok := station.Plattforms[i.IdPlat]
		if !ok {
			return nil, fmt.Errorf("could not find Plattform")
		}
		err := platform.Rename(i.Body.Name)
		if err != nil {
			return nil, fmt.Errorf("could not rename Plattform; %s", err.Error())
		}

		return ds.CreateGenericResponse("renamed Plattform"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Plattform einer station umbennenen"
	})

	// hier löscht man eine Plattform
	huma.Post(*api, "/station/{id}/plattform/{idPlat}/delete", func(ctx context.Context, i *struct {
		Id     int `path:"id" example:"1"`
		IdPlat int `path:"idPlat" example:"1"`
	}) (*ds.GenericResponse, error) {
		station, ok := gs.Stations[i.Id]
		if !ok {
			return nil, fmt.Errorf("Station does not exist")
		}
		platform, ok := station.Plattforms[i.IdPlat]
		if !ok {
			return nil, fmt.Errorf("could not find Plattform")
		}
		err := station.RemovePlattform(platform.Id, gs)
		if err != nil {
			return nil, fmt.Errorf("could not remove Plattform; %s", err.Error())
		}

		return ds.CreateGenericResponse("removed Plattform"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Plattform entfernen"
	})

	// hier kann man eine Station bauen
	huma.Post(*api, "/station", func(ctx context.Context, i *struct {
		Body struct {
			Position    *[2]int `example:"[1,1]" minLength:"2" maxLength:"2"`
			Position_to *[2]int `example:"[1,2]" required:"false" minLength:"2" maxLength:"2" nullable:"true"`
		}
	}) (*ds.GenericResponse, error) {
		var err error
		if i.Body.Position_to != nil {
			// mehrere bauen
			err = gs.AddStationTiles(*i.Body.Position, *i.Body.Position_to)
		} else {
			// Der hier gibt eigenlich noch die station mit
			_, err = gs.AddStationTile(*i.Body.Position)
		}
		if err != nil {
			return nil, fmt.Errorf("could not create statino; %s", err.Error())
		}

		return ds.CreateGenericResponse("build station"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"station"}
		o.Summary = "Station bauen"
	})

}

// endregion station
