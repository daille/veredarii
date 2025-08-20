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
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadMsg(r io.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	if n == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func WriteMsg(w io.Writer, payload []byte) error {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func BuildPayload(dataArg, dataFile string) ([]byte, error) {
	// 1) Si hay archivo, lo usamos
	if strings.TrimSpace(dataFile) != "" {
		b, err := os.ReadFile(dataFile)
		if err != nil {
			return nil, fmt.Errorf("leyendo archivo: %w", err)
		}
		return b, nil
	}
	// 2) Si hay --data, lo usamos
	if dataArg != "" {
		return []byte(dataArg), nil
	}
	// 3) Si no hay nada, leemos de STDIN (permite tuberías)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		br := bufio.NewReader(os.Stdin)
		return io.ReadAll(br)
	}
	// Si todo vacío, mandamos algo ejemplo
	return []byte("hola desde libp2p"), nil
}
