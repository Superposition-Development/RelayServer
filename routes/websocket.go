package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID   string
	Conn *websocket.Conn
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var (
	clients = make(map[string]*Client)
	mu      sync.RWMutex
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	var data map[string]string
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		json.Unmarshal(msg, &data)
		switch data["message"] {
		case "register":
			user, err := db.AuthValidation(data["authKey"])
			if err != nil {
				//do something
			}
			var rawUserID interface{} = user["userID"]
			userID := fmt.Sprintf("%v", rawUserID)
			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: ws,
			}
			mu.Unlock()
		case "sendMessage":
			SendMessage(data["serverID"], data["channelID"], data["content"], data["authKey"])
		}
	}

}

func SendWebsocketMessage(ws *websocket.Conn, msg WebsocketMessage) {
	ws.WriteJSON(msg)
}
