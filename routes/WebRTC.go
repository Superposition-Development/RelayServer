package routes

import (
	"encoding/json"
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
	return &Peer{
		id:    id,
		mutex: sync.RWMutex{},
	}
}

func (p *Peer) WriteJSON(json interface{}) error {
	p.wsMutex.Lock()
	defer p.wsMutex.Unlock()
	if p.socket == nil {
	}
	return p.socket.WriteJSON(json)
}

func (peer *Peer) SetSocket(socket *websocket.Conn) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	peer.socket = socket
}

func (peer *Peer) SetPeerConnection(conn *webrtc.PeerConnection) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	peer.connection = conn
}

func (peer *Peer) ReactOnOffer(offerStr string) (webrtc.SessionDescription, error) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerStr,
	}

	if err := peer.connection.SetRemoteDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	answer, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err := peer.connection.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	return answer, nil
}

func (peer *Peer) ReactOnAnswer(answerStr string) error {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerStr,
	}
	return peer.connection.SetRemoteDescription(answer)
}

type Call struct {
	id     string
	mutex  sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
}

func NewCall(id string) *Call {
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
	call.mutex.Unlock()
}

func (call *Call) RemovePeer(peerID string) {
	call.mutex.Lock()
	delete(call.peers, peerID)
	call.mutex.Unlock()

	call.Signal()
}

func (coordinator *Coordinator) RemoveUserFromCall(userID string, callID string) {
	if call, ok := coordinator.Sessions[callID]; ok {
		delete(call.peers, userID)
	}
}

func (call *Call) SendAnswer(message webrtc.SessionDescription, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		return
	}

	msg := WebsocketMessage{
		Type: "answer",
		Data: message,
	}

	if err := peer.WriteJSON(msg); err != nil {

	}
}

func (call *Call) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
	if err != nil {
		return nil
	}

	call.mutex.Lock()
	call.tracks[track.ID()] = trackLocal
	call.mutex.Unlock()

	call.Signal()
	return trackLocal
}

func (call *Call) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	if track == nil {
		return
	}
	call.mutex.Lock()
	delete(call.tracks, track.ID())
	call.mutex.Unlock()

	call.Signal()
}

func (call *Call) SendICE(message *webrtc.ICECandidate, peerID string) {
	peer, ok := call.GetPeer(peerID)
	if !ok {
		return
	}

	raw, err := json.Marshal(message.ToJSON())
	if err != nil {
		return
	}

	if err := peer.WriteJSON(WebsocketMessage{Type: "candidate", Data: string(raw)}); err != nil {
	}
}

func (call *Call) Signal() {
	call.mutex.Lock()
	defer call.mutex.Unlock()

	attemptSync := func() bool {
		for _, peer := range call.peers {
			if peer.connection == nil {
				continue
			}

			if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
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
					if err := peer.connection.RemoveTrack(sender); err == nil {
						return true
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
					if _, err := peer.connection.AddTrack(track); err == nil {
						return true
					}
				}
			}

			if peer.connection.SignalingState() == webrtc.SignalingStateStable {
				offer, err := peer.connection.CreateOffer(nil)
				if err != nil {
					return false
				}

				if err = peer.connection.SetLocalDescription(offer); err != nil {
					return false
				}

				offerString, err := json.Marshal(offer)
				if err != nil {
					return false
				}

				if err = peer.WriteJSON(WebsocketMessage{
					Type: "offer",
					Data: string(offerString),
				}); err != nil {
					return false
				}
			}
		}
		return false
	}

	for syncAttempt := 0; syncAttempt < 25; syncAttempt++ {
		if !attemptSync() {
			break
		}
	}
}

type Coordinator struct {
	mutex    sync.RWMutex
	Sessions map[string]*Call
}

func NewCoordinator() *Coordinator {
	return &Coordinator{Sessions: make(map[string]*Call)}
}

func (coordinator *Coordinator) AddUserToCall(userID string, callID string, socket *websocket.Conn) {
	coordinator.mutex.Lock()
	call, ok := coordinator.Sessions[callID]
	if !ok {
		call = NewCall(callID)
		coordinator.Sessions[callID] = call
	}
	coordinator.mutex.Unlock()

	peer := newPeer(userID)
	peer.SetSocket(socket)

	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}
	peer.SetPeerConnection(conn)

	call.AddPeer(peer)

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := conn.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			return
		}
	}

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateFailed:
			_ = conn.Close()
		case webrtc.PeerConnectionStateClosed:
			call.RemovePeer(userID)
		}
	})

	conn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice != nil {
			call.SendICE(ice, userID)
		}
	})

	conn.OnTrack(func(trackRemote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		trackLocal := call.AddTrack(trackRemote)
		if trackLocal == nil {
			return
		}
		defer call.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := trackRemote.Read(buf)
			if err != nil {
				return
			}
			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})

	call.Signal()
}
