package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var globalCoordinator = NewCoordinator()

type MutexConn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func (m *MutexConn) WriteJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ws == nil {
		return fmt.Errorf("nil websocket")
	}

	return m.ws.WriteJSON(v)
}

type Client struct {
	ID   string
	Conn *MutexConn
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var (
	clients = make(map[string]*Client)
	mu      sync.RWMutex
)

func parseString(val interface{}) string {
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", val)
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	rawWS, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	mutexWS := &MutexConn{ws: rawWS}

	var (
		registeredUserID string
		userMu           sync.Mutex
	)

	defer func() {
		_ = rawWS.Close()

		userMu.Lock()
		uid := registeredUserID
		userMu.Unlock()

		if uid != "" {
			mu.Lock()
			delete(clients, uid)
			mu.Unlock()

			log.Printf("[WS] Client disconnected: %s", uid)
		}
	}()

	log.Printf("[WS] New WebSocket connection from %s", r.RemoteAddr)

	for {
		_, msg, err := rawWS.ReadMessage()
		if err != nil {
			userMu.Lock()
			uid := registeredUserID
			userMu.Unlock()

			log.Printf("[WS %s] Read error: %v", uid, err)
			break
		}

		userMu.Lock()
		currentUID := registeredUserID
		userMu.Unlock()

		log.Printf("[WS %s] Recieved: %s", currentUID, string(msg))

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Printf("[WS] Failed to unmarshal JSON: %v", err)
			continue
		}

		authKey, _ := data["authKey"].(string)
		msgType, _ := data["message"].(string)

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

		switch msgType {
		case "register":
			userMu.Lock()
			registeredUserID = userID
			userMu.Unlock()

			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: mutexWS,
			}
			mu.Unlock()

			log.Printf("[WS] Registered websocket user: %s", userID)

		case "sendMessage":
			handleSendMessage(data, userID, authKey)

		case "joinCall":
			callID := parseString(data["callID"])
			log.Printf("[WS %s] Joining call %s", userID, callID)

			go globalCoordinator.AddUserToCall(userID, callID, rawWS)

		case "leaveCall":
			callID := parseString(data["callID"])
			log.Printf("[WS %s] Leaving call %s", userID, callID)

			go globalCoordinator.RemoveUserFromCall(userID, callID)

		case "offer":
			handleOffer(data, userID)

		case "answer":
			handleAnswer(data, userID)

		case "candidate", "ice-candidate":
			handleCandidate(data, userID)

		default:
			log.Printf("[WS %s] Unknown Message Type: %s", userID, msgType)
		}
	}
}

func handleOffer(data map[string]any, userID string) {
	callID := parseString(data["callID"])

	offerMap, ok := data["offer"].(map[string]any)
	if !ok {
		log.Printf("[WS %s] Invalid offer payload", userID)
		return
	}

	sdpStr, _ := offerMap["sdp"].(string)
	if sdpStr == "" {
		log.Printf("[WS %s] Empty offer SDP", userID)
		return
	}

	globalCoordinator.mutex.RLock()
	call, ok := globalCoordinator.Sessions[callID]
	globalCoordinator.mutex.RUnlock()

	if !ok {
		log.Printf("[WS %s] Call %s not found", userID, callID)
		return
	}

	peer, ok := call.GetPeer(userID)
	if !ok {
		log.Printf("[WS %s] Peer not in call %s", userID, callID)
		return
	}

	answer, err := peer.ReactOnOffer(sdpStr)
	if err != nil {
		log.Printf("[WS %s] ReactOnOffer failed: %v", userID, err)
		return
	}

	call.SendAnswer(answer, userID)
}

func handleAnswer(data map[string]any, userID string) {
	callID := parseString(data["callID"])

	answerMap, ok := data["answer"].(map[string]any)
	if !ok {
		log.Printf("[WS %s] Invalid answer payload", userID)
		return
	}

	sdpStr, _ := answerMap["sdp"].(string)
	if sdpStr == "" {
		log.Printf("[WS %s] Empty answer SDP", userID)
		return
	}

	globalCoordinator.mutex.RLock()
	call, ok := globalCoordinator.Sessions[callID]
	globalCoordinator.mutex.RUnlock()

	if !ok {
		log.Printf("[WS %s] Call %s not found", userID, callID)
		return
	}

	peer, ok := call.GetPeer(userID)
	if !ok {
		log.Printf("[WS %s] Peer not found", userID)
		return
	}

	if err := peer.ReactOnAnswer(sdpStr); err != nil {
		log.Printf("[WS %s] ReactOnAnswer failed: %v", userID, err)
		return
	}

	log.Printf("[WS %s] Answer success", userID)
}

func handleCandidate(data map[string]any, userID string) {
	callID := parseString(data["callID"])

	candidateMap, ok := data["candidate"].(map[string]any)
	if !ok {
		log.Printf("[WS %s] Invalid candidate payload: %T", userID, data["candidate"])
		return
	}

	candidateJSON, err := json.Marshal(candidateMap)
	if err != nil {
		log.Printf("[WS %s] Failed to marshal candidate: %v", userID, err)
		return
	}

	var init webrtc.ICECandidateInit
	if err := json.Unmarshal(candidateJSON, &init); err != nil {
		log.Printf("[WS %s] Failed to decode candidate: %v", userID, err)
		return
	}

	log.Printf("[WS %s] Received ICE candidate: %s", userID, init.Candidate)

	globalCoordinator.mutex.RLock()
	call, ok := globalCoordinator.Sessions[callID]
	globalCoordinator.mutex.RUnlock()

	if !ok {
		log.Printf("[WS %s] Call %s not found", userID, callID)
		return
	}

	peer, ok := call.GetPeer(userID)
	if !ok {
		log.Printf("[WS %s] Peer not found", userID)
		return
	}

	if err := peer.AddRemoteCandidate(init); err != nil {
		log.Printf("[WS %s] AddRemoteCandidate failed: %v", userID, err)
		return
	}

	log.Printf("[WS %s] ICE candidate success", userID)
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

	mu.RLock()
	defer mu.RUnlock()

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
