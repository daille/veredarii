// Package data contiene todos los datos dummy en memoria.
// Cuando el backend real esté listo, este paquete se reemplaza
// por llamadas a base de datos sin tocar los handlers.
package data

import "time"

// ─── Network ──────────────────────────────────────────────────────────────────

type NodeConfig struct {
	NodeID         string `json:"nodeId"`
	NodeName       string `json:"nodeName"`
	Organization   string `json:"organization"`
	Description    string `json:"description"`
	ListenAddr     string `json:"listenAddr"`
	Port           int    `json:"port"`
	PublicEndpoint string `json:"publicEndpoint"`
	BootstrapPeers string `json:"bootstrapPeers"`
	PubKeyPreview  string `json:"pubKeyPreview"`
	SigAlgo        string `json:"sigAlgo"`
	TLS            string `json:"tls"`
	AutoRotateKeys bool   `json:"autoRotateKeys"`
	MaxConnections int    `json:"maxConnections"`
	RateLimit      int    `json:"rateLimit"`
	MaxPayloadKb   int    `json:"maxPayloadKb"`
	ConnTimeout    int    `json:"connTimeout"`
}

type NodeStatus struct {
	Online  bool   `json:"online"`
	State   string `json:"state"`
	Version string `json:"version"`
	Peers   []Peer `json:"peers"`
	Uptime  string `json:"uptime"`
}

type Peer struct {
	ID      string `json:"id"`
	Org     string `json:"org"`
	Latency int    `json:"latency"`
	TX      int    `json:"tx"`
	Status  string `json:"status"`
}

type NetworkMetrics struct {
	InboundBw      string    `json:"inboundBw"`
	InboundDelta   string    `json:"inboundDelta"`
	OutboundBw     string    `json:"outboundBw"`
	OutboundDelta  string    `json:"outboundDelta"`
	TxTotal        string    `json:"txTotal"`
	TxDelta        string    `json:"txDelta"`
	ErrorCount     string    `json:"errorCount"`
	ErrorDelta     string    `json:"errorDelta"`
	LatencyP50     string    `json:"latencyP50"`
	LatencyDelta   string    `json:"latencyDelta"`
	Labels         []string  `json:"labels"`
	InboundSeries  []float64 `json:"inboundSeries"`
	OutboundSeries []float64 `json:"outboundSeries"`
	TxSeries       []float64 `json:"txSeries"`
	LatP50Series   []float64 `json:"latP50Series"`
	LatP95Series   []float64 `json:"latP95Series"`
}

type Member struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	NodeID    string `json:"nodeId"`
	Type      string `json:"type"`
	TypeKey   string `json:"typeKey"`
	Country   string `json:"country"`
	Resources int    `json:"resources"`
	LastSeen  string `json:"lastSeen"`
	Status    string `json:"status"`
}

// ─── Resources ────────────────────────────────────────────────────────────────

type Resource struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Path        string  `json:"path"`
	Version     string  `json:"version"`
	Controlador string  `json:"controlador"`
	Description string  `json:"description"`
	Consumers   int     `json:"consumers"`
	Enabled     bool    `json:"enabled"`
	Auth        string  `json:"auth"`
	Visibility  string  `json:"visibility"`
	RateLimit   int     `json:"rateLimit"`
	SLA         float64 `json:"sla"`
	BackendURL  string  `json:"backendUrl,omitempty"`
	Format      string  `json:"format,omitempty"`
	Endpoints   string  `json:"endpoints,omitempty"`
	Retention   int     `json:"retention,omitempty"`
	Schema      string  `json:"schema,omitempty"`
	Tags        string  `json:"tags"`
	Contact     string  `json:"contact"`
	DocsURL     string  `json:"docsUrl"`
}

type ExternalResource struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Org        string  `json:"org"`
	Path       string  `json:"path"`
	SLA        float64 `json:"sla"`
	Latency    int     `json:"latency"`
	Auth       string  `json:"auth"`
	Subscribed bool    `json:"subscribed"`
	Country    string  `json:"country"`
}

// ─── Map ──────────────────────────────────────────────────────────────────────

type MapNode struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	IsCenter  bool   `json:"isCenter"`
	Type      string `json:"type"`
	OutTx     int    `json:"outTx"`
	InTx      int    `json:"inTx"`
	Latency   int    `json:"latency"`
	Resources int    `json:"resources"`
}

type Topology struct {
	Nodes []MapNode `json:"nodes"`
}

type Flows struct {
	TotalOut string `json:"totalOut"`
	TotalIn  string `json:"totalIn"`
	Period   string `json:"period"`
}

// ─── Datos dummy en memoria ───────────────────────────────────────────────────

// Config es mutable (el PUT /network/config lo actualiza en memoria)
var Config = NodeConfig{
	NodeID:         "vrd-3a7f-9c2e-b41d-8052",
	NodeName:       "nodo-santiago-01",
	Organization:   "Mi Organización",
	Description:    "Nodo primario de interoperabilidad para servicios institucionales.",
	ListenAddr:     "0.0.0.0",
	Port:           8080,
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

var Status = NodeStatus{
	Online:  true,
	State:   "Activo",
	Version: "v0.9.4",
	Uptime:  "14d 03:22",
	Peers: []Peer{
		{ID: "vrd-a1b2-c3d4", Org: "Servicio Nacional X", Latency: 18, TX: 142, Status: "ok"},
		{ID: "vrd-e5f6-g7h8", Org: "Municipalidad Norte", Latency: 45, TX: 88, Status: "ok"},
		{ID: "vrd-i9j0-k1l2", Org: "Ministerio Salud", Latency: 23, TX: 210, Status: "ok"},
		{ID: "vrd-m3n4-o5p6", Org: "Gobierno Regional", Latency: 120, TX: 12, Status: "degraded"},
		{ID: "vrd-q7r8-s9t0", Org: "Registro Civil", Latency: 31, TX: 190, Status: "ok"},
	},
}

// generateSeries crea una serie de floats con variación aleatoria pseudo-determinista
func generateSeries(length int, base, variance float64) []float64 {
	s := make([]float64, length)
	v := base
	for i := range s {
		// variación simple sin import math/rand para mantenerlo liviano
		delta := variance * float64((i*7+3)%10-5) / 5.0
		v = base + delta
		if v < 0 {
			v = base * 0.1
		}
		s[i] = v
	}
	return s
}

func generateLabels(n int) []string {
	labels := make([]string, n)
	now := time.Now()
	for i := range labels {
		t := now.Add(-time.Duration(n-i) * 30 * time.Minute)
		labels[i] = t.Format("15:04")
	}
	return labels
}

func GetMetrics(period string) NetworkMetrics {
	n := 24
	switch period {
	case "6h":
		n = 12
	case "7d":
		n = 28
	}
	return NetworkMetrics{
		InboundBw:      "2.4 MB/s",
		InboundDelta:   "12% vs hora anterior",
		OutboundBw:     "1.8 MB/s",
		OutboundDelta:  "3% vs hora anterior",
		TxTotal:        "48,291",
		TxDelta:        "8% vs ayer",
		ErrorCount:     "7",
		ErrorDelta:     "2 vs hora anterior",
		LatencyP50:     "24ms",
		LatencyDelta:   "1ms mejor",
		Labels:         generateLabels(n),
		InboundSeries:  generateSeries(n, 2.4, 0.8),
		OutboundSeries: generateSeries(n, 1.8, 0.6),
		TxSeries:       generateSeries(n, 2000, 400),
		LatP50Series:   generateSeries(n, 24, 6),
		LatP95Series:   generateSeries(n, 55, 12),
	}
}

// ProvidedResources es mutable (POST/PUT/DELETE/PATCH lo modifican)
var ProvidedResources = []Resource{
	{ID: 1, Name: "API Ciudadanos", Type: "api", Path: "/api/v1/ciudadanos", Version: "v1.2", Description: "Consulta datos de ciudadanos registrados", Consumers: 8, Enabled: true, Auth: "token", Visibility: "public", RateLimit: 500, SLA: 99.9, BackendURL: "http://svc-ciudadanos:3000", Format: "json", Tags: "ciudadanos,identidad", Contact: "api@mi-org.cl", DocsURL: "https://docs.mi-org.cl/ciudadanos"},
	{ID: 2, Name: "Tópico Eventos", Type: "topic", Path: "vrdr://eventos.municipales", Version: "v1.0", Description: "Stream de eventos municipales", Consumers: 3, Enabled: true, Auth: "mtls", Visibility: "restricted", RateLimit: 0, SLA: 99.5, Retention: 24, Tags: "eventos,municipal", Contact: "ops@mi-org.cl", DocsURL: ""},
	{ID: 3, Name: "Catastro Predial", Type: "file", Path: "/files/catastro-2024.csv", Version: "v2.1", Description: "Dataset catastro predial actualizado", Consumers: 5, Enabled: true, Auth: "token", Visibility: "public", RateLimit: 100, SLA: 98.0, Tags: "catastro,predios", Contact: "datos@mi-org.cl", DocsURL: ""},
	{ID: 4, Name: "BD Permisos", Type: "db", Path: "vrdr://db/permisos-edif", Version: "v1.0", Description: "Base replicable de permisos de edificación", Consumers: 2, Enabled: false, Auth: "mtls", Visibility: "private", RateLimit: 0, SLA: 99.0, Tags: "permisos,edificacion", Contact: "db@mi-org.cl", DocsURL: ""},
	{ID: 5, Name: "Query Licencias", Type: "stream", Path: "vrdr://qs/licencias", Version: "v0.9", Description: "Stream de consultas sobre licencias", Consumers: 4, Enabled: true, Auth: "token", Visibility: "restricted", RateLimit: 200, SLA: 97.5, Tags: "licencias", Contact: "api@mi-org.cl", DocsURL: ""},
	{ID: 6, Name: "API Pagos", Type: "api", Path: "/api/v2/pagos", Version: "v2.0", Description: "Procesamiento de pagos municipales", Consumers: 12, Enabled: true, Auth: "mtls", Visibility: "restricted", RateLimit: 1000, SLA: 99.9, BackendURL: "http://svc-pagos:4000", Format: "json", Tags: "pagos,finanzas", Contact: "pagos@mi-org.cl", DocsURL: "https://docs.mi-org.cl/pagos"},
	{ID: 7, Name: "Arch. Ordenanzas", Type: "file", Path: "/files/ordenanzas/", Version: "v1.0", Description: "Colección de ordenanzas municipales", Consumers: 1, Enabled: true, Auth: "none", Visibility: "public", RateLimit: 50, SLA: 95.0, Tags: "ordenanzas,legal", Contact: "legal@mi-org.cl", DocsURL: ""},
}

var nextResourceID = 8 // contador simple para IDs nuevos

func NextResourceID() int {
	nextResourceID++
	return nextResourceID
}

var ExternalResources = []ExternalResource{
	{ID: 1, Name: "API Identidad Nacional", Type: "api", Org: "Registro Civil", Path: "/api/v2/identidad", SLA: 99.9, Latency: 18, Auth: "mTLS", Subscribed: true, Country: "Chile"},
	{ID: 2, Name: "Tópico Alertas Salud", Type: "topic", Org: "Ministerio Salud", Path: "vrdr://alertas-salud", SLA: 99.5, Latency: 22, Auth: "Token", Subscribed: true, Country: "Chile"},
	{ID: 3, Name: "Dataset Contribuyentes", Type: "file", Org: "SII", Path: "/files/contribuyentes.csv", SLA: 98.0, Latency: 45, Auth: "Token", Subscribed: false, Country: "Chile"},
	{ID: 4, Name: "BD Registro Propiedad", Type: "db", Org: "CBR Santiago", Path: "vrdr://db/propiedades", SLA: 99.0, Latency: 31, Auth: "mTLS", Subscribed: false, Country: "Chile"},
	{ID: 5, Name: "Stream Transacciones OIRS", Type: "stream", Org: "Gobierno Regional", Path: "vrdr://qs/oirs", SLA: 97.5, Latency: 55, Auth: "Token", Subscribed: true, Country: "Chile"},
	{ID: 6, Name: "API RENAPER Argentina", Type: "api", Org: "Agencia Digital AR", Path: "/api/v1/personas", SLA: 99.2, Latency: 120, Auth: "Token", Subscribed: false, Country: "Argentina"},
	{ID: 7, Name: "API CURP Colombia", Type: "api", Org: "MinTIC Colombia", Path: "/api/v1/curp", SLA: 98.5, Latency: 95, Auth: "mTLS", Subscribed: false, Country: "Colombia"},
	{ID: 8, Name: "Tópico Sismos Chile", Type: "topic", Org: "CSN", Path: "vrdr://sismos", SLA: 99.8, Latency: 12, Auth: "none", Subscribed: true, Country: "Chile"},
}

var TopoData = Topology{
	Nodes: []MapNode{
		{ID: "center", Label: "Mi Nodo", Name: "Mi Nodo (vrd-3a7f)", Icon: "⬡", IsCenter: true, Type: "center", OutTx: 0, InTx: 0, Latency: 0, Resources: 7},
		{ID: "n1", Label: "Reg. Civil", Name: "Registro Civil", Icon: "🏛", IsCenter: false, Type: "gov", OutTx: 142, InTx: 88, Latency: 18, Resources: 9},
		{ID: "n2", Label: "Min. Salud", Name: "Ministerio Salud", Icon: "🏥", IsCenter: false, Type: "gov", OutTx: 210, InTx: 55, Latency: 23, Resources: 14},
		{ID: "n3", Label: "SII", Name: "SII - Impuestos Int.", Icon: "💰", IsCenter: false, Type: "gov", OutTx: 190, InTx: 120, Latency: 31, Resources: 18},
		{ID: "n4", Label: "Municipalidad", Name: "Municipalidad Santiago", Icon: "🏙", IsCenter: false, Type: "regional", OutTx: 88, InTx: 200, Latency: 45, Resources: 6},
		{ID: "n5", Label: "Gob. Regional", Name: "Gobierno Regional", Icon: "🗺", IsCenter: false, Type: "regional", OutTx: 12, InTx: 65, Latency: 120, Resources: 3},
		{ID: "n6", Label: "Ag. Digital", Name: "Agencia Digital Argentina", Icon: "🌐", IsCenter: false, Type: "gov", OutTx: 30, InTx: 20, Latency: 120, Resources: 5},
		{ID: "n7", Label: "SERVIU", Name: "SERVIU RM", Icon: "🏗", IsCenter: false, Type: "regional", OutTx: 45, InTx: 180, Latency: 38, Resources: 8},
		{ID: "n8", Label: "CSN", Name: "Centro Sismológico Nacional", Icon: "📡", IsCenter: false, Type: "gov", OutTx: 500, InTx: 5, Latency: 12, Resources: 2},
	},
}
