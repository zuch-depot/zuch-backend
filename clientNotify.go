package main

import "zuch-backend/internal/ds"

func startNotifiyingClientsOfChanges(users *[]*ds.User, channel <-chan ds.WsEnvelope) {
	for _, user := range *users {
		if user.IsConnected {
			startNotifiyingClientsOfChanges(users, channel)

		}
	}
}
