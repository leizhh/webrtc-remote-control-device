package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"os/signal"
)

type Session struct {
	Type     string `json:"type"`
	Msg      string `json:"msg"`
	Data     string `json:"data"`
	DeviceId string `json:"device_id"`
}

func Reconnect() {
	u := url.URL{Scheme: "ws", Host: Conf.ServerAddr, Path: "/answer"}
	fmt.Println("connecting to ", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, os.Interrupt, os.Kill)
	go signalHandler(ws, signal_ch)

	req := &Session{}
	req.Type = "online"
	req.DeviceId = Conf.DeviceId

	if err := ws.WriteJSON(req); err != nil {
		fmt.Println(err)
		return
	}

	RTCReconnect(ws)
	signal.Stop(signal_ch)
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
			if Conf.Password != "" {
				pwResp := &Session{}
				pwResp.Type = "password"
				pwResp.DeviceId = Conf.DeviceId
				ws.WriteJSON(pwResp)

				var pwReq Session
				ws.ReadJSON(&pwReq)
				if pwReq.Type == "password" {
					if pwReq.Data != Conf.Password {
						pwResp := &Session{}
						pwResp.Type = "error"
						pwResp.Msg = "wrong password"
						ws.WriteJSON(pwResp)
						ws.Close()
						return
					}
				} else {
					ws.Close()
					return
				}
			}
			go startRTC(ws, resp.Data, stopRTC)
		}

		if resp.Type == "error" {
			fmt.Println(resp.Msg)
			ws.Close()
			return
		}
	}
}

func signalHandler(ws *websocket.Conn, signal_ch chan os.Signal) {
	for msg := range signal_ch {
		fmt.Println("signal:", msg)
		err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			fmt.Println("write close:", err)
		}
		ws.Close()
		os.Exit(0)
	}
}
