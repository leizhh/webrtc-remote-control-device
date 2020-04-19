package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"webrtc-device/config"
)

type Session struct {
	Type     string `json:"type"`
	Msg      string `json:"msg"`
	Data     string `json:"data"`
	DeviceId string `json:"device_id"`
}

func Reconnect(serverAddr *string) {
	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/answer"}
	fmt.Println("connecting to ", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req := &Session{}
	req.Type = "online"
	req.DeviceId = config.DeviceId

	if err := ws.WriteJSON(req); err != nil {
		fmt.Println(err)
		return
	}

	RTCReconnect(ws)
}

func RTCReconnect(ws *websocket.Conn) {
	var resp Session
	stopRTC := make(chan string, 1)

	ws.SetCloseHandler(func(code int, text string) error {
		fmt.Println("client is offline")
		stopRTC <- "close"
		return nil
	})

	for {
		if err := ws.ReadJSON(&resp); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				fmt.Printf("error: %v", err)
			}
			stopRTC <- "close"
			fmt.Println("Connection has been closed")
			return
		}

		if resp.Type == "offer" {
			go startRTC(ws, resp.Data, stopRTC)
		}

		if resp.Type == "error" {
			fmt.Println(resp.Msg)
			ws.Close()
			return
		}
	}
}
