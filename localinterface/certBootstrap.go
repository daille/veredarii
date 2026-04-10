package localinterface

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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	caKeyFile      = "ca.key"
	caCertFile     = "ca.crt"
	serverKeyFile  = "server.key"
	serverCertFile = "server.crt"
	caValidity     = 10 * 365 * 24 * time.Hour
	serverValidity = 2 * 365 * 24 * time.Hour
	orgName        = "Veredarii LAN"
)

var serverSANs = []string{
	"localhost",
	"127.0.0.1",
	"::1",
}

func generateCA() (*ecdsa.PrivateKey, *x509.Certificate, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generando clave CA: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generando número de serie CA: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{orgName},
			OrganizationalUnit: []string{"Certificate Authority"},
			CommonName:         orgName + " Root CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(caValidity),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creando certificado CA: %w", err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parseando certificado CA: %w", err)
	}

	return caKey, caCert, nil
}

func generateServerCert(caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*ecdsa.PrivateKey, *x509.Certificate, error) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generando clave servidor: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generando número de serie servidor: %w", err)
	}

	var dnsNames []string
	var ipAddresses []net.IP
	for _, san := range serverSANs {
		if ip := net.ParseIP(san); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, san)
		}
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{orgName},
			OrganizationalUnit: []string{"Server"},
			CommonName:         "Veredarii Server",
		},
		NotBefore:   time.Now().Add(-1 * time.Minute),
		NotAfter:    time.Now().Add(serverValidity),
		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creando certificado servidor: %w", err)
	}

	serverCert, err := x509.ParseCertificate(serverDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parseando certificado servidor: %w", err)
	}

	return serverKey, serverCert, nil
}

func saveECKey(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

func saveCert(path string, cert *x509.Certificate) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func loadECKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM inválido en " + path)
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func loadCert(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM inválido en " + path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func bootstrapTLS() (tls.Certificate, error) {
	serverExists := fileExists(serverCertFile) && fileExists(serverKeyFile)
	caExists := fileExists(caCertFile) && fileExists(caKeyFile)

	if serverExists && caExists {
		log.Println("[TLS] Cargando certificados existentes desde disco")
		cert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("cargando certificado servidor: %w", err)
		}
		log.Printf("[TLS] Certificado servidor cargado desde %s", serverCertFile)
		log.Printf("[TLS] Distribuir a clientes si no lo tienen: %s", caCertFile)
		return cert, nil
	}

	log.Println("[TLS] Generando CA privada...")
	caKey, caCert, err := generateCA()
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := saveECKey(caKeyFile, caKey); err != nil {
		return tls.Certificate{}, fmt.Errorf("guardando clave CA: %w", err)
	}
	if err := saveCert(caCertFile, caCert); err != nil {
		return tls.Certificate{}, fmt.Errorf("guardando certificado CA: %w", err)
	}
	log.Printf("[TLS] CA generada → %s (distribuir a clientes)", caCertFile)

	log.Println("[TLS] Generando certificado de servidor...")
	serverKey, serverCert, err := generateServerCert(caKey, caCert)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := saveECKey(serverKeyFile, serverKey); err != nil {
		return tls.Certificate{}, fmt.Errorf("guardando clave servidor: %w", err)
	}
	if err := saveCert(serverCertFile, serverCert); err != nil {
		return tls.Certificate{}, fmt.Errorf("guardando certificado servidor: %w", err)
	}
	log.Printf("[TLS] Certificado servidor generado → %s", serverCertFile)
	log.Printf("[TLS] SANs incluidos: %v", serverSANs)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{serverCert.Raw, caCert.Raw},
		PrivateKey:  serverKey,
		Leaf:        serverCert,
	}
	return tlsCert, nil
}

func tlsConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func tlsVersionName(cs *tls.ConnectionState) string {
	if cs == nil {
		return "ninguno"
	}
	switch cs.Version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	default:
		return "desconocido"
	}
}
