package targets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/stretchr/testify/assert"
)

const (
	NotImplementedResponse = "{\"status\":\"Not implemented.\"}\n"
	TargetsResponse        = "[{\"id\":\"b45192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185\",\"gun\":\"docker.io/marcofranssen/openjdk\"},{\"id\":\"b635efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8\",\"gun\":\"docker.io/marcofranssen/whalesay\"},{\"id\":\"d22b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd\",\"gun\":\"docker.io/marcofranssen/nginx\"}]\n"
	WhalesayResponse       = "{\"id\":\"b635efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8\",\"gun\":\"docker.io/marcofranssen/whalesay\"}\n"
)

func createRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	tr := NewTargetsResource(notary.NewService("../notary-config.json"))

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
	assert.Equal(TargetsResponse, rr.Body.String(), "Invalid response")
}

func TestGetTarget(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets/b635efe", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal(WhalesayResponse, rr.Body.String(), "Invalid response")
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

func TestDeleteTarget(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodDelete, "/targets/1234", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotImplemented, rr.Code, "Invalid status code")
	assert.Equal(NotImplementedResponse, rr.Body.String(), "Invalid response text")
}
