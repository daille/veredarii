package handler

import (
	"net/http"
	"strconv"
	"strings"

	"Veredarii/configuration"
	"Veredarii/localinterface/data"
	mw "Veredarii/localinterface/middleware"

	"github.com/go-chi/chi/v5"
)

// ─── GET /api/resources/provided ─────────────────────────────────────────────

func GetProvidedResources(w http.ResponseWriter, r *http.Request) {
	var ProvidedResources = []data.Resource{}
	cfg := configuration.CM.GetConfig().Networks[0]
	for _, res := range cfg.Resources.API {
		ProvidedResources = append(ProvidedResources, data.Resource{
			ID:          1,
			Name:        res.Name,
			Type:        "API",
			Controlador: res.Plugin,
			Path:        cfg.ResourcesPath,
			Description: res.Plugin,
			Consumers:   8,
			//Version:     "v1.2",
			//Enabled:     true,
			//Visibility:  "public",
			//RateLimit:   500,
			//SLA:         99.9,
			//BackendURL:  "http://svc-ciudadanos:3000",
			//Format:      "json",
			//Tags:        "ciudadanos,identidad",
			//Contact:     "api@mi-org.cl",
			//DocsURL:     "https://docs.mi-org.cl/ciudadanos",
		})
	}
	for _, res := range cfg.Resources.TOPIC {
		ProvidedResources = append(ProvidedResources, data.Resource{
			ID:          1,
			Name:        res.Name,
			Type:        "TOPIC",
			Controlador: res.Plugin,
			Path:        cfg.ResourcesPath,
			Description: res.Plugin,
			Consumers:   8,
		})
	}

	mw.JSON(w, http.StatusOK, ProvidedResources)
}

// ─── GET /api/resources/provided/{id} ────────────────────────────────────────

func GetProvidedResource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	for _, res := range data.ProvidedResources {
		if res.ID == id {
			mw.JSON(w, http.StatusOK, res)
			return
		}
	}
	mw.Error(w, http.StatusNotFound, "recurso no encontrado")
}

// ─── POST /api/resources/provided ────────────────────────────────────────────

func CreateProvidedResource(w http.ResponseWriter, r *http.Request) {
	var body data.Resource
	if !mw.DecodeBody(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Type) == "" {
		mw.Error(w, http.StatusBadRequest, "name y type son obligatorios")
		return
	}
	body.ID = data.NextResourceID()
	if body.Version == "" {
		body.Version = "v1.0"
	}
	data.ProvidedResources = append(data.ProvidedResources, body)
	mw.JSON(w, http.StatusCreated, body)
}

// ─── PUT /api/resources/provided/{id} ────────────────────────────────────────

func UpdateProvidedResource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	var body data.Resource
	if !mw.DecodeBody(w, r, &body) {
		return
	}
	for i, res := range data.ProvidedResources {
		if res.ID == id {
			body.ID = id // nunca permitir cambio de ID
			data.ProvidedResources[i] = body
			mw.JSON(w, http.StatusOK, body)
			return
		}
	}
	mw.Error(w, http.StatusNotFound, "recurso no encontrado")
}

// ─── PATCH /api/resources/provided/{id} ──────────────────────────────────────
// Sólo actualiza el campo "enabled" (toggle activo/inactivo)

func ToggleProvidedResource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if !mw.DecodeBody(w, r, &body) {
		return
	}
	for i, res := range data.ProvidedResources {
		if res.ID == id {
			data.ProvidedResources[i].Enabled = body.Enabled
			mw.JSON(w, http.StatusOK, data.ProvidedResources[i])
			return
		}
	}
	mw.Error(w, http.StatusNotFound, "recurso no encontrado")
}

// ─── DELETE /api/resources/provided/{id} ─────────────────────────────────────

func DeleteProvidedResource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	for i, res := range data.ProvidedResources {
		if res.ID == id {
			data.ProvidedResources = append(
				data.ProvidedResources[:i],
				data.ProvidedResources[i+1:]...,
			)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	mw.Error(w, http.StatusNotFound, "recurso no encontrado")
}

// ─── GET /api/resources/external ─────────────────────────────────────────────

func GetExternalResources(w http.ResponseWriter, r *http.Request) {
	var ExternalResources = []data.Resource{}
	cfg := configuration.CM.GetConfig().Networks[0]

	for _, res := range cfg.RemoteResources.API {
		ExternalResources = append(ExternalResources, data.Resource{
			ID:          1,
			Name:        res.Name,
			Type:        "API",
			Controlador: res.Plugin,
			Path:        cfg.ResourcesPath,
			Description: res.Plugin,
			Consumers:   8,
		})
	}
	for _, res := range cfg.RemoteResources.TOPIC {
		ExternalResources = append(ExternalResources, data.Resource{
			ID:          1,
			Name:        res.Name,
			Type:        "API",
			Controlador: res.Plugin,
			Path:        cfg.ResourcesPath,
			Description: res.Plugin,
			Consumers:   8,
		})
	}

	mw.JSON(w, http.StatusOK, ExternalResources)
}

// ─── GET /api/resources/search?q=...&type=...&country=... ────────────────────

func SearchResources(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	typeF := r.URL.Query().Get("type")
	country := r.URL.Query().Get("country")

	results := []data.ExternalResource{}
	for _, res := range data.ExternalResources {
		if q != "" && !strings.Contains(strings.ToLower(res.Name), q) &&
			!strings.Contains(strings.ToLower(res.Org), q) &&
			!strings.Contains(strings.ToLower(res.Path), q) {
			continue
		}
		if typeF != "" && res.Type != typeF {
			continue
		}
		if country != "" && res.Country != country {
			continue
		}
		results = append(results, res)
	}

	// Si no hay resultados en el catálogo local, devolvemos sugerencias simuladas
	// (en producción esto buscaría en la red Veredarii real)
	if len(results) == 0 && q != "" {
		results = []data.ExternalResource{
			{
				ID: 900, Name: "Recurso: " + q, Type: "api",
				Org: "Red Veredarii", Path: "/api/v1/" + strings.ReplaceAll(q, " ", "-"),
				SLA: 99.0, Latency: 50, Auth: "Token", Subscribed: false, Country: "Chile",
			},
			{
				ID: 901, Name: "Dataset " + q, Type: "file",
				Org: "Datos Públicos CL", Path: "/files/" + strings.ReplaceAll(q, " ", "_") + ".csv",
				SLA: 97.0, Latency: 80, Auth: "none", Subscribed: false, Country: "Chile",
			},
		}
	}

	mw.JSON(w, http.StatusOK, results)
}
