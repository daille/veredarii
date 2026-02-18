package global

/*
MIT License

# Copyright (c) 2026 Juan Carlos Daille

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
type KVType struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type TopicType struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type IdentityType struct {
	Entity      string `json:"entity"`
	PrivKeyFile string `json:"priv_key_file"`
}

type LocalInterfaceType struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
}

type RemoteResourceType struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ConfigType struct {
	Identity       IdentityType       `json:"identity"`
	LocalInterface LocalInterfaceType `json:"localInterface"`
	Networks       []NetworkType      `json:"networks"`
}

type NetworkType struct {
	Port                string        `json:"port"`
	FS                  string        `json:"filesystem"`
	Name                string        `json:"name"`
	Pivots              []string      `json:"pivots"`
	NetworkKey          string        `json:"network_key"`
	MyAddress           []string      `json:"myAddress"`
	Entities            []KVType      `json:"entities"`
	Topics              []TopicType   `json:"topics"`
	RemoteResourcesPath string        `json:"remote_resources"`
	RemoteResources     ResourcesType `json:"-"`
	ResourcesPath       string        `json:"resources"`
	Resources           ResourcesType `json:"-"`
}
