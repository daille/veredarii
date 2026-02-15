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
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

var RBAC *RBACType

type RBACType struct {
	Enforcer      *casbin.Enforcer
	PeerEntity    map[string]string
	MutexSesiones sync.RWMutex
}

func StartRBAC() {
	var err error
	RBAC = &RBACType{}
	RBAC.Enforcer, err = casbin.NewEnforcer("./model.conf", "./policy.csv")
	RBAC.PeerEntity = make(map[string]string)
	if err != nil {
		log.Fatal("Error cargando RBAC:", err)
		return
	}
}

func (rb *RBACType) Allowed(peerID peer.ID, dom string, obj string, act string) bool {
	if entity := rb.PeerEntity[peerID.String()]; entity == "" {
		log.Debug(fmt.Sprintf("No encontrado: %s %s %s %s", peerID.String(), dom, obj, act))
		return false
	} else {
		if res, _ := rb.Enforcer.Enforce(entity, dom, obj, act); res {
			log.Debug(fmt.Sprintf("Permitido: %s %s %s %s", peerID.String(), dom, obj, act))
			return true
		} else {
			return false
		}
	}
}

func (rb *RBACType) HasPermition2Protocol(peerID peer.ID, dom string, obj string) bool {
	if entity := rb.PeerEntity[peerID.String()]; entity == "" {
		log.Debug(fmt.Sprintf("No encontrado: %s %s %s %s", peerID.String(), dom, obj))
		return false
	} else {
		policies, err := rb.Enforcer.GetFilteredPolicy(0, entity, dom, obj)
		if err != nil {
			log.Error("Error al obtener la lista de políticas: ", err)
			return false
		}
		if len(policies) > 0 {
			return true
		} else {
			return false
		}
	}
}

func (rb *RBACType) SetPeer(rec EntidadRecord) {
	rb.MutexSesiones.Lock()
	defer rb.MutexSesiones.Unlock()
	rb.PeerEntity[rec.PeerID.String()] = rec.EntityName
}
