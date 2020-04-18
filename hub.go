package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"math/rand"
	"strconv"
	"strings"
	gst "webrtc-device/lib/gstreamer-src"
	"webrtc-device/lib/signal"
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
	stopRTC := make(chan string, 1)
	ws.SetCloseHandler(func(code int, text string) error {
		fmt.Println("client is offline")
		stopRTC <- "close"
		return nil
	})

	for {
		err := ws.ReadJSON(&resp)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
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
				fmt.Println("start rtc:", err)
			}
		}

		if resp.Type == "error" {
			fmt.Println(resp.Msg)
			ws.Close()
			return
		}
	}
}

func startRTC(ws *websocket.Conn, offer webrtc.SessionDescription, stopRTC chan string) {
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
			sshDataChannel(dc)
		}
		if dc.Label() == "Control" {
			controlDataChannel(dc)
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

	<-stopRTC
	close(stopRTC)

	audioPipeline.Stop()
	videoPipeline.Stop()
	peerConnection.Close()

	return
}

func controlDataChannel(dc *webrtc.DataChannel) {
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

func sshDataChannel(dc *webrtc.DataChannel) {
	var user string
	var password string

	dc.OnOpen(func() {
		for {
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
			sshSession, err := initSSH(SSHHost, user, password, SSHPort, dc, rtcin)
			if err != nil {
				dc.SendText(err.Error())
				continue
			}
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				msg_ := string(msg.Data)
				fmt.Println(msg_)

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
