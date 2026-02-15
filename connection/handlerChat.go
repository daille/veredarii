package connection

/*
MIT License

Copyright (c) 2026 Juan Carlos Daille

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Network) handleChatStream(s network.Stream) {
	remotePeer := s.Conn().RemotePeer()
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go n.readChats(rw, remotePeer)
}

func (n *Network) readChats(rw *bufio.ReadWriter, remotePeer peer.ID) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Printf("[Error] Peer %s desconectado\n", remotePeer)
			return
		}

		if str != "\n" {
			fmt.Printf("[Mensaje de %s]: %s", remotePeer, str)
		}
	}
}

func (n *Network) writeChats(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println("Error leyendo de stdin:", err)
			return
		}

		_, err = rw.WriteString(sendData)
		if err != nil {
			log.Println("Error escribiendo al peer:", err)
			return
		}

		err = rw.Flush()
		if err != nil {
			log.Println("Error haciendo flush:", err)
			return
		}
	}
}
