package rtmp

import (
	"fmt"
	"io"
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
	log.Println(fmt.Sprintf("Server is starting at port:%d", srv.Port))

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
		go srv.handleConnection(conn)
	}
}

func (srv *Server) handleConnection(conn *Connection) (err error) {
	if err := HandshakeServer(conn); err != nil {
		conn.Close()
		log.Error("handleConn HandshakeServer err: ", err)
		return err
	}

	connHandler := NewHandler(conn)

	if err := connHandler.InitConnection(); err != nil {
		conn.Close()
		log.Error("connHandler read msg err: ", err)
		return err
	}

	app, name, _ := connHandler.GetInfo()

	log.Println(fmt.Sprintf("handleConn: IsPublisher=%v", connHandler.IsPublisher()))
	log.Println(fmt.Sprintf("connection is initialized successfully %s %s", app, name))

	if connHandler.IsPublisher() {
		for {
			var buff [1024]byte

			if _, err = io.ReadFull(connHandler.conn.rw, buff[0:]); err != nil {
				return
			}

			fmt.Printf("%x\n", buff)
		}
	}

	return nil
}

func (srv Server) createPriorityThread() bool {
	return true
}
