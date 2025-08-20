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
	"context"
	"fmt"

	"Veredarii/general"
	"Veredarii/util"

	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	log "github.com/sirupsen/logrus"
)

var NetworkConnectionList []*NetworkConnectionType

var ctx context.Context

func InitNetworks() {
	ctx = context.Background()

	for {
		log.Debug("Esperando por inicializar network")
		select {
		case <-general.Chan.StartNetwork:
			log.Debug("Starting networks...")
			peerKey, err := util.LoadOrCreateKey(general.Configuration.Identification.Peer)
			if err != nil {
				log.Error(general.T("error_loading_key"), " ", err)
			}

			for _, network := range general.Configuration.Networks {
				networkConn := NewNetworkConnection(network)
				NetworkConnectionList = append(NetworkConnectionList, networkConn)
				ok := networkConn.CreateHost(peerKey)

				if ok {
					fmt.Println("Peer ID:", networkConn.Host.ID())
					pingService := &ping.PingService{Host: networkConn.Host}
					networkConn.Host.SetStreamHandler(ping.ID, pingService.PingHandler)
					networkConn.Host.SetStreamHandler("/data/1.0.0", DataHandler)
					networkConn.Host.SetStreamHandler("/miapp/proto/1.0.0", ProtoHandler)
					networkConn.PrintAddress()
				}
			}

		case <-general.Chan.StopNetwork:
			log.Debug("Stoping networks...")
			for _, networkConn := range NetworkConnectionList {
				networkConn.Host.Close()
			}
			NetworkConnectionList = []*NetworkConnectionType{}
		}
	}
}
