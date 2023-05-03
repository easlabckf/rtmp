package rtmp

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type Server struct {
	Host string
	Port int

	inited     bool
	serverPort int
	server     net.Listener

	Connections map[int]Connection
	Handlers    map[string]Handler
}

func (srv Server) Init() bool {
	log.Println(fmt.Printf("~Server -Init() [port:%d]", srv.Port))

	if srv.inited {
		log.Fatalln("~Server -Init() RTMP Server is already running.")

		return false
	}

	srv.serverPort = srv.Port
	srv.inited = true

	//Init server socket
	if !srv.bindServer() {
		return false
	}

	//Create threads
	srv.createPriorityThread()
	srv.server.Accept()

	//Return ok
	return true

}

func (srv Server) bindServer() bool {
	port := strconv.Itoa(srv.Port)
	listen, err := net.Listen("tcp", "localhost:"+port)

	if err != nil {
		log.Fatal(err)
		return false
	}

	srv.server = listen

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			return false
		}

		go echo_handler(conn)

	}
}

func echo_handler(conn net.Conn) {
	tmp := make([]byte, 4096)

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}
		fmt.Println(tmp[:n])
		fmt.Println(n)
	}
}

func (srv Server) createPriorityThread() bool {
	return true
}
