package targets

import (
	"context"
	"net/http"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	e "github.com/philips-labs/dct-notary-admin/errors"
)

type TargetsResource struct {
	notary *notary.Service
}

func NewTargetsResource(service *notary.Service) *TargetsResource {
	return &TargetsResource{service}
}

// RegisterRoutes registers the API routes
func (tr *TargetsResource) RegisterRoutes(r chi.Router) {
	r.Get("/targets", tr.listTargets)
	r.Post("/targets", tr.createTargets)
	r.Get("/targets/{target}", tr.getTarget)
	r.Delete("/targets/{target}", tr.deleteTarget)
}

func (tr *TargetsResource) listTargets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	targets, err := tr.notary.ListTargets(ctx)
	if err != nil {
		render.Render(w, r, e.ErrRender(err))
		return
	}
	if err := render.RenderList(w, r, NewTargetListResponse(targets)); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func (tr *TargetsResource) createTargets(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, e.ErrNotImplemented); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func (tr *TargetsResource) getTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetTarget(ctx, id)
	if err != nil {
		if err := render.Render(w, r, e.ErrInvalidRequest(err)); err != nil {
			render.Render(w, r, e.ErrRender(err))
		}
		return
	}

	if target == nil {
		if err := render.Render(w, r, e.ErrNotFound); err != nil {
			render.Render(w, r, e.ErrRender(err))
		}
	} else {
		if err := render.Render(w, r, NewTargetResponse(*target)); err != nil {
			render.Render(w, r, e.ErrRender(err))
		}
	}
}

func (tr *TargetsResource) deleteTarget(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, e.ErrNotImplemented); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}
