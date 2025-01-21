package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/garder500/safestore/database"
	"github.com/garder500/safestore/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	r := mux.NewRouter()
	manager, err := utils.NewManager()
	if err != nil {
		panic(err)
	}
	upgrader := websocket.Upgrader{}

	r.HandleFunc("/realtime", func(w http.ResponseWriter, r *http.Request) {
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

	r.PathPrefix("/database/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := r.URL.Path[10:]
		if r.Method == http.MethodGet {
			rows := make([]*database.SafeRow, 0)
			var err error

			if path == "" {
				err = manager.DB.Find(&rows).Error
			} else {
				err = database.StartWith(strings.ReplaceAll(path, "/", "."), manager.DB).Find(&rows).Error
			}
			if err != nil {
				jsonData, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
				http.Error(w, string(jsonData), http.StatusInternalServerError)
				return
			}

			data, err := database.FormatChildrenRecursive(rows, path)
			if err != nil {
				jsonData, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
				http.Error(w, string(jsonData), http.StatusInternalServerError)
				return
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				jsonData, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
				http.Error(w, string(jsonData), http.StatusInternalServerError)
				return
			}

			w.Write(jsonData)
		} else if r.Method == http.MethodPost {
			// tell that this method is not yet handled
			http.Error(w, "Method not yet implemented", http.StatusNotImplemented)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
