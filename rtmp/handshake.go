package rtmp

import (
	"fmt"
	"io"
)

/*
+-------------+                           +-------------+
|    Client   |       TCP/IP Network      |    Server   |
+-------------+            |              +-------------+
      |                    |                     |
 Uninitialized             |               Uninitialized
      |          C0        |                     |
      |------------------->|         C0          |
      |                    |-------------------->|
      |          C1        |                     |
      |------------------->|         S0          |
      |                    |<--------------------|
      |                    |         S1          |
 Version sent              |<--------------------|
      |          S0        |                     |
      |<-------------------|                     |
      |          S1        |                     |
      |<-------------------|                Version sent
      |                    |         C1          |
      |                    |-------------------->|
      |          C2        |                     |
      |------------------->|         S2          |
      |                    |<--------------------|
   Ack sent                |                  Ack Sent
      |          S2        |                     |
      |<-------------------|                     |
      |                    |         C2          |
      |                    |-------------------->|
 Handshake Done            |               Handshake Done
      |                    |                     |
*/

var GenuineFMSKey = [...]uint8{
	0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
	0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x20, 0x30, 0x30, 0x31, /* Genuine Adobe Flash Media Server 001 */

	0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8, 0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
	0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab, 0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
}

var GenuineFPKey = [...]uint8{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20, 0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79, 0x65, 0x72, 0x20, 0x30, 0x30, 0x31, /* Genuine Adobe Flash Player 001 */
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB, 0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xae,
}

type Handshake struct{}

func HandshakeServer(conn *Connection) (err error) {
	var clientData [1 + 1536*2]byte
	var serverData [1 + 1536*2]byte

	C0 := clientData[:1]
	C1 := clientData[1 : 1536+1]
	C0C1 := clientData[:1536+1]
	C2 := clientData[1536+1:]

	S0 := serverData[:1]
	S1 := serverData[1 : 1536+1]
	S0S1 := serverData[:1536+1]
	S2 := serverData[1536+1:]

	if _, err = io.ReadFull(conn.rw, C0C1); err != nil {
		return
	}

	/*
	   In C0, this field identifies the RTMP version requested by the client.
	   In S0, this field identifies the RTMP version selected by the server.
	   The version defined by this specification is 3.
	*/
	if C0[0] != 3 {
		err = fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
		return
	}

	S0[0] = 3

	//cliTs := bitops.U32BE(C1[0:4])

	copy(S1, C2)
	copy(S2, C1)

	if _, err = conn.rw.Write(S0S1); err != nil {
		return
	}

	if _, err = conn.rw.Write(S2); err != nil {
		return
	}

	if err = conn.rw.Flush(); err != nil {
		return
	}

	if _, err = io.ReadFull(conn.rw, C2); err != nil {
		return
	}

	fmt.Println(C2)

	return
}
