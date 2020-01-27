package notary

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/theupdateframework/notary"
	notaryclient "github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/cryptoservice"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
	tufutils "github.com/theupdateframework/notary/tuf/utils"
	"go.uber.org/zap"
)

// importRootKey imports the root key from path then adds the key to repo
// returns key ids
func importRootKey(log *zap.Logger, rootKey string, nRepo notaryclient.Repository, retriever notary.PassRetriever) ([]string, error) {
	var rootKeyList []string

	if rootKey != "" {
		privKey, err := readKey(data.CanonicalRootRole, rootKey, retriever)
		if err != nil {
			return nil, err
		}
		// add root key to repo
		err = nRepo.GetCryptoService().AddKey(data.CanonicalRootRole, "", privKey)
		if err != nil {
			return nil, fmt.Errorf("Error importing key: %v", err)
		}
		rootKeyList = []string{privKey.ID()}
	} else {
		rootKeyList = nRepo.GetCryptoService().ListKeys(data.CanonicalRootRole)
	}

	if len(rootKeyList) > 0 {
		// Chooses the first root key available, which is initialization specific
		// but should return the HW one first.
		rootKeyID := rootKeyList[0]
		log.Info("Root key found", zap.String("rootKeyID", rootKeyID))

		return []string{rootKeyID}, nil
	}

	return []string{}, nil
}

// importRootCert imports the base64 encoded public certificate corresponding to the root key
// returns empty slice if path is empty
func importRootCert(certFilePath string) ([]data.PublicKey, error) {
	publicKeys := make([]data.PublicKey, 0, 1)

	if certFilePath == "" {
		return publicKeys, nil
	}

	// read certificate from file
	certPEM, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file: %v", err)
	}
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("the provided file does not contain a valid PEM certificate %v", err)
	}

	// convert the file to data.PublicKey
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Parsing certificate PEM bytes to x509 certificate: %v", err)
	}
	publicKeys = append(publicKeys, tufutils.CertToKey(cert))

	return publicKeys, nil
}

// Attempt to read a role key from a file, and return it as a data.PrivateKey
// If key is for the Root role, it must be encrypted
func readKey(role data.RoleName, keyFilename string, retriever notary.PassRetriever) (data.PrivateKey, error) {
	pemBytes, err := ioutil.ReadFile(keyFilename)
	if err != nil {
		return nil, fmt.Errorf("Error reading input root key file: %v", err)
	}
	isEncrypted := true
	if err = cryptoservice.CheckRootKeyIsEncrypted(pemBytes); err != nil {
		if role == data.CanonicalRootRole {
			return nil, err
		}
		isEncrypted = false
	}
	var privKey data.PrivateKey
	if isEncrypted {
		privKey, _, err = trustmanager.GetPasswdDecryptBytes(retriever, pemBytes, "", data.CanonicalRootRole.String())
	} else {
		privKey, err = tufutils.ParsePEMPrivateKey(pemBytes, "")
	}
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

type passwordStore struct {
	anonymous bool
}

func (ps passwordStore) Basic(u *url.URL) (string, string) {
	// if it's not a terminal, don't wait on input
	if ps.anonymous {
		return "", ""
	}

	auth := os.Getenv("NOTARY_AUTH")
	if auth != "" {
		dec, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			// logrus.Error("Could not base64-decode authentication string")
			return "", ""
		}
		plain := string(dec)

		i := strings.Index(plain, ":")
		if i == 0 {
			// logrus.Error("Authentication string with zero-length username")
			return "", ""
		} else if i > -1 {
			username := plain[:i]
			password := plain[i+1:]
			password = strings.TrimSpace(password)
			return username, password
		}

		// logrus.Error("Malformatted authentication string; format must be <username>:<password>")
		return "", ""
	}

	return "", ""
}

// to comply with the CredentialStore interface
func (ps passwordStore) RefreshToken(u *url.URL, service string) string {
	return ""
}

// to comply with the CredentialStore interface
func (ps passwordStore) SetRefreshToken(u *url.URL, service string, token string) {
}

type httpAccess int

const (
	readOnly httpAccess = iota
	readWrite
	admin
)

// It correctly handles the auth challenge/credentials required to interact
// with a notary server over both HTTP Basic Auth and the JWT auth implemented
// in the notary-server
// The readOnly flag indicates if the operation should be performed as an
// anonymous read only operation. If the command entered requires write
// permissions on the server, readOnly must be false
func getTransport(config *Config, gun data.GUN, permission httpAccess) (http.RoundTripper, error) {
	// Attempt to get a root CA from the config file. Nil is the host defaults.
	rootCAFile := config.RemoteServer.RootCA
	clientCert := config.RemoteServer.TLSClientCert
	clientKey := config.RemoteServer.TLSClientKey
	insecureSkipVerify := config.RemoteServer.SkipTLSVerify
	trustServerURL := config.RemoteServer.URL

	if clientCert == "" && clientKey != "" || clientCert != "" && clientKey == "" {
		return nil, fmt.Errorf("either pass both client key and cert, or neither")
	}

	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		CAFile:             rootCAFile,
		InsecureSkipVerify: insecureSkipVerify,
		CertFile:           clientCert,
		KeyFile:            clientKey,
		ExclusiveRootPools: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to configure TLS: %s", err.Error())
	}

	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
		DisableKeepAlives:   true,
	}

	return tokenAuth(trustServerURL, base, gun, permission)
}

func tokenAuth(trustServerURL string, baseTransport *http.Transport, gun data.GUN,
	permission httpAccess) (http.RoundTripper, error) {

	// TODO(dmcgowan): add notary specific headers
	authTransport := transport.NewTransport(baseTransport)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
	}
	endpoint, err := url.Parse(trustServerURL)
	if err != nil {
		return nil, fmt.Errorf("Could not parse remote trust server url (%s): %s", trustServerURL, err.Error())
	}
	if endpoint.Scheme == "" {
		return nil, fmt.Errorf("Trust server url has to be in the form of http(s)://URL:PORT. Got: %s", trustServerURL)
	}
	subPath, err := url.Parse(path.Join(endpoint.Path, "/v2") + "/")
	if err != nil {
		return nil, fmt.Errorf("Failed to parse v2 subpath. This error should not have been reached. Please report it as an issue at https://github.com/theupdateframework/notary/issues: %s", err.Error())
	}
	endpoint = endpoint.ResolveReference(subPath)
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := pingClient.Do(req)
	if err != nil {
		// logrus.Errorf("could not reach %s: %s", trustServerURL, err.Error())
		// logrus.Info("continuing in offline mode")
		return nil, nil
	}
	// non-nil err means we must close body
	defer resp.Body.Close()
	if (resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices) &&
		resp.StatusCode != http.StatusUnauthorized {
		// If we didn't get a 2XX range or 401 status code, we're not talking to a notary server.
		// The http client should be configured to handle redirects so at this point, 3XX is
		// not a valid status code.
		// logrus.Errorf("could not reach %s: %d", trustServerURL, resp.StatusCode)
		// logrus.Info("continuing in offline mode")
		return nil, nil
	}

	challengeManager := challenge.NewSimpleManager()
	if err := challengeManager.AddResponse(resp); err != nil {
		return nil, err
	}

	ps := passwordStore{anonymous: permission == readOnly}

	var actions []string
	switch permission {
	case admin:
		actions = []string{"*"}
	case readWrite:
		actions = []string{"push", "pull"}
	case readOnly:
		actions = []string{"pull"}
	default:
		return nil, fmt.Errorf("Invalid permission requested for token authentication of gun %s", gun)
	}

	tokenHandler := auth.NewTokenHandler(authTransport, ps, gun.String(), actions...)
	basicHandler := auth.NewBasicHandler(ps)

	modifier := auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler)

	if permission != readOnly {
		return newAuthRoundTripper(transport.NewTransport(baseTransport, modifier)), nil
	}

	// Try to authenticate read only repositories using basic username/password authentication
	return newAuthRoundTripper(transport.NewTransport(baseTransport, modifier),
		transport.NewTransport(baseTransport, auth.NewAuthorizer(challengeManager, auth.NewTokenHandler(authTransport, passwordStore{anonymous: false}, gun.String(), actions...)))), nil
}

func getTrustPinning(config *Config) (trustpinning.TrustPinConfig, error) {
	var ok bool
	// Need to parse out Certs section from config
	certMap := config.TrustPinning.Certs
	resultCertMap := make(map[string][]string)
	for gun, certSlice := range certMap {
		var castedCertSlice []interface{}
		if castedCertSlice, ok = certSlice.([]interface{}); !ok {
			return trustpinning.TrustPinConfig{}, fmt.Errorf("invalid format for trust_pinning.certs")
		}
		certsForGun := make([]string, len(castedCertSlice))
		for idx, certIDInterface := range castedCertSlice {
			if certID, ok := certIDInterface.(string); ok {
				certsForGun[idx] = certID
			} else {
				return trustpinning.TrustPinConfig{}, fmt.Errorf("invalid format for trust_pinning.certs")
			}
		}
		resultCertMap[gun] = certsForGun
	}
	return trustpinning.TrustPinConfig{
		DisableTOFU: config.TrustPinning.DisableTofu,
		CA:          config.TrustPinning.CA,
		Certs:       resultCertMap,
	}, nil
}

// authRoundTripper tries to authenticate the requests via multiple HTTP transactions (until first succeed)
type authRoundTripper struct {
	trippers []http.RoundTripper
}

func newAuthRoundTripper(trippers ...http.RoundTripper) http.RoundTripper {
	return &authRoundTripper{trippers: trippers}
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	var resp *http.Response
	// Try all run all transactions
	for _, t := range a.trippers {
		var err error
		resp, err = t.RoundTrip(req)
		// Reject on error
		if err != nil {
			return resp, err
		}

		// Stop when request is authorized/unknown error
		if resp.StatusCode != http.StatusUnauthorized {
			return resp, nil
		}
	}

	// Return the last response
	return resp, nil
}

func maybeAutoPublish(log *zap.Logger, doPublish bool, gun data.GUN, config *Config, passRetriever notary.PassRetriever) error {

	if !doPublish {
		return nil
	}

	// We need to set up a http RoundTripper when publishing
	rt, err := getTransport(config, gun, readWrite)
	if err != nil {
		return err
	}

	trustPin, err := getTrustPinning(config)
	if err != nil {
		return err
	}

	nRepo, err := notaryclient.NewFileCachedRepository(config.TrustDir, gun, config.RemoteServer.URL, rt, passRetriever, trustPin)
	if err != nil {
		return err
	}

	log.Info("Auto-publishing changes", zap.Stringer("gun", nRepo.GetGUN()))
	return nRepo.Publish()
}
