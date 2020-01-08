package targets

import (
	"net/http"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/render"
)

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
