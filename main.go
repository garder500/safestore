package main

import "github.com/garder500/safestore/utils"

func main() {
	manager, err := utils.NewManager()
	if err != nil {
		panic(err)
	}

	// Add a channel to the listener
	manager.Notify("channel1", "payload1")

	for {
		// Wait for the next payload on the channel
		err := manager.Listen("channel1")
		if err != nil {
			panic(err)
		}
		payload, err := manager.ListenForNextPayload("channel1")
		if err != nil {
			panic(err)
		}
		// Print the payload
		println(payload)
	}

}
