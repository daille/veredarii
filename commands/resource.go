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
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

			connection.AddResourceAccessControl(network, resourceType, resource)
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

			connection.RemoveResourceAccessControl(network, resourceType, resource)
			return Mensaje{Salida: "Recurso eliminado correctamente"}
		}
	}
	return Mensaje{Salida: "Error: No se encontro la red"}
}

// Allow agrega una nueva política al archivo CSV
func Allow(entidad, network, protocolo, servicio, tps string) Mensaje {
	policyFile := network + ".policy"
	entitiesFile := network + ".entities"
	protocolo = connection.NormalizeCedarType(protocolo)
	entidad = strings.ToLower(entidad)
	servicio = strings.ToLower(servicio)

	// --- Entidades (igual que arriba) ---
	if err := updateEntities(entitiesFile, entidad, protocolo, servicio, network, tps); err != nil {
		return Mensaje{Salida: err.Error()}
	}

	// --- Policy: agregar bloque si no existe ---
	policyBytes, _ := os.ReadFile(policyFile)
	policyContent := string(policyBytes)

	newResource := fmt.Sprintf(`%s::"%s"`, protocolo, servicio)
	newPrincipal := fmt.Sprintf(`User::"%s"`, entidad)

	// Verificar si ya existe exactamente este bloque
	checkStr := fmt.Sprintf("principal == %s", newPrincipal)
	if strings.Contains(policyContent, checkStr) && strings.Contains(policyContent, newResource) {
		return Mensaje{Salida: "El servicio ya está permitido para este usuario"}
	}

	newBlock := fmt.Sprintf(`
permit (principal, action == Action::"call", resource)
when {
    principal == %s &&
    resource.network == "%s" &&
    resource == %s
};
`, newPrincipal, network, newResource)

	policyContent += newBlock

	if err := os.WriteFile(policyFile, []byte(policyContent), 0644); err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error escribiendo policy: %v", err)}
	}

	connection.SetupAccessControl()
	return Mensaje{Salida: "Política agregada correctamente"}
}

// Deny elimina una política existente buscando coincidencia exacta de los campos clave
func Deny(entidad, network, protocolo, servicio string) Mensaje {
	policyFile := network + ".policy"
	entitiesFile := network + ".entities"
	protocolo = connection.NormalizeCedarType(protocolo)
	entidad = strings.ToLower(entidad)
	servicio = strings.ToLower(servicio)

	// --- 1. Eliminar bloque del .policy ---
	policyBytes, err := os.ReadFile(policyFile)
	if err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error leyendo policy: %v", err)}
	}

	newPrincipal := fmt.Sprintf(`User::"%s"`, entidad)
	newResource := fmt.Sprintf(`%s::"%s"`, protocolo, servicio)

	// Separar en bloques por ";" y filtrar el que coincide
	blocks := strings.Split(string(policyBytes), "};")
	var kept []string
	found := false
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			continue
		}
		// Si el bloque contiene AMBOS (principal y recurso), es el que eliminamos
		if strings.Contains(block, newPrincipal) && strings.Contains(block, newResource) {
			found = true
			continue
		}
		kept = append(kept, trimmed)
	}

	if !found {
		return Mensaje{Salida: fmt.Sprintf("No se encontró policy para %s sobre %s", entidad, servicio)}
	}

	newPolicyContent := strings.Join(kept, "\n};\n")
	if len(kept) > 0 {
		newPolicyContent += "\n};"
	}

	if err := os.WriteFile(policyFile, []byte(newPolicyContent+"\n"), 0644); err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error escribiendo policy: %v", err)}
	}

	// --- 2. Actualizar entities.json ---
	entitiesBytes, err := os.ReadFile(entitiesFile)
	if err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error leyendo entidades: %v", err)}
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(entitiesBytes, &entities); err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error parseando entidades: %v", err)}
	}

	// Quitar el servicio de las quotas del usuario
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
		if uid["type"] == "User" && existingID == entidad {
			if attrs, ok := e["attrs"].(map[string]interface{}); ok {
				if quotas, ok := attrs["quotas"].(map[string]interface{}); ok {
					delete(quotas, servicio)
				}
			}
		}
	}

	// Quitar el recurso si ya ningún usuario tiene quota sobre él
	servicioUsado := false
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		if uid["type"] == "User" {
			if attrs, ok := e["attrs"].(map[string]interface{}); ok {
				if quotas, ok := attrs["quotas"].(map[string]interface{}); ok {
					if _, exists := quotas[servicio]; exists {
						servicioUsado = true
						break
					}
				}
			}
		}
	}

	if !servicioUsado {
		// Eliminar la entidad del recurso
		var filteredEntities []map[string]interface{}
		for _, e := range entities {
			uid := e["uid"].(map[string]interface{})
			existingID := strings.ToLower(fmt.Sprintf("%v", uid["id"]))
			existingType := strings.ToLower(fmt.Sprintf("%v", uid["type"]))
			if existingType == strings.ToLower(protocolo) && existingID == servicio {
				continue // eliminar
			}
			filteredEntities = append(filteredEntities, e)
		}
		entities = filteredEntities
	}

	out, err := json.MarshalIndent(entities, "", "    ")
	if err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error serializando entidades: %v", err)}
	}
	if err := os.WriteFile(entitiesFile, out, 0644); err != nil {
		return Mensaje{Salida: fmt.Sprintf("Error escribiendo entidades: %v", err)}
	}

	connection.SetupAccessControl()
	return Mensaje{Salida: "Política eliminada correctamente"}
}

func updateEntities(file, entidad, protocolo, servicio, network, tps string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error leyendo entidades: %v", err)
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal(data, &entities); err != nil {
		return fmt.Errorf("error parseando entidades: %v", err)
	}

	tpsInt, _ := strconv.Atoi(tps)

	// Usuario
	userFound := false
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		if uid["type"] == "User" && uid["id"] == entidad {
			userFound = true
			attrs := e["attrs"].(map[string]interface{})
			quotas, ok := attrs["quotas"].(map[string]interface{})
			if !ok {
				quotas = map[string]interface{}{}
				attrs["quotas"] = quotas
			}
			quotas[servicio] = tpsInt
		}
	}
	if !userFound {
		entities = append(entities, map[string]interface{}{
			"uid":     map[string]interface{}{"type": "User", "id": entidad},
			"attrs":   map[string]interface{}{"quotas": map[string]interface{}{servicio: tpsInt}},
			"parents": []interface{}{},
		})
	}

	// Recurso
	resFound := false
	for _, e := range entities {
		uid := e["uid"].(map[string]interface{})
		if uid["type"] == protocolo && uid["id"] == servicio {
			resFound = true
		}
	}
	if !resFound {
		entities = append(entities, map[string]interface{}{
			"uid":     map[string]interface{}{"type": protocolo, "id": servicio},
			"attrs":   map[string]interface{}{"network": network},
			"parents": []interface{}{},
		})
	}

	out, _ := json.MarshalIndent(entities, "", "    ")
	return os.WriteFile(file, out, 0644)
}
