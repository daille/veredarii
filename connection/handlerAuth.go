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
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"log/slog"

	"github.com/oklog/ulid/v2"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/record"
)

var cache = NonceCache{firmas: make(map[string]time.Time)}

type NonceCache struct {
	sync.RWMutex
	firmas map[string]time.Time
}

type EnvioMasterEntities struct {
	Entities map[string][]byte `json:"entities"`
}

type EntidadRecord struct {
	ID         string  `json:"id"`
	EntityName string  `json:"entity"`
	PeerID     peer.ID `json:"peer_id"`
	//ExpiresAt  int64   `json:"expires_at"`
	Signature []byte `json:"signature"`
}

func (r *EntidadRecord) Domain() string                 { return "pisee-auth-v1" }
func (r *EntidadRecord) Codec() []byte                  { return []byte("/pisee/entidad-auth/1.0.0") }
func (r *EntidadRecord) MarshalRecord() ([]byte, error) { return json.Marshal(r) }
func (r *EntidadRecord) UnmarshalRecord(b []byte) error { return json.Unmarshal(b, r) }

func (n *Network) handleAuthStream(s network.Stream) {
	defer s.Close()
	remotePeer := s.Conn().RemotePeer()
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	slog.Debug("[Validacion] Verificando sobre", "peer", remotePeer)

	envelope, err := recibirSobre(rw)
	if err != nil {
		slog.Error("Error al recibir sobre", "error", err)
		s.Write([]byte{0})
		s.Reset()
		return
	}

	envelopeBytes, err := envelope.Marshal()
	if err != nil {
		slog.Error("Error serializando sobre", "error", err)
		s.Write([]byte{0})
		s.Reset()
		return
	}

	if rec, err := n.verificarEntidad(envelopeBytes, remotePeer); err != nil {
		slog.Info("Error verificando entidad", "error", err)
		s.Write([]byte{0})
		s.Reset()
		return
	} else {
		RBAC.SetPeer(*rec)
		n.Peers[remotePeer] = PeerType{
			ID:     remotePeer,
			Entity: rec.EntityName,
		}
		if err != nil {
			slog.Error("Error serializando sobre", "error", err)
		} else {
			pr := PeerRequest{
				PeerID:     remotePeer.String(),
				EntityName: rec.EntityName,
				Address:    s.Conn().RemoteMultiaddr(),
			}
			peerRequest, err := json.Marshal(pr)
			if err != nil {
				slog.Error("Error serializando sobre", "error", err)
				return
			}
			preKey := sha256.Sum256([]byte(n.SwarmKey))
			key := preKey[:]
			peerRequest, err = global.Encrypt(peerRequest, key)
			if err != nil {
				slog.Error("Error al cifrar la solicitud", "error", err)
				return
			}
			ctx := context.Background()
			n.PutCRDT(PEERS, remotePeer.String(), rec.EntityName)
			n.NetworkPeerTopic.Publish(ctx, peerRequest)
		}
	}

	n.MutexSesiones.Lock()
	n.SesionesActivas[remotePeer] = "usuario_verificado"
	n.MutexSesiones.Unlock()
	slog.Info("ACEPTADO.")
}

func (n *Network) Authenticar(ctx context.Context, priv crypto.PrivKey, peerID peer.ID) error {
	rec, err := FirmarRecordConULID(
		configuration.CM.GetConfig().Identity.Entity,
		n.Host.ID(),
		time.Hour,
	)
	if err != nil {
		slog.Error("Error firmando el record", "error", err)
	}

	envelope, err := record.Seal(rec, priv)
	if err != nil {
		slog.Error("Error al sellar el sobre", "error", err)
	}

	sAuth, err := n.Host.NewStream(ctx, peerID, protocol.ID(global.ProtocolAuth))
	if err != nil {
		slog.Error("No se pudo abrir el stream de autenticación", "error", err)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(sAuth), bufio.NewWriter(sAuth))

	slog.Debug("Enviando credenciales de entidad...")
	envelopeBytes, err := envelope.Marshal()
	if err != nil {
		return err
	}

	length := uint32(len(envelopeBytes))
	if err := binary.Write(rw.Writer, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := rw.Writer.Write(envelopeBytes); err != nil {
		return err
	}
	slog.Debug("Autenticación enviada. Esperando validación...", "id", rec.ID, "entity", rec.EntityName, "peer", rec.PeerID.String())
	rw.Flush()
	resp, err := io.ReadAll(sAuth)
	if err != nil {
		slog.Error("Error al leer la respuesta", "error", err)
	}
	slog.Info("El servidor aceptó la autenticación", "respuesta", resp)
	sAuth.Close()
	return nil
}

func FirmarRecordConULID(name string, pID peer.ID, ttl time.Duration) (*EntidadRecord, error) {
	privKey, err := obtenerMasterKey(configuration.CM.GetConfig().Identity.PrivKeyFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo cargar la llave privada: %w", err)
	}
	id := ulid.Make().String()

	msgAuth := []byte(fmt.Sprintf("%s:%s:%s", id, name, pID.String()))
	signature, err := privKey.Sign(msgAuth)
	pkb_auto := privKey.GetPublic()
	valid, err := pkb_auto.Verify(msgAuth, signature)
	if err != nil {
		slog.Error("Error al verificar firma", "error", err)
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("firma inválida")
	}

	return &EntidadRecord{
		ID:         id,
		EntityName: name,
		PeerID:     pID,
		Signature:  signature,
	}, nil
}

func obtenerMasterKey(ruta string) (crypto.PrivKey, error) {
	rawBytes, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	if len(rawBytes) > 64 {
		rawBytes = rawBytes[len(rawBytes)-64:]
	}
	return crypto.UnmarshalEd25519PrivateKey(rawBytes)
}

func (n *Network) verificarEntidad(envelopeBytes []byte, remotePeer peer.ID) (*EntidadRecord, error) {
	envelope, recordObj, err := record.ConsumeEnvelope(envelopeBytes, "pisee-auth-v1")
	if err != nil {
		return nil, fmt.Errorf("error al procesar el sobre: %w", err)
	}
	if envelope == nil || envelope.PublicKey == nil {
		return nil, fmt.Errorf("sobre o llave pública nula")
	}

	idDelSobre, err := peer.IDFromPublicKey(envelope.PublicKey)
	if err != nil || idDelSobre != remotePeer {
		return nil, fmt.Errorf("el sobre no pertenece al PeerID conectado")
	}
	if recordObj == nil {
		return nil, fmt.Errorf("el contenido del sobre (recordObj) es nulo")
	}

	rec, ok := recordObj.(*EntidadRecord)
	if !ok || rec == nil {
		return nil, fmt.Errorf("el contenido del sobre no es un EntidadRecord válido o es nil")
	}

	if esReplay(rec.Signature) {
		return nil, fmt.Errorf("ataque de replay: este sobre ya fue utilizado")
	}

	masterPubKey, existe := n.MasterEntities[rec.EntityName]
	if existe {
		mpk, _ := masterPubKey.Raw()
		slog.Debug("Verificando entidad	", "Master", rec.EntityName, "entidad", rec.EntityName, "key", masterPubKey, "peer", rec.PeerID.String(), "mpk", hex.EncodeToString(mpk))
		if !existe || masterPubKey == nil {
			return nil, fmt.Errorf("la entidad '%s' no está configurada o su llave pública es nula", rec.EntityName)
		}
	} else {
		slog.Debug("Verificando entidad	", "Master", rec.EntityName, "entidad", rec.EntityName, "key", masterPubKey, "peer", rec.PeerID.String())
		return nil, fmt.Errorf("la entidad '%s' no está configurada o su llave pública es nula", rec.EntityName)
	}
	msgAuth := []byte(fmt.Sprintf("%s:%s:%s", rec.ID, rec.EntityName, rec.PeerID.String()))
	valid, err := masterPubKey.Verify(msgAuth, rec.Signature)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar verificación de firma: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("la firma de la entidad maestra es inválida")
	}

	return rec, nil
}

func esReplay(firma []byte) bool {
	cache.Lock()
	defer cache.Unlock()
	return false
}
