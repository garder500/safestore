package utils

import (
	"fmt"
)

type mapListener struct {
	// Map of channels to listeners
	listeners map[string]chan string
}

func (ml *mapListener) Notify(channel, payload string) error {
	// Check if channel exists
	if _, ok := ml.listeners[channel]; !ok {
		return fmt.Errorf("channel %s does not exist", channel)
	}

	fmt.Println("Sending payload to channel", channel)
	// Send payload to channel
	ml.listeners[channel] <- payload
	return nil
}

func (ml *mapListener) listenForNextPayload(channel string) (string, error) {
	// Check if channel exists
	if _, ok := ml.listeners[channel]; !ok {
		return "", fmt.Errorf("channel %s does not exist", channel)
	}

	// Wait for payload
	payload := <-ml.listeners[channel]
	fmt.Println("Received payload from channel", payload)
	return payload, nil
}

func newMapListener() *mapListener {
	return &mapListener{
		listeners: make(map[string]chan string),
	}
}

func (ml *mapListener) addChannel(channel string) {
	ml.listeners[channel] = make(chan string)
}

func (ml *mapListener) RemoveChannel(channel string) {
	delete(ml.listeners, channel)
}

func (ml *mapListener) Close() {
	for _, ch := range ml.listeners {
		close(ch)
	}
}
