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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

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
		err = os.WriteFile("./invitations/"+network+"."+entity+".vni", []byte(token), 0644)
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
		if network == "" || entity == "" || inviter == "" || file == "" {
			fmt.Println("❌ Error: Se requieren las flags --network, --entity y --invitation")
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
		invitation := string(invBytes)
		inv.Close()

		// Crea llaves de la entidad (si aun no tiene) o usa las existentes
		privateKey, err := global.ObtenerIdentidad("./" + entity + ".key")
		if err != nil {
			fmt.Println("❌ Error al obtener la identidad:", err)
			return
		}
		publicKey := global.GetPubKey(privateKey)

		// Se conecta a la red
		// @TODO obtener la llave de la red
		psk, err := global.DecodeV1PSK("7d44e2103328e75003666d3f23f858e376a9f0290130f1464303d8d515a676c8")
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

		// @TODO obtener la dirección del pivot
		maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWEgeYgNpgnbVFP3hyCog1kLF7RbmD1XmiopWjkaCtnV2b")
		if err != nil {
			fmt.Println("❌ Error parseando la dirección del pivot:", err)
			return
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			fmt.Println("❌ Error extraiendo el ID del peer:", err)
			return
		}

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

		// Espera el resultado
	},
}

func init() {
	inviteCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	inviteCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")

	joinCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	joinCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	joinCmd.PersistentFlags().StringVarP(&file, "invitation", "i", "", "Archivo de invitación (requerido)")
	joinCmd.PersistentFlags().StringVarP(&inviter, "inviter", "m", "", "Nombre del invitador (requerido)")

	createCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&port, "port", "p", "", "Puerto de la red (requerido)")
	createCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")

	newNetworkKeyCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	newPivotCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	networkCmd.AddCommand(inviteCmd, createCmd, newNetworkKeyCmd, newPivotCmd, joinCmd)
	rootCmd.AddCommand(networkCmd)
}
