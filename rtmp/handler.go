package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"rtmp-example/internal/amf"
)

/*
Types of publish commands
*/
var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

var (
	ErrReq = fmt.Errorf("req error")
)

/*
Types of commands
*/
var (
	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

type ConnectResp struct {
	FMSVer       string `amf:"fmsVer"`
	Capabilities int    `amf:"capabilities"`
}

type ConnectEvent struct {
	Level          string `amf:"level"`
	Code           string `amf:"code"`
	Description    string `amf:"description"`
	ObjectEncoding int    `amf:"objectEncoding"`
}

type PublishInfo struct {
	Name string
	Type string
}

type Handler struct {
	done          bool
	streamID      int
	isPublisher   bool
	conn          *Connection
	transactionID int
	ConnInfo      ConnectInfo
	PublishInfo   PublishInfo
	decoder       *amf.Decoder
	encoder       *amf.Encoder
	bytesw        *bytes.Buffer
	bufferSize    int
}

func NewHandler(conn *Connection) *Handler {
	return &Handler{
		conn:     conn,
		streamID: 1,
		bytesw:   bytes.NewBuffer(nil),
		decoder:  &amf.Decoder{},
		encoder:  &amf.Encoder{},
	}
}

func (handler *Handler) ReadMsg() error {
	var c ChunkStream

	for {
		if err := handler.Read(&c); err != nil {
			return err
		}

		switch c.TypeID {
		case 20, 17:
			if err := handler.handleCmdMsg(&c); err != nil {
				return err
			}
		}

		if handler.done {
			break
		}
	}
	return nil
}

func (handler *Handler) Read(c *ChunkStream) (err error) {
	return handler.conn.Read(c)
}

func (handler *Handler) handleCmdMsg(c *ChunkStream) error {
	amfType := amf.AMF0
	if c.TypeID == 17 {
		c.Data = c.Data[1:]
	}

	r := bytes.NewReader(c.Data)
	vs, err := handler.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}

	log.Println(fmt.Sprintf("rtmp req: %#v", vs))

	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case cmdConnect:
			if err = handler.connect(vs[1:]); err != nil {
				return err
			}
			if err = handler.connectResp(c); err != nil {
				return err
			}
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			log.Println(fmt.Sprint("no support command=", vs[0].(string)))
		}
	}

	return nil
}

func (handler *Handler) connect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return ErrReq
			}
			handler.transactionID = id
		case amf.Object:
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				handler.ConnInfo.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				handler.ConnInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				handler.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				handler.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}

	return nil
}

func (handler *Handler) connectResp(cur *ChunkStream) error {
	c := handler.conn.NewWindowAckSize(2500000)
	handler.conn.Write(&c)
	c = handler.conn.NewSetPeerBandwidth(2500000)
	handler.conn.Write(&c)
	c = handler.conn.NewSetChunkSize(uint32(1024))
	handler.conn.Write(&c)

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = handler.ConnInfo.ObjectEncoding

	return handler.writeMsg(cur.CSID, cur.StreamID, "_result", handler.transactionID, resp, event)
}

/*
func (handler *Handler) createStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			handler.transactionID = int(v.(float64))
		case amf.Object:
		}
	}

	return nil
}

func (handler *Handler) createStreamResp(cur *ChunkStream) error {
	return handler.writeMsg(cur.CSID, cur.StreamID, "_result", handler.transactionID, nil, handler.streamID)
}

func (handler *Handler) publishOrPlay(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				handler.PublishInfo.Name = v.(string)
			} else if k == 3 {
				handler.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			handler.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (handler *Handler) playResp(cur *ChunkStream) error {
	handler.conn.SetRecorded()
	handler.conn.SetBegin()

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if err := handler.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if err := handler.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if err := handler.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if err := handler.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	return handler.conn.Flush()
}

func (handler *Handler) releaseStream(vs []interface{}) error {
	return nil
}

func (handler *Handler) fcPublish(vs []interface{}) error {
	return nil
}

func (handler *Handler) publishResp(cur *ChunkStream) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publishing."
	return handler.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event)
}
*/

func (handler *Handler) writeMsg(csid, streamID uint32, args ...interface{}) error {
	handler.bytesw.Reset()

	log.Println(fmt.Sprintf("%#v", args))

	return handler.conn.Flush()
}

/*
func (handler *Handler) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.DEL); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}

	return handler.conn.Write(&c)
}
*/
func (handler *Handler) Flush() error {
	return handler.conn.Flush()
}
