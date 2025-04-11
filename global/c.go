package global

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/memberlist"
	"github.com/serialx/hashring"
)

type MyDelegate struct {
	MsgCh      chan []byte
	Broadcasts *memberlist.TransmitLimitedQueue
	Meta       NodoMetaData
	Num        int
	consistent *hashring.HashRing
}

func (d *MyDelegate) NodeMeta(limit int) []byte {
	return d.Meta.Bytes()
}

func (d *MyDelegate) NotifyMsg(msg []byte) {
	d.MsgCh <- msg
}
func (d *MyDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.Broadcasts.GetBroadcasts(overhead, limit)
}

func (d *MyDelegate) LocalState(join bool) []byte {
	// not use, noop
	return []byte("")
}
func (d *MyDelegate) MergeRemoteState(buf []byte, join bool) {
	// not use
}

type MyBroadcastMessage struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Origin string `json:"origin"`
}

func (m MyBroadcastMessage) Invalidates(other memberlist.Broadcast) bool {
	return false
}
func (m MyBroadcastMessage) Finished() {
	// nop
}
func (m MyBroadcastMessage) Message() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		return []byte("")
	}
	return data
}

func ParseMyBroadcastMessage(data []byte) (*MyBroadcastMessage, bool) {
	msg := new(MyBroadcastMessage)
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, false
	}
	return msg, true
}

func WaitSignal(cancel context.CancelFunc) {
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGINT)
	for {
		select {
		case s := <-signal_chan:
			log.Printf("signal %s happen", s.String())
			cancel()
		}
	}
}

func (m NodoMetaData) Bytes() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		return []byte("")
	}
	return data
}
func ParseMyMetaData(data []byte) (NodoMetaData, bool) {
	meta := NodoMetaData{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return meta, false
	}
	return meta, true
}

type MyEventDelegate struct {
	consistent *hashring.HashRing
	Num        int
}

func (d *MyEventDelegate) NotifyJoin(node *memberlist.Node) {
	hostPort := fmt.Sprintf("%s:%d", node.Addr.To4().String(), node.Port)
	meta, ok := ParseMyMetaData(node.Meta)
	if ok != true {
		log.Println("error?")
	}

	log.Debug(fmt.Sprintf("JOIN %s ID: %s, Country: %s, Organization: %s, System: %s",
		node.Name,
		meta.ID,
		meta.Country,
		meta.Organism,
		meta.System,
	))

	if d.consistent == nil {
		d.consistent = hashring.New([]string{hostPort})
	} else {
		d.consistent = d.consistent.AddNode(hostPort)
	}
}

func (d *MyEventDelegate) NotifyLeave(node *memberlist.Node) {
	hostPort := fmt.Sprintf("%s:%d", node.Addr.To4().String(), node.Port)
	log.Printf("leave %s", hostPort)
	if d.consistent != nil {
		d.consistent = d.consistent.RemoveNode(hostPort)
	}
}
func (d *MyEventDelegate) NotifyUpdate(node *memberlist.Node) {
	// skip
}
