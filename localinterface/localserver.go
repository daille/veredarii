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
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	log "github.com/sirupsen/logrus"

	configuration "Veredarii/configuration"
	"Veredarii/connection"
	handler "Veredarii/localinterface/handler"
	pluginmanager "Veredarii/pluginManager"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	manet "github.com/multiformats/go-multiaddr/net"
)

type LocalServer struct {
	Router *chi.Mux
}

func Start() {
	cert, err := bootstrapTLS()
	if err != nil {
		log.Fatalf("[TLS] Error fatal: %v", err)
	}

	LocalServer := &LocalServer{
		Router: chi.NewRouter(),
	}
	_ = LocalServer.setupRouter()

	go func() {
		server := &http.Server{
			Addr:              ":" + configuration.CM.GetConfig().LocalInterface.Server.Port,
			Handler:           LocalServer.Router,
			TLSConfig:         tlsConfig(cert),
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Printf("Error server HTTP: %v", err)
		}
	}()
	log.Info("Esperando peticiones de API en el puerto ", configuration.CM.GetConfig().LocalInterface.Server.Port)
}

func (n *LocalServer) setupRouter() (iplocal string) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Use(SecurityAPIMiddleware)
		// — Red —
		r.Route("/network", func(r chi.Router) {
			r.Get("/config", handler.GetNetworkConfig)
			r.Put("/config", handler.PutNetworkConfig)
			r.Get("/status", handler.GetNetworkStatus)
			r.Get("/metrics", handler.GetNetworkMetrics) // ?period=1h|6h|24h|7d
			r.Get("/members", handler.GetNetworkMembers)
			r.Get("/members/{id}", handler.GetNetworkMember)
		})

		// — Recursos —
		r.Route("/resources", func(r chi.Router) {
			r.Get("/provided", handler.GetProvidedResources)
			r.Post("/provided", handler.CreateProvidedResource)
			r.Get("/provided/{id}", handler.GetProvidedResource)
			r.Put("/provided/{id}", handler.UpdateProvidedResource)
			r.Patch("/provided/{id}", handler.ToggleProvidedResource) // { "enabled": true/false }
			r.Delete("/provided/{id}", handler.DeleteProvidedResource)

			r.Get("/external", handler.GetExternalResources)
			r.Get("/search", handler.SearchResources) // ?q=...&type=...&country=...
		})

		// — Mapa —
		r.Route("/map", func(r chi.Router) {
			r.Get("/topology", handler.GetTopology)
			r.Get("/flows", handler.GetFlows) // ?period=1h|24h|7d
		})
	})

	for _, network := range configuration.CM.GetConfig().Networks {
		// Rutas a tratar como proxy
		for _, service := range network.RemoteResources.API {
			r.Get("/"+network.Name+"/"+service.ResourcePath, func(w http.ResponseWriter, r *http.Request) {
				if service.Plugin == "" {
					requestDump, err := httputil.DumpRequest(r, true)
					if err != nil {
						http.Error(w, "Error capturando petición", 500)
						return
					}

					respuesta := connection.NM.Networks[network.Name].Send(service.Name, requestDump)
					w.Write(respuesta)
				} else {
					body, _ := io.ReadAll(r.Body)
					resultado, err := pluginmanager.PM.Execute(r.Context(), service.Plugin, "handle_request", body)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Write(resultado)
				}
			})
		}

		/*for _, datasource := range network.RemoteResources.DATASOURCE {
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
		}*/

		if connection.NM.Networks[network.Name].Host != nil {
			for _, addr := range connection.NM.Networks[network.Name].Host.Addrs() {
				if !manet.IsPublicAddr(addr) {
					iplocal = addr.String()
				}
			}
		}

	}
	n.Router = r
	return iplocal
}

func SecurityAPIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Verificar que el host sea local
		// Nota: r.Host puede ser "localhost:8080" o "127.0.0.1:8080"
		isLocal := r.Host == "localhost:8080" || r.Host == "127.0.0.1:8080"

		if !isLocal {
			http.Error(w, "Acceso denegado: Solo conexiones locales permitidas.", http.StatusForbidden)
			return
		}

		// 2. Opcional: Verificar el Header Origin si viene de un navegador
		origin := r.Header.Get("Origin")
		if origin != "" && origin != "http://localhost:3000" { // puerto de tu front
			http.Error(w, "Origin no autorizado", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

/*func generateSelfSignedCert(iplocal string) (tls.Certificate, error) {
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
}*/
