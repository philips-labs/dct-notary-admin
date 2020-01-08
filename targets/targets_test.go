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
	NotFoundResponse       = "{\"status\":\"Resource not found.\"}\n"
	InvalidIDResponse      = "{\"status\":\"Invalid request.\",\"error\":\"you must provide at least 7 characters of the path: invalid id\"}\n"
	EmptyResponse          = "[]\n"
)

func createRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	n, _ := notary.NewService("../notary-config.json")
	tr := NewTargetsResource(n)

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
