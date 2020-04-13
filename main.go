package main

import (
	"fmt"
	"flag"
	"net/url"
	"time"
	"github.com/gorilla/websocket"
)

type Session struct {
	Type string `json:"type"`
	Msg string `json:"msg"`
	Data string `json:"data"`
	DeviceId string `json:"device_id"`
}

var (
	wsAddr *string
	audioSrc *string
	videoSrc *string
)

func main(){
	audioSrc = flag.String("audio-src", defaultAudioSrc, "GStreamer audio src")
	videoSrc = flag.String("video-src", defaultVideoSrc, "GStreamer video src")
	wsAddr = flag.String("websocket-addr", defaultWSAddr, "websocket service address")
	flag.Parse()

	var ws *websocket.Conn
	for {
		ws = reconnect()
		hub(ws)
		time.Sleep(5 * time.Second)
		fmt.Println("Reconnect with the signaling server")
	}
}

func reconnect() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: *wsAddr, Path: "/answer"}
	fmt.Println("connecting to ", u.String())
	
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
		fmt.Println(err)
		return nil
	}

	req := &Session{}
	req.Type = "online"
	req.DeviceId = deviceId
 
	if err = ws.WriteJSON(req); err != nil {
		fmt.Println(err)
		return nil
	}

	return ws
}