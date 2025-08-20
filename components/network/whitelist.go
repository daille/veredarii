package components

/**
 *    Veredarii, software for interoperability.
 *    This file is part of Veredarii.
 *
 *    @author jcDaille
 *
 *
 *    MIT License
 *
 * Copyright (c) 2025 JC Daille
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/control"
	network "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

// whitelistConnGater implementa connGater filtrando solo peers en whitelist
type whitelistConnGater struct {
	allowed map[peer.ID]struct{}
	//allowedProtocols map[peer.ID]map[string]bool // peerID -> protocolo -> permitido
	//mu               sync.RWMutex
}

func newWhitelistConnGater(ids []peer.ID) *whitelistConnGater {
	m := make(map[peer.ID]struct{})
	for _, id := range ids {
		m[id] = struct{}{}
	}

	log.Debug("WhiteList:", m)
	return &whitelistConnGater{allowed: m}
}

// Interfaz connGater

func (w *whitelistConnGater) InterceptPeerDial(p peer.ID) bool {
	// Permitir dial solo a peers en whitelist
	fmt.Println(1)
	_, ok := w.allowed[p]
	return ok
}

func (w *whitelistConnGater) InterceptAddrDial(p peer.ID, addr multiaddr.Multiaddr) bool {
	// Similar: filtrar por dirección de dial si quieres
	fmt.Println(2)
	_, ok := w.allowed[p]
	return ok
}

func (w *whitelistConnGater) InterceptAccept(conn network.ConnMultiaddrs) bool {
	return true
}

func (w *whitelistConnGater) InterceptSecured(direction network.Direction, p peer.ID, conn network.ConnMultiaddrs) bool {
	// Después de cifrar, confirmar whitelist
	fmt.Println(4)
	_, ok := w.allowed[p]
	return ok
}

func (w *whitelistConnGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	fmt.Println(5)
	// Después de upgrade final
	_, ok := w.allowed[conn.RemotePeer()]
	return ok, 0
}

// ----------------------------------------------------------------------------------
