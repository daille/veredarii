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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/record"
)

type EntidadRecord struct {
	EntityName string  `json:"entity"`
	PeerID     peer.ID `json:"peer_id"`
	Signature  []byte  `json:"signature"`
}

func (r *EntidadRecord) Domain() string                 { return "pisee-auth-v1" }
func (r *EntidadRecord) Codec() []byte                  { return []byte("/pisee/entidad-auth/1.0.0") }
func (r *EntidadRecord) MarshalRecord() ([]byte, error) { return json.Marshal(r) }
func (r *EntidadRecord) UnmarshalRecord(b []byte) error { return json.Unmarshal(b, r) }

func (n *Network) handleAuthStream(s network.Stream) {
	defer s.Close()
	remotePeer := s.Conn().RemotePeer()
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	fmt.Printf("[Validacion] Verificando sobre de: %s...", remotePeer)

	envelope, err := recibirSobre(rw)
	if err != nil {
		fmt.Printf(" RECHAZADO: %v\n", err)
		s.Write([]byte{0})
		s.Reset()
		return
	}

	envelopeBytes, err := envelope.Marshal()
	if err != nil {
		fmt.Println("Error serializando sobre:", err)
		s.Write([]byte{0})
		s.Reset()
		return
	}

	if rec, err := n.verificarEntidad(envelopeBytes, remotePeer); err != nil {
		fmt.Printf(" RECHAZADO: %v\n", err)
		s.Write([]byte{0})
		s.Reset()
		return
	} else {
		RBAC.SetPeer(*rec)
	}

	n.MutexSesiones.Lock()
	n.SesionesActivas[remotePeer] = "usuario_verificado"
	n.MutexSesiones.Unlock()

	s.Write([]byte{1})
	fmt.Println(" ACEPTADO.")
}

func (n *Network) Authenticar(ctx context.Context, priv crypto.PrivKey, peerID peer.ID) error {

	firmaDeLaEntidad, err := hex.DecodeString(configuration.CM.GetConfig().Identity.Firma)
	if err != nil {
		log.Fatal("Error decodificando la firma hex:", err)
	}

	rec := &EntidadRecord{
		EntityName: configuration.CM.GetConfig().Identity.Entity,
		PeerID:     n.Host.ID(),
		Signature:  firmaDeLaEntidad,
	}

	// Sellar el sobre con la llave privada del cliente
	envelope, err := record.Seal(rec, priv)
	if err != nil {
		log.Fatal("Error al sellar el sobre:", err)
	}

	sAuth, err := n.Host.NewStream(ctx, peerID, global.ProtocolAuth)
	if err != nil {
		log.Fatal("No se pudo abrir el stream de autenticación:", err)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(sAuth), bufio.NewWriter(sAuth))

	// ENVIAR SOBRE INMEDIATAMENTE
	fmt.Println("Enviando credenciales de entidad...")
	envelopeBytes, err := envelope.Marshal()
	if err != nil {
		return err
	}

	// Prefijo de longitud para que el servidor sepa cuánto leer
	length := uint32(len(envelopeBytes))
	if err := binary.Write(rw.Writer, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := rw.Writer.Write(envelopeBytes); err != nil {
		return err
	}
	fmt.Println("Autenticación enviada. Esperando validación...")
	rw.Flush()

	// Leer respuesta del servidor
	resp := make([]byte, 1)
	_, err = sAuth.Read(resp)

	if err != nil || resp[0] != 1 {
		fmt.Println("❌ El servidor rechazó la autenticación")
		return err
	}
	sAuth.Close()

	return nil
}

// --- FUNCIONES DE SOPORTE ---
func recibirSobre(rw *bufio.ReadWriter) (*record.Envelope, error) {
	var length uint32
	if err := binary.Read(rw.Reader, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("error longitud: %w", err)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(rw.Reader, buf); err != nil {
		return nil, fmt.Errorf("error leyendo bytes: %w", err)
	}

	envelope, err := record.UnmarshalEnvelope(buf)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar sobre (wire-format): %w", err)
	}

	return envelope, nil
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
	if !existe || masterPubKey == nil {
		return nil, fmt.Errorf("la entidad '%s' no está configurada o su llave pública es nula", rec.EntityName)
	}

	msgAuth := []byte(rec.PeerID.String() + rec.EntityName)
	valid, err := masterPubKey.Verify(msgAuth, rec.Signature)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar verificación de firma: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("la firma de la entidad maestra es inválida")
	}

	return rec, nil
}

func (n *Network) cargarWhitelist() {
	ids := []string{
		"12D3KooWAkCUB7wJ1p8748BFFDZS1Vdx5N2ywp8BDh34jcrBPudA",
		"12D3KooWEdjx22FMtiWYuH7AAn365doDrAZN2GnttpnmEfPvs8hh",
		"12D3KooWEgeYgNpgnbVFP3hyCog1kLF7RbmD1XmiopWjkaCtnV2b",
	}

	for _, s := range ids {
		id, err := peer.Decode(s)
		if err != nil {
			log.Printf("Error decodificando PeerID %s: %v", s, err)
			continue
		}
		n.Whitelist[id] = true
	}
}

func (n *Network) obtenerIdentidad(path string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return crypto.UnmarshalPrivateKey(data)
	}
	fmt.Println("Generando nueva identidad...")
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, err
	}
	data, err = crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(path, data, 0600)
	return priv, err
}

type NonceCache struct {
	sync.RWMutex
	firmas map[string]time.Time
}

var cache = NonceCache{firmas: make(map[string]time.Time)}

func esReplay(firma []byte) bool {
	cache.Lock()
	defer cache.Unlock()

	return false
}

func limpiarCache() {
	for {
		time.Sleep(1 * time.Minute)
		ahora := time.Now()
		cache.Lock()
		for f, t := range cache.firmas {
			if ahora.Sub(t) > 1*time.Minute {
				delete(cache.firmas, f)
			}
		}
		cache.Unlock()
	}
}
