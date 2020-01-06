package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

type registeredRoute struct {
	method string
	route  string
}

func newTestConfig() *Config {
	return &Config{NotaryConfigFile: "./notary-config.json"}
}

func TestRoutes(t *testing.T) {
	assert := assert.New(t)
	expectedRoutes := []registeredRoute{
		{http.MethodGet, "/"},
		{http.MethodGet, "/ping"},
		{http.MethodGet, "/targets"},
		{http.MethodPost, "/targets"},
		{http.MethodGet, "/targets/{target}"},
		{http.MethodDelete, "/targets/{target}"},
	}

	router := configureAPI(newTestConfig(), zap.NewNop())

	routes := make([]registeredRoute, 0)
	err := chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		routes = append(routes, registeredRoute{method, route})
		return nil
	})

	assert.NoError(err, "Failed to walk handlers")
	assert.Equalf(len(expectedRoutes), len(routes), "Expected %v routes, but got %v", len(expectedRoutes), len(routes))
	assert.Subset(routes, expectedRoutes)
}

func TestGetRoot(t *testing.T) {
	assert := assert.New(t)
	router := configureAPI(newTestConfig(), zap.NewNop())

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal("", rr.Body.String(), "Invalid response text")
}

func TestGetPing(t *testing.T) {
	assert := assert.New(t)
	router := configureAPI(newTestConfig(), zap.NewNop())

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal("pong\n", rr.Body.String(), "Invalid response text")
}
