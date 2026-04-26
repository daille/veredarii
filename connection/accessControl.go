package connection

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
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"
)

var PolicySet *cedar.PolicySet
var Entities cedar.EntityMap

func SetupAccessControl() {
	PolicySet = cedar.NewPolicySet()
	files, err := filepath.Glob("./*.policy")
	if err != nil {
		slog.Error("Error leyendo políticas", "error", err)
	}

	for i, file := range files {
		policyBytes, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Error leyendo política", "file", file, "error", err)
		}

		cleanPolicy := strings.ReplaceAll(string(policyBytes), "\u00a0", " ")

		policyFromDB := strings.TrimSpace(cleanPolicy)
		if policyFromDB != "" {
			var p cedar.Policy
			if err := p.UnmarshalCedar([]byte(policyFromDB)); err != nil {
				slog.Error("Error parseando política", "file", file, "error", err)
			} else {
				policyID := cedar.PolicyID(fmt.Sprintf("policy%d", i))
				PolicySet.Add(policyID, &p)
			}
		} else {
			defaultDeny := `forbid(principal, action, resource);`
			var p cedar.Policy
			_ = p.UnmarshalCedar([]byte(defaultDeny))
			PolicySet.Add("default_deny", &p)
			slog.Info("⚠️ Sistema operando en modo 'Safe Deny': No se encontraron políticas.")
		}
	}

	entitiesBytes, err := os.ReadFile("red_interoperabilidad.entities")
	if err != nil {
		slog.Error("Error leyendo entidades", "error", err)
	}
	json.Unmarshal(entitiesBytes, &Entities)
}

func Authorize(usuario, protocolo, accion, recurso, ip string) bool {
	req := cedar.Request{
		Principal: cedar.EntityUID{Type: "User", ID: types.String(usuario)},
		Action:    cedar.EntityUID{Type: "Action", ID: types.String(accion)},
		Resource:  cedar.EntityUID{Type: types.EntityType(NormalizeCedarType(protocolo)), ID: types.String(recurso)},
		Context: cedar.NewRecord(cedar.RecordMap{
			"source_ip": types.String(ip),
			"now":       types.Long(time.Now().Unix()),
		}),
	}

	decision, _ := cedar.Authorize(PolicySet, Entities, req)

	return (decision == cedar.Allow)
}

func NormalizeCedarType(input string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	normalized := reg.ReplaceAllString(input, "_")
	matchNumber := regexp.MustCompile(`^[0-9]`)
	if matchNumber.MatchString(normalized) {
		normalized = "_" + normalized
	}
	multiDash := regexp.MustCompile(`_+`)
	normalized = multiDash.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		return "UnknownType"
	}
	return normalized
}

func GetTPSLimit(userID string, resourceID string) int64 {
	userUID := types.EntityUID{
		Type: types.EntityType("User"),
		ID:   types.String(userID),
	}

	userEntity, ok := Entities[userUID]
	if !ok {
		return 0
	}

	quotasAttr, ok := userEntity.Attributes.Get("quotas")
	if !ok {
		return 0
	}

	quotasRecord, ok := quotasAttr.(types.Record)
	if !ok {
		return 0
	}

	tpsVal, ok := quotasRecord.Get(types.String(resourceID))
	if !ok {
		return 0
	}

	if tps, ok := tpsVal.(types.Long); ok {
		return int64(tps)
	}
	return 0
}

// AddPrincipal agrega un nuevo User al entities.json
func AddPrincipalAccessControl(network, entidad string) bool {
	entitiesFile := fmt.Sprintf("./%s.entities", network)

	entidad = strings.ToLower(entidad)

	data, err := os.ReadFile(entitiesFile)
	if err != nil {
		return false
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(data, &entities); err != nil {
		return false
	}

	// Verificar si ya existe
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
		if uid["type"] == "User" && existingID == entidad {
			return false
		}
	}

	entities = append(entities, map[string]interface{}{
		"uid":     map[string]interface{}{"type": "User", "id": entidad},
		"attrs":   map[string]interface{}{"quotas": map[string]interface{}{}},
		"parents": []interface{}{},
	})

	out, _ := json.MarshalIndent(entities, "", "    ")
	if err := os.WriteFile(entitiesFile, out, 0644); err != nil {
		return false
	}

	SetupAccessControl()
	return true
}

// AddResource agrega un nuevo recurso al entities.json
func AddResourceAccessControl(protocolo, servicio, network string) bool {
	entitiesFile := fmt.Sprintf("./%s.entities", network)

	servicio = strings.ToLower(servicio)

	data, err := os.ReadFile(entitiesFile)
	if err != nil {
		return false
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(data, &entities); err != nil {
		return false
	}

	// Verificar si ya existe
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
		existingType := strings.ToLower(fmt.Sprintf("%v", uid["type"]))
		if existingType == strings.ToLower(protocolo) && existingID == servicio {
			return false
		}
	}

	entities = append(entities, map[string]interface{}{
		"uid":     map[string]interface{}{"type": protocolo, "id": servicio},
		"attrs":   map[string]interface{}{"network": network},
		"parents": []interface{}{},
	})

	out, _ := json.MarshalIndent(entities, "", "    ")
	if err := os.WriteFile(entitiesFile, out, 0644); err != nil {
		return false
	}

	SetupAccessControl()
	return true
}

// RemovePrincipal elimina un User del entities.json y todas sus policies del .policy
func RemovePrincipalAccessControl(network, entidad string) bool {
	policyFile := fmt.Sprintf("./%s.policy", network)
	entitiesFile := fmt.Sprintf("./%s.entities", network)

	entidad = strings.ToLower(entidad)

	// --- 1. Eliminar todas las policies del usuario ---
	policyBytes, err := os.ReadFile(policyFile)
	if err != nil {
		return false
	}

	principalClause := fmt.Sprintf(`User::"%s"`, entidad)
	blocks := strings.Split(string(policyBytes), "};")
	var kept []string
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			continue
		}
		if strings.Contains(block, principalClause) {
			continue // eliminar todos los bloques de este usuario
		}
		kept = append(kept, trimmed)
	}

	newPolicyContent := strings.Join(kept, "\n};\n")
	if len(kept) > 0 {
		newPolicyContent += "\n};"
	}

	if err := os.WriteFile(policyFile, []byte(newPolicyContent+"\n"), 0644); err != nil {
		return false
	}

	// --- 2. Eliminar el User del entities.json ---
	data, err := os.ReadFile(entitiesFile)
	if err != nil {
		return false
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(data, &entities); err != nil {
		return false
	}

	found := false
	var filtered []map[string]interface{}
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
		if uid["type"] == "User" && existingID == entidad {
			found = true
			continue // eliminar
		}
		filtered = append(filtered, e)
	}

	if !found {
		return false
	}

	out, _ := json.MarshalIndent(filtered, "", "    ")
	if err := os.WriteFile(entitiesFile, out, 0644); err != nil {
		return false
	}

	SetupAccessControl()
	return true
}

// RemoveResource elimina un recurso del entities.json y todas sus policies del .policy
func RemoveResourceAccessControl(network, protocolo, servicio string) bool {
	policyFile := fmt.Sprintf("./%s.policy", network)
	entitiesFile := fmt.Sprintf("./%s.entities", network)

	servicio = strings.ToLower(servicio)

	// --- 1. Eliminar todas las policies que referencien este recurso ---
	policyBytes, err := os.ReadFile(policyFile)
	if err != nil {
		return false
	}

	resourceClause := fmt.Sprintf(`%s::"%s"`, protocolo, servicio)
	blocks := strings.Split(string(policyBytes), "};")
	var kept []string
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			continue
		}
		if strings.Contains(block, resourceClause) {
			continue // eliminar bloques que referencien este recurso
		}
		kept = append(kept, trimmed)
	}

	newPolicyContent := strings.Join(kept, "\n};\n")
	if len(kept) > 0 {
		newPolicyContent += "\n};"
	}

	if err := os.WriteFile(policyFile, []byte(newPolicyContent+"\n"), 0644); err != nil {
		return false
	}

	// --- 2. Eliminar el recurso del entities.json ---
	data, err := os.ReadFile(entitiesFile)
	if err != nil {
		return false
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(data, &entities); err != nil {
		return false
	}

	found := false
	var filtered []map[string]interface{}
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
		existingType := strings.ToLower(fmt.Sprintf("%v", uid["type"]))
		if existingType == strings.ToLower(protocolo) && existingID == servicio {
			found = true
			continue // eliminar
		}
		filtered = append(filtered, e)
	}

	if !found {
		return false
	}

	// --- 3. Limpiar quotas del recurso en todos los usuarios ---
	for _, e := range filtered {
		uid := e["uid"].(map[string]interface{})
		if uid["type"] == "User" {
			if attrs, ok := e["attrs"].(map[string]interface{}); ok {
				if quotas, ok := attrs["quotas"].(map[string]interface{}); ok {
					delete(quotas, servicio)
				}
			}
		}
	}

	out, _ := json.MarshalIndent(filtered, "", "    ")
	if err := os.WriteFile(entitiesFile, out, 0644); err != nil {
		return false
	}

	SetupAccessControl()
	return true
}
