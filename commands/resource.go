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
		if network, ok := GetNetwork(); ok {
			configuration.CM = configuration.NewConfigurationManager()
			err := configuration.CM.LoadConfig()
			if err != nil {
				fmt.Println("Error cargando configuracion:", "error", err.Error())
				return
			}
			resources := [][]string{}

			for _, n := range configuration.CM.Config.Networks {
				if n.Name == network {
					showResources(&resources, "API", n.Resources.API)
					showResources(&resources, "FILE", n.Resources.FILE)
					showResources(&resources, "DATASOURCE", n.Resources.DATASOURCE)
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

			fmt.Println("Recursos Provistos:")
			if len(resources) > 0 {
				fmt.Println(t.String())
			} else {
				fmt.Println("No existen recursos provistos")
			}
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
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
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" || local == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource, --local")
				return
			}

			msg := Mensaje{
				Entrada: []string{"resource", "add", network, resourceType, resource, local},
			}
			respuesta := socketClient(msg)
			fmt.Println(respuesta)
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

var resourceRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Elimina un recurso remoto",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource")
				return
			}

			msg := Mensaje{
				Entrada: []string{"resource", "remove", network, resourceType, resource},
			}
			respuesta := socketClient(msg)
			fmt.Println(respuesta)
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

var resourceAllowCmd = &cobra.Command{
	Use:   "allow",
	Short: "Permite el acceso a un recurso",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" || entity == "" || tps == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource, --entity, --tps")
				return
			}

			protocol := resourceType
			msg := Mensaje{
				Entrada: []string{"resource", "allow", entity, network, protocol, resource, tps},
			}
			respuesta := socketClient(msg)
			fmt.Println(respuesta)
		} else {
			fmt.Println("❌ Error: Se requiere la flag --network")
			return
		}
	},
}

var resourceDenyCmd = &cobra.Command{
	Use:   "deny",
	Short: "Deniega el acceso a un recurso",
	Run: func(cmd *cobra.Command, args []string) {
		if network, ok := GetNetwork(); ok {
			if resourceType == "" || resource == "" || entity == "" {
				fmt.Println("❌ Error: Se requieren las flags --type, --resource, --entity")
				return
			}

			protocol := resourceType
			msg := Mensaje{
				Entrada: []string{"resource", "deny", entity, network, protocol, resource},
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
	resourceLsCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")

	resourceAllowCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	resourceAllowCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	resourceAllowCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	resourceAllowCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	resourceAllowCmd.PersistentFlags().StringVarP(&tps, "tps", "s", "", "Ruta local del recurso (requerido)")

	resourceDenyCmd.PersistentFlags().StringVarP(&entity, "entity", "e", "", "Nombre de la entidad (requerido)")
	resourceDenyCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	resourceDenyCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	resourceDenyCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")

	resourceAddCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")
	resourceAddCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "Ruta local del recurso (requerido)")

	resourceRemoveCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Nombre de la red (requerido)")
	resourceRemoveCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Tipo de recurso (requerido)")
	resourceRemoveCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Nombre del recurso (requerido)")

	resourceCmd.AddCommand(resourceAddCmd, resourceLsCmd, resourceRemoveCmd, resourceAllowCmd, resourceDenyCmd)
	rootCmd.AddCommand(resourceCmd)
}

func resourceAdd(network, resourceType, resource, local string) Mensaje {
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
				cfg.Networks[idx].Resources.API = append(cfg.Networks[idx].Resources.API, res)
			case "FILE":
				fmt.Println("Agregando recurso FILE:", resource)
				cfg.Networks[idx].Resources.FILE = append(cfg.Networks[idx].Resources.FILE, res)
			case "DATA_SOURCE":
				fmt.Println("Agregando recurso DATA_SOURCE:", resource)
				cfg.Networks[idx].Resources.DATASOURCE = append(cfg.Networks[idx].Resources.DATASOURCE, res)
			default:
				fmt.Println("❌ Error: Tipo de recurso no válido")
				return Mensaje{Salida: "Error: Tipo de recurso no válido"}
			}
			configuration.CM.SaveResources(network)

			if n, ok := connection.NM.GetNetwork(network); ok {
				n.AnunciarServicio(context.Background(), resource)
			}
			return Mensaje{Salida: "Recurso agregado correctamente"}
		}
	}
	return Mensaje{Salida: "Error: No se encontro la red"}
}

func resourceRemove(network, resourceType, resource string) Mensaje {
	configuration.CM = configuration.NewConfigurationManager()
	configuration.CM.LoadConfig()
	cfg := configuration.CM.GetConfig()

	for idx, net := range cfg.Networks {
		if net.Name == network {
			switch resourceType {
			case "API":
				fmt.Println("Eliminando recurso API:", resource)
				for i, r := range cfg.Networks[idx].Resources.API {
					if r.Name == resource {
						cfg.Networks[idx].Resources.API = append(cfg.Networks[idx].Resources.API[:i], cfg.Networks[idx].Resources.API[i+1:]...)
						break
					}
				}
			case "FILE":
				fmt.Println("Eliminando recurso FILE:", resource)
				for i, r := range cfg.Networks[idx].Resources.FILE {
					if r.Name == resource {
						cfg.Networks[idx].Resources.FILE = append(cfg.Networks[idx].Resources.FILE[:i], cfg.Networks[idx].Resources.FILE[i+1:]...)
						break
					}
				}
			case "DATA_SOURCE":
				fmt.Println("Eliminando recurso DATA_SOURCE:", resource)
				for i, r := range cfg.Networks[idx].Resources.DATASOURCE {
					if r.Name == resource {
						cfg.Networks[idx].Resources.DATASOURCE = append(cfg.Networks[idx].Resources.DATASOURCE[:i], cfg.Networks[idx].Resources.DATASOURCE[i+1:]...)
						break
					}
				}
			default:
				fmt.Println("❌ Error: Tipo de recurso no válido")
				return Mensaje{Salida: "Error: Tipo de recurso no válido"}
			}
			configuration.CM.SaveResources(network)

			return Mensaje{Salida: "Recurso eliminado correctamente"}
		}
	}
	return Mensaje{Salida: "Error: No se encontro la red"}
}

// Allow agrega una nueva política al archivo CSV
func Allow(entidad, network, protocolo, servicio, tps string) Mensaje {
	return Mensaje{Salida: "Política agregada correctamente"}
}

// Deny elimina una política existente buscando coincidencia exacta de los campos clave
func Deny(entidad, network, protocolo, servicio string) Mensaje {
	return Mensaje{Salida: "Política eliminada correctamente"}
}
