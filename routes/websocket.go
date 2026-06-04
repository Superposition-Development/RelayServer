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
		user, err := db.AuthValidation(data["authKey"])
		if err != nil {
			//do something
		}
		var rawUserID interface{} = user["userID"]
		userID := fmt.Sprintf("%v", rawUserID)
		switch data["message"] {
		case "register":
			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: ws,
			}
			mu.Unlock()
		case "sendMessage":
			messageID, err := SendMessage(data["serverID"], data["channelID"], data["content"], data["authKey"])
			users, err := GetServerUsers(data["serverID"])
			queryMap := map[string]string{
				"userID": userID,
			}
			senderData, err := db.QueryRow([]string{"pfp", "username"}, "user", queryMap)
			if err != nil {

			}
			for _, value := range users {
				client, ok := clients[value]
				if !ok {
					fmt.Printf("user %s not connected\n", value)
					continue
				}
				SendWebsocketMessage(client.Conn, WebsocketMessage{
					Type: "recieveMessage",
					Data: map[string]any{
						"id":        messageID,
						"name":      senderData["username"],
						"pfp":       senderData["pfp"],
						"content":   data["content"],
						"timestamp": data["timestamp"],
					},
				})
			}

		}
	}

}

func SendWebsocketMessage(ws *websocket.Conn, msg WebsocketMessage) {
	ws.WriteJSON(msg)
}
