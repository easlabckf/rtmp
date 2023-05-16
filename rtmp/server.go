package rtmp

import (
	"fmt"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
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

func (srv *Server) Init() bool {
	log.Println(fmt.Printf("Server is starting at port:%d", srv.Port))

	if srv.inited {
		log.Error("Server is already initialized")
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

func (srv *Server) bindServer() bool {
	port := strconv.Itoa(srv.Port)
	listen, err := net.Listen("tcp", "localhost:"+port)

	if err != nil {
		log.Error("network listen error: ", err)
		return false
	}

	defer listen.Close()

	for {
		netconn, err := listen.Accept()
		if err != nil {
			log.Error("network accept init error: ", err)
			return false
		}

		conn := NewConn(netconn, 4*1024)
		srv.handleConnection(conn)

	}
}

func (srv *Server) handleConnection(conn *Connection) (err error) {
	if err := HandshakeServer(conn); err != nil {
		conn.Close()
		log.Error("handleConn HandshakeServer err: ", err)
		return err
	}

	connHandler := NewHandler(conn, 5)
	connHandler.bufferSize = 2

	return
}

func (srv Server) createPriorityThread() bool {
	return true
}
