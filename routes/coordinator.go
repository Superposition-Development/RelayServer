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
	CreateRoom(id string)
	RemoveRoom(id string)
	AddUserToRoom(self_id string, room_id string, socket *websocket.Conn)
	RemoveUserFromRoom(self_id string, room_id string, socket *websocket.Conn)
	ShowSessions()
	ObtainEvent(message WebsocketMessage, socket *websocket.Conn)
}

type Coordinator struct {
	sessioins map[string]*Room
}

func NewCoordinator() *Coordinator {
	return &Coordinator{sessioins: map[string]*Room{}}
}

func (coordinator *Coordinator) ShowSessions() map[string]*Room {
	return coordinator.sessioins
}

func (coordinator *Coordinator) CreateRoom(id string) {
	coordinator.sessioins[id] = NewRoom(id)
}

func (coordinator *Coordinator) RemoveRoom(id string) {
	delete(coordinator.sessioins, id)
}

func (coordinator *Coordinator) AddUserToRoom(self_id string, room_id string, socket *websocket.Conn) {
	if _, ok := coordinator.sessioins[room_id]; !ok {
		fmt.Println("New Room was created: ", room_id)
		coordinator.CreateRoom(room_id)
	}
	if room, ok := coordinator.sessioins[room_id]; ok {
		room.AddPeer(newPeer(self_id))
		fmt.Println("Peer ", self_id, "was added to room ", room_id)
		if peer, ok := room.peers[self_id]; ok {
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
					room.Signal()
				default:
				}
			})

			peer.connection.OnICECandidate(func(i *webrtc.ICECandidate) {
				if i == nil {
					fmt.Println("ICEGatheringState: connected")
					return
				}
				fmt.Println("Ice: ", i)
				room.SendICE(i, self_id)
			})

			peer.connection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
				fmt.Println("Track added from peer: ", self_id)
				defer room.Signal()
				trackLocal := room.AddTrack(t)
				defer room.RemoveTrack(trackLocal)
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

func (coordinator *Coordinator) RemoveUserFromRoom(self_id string, room_id string) {
	if room, ok := coordinator.sessioins[room_id]; ok {
		delete(room.peers, self_id)
	}
}

func (coordinator *Coordinator) ObtainEvent(message WebsocketMessage, socket *MutexConn, userID string) {
	fmt.Println("im talkin bout a " + message.Type)
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

	case "sendMessage":
		m, ok := message.Data.(map[string]any)
		if ok {
			// serverID := parseString(m["serverID"])
			// channelID := parseString(m["channelID"])
			// content := parseString(m["content"])
			// authKey := parseString(m["authKey"])
			// SendMessage(serverID, channelID, content, authKey)
			handleSendMessage(m, userID, parseString(m["authKey"]))
		}

	case "joinRoom":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id := m["self_id"].(string)
				room_id := m["room_id"].(string)
				coordinator.AddUserToRoom(self_id, room_id, socket.ws)
			}
		}()
	case "leaveRoom":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id := m["self_id"].(string)
				room_id := m["room_id"].(string)
				coordinator.RemoveUserFromRoom(self_id, room_id)
			}
		}()
	case "offer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
				offer2 := m["offer"].(map[string]any)
				if room, ok := coordinator.sessioins[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
						answer, err2 := peer.ReactOnOffer(offer2["sdp"].(string))
						if err2 != nil {
							fmt.Println(err2)
							return
						}
						room.SendAnswer(answer, self_id)
					}
				}
			}
		}()
	case "answer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
				offer2 := m["answer"].(map[string]any)
				if room, ok := coordinator.sessioins[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
						err := peer.ReactOnAnswer(offer2["sdp"].(string))
						if err != nil {
							fmt.Println(err)
							return
						}
					}

				}
			}
		}()
	case "ice-candidate":
		go func() {
			//m, ok := message.Data.(CANDIDATE)
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
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
				if room, ok := coordinator.sessioins[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
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
				fmt.Println("nach")
			}
		}()
	default:
		fmt.Println("DEFAULT")
		fmt.Println(message)

	}
}
