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
		jsonOp := utils.WebSocketQuery{}
		for {
			// read json message
			err := c.ReadJSON(&jsonOp)
			if err != nil {
				log.Println("read:", err)
				manager.WebsocketManager.RemoveClient(userID)
				break
			}

			switch jsonOp.Op {
			case utils.AuthOp: // Authentication operation
				authPayload := jsonOp.Data.(utils.AuthPayload)

				if authPayload.Token != "" && authPayload.Token == "supersecret" {
					c.WriteJSON(utils.WebSocketQuery{Op: 0, Data: "Authorized"})
				} else {
					c.WriteJSON(utils.WebSocketQuery{Op: 0, Data: "Unauthorized"})
					manager.WebsocketManager.RemoveClient(userID)
					c.Close()
					break
				}
			case utils.InsertOp: // Insert operation in the database
				crudPayload := jsonOp.Data.(utils.CrudPayload)
				var paths []map[string]interface{}
				utils.GeneratePaths(crudPayload.Data, crudPayload.Path, &paths)
				err = database.InsertInSafeRow(manager.DB, &paths)
				if err != nil {
					log.Println(err)
					break
				}
				jsonData, err := json.Marshal(utils.WebSocketQuery{Op: 1, Data: crudPayload})
				if err != nil {
					log.Println(err)
					break
				}
				manager.WebsocketManager.Broadcast(jsonData, userID)
			}
			err = c.WriteJSON(jsonOp)
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
		port = "4789"
	}
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
