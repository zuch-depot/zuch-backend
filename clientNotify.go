package main

import "log/slog"

func startNotifiyingClientsOfChanges(users *[]*User, channel <-chan wsEnvelope) {
	for {
		event := <-channel
		for _, user := range *users {
			logger.Debug("Notifying client of event", slog.String("user", user.username), slog.String("Event Type", event.Type))
			user.connection.WriteJSON(event)
		}
	}
}
