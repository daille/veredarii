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

func init() {
	entityLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	entityCmd.AddCommand(entityLsCmd)
	rootCmd.AddCommand(entityCmd)
}
