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

	"github.com/oklog/ulid/v2"
	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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

	resp, err := SerializarMasterEntities(n.MasterEntities)
	if err != nil {
		fmt.Println("Error serializando master entities:", err)
		s.Write([]byte{0})
		s.Reset()
		return
	}
	s.Write(resp)
	fmt.Println(" ACEPTADO.")
}

func (n *Network) Authenticar(ctx context.Context, priv crypto.PrivKey, peerID peer.ID) error {

	rec, err := FirmarRecordConULID(
		configuration.CM.GetConfig().Identity.Entity,
		n.Host.ID(),
		time.Hour,
	)
	if err != nil {
		log.Fatal("Error firmando el record:", err)
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
	log.Debug(fmt.Sprintf("%s:%s:%s", rec.ID, rec.EntityName, rec.PeerID.String()))
	fmt.Println("Autenticación enviada. Esperando validación...")
	rw.Flush()

	// Leer respuesta del servidor
	resp, err := io.ReadAll(sAuth)
	if err != nil {
		log.Fatal(err)
	}
	masterEntities, err := DeserializarMasterEntities(resp)
	if err != nil {
		log.Error("Error deserializando master entities:", err)
		return err
	}
	fmt.Println("✅ El servidor aceptó la autenticación", masterEntities)

	n.MutexSesiones.Lock()
	for nombre, key := range masterEntities {
		n.MasterEntities[nombre] = key
	}
	n.MutexSesiones.Unlock()
	sAuth.Close()

	return nil
}

func DeserializarMasterEntities(data []byte) (map[string]crypto.PubKey, error) {
	var transporte EnvioMasterEntities
	if err := json.Unmarshal(data, &transporte); err != nil {
		return nil, err
	}

	master := make(map[string]crypto.PubKey)
	for nombre, keyBytes := range transporte.Entities {
		// Reconstruimos la interfaz PubKey desde los bytes
		pubKey, err := crypto.UnmarshalPublicKey(keyBytes)
		if err != nil {
			return nil, err
		}
		master[nombre] = pubKey
	}

	return master, nil
}

func FirmarRecordConULID(name string, pID peer.ID, ttl time.Duration) (*EntidadRecord, error) {

	privKey, err := obtenerMasterKey(configuration.CM.GetConfig().Identity.PrivKeyFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo cargar la llave privada: %w", err)
	}
	id := ulid.Make().String()
	//expiration := time.Now().Add(ttl).Unix()
	//dataToSign := fmt.Sprintf("%s:%s:%s:%d", id, name, pID.String(), expiration)
	//sigBytes := ed25519.Sign(privKey, []byte(dataToSign))

	// 3. Firmar usando la interfaz de libp2p
	msgAuth := []byte(fmt.Sprintf("%s:%s:%s", id, name, pID.String()))
	signature, err := privKey.Sign(msgAuth) // Esto devuelve []byte
	pkb_auto := privKey.GetPublic()
	valid, err := pkb_auto.Verify(msgAuth, signature)
	if err != nil {
		log.Fatalf("Error al verificar firma: %v", err)
	}

	pubKeyRaw, err := hex.DecodeString("080112202c06e7dbf218d0f26edb337c1e5f90dbc3f729bc2e08feb0a78863c1782e62af")
	if err != nil {
		log.Fatalf("Error al decodificar hexadecimal: %v", err)
	}
	pkb, err := crypto.UnmarshalPublicKey(pubKeyRaw)
	if err != nil {
		log.Fatalf("Error al procesar llave pública: %v", err)
	}

	valid, err = pkb.Verify(msgAuth, signature)
	if !valid {
		log.Debug("Firma inválida")
	} else {
		log.Debug("Firma válida")
	}

	return &EntidadRecord{
		ID:         id,
		EntityName: name,
		PeerID:     pID,
		//ExpiresAt:  expiration,
		Signature: signature,
	}, nil
}

func obtenerMasterKey(ruta string) (crypto.PrivKey, error) {
	rawBytes, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	// Si el archivo tiene el prefijo 08011240 (libp2p protobuf)
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
		log.Debug("Master: ", rec.EntityName, " : ", hex.EncodeToString(mpk))
		log.Debug(fmt.Sprintf("Verificando entidad '%s' con llave pública '%s'", rec.EntityName, masterPubKey))
		if !existe || masterPubKey == nil {
			return nil, fmt.Errorf("la entidad '%s' no está configurada o su llave pública es nula", rec.EntityName)
		}
	} else {
		return nil, fmt.Errorf("la entidad '%s' no está configurada o su llave pública es nula", rec.EntityName)
	}

	log.Debug(fmt.Sprintf("%s:%s:%s", rec.ID, rec.EntityName, rec.PeerID.String()))
	msgAuth := []byte(fmt.Sprintf("%s:%s:%s", rec.ID, rec.EntityName, rec.PeerID.String()))
	log.Debug(fmt.Sprintf("SERVER PAYLOAD HEX: %x", []byte(msgAuth)))
	log.Debug(fmt.Sprintf("SERVER SIG HEX: %x", rec.Signature))

	// 4. Verificar
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

}

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

func SerializarMasterEntities(master map[string]crypto.PubKey) ([]byte, error) {
	transporte := EnvioMasterEntities{
		Entities: make(map[string][]byte),
	}

	for nombre, pubKey := range master {
		// Convertimos la PubKey al formato estándar de libp2p (Protobuf)
		keyBytes, err := crypto.MarshalPublicKey(pubKey)
		if err != nil {
			return nil, err
		}
		transporte.Entities[nombre] = keyBytes
	}

	return json.Marshal(transporte)
}
