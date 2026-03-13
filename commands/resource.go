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
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Operaciones de recursos",
}

var resourceLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista los recursos remotos",
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
		resources := [][]string{}
		external := [][]string{}

		for _, n := range configuration.CM.Config.Networks {
			if n.Name == network {
				showResources(&resources, "API", n.Resources.API)
				showResources(&resources, "FILE", n.Resources.FILE)
				showResources(&resources, "DATASOURCE", n.Resources.DATASOURCE)

				showResources(&external, "API", n.RemoteResources.API)
				showResources(&external, "FILE", n.RemoteResources.FILE)
				showResources(&external, "DATASOURCE", n.RemoteResources.DATASOURCE)
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
			Headers("", "Tipo", "Nombre").
			Rows(resources...)

		t1 := table.New().
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
			Headers("", "Tipo", "Nombre").
			Rows(external...)

		fmt.Println("Recursos Provistos:")
		if len(resources) > 0 {
			fmt.Println(t.String())
		} else {
			fmt.Println("No existen recursos provistos")
		}

		fmt.Println("Recursos Externos:")
		if len(external) > 0 {
			fmt.Println(t1.String())
		} else {
			fmt.Println("No existen recursos externos")
		}
	},
}

func showResources(t *[][]string, tipo string, n []global.ResourceType) {
	start := len(*t) + 1
	for i, r := range n {
		*t = append(*t, []string{
			fmt.Sprintf("%d", i+start),
			tipo,
			r.Name,
		})
	}
}

var resourceAddCmd = &cobra.Command{
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

var resourceRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Elimina un recurso remoto",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	resourceLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	resourceAddCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "Ruta local del recurso (requerido)")

	resourceCmd.AddCommand(resourceAddCmd, resourceLsCmd, resourceRemoveCmd)
	rootCmd.AddCommand(resourceCmd)
}
