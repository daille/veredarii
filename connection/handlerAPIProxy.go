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
	"net/url"
	"strings"
	"time"

	"log/slog"

	"Veredarii/configuration"
	global "Veredarii/global"
	pluginmanager "Veredarii/pluginManager"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
)

func (n *Network) handleAPIProxyStream(s network.Stream) {
	slog.Debug(">>>> handleAPIProxyStream", "stream", s.ID())
	defer s.Close()
	remotePeer := s.Conn().RemotePeer()
	PID, err := peer.Decode(remotePeer.String())
	if err != nil {
		slog.Error("Error decodificando peer", "peer", remotePeer.String(), "error", err)
		return
	}
	entidad := n.Peers[PID].Entity

	if !RBAC.HasPermition2Protocol(entidad, n.Name, global.ProtocolAPIProxy) {
		slog.Warn("Denegado, sin permiso al protocolo", "entidad", entidad, "red", n.Name, "protocolo", global.ProtocolAPIProxy)
		s.Reset()
		return
	}

	for {
		msg := &global.Envelop{}
		data, err := readDelimited(s)
		if err != nil {
			if err != io.EOF {
				slog.Error("Error leyendo stream", "error", err.Error())
			}
			return
		}

		if err := proto.Unmarshal(data, msg); err != nil {
			slog.Error("Error unmarshal protobuf", "error", err.Error())
			slog.Error("Error", "data", string(data))
			return
		}
		tps := RBAC.GetTPS(entidad, global.ProtocolAPIProxy, n.Name, msg.Service)
		slog.Debug("TPS registrada", "tps", tps)
		limiter := n.RateLimiter.GetLimiter(entidad, global.ProtocolAPIProxy, tps)

		if !limiter.Allow() {
			slog.Warn("Entidad excedió su límite global de TPS", "entidad", entidad, "tps", tps)
			s.Reset()
			return
		}

		if !RBAC.Allowed(entidad, n.Name, global.ProtocolAPIProxy, msg.Service) {
			slog.Warn("Denegado, sin permiso al servicio: ", "entidad", entidad, "red", n.Name, "protocolo", global.ProtocolAPIProxy, "servicio", msg.Service)
			s.Reset()
			return
		}

		b := bufio.NewReader(bytes.NewReader(msg.Payload))
		req, err := http.ReadRequest(b)
		if err != nil {
			slog.Error("Error reconstruyendo petición:", "error", err.Error())
			return
		}
		fmt.Printf("📩 Recibido de %s: %s\n", s.Conn().RemotePeer().String()[:6], req.URL.String())

		founded := false
		for _, net := range configuration.CM.Config.Networks {
			if net.Name == n.Name {
				for _, service := range n.Resources.API {
					if service.Name == msg.Service {
						founded = true
						var bodyBytes []byte

						if service.Plugin == "" {
							urlDestino := fmt.Sprintf("%s?%s", service.ResourcePath, req.URL.RawQuery)
							slog.Debug("Llamando URL Local", "urlDestino: ", urlDestino)
							u, err := url.Parse(urlDestino)
							if err != nil {
								slog.Error("Error parseando url: ", "error", err.Error())
								return
							}

							proxyReq, err := http.NewRequest(req.Method, urlDestino, req.Body)
							if err != nil {
								slog.Error("Error creando nuevo request:", "error", err.Error())
								return
							}
							proxyReq.Header = req.Header
							proxyReq.Host = u.Host

							client := &http.Client{}
							resp, err := client.Do(proxyReq)
							if err != nil {
								slog.Error("Error replicando el llamado: %v", "error", err.Error())
								return
							}
							bodyBytes, err = io.ReadAll(resp.Body)
							if err != nil {
								slog.Error("Error leyendo el cuerpo de la respuesta: %v", "error", err.Error())
								return
							}
							resp.Body.Close()
						} else {
							slog.Debug("Llamando Plugin", "plugin", service.Plugin)
							ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer cancel()
							bodyBytes, err = pluginmanager.PM.Execute(ctx, service.Plugin, "handle_request", msg.Payload)
							if err != nil {
								slog.Error("Error ejecutando plugin: ", "error", err.Error())
							}
							slog.Debug("Respuesta del plugin", "valor", string(bodyBytes))
						}

						response := &global.Envelop{
							Id:      uuid.New().String(),
							Payload: bodyBytes,
						}

						resData, _ := proto.Marshal(response)
						if _, err := writeDelimited(s, resData); err != nil {
							slog.Error("Error respondiendo: %v", "error", err.Error())
							return
						}
					}
				}
			}
		}
		if !founded {
			slog.Error("No se encontro la red", "red", n.Name)
			s.Reset()
			return
		}
	}
}

func (n *Network) Send(service string, payload []byte) []byte {
	stream, err := n.getStream(global.ProtocolAPIProxy, service)
	if err != nil {
		slog.Error("Error obteniendo stream", "error", err.Error())
		return nil
	}
	s, err := n.Host.NewStream(context.Background(), stream.Target, protocol.ID(stream.Proto))
	if err != nil {
		return nil
	}
	defer s.Close()

	msg := &global.Envelop{
		Id:      uuid.New().String(),
		Service: service,
		Payload: payload,
	}
	data, _ := proto.Marshal(msg)

	_, err = writeDelimited(s, data)
	if err != nil {
		slog.Error("error escribiendo respuesta", "error", err)
		n.MutexStreams.Lock()
		delete(n.Streams[global.ProtocolAPIProxy], service)
		n.MutexStreams.Unlock()
		s.Reset()
		return nil
	}
	resData, err := readDelimited(s)
	if err == nil {
		res := &global.Envelop{}
		proto.Unmarshal(resData, res)
		return res.Payload
	} else {
		slog.Error("error leyendo respuesta", "error", err)
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
	const maxMessageSize = 32 * 1024 * 1024 // 32 MB
	if size > maxMessageSize {
		return nil, fmt.Errorf("mensaje demasiado grande: %d bytes", size)
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
		if strings.Contains(addrStr, "127.0.0.1") || strings.Contains(addrStr, "::1") {
			continue
		}

		if strings.Contains(addrStr, "/ip4/192.168.") ||
			strings.Contains(addrStr, "/ip4/10.") ||
			strings.Contains(addrStr, "/ip4/172.16.") {
			continue
		}

		if strings.Contains(addrStr, "p2p-circuit") {
			continue
		}

		return true
	}
	return false
}
