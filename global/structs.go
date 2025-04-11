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

type EnvironmentType struct {
	NODOCB_PORT string `env:"NODOCB_PORT" required:"false"`

	NODOCB_CLUSTER_KEY   string `env:"NODOCB_CLUSTER_KEY" required:"false"`
	NODOCB_ENDPOINT_JOIN string `env:"NODOCB_ENDPOINT_JOIN" required:"false"`
}

type ConfigurationType struct {
	Port         string
	ClusterKey   []byte
	EndpointJoin string
}

type KeysType struct {
	Certificate struct {
		Private string `json:"private"`
		Public  string `json:"public"`
	} `json:"certificate"`
	Keys struct {
		Private string `json:"private"`
		Public  string `json:"public"`
	} `json:"keys"`
}

type MemberType struct {
	Organization string `json:"organization"`
	PublicCert   string `json:"public_cert"`
	PublicKey    string `json:"public_key"`
	Signature    string `json:"signature"`
}

// ----------------------------------
type ConfigType struct {
	Servers struct {
		Internal struct {
			Port      string `json:"port,omitempty"`
			TPS       int    `json:"tps,omitempty"`
			Backlog   int    `json:"backlog,omitempty"`
			SelfPath  string `json:"selfpath,omitempty"`
			Seguridad struct {
				PasswordHTTPBasico string   `json:"passwordHTTPBasico,omitempty"`
				HTTPS              bool     `json:"https,omitempty"`
				IPPermitidas       []string `json:"ipAllowed,omitempty"`
				CORS               string   `json:"cors,omitempty"`
			} `json:"security,omitempty"`
		} `json:"internal,omitempty"`
		External struct {
			Port string `json:"port,omitempty"`
			//HerramientasDiagnostico bool   `json:"herramientasDiagnostico,omitempty"`
		} `json:"external,omitempty"`
	} `json:"server,omitempty"`
	Identity struct {
		PKI    ConfigPKIType    `json:"pki,omitempty"`
		Keys   ConfigPKIType    `json:"keys,omitempty"`
		CUSTOM []CustomAuthType `json:"custom,omitempty"`
	} `json:"identity,omitempty"`
	Database struct {
		Path string `json:"path,omitempty"`
	} `json:"database,omitempty"`
	Log struct {
		Level      string `json:"nivel,omitempty"`
		Path       string `json:"ruta,omitempty"`
		Color      bool   `json:"color,omitempty"`
		Megabytes  int    `json:"megabytes,omitempty"`
		MaxBackups int    `json:"MaxBackups,omitempty"`
		MaxAge     int    `json:"diasMaximo,omitempty"`
	} `json:"log,omitempty"`
	Consumer []ConfigConsumerType `json:"consumer,omitempty"`
	Provider []ConfigProviderType `json:"provider,omitempty"`
}

type Oauth2CredentialType struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type ConfigPKIType struct {
	Private string `json:"private,omitempty"`
	Public  string `json:"public,omitempty"`
	CA      string `json:"ca,omitempty"`
}

type ConversionType struct {
	Entrada   string `json:"entrada,omitempty"`
	Respuesta string `json:"respuesta,omitempty"`
}

type CabeceraFijaType struct {
	Cabecera string `json:"cabecera,omitempty"`
	Valor    string `json:"valor,omitempty"`
}

type CustomAuthType struct {
	Name     string `json:"nombre,omitempty"`
	ID       string `json:"id,omitempty"`
	Servicio string `json:"servicio,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Script   string `json:"script,omitempty"`
}

type ConfigConsumerType struct {
	Name           string             `json:"nombre,omitempty"`
	LocalPath      string             `json:"rutaLocal,omitempty"`
	Timeout        int                `json:"timeout,omitempty"`
	Tramite        string             `json:"tramite,omitempty"`
	TramitePISEE   string             `json:"tramitePisee,omitempty"`
	ServicioPISEE  string             `json:"servicioPisee,omitempty"`
	Auth           string             `json:"autenticacion,omitempty"`
	Callback       string             `json:"callback,omitempty"`
	Conversion     ConversionType     `json:"conversion,omitempty"`
	CabecerasFijas []CabeceraFijaType `json:"cabecerasFijas,omitempty"`
	Endpoint       string
}

type ConfigProviderType struct {
	Name         string `json:"nombre,omitempty"`
	ExternalPath string `json:"rutaExterna,omitempty"`
	TPS          int    `json:"tps,omitempty"`
	FullDuplex   bool   `json:"fullduplex,omitempty"`
	Origen       struct {
		Tipo                  string `json:"tipo,omitempty"`
		ConfiguracionServicio string `json:"configuracionServicio,omitempty"`
		RutaLocal             string `json:"rutaLocal,omitempty"`
		DestinationFolder     string `json:"carpetaDestino,omitempty"`
	} `json:"origen,omitempty"`
}
