package routes

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Peer struct {
	id         string
	connection *webrtc.PeerConnection
	mutex      sync.RWMutex
	socket     *websocket.Conn
	wsMutex    sync.Mutex
}

func newPeer(id string) *Peer {
	log.Printf("[Peer] Initializing new peer struct for peerID: %s", id)
	return &Peer{
		id:    id,
		mutex: sync.RWMutex{},
	}
}

func (p *Peer) WriteJSON(v interface{}) error {
	p.wsMutex.Lock()
	defer p.wsMutex.Unlock()

	if p.socket == nil {
		log.Printf("[Peer %s] Error: Attempted WriteJSON on nil WebSocket connection", p.id)
	}

	log.Printf("[Peer %s] Sending WebSocket message: %v", p.id, v)
	err := p.socket.WriteJSON(v)
	if err != nil {
		log.Printf("[Peer %s] Error writing JSON to WebSocket: %v", p.id, err)
	}
	return err
}

func (peer *Peer) SetSocket(socket *websocket.Conn) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	log.Printf("[Peer %s] Setting WebSocket connection", peer.id)
	peer.socket = socket
}

func (peer *Peer) SetPeerConnection(conn *webrtc.PeerConnection) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	log.Printf("[Peer %s] Setting PeerConnection instance", peer.id)
	peer.connection = conn
}

func (peer *Peer) ReactOnOffer(offerStr string) (webrtc.SessionDescription, error) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	log.Printf("[Peer %s] Reacting to received Offer", peer.id)

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerStr,
	}

	if err := peer.connection.SetRemoteDescription(offer); err != nil {
		log.Printf("[Peer %s] Error setting Remote Description (Offer): %v", peer.id, err)
		return webrtc.SessionDescription{}, err
	}
	log.Printf("[Peer %s] Successfully set Remote Description (Offer)", peer.id)

	answer, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		log.Printf("[Peer %s] Error creating Answer: %v", peer.id, err)
		return webrtc.SessionDescription{}, err
	}
	log.Printf("[Peer %s] Successfully created Answer", peer.id)

	if err := peer.connection.SetLocalDescription(answer); err != nil {
		log.Printf("[Peer %s] Error setting Local Description (Answer): %v", peer.id, err)
		return webrtc.SessionDescription{}, err
	}
	log.Printf("[Peer %s] Successfully set Local Description (Answer)", peer.id)

	return answer, nil
}

func (peer *Peer) ReactOnAnswer(answerStr string) error {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	log.Printf("[Peer %s] Reacting to received Answer", peer.id)

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerStr,
	}
	err := peer.connection.SetRemoteDescription(answer)
	if err != nil {
		log.Printf("[Peer %s] Error setting Remote Description (Answer): %v", peer.id, err)
		return err
	}
	log.Printf("[Peer %s] Successfully set Remote Description (Answer)", peer.id)
	return nil
}

type Call struct {
	id     string
	mutex  sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
}

func NewCall(id string) *Call {
	log.Printf("[Call %s] Initializing new Call session", id)
	return &Call{
		id:     id,
		peers:  make(map[string]*Peer),
		tracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

func (call *Call) GetPeer(peerID string) (*Peer, bool) {
	call.mutex.RLock()
	defer call.mutex.RUnlock()
	peer, ok := call.peers[peerID]
	log.Printf("[Call %s] GetPeer call for peerID: %s (Found: %t)", call.id, peerID, ok)
	return peer, ok
}

func (call *Call) AddPeer(peer *Peer) {
	call.mutex.Lock()
	call.peers[peer.id] = peer
	call.mutex.Unlock()
	log.Printf("[Call %s] Added peer: %s. Total peers: %d", call.id, peer.id, len(call.peers))
}

func (call *Call) RemovePeer(peerID string) {
	call.mutex.Lock()
	delete(call.peers, peerID)
	log.Printf("[Call %s] Removed peer: %s. Remaining peers: %d", call.id, peerID, len(call.peers))
	call.mutex.Unlock()

	log.Printf("[Call %s] Triggering Signal after removing peer: %s", call.id, peerID)
	call.Signal()
}

func (coordinator *Coordinator) RemoveUserFromCall(userID string, callID string) {
	log.Printf("[Coordinator] Request to remove user: %s from call: %s", userID, callID)
	if call, ok := coordinator.Sessions[callID]; ok {
		delete(call.peers, userID)
		log.Printf("[Coordinator] User %s deleted from call %s peers map", userID, callID)
	} else {
		log.Printf("[Coordinator] Call session %s not found for user removal", callID)
	}
}

func (call *Call) SendAnswer(message webrtc.SessionDescription, peerID string) {
	log.Printf("[Call %s] Attempting to send Answer to peer: %s", call.id, peerID)
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] Failed to send Answer: Peer %s not found", call.id, peerID)
		return
	}

	msg := WebsocketMessage{
		Type: "answer",
		Data: message,
	}

	if err := peer.WriteJSON(msg); err != nil {
		log.Printf("[Call %s] Error sending Answer message to peer %s: %v", call.id, peerID, err)
	} else {
		log.Printf("[Call %s] Answer successfully sent to peer %s", call.id, peerID)
	}
}

func (call *Call) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	log.Printf("[Call %s] Adding remote track ID: %s, StreamID: %s, Kind: %s", call.id, track.ID(), track.StreamID(), track.Kind().String())

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
	if err != nil {
		log.Printf("[Call %s] Error creating local track for track ID %s: %v", call.id, track.ID(), err)
		return nil
	}

	call.mutex.Lock()
	call.tracks[track.ID()] = trackLocal
	call.mutex.Unlock()

	log.Printf("[Call %s] Local track created and added. Total active call tracks: %d", call.id, len(call.tracks))
	call.Signal()
	return trackLocal
}

func (call *Call) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	if track == nil {
		log.Printf("[Call %s] Attempted to remove nil track", call.id)
		return
	}

	log.Printf("[Call %s] Removing track ID: %s", call.id, track.ID())
	call.mutex.Lock()
	delete(call.tracks, track.ID())
	call.mutex.Unlock()

	log.Printf("[Call %s] Track %s removed. Remaining tracks: %d", call.id, track.ID(), len(call.tracks))
	call.Signal()
}

func (call *Call) SendICE(message *webrtc.ICECandidate, peerID string) {
	log.Printf("[Call %s] Preparing to send ICE candidate to peer: %s", call.id, peerID)
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] SendICE failed: Peer %s not found", call.id, peerID)
		return
	}

	raw, err := json.Marshal(message.ToJSON())
	if err != nil {
		log.Printf("[Call %s] Error marshaling ICE candidate for peer %s: %v", call.id, peerID, err)
		return
	}

	if err := peer.WriteJSON(WebsocketMessage{Type: "candidate", Data: string(raw)}); err != nil {
		log.Printf("[Call %s] Error sending ICE candidate to peer %s: %v", call.id, peerID, err)
	} else {
		log.Printf("[Call %s] ICE candidate sent to peer %s", call.id, peerID)
	}
}

func (call *Call) Signal() {
	call.mutex.Lock()
	defer call.mutex.Unlock()

	log.Printf("[Call %s] Starting Signal renegotiation loop across %d peers", call.id, len(call.peers))

	attemptSync := func() bool {
		for _, peer := range call.peers {
			if peer.connection == nil {
				log.Printf("[Call %s] Signal loop: Peer %s has nil PeerConnection, skipping", call.id, peer.id)
				continue
			}

			if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				log.Printf("[Call %s] Peer %s connection state is Closed. Removing peer.", call.id, peer.id)
				delete(call.peers, peer.id)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range peer.connection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				if _, ok := call.tracks[sender.Track().ID()]; !ok {
					log.Printf("[Call %s] Track %s no longer in call. Removing sender from peer %s", call.id, sender.Track().ID(), peer.id)
					if err := peer.connection.RemoveTrack(sender); err == nil {
						log.Printf("[Call %s] Successfully removed sender track from peer %s. Requesting sync repeat.", call.id, peer.id)
						return true
					} else {
						log.Printf("[Call %s] Error removing sender track from peer %s: %v", call.id, peer.id, err)
					}
				}
			}

			for _, receiver := range peer.connection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}
				existingSenders[receiver.Track().ID()] = true
			}

			for trackID, track := range call.tracks {
				if _, ok := existingSenders[trackID]; !ok {
					log.Printf("[Call %s] Adding track %s to peer %s connection", call.id, trackID, peer.id)
					if _, err := peer.connection.AddTrack(track); err == nil {
						log.Printf("[Call %s] Successfully added track to peer %s. Requesting sync repeat.", call.id, peer.id)
						return true
					} else {
						log.Printf("[Call %s] Error adding track %s to peer %s: %v", call.id, trackID, peer.id, err)
					}
				}
			}

			if peer.connection.SignalingState() == webrtc.SignalingStateStable {
				log.Printf("[Call %s] Peer %s signaling state is Stable. Creating new SDP Offer...", call.id, peer.id)
				offer, err := peer.connection.CreateOffer(nil)
				if err != nil {
					log.Printf("[Call %s] Error creating SDP offer for peer %s: %v", call.id, peer.id, err)
					return false
				}

				if err = peer.connection.SetLocalDescription(offer); err != nil {
					log.Printf("[Call %s] Error setting local description for peer %s: %v", call.id, peer.id, err)
					return false
				}

				offerString, err := json.Marshal(offer)
				if err != nil {
					log.Printf("[Call %s] Error marshaling SDP offer for peer %s: %v", call.id, peer.id, err)
					return false
				}

				log.Printf("[Call %s] Sending SDP Offer via WebSocket to peer %s", call.id, peer.id)
				if err = peer.WriteJSON(WebsocketMessage{
					Type: "offer",
					Data: string(offerString),
				}); err != nil {
					log.Printf("[Call %s] Error writing SDP Offer to peer %s: %v", call.id, peer.id, err)
					return false
				}
			}
		}
		return false
	}

	for syncAttempt := 0; syncAttempt < 25; syncAttempt++ {
		log.Printf("[Call %s] Signal sync loop iteration: %d", call.id, syncAttempt+1)
		if !attemptSync() {
			log.Printf("[Call %s] Signal sync loop stabilized on iteration %d", call.id, syncAttempt+1)
			break
		}
	}
}

type Coordinator struct {
	mutex    sync.RWMutex
	Sessions map[string]*Call
}

func NewCoordinator() *Coordinator {
	log.Println("[Coordinator] Initializing new Coordinator")
	return &Coordinator{Sessions: make(map[string]*Call)}
}

func (coordinator *Coordinator) AddUserToCall(userID string, callID string, socket *websocket.Conn) {
	log.Printf("[Coordinator] Adding user %s to call %s", userID, callID)

	coordinator.mutex.Lock()
	call, ok := coordinator.Sessions[callID]
	if !ok {
		log.Printf("[Coordinator] Call %s does not exist. Creating new call session.", callID)
		call = NewCall(callID)
		coordinator.Sessions[callID] = call
	}
	coordinator.mutex.Unlock()

	peer := newPeer(userID)
	peer.SetSocket(socket)

	log.Printf("[Coordinator] Creating WebRTC PeerConnection for user %s", userID)
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					// "stun:stun.l.google.com:19302",
				},
			},
			{
				URLs: []string{
					"turn:global.relay.metered.ca:80",
					"turn:global.relay.metered.ca:443",
					"turn:global.relay.metered.ca:443?transport=tcp",
				},
				Username:   "a072cb146b471d7876e641dc",
				Credential: "AbV/kjuHbgOurcxl",
			},
		},
	}

	conn, err := webrtc.NewPeerConnection(config)
	// conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("[Coordinator] Failed to create PeerConnection for user %s: %v", userID, err)
		return
	}
	peer.SetPeerConnection(conn)

	call.AddPeer(peer)

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		log.Printf("[Coordinator] Adding recvonly transceiver (%s) for user %s", typ.String(), userID)
		if _, err := conn.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Printf("[Coordinator] Failed to add transceiver (%s) for user %s: %v", typ.String(), userID, err)
			return
		}
	}

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[PeerConnection %s] Connection state changed: %s", userID, state.String())
		switch state {
		case webrtc.PeerConnectionStateFailed:
			log.Printf("[PeerConnection %s] Connection failed. Closing PeerConnection.", userID)
			_ = conn.Close()
		case webrtc.PeerConnectionStateClosed:
			log.Printf("[PeerConnection %s] Connection closed. Removing user from call %s", userID, callID)
			call.RemovePeer(userID)
		}
	})

	conn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice != nil {
			log.Printf("[PeerConnection %s] Local ICE candidate generated: %s", userID, ice)
			call.SendICE(ice, userID)
		} else {
			log.Printf("[PeerConnection %s] ICE candidate gathering finished (nil candidate)", userID)
		}
	})

	conn.OnTrack(func(trackRemote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Printf("[PeerConnection %s] Received remote track (ID: %s, Kind: %s, StreamID: %s)", userID, trackRemote.ID(), trackRemote.Kind().String(), trackRemote.StreamID())

		trackLocal := call.AddTrack(trackRemote)
		if trackLocal == nil {
			log.Printf("[PeerConnection %s] Failed to create local track for remote track %s", userID, trackRemote.ID())
			return
		}
		defer func() {
			log.Printf("[PeerConnection %s] Deferred execution: Removing track %s from call %s", userID, trackLocal.ID(), callID)
			call.RemoveTrack(trackLocal)
		}()

		log.Printf("[PeerConnection %s] Starting RTP copy loop for track %s", userID, trackRemote.ID())
		buf := make([]byte, 1500)
		for {
			i, _, err := trackRemote.Read(buf)
			if err != nil {
				log.Printf("[PeerConnection %s] Error reading from track %s (loop ending): %v", userID, trackRemote.ID(), err)
				return
			}
			if _, err = trackLocal.Write(buf[:i]); err != nil {
				log.Printf("[PeerConnection %s] Error writing to local track %s (loop ending): %v", userID, trackRemote.ID(), err)
				return
			}
		}
	})

	log.Printf("[Coordinator] Initializing Signal phase for user %s joining call %s", userID, callID)
	call.Signal()
}
