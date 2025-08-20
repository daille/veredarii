package util

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

import (
	"Veredarii/general"
	"crypto/rand"
	"encoding/base64"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/pnet"
	log "github.com/sirupsen/logrus"
)

func LoadOrCreateKey(filename string) (crypto.PrivKey, error) {
	if _, err := os.Stat(filename); err == nil {
		// Cargar
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		raw, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
		return crypto.UnmarshalPrivateKey(raw)
	}

	// Generar nueva (Ed25519)
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	raw, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	enc := base64.StdEncoding.EncodeToString(raw)
	if err := os.WriteFile(filename, []byte(enc), 0o600); err != nil {
		return nil, err
	}
	return priv, nil
}

func LoadSwarmKey(filename string) (pnet.PSK, bool) {
	swarmKey, err := os.Open(filename)
	if err != nil {
		log.Debug(general.T("error_loading_swarm"), ":", err)
		return nil, false
	}
	defer swarmKey.Close()

	psk, err := pnet.DecodeV1PSK(swarmKey)
	if err != nil {
		log.Debug(general.T("error_loading_swarm"), ":", err)
		return nil, false
	}

	return psk, true
}
