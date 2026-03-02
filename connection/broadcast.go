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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

func (n *Network) InitBroadcast() {
	ctx := context.Background()
	ps, err := pubsub.NewGossipSub(ctx, n.Host)
	if err != nil {
		log.Error("Error al crear el pubsub:", err)
	}
	n.NetworkMemberTopic, err = ps.Join("members")
	if err != nil {
		log.Error("Error al unirse al topic:", err)
	}
	sub, err := n.NetworkMemberTopic.Subscribe()
	if err != nil {
		log.Error("Error al suscribirse al topic:", err)
	}
	go n.listenerMembers(sub)

	n.NetworkPeerTopic, err = ps.Join("peers")
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
	preKey := sha256.Sum256([]byte(n.SwarmKey))
	key := preKey[:]
	for {
		ctx := context.Background()
		msg, err := subPeer.Next(ctx)
		if err != nil {
			log.Error("Error al recibir el mensaje:", err)
			continue
		}
		descifrado, err := global.Decrypt(msg.Data, key)
		if err != nil {
			log.Error("Error: No pude descifrar el mensaje o no estoy autorizado. ", err)
			continue
		}
		fmt.Printf("Mensaje recibido de %s: %s\n", msg.ReceivedFrom, string(descifrado))

		var peerRequest PeerRequest
		err = json.Unmarshal(descifrado, &peerRequest)
		if err != nil {
			log.Error("❌ Error deserializando solicitud:", err)
			continue
		}
		log.Info("Solicitud deserializada:", peerRequest.EntityName, peerRequest.PeerID)
		PID, err := peer.Decode(peerRequest.PeerID)
		if err != nil {
			log.Error("❌ Error decodificando llave publica:", err)
			continue
		}

		n.PutCRDT("peers", peerRequest.PeerID, peerRequest.EntityName)

		n.MutexSesiones.Lock()
		n.Peers[PID] = PeerType{
			Entity: peerRequest.EntityName,
		}
		n.MutexSesiones.Unlock()
	}
}

func (n *Network) listenerMembers(sub *pubsub.Subscription) {

	preKey := sha256.Sum256([]byte(n.SwarmKey))
	key := preKey[:]
	for {
		ctx := context.Background()
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Error("Error al recibir el mensaje:", err)
			continue
		}
		descifrado, err := global.Decrypt(msg.Data, key)
		if err != nil {
			log.Error("Error: No pude descifrar el mensaje o no estoy autorizado. ", err)
			continue
		}
		fmt.Printf("Mensaje recibido de %s: %s\n", msg.ReceivedFrom, string(descifrado))

		var joinRequest JoinRequest
		err = json.Unmarshal(descifrado, &joinRequest)
		if err != nil {
			log.Error("❌ Error deserializando solicitud:", err)
			continue
		}
		log.Info("Solicitud deserializada:", joinRequest.EntityName)
		pubKey, err := global.ParsePubKeyRecibida(joinRequest.PublicKey)
		if err != nil {
			log.Error("❌ Error decodificando llave publica:", err)
			continue
		}

		// @TODO Validar que el miembro enviado sea válido segun los estándares propios

		// Ingresar entidad a la DB
		n.PutCRDT("members", joinRequest.EntityName, joinRequest.PublicKey)

		// Ingresar entidad a la lista de entidades activas
		n.MutexSesiones.Lock()
		n.MasterEntities[joinRequest.EntityName] = pubKey
		n.MutexSesiones.Unlock()

		// Escribe entidad en el archivo de configuracion (obsoleto)
		data, err := base64.StdEncoding.DecodeString(joinRequest.PublicKey)
		if err != nil {
			log.Fatal("Error decodificando Base64:", err)
		}
		configuration.CM.AddEntity(n.Name, global.KVType{
			Name: joinRequest.EntityName,
			Key:  hex.EncodeToString(data),
		})

	}
}
