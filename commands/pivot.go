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
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var pivotCmd = &cobra.Command{
	Use:   "pivot",
	Short: "Maneja los pivotes de la red",
	Long:  `Maneja los pivotes de la red`,
}

var pivotLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista las redes",
	Long:  `Lista los redes`,
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
		resultado := [][]string{}
		for _, n := range configuration.CM.Config.Networks {
			if n.Name == network {
				for i, p := range n.Pivots {
					resultado = append(resultado, []string{strconv.Itoa(i), p})
				}
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
			Headers("", "Ruta del Pivote").
			Rows(resultado...)

		if len(resultado) > 0 {
			fmt.Println(t.String())
		} else {
			fmt.Println("No existen pivotes")
		}
	},
}

var pivotAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Agrega un pivote",
	Long:  `Agrega un pivote`,
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" {
			fmt.Println("❌ Error: Se requieren las flags --network")
			return
		}
		if pivot == "" {
			fmt.Println("❌ Error: Se requieren las flags --pivot")
			return
		}

		msg := Mensaje{
			Entrada: []string{"pivot", "add", network, pivot},
		}
		respuesta := socketClient(msg)
		fmt.Println(respuesta)
	},
}

var pivotRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remueve un pivote",
	Long:  `Remueve un pivote`,
	Run: func(cmd *cobra.Command, args []string) {
		if network == "" {
			fmt.Println("❌ Error: Se requieren las flags --network")
			return
		}
		if pivot == "" {
			fmt.Println("❌ Error: Se requieren las flags --pivot")
			return
		}

		msg := Mensaje{
			Entrada: []string{"pivot", "remove", network, pivot},
		}
		respuesta := socketClient(msg)
		fmt.Println(respuesta)
	},
}

var pivot string

func init() {
	pivotLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	pivotCmd.AddCommand(pivotLsCmd)
	pivotAddCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	pivotAddCmd.PersistentFlags().StringVarP(&pivot, "pivot", "p", "", "Ruta del pivote (requerido)")
	pivotCmd.AddCommand(pivotAddCmd)
	pivotRemoveCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	pivotRemoveCmd.PersistentFlags().StringVarP(&pivot, "pivot", "p", "", "Ruta del pivote (requerido)")
	pivotCmd.AddCommand(pivotRemoveCmd)
	rootCmd.AddCommand(pivotCmd)
}

func pivotAdd(network, pivot string) Mensaje {
	configuration.CM = configuration.NewConfigurationManager()
	err := configuration.CM.LoadConfig()
	if err != nil {
		return Mensaje{Salida: "Error cargando configuracion:" + err.Error()}
	}
	for idx, nn := range configuration.CM.Config.Networks {
		if nn.Name == network {
			configuration.CM.Config.Networks[idx].Pivots = append(configuration.CM.Config.Networks[idx].Pivots, pivot)
			break
		}
	}
	err = configuration.CM.Save()
	if err != nil {
		return Mensaje{Salida: "Error guardando configuracion:" + err.Error()}
	}
	return Mensaje{Salida: "Pivote agregado correctamente"}
}

func pivotRemove(network, pivot string) Mensaje {
	configuration.CM = configuration.NewConfigurationManager()
	err := configuration.CM.LoadConfig()
	if err != nil {
		return Mensaje{Salida: "Error cargando configuracion:" + err.Error()}
	}
	for idx, nn := range configuration.CM.Config.Networks {
		if nn.Name == network {
			for i, p := range nn.Pivots {
				if p == pivot {
					configuration.CM.Config.Networks[idx].Pivots = append(configuration.CM.Config.Networks[idx].Pivots[:i], configuration.CM.Config.Networks[idx].Pivots[i+1:]...)
					break
				}
			}
			break
		}
	}
	err = configuration.CM.Save()
	if err != nil {
		return Mensaje{Salida: "Error guardando configuracion:" + err.Error()}
	}
	return Mensaje{Salida: "Pivote removido correctamente"}
}
