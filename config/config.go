package config

import (
	"github.com/pion/webrtc/v2"
)

var (
	RTCConfig = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:118.89.111.54:3478"},
			}, 
		},
	}

	DefaultDeviceId   = ""
	DefaultServerAddr = ""
	DefaultAudioSrc   = "audiotestsrc"
	DefaultVideoSrc = "videotestsrc"
	//DefaultVideoSrc = "autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue"
	DefaultSSHHost  = "127.0.0.1"
	DefaultSSHPort  = 22
)
