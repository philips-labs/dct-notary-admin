package cmd

// VaultConfig is a copy of https://github.com/hashicorp/vault/blob/master/command/agent/config/config.go#L34
//
// as per https://github.com/hashicorp/vault/issues/9575 we are not supposed to depend on the
// toplevel hashicorp/vault module, which breaks go modules in version 1.5.0
type VaultConfig struct {
	Address          string      `hcl:"address"`
	CACert           string      `hcl:"ca_cert"`
	CAPath           string      `hcl:"ca_path"`
	TLSSkipVerify    bool        `hcl:"-"`
	TLSSkipVerifyRaw interface{} `hcl:"tls_skip_verify"`
	ClientCert       string      `hcl:"client_cert"`
	ClientKey        string      `hcl:"client_key"`
	TLSServerName    string      `hcl:"tls_server_name"`
}
