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
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

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
		config := configuration.CM.GetConfig()

		passphrase := "mi frase super secreta para la red"
		salt := "mi-red-p2p-secreta-unique-salt"

		aesKey := global.GenerarLlaveDesdeFrase(passphrase, salt)

		inv := global.InvitacionType{
			Inviter:    configuration.CM.GetConfig().Identity.Entity,
			PeerID:     "QmYy6libp2pID",
			Guest:      entity,
			Network:    network,
			Expiration: time.Now().Add(24 * time.Hour),
		}

		token := global.CipherInvitation(inv, aesKey)
		fmt.Printf("Token: %s\n\n", token)

		err = os.MkdirAll("./invitations", 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:", err)
			return
		}
		pivot := base64.StdEncoding.EncodeToString([]byte(config.Networks[0].Pivots[0]))

		// Firma digital de (token|pivot) con la clave privada del invitador
		privKey, err := global.ObtenerIdentidad(config.Identity.PrivKeyFile)
		if err != nil {
			fmt.Println("❌ Error obteniendo la identidad del invitador:", err)
			return
		}
		datosFirmados := []byte(token + separator + pivot)
		sigBytes, err := privKey.Sign(datosFirmados)
		if err != nil {
			fmt.Println("❌ Error firmando la invitación:", err)
			return
		}
		firma := base64.StdEncoding.EncodeToString(sigBytes)

		// Formato final: token|pivot|firma
		vni := []byte(token + separator + pivot + separator + firma)
		err = os.WriteFile("./invitations/"+network+"."+entity+".vni", vni, 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo:", err)
			return
		}

		fmt.Printf("🚀 No more time to waste '%s' creada con éxito '%s'.\n", network, entity)
	},
}

func init() {
	inviteCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	inviteCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")

	rootCmd.AddCommand(versionCmd)
}
