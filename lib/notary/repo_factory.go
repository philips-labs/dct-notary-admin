package notary

import (
	"net/http"

	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
)

const remoteConfigField = "api"

// RepoFactory takes a GUN and returns an initialized client.Repository, or an error.
type RepoFactory func(gun data.GUN) (client.Repository, error)

// ConfigureRepo takes in the configuration parameters and returns a repoFactory that can
// initialize new client.Repository objects with the correct upstreams and password
// retrieval mechanisms.
func ConfigureRepo(config *Config, retriever notary.PassRetriever, onlineOperation bool, permission httpAccess) RepoFactory {
	localRepo := func(gun data.GUN) (client.Repository, error) {
		var rt http.RoundTripper
		trustPin, err := getTrustPinning(config)
		if err != nil {
			return nil, err
		}
		if onlineOperation {
			rt, err = getTransport(config, gun, permission)
			if err != nil {
				return nil, err
			}
		}
		return client.NewFileCachedRepository(
			config.TrustDir,
			gun,
			config.RemoteServer.URL,
			rt,
			retriever,
			trustPin,
		)
	}

	return localRepo
}
