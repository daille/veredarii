package components

import (
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

func ProtoHandler(s network.Stream) {
	defer s.Close()
	if allowedForProto["/miapp/proto/1.0.0"][s.Conn().RemotePeer()] {
		fmt.Println("POROTO")

		buf, _ := io.ReadAll(s)
		var msg Mensaje
		if err := proto.Unmarshal(buf, &msg); err != nil {
			log.Println("Error decoding proto:", err)
			return
		}
		log.Printf("Recibí: %+v\n", msg)
	} else {
		log.Println("Protocolo no autorizado")
	}
}
