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
	global "Veredarii/global"

	log "github.com/sirupsen/logrus"
)

var NM *NetworkManager

type NetworkManager struct {
	Networks        map[string]*Network
	ChannelNetworks chan string
}

func NewNetworkManager() *NetworkManager {
	N := &NetworkManager{
		Networks: make(map[string]*Network),
	}

	N.ChannelNetworks = make(chan string)

	return N
}

func (nm *NetworkManager) StartProcess() {
	select {
	case order := <-nm.ChannelNetworks:
		switch order {
		case "init":
			log.Debug("Iniciando redes...")
			StartRBAC()
			for _, network := range nm.Networks {
				log.Debug("Iniciando red: ", network.Name)
				network.Connect()
			}
		}
	}
}

func (nm *NetworkManager) AddNetwork(network global.NetworkType) {
	nm.Networks[network.Name] = NewNetwork(
		network.Name,
		network.Port,
		network.NetworkKey,
		network.Pivots,
		network.MyAddress,
		network.Topics,
		network.Entities,
		network.Resources,
		network.RemoteResources,
	)
}

func (nm *NetworkManager) GetNetwork(name string) (*Network, bool) {
	if nm.Networks == nil {
		return nil, false
	}
	if network, ok := nm.Networks[name]; ok {
		return network, true
	}
	return nil, false
}
