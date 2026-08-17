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
			URLs: []string{
				"turn:global.relay.metered.ca:80",
				"turn:global.relay.metered.ca:443",
				"turn:global.relay.metered.ca:443?transport=tcp",
			},
			Username:   "009867250f59e795a4229efe",
			Credential: "5zIER+2OTGZSow0c",
		},
	},
	ICETransportPolicy: webrtc.ICETransportPolicyRelay,
}

type Coordinator struct {
	mutex    sync.RWMutex
	Sessions map[string]*Call
}

func NewCoordinator() *Coordinator {
	log.Println("[Coordinator] Init Coordinator")
	return &Coordinator{
		Sessions: make(map[string]*Call),
	}
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
	id    string
	mutex sync.RWMutex

	peers  map[string]*Peer
	tracks map[string]*LocalTrack
}

type LocalTrack struct {
	Track *webrtc.TrackLocalStaticRTP
	Owner string
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
		log.Printf("[Peer %s] Error writing JSON: %v", p.id, err)
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
		log.Printf("[Peer %s] Error setting Remote Description: %v", p.id, err)
		return webrtc.SessionDescription{}, err
	}

	p.remoteDescriptionSet = true
	p.clearPendingICECandidatesLocked()

	answer, err := p.connection.CreateAnswer(nil)
	if err != nil {
		log.Printf("[Peer %s] Error creating Answer: %v", p.id, err)
		return webrtc.SessionDescription{}, err
	}

	if err := p.connection.SetLocalDescription(answer); err != nil {
		log.Printf("[Peer %s] Error setting Local Description: %v", p.id, err)
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
		log.Printf("[Peer %s] Error setting Remote Description: %v", p.id, err)
		return err
	}

	p.remoteDescriptionSet = true
	p.clearPendingICECandidatesLocked()

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
		log.Printf("[Peer %s] ICE queued; remote desc unset: %s", p.id, candidate.Candidate)
		p.pendingCandidates = append(p.pendingCandidates, candidate)
		return nil
	}

	log.Printf("[Peer %s] Adding remote ICE candidate: %s", p.id, candidate.Candidate)
	if err := p.connection.AddICECandidate(candidate); err != nil {
		log.Printf("[Peer %s] AddICECandidate failed: %v", p.id, err)
		return err
	}

	log.Printf("[Peer %s] ICE candidate added", p.id)
	return nil
}

func (p *Peer) clearPendingICECandidatesLocked() {
	if len(p.pendingCandidates) == 0 {
		return
	}

	log.Printf("[Peer %s] Clearing %d queued ICE candidates", p.id, len(p.pendingCandidates))

	for _, candidate := range p.pendingCandidates {
		if err := p.connection.AddICECandidate(candidate); err != nil {
			log.Printf("[Peer %s] Error adding queued ICE candidate: %v", p.id, err)
		} else {
			log.Printf("[Peer %s] queued ICE added : %s", p.id, candidate.Candidate)
		}
	}

	p.pendingCandidates = nil
}

func NewCall(id string) *Call {
	log.Printf("[Call %s] Init new Call", id)
	return &Call{
		id:     id,
		peers:  make(map[string]*Peer),
		tracks: make(map[string]*LocalTrack),
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
	peer, exists := call.peers[peerID]
	delete(call.peers, peerID)
	remaining := len(call.peers)
	call.mutex.Unlock()

	if exists && peer != nil {
		if conn := peer.GetPeerConnection(); conn != nil {
			_ = conn.Close()
		}
	}

	log.Printf("[Call %s] Removed peer: %s. Total peers: %d", call.id, peerID, remaining)
}

func (call *Call) SendAnswer(message webrtc.SessionDescription, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] Failed to send Answer: peer %s not found", call.id, peerID)
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
		log.Printf("[Call %s] Error sending Answer: %v", call.id, err)
		return
	}

	log.Printf("[Call %s] Answer sent to peer %s", call.id, peerID)
}

func (call *Call) AddTrack(track *webrtc.TrackRemote, owner string) *webrtc.TrackLocalStaticRTP {
	log.Printf("[Call %s] Adding remote track ID=%s StreamID=%s Kind=%s", call.id, track.ID(), track.StreamID(), track.Kind().String())

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
	call.tracks[track.ID()] = &LocalTrack{
		Track: trackLocal,
		Owner: owner,
	}
	total := len(call.tracks)
	call.mutex.Unlock()

	log.Printf("[Call %s] Local track added. Total active tracks: %d", call.id, total)
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
}

func (call *Call) SendICE(message *webrtc.ICECandidate, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		log.Printf("[Call %s] SendICE failed: peer %s not found", call.id, peerID)
		return
	}

	candidate := message.ToJSON()
	log.Printf("[Call %s] Sending ICE to %s: %s", call.id, peerID, candidate.Candidate)

	err := peer.WriteJSON(WebsocketMessage{
		Type: "candidate",
		Data: candidate,
	})
	if err != nil {
		log.Printf("[Call %s] Error sending ICE: %v", call.id, err)
	}
}

func (call *Call) Signal() {
	call.mutex.RLock()
	peers := make([]*Peer, 0, len(call.peers))
	for _, peer := range call.peers {
		peers = append(peers, peer)
	}

	tracks := make([]*LocalTrack, 0, len(call.tracks))
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

		for _, localTrack := range tracks {
			if localTrack == nil || localTrack.Track == nil {
				continue
			}
			if localTrack.Owner == peer.id {
				continue
			}

			track := localTrack.Track

			if existingTracks[track.ID()] {
				continue
			}

			log.Printf("[Call %s] Adding track %s to peer %s", call.id, track.ID(), peer.id)
			if _, err := conn.AddTrack(track); err != nil {
				log.Printf("[Call %s] Error adding track: %v", call.id, err)
			}
		}

		if conn.SignalingState() != webrtc.SignalingStateStable {
			log.Printf("[Call %s] Peer %s signaling state is %s; skipping offer", call.id, peer.id, conn.SignalingState().String())
			continue
		}

		offer, err := conn.CreateOffer(nil)
		if err != nil {
			log.Printf("[Call %s] CreateOffer failed: %v", call.id, err)
			continue
		}

		if err := conn.SetLocalDescription(offer); err != nil {
			log.Printf("[Call %s] SetLocalDescription failed: %v", call.id, err)
			continue
		}
		gatherComplete := webrtc.GatheringCompletePromise(conn)
		<-gatherComplete

		localDescription := conn.LocalDescription()
		if localDescription == nil {
			log.Printf("[Call %s] LocalDescription is nil", call.id)
			continue
		}

		log.Printf("[Call %s] ICE gathering complete; sending complete SDP offer", call.id)

		err = peer.WriteJSON(WebsocketMessage{
			Type: "offer",
			Data: map[string]interface{}{
				"type": localDescription.Type.String(),
				"sdp":  localDescription.SDP,
			},
		})
		if err != nil {
			log.Printf("[Call %s] Failed sending offer: %v", call.id, err)
		}
	}
}

func (coordinator *Coordinator) RemoveUserFromCall(userID, callID string) {
	log.Printf("[Coordinator] Removing user %s from call %s", userID, callID)

	coordinator.mutex.RLock()
	call, ok := coordinator.Sessions[callID]
	coordinator.mutex.RUnlock()

	if !ok {
		log.Printf("[Coordinator] Call %s not found", callID)
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
		log.Printf("[Coordinator] Failed to create PeerConnection: %v", err)
		return
	}

	peer.SetPeerConnection(conn)
	call.AddPeer(peer)

	_, err = conn.AddTransceiverFromKind(
		webrtc.RTPCodecTypeAudio,
		webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		},
	)
	if err != nil {
		log.Printf("[Coordinator] Failed to add audio transceiver: %v", err)
	}

	conn.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[PeerConnection %s] ICE state: %s", userID, state.String())
	})

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[PeerConnection %s] Connection state: %s", userID, state.String())
		switch state {
		case webrtc.PeerConnectionStateFailed:
			log.Printf("[PeerConnection %s] Connection Failed", userID)
		case webrtc.PeerConnectionStateConnected:
			log.Printf("[PeerConnection %s] Connection Success", userID)
		case webrtc.PeerConnectionStateDisconnected:
			log.Printf("[PeerConnection %s] disconnected", userID)
		}
	})

	conn.OnICEGatheringStateChange(func(state webrtc.ICEGathererState) {
		log.Printf("[PeerConnection %s] Ice gathering: %s", userID, state.String())
	})

	conn.OnTrack(func(trackRemote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("[PeerConnection %s] remote track: ID=%s Kind=%s StreamID=%s", userID, trackRemote.ID(), trackRemote.Kind().String(), trackRemote.StreamID())

		trackLocal := call.AddTrack(trackRemote, userID)
		if trackLocal == nil {
			return
		}

		log.Printf("[Call %s] New track %s; renegotiating", call.id, trackLocal.ID())

		go func() {
			call.Signal()
		}()

		defer call.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			n, _, readErr := trackRemote.Read(buf)
			if readErr != nil {
				if readErr != io.EOF {
					log.Printf("[PeerConnection %s] Track read error: %v", userID, readErr)
				}
				return
			}

			if _, writeErr := trackLocal.Write(buf[:n]); writeErr != nil {
				if writeErr != io.EOF {
					log.Printf("[PeerConnection %s] Track write error: %v", userID, writeErr)
				}
				return
			}
		}
	})

	log.Printf("[Coordinator] Starting Signal phase for %s", userID)
	call.Signal()
}
