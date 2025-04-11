package command

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
	"crypto/rand"
	"encoding/base64"
	"strings"

	"NodoCb/global"
	"NodoCb/manager/database"
	"NodoCb/util"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Administra la red",
	Long:  `.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		CreateNetwork()
	},
}

func init() {
	networkCmd.Flags().BoolVar(&parameter_new, "new", false, "create a new network")
	networkCmd.Flags().StringVar(&parameter_name, "name", "", "The name of the network (required)")
	networkCmd.Flags().StringVar(&parameter_description, "description", "", "The description of the network (required)")
	networkCmd.Flags().StringVar(&parameter_country, "country", "", "country of the network")

	networkCmd.Flags().StringVar(&parameter_endpoint, "endpoint", "", "endpoint of the network")
	networkCmd.Flags().StringVar(&parameter_port, "port", "", "port of the network")

	/*createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("description")
	createCmd.MarkFlagRequired("port")
	createCmd.MarkFlagRequired("endpoint")*/
	rootCmd.AddCommand(networkCmd)
}

func CreateNetwork() {

	fmt.Printf("****** Crea una red ******")
	fmt.Println("Name: ", parameter_name)
	fmt.Println("description: ", parameter_description)
	fmt.Println("port: ", parameter_port)
	fmt.Println("networkkey: ", parameter_networkkey)
	fmt.Println("endpoint: ", parameter_endpoint)
	fmt.Println("country: ", parameter_country)

	//Crea una CA
	pub, priv := crearCA(parameter_name, parameter_country)
	util.SaveFile("./ca.pem", pub)
	util.SaveFile("./caPriv.pem", priv)

	var ck []byte
	var err error
	if parameter_networkkey == "" {
		ck = make([]byte, 32)
		_, err = rand.Read(ck)
		fmt.Println("Creando NetworkKey:", base64.StdEncoding.EncodeToString(ck))
	} else {
		ck, err = base64.StdEncoding.DecodeString(parameter_networkkey)
		if err != nil {
			fmt.Println("ERROR: ", err)
			return
		}
	}

	uid, _ := uuid.NewUUID()
	msgK := uid.String()
	a := strings.Split(msgK, "-")
	b := a[3] + a[4]
	log.Info("MSG-K: ", util.Yellow(b))

	database.UpdateNetwork(
		global.ClusterDataType{
			Name:         parameter_name,
			Description:  parameter_description,
			ClusterKey:   ck,
			MessageKey:   b,
			Port:         parameter_port,
			EndpointJoin: parameter_endpoint,
		},
	)

}
