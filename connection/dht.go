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
	"time"

	"log/slog"

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
		slog.Error("Error al crear el DHT", "error", err)
		return
	}

	if err = n.DHT.Bootstrap(ctx); err != nil {
		slog.Error("Error al hacer bootstrap del DHT", "error", err)
		return
	}

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()
	done := make(chan bool)
	n.publishServices()

	go func() {
		<-n.DHT.RefreshRoutingTable()
		slog.Debug("Tabla de ruteo refrescada")
	}()

	for {
		select {
		case _ = <-ticker.C:
			n.publishServices()
		case <-done:
			slog.Info("Deteniendo ticker...")
			return
		}
	}
}

func (n *Network) publishServices() {
	slog.Info("Anunciando servicios en la DHT...")
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
	slog.Info("Anunciando servicio en la DHT", "service", serviceName)
}

func (n *Network) BuscarServicio(ctx context.Context, serviceName string) *peer.AddrInfo {
	peerChan, err := n.RoutingDiscovery.FindPeers(ctx, serviceName, discovery.Limit(10))
	if err != nil {
		slog.Error("Error al buscar servicio", "error", err)
		return nil
	}
	slog.Info("Buscando proveedores de servicio", "service", serviceName)

	for peerInfo := range peerChan {
		if peerInfo.ID == n.Host.ID() {
			continue
		}
		slog.Info("Encontrado servicio en Peer", "peer", peerInfo.ID)
		return &peerInfo
	}
	return nil
}
