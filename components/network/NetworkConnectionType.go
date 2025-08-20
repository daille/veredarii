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
	"Veredarii/general"
	"Veredarii/util"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	basicconnmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	log "github.com/sirupsen/logrus"
)

/**
 *    Interop software for interoperability
 *
 *    @author jcDaille
 *
 *    This file is part of Interop.
 *
 *    Interop is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License.
 *
 *    Interop is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with Foobar.  If not, see <https://www.gnu.org/licenses/>.
 */
type NetworkConnectionType struct {
	Config       general.NetworkType
	NetworkPeers []peer.AddrInfo
	Host         host.Host
	DHT          *dht.IpfsDHT
	CID          cid.Cid
}

func NewNetworkConnection(networkConfig general.NetworkType) *NetworkConnectionType {
	ncl := &NetworkConnectionType{
		Config: networkConfig,
	}
	return ncl
}

func (NC *NetworkConnectionType) CreateHost(priv crypto.PrivKey) (ok bool) {
	var low, high = 400, 800
	if psk, ok := util.LoadSwarmKey(NC.Config.KeyNetwork); ok {
		listen := []string{
			"/ip4/0.0.0.0/tcp/" + NC.Config.Port,
			"/ip6/::/tcp/" + NC.Config.Port,
		}
		var maddrs []ma.Multiaddr
		for _, s := range listen {
			m, err := ma.NewMultiaddr(s)
			if err != nil {
				log.Error(err)
				return false
			}
			maddrs = append(maddrs, m)
		}

		gater := newWhitelistConnGater(NC.loadWhiteListPeer(NC.Config))
		cm, err := basicconnmgr.NewConnManager(low, high)
		if err != nil {
			log.Error(err)
			return false
		}

		NC.Host, err = libp2p.New(
			libp2p.Identity(priv),
			libp2p.ListenAddrs(maddrs...),
			libp2p.PrivateNetwork(psk),
			libp2p.Security(tls.ID, tls.New),
			libp2p.Security(noise.ID, noise.New),
			libp2p.ConnectionGater(gater),
			libp2p.ConnectionManager(cm),
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			//libp2p.Transport(tcp.NewTCPTransport),
			libp2p.EnableNATService(),
			libp2p.EnableRelay(),
			libp2p.Ping(false),
		)
		if err != nil {
			log.Error(err)
			return false
		} else {
			NC.MakeDHT(ctx)
			if len(NC.Config.Pivots) > 0 {
				log.Debug(util.Yellow("Connecting to the pivot"))
				NC.Connect(NC.Config.Pivots[0])
				NC.GetDHTPeers()
			} else {
				// i'm the pivot
				log.Debug(util.Yellow("i'am the Pivot"))
			}
		}

	} else {
		log.Error("No swarm key")
		return false
	}

	return true
}

func (NC *NetworkConnectionType) Connect(peerAddrStr string) {
	maddr, err := ma.NewMultiaddr(peerAddrStr)
	if err != nil {
		log.Debug("multiaddr inválida: ", err)
		return
	}
	ai, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Debug("no pude extraer AddrInfo: ", err)
		return
	}

	// Conectar
	NC.Host.Peerstore().AddAddrs(ai.ID, ai.Addrs, time.Hour)
	if err := NC.Host.Connect(ctx, *ai); err != nil {
		log.Debug(fmt.Sprintf("connect falló: %w", err))
		return
	}
}

func (NC *NetworkConnectionType) MakeDHT(ctx context.Context) {
	var err error
	NC.DHT, err = dht.New(
		ctx,
		NC.Host,
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix("/mi-red-privada/kad"),
		dht.NamespacedValidator("app", allowAllValidator{}),
	)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := "/private/peers/myapp"
	mh, _ := multihash.Sum([]byte(keyStr), multihash.SHA2_256, -1)
	NC.CID = cid.NewCidV1(cid.Raw, mh)

	info := InstitutionInfo{
		Name:     "Institución ACME",
		Location: "Santiago, Chile",
		Roles:    []string{"API", "Validator", "Storage"},
	}
	data, _ := json.Marshal(info)
	NC.DHT.Provide(ctx, NC.CID, true) // true = anounce ourselves
	fmt.Println("->", NC.DHT.PutValue(ctx, "/app/institucion/ACME", data))
}

func (NC *NetworkConnectionType) loadWhiteListPeer(network general.NetworkType) (allowedPeers []peer.ID) {
	for _, peerId := range network.Whitelist {
		pid, err := peer.Decode(peerId)
		if err != nil {
			log.Error(general.T("error_whitelist"), ":", err)
			return nil
		}
		allowedPeers = append(allowedPeers, pid)
	}
	return
}

func (NC *NetworkConnectionType) PrintAddress() {
	log.Info(util.Yellow("ID: "), util.Teal(NC.Host.ID()))
	log.Debug(general.T("address_to_share"))
	for _, a := range NC.Host.Addrs() {
		log.Debug(fmt.Sprintf("    %s/p2p/%s", a, NC.Host.ID()))
	}
}

func (NC *NetworkConnectionType) GetDHTPeers() {
	var err error
	NC.NetworkPeers, err = NC.DHT.FindProviders(ctx, NC.CID)
	if err != nil {
		log.Error(err)
	}
	for i, p := range NC.NetworkPeers {
		fmt.Println("# [", i, "] ", p.ID, p.Addrs)
	}
	return
}

func (NC *NetworkConnectionType) FindPeer(search string) string {
	NC.GetDHTPeers()
	for _, p := range NC.NetworkPeers {
		if p.ID.String() == search {
			return p.Addrs[0].String()
		}
	}
	return ""
}

func (NC *NetworkConnectionType) SendMessage(payload []byte, to string) {
	peerAddrStr := NC.FindPeer(to) + "/p2p/" + to
	maddr, err := ma.NewMultiaddr(peerAddrStr)
	if err != nil {
		log.Debug("multiaddr inválida: ", err)
		return
	}
	ai, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Debug("no pude extraer AddrInfo: ", err)
		return
	}

	// Conectar
	NC.Host.Peerstore().AddAddrs(ai.ID, ai.Addrs, time.Hour)
	if err := NC.Host.Connect(ctx, *ai); err != nil {
		log.Debug(fmt.Sprintf("connect falló: %w", err))
		return
	}

	s, err := NC.Host.NewStream(ctx, ai.ID, "/data/1.0.0")
	if err != nil {
		log.Debug(fmt.Sprintf("no pude abrir stream: %w", err))
		return
	}
	defer s.Close()

	// Enviar payload (len+data)
	if err := util.WriteMsg(s, payload); err != nil {
		log.Debug("no pude enviar: %w", err)
	}

	// Leer respuesta (echo u otro)
	resp, err := util.ReadMsg(s)
	if err != nil {
		log.Debug("no pude leer respuesta: %w", err)
	}

	log.Debug("Respuesta (", len(resp), "):")
	log.Debug(string(resp))
}
