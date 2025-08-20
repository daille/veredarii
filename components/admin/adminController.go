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
	"bufio"
	"io"
	"net"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	StopCharacter         = "\r\n\r\n"
	SocketAddress         = "./admin_terminal.socket"
	SocketAddressWindows8 = ":3333"
)

func SocketServer() {
	if err := os.RemoveAll(SocketAddress); err != nil {
		log.Error(err)
	}
	listen, err := net.Listen("unix", SocketAddress)
	if err != nil {
		listen, err = net.Listen("tcp4", SocketAddressWindows8)
		if err != nil {
			log.Error("Socket listen failed.", err)
			return
		}
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		SocketHandler(conn)
	}
}

func SocketHandler(conn net.Conn) {
	var (
		buf = make([]byte, 16384)
		r   = bufio.NewReader(conn)
		w   = bufio.NewWriter(conn)
	)

	command1 := regexp.MustCompile(`^StopNodo (?P<llave>[a-z0-9]*)`)
	command2 := regexp.MustCompile(`^ForceServiceDiscoveryUpdate`)
	defer conn.Close()

EOF:
	for {
		r, err := r.Read(buf)
		data := string(buf[:r])
		message := strings.Trim(data, "\t\n\v\f\r ")

		switch err {
		case io.EOF:
			break EOF
		case nil:
			if len(command1.FindStringSubmatch(message)) > 0 {

			} else if len(command2.FindStringSubmatch(message)) > 0 {

			}

			if strings.HasSuffix(data, "\r\n\r\n") {
				break EOF
			}
		default:
			return
		}
	}

	w.Write([]byte("return"))
	w.Flush()
}
