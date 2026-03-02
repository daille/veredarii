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
	"strings"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
)

func (n *Network) GetReachability() network.Reachability {
	// Analizamos las direcciones actuales del host
	hasPublicIP := false
	for _, addr := range n.Host.Addrs() {
		addrStr := addr.String()
		// Si tiene una IP que no es privada ni loopback ni relay
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

// Helper simple para filtrar IPs
func isExternalAddr(addr string) bool {
	// Si contiene p2p-circuit (Relay) o IPs privadas, no es "puro" público
	isLocal := strings.Contains(addr, "127.0.0.1") ||
		strings.Contains(addr, "192.168.") ||
		strings.Contains(addr, "10.") ||
		strings.Contains(addr, "p2p-circuit")
	return !isLocal
}

func (n *Network) Reachability() {
	fmt.Println("reachability")
	status := n.GetReachability()

	switch status {
	case network.ReachabilityPublic:
		fmt.Println("🚀 Conectividad Óptima: Los peers pueden entrar directo.")
		// Aquí podrías, por ejemplo, aumentar el límite de conexiones del Connection Manager

	case network.ReachabilityPrivate:
		fmt.Println("🛡️ Conectividad Protegida: Dependemos del Pivote (Relay).")
		// Aquí podrías avisar al usuario que la latencia será un poco mayor

	default:
		fmt.Println("⏳ Determinando estado de red...")
	}

	sub, _ := n.Host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	go func() {
		for e := range sub.Out() {
			evt := e.(event.EvtLocalReachabilityChanged)
			if evt.Reachability == network.ReachabilityPublic {
				fmt.Println("🚀 ¡Evento: Ahora soy alcanzable desde el exterior!")
			}
		}
	}()
}
