package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	db "RelayServer/database"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var globalCoordinator = NewCoordinator()

type MutexConn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func (s *MutexConn) WriteJSON(v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ws.WriteJSON(v)
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

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	rawWS, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	mutexWS := &MutexConn{ws: rawWS}
	var registeredUserID string

	defer func() {
		rawWS.Close()
		if registeredUserID != "" {
			mu.Lock()
			delete(clients, registeredUserID)
			mu.Unlock()
		}
	}()

	for {
		_, msg, err := rawWS.ReadMessage()
		if err != nil {
			break
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			continue
		}

		authKey := fmt.Sprintf("%v", data["authKey"])

		user, err := db.AuthValidation(authKey)
		if err != nil || user == nil {

			continue
		}

		userID := fmt.Sprintf("%v", user["userID"])
		if userID == "<nil>" || userID == "" {
			continue
		}

		msgType, _ := data["message"].(string)
		userID, ok := data["userID"].(string)
		if !ok || userID == "" {
			continue
		}

		switch msgType {
		case "register":
			registeredUserID = userID
			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: mutexWS,
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

		case "joinCall":
			go func(d map[string]any) {
				callID := fmt.Sprintf("%v", d["callID"])

				globalCoordinator.AddUserToCall(userID, callID, mutexWS.ws)
			}(data)

		case "leaveCall":
			go func(d map[string]any) {
				callID := fmt.Sprintf("%v", d["callID"])
				globalCoordinator.RemoveUserFromCall(userID, callID)
			}(data)

		case "offer":
			go func(d map[string]any) {
				callID := fmt.Sprintf("%v", d["callID"])

				offerMap, ok := d["offer"].(map[string]any)
				if !ok {
					return
				}
				sdpStr, _ := offerMap["sdp"].(string)

				globalCoordinator.mutex.RLock()
				call, ok := globalCoordinator.Sessions[callID]
				globalCoordinator.mutex.RUnlock()

				if ok {
					if peer, ok := call.GetPeer(userID); ok {
						answer, err := peer.ReactOnOffer(sdpStr)
						if err != nil {
							return
						}
						call.SendAnswer(answer, userID)
					}
				}
			}(data)

		case "answer":
			go func(d map[string]any) {
				callID := fmt.Sprintf("%v", d["callID"])

				answerMap, ok := d["answer"].(map[string]any)
				if !ok {
					return
				}
				sdpStr, _ := answerMap["sdp"].(string)

				globalCoordinator.mutex.RLock()
				call, ok := globalCoordinator.Sessions[callID]
				globalCoordinator.mutex.RUnlock()

				if ok {
					if peer, ok := call.GetPeer(userID); ok {
						if err := peer.ReactOnAnswer(sdpStr); err != nil {
							log.Printf("ReactOnAnswer Error: %v", err)
						}
					}
				}
			}(data)

		case "candidate", "ice-candidate":
			go func(d map[string]any) {
				callID := fmt.Sprintf("%v", d["callID"])

				candMap, ok := d["candidate"].(map[string]any)
				if !ok {
					return
				}

				candStr, _ := candMap["candidate"].(string)
				var sdpMid *string
				if v, ok := candMap["sdpMid"].(string); ok {
					sdpMid = &v
				}

				var sdpMLineIndex *uint16
				if v, ok := candMap["sdpMLineIndex"].(float64); ok {
					idx := uint16(v)
					sdpMLineIndex = &idx
				}

				init := webrtc.ICECandidateInit{
					Candidate:     candStr,
					SDPMid:        sdpMid,
					SDPMLineIndex: sdpMLineIndex,
				}

				globalCoordinator.mutex.RLock()
				call, ok := globalCoordinator.Sessions[callID]
				globalCoordinator.mutex.RUnlock()

				if ok {
					if peer, ok := call.GetPeer(userID); ok {
						if err := peer.connection.AddICECandidate(init); err != nil {

						}
					}
				}
			}(data)
		}
	}
}

func SendWebsocketMessage(mutexWS *MutexConn, msg WebsocketMessage) error {
	return mutexWS.WriteJSON(msg)
}
