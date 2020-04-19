package main

import (
	"fmt"
	"time"
	"webrtc-device/client"
)

func main() {
	if err := client.InitConfig(); err != nil {
		fmt.Println(err)
		return
	}

	for {
		client.Reconnect()
		time.Sleep(5 * time.Second)
		fmt.Println("Reconnect with the signaling server")
	}
}
