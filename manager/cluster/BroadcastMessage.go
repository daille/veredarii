package cluster

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
	"NodoCb/global"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BroadcastMessageType struct {
	Body BodyBroadcastMessageType
	Node global.NodoMetaData
}

type BodyBroadcastMessageType struct {
	Meta struct {
		Key       string `json:"key"`
		Timestamp string `json:"timestamp"`
		Origin    string `json:"origin"`
		Destiny   string `json:"destiny"`
	} `json:"meta"`
	Update struct {
		Topic   string                 `json:"topic"`
		Action  string                 `json:"action"`
		Network global.ClusterDataType `json:"network,omitempty"`
		Member  global.MemberType      `json:"member,omitempty"`
		Catalog struct {
			Service struct {
				ID          string   `json:"id"`
				Description string   `json:"description"`
				Endpoint    []string `json:"endpoint"`
				Tags        []string `json:"tags"`
				Consumer    []string `json:"consumer"`
			} `json:"service"`
		} `json:"catalog,omitempty"`
	} `json:"update"`
}

func NewBroadcastMessage(body BodyBroadcastMessageType, node global.NodoMetaData) *BroadcastMessageType {
	bm := &BroadcastMessageType{
		Body: body,
		Node: node,
	}
	return bm
}

func (BM *BroadcastMessageType) CreateString() string {
	BM.Body.Meta.Origin = BM.Node.Organism
	BM.Body.Meta.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)

	if value, err := json.Marshal(BM.Body); err == nil {
		return string(value)
	} else {
		log.Error("ERR: ", err)
		return ""
	}
}

func (BM *BroadcastMessageType) SetString(json string) {

}

func (BM *BroadcastMessageType) Encrypt() string {
	// @TODO encryptar
	return BM.CreateString()
}

func (BM *BroadcastMessageType) Decrypt() {
	// @TODO desencryptar
}
