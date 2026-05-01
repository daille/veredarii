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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"log/slog"

	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
	"golang.org/x/time/rate"
)

type JoinRequest struct {
	EntityName  string `json:"entity"`
	InviterName string `json:"inviter"`
	Network     string `json:"network"`
	PublicKey   string `json:"pubkey"`
	Invitation  string `json:"invitation"`
}

type PeerRequest struct {
	EntityName string       `json:"entity"`
	PeerID     string       `json:"peerid"`
	PublicKey  string       `json:"pubkey"`
	Address    ma.Multiaddr `json:"address"`
}

var globalUnionLimiter = rate.NewLimiter(rate.Limit(0.2), 3)

func (n *Network) handleJoinStream(s network.Stream) {
	if !globalUnionLimiter.Allow() {
		slog.Error("TPS Global excedido. Rechazando conexión de: ", "peer", s.Conn().RemotePeer())
		s.Reset()
		return
	}
	defer s.Close()
	slog.Info("Procesando solicitud legítima de: ", "peer", s.Conn().RemotePeer())

	limitReader := io.LimitReader(s, 4096)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		slog.Error("Error leyendo solicitud:", "error", err)
		return
	}
	slog.Debug("Solicitud recibida:", "body", string(body))

	var joinRequest JoinRequest
	err = json.Unmarshal(body, &joinRequest)
	if err != nil {
		slog.Error("Error deserializando solicitud:", "error", err)
		return
	}
	slog.Info("Solicitud deserializada:", "entity", joinRequest.EntityName)

	slog.Debug(fmt.Sprintf("Cargando entidad '%s' con llave pública '%s'", joinRequest.EntityName, joinRequest.PublicKey))
	_, err = global.ParsePubKeyRecibida(joinRequest.PublicKey)
	if err != nil {
		slog.Error("Error decodificando llave publica:", "error", err)
		return
	}

	invitationSplit := strings.Split(joinRequest.Invitation, "|")
	passphrase := n.GetValidator()
	salt := joinRequest.EntityName + ExtraerSalt(invitationSplit[1])
	key := global.GenerarLlaveDesdeFrase(passphrase, salt)
	invitation := global.DecipherInvitation(invitationSplit[0]+"|"+invitationSplit[1], joinRequest.InviterName, key)
	slog.Info("Invitación descifrada", "invitation", invitation)
	invitationSplit = strings.Split(invitation, ";")

	if invitationSplit[0] != joinRequest.InviterName {
		slog.Debug(fmt.Sprintf("Invitación invitador inválida: %s != %s", invitationSplit[0], joinRequest.InviterName))
		return
	}
	if invitationSplit[3] != joinRequest.Network {
		slog.Debug(fmt.Sprintf("Invitación Red inválida: %s != %s", invitationSplit[3], joinRequest.Network))
		return
	}
	if invitationSplit[2] != joinRequest.EntityName {
		slog.Debug(fmt.Sprintf("Invitación invitado inválida: %s != %s", invitationSplit[2], joinRequest.EntityName))
		return
	}

	a := false
	for _, b := range configuration.CM.GetConfig().Networks[0].Entities {
		if b.Name == invitationSplit[0] {
			a = true
			if !global.VerifyInvitation(invitation, b.Key) {
				slog.Error("Error verificando firma de la invitacion")
				return
			}
			break
		}
	}
	if !a {
		slog.Error("Error verificando firma, entidad no encontrada")
		return
	}

	expiracion, err := time.Parse(time.RFC3339, invitationSplit[4])
	if err != nil {
		slog.Error("Error al parsear la fecha:", "error", err)
		return
	}
	if time.Now().After(expiracion) {
		slog.Error("La invitación ha expirado.")
		return
	} else {
		slog.Info("La invitación es válida y se acepta")
		AddPrincipalAccessControl(joinRequest.Network, joinRequest.EntityName)
		n.PutCRDT(MEMBERS, joinRequest.EntityName, joinRequest.PublicKey)
	}
}

func ExtraerSalt(b64Input string) string {
	decodedBytes, err := base64.StdEncoding.DecodeString(b64Input)
	if err != nil {
		return ""
	}
	decodedStr := string(decodedBytes)
	decodedStr = strings.TrimSuffix(decodedStr, "/")
	partes := strings.Split(decodedStr, "/")
	if len(partes) == 0 {
		return ""
	}
	return partes[len(partes)-1]
}
