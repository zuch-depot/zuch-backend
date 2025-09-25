package main

func startNotifiyingClientsOfChanges(users *[]*User, channel <-chan wsEnvelope) {
	for _, user := range *users {
		if user.isConnected {
			startNotifiyingClientsOfChanges(users, channel)

		}
	}
}
