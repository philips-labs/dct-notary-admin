package targets

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	e "github.com/philips-labs/dct-notary-admin/errors"
)

// RegisterRoutes registers the API routes
func RegisterRoutes(r chi.Router) {
	r.Get("/targets", listTargets)
	r.Post("/targets", createTargets)
	r.Get("/targets/{target}", getTarget)
	r.Delete("/targets/{target}", deleteTarget)
}

func listTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := listNotaryTargets()
	if err != nil {
		render.Render(w, r, e.ErrRender(err))
		return
	}
	if err := render.RenderList(w, r, NewTargetListResponse(targets)); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func createTargets(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, e.ErrNotImplemented); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func getTarget(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, e.ErrNotImplemented); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func deleteTarget(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, e.ErrNotImplemented); err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}
