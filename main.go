package main

import(
	"fmt"
	"flag"
	"math/rand"
	"github.com/pion/webrtc/v2"
	"net/url"
	"github.com/gorilla/websocket"
	"encoding/json"
	gst "github.com/pion/example-webrtc-applications/internal/gstreamer-src"
	"github.com/pion/example-webrtc-applications/internal/signal"
)

type WSRequest() struct {
	Type string `json:"type"`
	Msg string `json:"msg"`
	Data string `json:"data"`
	DeviceId string `json:"device_id"`
}

var (
	deviceId = "51a8da4c-a9ce-403f-999b-6f0445e52d74"
	peerConnection webrtc.*PeerConnection
	audioSrc string
	videoSrc string
	wsAddr string
	audioTrack webrtc.Track
	videoTrack webrtc.Track
	config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:118.89.111.54:3478"},
			},
		},
	}
)

func main(){
	audioSrc = flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc = flag.String("video-src", "autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue", "GStreamer video src")
	wsAddr = flag.String("websocket-addr", "118.89.111.54:8080", "websocket service address")
	flag.Parse()

	initWebRTC()
	go reConnection()

	// Block forever
	select {}
}

func initWebRTC(){
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	
	// Create a audio track
	audioTrack, err = peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pion1")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// Create a video track
	videoTrack, err = peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "pion2")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}
}

func reConnection(){
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/answer"}
	fmt.Println("connecting to ", u.String())
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
		return
	}
	defer conn.Close()

	conn.SetCloseHandler(func(code int, text string) error {
		initWebRTC()
		go reConnection()
		return nil
	})

	req := &WSRequest{}
	req.Type = "online"
	req.DeviceId = deviceId
 
    data, _ := json.Marshal(req)
	conn.WriteMessage(websocket.TextMessage, data)
	
	for{
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,websocket.CloseNoStatusReceived) {
				fmt.Printf("error: %v", err)
			}
			return
		}
	
		resp := make(map[string]string)
		err = json.Unmarshal(msg, &resp)

		if resp["type"] == "offer" {
			offer := webrtc.SessionDescription{}
			signal.Decode(resp["data"], &offer)

			// Set the remote SessionDescription
			err = peerConnection.SetRemoteDescription(offer)
				if err != nil {
				panic(err)
			}

			// Create an answer
			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}

			// Sets the LocalDescription, and starts our UDP listeners
			err = peerConnection.SetLocalDescription(answer)
			if err != nil {
				panic(err)
			}

			// Output the answer in base64 so we can paste it in browser
			req := &WSRequest{}
			req.Type = "answer"
			req.DeviceId = deviceId
			req.Data = signal.Encode(answer)
		 
			data, _ := json.Marshal(req)
			conn.WriteMessage(websocket.TextMessage, data)

			// Start pushing buffers on these tracks
			gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc).Start()
			gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{videoTrack}, *videoSrc).Start()
		}
	}
}