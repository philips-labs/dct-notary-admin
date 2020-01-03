package targets

import (
	"io"
	"net/http"
	"os/exec"

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
	cmd := exec.Command("notary", "key", "export")
	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter
	go pipeResponse(w, pipeReader)

	err := cmd.Run()
	if err != nil {
		render.Render(w, r, e.ErrRender(err))
	}
}

func pipeResponse(w http.ResponseWriter, reader io.ReadCloser) {
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			reader.Close()
			break
		}
		data := buf[0:n]
		w.Write(data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		for i := 0; i < n; i++ {
			buf[i] = 0
		}
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
