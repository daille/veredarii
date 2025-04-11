package configuration

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

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
	"NodoCb/global"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ed25519"
)

func ObtenerLlaves() global.KeysType {
	var keys global.KeysType

	return keys
}

func CrearLlavesFirma() (string, string) {
	pubKey, priKey, _ := ed25519.GenerateKey(nil)
	//log.Println("PUB:", hex.EncodeToString(pubKey))
	//log.Println("PRIV:", hex.EncodeToString(priKey))
	return hex.EncodeToString(pubKey), hex.EncodeToString(priKey)
}

func CrearLlavesTLS(nombre string) (string, string) {

	result := strings.Split(nombre, " ")
	soloNombre := result[len(result)-1]

	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{nombre},
			Country:      []string{"CL"},
			Province:     []string{""},
			Locality:     []string{"Santiago"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		DNSNames:              []string{"*." + soloNombre + ".cl"},
		SubjectKeyId:          []byte{1, 2, 3, 4, 6},
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", ""
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return "", ""
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return caPEM.String(), caPrivKeyPEM.String()
}

func LoadX509Certificate(certFile string) *x509.Certificate {
	cf, e := os.ReadFile(certFile)
	if e != nil {
		fmt.Println("Load File Cert:", e.Error())
		return nil
	}
	cpb, _ := pem.Decode(cf)
	crt, e := x509.ParseCertificate(cpb.Bytes)
	if e != nil {
		fmt.Println("parsex509:", e.Error())
		return nil
	}
	return crt
}

func LoadX509CertificateFromString(certBody string) *x509.Certificate {
	cpb, _ := pem.Decode([]byte(certBody))
	crt, e := x509.ParseCertificate(cpb.Bytes)
	if e != nil {
		fmt.Println("parsex509:", e.Error())
		return nil
	}
	return crt
}

func LoadX509PrivateKey(keyFile string) *rsa.PrivateKey {
	kf, e := os.ReadFile(keyFile)
	if e != nil {
		fmt.Println("kfload:", e.Error())
	}
	kpb, kr := pem.Decode(kf)
	fmt.Println(string(kr))

	key, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if e != nil {
		fmt.Println("parsekey:", e.Error())
	}
	return key
}

func LoadX509PrivateKeyFromString(keyFile string) *rsa.PrivateKey {
	cpb, _ := pem.Decode([]byte(keyFile))
	crt, e := x509.ParsePKCS1PrivateKey(cpb.Bytes)
	if e != nil {
		fmt.Println("parsex509:", e.Error())
		return nil
	}
	return crt
}
