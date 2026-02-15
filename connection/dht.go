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
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func (n *Network) initDHT() {
	ctx := context.Background()
	mode := dht.Mode(dht.ModeAuto)
	if len(n.Pivots) == 0 {
		mode = dht.Mode(dht.ModeServer)
	} else {
		mode = dht.Mode(dht.ModeClient)
	}

	var err error
	n.DHT, err = dht.New(ctx, n.Host, mode, dht.ProtocolPrefix("/mi-app-servicios"))
	if err != nil {
		log.Error(err)
		return
	}

	if err = n.DHT.Bootstrap(ctx); err != nil {
		log.Error(err)
		return
	}

	for _, topic := range n.Resources.API {
		go n.anunciarServicio(ctx, topic.Name)
	}
	for _, topic := range n.Resources.FILE {
		go n.anunciarServicio(ctx, topic.Name)
	}
	for _, topic := range n.Resources.DATASOURCE {
		go n.anunciarServicio(ctx, topic.Name)
	}
}

func (n *Network) anunciarServicio(ctx context.Context, serviceName string) {
	routingDiscovery := routing.NewRoutingDiscovery(n.DHT)
	util.Advertise(ctx, routingDiscovery, serviceName)
	fmt.Printf("Anunciando servicio '%s' en la DHT...\n", serviceName)
}

func (n *Network) BuscarServicio(ctx context.Context, serviceName string) peer.ID {
	routingDiscovery := routing.NewRoutingDiscovery(n.DHT)
	peerChan, err := routingDiscovery.FindPeers(ctx, serviceName, discovery.Limit(10))
	if err != nil {
		fmt.Printf("Error al buscar servicio: %v\n", err)
		return ""
	}

	fmt.Printf("Buscando proveedores de '%s'...\n", serviceName)

	for peerInfo := range peerChan {
		if peerInfo.ID == n.Host.ID() {
			continue
		}

		fmt.Printf("✨ Encontrado servicio en Peer: %s\n", peerInfo.ID)
		return peerInfo.ID
	}

	return ""
}
