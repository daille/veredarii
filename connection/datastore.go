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
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ipfs/boxo/bitswap"
	bsnet "github.com/ipfs/boxo/bitswap/network/bsnet"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	crdt "github.com/ipfs/go-ds-crdt"
	dspebble "github.com/ipfs/go-ds-pebble"
	routinghelpers "github.com/libp2p/go-libp2p-routing-helpers"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

const PEERS string = "peers"
const MEMBERS string = "members"

func (n *Network) NewDataStore() (*crdt.Datastore, *dspebble.Datastore) {

	ds, err := dspebble.NewDatastore("store/"+n.Name+".db", nil)
	if err != nil {
		log.Fatal("No se pudo inicializar el datastore:", err)
	}

	blockDs := namespace.Wrap(ds, datastore.NewKey("/blocks"))
	bstore := blockstore.NewBlockstore(blockDs)

	// Bitswap para intercambiar bloques con otros peers
	bsnetwork := bsnet.NewFromIpfsHost(n.Host)
	exchange := bitswap.New(context.Background(), bsnetwork, routinghelpers.Null{}, bstore)

	bsvc := blockservice.New(bstore, exchange)
	dagService := merkledag.NewDAGService(bsvc)
	topic := "mi-aplicacion-chat"
	opts := crdt.DefaultOptions()
	opts.RebroadcastInterval = 10 * time.Second
	n.chPutHook = make(chan global.KVType)
	go n.PutHookHandler()
	opts.PutHook = func(k datastore.Key, v []byte) {
		log.Infof("[CRDT COMMIT] %s = %s", k, string(v))
		n.chPutHook <- global.KVType{Key: k.String(), Name: string(v)}
	}

	ctx := context.Background()
	crdtNamespace := datastore.NewKey("/crdt-data")
	broadcaster, err := crdt.NewPubSubBroadcaster(ctx, n.PS, topic)
	if err != nil {
		log.Fatal(err)
	}

	crdtStore, err := crdt.New(
		ds,            // 1. datastore.Datastore (Pebble)
		crdtNamespace, // 2. datastore.Key (El prefijo)
		dagService,    // 3. format.DAGService
		broadcaster,   // 4. crdt.Broadcaster (El que creamos arriba)
		opts,          // 5. *crdt.Options
	)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if err := n.DataStore.Sync(context.Background(), datastore.NewKey("/")); err != nil {
				log.Errorf("Error en sync periódico: %v", err)
			}
			if err := n.PebbleStore.Sync(context.Background(), datastore.NewKey("/")); err != nil {
				log.Errorf("Error en sync pebble: %v", err)
			}
		}
	}()

	return crdtStore, ds
}

func (n *Network) PutCRDT(bucket string, key string, value string) {
	fmt.Println("Guardando en CRDT...")
	k := datastore.NewKey("/" + bucket + "/" + key)
	err := n.DataStore.Put(context.Background(), k, []byte(value))
	if err != nil {
		log.Printf("Error al guardar: %v", err)
	}
	if err := n.PebbleStore.Sync(context.Background(), datastore.NewKey("/")); err != nil {
		log.Printf("Error en sync: %v", err)
	}
}

func (n *Network) Query(bucket string) []global.KVType {
	resp := []global.KVType{}
	results, err := n.PebbleStore.Query(context.Background(), query.Query{
		Prefix: "/crdt-data/s/k/" + bucket,
	})
	if err != nil {
		log.Error("error en query: %w", err)
	}
	defer results.Close()

	for r := range results.Next() {
		if r.Error != nil {
			log.Error("error en query: %w", r.Error)
			continue
		}
		if !strings.HasSuffix(r.Key, "/v") {
			continue
		}
		cleanKey := strings.TrimPrefix(r.Key, "/crdt-data/s/k/"+bucket+"/")
		cleanKey = strings.TrimSuffix(cleanKey, "/v")
		resp = append(resp, global.KVType{Key: cleanKey, Name: string(r.Entry.Value)})
	}
	return resp
}

func (n *Network) QueryCRDT() {
	q := query.Query{}

	ctx := context.Background()
	results, err := n.DataStore.Query(ctx, q)
	if err != nil {
		fmt.Printf("Error al consultar: %v\n", err)
		return
	}

	defer results.Close()

	fmt.Println("Listado de registros en CRDT:")
	for result := range results.Next() {
		if result.Error != nil {
			fmt.Printf("Error en el registro: %v\n", result.Error)
			continue
		}

		fmt.Printf("Clave: %s | Valor: %s\n", result.Key, string(result.Value))
	}

}

func (n *Network) PutHookHandler() {

	rxPeers, _ := regexp.Compile("/peers/.*")
	rxMembers, _ := regexp.Compile("/members/.*")

	for kv := range n.chPutHook {
		// Members
		if rxMembers.Match([]byte(kv.Key)) {

			pubBytes, err := hex.DecodeString(kv.Name)
			if err != nil {
				fmt.Printf("❌ Error decodificando hex: %v\n", err)
				return
			}

			pubKey, err := crypto.UnmarshalPublicKey(pubBytes)
			if err != nil {
				fmt.Printf("❌ Error unmarshal libp2p: %v\n", err)
				return
			}
			n.MutexSesiones.Lock()
			n.MasterEntities[kv.Key] = pubKey
			n.MutexSesiones.Unlock()
			continue
		}

		// Peers
		if rxPeers.Match([]byte(kv.Key)) {
			cleanKey := strings.TrimPrefix(kv.Key, "/peers/")
			log.Debug("✅ Peer {", cleanKey, "}")
			id, err := peer.Decode(cleanKey)
			if err != nil {
				log.Error("❌ Error decodificando peer: ", cleanKey, " : ", err)
				continue
			}

			n.MutexSesiones.Lock()
			n.Peers[id] = PeerType{
				Entity: kv.Name,
			}
			n.MutexSesiones.Unlock()
			continue
		}
	}
}
