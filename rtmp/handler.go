package rtmp

import (
	"bytes"
)

/*
Types of publish commands
*/
var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
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
