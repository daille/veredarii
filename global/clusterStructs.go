package global

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

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

type NodoMetaData struct {
	ID       string `json:"id"`
	Country  string `json:"country"`
	Organism string `json:"organism"`
	System   string `json:"system"`
}

type ClusterDataType struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	EndpointJoin string `json:"endpoint_join"`
	ClusterKey   []byte `json:"cluster_key"`
	MessageKey   string `json:"message_key"`
	Port         string `json:"port"`
}

type MemberDataType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoints   string `json:"endpoint_join"`
	ClusterKey  []byte `json:"cluster_key"`
	Port        string `json:"port"`
}
