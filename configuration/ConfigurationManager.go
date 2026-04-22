package configuration

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
	global "Veredarii/global"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

var CM *ConfigurationManager
var DefaultNetwork string

type ConfigurationManager struct {
	Config *global.ConfigType
}

const ConfigFilename = "config.json"

func NewConfigurationManager() *ConfigurationManager {
	DefaultNetwork = ""
	return &ConfigurationManager{
		Config: &global.ConfigType{},
	}
}

func (cm *ConfigurationManager) NewConfig() *global.ConfigType {
	cm.Config = &global.ConfigType{}
	return cm.Config
}

func (cm *ConfigurationManager) LoadConfig() error {
	var err error
	if err = cm.loadJson(ConfigFilename, cm.Config); err != nil {
		slog.Error("Error cargando configuracion", "error", err)
		return err
	}

	for idx, network := range cm.Config.Networks {
		// resources
		if network.ResourcesPath != "" {
			if err = cm.loadJson(network.ResourcesPath, &cm.Config.Networks[idx].Resources); err != nil {
				slog.Error("Error cargando recursos", "error", err)
				return err
			}
		}

		// remote resources
		if network.RemoteResourcesPath != "" {
			if err = cm.loadJson(network.RemoteResourcesPath, &cm.Config.Networks[idx].RemoteResources); err != nil {
				slog.Error("Error cargando recursos remotos", "error", err)
				return err
			}
		}
	}

	// Load environment variables
	cm.Config.Networks[0].NetworkKey = os.Getenv("NETWORK_KEY")

	return nil
}

func (cm *ConfigurationManager) GetConfig() *global.ConfigType {
	return cm.Config
}

func (cm *ConfigurationManager) loadJson(file string, obj interface{}) error {
	buf, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Error leyendo archivo", "file", file, "error", err)
		return err
	}

	err = json.Unmarshal(buf, obj)
	if err != nil {
		slog.Error("Error deserializando archivo:", "file", file, "error", err)
		return err
	}

	return nil
}
func (cm *ConfigurationManager) AddEntity(network string, entity global.KVType) {
	for idx, nn := range cm.Config.Networks {
		if nn.Name == network {
			cm.Config.Networks[idx].Entities = append(cm.Config.Networks[idx].Entities, entity)
			cm.Save()
			break
		}
	}
}

func (cm *ConfigurationManager) Save() error {
	buf, err := json.MarshalIndent(cm.Config, "", "    ")
	if err != nil {
		slog.Error("Error serializando archivo:", "file", ConfigFilename, "error", err)
		return err
	}

	return os.WriteFile(ConfigFilename, buf, 0644)
}

func (cm *ConfigurationManager) SaveResources(network string) error {
	fmt.Println("......", network)
	if network == "" {
		return errors.New("red no especificada")
	}

	for idx, nn := range cm.Config.Networks {
		if nn.Name == network {
			resourcesPath := "./resources_" + network + ".json"
			buf, err := json.MarshalIndent(cm.Config.Networks[idx].Resources, "", "    ")
			if err != nil {
				slog.Error("Error serializando archivo:", "file", resourcesPath, "error", err)
				return err
			}
			return os.WriteFile(resourcesPath, buf, 0644)
		}
	}

	return errors.New("red no encontrada")
}

func (cm *ConfigurationManager) SaveRemoteResources(network string) error {
	if network == "" {
		return errors.New("red no especificada")
	}

	for idx, nn := range cm.Config.Networks {
		if nn.Name == network {
			remoteResourcesPath := "./remote_resources_" + network + ".json"
			buf, err := json.MarshalIndent(cm.Config.Networks[idx].RemoteResources, "", "    ")
			if err != nil {
				slog.Error("Error serializando archivo:", "file", remoteResourcesPath, "error", err)
				return err
			}
			return os.WriteFile(remoteResourcesPath, buf, 0644)
		}
	}

	return errors.New("red no encontrada")
}
