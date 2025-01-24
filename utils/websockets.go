package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	Token string `json:"token"`
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

func (wm *WebsocketManager) Broadcast(message []byte, exclude ...string) {
	for userID, conn := range wm.clients {
		for _, ex := range exclude {
			if userID == ex {
				continue
			} else {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("error writing message:", err)
				}
			}
		}
	}
}

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (wm *WebsocketManager) SendToUser(userID string, message []byte) error {
	conn, ok := wm.clients[userID]
	if !ok {
		return errors.New("user not found")
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}

func (wm *WebsocketManager) SendToMultipleUsers(userIDs []string, message []byte) {
	// send to existing clients only
	for _, userID := range userIDs {
		conn, ok := wm.clients[userID]
		if !ok {
			continue
		}
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("error writing message:", err)
		}
	}
}
