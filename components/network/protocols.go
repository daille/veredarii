package components

/**
 *    Interop software for interoperability
 *
 *    @author jcDaille
 *
 *    This file is part of Interop.
 *
 *    Interop is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License.
 *
 *    Interop is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with Foobar.  If not, see <https://www.gnu.org/licenses/>.
 */
import "github.com/libp2p/go-libp2p/core/peer"

var allowedForProto = map[string]map[peer.ID]bool{
	"/chat/1.0.0":  {},
	"/proto/1.0.0": {},
}

func AllowPeerForProtocol(protoID string, id peer.ID) {
	if allowedForProto[protoID] == nil {
		allowedForProto[protoID] = make(map[peer.ID]bool)
	}
	allowedForProto[protoID][id] = true
}

type allowAllValidator struct{}

func (allowAllValidator) Validate(key string, value []byte) error         { return nil }
func (allowAllValidator) Select(key string, values [][]byte) (int, error) { return 0, nil }

type InstitutionInfo struct {
	Name     string   `json:"name"`
	Location string   `json:"location"`
	Roles    []string `json:"roles"`
}
