package routes

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v3"
)

type Session interface {
	JoinCall(id string)
	AddPeer(peer *Peer)
	RemovePeer(userID string)
	AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP
	RemoveTrack(track *webrtc.TrackLocalStaticRTP)
	SendAnswer(message webrtc.SessionDescription, userID string)
	SendOffer(message webrtc.SessionDescription, userID string)
	SendICE(candidate *webrtc.ICECandidate, userID string)
	Signal()
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

func (call *Call) AddPeer(peer *Peer) {
	call.mutex.Lock()
	defer call.mutex.Unlock()

	call.peers[peer.id] = peer
}

func (call *Call) RemovePeer(userID string) {
	call.mutex.Lock()
	defer call.mutex.Unlock()
	delete(call.peers, userID)
}

func (call *Call) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {

	call.mutex.Lock()
	defer call.mutex.Unlock()
	existing, ok := call.tracks[track.ID()]

	if ok {
		fmt.Println("Track already exists ", track.ID())
		return existing
	}

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(
		track.Codec().RTPCodecCapability,
		track.ID(),
		track.StreamID(),
	)

	if err != nil {
		panic(err)
	}

	call.tracks[track.ID()] = trackLocal

	fmt.Println("New track ", track.ID(), "stream: ", track.StreamID())

	return trackLocal
}

func (call *Call) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	call.mutex.Lock()
	defer call.mutex.Unlock()
	delete(call.tracks, track.ID())
	fmt.Println("Track", track.ID(), "removed")
}

func (call *Call) SendActionMutexCheck(userID string) (*Peer, bool) {
	call.mutex.RLock()
	peer, ok := call.peers[userID]
	call.mutex.RUnlock()

	return peer, ok
}

func (call *Call) SendAnswer(answer webrtc.SessionDescription, userID string) {
	peer, ok := call.SendActionMutexCheck(userID)

	if !ok {
		return
	}

	raw, err := json.Marshal(answer)
	if err != nil {
		fmt.Println("Decode answer error:", err)
		return
	}

	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	message := WebsocketMessage{
		Type: "answer",
		Data: string(raw),
	}

	err = SendWebsocketMessagePeer(peer.socket, message)

	if err != nil {
		fmt.Println("Answer error:", err)
	}
}

func (call *Call) SendOffer(offer webrtc.SessionDescription, userID string) {
	peer, ok := call.SendActionMutexCheck(userID)

	if !ok {
		return
	}

	raw, err := json.Marshal(offer)
	if err != nil {
		fmt.Println("decoding offer error:", err)
		return
	}

	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	message := WebsocketMessage{
		Type: "offer",
		Data: string(raw),
	}

	err = SendWebsocketMessagePeer(peer.socket, message)

	if err != nil {
		fmt.Println("send pffer error:", err)
	}
}

func (call *Call) SendICE(candidate *webrtc.ICECandidate, userID string) {
	peer, ok := call.SendActionMutexCheck(userID)

	if !ok {
		return
	}

	data, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		fmt.Println("Marshal ICE:", err)
		return
	}

	message := WebsocketMessage{
		Type: "candidate",
		Data: string(data),
	}

	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	err = SendWebsocketMessagePeer(peer.socket, message)

	if err != nil {
		fmt.Println("send ice error:", err)
	}
}

func (call *Call) Signal() {
	call.mutex.RLock()

	peers := make([]*Peer, 0, len(call.peers))
	tracks := make(map[string]*webrtc.TrackLocalStaticRTP)

	for _, peer := range call.peers {
		peers = append(peers, peer)
	}

	for id, track := range call.tracks {
		tracks[id] = track
	}

	call.mutex.RUnlock()

	for _, peer := range peers {
		if peer.connection == nil {
			continue
		}

		existingSenders := make(map[string]bool)

		for _, sender := range peer.connection.GetSenders() {
			track := sender.Track()
			if track == nil {
				continue
			}
			existingSenders[track.ID()] = true
		}

		changed := false

		for _, sender := range peer.connection.GetSenders() {
			track := sender.Track()

			if track == nil {
				continue
			}

			trackID := track.ID()

			_, exists := tracks[trackID]

			if !exists {
				fmt.Println("Removing track", trackID, "from peer", peer.id)
				err := peer.connection.RemoveTrack(sender)

				if err != nil {
					fmt.Println("RemoveTrack error", err)
					continue
				}

				changed = true
			}
		}

		for trackID, track := range tracks {
			if existingSenders[trackID] {
				continue
			}

			fmt.Println("Adding ", trackID, " to ", peer.id)

			_, err := peer.connection.AddTrack(track)

			if err != nil {
				fmt.Println("add track error :", err)
				continue
			}

			changed = true
		}

		if !changed {
			continue
		}

		if peer.connection.SignalingState() != webrtc.SignalingStateStable {
			fmt.Println("Couldn't negotiate ", peer.id, " signaling state: ", peer.connection.SignalingState())
			continue
		}

		offer, err := peer.connection.CreateOffer(nil)
		if err != nil {
			fmt.Println("CreateOffer error :", err)
			continue
		}

		err = peer.connection.SetLocalDescription(offer)

		if err != nil {
			fmt.Println("SetLocalDescription error:", err)
			continue
		}

		fmt.Println("SFU offer to ", peer.id)
		call.SendOffer(offer, peer.id)
	}
}
