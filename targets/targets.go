package targets

import (
	"context"
	"errors"
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
		respond(w, r, e.ErrRender(err))
		return
	}
	respondList(w, r, NewTargetListResponse(targets))
}

func (tr *TargetsResource) createTargets(w http.ResponseWriter, r *http.Request) {
	respond(w, r, e.ErrNotImplemented)
}

func (tr *TargetsResource) getTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetTarget(ctx, id)
	if err != nil {
		if errors.Is(err, notary.ErrNotFound) {
			respond(w, r, e.ErrNotFound)
		} else {
			respond(w, r, e.ErrInvalidRequest(err))
		}
		return
	}

	if target == nil {
		respond(w, r, e.ErrNotFound)
	} else {
		respond(w, r, NewTargetResponse(*target))
	}
}

func (tr *TargetsResource) deleteTarget(w http.ResponseWriter, r *http.Request) {
	respond(w, r, e.ErrNotImplemented)
}

func respond(w http.ResponseWriter, r *http.Request, renderer render.Renderer) {
	if err := render.Render(w, r, renderer); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func respondList(w http.ResponseWriter, r *http.Request, renderers []render.Renderer) {
	if err := render.RenderList(w, r, renderers); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}
