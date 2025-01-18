package main

import (
	"log"
	"net/http"

	"github.com/garder500/safestore/utils"
	"github.com/gorilla/websocket"
)

func main() {
	_, err := utils.NewManager()
	if err != nil {
		panic(err)
	}

	upgrader := websocket.Upgrader{}

	http.HandleFunc("/realtime", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
