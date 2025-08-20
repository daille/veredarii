package general

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

var messages = map[string]map[string]string{
	"en": {
		"Loading_configuration":   "Loading configuration...",
		"configuration":           "Configuration",
		"address_to_share":        "Address to share",
		"error_loading_key":       "Impossible to load the key.",
		"error_loading_swarm":     "Impossible to load the swarm key",
		"error_whitelist":         "Impossible to charge the whitelist",
		"configuration_error":     "Configuration error",
		"database_error_starting": "Error starting local database",
		"database_started":        "Local database started",
	},
	"es": {
		"Loading_configuration":   "Cargando configuracion",
		"configuration":           "Configuracion",
		"address_to_share":        "Direcciones para compartir",
		"error_loading_key":       "No fue posible cargar la llave",
		"error_loading_swarm":     "No se pudo abrir la swarm key",
		"error_whitelist":         "No se pudo cargar la lista blanca",
		"configuration_error":     "Error de configuracion",
		"database_error_starting": "Error Inicializando Base de Datos local",
		"database_started":        "Base de datos local inicializada",
	},
}

var currentLang = "es"

func T(key string) string {
	if msg, ok := messages[currentLang][key]; ok {
		return msg
	}
	return messages["en"][key]
}
