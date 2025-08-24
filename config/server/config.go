package config

type Config struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Name      string     `json:"name"`
	Listen    Listen     `json:"listen"`
	Hosts     []string   `json:"hosts"`
	AccessLog string     `json:"access_log"`
	SSL       *SSLConfig `json:"ssl,omitempty"`
	Locations []Location `json:"locations"`
}

type Listen struct {
	Port int  `json:"port"`
	SSL  bool `json:"ssl"`
}

type SSLConfig struct {
	Certificate    string `json:"certificate"`
	CertificateKey string `json:"certificate_key"`
}

type Location struct {
	Path      string `json:"path"`
	Root      string `json:"root,omitempty"`
	ProxyPass string `json:"proxy_pass,omitempty"`
}
