package oidc

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/krateoplatformops/authn/internal/helpers/decode"
	"github.com/krateoplatformops/authn/internal/helpers/encode"
	kubeconfig "github.com/krateoplatformops/authn/internal/helpers/kube/config"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"github.com/krateoplatformops/authn/internal/helpers/userinfo"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/krateoplatformops/authn/internal/shortid"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
)

const (
	Path = "/oidc/login"
)

func Login(rc *rest.Config, gen kubeconfig.Generator) routes.Route {
	return &loginRoute{
		rc: rc, gen: gen,
	}
}

var _ routes.Route = (*loginRoute)(nil)

type loginRoute struct {
	rc  *rest.Config
	gen kubeconfig.Generator
}

func (r *loginRoute) Name() string {
	return "oidc.login"
}

func (r *loginRoute) Pattern() string {
	return Path
}

func (r *loginRoute) Method() string {
	return http.MethodGet
}

func (r *loginRoute) Handler() http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		log := zerolog.Ctx(req.Context()).With().
			Str("namespace", os.Getenv(util.NamespaceEnvVar)).
			Logger()

		name := req.URL.Query().Get("name")
		if len(name) == 0 {
			err := fmt.Errorf("OIDCConfig 'name' must be specified")
			log.Err(err).Msgf("empty 'name' parameter in query string")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}
		var lo loginInfo
		err := decode.JSONBody(wri, req, &lo)
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			encode.Error(wri, http.StatusBadRequest, err)
			return
		}

		cfg, err := getConfig(r.rc, name)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to fetch oidc configuration")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}

		idToken, err := doLogin(lo.Username, lo.Password, cfg)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to decode id token from jwt")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}

		nfo, err := r.validate(idToken)
		if err != nil {
			log.Err(err).Str("name", name).
				Str("tokenURL", cfg.TokenURL).
				Str("user", lo.Username).
				Msg("user info default user error for oidc")
			code := http.StatusForbidden
			encode.Error(wri, code, err)
			return
		}

		dat, err := r.gen.Generate(nfo)
		if err != nil {
			log.Err(err).Msg("kubeconfig creation failure")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}

		encode.Success(wri, nfo, dat)

	}
}

func (r *loginRoute) validate(idToken idToken) (userinfo.Info, error) {
	exts := userinfo.Extensions{}
	exts.Add("name", idToken.name)
	exts.Add("avatarUrl", idToken.avatarURL)
	exts.Add("email", idToken.email)

	uid, _ := shortid.Generate()
	nfo := userinfo.NewDefaultUser(strings.Replace(idToken.preferredUsername, "@", "-", 1), uid, idToken.groups, exts)
	return nfo, nil
}

type loginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
