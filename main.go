package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/garder500/safestore/database"
	"github.com/garder500/safestore/utils"
	"github.com/gorilla/websocket"
)

func main() {
	manager, err := utils.NewManager()
	if err != nil {
		panic(err)
	}

	upgrader := websocket.Upgrader{}
	rows := &database.SafeRow{}
	err = database.StartWith("posts", manager.DB).First(rows).Error
	if err != nil {
		log.Println(err)
		return
	}

	formattedChildren, err := rows.FormatChildren(manager.DB)
	if err != nil {
		log.Println(err)
		return
	}
	data, err := rows.ToJson()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(data))
	jsonData, err := json.Marshal(formattedChildren)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(jsonData))

	http.HandleFunc("/realtime", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		userID, err := utils.GenerateRandomString()
		if err != nil {
			log.Println(err)
			return
		}

		manager.WebsocketManager.AddClient(userID, c)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				manager.WebsocketManager.RemoveClient(userID)
				break
			}
			log.Printf("recv: %s by %s", message, userID)
			// search if the first word is broadcast
			if string(message[:9]) == "broadcast" {
				manager.WebsocketManager.Broadcast(message[10:], userID)
				// we need to continue to avoid sending the message to the client
				continue
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
