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
	components "Veredarii/components/network"
	server "Veredarii/components/server"
	"Veredarii/general"
	"os"
	"time"
)

type InteropControllerType struct {
	ChSigs chan os.Signal
	ChInit chan bool
}

func NewInteropController() *InteropControllerType {
	ic := &InteropControllerType{
		ChSigs: make(chan os.Signal, 1),
		ChInit: make(chan bool),
	}

	go ic.LoadConfiguration()
	go ic.Servers()
	go ic.NetworkConnection()

	return ic
}

func (IC InteropControllerType) Start() {
	general.Chan.LoadConfiguration <- true
	time.Sleep(1 * time.Second)
	general.Chan.StartNetwork <- true
	general.Chan.StartLocalServer <- true
}

func (IC InteropControllerType) Servers() {
	server.CreateLocalServer()
}

func (IC InteropControllerType) LoadConfiguration() {
	for {
		<-general.Chan.LoadConfiguration
		general.Load("configuration.json")
	}
}

func (IC InteropControllerType) NetworkConnection() {
	components.InitNetworks()
}
