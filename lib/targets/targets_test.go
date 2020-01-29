package targets

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/theupdateframework/notary/tuf/data"

	"github.com/stretchr/testify/assert"

	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

const (
	NotImplementedResponse = "{\"status\":\"Not implemented.\"}\n"
	NotFoundResponse       = "{\"status\":\"Resource not found.\"}\n"
	InvalidIDResponse      = "{\"status\":\"Invalid request.\",\"error\":\"you must provide at least 7 characters of the path: invalid id\"}\n"
	EmptyResponse          = "[]\n"
)

var (
	n            *notary.Service
	router       *chi.Mux
	ListResponse = []KeyResponse{
		*NewKeyResponse(notary.Key{ID: "c3b49d8c15f339864a21c90a0b7c242e737e6a8a4d1ad73603bfdf0709f01241", GUN: "localhost:5000/dct-notary-admin", Role: "targets"}),
	}
)

func parseSingle(body *bytes.Buffer) (KeyResponse, error) {
	var resp KeyResponse
	return resp, json.Unmarshal(body.Bytes(), &resp)
}

func parseList(body *bytes.Buffer) ([]KeyResponse, error) {
	var resp []KeyResponse
	return resp, json.Unmarshal(body.Bytes(), &resp)
}

func init() {
	os.Setenv("NOTARY_ROOT_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_TARGETS_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_SNAPSHOT_PASSPHRASE", "test1234")

	n = notary.NewService(&notary.Config{
		TrustDir: "../../.notary",
		RemoteServer: notary.RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())

	router = chi.NewRouter()

	router.Use(middleware.Recoverer)
	tr := NewResource(n)

	tr.RegisterRoutes(router)
}

func TestGetTargets(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest(http.MethodGet, "/targets", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	res, err := parseList(rr.Body)
	assert.NoError(err)
	assert.GreaterOrEqual(len(res), len(ListResponse))
	assert.Contains(res, ListResponse[0])
}

func TestGetTarget(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest(http.MethodGet, "/targets/c3b49d8", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	res, err := parseSingle(rr.Body)
	assert.NoError(err)
	assert.NotNil(res)
	assert.Equal(ListResponse[0], res, "Invalid response")
}

func TestGetUnknownTarget(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest(http.MethodGet, "/targets/b635efe", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotFound, rr.Code, "Invalid status code")
	assert.Equal(NotFoundResponse, rr.Body.String(), "Invalid response")
}

func TestGetTargetWithInvalidID(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest(http.MethodGet, "/targets/c3b4", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusBadRequest, rr.Code, "Invalid status code")
	assert.Equal(InvalidIDResponse, rr.Body.String(), "Invalid response")
}

func TestCreateTarget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	reqData := RepositoryRequest{GUN: "localhost:5000/api-create-test/dct-notary-admin"}
	jsonData, _ := json.Marshal(reqData)
	req, err := http.NewRequest(http.MethodPost, "/targets", bytes.NewBuffer(jsonData))
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusCreated, rr.Code, "Invalid status code")

	resp, err := parseSingle(rr.Body)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(reqData.GUN, resp.GUN)
	assert.NotEmpty(resp.ID)
	assert.Equal("targets", resp.Role)

	snapshotKeys, err := n.ListKeys(ctx, notary.AndFilter(notary.SnapshotsFilter, notary.GUNFilter(reqData.GUN)))
	assert.NoError(err)

	err = n.DeleteRepository(ctx, notary.DeleteRepositoryCommand{GUN: data.GUN(reqData.GUN)})
	assert.NoError(err)

	keyIds := make([]string, len(snapshotKeys))
	for i, key := range snapshotKeys {
		keyIds[i] = key.ID
	}
	err = notary.CleanupKeys("../../.notary", append(keyIds, resp.ID)...)
	assert.NoError(err)
}

func TestListTargetDelegates(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest(http.MethodGet, "/targets/c3b49d8/delegates", nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	assert.Equal(EmptyResponse, rr.Body.String(), "Invalid response text")
}
