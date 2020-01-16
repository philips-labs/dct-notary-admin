package lib

// ServerConfig holds configuration options
type ServerConfig struct {
	ListenAddr    string `json:"listen_addr" mapstructure:"listen_addr"`
	ListenAddrTLS string `json:"listen_addr_tls" mapstructure:"listen_addr_tls"`
}
