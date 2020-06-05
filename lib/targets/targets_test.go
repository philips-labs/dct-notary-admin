package targets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/theupdateframework/notary/tuf/data"

	"github.com/stretchr/testify/assert"

	m "github.com/philips-labs/dct-notary-admin/lib/middleware"
	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

const (
	chars             = "abcdefghijklmnopqrstuvwxyz_"
	NotFoundResponse  = "{\"status\":\"Resource not found.\"}\n"
	InvalidIDResponse = "{\"status\":\"Invalid request.\",\"error\":\"you must provide at least 7 characters of the path: invalid id\"}\n"
	pubKey            = `-----BEGIN PUBLIC KEY-----
role: marcofranssen

MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEmI6bhcF0aqKobYIgBD/wHg/vhjW2
E+C9PEdgfom/x+XxcrFLxvPz1jl7sH8yj315Tr3C5dcE9GhDDlNyJcNC/g==
-----END PUBLIC KEY-----
`
)

var (
	n            *notary.Service
	router       *chi.Mux
	ListResponse = []KeyResponse{
		*NewKeyResponse(notary.Key{ID: "4ea1fec36392486d4bd99795ffc70f3ffa4a76185b39c8c2ab1d9cf5054dbbc9", GUN: "localhost:5000/dct-notary-admin", Role: "targets"}),
	}
)

func parseMetadataResponse(body *bytes.Buffer) (MetadataResponse, error) {
	var resp MetadataResponse
	return resp, json.Unmarshal(body.Bytes(), &resp)
}

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

	nopLogger := zap.NewNop()

	n = notary.NewService(&notary.Config{
		TrustDir: "../../.notary",
		RemoteServer: notary.RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, notary.GetPassphraseRetriever(), nopLogger)

	router = chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(m.ZapLogger(nopLogger))
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

	req, err := http.NewRequest(http.MethodGet, "/targets/4ea1fec", nil)
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

func TestFetchMetadata(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	id, err := createTestTarget(ctx, gun)
	if !assert.NoError(err) {
		return
	}
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()

	jsonData, _ := json.Marshal(RepositoryRequest{GUN: gun.String()})
	req, err := http.NewRequest(http.MethodPost, "/targets/fetchmeta", bytes.NewBuffer(jsonData))
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	resp, err := parseMetadataResponse(rr.Body)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.NotNil(resp.Data)
	assert.NotNil(resp.Data.Root)
	assert.NotNil(resp.Data.Root.Signed.Keys[id])
	assert.NotNil(resp.Data.Targets)
	assert.NotNil(resp.Data.Snapshot)
	assert.NotNil(resp.Data.Timestamp)
}

func TestCreateTarget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	jsonData, _ := json.Marshal(RepositoryRequest{GUN: gun.String()})
	req, err := http.NewRequest(http.MethodPost, "/targets", bytes.NewBuffer(jsonData))
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusCreated, rr.Code, "Invalid status code")
	resp, err := parseSingle(rr.Body)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(gun.String(), resp.GUN)
	assert.NotEmpty(resp.ID)
	assert.Equal("targets", resp.Role)

	err = cleanupTarget(ctx, gun, resp.ID)
	assert.NoError(err)
}

func TestAddDelegation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	id, err := createTestTarget(ctx, gun)
	if !assert.NoError(err) {
		return
	}
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()

	data := DelegationRequest{
		DelegationName:      "marcofranssen",
		DelegationPublicKey: pubKey,
	}
	body, _ := json.Marshal(data)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/targets/%s/delegations", id), bytes.NewReader(body))
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusCreated, rr.Code, "Invalid status code")

	resp, err := parseSingle(rr.Body)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(gun.String(), resp.GUN)
	assert.NotEmpty(resp.ID)
	assert.Equal(data.DelegationName, resp.Role)
}

func TestListTargetDelegates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	id, err := createTestTarget(ctx, gun)
	assert.NoError(err)
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()
	delID, delName, err := addDelegation(ctx, gun)
	assert.NoError(err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/targets/%s/delegations", id), nil)
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")

	resp, err := parseList(rr.Body)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Len(resp, 1, "Expected response to have one delegation")
	assert.Empty(resp[0].GUN)
	assert.Equal(delID, resp[0].ID)
	assert.Equal(strings.TrimPrefix(delName.String(), "targets/"), resp[0].Role)
}

func TestRemoveDelegation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	id, err := createTestTarget(ctx, gun)
	assert.NoError(err)
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()
	delID, delName, err := addDelegation(ctx, gun)
	assert.NoError(err)

	dr := DelegationRequest{DelegationName: string(delName[8:])}
	body, _ := json.Marshal(dr)
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/targets/%s/delegations/%s", id, delID), bytes.NewReader(body))
	assert.NoError(err, "Failed to create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "Invalid status code")
	resp, err := parseSingle(rr.Body)

	if assert.NoError(err) && assert.NotNil(resp) {
		assert.Equal(delName[8:], data.RoleName(resp.Role))
		assert.Equal(delID, resp.ID)
		assert.Equal(gun, data.GUN(resp.GUN))
	}
}

func randomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[seededRand.Intn(len(chars))]
	}
	return string(b)
}

func randomGUN() data.GUN {
	return data.GUN(fmt.Sprintf("localhost:5000/random-test-guns/dctna-%s", randomString(10)))
}

func createTestTarget(ctx context.Context, gun data.GUN) (string, error) {
	err := n.CreateRepository(ctx, notary.CreateRepoCommand{
		TargetCommand: notary.TargetCommand{GUN: gun},
		AutoPublish:   true,
	})
	if err != nil {
		return "", err
	}
	targetKey, err := n.GetTargetByGUN(ctx, gun)
	if err != nil {
		return "", err
	}
	return targetKey.ID, nil
}

func addDelegation(ctx context.Context, gun data.GUN) (string, data.RoleName, error) {
	pubKey, pubKeyID, err := readPublicKey([]byte(pubKey))
	if err != nil {
		return "", data.RoleName(""), nil
	}
	role := notary.DelegationPath(randomString(8))
	err = n.AddDelegation(ctx, notary.AddDelegationCommand{
		AutoPublish:    true,
		Role:           role,
		DelegationKeys: []data.PublicKey{pubKey},
		Paths:          []string{""},
		TargetCommand:  notary.TargetCommand{GUN: gun},
	})
	if err != nil {
		return "", data.RoleName(""), nil
	}
	return pubKeyID, role, nil
}

func cleanupTarget(ctx context.Context, gun data.GUN, keyID string) error {
	snapshotKeys, err := n.ListKeys(ctx, notary.AndFilter(notary.SnapshotsFilter, notary.GUNFilter(gun.String())))
	if err != nil {
		return err
	}
	err = n.DeleteRepository(ctx, notary.DeleteRepositoryCommand{TargetCommand: notary.TargetCommand{GUN: gun}})
	if err != nil {
		return err
	}

	keyIds := make([]string, len(snapshotKeys))
	for i, key := range snapshotKeys {
		keyIds[i] = key.ID
	}
	return notary.CleanupKeys("../../.notary", append(keyIds, keyID)...)
}
