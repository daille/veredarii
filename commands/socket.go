package cmd

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
	"Veredarii/configuration"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
)

type Mensaje struct {
	Entrada []string
	Salida  string
}

func SocketListener() {
	socketPath := "./veredarii.sock"
	if runtime.GOOS == "windows" {
		socketPath = `veredarii.sock`
	}
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}
	os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error Listen:", err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Servidor esperando mensajes...")

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			// --- BLOQUE DE SEGURIDAD ---
			// 2. Verificar que el proceso que conecta sea el mismo usuario (UID)
			unixConn := c.(*net.UnixConn)
			rawConn, err := unixConn.SyscallConn()
			if err != nil {
				return
			}

			var isAuthorized bool
			rawConn.Control(func(fd uintptr) {
				// Obtener credenciales del cliente (Linux)
				creds, err := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
				if err == nil {
					// Solo permitimos si el que conecta tiene nuestro mismo UID
					if int(creds.Uid) == os.Getuid() {
						isAuthorized = true
					}
				}
			})

			if !isAuthorized {
				fmt.Println("Conexión rechazada: Usuario no autorizado")
				return
			}
			// --- FIN BLOQUE DE SEGURIDAD ---

			var m, respuesta Mensaje
			decoder := gob.NewDecoder(c)
			if err := decoder.Decode(&m); err == nil {
				switch m.Entrada[0] {
				case "set":
					configuration.DefaultNetwork = m.Entrada[1]
					respuesta = Mensaje{Salida: "Red fijada a: " + m.Entrada[1]}
				case "get":
					respuesta = Mensaje{Salida: configuration.DefaultNetwork}
				case "pivot":
					switch m.Entrada[1] {
					case "add":
						respuesta = pivotAdd(m.Entrada[2], m.Entrada[3])
					case "remove":
						respuesta = pivotRemove(m.Entrada[2], m.Entrada[3])
					default:
						respuesta = Mensaje{Salida: "Comando no reconocido"}
					}
				case "resource":
					switch m.Entrada[1] {
					case "add":
						respuesta = resourceAdd(m.Entrada[2], m.Entrada[3], m.Entrada[4], m.Entrada[5])
					case "remove":
						respuesta = resourceRemove(m.Entrada[2], m.Entrada[3], m.Entrada[4])
					case "allow":
						respuesta = Allow(m.Entrada[2], m.Entrada[3], m.Entrada[4], m.Entrada[5], m.Entrada[6])
					case "deny":
						respuesta = Deny(m.Entrada[2], m.Entrada[3], m.Entrada[4], m.Entrada[5])
					default:
						respuesta = Mensaje{Salida: "Comando no reconocido"}
					}
				case "external":
					switch m.Entrada[1] {
					case "add":
						respuesta = resourceExternalAdd(m.Entrada[2], m.Entrada[3], m.Entrada[4], m.Entrada[5])
					case "remove":
						respuesta = resourceExternalRemove(m.Entrada[2], m.Entrada[3], m.Entrada[4])
					default:
						respuesta = Mensaje{Salida: "Comando no reconocido"}
					}
				case "connections":
					respuesta = listPeers()
				}

				encoder := gob.NewEncoder(c)
				encoder.Encode(respuesta)
			}
		}(conn)
	}
}

func socketClient(msg Mensaje) Mensaje {
	socketPath := "./veredarii.sock"
	if runtime.GOOS == "windows" {
		socketPath = `veredarii.sock`
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return Mensaje{Salida: ""}
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		fmt.Println("Error al codificar mensaje:", err)
		return Mensaje{}
	}
	var respuesta Mensaje
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&respuesta); err != nil {
		fmt.Println("Error al decodificar respuesta:", err)
		return Mensaje{}
	}

	return respuesta
}
