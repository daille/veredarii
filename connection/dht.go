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
	"time"

	log "github.com/sirupsen/logrus"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const protocolPrefix = "/KoNfE4RRSjvPfYrq9"

func (n *Network) initDHT() {
	ctx := context.Background()
	mode := dht.Mode(dht.ModeServer)

	var err error
	n.DHT, err = dht.New(ctx, n.Host, mode, dht.ProtocolPrefix(protocolPrefix))
	if err != nil {
		log.Error(err)
		return
	}

	if err = n.DHT.Bootstrap(ctx); err != nil {
		log.Error(err)
		return
	}

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()
	done := make(chan bool)
	n.publishServices()

	go func() {
		<-n.DHT.RefreshRoutingTable()
		log.Debug("✅ Tabla de ruteo refrescada")
	}()

	for {
		select {
		case _ = <-ticker.C:
			n.publishServices()
		case <-done:
			log.Info("Stopping ticker...")
			return
		}
	}
}

func (n *Network) publishServices() {
	log.Info("Anunciando servicios en la DHT...")
	ctx := context.Background()
	n.RoutingDiscovery = routing.NewRoutingDiscovery(n.DHT)
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
	util.Advertise(ctx, n.RoutingDiscovery, serviceName)
	fmt.Printf("Anunciando servicio '%s' en la DHT...\n", serviceName)
}

func (n *Network) BuscarServicio(ctx context.Context, serviceName string) *peer.AddrInfo {

	fmt.Println(len(n.DHT.RoutingTable().ListPeers()))
	peerChan, err := n.RoutingDiscovery.FindPeers(ctx, serviceName, discovery.Limit(10))
	if err != nil {
		fmt.Printf("Error al buscar servicio: %v\n", err)
		return nil
	}

	fmt.Printf("Buscando proveedores de '%s'...\n", serviceName)

	for peerInfo := range peerChan {
		if peerInfo.ID == n.Host.ID() {
			continue
		}

		fmt.Printf("✨ Encontrado servicio en Peer: %s\n", peerInfo.ID)
		return &peerInfo
	}

	return nil
}
