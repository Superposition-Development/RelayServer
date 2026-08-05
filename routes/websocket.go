package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

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

		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {

			break
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {

			continue
		}

		authKey, _ := data["authKey"].(string)
		msgType, _ := data["message"].(string)

		user, err := db.AuthValidation(authKey)
		if err != nil || user == nil {

			continue
		}

		userID := fmt.Sprintf("%v", user["userID"])
		if userID == "<nil>" || userID == "" {

			continue
		}

		switch msgType {
		case "register":
			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: ws,
			}
			mu.Unlock()

		case "sendMessage":
			serverID := fmt.Sprintf("%v", data["serverID"])
			channelID := fmt.Sprintf("%v", data["channelID"])
			content := fmt.Sprintf("%v", data["content"])

			messageID, err := SendMessage(serverID, channelID, content, authKey)
			if err != nil {

			}

			users, err := GetServerUsers(serverID)
			if err != nil {

				continue
			}

			queryMap := map[string]string{"userID": userID}
			senderData, err := db.QueryRow([]string{"pfp", "username"}, "user", queryMap)
			if err != nil {

			}

			mu.RLock()
			for _, value := range users {
				client, ok := clients[value]
				if !ok {

					continue
				}

				SendWebsocketMessage(client.Conn, WebsocketMessage{
					Type: "recieveMessage",
					Data: map[string]any{
						"id":        messageID,
						"serverID":  serverID,
						"channelID": channelID,
						"name":      senderData["username"],
						"pfp":       senderData["pfp"],
						"content":   content,
						"timestamp": time.Now().Unix(),
					},
				})
			}
			mu.RUnlock()

		default:

		}
	}
}

func SendWebsocketMessage(ws *websocket.Conn, msg WebsocketMessage) {
	ws.WriteJSON(msg)
}
