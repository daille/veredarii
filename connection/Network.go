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
	"Veredarii/configuration"
	global "Veredarii/global"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/record"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

type Network struct {
	Name            string
	Host            host.Host
	Port            string
	SwarmKey        string
	Pivots          []string
	Address         []string
	Resources       global.ResourcesType
	RemoteResources global.ResourcesType
	Topics          []global.TopicType
	Entities        []global.KVType
	//
	SesionesActivas map[peer.ID]string
	MutexSesiones   sync.RWMutex
	MasterEntities  map[string]crypto.PubKey
	Whitelist       map[peer.ID]bool
	DHT             *dht.IpfsDHT
}

func NewNetwork(name string, port string, swarmKey string, pivots []string, address []string, topics []global.TopicType, entities []global.KVType, resources global.ResourcesType, remoteResources global.ResourcesType) *Network {
	N := Network{
		Name:            name,
		Port:            port,
		SwarmKey:        swarmKey,
		Pivots:          pivots,
		Address:         address,
		Topics:          topics,
		Entities:        entities,
		Resources:       resources,
		RemoteResources: remoteResources,
		SesionesActivas: make(map[peer.ID]string),
		MutexSesiones:   sync.RWMutex{},
		MasterEntities:  map[string]crypto.PubKey{},
		Whitelist:       map[peer.ID]bool{},
	}

	return &N
}

func (n *Network) Connect() {
	psk, priv := n.LoadConfig()
	n.cargarWhitelist()
	miGater := &MiGater{whitelist: n.Whitelist}

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
		panic(err)
	}

	n.Host, err = libp2p.New(
		libp2p.ListenAddrStrings(n.Address...),
		libp2p.Identity(priv),
		libp2p.ConnectionManager(cmgr),
		libp2p.ConnectionGater(miGater),
		libp2p.ResourceManager(rmgr),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(tls.ID, tls.New),
		libp2p.DefaultMuxers,
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelayService(),
		libp2p.EnableNATService(),
		libp2p.PrivateNetwork(psk),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Error(err)
		return
	}
	n.Host.Network().Notify(&networkNotifiee{n: n})
	defer n.Host.Close()
	fmt.Println("ID del peer:", n.Host.ID())
	peerID := n.Host.ID().String()
	for _, addr := range n.Host.Addrs() {
		fmt.Printf("👉 %s/p2p/%s\n", addr, peerID)
	}
	go n.MonitorConnections(priv)
	n.initDHT()

	n.Host.SetStreamHandler(global.ProtocolAuth, n.handleAuthStream)
	n.Host.SetStreamHandler(global.ProtocolChat, n.handleChatStream)
	n.Host.SetStreamHandler(global.ProtocolAPIProxy, n.handleAPIProxyStream)
	n.Host.SetStreamHandler(global.ProtocolFileSystem, n.handleFileFetch)
	n.Host.SetStreamHandler(global.ProtocolFileSystemStat, n.handleFileStat)
	n.Host.SetStreamHandler(global.ProtocolQuery, n.HandleSearch)
	go n.FileSystem()

	fmt.Println("\nServidor esperando conexiones...")
	select {}
}

func (n *Network) MonitorConnections(priv crypto.PrivKey) {
	for {
		peerCount := len(n.Host.Network().Peers())

		if peerCount == 0 {
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

func (n *Network) InitBroadcast() {
	ctx := context.Background()
	ps, err := pubsub.NewGossipSub(ctx, n.Host)
	if err != nil {
		panic(err)
	}
	topic, err := ps.Join("anuncios-globales")
	sub, err := topic.Subscribe()

	go func() {
		preKey := sha256.Sum256([]byte(n.SwarmKey))
		key := preKey[:]
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			descifrado, err := global.Decrypt(msg.Data, key)
			if err != nil {
				fmt.Println("Error: No pude descifrar el mensaje o no estoy autorizado. ", err)
				continue
			}
			fmt.Printf("Mensaje recibido de %s: %s\n", msg.ReceivedFrom, string(descifrado))
		}
	}()

}

func (n *Network) LoadConfig() (pnet.PSK, crypto.PrivKey) {
	var err error

	record.RegisterType(&EntidadRecord{})
	for _, entity := range n.Entities {
		pubKeyRaw, err := hex.DecodeString(entity.Key)
		if err != nil {
			log.Fatalf("Error al decodificar hexadecimal: %v", err)
		}
		pkb, err := crypto.UnmarshalPublicKey(pubKeyRaw)
		if err != nil {
			log.Fatalf("Error al procesar llave pública: %v", err)
		}
		n.MasterEntities[entity.Name] = pkb
	}

	priv, err := n.obtenerIdentidad(configuration.CM.GetConfig().Identity.PrivKeyFile)
	if err != nil {
		log.Fatal("Error con la identidad:", err)
	}

	keyStr := n.SwarmKey
	psk, err := global.DecodeV1PSK(keyStr)
	if err != nil {
		log.Fatal("Error cargando PSK:", err)
	}

	return psk, priv
}
