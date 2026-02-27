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
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	global "Veredarii/global"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

func (n *Network) handleAPIProxyStream(s network.Stream) {
	defer s.Close()
	remotePeer := s.Conn().RemotePeer()
	fmt.Println("remotePeer -> handleAPIProxyStream -> ", remotePeer)

	pID := s.Conn().RemotePeer()
	entidad := n.Peers[pID].Entity

	if !RBAC.HasPermition2Protocol(entidad, n.Name, global.ProtocolAPIProxy) {
		log.Debug("Denegado, sin permiso al protocolo: ", entidad, n.Name, global.ProtocolAPIProxy)
		s.Reset()
		return
	}

	for {
		msg := &global.Envelop{}
		data, err := readDelimited(s)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error leyendo stream: %v", err)
			}
			return
		}

		if err := proto.Unmarshal(data, msg); err != nil {
			log.Printf("Error unmarshal protobuf: %v", err)
			return
		}
		tps := RBAC.GetTPS(entidad, global.ProtocolAPIProxy, n.Name, msg.Service)
		log.Debug("TPS: ", tps)
		limiter := n.RateLimiter.GetLimiter(entidad, global.ProtocolAPIProxy, tps)

		if !limiter.Allow() {
			fmt.Printf("⚠️ Entidad '%s' excedió su límite global de %v TPS\n", entidad, tps)
			s.Reset()
			return
		}

		if !RBAC.Allowed(entidad, n.Name, global.ProtocolAPIProxy, msg.Service) {
			log.Debug("Denegado, sin permiso al servicio: ", remotePeer.String(), n.Name, global.ProtocolAPIProxy, msg.Service)
			s.Reset()
			return
		}

		b := bufio.NewReader(bytes.NewReader(msg.Payload))
		req, err := http.ReadRequest(b)
		if err != nil {
			log.Println("Error reconstruyendo petición:", err)
			return
		}
		fmt.Printf("📩 Recibido de %s: %s\n", s.Conn().RemotePeer().String()[:6], req.URL.String())

		nuevoHost := "localhost:3000"
		nuevaRuta := "/echo"
		urlDestino := fmt.Sprintf("http://%s%s?%s", nuevoHost, nuevaRuta, req.URL.RawQuery)
		proxyReq, err := http.NewRequest(req.Method, urlDestino, req.Body)
		if err != nil {
			log.Printf("Error creando nuevo request: %v", err)
			return
		}
		proxyReq.Header = req.Header
		proxyReq.Host = nuevoHost

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Printf("Error replicando el llamado: %v", err)
			return
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error leyendo el cuerpo de la respuesta: %v", err)
			return
		}
		defer resp.Body.Close()

		response := &global.Envelop{
			Id:      uuid.New().String(),
			Payload: bodyBytes,
		}

		resData, _ := proto.Marshal(response)
		if _, err := writeDelimited(s, resData); err != nil {
			log.Printf("Error respondiendo: %v", err)
			return
		}
	}
}

func (n *Network) Conversar(targetID peer.AddrInfo, service string, payload []byte) []byte {
	s, err := n.Host.NewStream(context.Background(), targetID.ID, global.ProtocolAPIProxy)
	if err != nil {
		log.Printf("Error abriendo stream: %v", err)
		return nil
	}
	defer s.Close()

	msg := &global.Envelop{
		Id:      uuid.New().String(),
		Service: service,
		Payload: payload,
	}
	data, _ := proto.Marshal(msg)
	writeDelimited(s, data)

	resData, err := readDelimited(s)
	if err == nil {
		res := &global.Envelop{}
		proto.Unmarshal(resData, res)
		fmt.Printf("🔄 Respuesta del nodo: %s\n", res.Payload)
		return res.Payload
	}

	return nil
}

func writeDelimited(w io.Writer, data []byte) (int, error) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(data)))
	if _, err := w.Write(buf[:n]); err != nil {
		return 0, err
	}
	return w.Write(data)
}

func readDelimited(r io.Reader) ([]byte, error) {
	br := bufio.NewReader(r)
	size, err := binary.ReadUvarint(br)
	if err != nil {
		return nil, err
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(br, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (n *Network) IsPublic() bool {
	addrs := n.Host.Addrs()
	for _, addr := range addrs {
		addrStr := addr.String()

		// 1. Ignoramos lo local
		if strings.Contains(addrStr, "127.0.0.1") || strings.Contains(addrStr, "::1") {
			continue
		}

		// 2. Ignoramos IPs privadas típicas de router/NAT
		if strings.Contains(addrStr, "/ip4/192.168.") ||
			strings.Contains(addrStr, "/ip4/10.") ||
			strings.Contains(addrStr, "/ip4/172.16.") { // (puedes añadir más si quieres)
			continue
		}

		// 3. Ignoramos si es una dirección de Relay (p2p-circuit)
		// Porque si usas relay, no eres "realmente" público, el pivote te está ayudando
		if strings.Contains(addrStr, "p2p-circuit") {
			continue
		}

		// Si llegamos aquí y hay una dirección IP4 o IP6, ¡es externa!
		return true
	}
	return false
}
