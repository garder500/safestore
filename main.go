package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"safestore/controllers"
	"safestore/database"
	"safestore/utils"

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
			controllers.PostSafeRow(w, r, manager.DB, &path)
		} else if r.Method == http.MethodDelete {
			controllers.DeleteSafeRow(w, r, manager.DB, &path)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3478"
	}
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
