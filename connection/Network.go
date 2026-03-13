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
	global "Veredarii/global"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	dspebble "github.com/ipfs/go-ds-pebble"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	crdt "github.com/ipfs/go-ds-crdt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	ma "github.com/multiformats/go-multiaddr"
)

const Timeout = 5 * time.Minute

type Network struct {
	Name            string
	Host            host.Host
	Port            string
	SwarmKey        string
	JoinKey         string
	Pivots          []string
	Address         []string
	ExternalAddress string
	Resources       global.ResourcesType
	RemoteResources global.ResourcesType
	Topics          []global.TopicType
	Entities        []global.KVType
	//
	SesionesActivas    map[peer.ID]string
	MutexSesiones      sync.RWMutex
	MasterEntities     map[string]crypto.PubKey
	Peers              map[peer.ID]PeerType
	DHT                *dht.IpfsDHT
	RoutingDiscovery   *routing.RoutingDiscovery
	NetworkMemberTopic *pubsub.Topic
	NetworkPeerTopic   *pubsub.Topic
	RateLimiter        *RateManager
	DataStore          *crdt.Datastore
	PebbleStore        *dspebble.Datastore
	chPutHook          chan (global.KVType)
	PS                 *pubsub.PubSub
	PK                 crypto.PrivKey
	PSK                pnet.PSK
	Streams            map[string]map[string]StreamSession
	MutexStreams       sync.Mutex
}

type StreamSession struct {
	stream network.Stream
	timer  *time.Timer
}

type PeerType struct {
	ID     peer.ID
	PubKey crypto.PubKey
	Entity string
}

func NewNetwork(name string, port string, swarmKey string, pivots []string, address []string, externalAddress string, topics []global.TopicType, entities []global.KVType, resources global.ResourcesType, remoteResources global.ResourcesType) *Network {

	s := make(map[string]map[string]StreamSession)
	s[global.ProtocolAPIProxy] = make(map[string]StreamSession)
	s[global.ProtocolFileSystem] = make(map[string]StreamSession)
	s[global.ProtocolFileSystemStat] = make(map[string]StreamSession)

	N := Network{
		Name:            name,
		Port:            port,
		SwarmKey:        swarmKey,
		JoinKey:         ":",
		Pivots:          pivots,
		Address:         address,
		ExternalAddress: externalAddress,
		Topics:          topics,
		Entities:        entities,
		Resources:       resources,
		RemoteResources: remoteResources,
		SesionesActivas: make(map[peer.ID]string),
		MutexSesiones:   sync.RWMutex{},
		MasterEntities:  map[string]crypto.PubKey{},
		Peers:           map[peer.ID]PeerType{},
		RateLimiter:     &RateManager{entidadLimits: make(map[string]map[string]*rate.Limiter)},
		DataStore:       &crdt.Datastore{},
		PebbleStore:     &dspebble.Datastore{},
		Streams:         s,
		MutexStreams:    sync.Mutex{},
	}

	return &N
}

func (n *Network) Connect() {
	n.LoadConfig()
	//n.cargarWhitelist()
	miGater := &MiGater{peers: n.Peers}

	rmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale()))
	if err != nil {
		log.Error(err)
		return
	}

	cmgr, err := connmgr.NewConnManager(
		20,
		50,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		log.Error(err)
		return
	}

	if n.Pivots != nil {
		log.Info("Iniciando libp2p como pivote")
		n.Host, err = libp2p.New(
			libp2p.ListenAddrStrings(n.Address...),
			libp2p.Identity(n.PK),
			libp2p.ConnectionManager(cmgr),
			libp2p.ConnectionGater(miGater),
			libp2p.ResourceManager(rmgr),
			libp2p.Security(noise.ID, noise.New),
			libp2p.Security(tls.ID, tls.New),
			libp2p.DefaultMuxers,
			libp2p.EnableRelayService(),
			libp2p.EnableNATService(),
			libp2p.PrivateNetwork(n.PSK),
			libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
				if n.ExternalAddress == "" {
					return addrs
				}
				externalAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", n.ExternalAddress, n.Port))
				return append(addrs, externalAddr)
			}),
		)
	} else {
		log.Info("Iniciando libp2p sin pivote")
		n.Host, err = libp2p.New(
			libp2p.ListenAddrStrings(n.Address...),
			libp2p.Identity(n.PK),
			libp2p.ConnectionManager(cmgr),
			libp2p.ConnectionGater(miGater),
			libp2p.ResourceManager(rmgr),
			libp2p.Security(noise.ID, noise.New),
			libp2p.Security(tls.ID, tls.New),
			libp2p.DefaultMuxers,
			libp2p.NATPortMap(),
			libp2p.EnableHolePunching(),
			libp2p.EnableNATService(),
			libp2p.PrivateNetwork(n.PSK),
			libp2p.EnableRelay(),
			libp2p.ForceReachabilityPublic(),
		)
	}
	if err != nil {
		log.Error(err)
		return
	}
	defer n.Host.Close()

	// Protocolos de funcionamiento de la red
	n.Host.SetStreamHandler(protocol.ID(global.ProtocolAuth), n.handleAuthStream)
	n.Host.SetStreamHandler(protocol.ID(global.ProtocolJoin), n.handleJoinStream)
	// Protocolos de comunicación
	n.Host.SetStreamHandler(protocol.ID(global.ProtocolAPIProxy), n.handleAPIProxyStream)
	n.Host.SetStreamHandler(protocol.ID(global.ProtocolFileSystem), n.handleFileFetch)
	n.Host.SetStreamHandler(protocol.ID(global.ProtocolFileSystemStat), n.handleFileStat)

	fmt.Println("ID del peer:", n.Host.ID())
	peerID := n.Host.ID().String()
	for _, addr := range n.Host.Addrs() {
		fmt.Printf("👉 %s/p2p/%s\n", addr, peerID)
	}

	n.PS, err = pubsub.NewGossipSub(context.Background(), n.Host,
		pubsub.WithPeerExchange(true),
		pubsub.WithFloodPublish(true), // publica a TODOS los peers, no solo el mesh
	)
	if err != nil {
		log.Fatal(err)
	}

	go n.InitBroadcast()
	go n.MonitorConnections(n.PK)
	go n.initDHT()
	go n.FileSystem()

	go func() {
		n.DataStore, n.PebbleStore = n.NewDataStore()
		peers := n.Query(PEERS)
		for _, peerFounded := range peers {
			id, err := peer.Decode(peerFounded.Key)
			if err != nil {
				log.Error("❌ Error decodificando peer: ", peerFounded.Key, " : ", err)
				continue
			}
			n.Peers[id] = PeerType{ID: id, Entity: peerFounded.Name}
		}

		members := n.Query(MEMBERS)
		for _, memberFounded := range members {

			// 1. Convertir HEX string a []byte
			pubBytes, err := hex.DecodeString(memberFounded.Name)
			if err != nil {
				log.Error("❌ Error decodificando llave publica del miembro :", memberFounded.Key, " Error: ", err)
				log.Error("✅ Key {", memberFounded.Name, "}")
				return
			}

			// 2. Unmarshal usando la librería de libp2p
			pubKey, err := crypto.UnmarshalPublicKey(pubBytes)
			if err != nil {
				// Aquí es donde te daba el error "invalid wire-format"
				fmt.Printf("❌ Error unmarshal libp2p: %v\n", err)
				return
			}

			if err != nil {
				log.Error("❌ Error decodificando llave publica del miembro :", memberFounded.Key, " Error: ", err)
				log.Error("✅ Key {", memberFounded.Name, "}")

				continue
			}
			n.MasterEntities[memberFounded.Key] = pubKey
		}

		n.Host.Network().Notify(&network.NotifyBundle{
			ConnectedF: func(net network.Network, conn network.Conn) {
				log.Infof("Peer conectado: %s", conn.RemotePeer())
				go func() {
					time.Sleep(2 * time.Second)
					ctx := context.Background()
					n.DataStore.MarkDirty(ctx)
					if err := n.DataStore.Repair(ctx); err != nil {
						log.Errorf("Error en repair: %v", err)
					}
				}()
			},
		})
	}()

	// Cargando Peers
	for _, addr := range n.Host.Addrs() {
		fmt.Printf("Dirección activa: %s\n", addr)
	}
	fmt.Println("\nServidor esperando conexiones...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		n.DataStore.Close()
		time.Sleep(500 * time.Millisecond)
		n.PebbleStore.Close()
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	select {}
}

/*func (n *Network) getOrCreateStream(protocol string, targetID peer.AddrInfo) (network.Stream, error) {
	n.MutexStreams.Lock()
	defer n.MutexStreams.Unlock()

	if s, ok := n.Streams[protocol][targetID.ID]; ok {
		log.Debug("Stream ya existe para: ", targetID.ID, " Protocolo: ", protocol)
		return s, nil
	}

	log.Debug("Abriendo Stream para: ", targetID.ID, " Protocolo: ", protocol)
	s, err := n.Host.NewStream(context.Background(), targetID.ID, global.ProtocolAPIProxy)
	if err != nil {
		return nil, err
	}

	n.Streams[protocol][targetID.ID] = s
	return s, nil
}*/

func (n *Network) getStream(proto string, service string) (network.Stream, error) {
	n.MutexStreams.Lock()
	defer n.MutexStreams.Unlock()

	if sess, ok := n.Streams[proto][service]; ok {
		if !sess.timer.Stop() {
			select {
			case <-sess.timer.C:
			default:
			}
		}
		sess.timer.Reset(Timeout)
		return sess.stream, nil
	}

	target := n.BuscarServicio(context.Background(), service)
	if target == nil {
		return nil, errors.New("Servicio no encontrado")
	}

	log.Debug("Abriendo Stream para: ", target.ID, " Protocolo: ", proto, " servicio: ", service)
	s, err := n.Host.NewStream(context.Background(), target.ID, protocol.ID(proto))
	if err != nil {
		return nil, err
	}

	sess := StreamSession{
		stream: s,
		timer:  time.NewTimer(Timeout),
	}

	go func(service string, st network.Stream, t *time.Timer) {
		<-t.C
		n.MutexStreams.Lock()
		log.Printf("⏳ Timeout: Cerrando stream inactivo con %s", service)
		st.Close()
		delete(n.Streams[proto], service)
		n.MutexStreams.Unlock()
	}(service, s, sess.timer)

	n.Streams[proto][service] = sess
	return s, nil
}

func (n *Network) MonitorConnections(priv crypto.PrivKey) {
	for {
		peerCount := len(n.Host.Network().Peers())

		if peerCount == 0 && len(n.Pivots) > 0 {
			log.Warn("¡Nodo aislado! Reconectando a los pivotes...")
			for _, addr := range n.Pivots {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				info, _ := peer.AddrInfoFromString(addr)
				if err := n.Host.Connect(ctx, *info); err != nil {
					log.Error("Fallo reconexión al pivote:", err)
				} else {
					log.Info("Conexión exitosa al pivote:", addr)
					n.Authenticar(ctx, priv, info.ID)
				}
				cancel()
			}
		}
		time.Sleep(30 * time.Second)
	}
}
