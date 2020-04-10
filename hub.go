package main

import (
	"fmt"
	"github.com/pion/webrtc/v2"
	"github.com/gorilla/websocket"
	"webrtc-device/lib/signal"
	"math/rand"
	"net"
	"io"
	gst "webrtc-device/lib/gstreamer-src"
)

type Wrap struct {
	*webrtc.DataChannel
}

var (
	peerConnection *webrtc.PeerConnection
)

func (rtc *Wrap) Write(data []byte) (int, error) {
	err := rtc.DataChannel.Send(data)
	return len(data), err
}

func hub(ws *websocket.Conn) {
	var resp Session
	for {
		err := ws.ReadJSON(&resp)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,websocket.CloseNoStatusReceived) {
				fmt.Printf("error: %v", err)
			}
			return
		}

		if resp.Type == "offer" {
			offer := webrtc.SessionDescription{}
			signal.Decode(resp.Data, &offer)
			err = startRTC(ws, offer)
			if err != nil {
				fmt.Println("start rtc:",err)
			}
		}
	}
}

func startRTC(ws *websocket.Conn, offer webrtc.SessionDescription) error{
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(configRTC)
	if err != nil {
		return err
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	
	// Create a audio track
	audioTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pion1")
	if err != nil {
		return err
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "pion2")
	if err != nil {
		return err
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
		return err
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return err
	}

	// Output the answer in base64 so we can paste it in browser
	req := &Session{}
	req.Type = "answer"
	req.DeviceId = deviceId
	req.Data = signal.Encode(answer)
 
	if err = ws.WriteJSON(req); err != nil {
		return err
	}

	// Start pushing buffers on these tracks
	gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc).Start()
	gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{videoTrack}, *videoSrc).Start()

	for{}
	//return nil
}

func DataChannel(dc *webrtc.DataChannel, ssh net.Conn) {
	dc.OnOpen(func() {	
		err := dc.SendText("OPEN_RTC_CHANNEL")
		if err != nil{
			fmt.Println("write data error:", err)
		}
		io.Copy(&Wrap{dc}, ssh)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		ssh.Write(msg.Data)
	})
	dc.OnClose(func() {
		fmt.Printf("Close SSH socket")
		ssh.Close()
	})
}