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
	"os"

	log "github.com/sirupsen/logrus"
)

var CM *ConfigurationManager

type ConfigurationManager struct {
	Config *global.ConfigType
}

const filename = "config.json"

func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{
		Config: &global.ConfigType{},
	}
}

func (cm *ConfigurationManager) LoadConfig() error {
	var err error
	if err = cm.loadJson(filename, cm.Config); err != nil {
		log.Error("Error cargando configuracion:", err)
		return err
	}

	for idx, network := range cm.Config.Networks {
		// resources
		if err = cm.loadJson(network.ResourcesPath, &cm.Config.Networks[idx].Resources); err != nil {
			log.Error("Error cargando recursos:", err)
			return err
		}

		// remote resources
		if err = cm.loadJson(network.RemoteResourcesPath, &cm.Config.Networks[idx].RemoteResources); err != nil {
			log.Error("Error cargando recursos remotos:", err)
			return err
		}
	}

	return nil
}

func (cm *ConfigurationManager) GetConfig() *global.ConfigType {
	return cm.Config
}

func (cm *ConfigurationManager) loadJson(file string, obj interface{}) error {
	buf, err := os.ReadFile(file)
	if err != nil {
		log.Error("Error leyendo archivo:", file, err)
		return err
	}

	err = json.Unmarshal(buf, obj)
	if err != nil {
		log.Error("Error deserializando archivo:", err)
		return err
	}

	return nil
}
