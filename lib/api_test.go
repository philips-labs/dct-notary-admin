package lib

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"

	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

type registeredRoute struct {
	method string
	route  string
}

func bootstrapAPI() *chi.Mux {
	n := notary.NewService(&notary.Config{
		TrustDir: "./.notary",
		RemoteServer: notary.RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, notary.GetPassphraseRetriever(), zap.NewNop())
	return configureAPI(n, zap.NewNop())
}

func TestRoutes(t *testing.T) {
	assert := assert.New(t)
	expectedRoutes := []registeredRoute{
		{http.MethodGet, "/"},
		{http.MethodGet, "/ping"},
		{http.MethodGet, "/api/targets/"},
		{http.MethodPost, "/api/targets/"},
		{http.MethodGet, "/api/targets/{target}"},
		{http.MethodGet, "/api/targets/{target}/delegations/"},
		{http.MethodPost, "/api/targets/{target}/delegations/"},
		{http.MethodDelete, "/api/targets/{target}/delegations/{delegation}"},
	}

	router := bootstrapAPI()

	routes := make([]registeredRoute, 0)
	err := chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		routes = append(routes, registeredRoute{method, route})
		return nil
	})

	assert.NoError(err, "Failed to walk handlers")
	assert.ElementsMatch(expectedRoutes, routes)
}

func TestGetRoot(t *testing.T) {
	assert := assert.New(t)
	router := bootstrapAPI()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal("", rr.Body.String(), "Invalid response text")
}

func TestGetPing(t *testing.T) {
	assert := assert.New(t)
	router := bootstrapAPI()

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal("pong\n", rr.Body.String(), "Invalid response text")
}
