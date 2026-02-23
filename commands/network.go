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
	"Veredarii/connection"
	"Veredarii/global"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

const separator = "|"

// 2. Comando Padre: entity
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Operaciones de red",
}

// 3. Subcomando: create
var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invita una entidad a la red",
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" || entity == "" {
			fmt.Println("❌ Error: Se requieren las flags --network y --entity")
			return
		}

		fmt.Println("Cargando configuracion...")
		configuration.CM = configuration.NewConfigurationManager()
		err := configuration.CM.LoadConfig()
		if err != nil {
			fmt.Println("Error cargando configuracion:", err)
			return
		}

		passphrase := "mi frase super secreta para la red"
		salt := "mi-red-p2p-secreta-unique-salt"

		key := global.GenerarLlaveDesdeFrase(passphrase, salt)

		inv := global.InvitacionType{
			Inviter:    configuration.CM.GetConfig().Identity.Entity,
			PeerID:     "QmYy6libp2pID",
			Guest:      entity,
			Network:    network,
			Expiration: time.Now().Add(24 * time.Hour),
		}

		token := global.CipherInvitation(inv, key)
		fmt.Printf("Token: %s\n\n", token)

		err = os.MkdirAll("./invitations", 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:", err)
			return
		}
		pivot := base64.StdEncoding.EncodeToString([]byte(configuration.CM.GetConfig().Networks[0].Pivots[0]))
		vni := []byte(token + separator + pivot)
		err = os.WriteFile("./invitations/"+network+"."+entity+".vni", vni, 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo:", err)
			return
		}

		fmt.Printf("🚀 No more time to waste '%s' creada con éxito '%s'.\n", network, entity)
	},
}

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

// 4. Subcomando: newkey
var newNetworkKeyCmd = &cobra.Command{
	Use:   "newkey",
	Short: "Genera una nueva llave",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔑 Generando nueva llave para: %s\n", name)
	},
}

// 5. Subcomando: newname
var newPivotCmd = &cobra.Command{
	Use:   "newpivot",
	Short: "Cambia el nombre de la entidad",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" || newName == "" {
			fmt.Println("❌ Error: Se requieren las flags --name y --newname")
			return
		}
		fmt.Printf("📝 Renombrando %s a %s\n", name, newName)
	},
}

// 4. Subcomando: newkey
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Une a la red",
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" || entity == "" || inviter == "" || file == "" || key == "" {
			fmt.Println("❌ Error: Se requieren las flags --network, --entity, --invitation, --inviter y --key")
			return
		}

		inv, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			fmt.Println("❌ Error al abrir el archivo:", err)
			return
		}
		invBytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("❌ Error al leer el archivo:", err)
			return
		}
		cruda := string(invBytes)
		inv.Close()
		both := strings.Split(cruda, separator)
		if len(both) < 2 {
			fmt.Println("❌ Error: El archivo de invitación no es válido")
			return
		}
		invitation := both[0]
		pivotEncoded := strings.ReplaceAll(both[1], "\n", "")
		pivot, err := base64.StdEncoding.DecodeString(pivotEncoded)
		if err != nil {
			fmt.Println("❌ Error al decodificar el pivot:", err)
			return
		}

		// Crea llaves de la entidad (si aun no tiene) o usa las existentes
		privateKey, err := global.ObtenerIdentidad("./" + entity + ".key")
		if err != nil {
			fmt.Println("❌ Error al obtener la identidad:", err)
			return
		}
		publicKey := global.GetPubKey(privateKey)

		// Se conecta a la red
		psk, err := global.DecodeV1PSK(key)
		if err != nil {
			log.Fatal("Error cargando PSK:", err)
		}

		h, err := libp2p.New(
			libp2p.NoListenAddrs,
			libp2p.Identity(privateKey),
			libp2p.PrivateNetwork(psk),
		)
		if err != nil {
			fmt.Println("❌ Error creando host:", err)
			return
		}
		defer h.Close()

		maddr, err := multiaddr.NewMultiaddr(string(pivot))
		if err != nil {
			fmt.Println("❌ Error parseando la dirección del pivot:", err)
			return
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			fmt.Println("❌ Error extraiendo el ID del peer:", err)
			return
		}

		fmt.Println("🌐 pivot ", string(pivot))
		fmt.Println("🌐 Conectando al pivot ", info)
		ctx := context.Background()
		if err := h.Connect(ctx, *info); err != nil {
			fmt.Println("❌ Error conectándose al pivot:", err)
			return
		}

		s, err := h.NewStream(ctx, info.ID, global.ProtocolJoin)
		if err != nil {
			fmt.Println("❌ Error abriendo stream:", err)
			return
		}
		defer s.Close()

		// Envía invitación + llave publica
		joinRequest := connection.JoinRequest{
			EntityName:  entity,
			InviterName: inviter,
			Network:     network,
			PublicKey:   publicKey,
			Invitation:  invitation,
		}
		jsonRequest, err := json.Marshal(joinRequest)
		if err != nil {
			fmt.Println("❌ Error serializando solicitud:", err)
			return
		}
		_, err = s.Write(jsonRequest)
		if err != nil {
			fmt.Println("❌ Error enviando solicitud:", err)
			return
		}
		fmt.Println("Solicitud enviada con éxito al pivot")

		// Espera el resultado y si esta ok crea la red
		resourcesPath := "./resources_" + network + ".json"
		f, err := os.OpenFile(resourcesPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)

		if err != nil {
			if os.IsExist(err) {
				log.Println("El archivo ya existe, no se hizo nada.")
			}
			log.Fatal("Error al intentar crear el archivo:", err)
		} else {
			_, err = f.Write([]byte("{\n    \"API\": [\n        {}\n    ],\n    \"FILE\": [\n        {}\n    ],\n    \"DATA_SOURCE\": [\n        {}\n    ]\n}"))
			if err != nil {
				log.Fatal("Error escribiendo contenido:", err)
			}
		}
		f.Close()

		remoteResourcesPath := "./remote_resources_" + network + ".json"
		f, err = os.OpenFile(remoteResourcesPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)

		if err != nil {
			if os.IsExist(err) {
				log.Println("El archivo ya existe, no se hizo nada.")
			}
			log.Fatal("Error al intentar crear el archivo:", err)
		} else {
			_, err = f.Write([]byte("{\n    \"API\": [\n        {}\n    ],\n    \"FILE\": [\n        {}\n    ],\n    \"DATA_SOURCE\": [\n        {}\n    ]\n}"))
			if err != nil {
				log.Fatal("Error escribiendo contenido:", err)
			}
		}
		f.Close()

		configuration.CM = configuration.NewConfigurationManager()
		configuration.CM.LoadConfig()
		cfg := configuration.CM.GetConfig()
		port := "9100"
		cfg.Networks = append(cfg.Networks, global.NetworkType{
			Name:                network,
			Port:                port,
			Pivots:              []string{string(pivot)},
			NetworkKey:          key,
			MyAddress:           []string{"/ip4/0.0.0.0/tcp/" + port, "/ip4/0.0.0.0/udp/" + port + "/quic"},
			Entities:            []global.KVType{},
			Topics:              []global.TopicType{},
			ResourcesPath:       resourcesPath,
			RemoteResourcesPath: remoteResourcesPath,
		})
		configuration.CM.Save()

	},
}

// 4. Subcomando: remote
// 2. Comando Padre: entity
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Operaciones de recursos remotos",
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Agrega un recurso remoto",
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" || resourceType == "" || resource == "" || local == "" {
			fmt.Println("❌ Error: Se requieren las flags --network, --type, --resource, --local")
			return
		}

		configuration.CM = configuration.NewConfigurationManager()
		configuration.CM.LoadConfig()
		cfg := configuration.CM.GetConfig()

		for idx, net := range cfg.Networks {
			if net.Name == network {
				res := global.ResourceType{
					Name:         resource,
					ResourcePath: local,
				}

				switch resourceType {
				case "API":
					fmt.Println("Agregando recurso remoto API:", resource)
					cfg.Networks[idx].RemoteResources.API = append(cfg.Networks[idx].RemoteResources.API, res)
				case "FILE":
					fmt.Println("Agregando recurso remoto FILE:", resource)
					cfg.Networks[idx].RemoteResources.FILE = append(cfg.Networks[idx].RemoteResources.FILE, res)
				case "DATA_SOURCE":
					fmt.Println("Agregando recurso remoto DATA_SOURCE:", resource)
					cfg.Networks[idx].RemoteResources.DATASOURCE = append(cfg.Networks[idx].RemoteResources.DATASOURCE, res)
				default:
					fmt.Println("❌ Error: Tipo de recurso no válido")
					return
				}

				configuration.CM.SaveRemoteResources(network)
				break
			}
		}
	},
}

func init() {
	inviteCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	inviteCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")

	joinCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	joinCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	joinCmd.PersistentFlags().StringVarP(&file, "invitation", "i", "", "Archivo de invitación (requerido)")
	joinCmd.PersistentFlags().StringVarP(&inviter, "inviter", "m", "", "Nombre del invitador (requerido)")
	joinCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "Llave de la red (requerido)")

	createCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&port, "port", "p", "", "Puerto de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")

	newNetworkKeyCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	newPivotCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	addCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	addCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	addCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	addCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "Ruta local del recurso (requerido)")
	remoteCmd.AddCommand(addCmd)
	networkCmd.AddCommand(inviteCmd, createCmd, newNetworkKeyCmd, newPivotCmd, joinCmd, remoteCmd)
	rootCmd.AddCommand(networkCmd)
}
