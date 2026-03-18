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
	"log/slog"
	"strings"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
)

func (n *Network) GetReachability() network.Reachability {
	hasPublicIP := false
	for _, addr := range n.Host.Addrs() {
		addrStr := addr.String()
		if !strings.Contains(addrStr, "127.0.0.1") &&
			!strings.Contains(addrStr, "192.168.") &&
			!strings.Contains(addrStr, "10.") &&
			!strings.Contains(addrStr, "p2p-circuit") {
			hasPublicIP = true
			break
		}
	}
	if hasPublicIP {
		return network.ReachabilityPublic
	}
	return network.ReachabilityPrivate
}

func isExternalAddr(addr string) bool {
	isLocal := strings.Contains(addr, "127.0.0.1") ||
		strings.Contains(addr, "192.168.") ||
		strings.Contains(addr, "10.") ||
		strings.Contains(addr, "p2p-circuit")
	return !isLocal
}

func (n *Network) Reachability() {
	status := n.GetReachability()

	switch status {
	case network.ReachabilityPublic:
		slog.Debug("Conectividad Óptima: Los peers pueden entrar directo.")
	case network.ReachabilityPrivate:
		slog.Debug("Conectividad Protegida: Dependemos del Pivote (Relay).")
	default:
		slog.Debug("Determinando estado de red.")
	}

	sub, _ := n.Host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	go func() {
		for e := range sub.Out() {
			evt := e.(event.EvtLocalReachabilityChanged)
			if evt.Reachability == network.ReachabilityPublic {
				slog.Debug("Ahora soy alcanzable desde el exterior.")
			}
		}
	}()
}
