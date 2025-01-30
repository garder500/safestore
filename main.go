package main

import (
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

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
	outer:
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
					break outer
				}
			case utils.InsertOp: // Insert operation in the database
				crudPayload := jsonOp.Data.(utils.CrudPayload)
				var paths []map[string]interface{}
				utils.GeneratePaths(crudPayload.Data, crudPayload.Path, &paths)
				err = database.InsertInSafeRow(manager.DB, &paths)
				if err != nil {
					log.Println(err)
					c.WriteJSON(utils.WebSocketQuery{Op: 0, Data: err.Error()})
				}
				manager.WebsocketManager.Broadcast(jsonOp)
			case utils.DeleteOp: // Delete operation in the database
				crudPayload := jsonOp.Data.(utils.CrudPayload)
				err := database.DeleteInSafeRow(manager.DB, &crudPayload.Path)
				if err != nil {
					c.WriteJSON(utils.WebSocketQuery{Op: 0, Data: err.Error()})
				} else {
					manager.WebsocketManager.Broadcast(jsonOp)
				}
			case utils.GetOp:
				crudPayload := jsonOp.Data.(utils.CrudPayload)
				rows := make([]*database.SafeRow, 0)
				path := crudPayload.Path
				var err error

				if path == "" {
					err = manager.DB.Find(&rows).Error
				} else {
					err = database.StartWith(strings.ReplaceAll(path, "/", "."), manager.DB).Find(&rows).Error
				}
				if err != nil {
					c.WriteJSON(utils.WebSocketQuery{
						Op: 500,
						Data: map[string]interface{}{
							"error": err.Error(),
						},
					})
				}

				data, err := database.FormatChildrenRecursive(rows, path)
				if err != nil {
					c.WriteJSON(utils.WebSocketQuery{
						Op: 500,
						Data: map[string]interface{}{
							"error": err.Error(),
						},
					})
				}

				manager.WebsocketManager.Broadcast(utils.WebSocketQuery{
					Op:   200,
					Data: data,
				})

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
		if r.Method == http.MethodGet {
			controllers.GetController(w, r, manager.DB)
			return
		} else if r.Method == http.MethodPost {
			controllers.PostController(w, r, manager.DB)
			return
		}
		utils.FormatHttpError(w, http.StatusNotImplemented, "Not implemented", "This endpoint is not implemented yet")
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "4789"
	}
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
