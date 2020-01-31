package targets

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/theupdateframework/notary/tuf/data"

	e "github.com/philips-labs/dct-notary-admin/lib/errors"
	"github.com/philips-labs/dct-notary-admin/lib/notary"
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
	r.Route("/targets", func(rr chi.Router) {
		rr.Use(render.SetContentType(render.ContentTypeJSON))
		rr.Get("/", tr.listTargets)
		rr.Post("/", tr.createTarget)
		rr.Get("/{target}", tr.getTarget)
		rr.Get("/{target}/delegates", tr.listDelegates)
	})
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

func (tr *Resource) createTarget(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	body := &RepositoryRequest{}
	if err := render.Bind(r, body); err != nil {
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	err := tr.notary.CreateRepository(ctx, notary.CreateRepoCommand{
		TargetCommand: notary.TargetCommand{GUN: data.GUN(body.GUN)},
		AutoPublish:   true,
	})
	if err != nil {
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	newKey, err := tr.notary.ListKeys(ctx, notary.AndFilter(notary.TargetsFilter, notary.GUNFilter(body.GUN)))
	if err != nil {
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	respond(w, r, NewKeyResponse(newKey[0]))
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
