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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/cobra"
)

// Variables para capturar los valores de las flags
var name string
var newName string

// 2. Comando Padre: entity
var entityCmd = &cobra.Command{
	Use:   "entity",
	Short: "Operaciones de entidad",
	// No definimos Run aquí para obligar a usar un subcomando
}

// 3. Subcomando: create
var createEntityCmd = &cobra.Command{
	Use:   "create",
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

		fmt.Printf("🚀 Entidad '%s' creada con éxito.\n", name)
	},
}

// 4. Subcomando: newkey
var newKeyCmd = &cobra.Command{
	Use:   "newkey",
	Short: "Genera una nueva llave",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔑 Generando nueva llave para: %s\n", name)
	},
}

// 5. Subcomando: newname
var newNameCmd = &cobra.Command{
	Use:   "newname",
	Short: "Cambia el nombre de la entidad",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" || newName == "" {
			fmt.Println("❌ Error: Se requieren las flags --name y --newname")
			return
		}
		fmt.Printf("📝 Renombrando %s a %s\n", name, newName)
	},
}

func init() {
	entityCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Nombre de la entidad (requerido)")
	newNameCmd.Flags().StringVarP(&newName, "newname", "m", "", "Nuevo nombre de la entidad")
	entityCmd.AddCommand(createEntityCmd, newKeyCmd, newNameCmd)
	rootCmd.AddCommand(entityCmd)
}
