package routes

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

type Session interface {
	JoinCall(id string)
	AddPeer(peer *Peer)
	RemovePeer(userID string)
	AddTrack(track *webrtc.TrackRemote)
	RemoveTrack(track *webrtc.TrackRemote)
	SendAnswer(message webrtc.SessionDescription, userID string)
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
		mutex:  sync.RWMutex{},
		peers:  map[string]*Peer{},
		tracks: map[string]*webrtc.TrackLocalStaticRTP{},
	}
}

func (call *Call) AddPeer(peer *Peer) {
	call.mutex.Lock()
	defer func() {
		call.mutex.Unlock()
	}()

	call.peers[peer.id] = peer
}

func (call *Call) RemovePeer(userID string) {
	call.mutex.Lock()
	defer func() {
		call.mutex.Unlock()
		call.Signal()
	}()

	delete(call.peers, userID)
}

func (call *Call) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	call.mutex.Lock()
	defer func() {
		call.mutex.Unlock()
		call.Signal()
	}()
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
	if err != nil {
		panic(err)
	}

	call.tracks[track.ID()] = trackLocal
	fmt.Println("Track ", track.ID(), " was added")
	return trackLocal
}

func (call *Call) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	call.mutex.Lock()
	defer func() {
		call.mutex.Unlock()
		call.Signal()
	}()

	delete(call.tracks, track.ID())
}

func (call *Call) SendAnswer(message webrtc.SessionDescription, userID string) {
	if peer, ok := call.peers[userID]; ok {
		raw, parse_err := json.Marshal(message)
		if err := peer.socket.WriteJSON(WebsocketMessage{Type: "answer", Data: string(raw)}); err != nil && parse_err != nil {
			fmt.Println(err)
			fmt.Println(parse_err)
		}
	}
}

func (call *Call) SendOffer(message webrtc.SessionDescription, userID string) {
	if peer, ok := call.peers[userID]; ok {
		raw, parse_err := json.Marshal(message)
		if err := peer.socket.WriteJSON(WebsocketMessage{Type: "offer", Data: string(raw)}); err != nil && parse_err != nil {
			fmt.Println(err)
			fmt.Println(parse_err)
		}
	}
}

func (call *Call) SendICE(message *webrtc.ICECandidate, userID string) {
	if peer, ok := call.peers[userID]; ok {
		fmt.Println("SENDED |ICE|: ", message.ToJSON())
		raw, parse_err := json.Marshal(message.ToJSON())
		if err := peer.socket.WriteJSON(WebsocketMessage{Type: "candidate", Data: string(raw)}); err != nil && parse_err != nil {
			fmt.Println(err)
			fmt.Println(parse_err)
		}
	}
}

func (call *Call) BroadCast(message WebsocketMessage, userID string) {
	call.mutex.Lock()
	defer call.mutex.Unlock()
	for _, rec := range call.peers {
		if rec.id != userID {
			if err := rec.socket.WriteJSON(message); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (call *Call) JoinCall(id string) {
	call.mutex.Lock()
	defer call.mutex.Unlock()
	call.peers[id] = newPeer(id)
}

func (call *Call) Signal() {
	call.mutex.Lock()
	defer call.mutex.Unlock()
	attemptSync := func() (again bool) {
		for _, peer := range call.peers {

			if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				fmt.Println("Peer with userID", peer.id, "was disconnected")
				call.RemovePeer(peer.id)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range peer.connection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true
				if _, ok := call.tracks[sender.Track().ID()]; !ok {
					if err := peer.connection.RemoveTrack(sender); err != nil {
						fmt.Println("Track", sender.Track().ID(), "was removed")
						return true
					}
				}
			}

			///look into this, make sure streamers can recieve video from other
			for _, receiver := range peer.connection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			for trackID := range call.tracks {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := peer.connection.AddTrack(call.tracks[trackID]); err == nil {
						fmt.Println("New track are sending for peer", peer.id)
						return true
					} else {
						fmt.Println(err)
					}
				}
			}

			if peer.connection.PendingLocalDescription() != nil {
				fmt.Println(peer.connection.PendingLocalDescription())
				offer, err := peer.connection.CreateOffer(&webrtc.OfferOptions{
					OfferAnswerOptions: webrtc.OfferAnswerOptions{},
					ICERestart:         true,
				})
				if err != nil {
					fmt.Println("Error in CreateOffer: ", err)
					return true
				}
				if err = peer.connection.SetLocalDescription(offer); err != nil {
					fmt.Println("Offer: ", offer)
					fmt.Println("Cannot set LocalDescription: ", err)
					return false
				}

				offerString, err := json.Marshal(offer)
				if err != nil {
					fmt.Println("Marshalling failed: ", err)
					return true
				}

				if err = peer.socket.WriteJSON(&WebsocketMessage{
					Type: "offer",
					Data: string(offerString),
				}); err != nil {
					fmt.Println("Cannot write message in WsMessage: ", err)
					return true
				}
			}

		}
		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				time.Sleep(time.Second * 3)
				call.Signal()
			}()
			return
		}

		if !attemptSync() {
			fmt.Println("Signalling finished")
			break
		}
	}
}
