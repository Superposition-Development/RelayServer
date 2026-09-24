package routes

import (
	// "RelayServer/SFU"
	// "RelayServer/SFU"
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"time"

	"github.com/gorilla/websocket"
)

var (
	clients = make(map[string]*Client)
)

type MutexConn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

type Client struct {
	ID   string
	Conn *MutexConn
}

func (m *MutexConn) WriteJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ws == nil {
		return fmt.Errorf("nil websocket")
	}

	return m.ws.WriteJSON(v)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func parseString(val interface{}) string {
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", val)
}

type WebsocketMessage struct {
	Type string      `json:"message"`
	Data interface{} `json:"data"`
}

type wsServer struct {
	clients     map[*websocket.Conn]bool
	coordinator Coordinator
}

func StartServer() *wsServer {
	server := wsServer{
		make(map[*websocket.Conn]bool),
		*NewCoordinator(),
	}

	return &server
}

func (ws *wsServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	mutexWS := &MutexConn{ws: conn}

	defer conn.Close()

	fmt.Printf("Client connected")

	if err != nil {
		fmt.Printf(" with error %s", err)
		return
	}

	fmt.Println(" successfully")

	message := WebsocketMessage{}

	for {
		messageType, bmessage, err := conn.ReadMessage()

		if err != nil {
			fmt.Println(err)
			return
		}
		if messageType == websocket.CloseMessage {
			break
		}

		var data map[string]any
		if err := json.Unmarshal(bmessage, &data); err != nil {
			log.Printf("[WS] Failed to unmarshal JSON: %v", err)
			continue
		}

		authKey, _ := data["authKey"].(string)

		user, err := db.AuthValidation(authKey)
		if err != nil || user == nil {
			log.Printf("[WS] Auth failed")
			continue
		}

		userID := parseString(user["userID"])
		if userID == "" || userID == "<nil>" {
			log.Printf("[WS] Invalid userID")
			continue
		}

		if err != nil {
			fmt.Println("DROP")
			fmt.Println(message.Data)
			fmt.Println(err)
			return
		}
		message.Data = data
		message.Type = data["message"].(string)
		ws.coordinator.ObtainEvent(message, mutexWS, userID)
	}
}

func handleSendMessage(data map[string]any, userID, authKey string) {
	serverID := parseString(data["serverID"])
	channelID := parseString(data["channelID"])
	content := parseString(data["content"])

	messageID, err := SendMessage(serverID, channelID, content, authKey)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	users, err := GetServerUsers(serverID)
	if err != nil {
		return
	}

	senderData, err := db.QueryRow(
		[]string{"pfp", "username"},
		"user",
		map[string]string{"userID": userID},
	)
	if err != nil {
		log.Printf("Error fetching sender details: %v", err)
	}

	outboundMsg := WebsocketMessage{
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
	}

	// mu.RLock()
	// defer mu.RUnlock()

	for _, targetID := range users {
		client, ok := clients[targetID]
		if !ok {
			continue
		}

		if err := SendWebsocketMessage(client.Conn, outboundMsg); err != nil {
			log.Printf("Failed to send message to  %s: %v", targetID, err)
		}
	}
}

func SendWebsocketMessage(mutexWS *MutexConn, msg WebsocketMessage) error {
	log.Printf("[WS] Sending Message: type=%s data=%+v", msg.Type, msg.Data)
	return mutexWS.WriteJSON(msg)
}
