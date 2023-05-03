package main

import (
	"fmt"

	"rtmp-example/rtmp"
)

const (
	PORT = 1935
)

func main() {
	fmt.Println("asd")
	rtmpserver := rtmp.RTMPServer{Port: PORT}
	rtmpserver.Init()
}
