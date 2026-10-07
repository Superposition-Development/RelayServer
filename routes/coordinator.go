package routes

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Lobby interface {
	CreateCall(id string)
	RemoveCall(id string)
	AddUserToCall(userID string, callID string, socket *websocket.Conn)
	RemoveUserFromCall(userID string, callID string)
	ObtainEvent(message WebsocketMessage, socket *MutexConn, userID string)
}

//	var config = webrtc.Configuration{
//		ICEServers: []webrtc.ICEServer{
//			{
//				URLs: []string{
//					"[2001:4860:4864:5:8000::1]:19302",
//				},
//			},
//		},
//	}
var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{
				"turn:global.relay.metered.ca:80",
				"turn:global.relay.metered.ca:443",
				"turn:global.relay.metered.ca:443?transport=tcp",
			},
			Username:   "c1bea89d980d944a146c66a3",
			Credential: "RbBZdljmQEoTFBC+",
		},
	},
	ICETransportPolicy: webrtc.ICETransportPolicyRelay,
}

// var config = webrtc.Configuration{
// ICEServers: []webrtc.ICEServer{
// 	{URLs: []string{"stun:stun.l.google.com:19302"}},
// },
// ICETransportPolicy: webrtc.ICETransportPolicyRelay,
// }

type Coordinator struct {
	sessions map[string]*Call
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		sessions: make(map[string]*Call),
	}
}

func (coordinator *Coordinator) CreateCall(id string) {
	coordinator.sessions[id] = NewCall(id)
}

func (coordinator *Coordinator) RemoveCall(id string) {
	delete(coordinator.sessions, id)
}

func (coordinator *Coordinator) AddUserToCall(userID string, callID string, socket *websocket.Conn) {
	call, ok := coordinator.sessions[callID]

	if !ok {
		fmt.Println("New call was created:", callID)
		coordinator.CreateCall(callID)
		call = coordinator.sessions[callID]
	}

	_, exists := call.peers[userID]
	if exists {
		fmt.Println("Peer already exists:", userID)
		return
	}

	peer := newPeer(userID)
	peer.SetSocket(socket)

	conn, err := webrtc.NewPeerConnection(config)

	if err != nil {
		fmt.Println("Failed to create PeerConnection:", err)
		return
	}

	peer.SetPeerConnection(conn)
	call.AddPeer(peer)

	fmt.Println("Peer ", userID, " was added to call ", callID)

	conn.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

		fmt.Println("connection detected a track:", userID, track.ID(), track.StreamID(), track.Kind())
		trackLocal := call.AddTrack(track)
		call.Signal()

		defer func() {
			fmt.Println("track end:", track.ID())
			call.RemoveTrack(trackLocal)
			call.Signal()
		}()

		buf := make([]byte, 1500)
		for {
			n, _, err := track.Read(buf)

			if err != nil {
				fmt.Println("Track read error: ", track.ID(), err)
				return
			}

			_, err = trackLocal.Write(buf[:n])

			if err != nil {
				fmt.Println("Track write error:", track.ID(), err)
				return
			}
		}
	})

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			fmt.Println("ice gathering complete:", userID)
			return
		}
		call.SendICE(candidate, userID)
	})

	conn.OnConnectionStateChange(
		func(state webrtc.PeerConnectionState) {
			fmt.Println("Peer", userID, "connection state:", state)

			switch state {
			case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
				go func() {
					fmt.Println("Removing peer: ", userID)
					peer.connection.Close()
					coordinator.RemoveUserFromCall(userID, callID)
				}()
			}
		},
	)

	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeAudio, webrtc.RTPCodecTypeVideo} {
		_, err := conn.AddTransceiverFromKind(kind, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})

		if err != nil {
			fmt.Println("add transciever error:", err)
			return
		}
	}
}

func (coordinator *Coordinator) RemoveUserFromCall(userID string, callID string) {
	call, ok := coordinator.sessions[callID]

	if !ok {
		return
	}

	peer, ok := call.peers[userID]

	if !ok {
		return
	}

	if peer.connection != nil {
		peer.connection.Close()
	}

	call.RemovePeer(userID)

	fmt.Println("Removed ", userID, " from call ", callID)

	if len(call.peers) == 0 {
		coordinator.RemoveCall(callID)
		fmt.Println("Removed empty call:", callID)
	}
}

func (coordinator *Coordinator) ObtainEvent(message WebsocketMessage, socket *MutexConn, userID string) {
	fmt.Println("Message:", message.Type)
	switch message.Type {
	case "register":
		clients[userID] = &Client{
			ID:   userID,
			Conn: socket,
		}
		log.Printf("[WS] Registered websocket user: %s", userID)

	case "joinCall":
		go func() {
			m, ok := message.Data.(map[string]any)
			if !ok {
				fmt.Println("Invalid joinCall data")
				return
			}

			callID, ok := m["callID"].(string)
			if !ok {
				fmt.Println("Invalid callID")
				return
			}

			offer, ok := m["offer"].(map[string]any)
			if !ok {
				fmt.Println("Invalid offer")
				return
			}

			sdp, ok := offer["sdp"].(string)
			if !ok {
				fmt.Println("Invalid SDP")
				return
			}

			coordinator.AddUserToCall(userID, callID, socket.ws)

			call, ok := coordinator.sessions[callID]
			if !ok {
				return
			}

			peer, ok := call.peers[userID]
			if !ok {
				return
			}

			answer, err := peer.ReactOnOffer(sdp)

			if err != nil {
				fmt.Println("ReactOnOffer error:", err)
				return
			}

			fmt.Println("Sending initial answer to", userID)

			call.SendAnswer(answer, userID)
		}()

	case "answer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if !ok {
				return
			}

			callID, ok := m["callID"].(string)
			if !ok {
				return
			}

			answerData, ok := m["answer"].(map[string]any)
			if !ok {
				return
			}

			sdp, ok := answerData["sdp"].(string)
			if !ok {
				return
			}

			call, ok := coordinator.sessions[callID]
			if !ok {
				return
			}

			peer, ok := call.peers[userID]
			if !ok {
				return
			}

			if peer.connection.SignalingState() != webrtc.SignalingStateHaveLocalOffer {
				fmt.Println("Discarding answer from", userID, " because signaling state is ", peer.connection.SignalingState())
				return
			}

			err := peer.ReactOnAnswer(sdp)

			if err != nil {
				fmt.Println("ReactOnAnswer error:", err)
				return
			}

			fmt.Println("Answer accepted from ", userID, " new state: ", peer.connection.SignalingState())
		}()

	case "candidate":
		go func() {
			m, ok := message.Data.(map[string]any)
			if !ok {
				return
			}

			callID, ok := m["callID"].(string)
			if !ok {
				return
			}

			candidate, ok := m["candidate"].(map[string]any)
			if !ok {
				return
			}

			candidateString, ok := candidate["candidate"].(string)
			if !ok {
				return
			}

			sdpMid, ok := candidate["sdpMid"].(string)
			if !ok {
				return
			}

			lineIndex, ok := candidate["sdpMLineIndex"].(float64)
			if !ok {
				return
			}

			lineIndexUint16 := uint16(lineIndex)

			usernameFragment := ""

			value, exists := candidate["usernameFragment"]

			if exists && value != nil {
				usernameFragment = value.(string)
			}

			init := webrtc.ICECandidateInit{
				Candidate:        candidateString,
				SDPMid:           &sdpMid,
				SDPMLineIndex:    &lineIndexUint16,
				UsernameFragment: &usernameFragment,
			}

			call, ok := coordinator.sessions[callID]
			if !ok {
				return
			}

			peer, ok := call.peers[userID]
			if !ok {
				return
			}

			err := peer.connection.AddICECandidate(init)

			if err != nil {
				fmt.Println("add ice candiate error:", err)
				return
			}

			fmt.Println("ICE candidate added:", userID)
		}()

	case "leaveCall":
		go func() {
			m, ok := message.Data.(map[string]any)
			if !ok {
				return
			}
			callID, ok := m["callID"].(string)
			if !ok {
				return
			}
			coordinator.RemoveUserFromCall(userID, callID)
		}()

	case "sendMessageServer":
		m, ok := message.Data.(map[string]any)
		if !ok {
			return
		}
		handleSendMessage(m, userID, parseString(m["authKey"]))

	case "sendMessageDM":
		m, ok := message.Data.(map[string]any)
		if !ok {
			return
		}
		handleSendMessage(m, userID, parseString(m["authKey"]))

	default:
		fmt.Println("Unknown message:", message)
	}
}
