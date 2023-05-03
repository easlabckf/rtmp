package main

import (
	"rtmp-example/rtmp"
)

const (
	PORT = 1935
)

func main() {
	rtmpserver := rtmp.Server{Port: PORT}
	rtmpserver.Init()
}
