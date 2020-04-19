package main

import (
	"flag"
	"fmt"
	"time"
	"webrtc-device/client"
	"webrtc-device/config"
)

func main() {
	audioSrc := flag.String("audio-src", config.DefaultAudioSrc, "GStreamer audio src")
	videoSrc := flag.String("video-src", config.DefaultVideoSrc, "GStreamer video src")
	serverAddr := flag.String("websocket-addr", config.DefaultServerAddr, "websocket service address")
	flag.Parse()

	client.SetMediaSrc(audioSrc, videoSrc)

	for {
		client.Reconnect(serverAddr)
		time.Sleep(5 * time.Second)
		fmt.Println("Reconnect with the signaling server")
	}
}
