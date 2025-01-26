package utils

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
)

type WebsocketManager struct {
	clients map[string]*websocket.Conn
}

type OpEnum int // OpEnum is an enum for the websocket operations
const (
	AuthOp OpEnum = iota
	InsertOp
	DeleteOp
	UpdateOp
	GetOp
)

type WebSocketQuery struct {
	Op   OpEnum      `json:"op"`
	Data interface{} `json:"data"`
}

type AuthPayload struct {
	Token         string `json:"token"`
	Authorization string `json:"authorization"`
}

type CrudPayload struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

func NewWebsocketManager() *WebsocketManager {
	return &WebsocketManager{
		clients: make(map[string]*websocket.Conn),
	}
}

func (wm *WebsocketManager) AddClient(userID string, conn *websocket.Conn) {
	wm.clients[userID] = conn
}

func (wm *WebsocketManager) RemoveClient(userID string) {
	delete(wm.clients, userID)
}

func (wm *WebsocketManager) Broadcast(message interface{}, exclude ...string) {
	for userID, conn := range wm.clients {
		for _, ex := range exclude {
			if userID == ex {
				continue
			} else {
				err := conn.WriteJSON(message)
				if err != nil {
					log.Printf("error writing message to %s: %s", userID, err)
				}
			}
		}
	}
}

func (wm *WebsocketManager) SendToUser(userID string, message interface{}) error {
	conn, ok := wm.clients[userID]
	if !ok {
		return errors.New("user not found")
	}
	return conn.WriteJSON(message)
}

func (wm *WebsocketManager) SendToMultipleUsers(userIDs []string, message interface{}) {
	// send to existing clients only
	for _, userID := range userIDs {
		conn, ok := wm.clients[userID]
		if !ok {
			continue
		}
		err := conn.WriteJSON(message)
		if err != nil {
			log.Printf("error writing message to %s: %s", userID, err)
		}
	}
}
