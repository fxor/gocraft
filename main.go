package main

import (
	"io"
	"log"
	"net"
	_ "testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/zaluska/gocraft/protocol"
)

type ByteConn struct {
	net.Conn
}

func (b *ByteConn) ReadByte() (byte, error) {
	bs := make([]byte, 1)
	_, err := b.Read(bs)
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func main() {
	server, err := net.Listen("tcp", "localhost:25565")
	if err != nil {
		log.Panicln(err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		for {
			packet, err := protocol.ReadRawPacket(conn)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
			}

			log.Println(spew.Sdump(packet))
			p := protocol.ParseRawPacket(&packet)
			log.Println(spew.Sdump(p))
			switch p := p.(type) {
			case *protocol.HandshakePacket:
				if p.NextState == 1 {
					l := len(protocol.ExampleJson)
					l += protocol.SizeOfSerializedData(protocol.VarInt(len(protocol.ExampleJson)))
					l += protocol.SizeOfSerializedData(protocol.VarInt(0))
					_, err := protocol.Write(conn, protocol.VarInt(l))
					if err != nil {
						log.Println(err)
					}
					_, err = protocol.Write(conn, protocol.VarInt(0))
					if err != nil {
						log.Println(err)
					}
					_, err = protocol.Write(conn, protocol.ExampleJson)
					if err != nil {
						log.Println(err)
					}
				}
			case *protocol.PingPacket:
				pong := protocol.PongPacket{}
				pong.Payload = p.Payload

				err := protocol.WritePacket(conn, pong)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}

}
