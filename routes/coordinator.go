package routes

import (
	// "RelayServer/routes"
	"fmt"
	"log"

	// "RelayServer/routes"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Lobby interface {
	CreateCall(id string)
	RemoveCall(id string)
	AddUserToCall(userID string, callID string, socket *websocket.Conn)
	RemoveUserFromCall(userID string, callID string, socket *websocket.Conn)
	ShowSessions()
	ObtainEvent(message WebsocketMessage, socket *websocket.Conn)
}

type Coordinator struct {
	sessions map[string]*Call
}

func NewCoordinator() *Coordinator {
	return &Coordinator{sessions: map[string]*Call{}}
}

func (coordinator *Coordinator) ShowSessions() map[string]*Call {
	return coordinator.sessions
}

func (coordinator *Coordinator) CreateCall(id string) {
	coordinator.sessions[id] = NewCall(id)
}

func (coordinator *Coordinator) RemoveCall(id string) {
	delete(coordinator.sessions, id)
}

func (coordinator *Coordinator) AddUserToCall(userID string, callID string, socket *websocket.Conn) {
	if _, ok := coordinator.sessions[callID]; !ok {
		fmt.Println("New call was created: ", callID)
		coordinator.CreateCall(callID)
	}
	if call, ok := coordinator.sessions[callID]; ok {
		call.AddPeer(newPeer(userID))
		fmt.Println("Peer ", userID, "was added to call ", callID)
		if peer, ok := call.peers[userID]; ok {
			peer.SetSocket(socket)

			conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
			if err != nil {
				fmt.Println("Failed to establish peer connection")
			}

			peer.SetPeerConnection(conn)
			fmt.Println("Peer connection was established")
			for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
				if _, err := peer.connection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionRecvonly,
				}); err != nil {
					log.Print(err)
					return
				}
			}

			peer.connection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
				switch p {
				case webrtc.PeerConnectionStateFailed:
					if err := peer.connection.Close(); err != nil {
						log.Print(err)
					}
				case webrtc.PeerConnectionStateClosed:
					call.Signal()
				default:
				}
			})

			peer.connection.OnICECandidate(func(i *webrtc.ICECandidate) {
				if i == nil {
					fmt.Println("ICEGatheringState: connected")
					return
				}
				fmt.Println("Ice: ", i)
				call.SendICE(i, userID)
			})

			peer.connection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
				fmt.Println("Track added from peer: ", userID)
				defer call.Signal()
				trackLocal := call.AddTrack(t)
				defer call.RemoveTrack(trackLocal)
				defer fmt.Println("Track", trackLocal, "was removed")
				buf := make([]byte, 1500)
				for {
					i, _, err := t.Read(buf)
					if err != nil {
						return
					}

					if _, err = trackLocal.Write(buf[:i]); err != nil {
						return
					}
				}
			})
		}

	}
}

func (coordinator *Coordinator) RemoveUserFromCall(userID string, callID string) {
	if call, ok := coordinator.sessions[callID]; ok {
		delete(call.peers, userID)
	}
}

func (coordinator *Coordinator) ObtainEvent(message WebsocketMessage, socket *MutexConn, userID string) {
	fmt.Println("Message: " + message.Type)
	switch message.Type {

	case "register":
		// userMu.Lock()
		// registeredUserID = userID
		// userMu.Unlock()

		// mu.Lock()
		clients[userID] = &Client{
			ID:   userID,
			Conn: socket,
		}
		// mu.Unlock()

		log.Printf("[WS] Registered websocket user: %s", userID)

	case "sendMessageServer":
		m, ok := message.Data.(map[string]any)
		if ok {
			// serverID := parseString(m["serverID"])
			// channelID := parseString(m["channelID"])
			// content := parseString(m["content"])
			// authKey := parseString(m["authKey"])
			// SendMessage(serverID, channelID, content, authKey)
			handleSendMessage(m, userID, parseString(m["authKey"]))
		}

	case "sendMessageDM":
		m, ok := message.Data.(map[string]any)
		if ok {
			// serverID := parseString(m["serverID"])
			// channelID := parseString(m["channelID"])
			// content := parseString(m["content"])
			// authKey := parseString(m["authKey"])
			// SendMessage(serverID, channelID, content, authKey)
			handleSendMessage(m, userID, parseString(m["authKey"]))
		}

	// case "joinCall":
	// 	go func() {
	// 		m, ok := message.Data.(map[string]any)
	// 		if ok {
	// 			callID := m["callID"].(string)
	// 			coordinator.AddUserToCall(userID, callID, socket.ws)
	// 		}
	// 	}()
	case "leaveCall":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				callID := m["callID"].(string)
				coordinator.RemoveUserFromCall(userID, callID)
			}
		}()
	case "joinCall":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				callID := m["callID"].(string)
				coordinator.AddUserToCall(userID, callID, socket.ws)
				fmt.Println(m)
				offer2 := m["offer"].(map[string]any)
				if call, ok := coordinator.sessions[callID]; ok {
					if peer, ok := call.peers[userID]; ok {
						answer, err2 := peer.ReactOnOffer(offer2["sdp"].(string))
						if err2 != nil {
							fmt.Println(err2)
							return
						}
						call.SendAnswer(answer, userID)
					}
				}
			}
		}()
	case "answer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				callID, _ := m["callID"].(string)
				offer2 := m["answer"].(map[string]any)
				if call, ok := coordinator.sessions[callID]; ok {
					if peer, ok := call.peers[userID]; ok {
						err := peer.ReactOnAnswer(offer2["sdp"].(string))
						if err != nil {
							fmt.Println(err)
							return
						}
					}

				}
			}
		}()
	case "candidate":
		go func() {
			//m, ok := message.Data.(CANDIDATE)
			m, ok := message.Data.(map[string]any)
			if ok {
				callID, _ := m["callID"].(string)
				candidate := m["candidate"].(map[string]any)
				i_candidate := candidate["candidate"].(string)
				sdp_mid := candidate["sdpMid"].(string)
				sdp_m_line_index := uint16(candidate["sdpMLineIndex"].(float64))
				var username_fragment string
				if candidate["usernameFragment"] != nil {
					username_fragment = candidate["usernameFragment"].(string)
				} else {
					username_fragment = ""
				}
				init := webrtc.ICECandidateInit{
					Candidate:        i_candidate,
					SDPMid:           &sdp_mid,
					SDPMLineIndex:    &sdp_m_line_index,
					UsernameFragment: &username_fragment,
				}
				if call, ok := coordinator.sessions[callID]; ok {
					if peer, ok := call.peers[userID]; ok {
						if err := peer.connection.AddICECandidate(init); err != nil {
							log.Println(err)
							return
						}
						fmt.Println("ICE-CANDIDATE added for peer", peer.id)
						fmt.Println(peer.connection.ICEConnectionState())
						fmt.Println(peer.connection.ICEGatheringState())
					}
				}
			} else {
				fmt.Println(m)
			}
		}()
	default:
		fmt.Println("DEFAULT")
		fmt.Println(message)

	}
}
