package routes

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
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

type Coordinator struct {
	mutex    sync.RWMutex
	Sessions map[string]*Call
}
type Peer struct {
	id string

	stateMutex           sync.RWMutex
	connection           *webrtc.PeerConnection
	remoteDescriptionSet bool
	pendingCandidates    []webrtc.ICECandidateInit

	wsMutex sync.Mutex
	socket  *websocket.Conn
}

type Call struct {
	id     string
	mutex  sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
}

func newPeer(id string) *Peer {
	log.Printf("[Peer] Init peer: %s", id)
	return &Peer{
		id:                id,
		pendingCandidates: make([]webrtc.ICECandidateInit, 0),
	}
}

func (p *Peer) WriteJSON(v interface{}) error {
	p.wsMutex.Lock()
	defer p.wsMutex.Unlock()

	if p.socket == nil {
		err := fmt.Errorf("nil websocket connection")
		log.Printf("[Peer %s] %v", p.id, err)
		return err
	}

	log.Printf("[Peer %s] Sending WebSocket message: %+v", p.id, v)
	if err := p.socket.WriteJSON(v); err != nil {
		log.Printf("[Peer %s] Error writing JSON to WebSocket: %v", p.id, err)
		return err
	}

	return nil
}

func (p *Peer) SetSocket(socket *websocket.Conn) {
	p.wsMutex.Lock()
	defer p.wsMutex.Unlock()

	log.Printf("[Peer %s] Setting WebSocket connection", p.id)
	p.socket = socket
}

func (p *Peer) SetPeerConnection(conn *webrtc.PeerConnection) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	log.Printf("[Peer %s] Setting PeerConnection", p.id)
	p.connection = conn
}

func (p *Peer) GetPeerConnection() *webrtc.PeerConnection {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()

	return p.connection
}

func (p *Peer) ReactOnOffer(offerStr string) (webrtc.SessionDescription, error) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.connection == nil {
		return webrtc.SessionDescription{}, fmt.Errorf("PeerConnection nil for %s", p.id)
	}

	log.Printf("[Peer %s] Reacting to received Offer", p.id)

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerStr,
	}

	if err := p.connection.SetRemoteDescription(offer); err != nil {
		log.Printf("[Peer %s] Error setting Remote Description (Offer): %v", p.id, err)
		return webrtc.SessionDescription{}, err
	}

	p.remoteDescriptionSet = true
	p.ClearPendingICECandidatesLocked()

	answer, err := p.connection.CreateAnswer(nil)
	if err != nil {
		log.Printf("[Peer %s] Error creating Answer: %v", p.id, err)
		return webrtc.SessionDescription{}, err
	}

	if err := p.connection.SetLocalDescription(answer); err != nil {
		log.Printf("[Peer %s] Error setting Local Description (Answer): %v", p.id, err)
		return webrtc.SessionDescription{}, err
	}

	log.Printf("[Peer %s] Processed Offer and created Answer", p.id)
	return answer, nil
}

func (p *Peer) ReactOnAnswer(answerStr string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.connection == nil {
		return fmt.Errorf("PeerConnection nil for %s", p.id)
	}

	log.Printf("[Peer %s] Reacting to Answer", p.id)

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerStr,
	}

	if err := p.connection.SetRemoteDescription(answer); err != nil {
		log.Printf("[Peer %s] Error setting Remote Description (Answer): %v", p.id, err)
		return err
	}

	p.remoteDescriptionSet = true
	p.ClearPendingICECandidatesLocked()

	log.Printf("[Peer %s] Set Remote Description (Answer)", p.id)
	return nil
}

func (p *Peer) AddRemoteCandidate(candidate webrtc.ICECandidateInit) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.connection == nil {
		return fmt.Errorf("PeerConnection nil for %s", p.id)
	}

	if !p.remoteDescriptionSet {
		log.Printf("[Peer %s] ICE queued, remote desc unset", p.id)
		p.pendingCandidates = append(p.pendingCandidates, candidate)
		return nil
	}

	log.Printf("[Peer %s] Adding remote ICE candidate", p.id)
	return p.connection.AddICECandidate(candidate)
}

func (p *Peer) ClearPendingICECandidatesLocked() {
	if len(p.pendingCandidates) == 0 {
		return
	}

	log.Printf("[Peer %s] Clearing %d queued ICE candidates", p.id, len(p.pendingCandidates))
	for _, candidate := range p.pendingCandidates {
		if err := p.connection.AddICECandidate(candidate); err != nil {
			log.Printf("[Peer %s] Error adding queued ICE candidate: %v", p.id, err)
		}
	}
	p.pendingCandidates = nil
}

func NewCall(id string) *Call {
	log.Printf("[Call %s] Init new Call", id)
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
	return peer, ok
}

func (call *Call) AddPeer(peer *Peer) {
	call.mutex.Lock()
	call.peers[peer.id] = peer
	total := len(call.peers)
	call.mutex.Unlock()

	log.Printf("[Call %s] Added peer: %s. Total peers: %d", call.id, peer.id, total)
}

func (call *Call) RemovePeer(peerID string) {
	call.mutex.Lock()
	delete(call.peers, peerID)
	remaining := len(call.peers)
	call.mutex.Unlock()

	log.Printf("[Call %s] Removed peer: %s. Total peers: %d", call.id, peerID, remaining)
	call.Signal()
}

func (call *Call) SendAnswer(message webrtc.SessionDescription, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] Failed to send Answer: Peer %s not found", call.id, peerID)
		return
	}

	msg := WebsocketMessage{
		Type: "answer",
		Data: map[string]interface{}{
			"type": message.Type.String(),
			"sdp":  message.SDP,
		},
	}

	if err := peer.WriteJSON(msg); err != nil {
		log.Printf("[Call %s] Error sending Answer to peer %s: %v", call.id, peerID, err)
		return
	}

	log.Printf("[Call %s] Answer sent to peer %s", call.id, peerID)
}

func (call *Call) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	log.Printf("[Call %s] Adding remote track ID: %s, StreamID: %s, Kind: %s",
		call.id, track.ID(), track.StreamID(), track.Kind().String())

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(
		track.Codec().RTPCodecCapability,
		track.ID(),
		track.StreamID(),
	)

	if err != nil {
		log.Printf("[Call %s] Error creating local track: %v", call.id, err)
		return nil
	}

	call.mutex.Lock()
	call.tracks[track.ID()] = trackLocal
	total := len(call.tracks)
	call.mutex.Unlock()

	log.Printf("[Call %s] Local track added. Total active tracks: %d", call.id, total)
	call.Signal()

	return trackLocal
}

func (call *Call) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	if track == nil {
		return
	}

	call.mutex.Lock()
	delete(call.tracks, track.ID())
	remaining := len(call.tracks)
	call.mutex.Unlock()

	log.Printf("[Call %s] Track %s removed. Total tracks: %d", call.id, track.ID(), remaining)
	call.Signal()
}

func (call *Call) SendICE(message *webrtc.ICECandidate, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] SendICE failed: Peer %s not found", call.id, peerID)
		return
	}

	candidate := message.ToJSON()
	err := peer.WriteJSON(WebsocketMessage{
		Type: "candidate",
		Data: candidate,
	})

	if err != nil {
		log.Printf("[Call %s] Error sending ICE to %s: %v", call.id, peerID, err)
	}
}

func (call *Call) Signal() {
	call.mutex.RLock()
	peers := make([]*Peer, 0, len(call.peers))
	for _, peer := range call.peers {
		peers = append(peers, peer)
	}

	tracks := make([]*webrtc.TrackLocalStaticRTP, 0, len(call.tracks))
	for _, track := range call.tracks {
		tracks = append(tracks, track)
	}
	call.mutex.RUnlock()

	log.Printf("[Call %s] Signaling across %d peers for %d tracks", call.id, len(peers), len(tracks))

	for _, peer := range peers {
		conn := peer.GetPeerConnection()
		if conn == nil || conn.ConnectionState() == webrtc.PeerConnectionStateClosed {
			continue
		}

		existingTracks := make(map[string]bool)
		for _, sender := range conn.GetSenders() {
			if sender.Track() != nil {
				existingTracks[sender.Track().ID()] = true
			}
		}

		for _, track := range tracks {
			if existingTracks[track.ID()] {
				continue
			}

			log.Printf("[Call %s] Adding track %s to peer %s", call.id, track.ID(), peer.id)
			if _, err := conn.AddTrack(track); err != nil {
				log.Printf("[Call %s] Error adding track %s to peer %s: %v", call.id, track.ID(), peer.id, err)
			}
		}

		if conn.SignalingState() != webrtc.SignalingStateStable {
			log.Printf("[Call %s] Peer %s signaling state is %s; deferring offer generation",
				call.id, peer.id, conn.SignalingState().String())
			continue
		}

		offer, err := conn.CreateOffer(nil)
		if err != nil {
			log.Printf("[Call %s] Error creating offer for %s: %v", call.id, peer.id, err)
			continue
		}

		if err := conn.SetLocalDescription(offer); err != nil {
			log.Printf("[Call %s] Error setting local description for %s: %v", call.id, peer.id, err)
			continue
		}

		err = peer.WriteJSON(WebsocketMessage{
			Type: "offer",
			Data: map[string]interface{}{
				"type": offer.Type.String(),
				"sdp":  offer.SDP,
			},
		})

		if err != nil {
			log.Printf("[Call %s] Failed to send offer to %s: %v", call.id, peer.id, err)
		}
	}
}

func NewCoordinator() *Coordinator {
	log.Println("[Coordinator] Init Coordinator")
	return &Coordinator{
		Sessions: make(map[string]*Call),
	}
}

func (coordinator *Coordinator) RemoveUserFromCall(userID, callID string) {
	log.Printf("[Coordinator] Attempting to remove user: %s from call: %s", userID, callID)

	coordinator.mutex.RLock()
	call, ok := coordinator.Sessions[callID]
	coordinator.mutex.RUnlock()

	if !ok {
		log.Printf("[Coordinator] Call session %s not found (removing user)", callID)
		return
	}

	call.RemovePeer(userID)
}

func (coordinator *Coordinator) AddUserToCall(userID, callID string, socket *websocket.Conn) {
	log.Printf("[Coordinator] Adding user %s to call %s", userID, callID)

	coordinator.mutex.Lock()
	call, ok := coordinator.Sessions[callID]
	if !ok {
		log.Printf("[Coordinator] Creating new call session: %s", callID)
		call = NewCall(callID)
		coordinator.Sessions[callID] = call
	}
	coordinator.mutex.Unlock()

	peer := newPeer(userID)
	peer.SetSocket(socket)

	conn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Printf("[Coordinator] Failed to create PeerConnection for %s: %v", userID, err)
		return
	}

	peer.SetPeerConnection(conn)
	call.AddPeer(peer)

	if _, err := conn.AddTransceiverFromKind(
		webrtc.RTPCodecTypeAudio,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly},
	); err != nil {
		log.Printf("[Coordinator] Failed to add audio transceiver for %s: %v", userID, err)
	}

	conn.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[PeerConnection %s] ICE state: %s", userID, state.String())
	})

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[PeerConnection %s] Connection state: %s", userID, state.String())
		switch state {
		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
			call.RemovePeer(userID)
		}
	})

	conn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice == nil {
			log.Printf("[PeerConnection %s] ICE gathering finished", userID)
			return
		}
		call.SendICE(ice, userID)
	})

	conn.OnTrack(func(trackRemote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("[PeerConnection %s] Received remote track ID=%s Kind=%s StreamID=%s",
			userID, trackRemote.ID(), trackRemote.Kind().String(), trackRemote.StreamID())

		trackLocal := call.AddTrack(trackRemote)
		if trackLocal == nil {
			return
		}

		defer call.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			n, _, err := trackRemote.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[PeerConnection %s] Track read ended with error: %v", userID, err)
				}
				return
			}

			if _, err := trackLocal.Write(buf[:n]); err != nil {
				if err != io.EOF {
					log.Printf("[PeerConnection %s] Track write failed: %v", userID, err)
				}
				return
			}
		}
	})

	log.Printf("[Coordinator] Init Signal phase for %s", userID)
	call.Signal()
}
