package general

import (
	"encoding/json"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

/**
 *    Veredarii, software for interoperability.
 *    This file is part of Veredarii.
 *
 *    @author jcDaille
 *
 *
 *    MIT License
 *
 * Copyright (c) 2025 JC Daille
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

type ConfigurationType struct {
	Identification struct {
		Entity string `json:"entity"`
		Peer   string `json:"peer"`
	} `json:"identification"`
	Behavior struct {
		Local struct {
			Port string `json:"port"`
		} `json:"local"`
		Database struct {
			Path string `json:"path"`
		} `json:"database"`
		Log struct {
			Level      string `json:"level"`
			Path       string `json:"path"`
			Color      bool   `json:"color"`
			Megabytes  int    `json:"megabytes"`
			MaxBackups int    `json:"MaxBackups"`
			MaxDays    int    `json:"MaxDays"`
		} `json:"log"`
	} `json:"behavior"`
	Networks []NetworkType `json:"networks"`
}

type NetworkType struct {
	Port       string   `json:"port"`
	Path       string   `json:"path"`
	KeyNetwork string   `json:"keyNetwork"`
	Whitelist  []string `json:"whitelist"`
	Pivots     []string `json:"pivots"`
}

func Load(file string) {
	log.Debug(T("Loading_configuration"))
	jsonFile, err := os.Open(file)
	if err != nil {
		log.Debug(T("configuration_error"), ":", err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &Configuration)
	if err != nil {
		log.Error(err)
	}

	/*var tmp g.EnvironmentType
	if err := env.Set(&tmp); err != nil {
		log.Error(u.Fatal(err))
	} else {
		log.Debug("Lectura de variables de entorno")
	}*/
	log.Debug(T("configuration"), ":", Configuration)
}
