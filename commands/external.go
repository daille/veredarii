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

var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "Operaciones de recursos externos",
}

var externalLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista los recursos remotos",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			configuration.CM = configuration.NewConfigurationManager()
			err := configuration.CM.LoadConfig()
			if err != nil {
				fmt.Println("Error cargando configuracion:", "error", err.Error())
				return
			}
			external := [][]string{}

			for _, n := range configuration.CM.Config.Networks {
				if n.Name == network {
					showResources(&external, "API", n.RemoteResources.API)
					showResources(&external, "FILE", n.RemoteResources.FILE)
					showResources(&external, "DATASOURCE", n.RemoteResources.DATASOURCE)
					break
				}
			}

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

			fmt.Println("Recursos Externos:")
			if len(external) > 0 {
				fmt.Println(t1.String())
			} else {
				fmt.Println("No existen recursos externos")
			}
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

var externalAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Agrega un recurso remoto externo",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" || local == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource, --local")
				return
			}

			msg := Mensaje{
				Entrada: []string{"external", "add", network, resourceType, resource, local},
			}
			respuesta := socketClient(msg)
			fmt.Println(respuesta)
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

var externalRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Elimina un recurso remoto externo",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource")
				return
			}

			msg := Mensaje{
				Entrada: []string{"external", "remove", network, resourceType, resource},
			}
			respuesta := socketClient(msg)
			fmt.Println(respuesta)
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

func init() {
	externalLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	externalAddCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	externalAddCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	externalAddCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	externalAddCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "Ruta local del recurso (requerido)")

	externalRemoveCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	externalRemoveCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	externalRemoveCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")

	externalCmd.AddCommand(externalAddCmd, externalLsCmd, externalRemoveCmd)
	rootCmd.AddCommand(externalCmd)
}

func resourceExternalAdd(network, resourceType, resource, local string) Mensaje {
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
				fmt.Println("Agregando recurso API:", resource)
				cfg.Networks[idx].RemoteResources.API = append(cfg.Networks[idx].RemoteResources.API, res)
			case "FILE":
				fmt.Println("Agregando recurso FILE:", resource)
				cfg.Networks[idx].RemoteResources.FILE = append(cfg.Networks[idx].RemoteResources.FILE, res)
			case "DATA_SOURCE":
				fmt.Println("Agregando recurso DATA_SOURCE:", resource)
				cfg.Networks[idx].RemoteResources.DATASOURCE = append(cfg.Networks[idx].RemoteResources.DATASOURCE, res)
			default:
				fmt.Println("❌ Error: Tipo de recurso no válido")
				return Mensaje{Salida: "Error: Tipo de recurso no válido"}
			}
			configuration.CM.SaveRemoteResources(network)

			return Mensaje{Salida: "Recurso agregado correctamente"}
		}
	}
	return Mensaje{Salida: "Error: No se encontro la red"}
}

func resourceExternalRemove(network, resourceType, resource string) Mensaje {
	configuration.CM = configuration.NewConfigurationManager()
	configuration.CM.LoadConfig()
	cfg := configuration.CM.GetConfig()

	for idx, net := range cfg.Networks {
		if net.Name == network {
			switch resourceType {
			case "API":
				fmt.Println("Eliminando recurso API:", resource)
				for i, r := range cfg.Networks[idx].RemoteResources.API {
					if r.Name == resource {
						cfg.Networks[idx].RemoteResources.API = append(cfg.Networks[idx].RemoteResources.API[:i], cfg.Networks[idx].RemoteResources.API[i+1:]...)
						break
					}
				}
			case "FILE":
				fmt.Println("Eliminando recurso FILE:", resource)
				for i, r := range cfg.Networks[idx].RemoteResources.FILE {
					if r.Name == resource {
						cfg.Networks[idx].RemoteResources.FILE = append(cfg.Networks[idx].RemoteResources.FILE[:i], cfg.Networks[idx].RemoteResources.FILE[i+1:]...)
						break
					}
				}
			case "DATA_SOURCE":
				fmt.Println("Eliminando recurso DATA_SOURCE:", resource)
				for i, r := range cfg.Networks[idx].RemoteResources.DATASOURCE {
					if r.Name == resource {
						cfg.Networks[idx].RemoteResources.DATASOURCE = append(cfg.Networks[idx].RemoteResources.DATASOURCE[:i], cfg.Networks[idx].RemoteResources.DATASOURCE[i+1:]...)
						break
					}
				}
			default:
				fmt.Println("❌ Error: Tipo de recurso no válido")
				return Mensaje{Salida: "Error: Tipo de recurso no válido"}
			}
			configuration.CM.SaveRemoteResources(network)

			return Mensaje{Salida: "Recurso eliminado correctamente"}
		}
	}
	return Mensaje{Salida: "Error: No se encontro la red"}
}
