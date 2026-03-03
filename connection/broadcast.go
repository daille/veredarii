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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

func (n *Network) InitBroadcast() {
	var err error
	n.NetworkPeerTopic, err = n.PS.Join("peers")
	if err != nil {
		log.Error("Error al unirse al topic:", err)
	}
	subPeer, err := n.NetworkPeerTopic.Subscribe()
	if err != nil {
		log.Error("Error al suscribirse al topic:", err)
	}
	go n.listenerPeers(subPeer)
}

func (n *Network) listenerPeers(subPeer *pubsub.Subscription) {
	log.Debug("Iniciando listener de PEERS")
	preKey := sha256.Sum256([]byte(n.SwarmKey))
	key := preKey[:]
	for {
		ctx := context.Background()
		msg, err := subPeer.Next(ctx)
		if err != nil {
			log.Error(global.Red("Error al recibir el mensaje:"), err)
			continue
		}
		descifrado, err := global.Decrypt(msg.Data, key)
		if err != nil {
			log.Error(global.Red("Error: No pude descifrar el mensaje o no estoy autorizado."), err)
			continue
		}
		fmt.Printf(global.Green("Mensaje recibido de %s: %s\n"), msg.ReceivedFrom, string(descifrado))

		var peerRequest PeerRequest
		err = json.Unmarshal(descifrado, &peerRequest)
		if err != nil {
			log.Error(global.Red("❌ Error deserializando solicitud:"), err)
			continue
		}
		fmt.Printf(global.Green("Solicitud deserializada:"), peerRequest.EntityName, peerRequest.PeerID)
		PID, err := peer.Decode(peerRequest.PeerID)
		if err != nil {
			log.Error("❌ Error decodificando llave publica:", err)
			continue
		}

		if peerRequest.PeerID == n.Host.ID().String() {
			log.Debug(global.Yellow("SELF"))
			continue
		}

		n.MutexSesiones.Lock()
		n.Peers[PID] = PeerType{
			Entity: peerRequest.EntityName,
		}
		n.MutexSesiones.Unlock()

		n.connection2RelatedPeer(peerRequest)

	}
}

func (n *Network) connection2RelatedPeer(peerRequest PeerRequest) {
	ctx := context.Background()
	PID, err := peer.Decode(peerRequest.PeerID)
	if err != nil {
		log.Error("❌ Error decodificando llave publica:", err)
		return
	}

	ip, _ := peerRequest.Address.ValueForProtocol(ma.P_IP4)
	correctAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, "9200"))

	info := peer.AddrInfo{
		ID:    PID,
		Addrs: []ma.Multiaddr{correctAddr},
	}
	// 3. Intentar la conexión
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := n.Host.Connect(ctx, info); err != nil {
		log.Error("❌ Error al conectar al peer: ", err)
		return
	}

	log.Info(fmt.Sprintf("✅ Conectado exitosamente a %s (%s)", peerRequest.EntityName, PID))
}
