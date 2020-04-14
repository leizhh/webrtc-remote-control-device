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
	stopRTC := make(chan string,1)
	ws.SetCloseHandler(func(code int, text string) error {
		fmt.Println("client is offline")
		stopRTC <- "close"
		return nil
	})

	for {
		err := ws.ReadJSON(&resp)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,websocket.CloseNoStatusReceived) {
				fmt.Printf("error: %v", err)
			}
			stopRTC <- "close"
			fmt.Println("RTC has been closed")
			return
		}

		if resp.Type == "offer" {
			offer := webrtc.SessionDescription{}
			signal.Decode(resp.Data, &offer)
			go startRTC(ws, offer, stopRTC)
			if err != nil {
				fmt.Println("start rtc:",err)
			}
		}
	}
}

func startRTC(ws *websocket.Conn, offer webrtc.SessionDescription, stopRTC chan string){
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(configRTC)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	
	// Create a audio track
	audioTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pion1")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "pion2")
	if err != nil {
		fmt.Println(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		fmt.Println(err)
		return
	}

	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "SSH" {
			ssh, err := net.Dial("tcp", fmt.Sprintf("%s:%d", SSHHost, SSHPort))
			if err != nil {
				fmt.Println("ssh dial failed:", err)
				peerConnection.Close()
			} else {
				fmt.Println("Connect SSH socket")
				sshDataChannel(dc, ssh)
			}
		}
		if dc.Label() == "Control" {
			
		}
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		fmt.Println(err)
	}

	// Output the answer in base64 so we can paste it in browser
	req := &Session{}
	req.Type = "answer"
	req.DeviceId = deviceId
	req.Data = signal.Encode(answer)
 
	if err = ws.WriteJSON(req); err != nil {
		fmt.Println(err)
		return
	}

	// Start pushing buffers on these tracks
	audioPipeline := gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc)
	videoPipeline := gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{videoTrack}, *videoSrc)
	audioPipeline.Start()
	videoPipeline.Start()

	
	<- stopRTC
	close(stopRTC)

	audioPipeline.Stop()
	videoPipeline.Stop()
	peerConnection.Close()

	return
}

func controlDataChannel(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {	
		err := dc.SendText("please input command")
		if err != nil{
			fmt.Println("write data error:", err)
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Println(msg)
	})
	dc.OnClose(func() {
		fmt.Printf("Close Control socket")
		ssh.Close()
	})
}

func sshDataChannel(dc *webrtc.DataChannel, ssh net.Conn) {
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