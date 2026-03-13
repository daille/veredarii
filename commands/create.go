package cmd

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
	"Veredarii/configuration"
	"Veredarii/global"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// 4. Subcomando: create
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea una nueva red",
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" && port == "" {
			fmt.Println("❌ Error: Se requiere la flag --network y --port")
			return
		}

		fmt.Println("Creando red " + network + "...")

		// crea una llave de la red
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			fmt.Println("Error al generar la llave:", err)
			return
		}

		fmt.Println("Llave de la red:", hex.EncodeToString(key))

		fmt.Println("Cargando configuracion...")
		configuration.CM = configuration.NewConfigurationManager()
		err := configuration.CM.LoadConfig()
		if err != nil {
			fmt.Println("Error cargando configuracion:", err)
			return
		}
		config := configuration.CM.GetConfig()

		// Crea el peerID
		if entity != "" {
			pathIdentity := "./" + entity + ".key"
			global.ObtenerIdentidad(pathIdentity)

			// Modifica el config.json
			config.Identity.Entity = entity
			config.Identity.PrivKeyFile = pathIdentity
		}
		if config.LocalInterface.Server.Port == "" {
			config.LocalInterface.Server.Port = "8000"
		}

		var Network global.NetworkType
		Network.Name = network
		Network.Port = port
		Network.MyAddress = []string{"/ip4/0.0.0.0/tcp/" + port, "/ip4/0.0.0.0/udp/" + port + "/quic"}
		Network.NetworkKey = hex.EncodeToString(key)
		Network.Pivots = []string{}
		Network.Entities = []global.KVType{}
		Network.Topics = []global.TopicType{}

		// 4. Guardar los cambios
		if Network.ResourcesPath != "" {
			Network.ResourcesPath = "./resources_" + network + ".json"
			err = os.WriteFile("./resources_"+network+".json", []byte("{\n    \"API\": [\n        {}\n    ],\n    \"FILE\": [\n        {}\n    ],\n    \"DATA_SOURCE\": [\n        {}\n    ]\n}"), 0644)
			if err != nil {
				log.Fatalf("Error escribiendo archivo: %v", err)
			}
		}

		// 4. Guardar los cambios
		if Network.RemoteResourcesPath != "" {
			Network.RemoteResourcesPath = "./remote_resources_" + network + ".json"
			err = os.WriteFile("./remote_resources_"+network+".json", []byte("{\n    \"API\": [\n        {}\n    ],\n    \"FILE\": [\n        {}\n    ],\n    \"DATA_SOURCE\": [\n        {}\n    ]\n}"), 0644)
			if err != nil {
				log.Fatalf("Error escribiendo archivo: %v", err)
			}
		}

		config.Networks = append(config.Networks, Network)

		// 3. Convertir de vuelta a JSON con indentación para que sea legible
		updatedJSON, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			log.Fatalf("Error creando JSON: %v", err)
		}

		// 4. Guardar los cambios
		err = os.WriteFile(configuration.ConfigFilename, updatedJSON, 0644)
		if err != nil {
			log.Fatalf("Error escribiendo archivo: %v", err)
		}

	},
}

func init() {
	createCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&port, "port", "p", "", "Puerto de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	rootCmd.AddCommand(createCmd)
}
