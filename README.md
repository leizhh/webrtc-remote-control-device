# <center>WebRTC Remote Control</center>
<hr>
<center>基于WebRTC的实时远程控制系统,支持视频监控、远程控制与SSH登录</center>
<br>

## 使用说明
### [server端](https://github.com/leizhh/webrtc-remote-control-server)
### 下载：
```
git clone https://github.com/leizhh/webrtc-remote-control-server.git
```
### 运行：
```
go run main.go
```

### [device端](https://github.com/leizhh/webrtc-remote-control-device)
### 安装 GStreamer
项目依赖gstreamer，安装方式：
#### Debian/Ubuntu
`sudo apt-get install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly`
#### Windows MinGW64/MSYS2
`pacman -S mingw-w64-x86_64-gstreamer mingw-w64-x86_64-gst-libav mingw-w64-x86_64-gst-plugins-good mingw-w64-x86_64-gst-plugins-bad mingw-w64-x86_64-gst-plugins-ugly`
#### macOS
` brew install gst-plugins-good gst-plugins-ugly pkg-config && export PKG_CONFIG_PATH="/usr/local/opt/libffi/lib/pkgconfig"`

### 下载：
```
git clone https://github.com/leizhh/webrtc-remote-control-device.git
```

### 运行：
```
go run main.go -server-addr="$your_server_addr"
```

## 依赖
GO VERSION >= 1.13  

## 感谢
WebRTC: https://github.com/pions/webrtc  
TURN/STUN: https://github.com/pion/turn  
Terminal: https://xtermjs.org/
