package targets

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

type RepositoryRequest struct {
	GUN string `json:"gun"`
}

type DelegationRequest struct {
	DelegationPublicKey string `json:"delegationPublicKey"`
	DelegationName      string `json:"delegationName"`
}

func (rr *DelegationRequest) Bind(r *http.Request) error {
	return nil
}

// Bind unmarshals request into structure and validates / cleans input
func (rr *RepositoryRequest) Bind(r *http.Request) error {
	rr.GUN = strings.Trim(rr.GUN, " \t")
	if rr.GUN == "" {
		return errors.New("gun is required")
	}

	return nil
}

// KeyResponse returns a notary.Key structure
type KeyResponse struct {
	*notary.Key
}

// KeyDataResponse returns the key data
type KeyDataResponse struct {
	Data map[string]notary.KeyData `json:"data,omitempty"`
}

// NewKeyResponse creates a KeyResponse from a notary.Key structure
func NewKeyResponse(target notary.Key) *KeyResponse {
	return &KeyResponse{&target}
}

// Render renders a KeyResponse
func (t *KeyResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// preprocessing possible here
	return nil
}

// NewKeyListResponse returns a slice of KeyResponse
func NewKeyListResponse(targets []notary.Key) []render.Renderer {
	list := make([]render.Renderer, len(targets))

	for i, t := range targets {
		list[i] = NewKeyResponse(t)
	}

	return list
}

// Render renders a KeyDataResponse
func (t *KeyDataResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// preprocessing possible here
	return nil
}
