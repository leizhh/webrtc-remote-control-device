package client

import (
	"errors"
	"flag"
	"github.com/pion/webrtc/v2"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"strings"
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
	//获取项目的执行路径
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config := viper.New()
	config.AddConfigPath(path) //设置读取的文件路径
	config.AddConfigPath(path + "/config")
	config.SetConfigName("config") //设置读取的文件名
	config.SetConfigType("yaml")   //设置文件的类型
	//尝试进行配置读取
	if err := config.ReadInConfig(); err != nil {
		return err
	}

	deviceId := flag.String("device-id", config.GetString("DeviceId"), "device id")
	password := flag.String("password", config.GetString("Password"), "password")
	serverAddr := flag.String("server-addr", config.GetString("ServerAddr"), "WebRTC-Remote-Control server address")
	sshHost := flag.String("ssh-host", config.GetString("SSHHost"), "local ssh host")
	sshPort := flag.Int("ssh-port", config.GetInt("SSHPort"), "local ssh port")
	audioSrc := flag.String("audio-src", config.GetString("AudioSrc"), "GStreamer audio src")
	videoSrc := flag.String("video-src", config.GetString("VideoSrc"), "GStreamer video src")
	iceServerUrl := flag.String("ice-url", "", "ICE server url")
	iceServerUsername := flag.String("ice-username", config.GetString("AudioSrc"), "ICE server username")
	iceServerCredential := flag.String("ice-credential", config.GetString("VideoSrc"), "ICE server credential")

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

	iceServers := []webrtc.ICEServer{}
	if *iceServerUrl != "" {
		iceServer := webrtc.ICEServer{}
		iceServer.URLs = []string{*iceServerUrl}
		if strings.Index(*iceServerUrl, "turn") != -1 {
			iceServer.Username = *iceServerUsername
			iceServer.Credential = *iceServerCredential
			iceServer.CredentialType = webrtc.ICECredentialTypePassword
		}
		iceServers = append(iceServers, iceServer)
	} else {
		config.UnmarshalKey("ICEServers", iceServers)
	}

	Conf = &Config{
		RTCConfig:  webrtc.Configuration{ICEServers: iceServers},
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
