package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"math/rand"
	"strconv"
	"strings"
	"webrtc-device/config"
	gst "webrtc-device/lib/gstreamer-src"
	"webrtc-device/lib/signal"
)

var (
	audioSrc *string
	videoSrc *string
)

func SetMediaSrc(audio *string, video *string) {
	audioSrc = audio
	videoSrc = video
}

func startRTC(ws *websocket.Conn, offer string, stopRTC chan string) {
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config.RTCConfig)
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
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		fmt.Println(err)
		return
	}

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "pion2")
	if err != nil {
		fmt.Println(err)
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		fmt.Println(err)
		return
	}

	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "SSH" {
			sshDataChannelHandler(dc)
		}
		if dc.Label() == "Control" {
			controlDataChannelHandler(dc)
		}
	})

	// Set the remote SessionDescription
	offer_ := webrtc.SessionDescription{}
	signal.Decode(offer, &offer_)
	err = peerConnection.SetRemoteDescription(offer_)
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
	req.DeviceId = config.DeviceId
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

	<-stopRTC
	close(stopRTC)

	audioPipeline.Stop()
	videoPipeline.Stop()
	peerConnection.Close()

	return
}

func controlDataChannelHandler(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		err := dc.SendText("please input command")
		if err != nil {
			fmt.Println("write data error:", err)
			dc.Close()
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		result := controlHandler(msg.Data)
		dc.SendText(result)
	})
	dc.OnClose(func() {
		fmt.Printf("Close Control socket")
	})
}

func sshDataChannelHandler(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		for {
			var user string
			var password string
			rtcin := make(chan string)
			step := make(chan string)

			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				user = string(msg.Data)
				fmt.Println(user)
				dc.OnMessage(func(msg webrtc.DataChannelMessage) {
					password = string(msg.Data)
					fmt.Println(password)
					step <- ""
				})
			})

			<-step
			sshSession, err := initSSH(user, password, config.SSHHost, config.SSHPort, dc, rtcin)
			if err != nil {
				dc.SendText(err.Error())
				continue
			}
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				msg_ := string(msg.Data)

				if len(msg_) >= 10 {
					ss := strings.Fields(msg_)
					if ss[0] == "resize" {
						cols, _ := strconv.Atoi(ss[1])
						rows, _ := strconv.Atoi(ss[2])
						sshSession.WindowChange(cols, rows)
						fmt.Println(msg_)
						return
					}
				}

				rtcin <- msg_
			})
			break
		}
	})
	dc.OnClose(func() {
		fmt.Printf("Close SSH socket")
	})
}
