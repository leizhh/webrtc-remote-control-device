package main

import (
	"github.com/pion/webrtc/v2"
)

var (
	configRTC = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:118.89.111.54:3478"},
			},
		},
	}

	deviceId = "51a8da4c-a9ce-403f-999b-6f0445e52d74"
	defaultWSAddr = "118.89.111.54:8080"
	defaultAudioSrc = "audiotestsrc"
	defaultVideoSrc = "autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue"
)