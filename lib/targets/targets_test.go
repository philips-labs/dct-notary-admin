package targets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philips-labs/dct-notary-admin/lib/notary"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/stretchr/testify/assert"
)

const (
	NotImplementedResponse = "{\"status\":\"Not implemented.\"}\n"
	NotFoundResponse       = "{\"status\":\"Resource not found.\"}\n"
	InvalidIDResponse      = "{\"status\":\"Invalid request.\",\"error\":\"you must provide at least 7 characters of the path: invalid id\"}\n"
	EmptyResponse          = "[]\n"
)

func createRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	n := notary.NewService(&notary.NotaryConfig{
		TrustDir: "~/.dctna/trust",
		RemoteServer: notary.RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	})
	tr := NewResource(n)

	tr.RegisterRoutes(r)
	return r
}

func TestGetTargets(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal(EmptyResponse, rr.Body.String(), "Invalid response")
}

func TestGetTarget(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets/b635efe", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotFound, rr.Code, "Invalid status code")
	assert.Equal(NotFoundResponse, rr.Body.String(), "Invalid response")
}

func TestGetTargetWithInvalidID(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets/b635", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusBadRequest, rr.Code, "Invalid status code")
	assert.Equal(InvalidIDResponse, rr.Body.String(), "Invalid response")
}

func TestCreateTarget(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodPost, "/targets", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotImplemented, rr.Code, "Invalid status code")
	assert.Equal(NotImplementedResponse, rr.Body.String(), "Invalid response text")
}

func TestListTargetDelegates(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets/d22b2a4/delegates", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotFound, rr.Code, "Invalid status code")
	assert.Equal(NotFoundResponse, rr.Body.String(), "Invalid response text")
}
