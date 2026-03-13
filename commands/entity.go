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
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

// Variables para capturar los valores de las flags
var name string
var newName string

var entityCmd = &cobra.Command{
	Use:   "entity",
	Short: "Operaciones de entidad",
}

var entityLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista las entidades",
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" {
			fmt.Println("❌ Error: Se requieren las flags --network")
			return
		}

		configuration.CM = configuration.NewConfigurationManager()
		err := configuration.CM.LoadConfig()
		if err != nil {
			fmt.Println("Error cargando configuracion:", "error", err.Error())
			return
		}
		results := [][]string{}

		for _, n := range configuration.CM.Config.Networks {
			if n.Name == network {

				for i, e := range n.Entities {
					results = append(results, []string{
						fmt.Sprintf("%d", i),
						e.Name,
						e.Key,
					})
				}
				break
			}
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(teal)).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case row == table.HeaderRow:
					return headerStyle
				case row%2 == 0:
					return evenRowStyle
				default:
					return oddRowStyle
				}
			}).
			Headers("", "Entidad", "Public key").
			Rows(results...)

		fmt.Println("Entidades de la red ", network)
		if len(results) > 0 {
			fmt.Println(t.String())
		} else {
			fmt.Println("No existen entidades")
		}

	},
}

/*
var newEntityCmd = &cobra.Command{
	Use:   "new",
	Short: "Crea una nueva entidad",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			fmt.Println("❌ Error: Se requiere la flag --name")
			return
		}

		priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			fmt.Printf("❌ Error generando llaves: %v\n", err)
			return
		}
		// 2. Serializar llave pública a formato Protobuf (el que tú quieres)
		pubBytes, _ := crypto.MarshalPublicKey(pub)
		pubHex := hex.EncodeToString(pubBytes)

		// 3. Serializar llave privada a bytes crudos para el archivo
		privBytes, _ := crypto.MarshalPrivateKey(priv)

		// 4. Guardar la privada en un archivo nombrado según la entidad
		fileName := fmt.Sprintf("%s.key", name)
		err = os.WriteFile(fileName, privBytes, 0600) // 0600: solo lectura/escritura para el dueño
		if err != nil {
			fmt.Printf("❌ Error al guardar el archivo: %v\n", err)
			return
		}

		// 5. Mostrar resultados
		fmt.Println("✅ Entidad creada exitosamente")
		fmt.Printf("📂 Llave privada guardada en: %s\n", fileName)
		fmt.Printf("🌐 Llave pública (libp2p): %s\n", pubHex)

		configuration.CM = configuration.NewConfigurationManager()
		cfg := configuration.CM.NewConfig()

		cfg.Identity.Entity = name
		cfg.Identity.PrivKeyFile = fileName
		cfg.LocalInterface.Server.Port = "8100"
		configuration.CM.Save()

		// policy y model
		err = os.WriteFile("policy.csv", []byte("#  Sujeto     | dominio                 | Objeto (Servicio)     | Acción | TPS\n#p,alice,red_interoperabilidad,api-proxy/1.0.0,hola"), 0600)
		if err != nil {
			fmt.Printf("❌ Error al guardar el archivo: %v\n", err)
			return
		}

		err = os.WriteFile("model.conf", []byte(`[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, tps

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.dom == p.dom && r.obj == p.obj && r.act == p.act`), 0600)
		if err != nil {
			fmt.Printf("❌ Error al guardar el archivo: %v\n", err)
			return
		}

		fmt.Printf("🚀 Entidad '%s' creada con éxito.\n", name)
	},
}*/

func init() {
	//newEntityCmd.PersistentFlags().StringVarP(&name, "entity", "e", "", "Nombre de la entidad (requerido)")
	entityLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	entityCmd.AddCommand(entityLsCmd)
	rootCmd.AddCommand(entityCmd)
}
