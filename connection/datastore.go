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
	"log/slog"
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
		slog.Error("No se pudo inicializar el datastore", "error", err)
		return nil, nil
	}

	blockDs := namespace.Wrap(ds, datastore.NewKey("/blocks"))
	bstore := blockstore.NewBlockstore(blockDs)
	bsnetwork := bsnet.NewFromIpfsHost(n.Host)
	exchange := bitswap.New(context.Background(), bsnetwork, routinghelpers.Null{}, bstore)

	bsvc := blockservice.New(bstore, exchange)
	dagService := merkledag.NewDAGService(bsvc)
	topic := "datastore-topic-crdt"
	opts := crdt.DefaultOptions()
	opts.RebroadcastInterval = 10 * time.Second
	n.chPutHook = make(chan global.KVType)
	go n.PutHookHandler()
	opts.PutHook = func(k datastore.Key, v []byte) {
		slog.Debug("[CRDT COMMIT]", "key", k, "value", string(v))
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
				slog.Error("Error en sync periódico", "error", err)
			}
			if err := n.PebbleStore.Sync(context.Background(), datastore.NewKey("/")); err != nil {
				slog.Error("Error en sync pebble", "error", err)
			}
		}
	}()

	return crdtStore, ds
}

func (n *Network) PutCRDT(bucket string, key string, value string) {
	slog.Debug("Guardando en CRDT", "bucket", bucket, "key", key, "value", value)
	k := datastore.NewKey("/" + bucket + "/" + key)
	err := n.DataStore.Put(context.Background(), k, []byte(value))
	if err != nil {
		slog.Error("Error al guardar", "error", err)
	}
	if err := n.PebbleStore.Sync(context.Background(), datastore.NewKey("/")); err != nil {
		slog.Error("Error en sync", "error", err)
	}
}

func (n *Network) Query(bucket string) []global.KVType {
	resp := []global.KVType{}
	results, err := n.PebbleStore.Query(context.Background(), query.Query{
		Prefix: "/crdt-data/s/k/" + bucket,
	})
	if err != nil {
		slog.Error("error en query", "error", err)
	}
	defer results.Close()

	for r := range results.Next() {
		if r.Error != nil {
			slog.Error("error en query", "error", r.Error)
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
		slog.Error("Error al consultar", "error", err)
		return
	}
	defer results.Close()

	slog.Debug("Listado de registros en CRDT:")
	for result := range results.Next() {
		if result.Error != nil {
			slog.Error("Error en el registro", "error", result.Error)
			continue
		}
		slog.Debug("CRDT", "clave", result.Key, "valor", string(result.Value))
	}
}

func (n *Network) PutHookHandler() {
	rxPeers, _ := regexp.Compile("/peers/.*")
	rxMembers, _ := regexp.Compile("/members/.*")

	for kv := range n.chPutHook {
		if rxMembers.Match([]byte(kv.Key)) {
			pubBytes, err := hex.DecodeString(kv.Name)
			if err != nil {
				slog.Error("Error decodificando hex", "error", err)
				return
			}

			pubKey, err := crypto.UnmarshalPublicKey(pubBytes)
			if err != nil {
				slog.Error("Error unmarshal libp2p", "error", err)
				return
			}
			n.MutexSesiones.Lock()
			n.MasterEntities[kv.Key] = pubKey
			n.MutexSesiones.Unlock()
			continue
		}

		if rxPeers.Match([]byte(kv.Key)) {
			cleanKey := strings.TrimPrefix(kv.Key, "/peers/")
			id, err := peer.Decode(cleanKey)
			if err != nil {
				slog.Error("Error decodificando peer", "key", cleanKey, "error", err)
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

func (n *Network) GetValidator() string {
	k := datastore.NewKey("/innercfg/validator")
	validator, err := n.PebbleStore.Get(context.Background(), k)
	if err != nil {
		slog.Error("Error al obtener el validator", "error", err)
		return ""
	}

	return string(validator)
}

func GetValidator(network string) string {
	ds, err := dspebble.NewDatastore("store/"+network+".db", nil)
	if err != nil {
		slog.Error("No se pudo inicializar el datastore", "error", err)
		return ""
	}
	defer ds.Close()
	k := datastore.NewKey("/innercfg/validator")
	validator, err := ds.Get(context.Background(), k)
	if err != nil {
		slog.Error("Error al obtener el validator", "error", err)
		return ""
	}

	return string(validator)
}

func (n *Network) SetValidator(validator string) {
	k := datastore.NewKey("/innercfg/validator")
	err := n.PebbleStore.Put(context.Background(), k, []byte(validator))
	if err != nil {
		slog.Error("Error al guardar", "error", err)
	}
}

func SetValidator(network string, validator string) {
	ds, err := dspebble.NewDatastore("store/"+network+".db", nil)
	if err != nil {
		slog.Error("No se pudo inicializar el datastore", "error", err)
		return
	}

	k := datastore.NewKey("/innercfg/validator")
	err = ds.Put(context.Background(), k, []byte(validator))
	if err != nil {
		slog.Error("Error al guardar", "error", err)
	}
	ds.Close()
}
