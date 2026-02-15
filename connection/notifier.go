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
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

type networkNotifiee struct {
	n *Network
}

func (nn *networkNotifiee) Listen(net network.Network, addr multiaddr.Multiaddr)      {}
func (nn *networkNotifiee) ListenClose(net network.Network, addr multiaddr.Multiaddr) {}
func (nn *networkNotifiee) Connected(net network.Network, c network.Conn)             {}
func (nn *networkNotifiee) OpenedStream(net network.Network, s network.Stream)        {}
func (nn *networkNotifiee) ClosedStream(net network.Network, s network.Stream)        {}

func (nn *networkNotifiee) Disconnected(net network.Network, c network.Conn) {
	peerID := c.RemotePeer()

	nn.n.MutexSesiones.Lock()
	RBAC.MutexSesiones.Lock()
	defer nn.n.MutexSesiones.Unlock()
	defer RBAC.MutexSesiones.Unlock()

	if _, existe := nn.n.SesionesActivas[peerID]; existe {
		delete(nn.n.SesionesActivas, peerID)
		delete(RBAC.PeerEntity, peerID.String())
		fmt.Printf("🧹 Sesión eliminada: el peer %s se ha desconectado\n", peerID)
	}
}
