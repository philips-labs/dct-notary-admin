package targets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/stretchr/testify/assert"
)

const (
	NotImplementedResponse = "{\"status\":\"Not implemented.\"}\n"
)

func createRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	RegisterRoutes(r)
	return r
}

func TestGetTargets(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotImplemented, rr.Code, "Invalid status code")
	assert.Equal(NotImplementedResponse, rr.Body.String(), "Invalid response text")
}

func TestGetTarget(t *testing.T) {
	assert := assert.New(t)
	router := createRouter()

	req, err := http.NewRequest(http.MethodGet, "/targets/1234", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotImplemented, rr.Code, "Invalid status code")
	assert.Equal(NotImplementedResponse, rr.Body.String(), "Invalid response text")
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
