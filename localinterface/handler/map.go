package handler

import (
	"fmt"
	"net/http"

	"Veredarii/localinterface/data"
	mw "Veredarii/localinterface/middleware"
)

// ─── GET /api/map/topology ────────────────────────────────────────────────────

func GetTopology(w http.ResponseWriter, r *http.Request) {
	mw.JSON(w, http.StatusOK, data.TopoData)
}

// ─── GET /api/map/flows?period=24h ───────────────────────────────────────────

func GetFlows(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	// Totales cambian según el período
	multipliers := map[string]int{"1h": 1, "24h": 24, "7d": 168}
	mult, ok := multipliers[period]
	if !ok {
		mult = 24
	}

	flows := data.Flows{
		TotalOut: fmt.Sprintf("%s", formatTx(1217*mult)),
		TotalIn:  fmt.Sprintf("%s", formatTx(733*mult)),
		Period:   period,
	}
	mw.JSON(w, http.StatusOK, flows)
}

func formatTx(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
