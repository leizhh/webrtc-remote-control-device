package client

import (
	"errors"
	"flag"
	"github.com/pion/webrtc/v2"
	"os/user"
	"webrtc-device/config"
)

var Conf *Config

type Config struct {
	RTCConfig  webrtc.Configuration
	DeviceId   string
	Password   string
	ServerAddr string
	AudioSrc   string
	VideoSrc   string
	SSHHost    string
	SSHPort    int
}

func InitConfig() error {
	deviceId := flag.String("device-id", config.DefaultDeviceId, "set device id")
	password := flag.String("password", config.DefaultPassword, "set password")
	serverAddr := flag.String("server-addr", config.DefaultServerAddr, "set server address")
	sshHost := flag.String("ssh-host", config.DefaultSSHHost, "set ssh host")
	sshPort := flag.Int("ssh-port", config.DefaultSSHPort, "set ssh-port")
	audioSrc := flag.String("audio-src", config.DefaultAudioSrc, "set GStreamer audio src")
	videoSrc := flag.String("video-src", config.DefaultVideoSrc, "set GStreamer video src")
	flag.Parse()

	if *deviceId == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		device_id := "device-" + u.Username
		deviceId = &device_id
	}

	if *serverAddr == "" {
		return errors.New("server can not be empty")
	}

	Conf = &Config{
		RTCConfig:  config.RTCConfig,
		DeviceId:   *deviceId,
		ServerAddr: *serverAddr,
		AudioSrc:   *audioSrc,
		VideoSrc:   *videoSrc,
		SSHHost:    *sshHost,
		SSHPort:    *sshPort,
		Password:   *password,
	}

	return nil
}
