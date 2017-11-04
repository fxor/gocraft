package protocol

import (
	"bytes"
	"io"
	"log"
)

const (
	Handshake = 0x00
)

type RawPacket struct {
	Id   int32
	data []byte
}

const (
	_                  = iota
	StateStatus VarInt = iota
	StateLogin
)

type HandshakePacket struct {
	ProtocolVersion VarInt
	ServerAddress   string
	ServerPort      uint16
	NextState       VarInt
}

type PingPacket struct {
	Payload int64
}

type PongPacket struct {
	Payload int64
}

var ExampleJson = `
{
    "version": {
        "name": "1.12.2",
        "protocol": 340
    },
    "players": {
        "max": 100,
        "online": 5,
		"sample": []
    },
    "description": {
        "text": "Hello world"
    }
}
`

type StatusResponse struct{}

func ReadRawPacket(r io.Reader) (p RawPacket, err error) {
	length, _, err := ReadVarInt(r)
	if err != nil {
		return
	}

	packetId, k, err := ReadVarInt(r)
	if err != nil {
		return
	}
	p.Id = packetId

	rest := int64(length - int32(k))

	var w bytes.Buffer

	io.CopyN(&w, r, rest)

	p.data = w.Bytes()

	return
}

func ParseRawPacket(rp *RawPacket) interface{} {
	br := bytes.NewReader(rp.data)
	switch rp.Id {
	case 0:
		if len(rp.data) == 0 {
			return nil
		}
		hp := new(HandshakePacket)
		_, err := Read(br, hp)
		if err != nil {
			log.Printf("Error while reading handshake packet: %s", err)
			return nil
		}
		return hp
	case 1:
		pp := new(PingPacket)
		_, err := Read(br, pp)
		if err != nil {
			log.Printf("Error while reading ping packet: %s", err)
			return nil
		}
		return pp
	default:
		return nil
	}
}
