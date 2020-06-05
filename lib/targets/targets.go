package targets

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"

	e "github.com/philips-labs/dct-notary-admin/lib/errors"
	m "github.com/philips-labs/dct-notary-admin/lib/middleware"
	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

const (
	ErrMsgFailedParseBody          = "failed to parse request body"
	ErrMsgFailedFetchMetadata      = "failed fetching TUF metadata"
	ErrMsgFailedListTargetKeys     = "failed to list target keys"
	ErrMsgFailedListDelegationKeys = "faild to list delegation keys"
	ErrMsgFailedGetTargetKey       = "failed getting target key"
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
		rr.Post("/fetchmeta", tr.fetchMetadata)
		rr.Get("/{target}", tr.getTarget)
		rr.Route("/{target}/delegations", func(rrr chi.Router) {
			rrr.Get("/", tr.listDelegates)
			rrr.Post("/", tr.addDelegation)
			rrr.Delete("/{delegation}", tr.removeDelegation)
		})
	})
}

func (tr *Resource) listTargets(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	targets, err := tr.notary.ListTargets(ctx)
	if err != nil {
		log.Error(ErrMsgFailedListTargetKeys, zap.Error(err))
		respond(w, r, e.ErrRender(err))
		return
	}
	respondList(w, r, NewKeyListResponse(targets))
}

func (tr *Resource) createTarget(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	body := &RepositoryRequest{}
	if err := render.Bind(r, body); err != nil {
		log.Error(ErrMsgFailedParseBody, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	err := tr.notary.CreateRepository(ctx, notary.CreateRepoCommand{
		TargetCommand: notary.TargetCommand{GUN: data.GUN(body.GUN)},
		AutoPublish:   true,
	})
	if err != nil {
		log.Error("failed creating target", zap.Error(err))
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	newKey, err := tr.notary.GetTargetByGUN(ctx, data.GUN(body.GUN))
	if err != nil || newKey == nil {
		log.Error(ErrMsgFailedGetTargetKey, zap.Error(err))
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	respond(w, r, NewKeyResponse(*newKey))
}

func (tr *Resource) fetchMetadata(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	body := &RepositoryRequest{}
	if err := render.Bind(r, body); err != nil {
		log.Error(ErrMsgFailedParseBody, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	metadata, err := tr.notary.FetchMetadata(ctx, data.GUN(body.GUN))
	if err != nil {
		log.Error(ErrMsgFailedFetchMetadata, zap.Error(err))
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	respond(w, r, &MetadataResponse{Data: metadata})
}

func (tr *Resource) getTarget(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetKeyByID(ctx, id)
	if err != nil {
		log.Error(ErrMsgFailedGetTargetKey, zap.Error(err))
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
	log := m.GetZapLogger(r)
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetKeyByID(ctx, id)
	if err != nil {
		log.Error(ErrMsgFailedGetTargetKey, zap.Error(err))
		respond(w, r, e.ErrRender(err))
		return
	}
	if target == nil {
		respond(w, r, e.ErrNotFound)
		return
	}
	delegates, err := tr.notary.ListDelegates(ctx, target)
	if err != nil {
		log.Error(ErrMsgFailedListDelegationKeys, zap.Error(err))
		respond(w, r, e.ErrRender(err))
		return
	}

	response := make([]notary.Key, 0)
	for _, v := range delegates {
		response = append(response, v...)
	}
	respondList(w, r, NewKeyListResponse(response))
}

func (tr *Resource) addDelegation(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)
	id := chi.URLParam(r, "target")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetKeyByID(ctx, id)
	if err != nil {
		log.Error(ErrMsgFailedGetTargetKey, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}
	body := &DelegationRequest{}
	if err := render.Bind(r, body); err != nil {
		log.Error(ErrMsgFailedParseBody, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	pubKey, pubKeyID, err := readPublicKey([]byte(body.DelegationPublicKey))
	if err != nil {
		log.Error("failed to read public key", zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	role := notary.DelegationPath(body.DelegationName)
	err = tr.notary.AddDelegation(ctx, notary.AddDelegationCommand{
		AutoPublish:    true,
		Role:           role,
		DelegationKeys: []data.PublicKey{pubKey},
		Paths:          []string{""},
		TargetCommand:  notary.TargetCommand{GUN: data.GUN(target.GUN)},
	})
	if err != nil {
		log.Error("failed to add delegation", zap.Error(err))
		respond(w, r, e.ErrInternalServer(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	respond(w, r, NewKeyResponse(notary.Key{ID: pubKeyID, GUN: target.GUN, Role: body.DelegationName}))
}

func (tr *Resource) removeDelegation(w http.ResponseWriter, r *http.Request) {
	log := m.GetZapLogger(r)

	targetID := chi.URLParam(r, "target")
	delegationID := chi.URLParam(r, "delegation")

	if targetID == "" || delegationID == "" {
		err := errors.New("no target or delegation provided")
		log.Error("", zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	target, err := tr.notary.GetKeyByID(ctx, targetID)
	if err != nil {
		log.Error(ErrMsgFailedGetTargetKey, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}
	if target == nil {
		respond(w, r, e.ErrNotFound)
		return
	}

	body := &DelegationRequest{}
	if err := render.Bind(r, body); err != nil {
		log.Error(ErrMsgFailedParseBody, zap.Error(err))
		respond(w, r, e.ErrInvalidRequest(err))
		return
	}

	role := notary.DelegationPath(body.DelegationName)
	delegation, err := tr.notary.GetDelegation(ctx, target, role, delegationID)
	if err != nil {
		respond(w, r, e.ErrInternalServer(err))
		return
	}
	if delegation == nil {
		respond(w, r, e.ErrNotFound)
		return
	}

	err = tr.notary.RemoveDelegation(ctx, notary.RemoveDelegationCommand{
		TargetCommand: notary.TargetCommand{GUN: data.GUN(target.GUN)},
		AutoPublish:   true,
		KeyID:         delegation.ID,
		Role:          notary.DelegationPath(delegation.Role),
	})
	if err != nil {
		log.Error("failed to remove delegation", zap.Error(err))
		respond(w, r, e.ErrInternalServer(err))
		return
	}
	respond(w, r, NewKeyResponse(notary.Key{ID: delegation.ID, GUN: target.GUN, Role: delegation.Role}))
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

func readPublicKey(pubKeyBytes []byte) (data.PublicKey, string, error) {
	pubKey, err := utils.ParsePEMPublicKey(pubKeyBytes)
	if err != nil {
		return nil, "", fmt.Errorf("can't parse public key: %w", err)
	}

	pubKeyID, err := utils.CanonicalKeyID(pubKey)
	if err != nil {
		return pubKey, "", fmt.Errorf("can't determine public Key ID: %w", err)
	}

	return pubKey, pubKeyID, nil
}
