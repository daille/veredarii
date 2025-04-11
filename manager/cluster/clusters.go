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
	"NodoCb/manager/configuration"
	"NodoCb/manager/database"
	"NodoCb/util"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/memberlist"
)

var Clusters []ClusterType

type ClusterType struct {
	Data global.ClusterDataType `json:"data"`
	D    *global.MyDelegate
	Meta global.NodoMetaData
	List *memberlist.Memberlist
}

func NewCluster(data global.ClusterDataType, nmd global.NodoMetaData) *ClusterType {
	t := &ClusterType{
		Data: data,
		Meta: nmd,
	}
	t.D = new(global.MyDelegate)
	return t
}

func (T *ClusterType) Connect(newcluster bool, firstConnection bool) {
	var err error
	msgCh := make(chan []byte)
	e := new(global.MyEventDelegate)
	e.Num = 0

	T.D.Meta = T.Meta
	T.D.Meta.ID = uuid.New().String()
	log.Info("NODE NAME: ", T.D.Meta.ID)
	T.D.MsgCh = msgCh
	T.D.Broadcasts = new(memberlist.TransmitLimitedQueue)
	T.D.Broadcasts.NumNodes = func() int {
		return e.Num
	}
	T.D.Broadcasts.RetransmitMult = 1

	conf := memberlist.DefaultLocalConfig()
	conf.Name = T.D.Meta.ID
	conf.BindPort, _ = strconv.Atoi(T.Data.Port)
	conf.SecretKey = T.Data.ClusterKey
	conf.AdvertisePort = conf.BindPort
	conf.Events = e
	conf.Delegate = T.D

	T.List, err = memberlist.Create(conf)
	if err != nil {
		log.Fatal(err)
	}

	local := T.List.LocalNode()
	fmt.Printf("%s:%d  \n", local.Addr.To4().String(), local.Port)
	T.List.Join([]string{
		fmt.Sprintf("%s:%d", local.Addr.To4().String(), local.Port),
	})
	e.Num = T.List.NumMembers()

	if !newcluster {
		log.Debug(util.Red("No es nuevo."), " conectando a ", T.Data.EndpointJoin)
		if _, err := T.List.Join([]string{T.Data.EndpointJoin}); err != nil {
			log.Error(err)
		}

		if firstConnection {
			if !T.DoFirstConnection() {
				log.Error("Can't comunicate with other members")
			}
		}
	}

	stopCtx, cancel := context.WithCancel(context.TODO())
	go global.WaitSignal(cancel)

	run := true
	for run {
		select {
		case data := <-T.D.MsgCh:
			log.Info(util.Green("Llego mensaje: ", string(data)))
			msg, ok := global.ParseMyBroadcastMessage(data)
			if ok != true {
				log.Error("No pudo parsear", ok)
				continue
			}
			T.ProcessMessage(msg)

		case <-stopCtx.Done():
			log.Debug("stop")
			run = false
		}
	}
}

func (T *ClusterType) ProcessMessage(msg *global.MyBroadcastMessage) {
	if msg.Key == global.REQUEST_MEMBERSHIP {
		var body BodyBroadcastMessageType
		if err := json.Unmarshal([]byte(msg.Value), &body); err == nil {
			log.Info("Adding a new organism ", body.Meta.Origin, " to the network")
			if configuration.VerifyIssuedBy(body.Update.Member.PublicCert) {
				log.Debug("The cert of the organism is part of the network")
				pubCert := configuration.LoadX509CertificateFromString(body.Update.Member.PublicCert)
				var spkiKey *rsa.PublicKey
				spkiKey = pubCert.PublicKey.(*rsa.PublicKey)
				msgHash := sha256.New()
				_, err = msgHash.Write([]byte(body.Update.Member.PublicKey))
				if err != nil {
					log.Error(err)
				}
				msgHashSum := msgHash.Sum(nil)
				sign, err := hex.DecodeString(body.Update.Member.Signature)
				if err != nil {
					fmt.Println("could not verify signature: ", err)
					return
				}
				err = rsa.VerifyPSS(spkiKey, crypto.SHA256, msgHashSum, sign, nil)
				if err != nil {
					fmt.Println("could not verify signature: ", err)
					return
				}
				log.Debug("Valid signature")

				msToBrc := NewBroadcastMessage(BodyBroadcastMessageType{}, T.Meta)
				msToBrc.Body.Update.Action = global.UPDATE_MEMBER
				msToBrc.Body.Update.Member = body.Update.Member
				sbmMsg, _ := json.Marshal(msToBrc)
				T.SendBroadcastMesage(string(sbmMsg))

				T.SendApproval(msg.Origin, global.ACTION_WELCOME, T.Data)
				time.Sleep(2)
				ec, err := util.EncryptMessage([]byte(T.Data.MessageKey), database.GetDatabaseToShare())
				if err != nil {
					log.Error(err)
					return
				}
				db := hex.EncodeToString([]byte(ec))
				T.SendDatabase(msg.Origin, global.ACTION_DATABASE, db)
			} else {
				log.Debug("Is NOT a member of the network")
			}
		}
	} else if msg.Key == global.APPROVAL_MEMBERSHIP {
		var body BodyBroadcastMessageType
		if err := json.Unmarshal([]byte(msg.Value), &body); err == nil {
			log.Info(util.Teal("Welcome to the network ", body.Update.Network.Name))
			body.Update.Network.Port = T.Data.Port
			body.Update.Network.EndpointJoin = T.Data.EndpointJoin
			database.UpdateNetwork(body.Update.Network)
			T.Data = body.Update.Network
		}
	} else if msg.Key == global.UPDATE_NETWORK {
		de, _ := hex.DecodeString(msg.Value)
		bodyDecrypted, _ := util.DecryptMessage([]byte(T.Data.MessageKey), string(de))
		a := strings.Split(bodyDecrypted, database.JUMPLINE)
		for _, j := range a {
			b := strings.Split(j, database.BREAKE)
			log.Debug(b)
			if len(b) > 1 {
				database.UpdateRaw(b[0], b[1])
			}
		}
		log.Debug("Database updated. ", len(a), " rows added")
	}
}

func (T *ClusterType) SendBroadcastMesage(msg string) {
	m := global.MyBroadcastMessage{
		Key:   "Node Sender:" + T.D.Meta.ID,
		Value: msg,
	}

	log.Printf(util.Yellow("send broadcast msg: key=", m.Key, " value=", m.Value))
	T.D.Broadcasts.QueueBroadcast(m)
}

func (T *ClusterType) DoFirstConnection() bool {
	m := new(global.MyBroadcastMessage)
	m.Key = global.REQUEST_MEMBERSHIP

	msg := NewBroadcastMessage(BodyBroadcastMessageType{}, T.Meta)
	msg.Body.Update.Action = "MEMBER"
	msg.Body.Update.Member.Organization = T.Meta.Organism

	dat, _ := os.ReadFile(global.ConfigFile.Identity.PKI.Public)
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(global.ConfigFile.Identity.Keys.Public))
	if err != nil {
		log.Error(err)
		return false
	}
	msgHashSum := msgHash.Sum(nil)
	msg.Body.Update.Member.PublicCert = string(dat)
	pk := configuration.LoadX509PrivateKey(global.ConfigFile.Identity.PKI.Private)
	signature, err := rsa.SignPSS(rand.Reader, pk, crypto.SHA256, msgHashSum, nil)
	if err != nil {
		log.Error(err)
		return false
	}
	msg.Body.Update.Member.Signature = hex.EncodeToString(signature)
	msg.Body.Update.Member.PublicKey = global.ConfigFile.Identity.Keys.Public

	pubCert := configuration.LoadX509Certificate(global.ConfigFile.Identity.PKI.Public)
	var spkiKey *rsa.PublicKey
	spkiKey = pubCert.PublicKey.(*rsa.PublicKey)
	err = rsa.VerifyPSS(spkiKey, crypto.SHA256, msgHashSum, signature, nil)
	if err != nil {
		fmt.Println("could not verify signature: ", err)
		return false
	}

	for _, node := range T.List.Members() {
		if node.Name == T.Data.Name {
			continue
		}
		msg.Body.Meta.Destiny = node.Name
		m.Value = msg.CreateString()
		m.Origin = T.D.Meta.ID
		log.Debug(fmt.Sprintf("DIRECT to %s msg: key=%s value=%s", node.Name, m.Key, m.Value))
		if err := T.List.SendReliable(node, m.Message()); err == nil {
			break
		}
	}
	return true
}

func (T *ClusterType) SendApproval(nodeOrigin string, action string, network global.ClusterDataType) {
	m := new(global.MyBroadcastMessage)
	m.Key = global.APPROVAL_MEMBERSHIP

	msg := NewBroadcastMessage(BodyBroadcastMessageType{}, T.Meta)
	msg.Body.Update.Action = action
	msg.Body.Update.Network = network

	for _, node := range T.List.Members() {
		if node.Name == nodeOrigin {
			msg.Body.Meta.Destiny = node.Name
			m.Value = msg.CreateString()
			m.Origin = T.D.Meta.ID
			log.Debug(fmt.Sprintf("DIRECT to %s msg: key=%s value=%s", node.Name, m.Key, m.Value))
			T.List.SendReliable(node, m.Message())
			break
		}
	}
}

func (T *ClusterType) SendDatabase(nodeOrigin string, action string, data string) {
	m := new(global.MyBroadcastMessage)
	m.Key = global.UPDATE_NETWORK
	m.Value = data

	for _, node := range T.List.Members() {
		if node.Name == nodeOrigin {
			m.Origin = T.D.Meta.ID
			log.Debug(fmt.Sprintf("DIRECT to %s msg: key=%s value=%s", node.Name, m.Key, m.Value))
			T.List.SendReliable(node, m.Message())
			break
		}
	}
}
