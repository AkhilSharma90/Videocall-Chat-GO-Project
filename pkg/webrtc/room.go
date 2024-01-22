package webrtc

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/webrtc/v3"
)
// RoomConn handles the WebSocket connection and WebRTC setup for a room
func RoomConn(c *websocket.Conn, p *Peers) {
	// Configure WebRTC based on environment (development or production)
	var config webrtc.Configuration
	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		config = turnConfig
	}

	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Print(err)
		return
	}
	defer peerConnection.Close()

	// Add video and audio transceivers for receiving streams
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return
		}
	}

	// Create a new PeerConnectionState
	newPeer := PeerConnectionState{
		PeerConnection: peerConnection,
		Websocket: &ThreadSafeWriter{
			Conn:  c,
			Mutex: sync.Mutex{},
		}}

	// Add our new PeerConnection to global list
	p.ListLock.Lock()
	p.Connections = append(p.Connections, newPeer)
	p.ListLock.Unlock()

	log.Println(p.Connections)

	// Trickle ICE. Emit server candidate to client
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		// Convert ICECandidate to JSON format
		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}
		// Send ICE candidate to the client over WebSocket
		if writeErr := newPeer.Websocket.WriteJSON(&websocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	// If PeerConnection is closed remove it from global list
	peerConnection.OnConnectionStateChange(func(pp webrtc.PeerConnectionState) {
		switch pp {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			p.SignalPeerConnections()
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// Create a track to fan out our incoming video to all peers
		trackLocal := p.AddTrack(t)
		if trackLocal == nil {
			return
		}
		defer p.RemoveTrack(trackLocal)
 		// Read from the remote track and write to the local track
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

	// Signal existing peers about the new connection
	p.SignalPeerConnections()
	// Handle WebSocket messages
	message := &websocketMessage{}
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}

		switch message.Event {
		case "candidate":
			// Handle ICE candidate message from the client
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}
			
			// Add ICE candidate to the PeerConnection
                       if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		case "answer":
			// Handle answer message from the client
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}
			
			// Set remote description to complete the offer/answer exchange
                      if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
