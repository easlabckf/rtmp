package rtmp

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type RTMPServer struct {
	Host string
	Port int

	inited     bool
	serverPort int
	server     net.Listener

	Connections  map[int]RTMPConnection
	Applications map[string]RTMPApplication
}

func (rtmpserver RTMPServer) Init() bool {
	log.Println(fmt.Printf("~RTMPServer -Init() [port:%d]", rtmpserver.Port))

	if rtmpserver.inited {
		log.Fatalln("~RTMPServer -Init() RTMP Server is already running.")

		return false
	}

	rtmpserver.serverPort = rtmpserver.Port
	rtmpserver.inited = true

	//Init server socket
	if !rtmpserver.bindServer() {
		return false
	}

	//Create threads
	rtmpserver.createPriorityThread()
	rtmpserver.server.Accept()

	//Return ok
	return true

}

func (rtmpserver RTMPServer) bindServer() bool {
	port := strconv.Itoa(rtmpserver.Port)
	listen, err := net.Listen("tcp", "localhost:"+port)

	if err != nil {
		log.Fatal(err)
		return false
	}

	rtmpserver.server = listen

	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			return false
		}
		conn.Write(make([]byte, 1))
	}
}

func (rtmpserver RTMPServer) createPriorityThread() bool {
	port := strconv.Itoa(rtmpserver.Port)
	listen, err := net.Listen("tcp", "localhost:"+port)

	if err != nil {
		log.Fatal(err)
		return false
	}

	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			return false
		}
		conn.Write(make([]byte, 1))
	}
}
