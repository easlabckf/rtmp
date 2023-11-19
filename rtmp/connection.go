package rtmp

import (
	"encoding/binary"
	"net"
	"rtmp-example/internal/bitops"
	"rtmp-example/internal/pool"
	"time"
)

const (
	_ = iota
	idSetChunkSize
	idAbortMessage
	idAck
	idUserControlMessages
	idWindowAckSize
	idSetPeerBandwidth
)

type Connection struct {
	net.Conn
	bufferSize          int
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReadWriter
	chunks              map[uint32]ChunkStream
	pool                *pool.Pool
}

func NewConn(c net.Conn, buffSize int) *Connection {
	return &Connection{
		Conn:                c,
		bufferSize:          buffSize,
		rw:                  NewReadWriter(c, buffSize),
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		chunks:              make(map[uint32]ChunkStream),
		pool:                pool.NewPool(),
	}
}

func (conn *Connection) Read(c *ChunkStream) error {
	for {
		h, _ := conn.rw.ReadUintBE(1)

		format := h >> 6
		csid := h & 0x3f

		cs, ok := conn.chunks[csid]

		if !ok {
			cs = ChunkStream{}
			conn.chunks[csid] = cs
		}

		cs.tmpFromat = format
		cs.CSID = csid

		err := cs.readChunk(conn.rw, conn.remoteChunkSize, conn.pool)
		if err != nil {
			return err
		}

		conn.chunks[csid] = cs

		//fmt.Printf("%+v\n", cs.got)

		if cs.full() {
			*c = cs
			break
		}
	}

	conn.handleControlMsg(c)

	conn.ack(c.Length)

	return nil
}

func (conn *Connection) handleControlMsg(c *ChunkStream) {
	if c.TypeID == idSetChunkSize {
		conn.remoteChunkSize = binary.BigEndian.Uint32(c.Data)
	} else if c.TypeID == idWindowAckSize {
		conn.remoteWindowAckSize = binary.BigEndian.Uint32(c.Data)
	}
}

func (conn *Connection) ack(size uint32) {
	conn.received += uint32(size)
	conn.ackReceived += uint32(size)

	if conn.received >= 0xf0000000 {
		conn.received = 0
	}

	if conn.ackReceived >= conn.remoteWindowAckSize {
		cs := conn.NewAck(conn.ackReceived)
		cs.writeChunk(conn.rw, int(conn.chunkSize))
		conn.ackReceived = 0
	}
}

func (conn *Connection) NewAck(size uint32) ChunkStream {
	return initControlMsg(idAck, 4, size)
}

func initControlMsg(id, size, value uint32) ChunkStream {
	ret := ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   id,
		StreamID: 0,
		Length:   size,
		Data:     make([]byte, size),
	}

	bitops.PutU32BE(ret.Data[:size], value)
	return ret
}

func (conn *Connection) Flush() error {
	return conn.rw.Flush()
}

func (conn *Connection) Close() error {
	return conn.Conn.Close()
}

func (conn *Connection) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *Connection) LocalAddr() net.Addr {
	return conn.Conn.LocalAddr()
}

func (conn *Connection) SetDeadline(t time.Time) error {
	return conn.Conn.SetDeadline(t)
}

func (conn *Connection) NewSetChunkSize(size uint32) ChunkStream {
	return initControlMsg(idSetChunkSize, 4, size)
}

func (conn *Connection) NewWindowAckSize(size uint32) ChunkStream {
	return initControlMsg(idWindowAckSize, 4, size)
}

func (conn *Connection) NewSetPeerBandwidth(size uint32) ChunkStream {
	ret := initControlMsg(idSetPeerBandwidth, 5, size)
	ret.Data[4] = 2
	return ret
}

func (conn *Connection) Write(c *ChunkStream) error {
	if c.TypeID == idSetChunkSize {
		conn.chunkSize = binary.BigEndian.Uint32(c.Data)
	}
	return c.writeChunk(conn.rw, int(conn.chunkSize))
}
