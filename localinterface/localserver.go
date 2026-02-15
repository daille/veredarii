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
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	log "github.com/sirupsen/logrus"

	configuration "Veredarii/configuration"
	"Veredarii/connection"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	manet "github.com/multiformats/go-multiaddr/net"
)

type LocalServer struct {
	Router *chi.Mux
}

func Start() {
	LocalServer := &LocalServer{
		Router: chi.NewRouter(),
	}
	_ = LocalServer.setupRouter()

	go func() {

		server := &http.Server{
			Addr:    ":" + configuration.CM.GetConfig().LocalInterface.Server.Port,
			Handler: LocalServer.Router,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error server HTTP: %v", err)
		}
	}()

	fmt.Println("Esperando señales o peticiones API...")

}

func (n *LocalServer) setupRouter() (iplocal string) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	for _, network := range configuration.CM.GetConfig().Networks {
		for _, service := range network.RemoteResources.API {
			r.Get("/"+network.Name+"/"+service.Name, func(w http.ResponseWriter, r *http.Request) {

				requestDump, err := httputil.DumpRequest(r, true)
				if err != nil {
					http.Error(w, "Error capturando petición", 500)
					return
				}

				targetID := connection.NM.Networks[network.Name].BuscarServicio(context.Background(), service.Name)
				if targetID == "" {
					log.Error("Servicio no encontrado")
					w.WriteHeader(http.StatusNotFound)
					return
				}
				respuesta := connection.NM.Networks[network.Name].Conversar(targetID, service.Name, requestDump)
				w.Write(respuesta)
			})
		}

		for _, datasource := range network.RemoteResources.DATASOURCE {
			r.Post("/"+network.Name+"/ds/"+datasource.Name, func(w http.ResponseWriter, r *http.Request) {
				var query connection.QueryType
				if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
					http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				// Validaciones opcionales
				if query.Query == "" {
					http.Error(w, "Query is required", http.StatusBadRequest)
					return
				}

				targetID := connection.NM.Networks[network.Name].BuscarServicio(context.Background(), datasource.Name)
				if targetID == "" {
					log.Error("Datasource no encontrado")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				connection.NM.Networks[network.Name].Query(targetID, query, datasource.Name)
			})
		}

		for _, addr := range connection.NM.Networks["red_interoperabilidad"].Host.Addrs() {
			if !manet.IsPublicAddr(addr) {
				iplocal = addr.String()
			}
		}

	}
	n.Router = r
	return iplocal
}

func generateSelfSignedCert(iplocal string) (tls.Certificate, error) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Veredarii"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP(iplocal)},
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}, nil
}
