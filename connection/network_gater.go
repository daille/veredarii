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
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type MiGater struct {
	whitelist map[peer.ID]bool
}

func (g *MiGater) InterceptPeerDial(p peer.ID) bool {
	return true //g.whitelist[p]
}

func (g *MiGater) InterceptAddrDial(p peer.ID, m multiaddr.Multiaddr) bool {
	return true //g.whitelist[p]
}

func (g *MiGater) InterceptAccept(n network.ConnMultiaddrs) bool {
	return true
}

func (g *MiGater) InterceptSecured(dir network.Direction, p peer.ID, n network.ConnMultiaddrs) bool {
	return true
}

func (g *MiGater) InterceptUpgraded(n network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
