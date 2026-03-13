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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

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
		partes := strings.Split(cruda, separator)
		if len(partes) < 2 {
			fmt.Println("❌ Error: El archivo de invitación no es válido")
			return
		}
		invitation := partes[0]
		pivotEncoded := strings.ReplaceAll(partes[1], "\n", "")
		pivot, err := base64.StdEncoding.DecodeString(pivotEncoded)
		if err != nil {
			fmt.Println("❌ Error al decodificar el pivot:", err)
			return
		}
		// Firma digital incluida en el archivo (campo 3), disponible para verificación futura
		// var firmaInvitacion string
		// if len(partes) >= 3 {
		// 	firmaInvitacion = strings.ReplaceAll(partes[2], "\n", "")
		// }

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

		s, err := h.NewStream(ctx, info.ID, protocol.ID(global.ProtocolJoin))
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

func init() {
	joinCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	joinCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	joinCmd.PersistentFlags().StringVarP(&file, "invitation", "i", "", "Archivo de invitación (requerido)")
	joinCmd.PersistentFlags().StringVarP(&inviter, "inviter", "m", "", "Nombre del invitador (requerido)")
	joinCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "Llave de la red (requerido)")

	rootCmd.AddCommand(joinCmd)
}
