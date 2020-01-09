package targets

import (
	"context"
	"net/http"

	"github.com/philips-labs/dct-notary-admin/notary"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	e "github.com/philips-labs/dct-notary-admin/errors"
)

// Resource holds api endpoints for the /targets urls
type Resource struct {
	notary *notary.Service
}

// NewResource create a new instance of Resource
func NewResource(service *notary.Service) *Resource {
	return &Resource{service}
}

// RegisterRoutes registers the API routes
func (tr *Resource) RegisterRoutes(r chi.Router) {
	r.Get("/targets", tr.listTargets)
	r.Post("/targets", tr.createTargets)
	r.Get("/targets/{target}", tr.getTarget)
	r.Get("/targets/{target}/delegates", tr.listDelegates)
}

func (tr *Resource) listTargets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	targets, err := tr.notary.ListTargets(ctx)
	if err != nil {
		respond(w, r, e.ErrRender(err))
		return
	}
	respondList(w, r, NewKeyListResponse(targets))
}

func (tr *Resource) createTargets(w http.ResponseWriter, r *http.Request) {
	respond(w, r, e.ErrNotImplemented)
}

func (tr *Resource) getTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetTarget(ctx, id)
	if err != nil {
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	if target == nil {
		respond(w, r, e.ErrNotFound)
	} else {
		respond(w, r, NewKeyResponse(*target))
	}
}

func (tr *Resource) listDelegates(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetTarget(ctx, id)
	if err != nil {
		respond(w, r, e.ErrRender(err))
		return
	}
	if target == nil {
		respond(w, r, e.ErrNotFound)
		return
	}
	delegates, err := tr.notary.ListDelegates(ctx, target)
	if err != nil {
		respond(w, r, e.ErrRender(err))
		return
	}

	response := make([]notary.Key, 0)
	for _, v := range delegates {
		response = append(response, v...)
	}
	respondList(w, r, NewKeyListResponse(response))
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
