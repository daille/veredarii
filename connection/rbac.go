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
	"strconv"
	"sync"

	"log/slog"

	"github.com/casbin/casbin/v2"
)

var RBAC *RBACType

type RBACType struct {
	Enforcer      *casbin.Enforcer
	PeerEntity    map[string]string
	MutexSesiones sync.RWMutex
}

func StartRBAC() {
	slog.Debug("Iniciando RBAC...")
	var err error
	RBAC = &RBACType{}
	RBAC.Enforcer, err = casbin.NewEnforcer("./model.conf", "./policy.csv")
	RBAC.PeerEntity = make(map[string]string)
	if err != nil {
		slog.Error("Error cargando RBAC", "error", err)
		return
	}
}

func (rb *RBACType) Allowed(entity string, dom string, obj string, act string) bool {
	if res, _ := rb.Enforcer.Enforce(entity, dom, obj, act); res {
		slog.Debug("Permitido", "entity", entity, "dom", dom, "obj", obj, "act", act)
		return true
	}
	slog.Debug("Denegado", "entity", entity, "dom", dom, "obj", obj, "act", act)
	return false
}

func (rb *RBACType) GetTPS(entity string, obj string, dom string, act string) int64 {
	ok, rule, err := rb.Enforcer.EnforceEx(entity, dom, obj, act)
	if !ok {
		slog.Debug("No se encontró la política", "entity", entity, "dom", dom, "obj", obj, "act", act)
		return 0
	} else if err != nil {
		slog.Error("Error al obtener la lista de políticas", "error", err)
		return 0
	}

	tps, _ := strconv.ParseInt(rule[4], 10, 64)
	slog.Debug("TPS", "tps", tps, "entity", entity, "dom", dom, "obj", obj, "act", act)
	return tps
}

func (rb *RBACType) HasPermition2Protocol(entity string, dom string, obj string) bool {
	policies, err := rb.Enforcer.GetFilteredPolicy(0, entity, dom, obj)
	if err != nil {
		slog.Error("Error al obtener la lista de políticas", "error", err)
		return false
	}
	if len(policies) > 0 {
		return true
	} else {
		return false
	}
}

func (rb *RBACType) SetPeer(rec EntidadRecord) {
	rb.MutexSesiones.Lock()
	defer rb.MutexSesiones.Unlock()
	rb.PeerEntity[rec.PeerID.String()] = rec.EntityName
}
