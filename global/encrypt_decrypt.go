package global

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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/pnet"
)

func ObtenerIdentidad(path string) (crypto.PrivKey, error) {
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

func GenerarLlaveDesdeFrase(passphrase string, salt string) []byte {
	saltByte := []byte(salt)

	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)
	key := argon2.IDKey([]byte(passphrase), saltByte, time, memory, threads, keyLen)
	return key
}

func DecodeV1PSK(hexKey string) (pnet.PSK, error) {
	header := "/key/swarm/psk/1.0.0/\n/base16/\n"
	return pnet.DecodeV1PSK(strings.NewReader(header + hexKey))
}

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext demasiado corto")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

func GetPubKey(priv crypto.PrivKey) string {
	pub := priv.GetPublic()
	pubBytes, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		log.Error(err)
		return ""
	}
	pubString := base64.StdEncoding.EncodeToString(pubBytes)
	return pubString
}

func CipherInvitation(i InvitacionType, key []byte) string {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Error creando cipher: %v", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Error creando GCM: %v", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatalf("Error generando nonce: %v", err)
	}
	aad := []byte("auth:" + i.Inviter)

	plaintext := fmt.Sprintf("%s;%s;%s;%s;%s", i.Inviter, i.PeerID, i.Guest, i.Network, i.Expiration.Format(time.RFC3339))
	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), aad)

	return hex.EncodeToString(ciphertext)
}

func DecipherInvitation(tokenHex string, invitador string, key []byte) string {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Error creando cipher: %v", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Error creando GCM: %v", err)
	}

	data, err := hex.DecodeString(tokenHex)
	if err != nil {
		log.Fatalf("Error decodificando hex: %v", err)
	}

	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		log.Fatalf("Token demasiado corto")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	aad := []byte("auth:" + invitador)

	decrypted, err := aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		log.Fatalf("Error de autenticación: El token fue alterado o la llave es incorrecta")
	}

	return string(decrypted)
}

func ParsePubKeyRecibida(pubString string) (crypto.PubKey, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(pubString)
	if err != nil {
		return nil, fmt.Errorf("error decodificando base64: %w", err)
	}

	pubKey, err := crypto.UnmarshalPublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal libp2p: %w", err)
	}

	return pubKey, nil
}
