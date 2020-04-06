package targets

import (
	"errors"
	"net/http"
	"strings"

	"github.com/philips-labs/dct-notary-admin/lib/notary"

	"github.com/go-chi/render"
)

type RepositoryRequest struct {
	GUN string `json:"gun"`
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