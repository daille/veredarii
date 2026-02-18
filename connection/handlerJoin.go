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
	"io"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type JoinRequest struct {
	EntityName  string `json:"entity"`
	InviterName string `json:"inviter"`
	Network     string `json:"network"`
	PublicKey   string `json:"pubkey"`
	Invitation  string `json:"invitation"`
}

var globalUnionLimiter = rate.NewLimiter(rate.Limit(0.2), 3)

func (n *Network) handleJoinStream(s network.Stream) {
	if !globalUnionLimiter.Allow() {
		log.Error(" [!] TPS Global excedido. Rechazando conexión de: ", s.Conn().RemotePeer())
		s.Reset()
		return
	}
	defer s.Close()
	log.Info(" [+] Procesando solicitud legítima de: ", s.Conn().RemotePeer())

	limitReader := io.LimitReader(s, 4096)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		log.Error("❌ Error leyendo solicitud:", err)
		return
	}

	log.Info("Solicitud recibida:", string(body))

	var joinRequest JoinRequest
	err = json.Unmarshal(body, &joinRequest)
	if err != nil {
		log.Error("❌ Error deserializando solicitud:", err)
		return
	}
	log.Info("Solicitud deserializada:", joinRequest.EntityName)

	log.Debug(fmt.Sprintf("Cargando entidad '%s' con llave pública '%s'", joinRequest.EntityName, joinRequest.PublicKey))
	pubKey, err := global.ParsePubKeyRecibida(joinRequest.PublicKey)
	if err != nil {
		log.Error("❌ Error decodificando llave publica:", err)
		return
	}

	passphrase := "mi frase super secreta para la red"
	salt := "mi-red-p2p-secreta-unique-salt"
	key := global.GenerarLlaveDesdeFrase(passphrase, salt)

	invitation := global.DecipherInvitation(joinRequest.Invitation, joinRequest.InviterName, key)
	log.Info("Invitación descifrada:", invitation)
	invitationSplit := strings.Split(invitation, ";")

	if invitationSplit[0] != joinRequest.InviterName {
		log.Debug(fmt.Sprintf("Invitación invitador inválida: %s != %s", invitationSplit[0], joinRequest.InviterName))
		log.Error("❌ Error descifrando invitación")
		return
	}
	if invitationSplit[3] != joinRequest.Network {
		log.Debug(fmt.Sprintf("Invitación Red inválida: %s != %s", invitationSplit[3], joinRequest.Network))
		log.Error("❌ Error descifrando invitación")
		return
	}
	if invitationSplit[2] != joinRequest.EntityName {
		log.Debug(fmt.Sprintf("Invitación invitado inválida: %s != %s", invitationSplit[2], joinRequest.EntityName))
		log.Error("❌ Error descifrando invitación")
		return
	}

	expiracion, err := time.Parse(time.RFC3339, invitationSplit[4])
	if err != nil {
		log.Error("❌ Error al parsear la fecha:", err)
		return
	}
	if time.Now().After(expiracion) {
		log.Error("❌ La invitación ha expirado.")
		return
	} else {
		log.Info("✅ La invitación aún es válida.")
		faltante := time.Until(expiracion)
		log.Info(fmt.Sprintf("Expira en: %v\n", faltante.Round(time.Minute)))

		n.MutexSesiones.Lock()
		n.MasterEntities[invitationSplit[2]] = pubKey
		n.MutexSesiones.Unlock()

		ctx := context.Background()
		preKey := sha256.Sum256([]byte(n.SwarmKey))
		key := preKey[:]
		body, err := global.Encrypt(body, key)
		if err != nil {
			log.Error("❌ Error al cifrar la solicitud:", err)
			return
		}
		n.NetworkMemberTopic.Publish(ctx, []byte(body))
	}
}
