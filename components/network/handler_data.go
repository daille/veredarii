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

import (
	"Veredarii/util"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
)

func DataHandler(s network.Stream) {
	defer s.Close()
	if !allowedForProto["/data/1.0.0"][s.Conn().RemotePeer()] {

		// Leer un mensaje (len+payload), procesar y responder
		payload, err := util.ReadMsg(s)
		if err != nil {
			log.Printf("handler: error leyendo msg: %v", err)
			_ = s.Reset()
			return
		}
		log.Printf("📩 Recibidos %d bytes desde %s", len(payload), s.Conn().RemotePeer())

		// Aquí haces lo que quieras con 'payload'.
		// Por ejemplo, solo lo devolvemos (echo) para que sea genérico.
		payload = []byte("Este fue tu mensaje original -> " + string(payload))
		if err := util.WriteMsg(s, payload); err != nil {
			log.Printf("handler: error respondiendo: %v", err)
			return
		}
	}
}
