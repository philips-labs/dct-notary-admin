package lib

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"go.uber.org/zap"

	m "github.com/philips-labs/dct-notary-admin/lib/middleware"
	"github.com/philips-labs/dct-notary-admin/lib/notary"
	"github.com/philips-labs/dct-notary-admin/lib/targets"
)

func configureAPI(n *notary.Service, l *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(m.ZapLogger(l))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong\n"))
	})
	r.Route("/api", func(rr chi.Router) {
		rr.Use(render.SetContentType(render.ContentTypeJSON))

		tr := targets.NewResource(n)
		tr.RegisterRoutes(rr)
	})

	logRoutes(r, l)

	return r
}

func logRoutes(r *chi.Mux, logger *zap.Logger) {
	if err := chi.Walk(r, zapPrintRoute(logger)); err != nil {
		logger.Error("Failed to walk routes:", zap.Error(err))
	}
}

func zapPrintRoute(logger *zap.Logger) chi.WalkFunc {
	return func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		logger.Debug("Registering route", zap.String("method", method), zap.String("route", route))
		return nil
	}
}
