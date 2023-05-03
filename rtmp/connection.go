package rtmp

import (
	"bufio"
	"io"
	"net"
)

type Connection struct {
	net.Conn
	bufferSize int
	rw         *ReadWriter
}

type ReadWriter struct {
	*bufio.ReadWriter
	readError  error
	writeError error
}

func NewReadWriter(rw io.ReadWriter, bufSize int) *ReadWriter {
	return &ReadWriter{
		ReadWriter: bufio.NewReadWriter(bufio.NewReaderSize(rw, bufSize), bufio.NewWriterSize(rw, bufSize)),
	}
}

func NewConn(c net.Conn, buffSize int) *Connection {
	return &Connection{
		Conn:       c,
		bufferSize: buffSize,
		rw:         NewReadWriter(c, buffSize),
	}
}
