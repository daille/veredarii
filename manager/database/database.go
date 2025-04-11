package database

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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var db *leveldb.DB

const BREAKE string = "©Æ§"
const JUMPLINE string = "‹‡›"
const NETWORK string = "NETWORK"
const MEMBERS string = "MEMBERS"

func DBInit(path string) {
	var err error
	db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func UpdateRaw(key string, value string) {
	err := db.Put([]byte(key), []byte(value), nil)
	if err != nil {
		log.Error(err)
	}
}

func UpdateNetwork(nt global.ClusterDataType) {
	j, err := json.Marshal(nt)
	if err != nil {
		log.Error(err)
	}

	UpdateRaw(NETWORK+":"+nt.Name, string(j))
}

func GetAllNetworks() []global.ClusterDataType {
	var nts []global.ClusterDataType

	iter := db.NewIterator(util.BytesPrefix([]byte(NETWORK+":")), nil)
	for iter.Next() {
		var nt global.ClusterDataType
		err := json.Unmarshal(iter.Value(), &nt)
		if err != nil {
			log.Error(err)
		}
		nts = append(nts, nt)
	}
	iter.Release()

	return nts
}

func GetNetwork(name string) global.ClusterDataType {
	var nt global.ClusterDataType
	data, err := db.Get([]byte(NETWORK+":"+name), nil)
	err = json.Unmarshal([]byte(data), &nt)
	if err != nil {
		log.Error(err)
	}

	return nt
}

func GetAllMembers() []global.MemberType {
	var mems []global.MemberType

	iter := db.NewIterator(util.BytesPrefix([]byte(MEMBERS+":")), nil)
	for iter.Next() {
		var mem global.MemberType
		err := json.Unmarshal([]byte(iter.Value()), &mem)
		if err != nil {
			log.Error(err)
		} else {
			mems = append(mems, mem)
		}
	}
	iter.Release()

	return mems
}

func UpdateMember(mem global.MemberType) {
	j, err := json.Marshal(mem)
	if err != nil {
		log.Error(err)
	}
	UpdateRaw(MEMBERS+":"+mem.Organization, string(j))
}

func GetDatabaseToShare() string {
	strFinal := ""
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		strFinal += string(iter.Key()) + BREAKE + string(iter.Value()) + JUMPLINE
	}
	iter.Release()
	if iter.Error() != nil {
		log.Error((iter.Error()))
	}

	return strFinal
}

func InsertDatabaseShared(sharedDb string) {
	a := strings.Split(sharedDb, "\n")

	for _, j := range a {
		b := strings.Split(j, BREAKE)
		err := db.Put([]byte(b[0]), []byte(b[1]), nil)
		if err != nil {
			log.Error(err)
		}
	}
}
