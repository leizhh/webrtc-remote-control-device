package main

import (
	"github.com/pion/webrtc/v2"
)

var (
	configRTC = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:118.89.111.54:3478"},
			},{
				URLs: []string{"turn:118.89.111.54:3478"},
				Username:"leizhh",
				Credential:"leizhh",
				CredentialType:webrtc.ICECredentialTypePassword,
			},
		},
	}

	deviceId = "51a8da4c-a9ce-403f-999b-6f0445e52d74"
	defaultWSAddr = "118.89.111.54:8080"
	//defaultAudioSrc = "audiotestsrc"
	//defaultVideoSrc = "videotestsrc"
	defaultAudioSrc = "audiotestsrc"
	defaultVideoSrc = "autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue"
	SSHHost = "127.0.0.1"
	SSHPort = 22
)