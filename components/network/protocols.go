package components

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
