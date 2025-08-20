package components

/**
 *    Veredarii, software for interoperability.
 *    This file is part of Veredarii.
 *
 *    @author jcDaille
 *
 *
 *    MIT License
 *
 * Copyright (c) 2025 JC Daille
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import (
	nw "Veredarii/components/network"
	"Veredarii/general"
	"Veredarii/util"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

type ServiceType struct{}

func middlewareSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("HSTS", "true")
		w.Header().Add("HSTSMaxAge", "31536000")
		w.Header().Add("HSTSIncludeSubdomains", "true")
		w.Header().Add("HSTSPreload", "false")
		next.ServeHTTP(w, r)
	})
}

func CreateLocalServer() {
	var ServidorInterno *http.Server
	for {
		log.Debug("Esperando para crear local server")
		select {
		case <-general.Chan.StartLocalServer:
			log.Debug("Starting local server...")
			router := chi.NewRouter()
			router.Use(middleware.ThrottleBacklog(100, 100, time.Second*1))
			router.Use(middlewareSecurity)
			router.Route("/local", func(subrouter chi.Router) {
				srv := ServiceType{}
				subrouter.Get("/", func(w http.ResponseWriter, r *http.Request) { localHandler(srv, w, r) })
				subrouter.Post("/", func(w http.ResponseWriter, r *http.Request) { localHandler(srv, w, r) })
				subrouter.Put("/", func(w http.ResponseWriter, r *http.Request) { localHandler(srv, w, r) })
				subrouter.Delete("/", func(w http.ResponseWriter, r *http.Request) { localHandler(srv, w, r) })
				subrouter.Head("/", func(w http.ResponseWriter, r *http.Request) {})
				subrouter.Options("/", func(w http.ResponseWriter, r *http.Request) {})
			})

			ServidorInterno = &http.Server{Addr: ":" + general.Configuration.Behavior.Local.Port, Handler: router}
			log.Info(util.Info("[Servidor Interno] Servidor Interno desplegado (puerto ", general.Configuration.Behavior.Local.Port, ")"))
			go ServidorInterno.ListenAndServe()
			/*if err != nil {
				log.Error("NO SE PUDO INICIAR EL NODO.\n", err)
				os.Exit(1)
			}*/
		case <-general.Chan.StopLocalServer:
			fmt.Println("Stoping server")
			ServidorInterno.Close()
		}
	}
}

func localHandler(srv ServiceType, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Llegó peticion local")
	b, _ := io.ReadAll(r.Body)
	nw.NetworkConnectionList[0].SendMessage(b, "12D3KooWJLyx7JzpAYqYgQk4gtme4aRpYMC3wQesxugHFuVSDQR7")
}
