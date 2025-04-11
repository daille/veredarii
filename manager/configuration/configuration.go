package configuration

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
	"NodoCb/manager/database"
	"NodoCb/util"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

const CONFIG string = "./config.json"

func LoadConfiguration() {
	jsonFile, err := os.Open(CONFIG)
	if err != nil {
		log.Debug(util.Fatal("## CONF ERROR: ", err))
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &global.ConfigFile)
	if err != nil {
		log.Error(util.Fatal(err))
	}

	database.DBInit(global.ConfigFile.Database.Path)

	// Buscar en las variables de entorno

	// Buscar en el comando

	// Buscar en la DB local

	/*var tmp global.EnvironmentType
	if err := env.Set(&tmp); err != nil {
		log.Error(util.Red(err))
	} else {
		log.Debug("Lectura de variables de entorno")

		conf.ClusterKey, err = base64.StdEncoding.DecodeString(tmp.NODOCB_CLUSTER_KEY)
		if err != nil {
			log.Error(util.Red(err))
		}
		conf.EndpointJoin = tmp.NODOCB_ENDPOINT_JOIN
		conf.Port = tmp.NODOCB_PORT
	}*/
}

func SaveConfig() {
	f, err := os.Create(CONFIG)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	b, err := json.Marshal(global.ConfigFile)
	_, err = f.Write(b)
	if err != nil {
		log.Error(err)
	}
}

func GetNodoMetadata() global.NodoMetaData {

	crt := LoadX509Certificate(global.ConfigFile.Identity.PKI.Public)

	nmd := global.NodoMetaData{}
	if len(crt.Subject.Organization) > 0 {
		nmd.Organism = crt.Subject.Organization[0]
	}
	if len(crt.Subject.Country) > 0 {
		nmd.Country = crt.Subject.Country[0]
	}

	return nmd
}

func NewClusterKey() []byte {
	clusterKey := make([]byte, 32)
	_, err := rand.Read(clusterKey)
	fmt.Println("ClusterKey:", base64.StdEncoding.EncodeToString(clusterKey), err)
	return clusterKey
}

func VerifyIssuedBy(rawcert string) bool {
	caCertPEM, err := os.ReadFile("./ca.pem")
	if err != nil {
		log.Error("read CA PEM file")
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(caCertPEM)

	cert := LoadX509CertificateFromString(rawcert)
	chain, err := cert.Verify(x509.VerifyOptions{Roots: roots})
	if err != nil {
		log.Error("failed to verify cert")
		return false
	} else {
		log.Debug("issuing chain: ", chain)
		return true
	}
}
