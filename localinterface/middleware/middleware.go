// Package middleware provee utilidades transversales para los handlers.
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// JSON escribe una respuesta JSON con el status code dado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}

// Error escribe una respuesta de error en formato estándar { "message": "..." }.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"message": msg})
}

// DecodeBody decodifica el body JSON de la request en v.
// Retorna false y escribe el error si falla.
func DecodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		Error(w, http.StatusBadRequest, "body JSON inválido: "+err.Error())
		return false
	}
	return true
}

// Logger es un middleware que imprime cada request con método, path y duración.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		fmt.Printf("[%s] %s %s → %d (%s)\n",
			time.Now().Format("15:04:05"),
			r.Method,
			r.URL.Path,
			wrapped.status,
			time.Since(start),
		)
	})
}

// responseWriter captura el status code para poder loguearlo.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
