// Package handler contiene los handlers HTTP agrupados por dominio.
package handler

import (
	"net/http"
	"strconv"

	"Veredarii/configuration"
	"Veredarii/localinterface/data"
	mw "Veredarii/localinterface/middleware"
)

// ─── GET /api/network/config ──────────────────────────────────────────────────

func GetNetworkConfig(w http.ResponseWriter, r *http.Request) {

	cfg := configuration.CM.GetConfig()
	port, _ := strconv.Atoi(cfg.Networks[0].Port)
	var Config = data.NodeConfig{
		NodeID:         cfg.Networks[0].Port,
		NodeName:       cfg.Networks[0].Name,
		Organization:   cfg.Identity.Entity,
		Description:    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		ListenAddr:     "0.0.0.0",
		Port:           port,
		PublicEndpoint: "https://nodo.mi-org.cl",
		BootstrapPeers: "peer1.veredarii.net:8080\npeer2.veredarii.net:8080",
		PubKeyPreview:  "ed25519:3J7kL...pQr9X",
		SigAlgo:        "ed25519",
		TLS:            "mutual",
		AutoRotateKeys: true,
		MaxConnections: 200,
		RateLimit:      1000,
		MaxPayloadKb:   512,
		ConnTimeout:    5000,
	}

	mw.JSON(w, http.StatusOK, Config)
}

// ─── PUT /api/network/config ──────────────────────────────────────────────────

func PutNetworkConfig(w http.ResponseWriter, r *http.Request) {
	var body data.NodeConfig
	if !mw.DecodeBody(w, r, &body) {
		return
	}
	// El nodeId no es editable desde el cliente
	body.NodeID = data.Config.NodeID
	data.Config = body
	mw.JSON(w, http.StatusOK, data.Config)
}

// ─── GET /api/network/status ──────────────────────────────────────────────────

func GetNetworkStatus(w http.ResponseWriter, r *http.Request) {
	mw.JSON(w, http.StatusOK, data.Status)
}

// ─── GET /api/network/metrics?period=1h ──────────────────────────────────────

func GetNetworkMetrics(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1h"
	}
	mw.JSON(w, http.StatusOK, data.GetMetrics(period))
}

// ─── GET /api/network/members ─────────────────────────────────────────────────

func GetNetworkMembers(w http.ResponseWriter, r *http.Request) {
	var Members = []data.Member{}

	for i, e := range configuration.CM.GetConfig().Networks[0].Entities {
		Members = append(Members, data.Member{
			ID:     i,
			NodeID: e.Key,
			Name:   e.Name,
		})
	}

	mw.JSON(w, http.StatusOK, Members)
}

// ─── GET /api/network/members/{id} ───────────────────────────────────────────

func GetNetworkMember(w http.ResponseWriter, r *http.Request) {
	/*
		// chi pone el id en la URL: /api/network/members/1
		// lo extraemos del path manualmente para no importar chi aquí
		// (los handlers son agnósticos al router)
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		for _, m := range configuration.CM.GetConfig().Networks[0].Entities {

			if strings.EqualFold(id, strings.TrimSpace(
				// comparamos contra el string del ID numérico
				func() string {
					return http.StatusText(m.Key) // truco: usamos el campo ID
				}(),
			)) {
				mw.JSON(w, http.StatusOK, m)
				return
			}
		}
		// búsqueda por ID numérico desde el path param inyectado por chi*/
	mw.Error(w, http.StatusNotFound, "miembro no encontrado")
}
