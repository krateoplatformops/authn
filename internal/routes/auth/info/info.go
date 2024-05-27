package info

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krateoplatformops/authn/internal/helpers/encode"
	"github.com/krateoplatformops/authn/internal/helpers/kube/config/storage"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
)

const (
	Path = "/info"
)

func Info(rc *rest.Config) routes.Route {
	return &infoRoute{
		store: storage.Default(rc),
	}
}

var _ routes.Route = (*infoRoute)(nil)

type infoRoute struct {
	store storage.AuthInfoStorage
}

func (r *infoRoute) Name() string {
	return "info"
}

func (r *infoRoute) Pattern() string {
	return Path
}

func (r *infoRoute) Method() string {
	return http.MethodGet
}

func (r *infoRoute) Handler() http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		log := zerolog.Ctx(req.Context()).With().Logger()

		qs := req.URL.Query()
		name := qs.Get("name")
		if len(name) == 0 {
			err := fmt.Errorf("missing 'name' param")
			log.Err(err).Msg("required query string param non found")
			encode.Error(wri, http.StatusBadRequest, err)
			return
		}

		nfo, err := r.store.Get(name)
		if err != nil {
			log.Err(err).
				Str("name", name).
				Msg("unable to resolve authinfo secret")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}

		dat, err := json.Marshal(nfo)
		if err != nil {
			log.Err(err).
				Str("name", name).
				Msg("unable to encode authinfo")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}

		encode.Success(wri, nil, dat)
	}
}
