package targets

import (
	"net/http"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/render"
)

type TargetResponse struct {
	*notary.Key
}

func NewTargetResponse(target notary.Key) *TargetResponse {
	return &TargetResponse{&target}
}

func (t *TargetResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// preprocessing possible here
	return nil
}

func NewTargetListResponse(targets []notary.Key) []render.Renderer {
	list := make([]render.Renderer, len(targets))

	for i, t := range targets {
		list[i] = NewTargetResponse(t)
	}

	return list
}
