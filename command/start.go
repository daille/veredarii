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
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"NodoCb/global"
	"NodoCb/manager/cluster"
	"NodoCb/manager/configuration"
	"NodoCb/manager/database"
	"NodoCb/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start",
	Long:  `Start the node of comunication`,
	Run: func(cmd *cobra.Command, args []string) {
		Iniciar()
	},
}

func init() {
	startCmd.Flags().BoolVar(&parameter_new, "new", false, "Start a new network")
	startCmd.Flags().StringVar(&parameter_endpoint, "endpoint", "", "")
	startCmd.Flags().StringVar(&parameter_networkkey, "key", "", "")
	startCmd.Flags().StringVar(&parameter_port, "port", "", "The port where others will joins to this node (required)")
	rootCmd.AddCommand(startCmd)
}

func Iniciar() {
	fmt.Printf(util.Green("\n\n",
		"┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐\n",
		"│ VEREDARII                                                                                              │\n",
		"│                                                                                                        │\n",
		"│ Version: ", global.Version, "                                                                                 │\n",
		"└────────────────────────────────────────────────────────────────────────────────────────────────────────┘\n\n",
	))

	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
		DisableColors:   true,
	})
	log.SetLevel(log.DebugLevel)

	nmd := configuration.GetNodoMetadata()

	// Conectarse a las redes a las que se encuentre suscrito
	networks := database.GetAllNetworks()
	if len(networks) > 0 {
		for _, networkData := range networks {
			if parameter_port != "" {
				networkData.Port = parameter_port
			}
			network := cluster.NewCluster(networkData, nmd)
			log.Info(util.Teal("Conecting to the net: ", network.Data.Name))
			/*go func() {
				time.Sleep(5 * time.Second)
				network.SendBroadcastMesage("mine mine mine (" + nmd.ID + ")")
			}()*/
			network.Connect(parameter_new, false)
		}
	} else {
		if parameter_endpoint != "" {
			k, err := base64.StdEncoding.DecodeString(parameter_networkkey)
			if err != nil {
				log.Error(err)
			} else {
				network := cluster.NewCluster(
					global.ClusterDataType{
						EndpointJoin: parameter_endpoint,
						ClusterKey:   k,
						Port:         parameter_port,
					},
					nmd,
				)
				log.Info(util.Teal("First time connected to the net: ", parameter_endpoint))
				log.Debug("port:", parameter_port)
				pubKey, priKey, _ := ed25519.GenerateKey(nil)
				global.ConfigFile.Identity.Keys.Public = hex.EncodeToString((pubKey))
				global.ConfigFile.Identity.Keys.Private = hex.EncodeToString(priKey)
				configuration.SaveConfig()
				network.Connect(false, true)
			}
		} else {
			log.Info("No avilable nets.")
		}
	}
}
