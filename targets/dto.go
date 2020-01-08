package targets

import (
	"net/http"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/render"
)

type KeyResponse struct {
	*notary.Key
}

func NewKeyResponse(target notary.Key) *KeyResponse {
	return &KeyResponse{&target}
}

func (t *KeyResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// preprocessing possible here
	return nil
}

func NewKeyListResponse(targets []notary.Key) []render.Renderer {
	list := make([]render.Renderer, len(targets))

	for i, t := range targets {
		list[i] = NewKeyResponse(t)
	}

	return list
}
