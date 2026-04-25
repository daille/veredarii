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
	"log"
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
		log.Fatal(err)
	}

	for i, file := range files {
		policyBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error leyendo %s: %v", file, err)
		}

		cleanPolicy := strings.ReplaceAll(string(policyBytes), "\u00a0", " ")

		var p cedar.Policy
		if err := p.UnmarshalCedar([]byte(cleanPolicy)); err != nil {
			log.Fatalf("Error parseando %s: %v", file, err)
		}

		policyID := cedar.PolicyID(fmt.Sprintf("policy%d", i))
		PolicySet.Add(policyID, &p)
	}

	entitiesBytes, err := os.ReadFile("red_interoperabilidad.entities")
	if err != nil {
		log.Fatal(err)
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
