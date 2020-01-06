package targets

import (
	"net/http"

	"github.com/go-chi/render"
)

type Target struct {
	Path string `json:"id"`
	Gun  string `json:"gun"`
}

type TargetResponse struct {
	*Target
}

func NewTargetResponse(target Target) *TargetResponse {
	return &TargetResponse{&target}
}

func (t *Target) Render(w http.ResponseWriter, r *http.Request) error {
	// preprocessing possible here
	return nil
}

func NewTargetListResponse(targets []Target) []render.Renderer {
	list := make([]render.Renderer, len(targets))

	for i, t := range targets {
		list[i] = NewTargetResponse(t)
	}

	return list
}
