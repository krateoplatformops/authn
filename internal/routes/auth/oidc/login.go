package oidc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/krateoplatformops/authn/internal/helpers/encode"
	kubeconfig "github.com/krateoplatformops/authn/internal/helpers/kube/config"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"github.com/krateoplatformops/authn/internal/helpers/restaction"
	"github.com/krateoplatformops/authn/internal/helpers/userinfo"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/krateoplatformops/authn/internal/shortid"
	"github.com/krateoplatformops/plumbing/kubeutil"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
)

const (
	Path = "/oidc/login"

	authCodeKey = "X-Auth-Code"
)

type LoginOptions struct {
	KubeconfigGenerator kubeconfig.Generator
	JwtDuration         time.Duration
	JwtSingKey          string
}

func Login(ctx context.Context, rc *rest.Config, opts LoginOptions) routes.Route {
	return &loginRoute{
		rc: rc, ctx: ctx,
		gen:         opts.KubeconfigGenerator,
		jwtDuration: opts.JwtDuration,
		jwtSignKey:  opts.JwtSingKey,
	}
}

var _ routes.Route = (*loginRoute)(nil)

type loginRoute struct {
	rc          *rest.Config
	gen         kubeconfig.Generator
	ctx         context.Context
	jwtDuration time.Duration
	jwtSignKey  string
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
			encode.BadRequest(wri, err)
			return
		}

		cfg, err := getConfig(r.rc, name)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to fetch oidc configuration")
			encode.ExpectationFailed(wri, err)
			return
		}

		log.Debug().Str("name", name).Msg("starting oidc login")
		idToken, err := doLogin(req.Header.Get(authCodeKey), cfg)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to complete login")
			encode.InternalError(wri, err)
			return
		}

		log.Debug().Str("name", name).Msg("resolving restaction")
		if cfg.RESTActionRef != nil {
			additionalFieldstoReplace, err := restaction.Resolve(r.ctx, r.rc, cfg.RESTActionRef, idToken.email, idToken.bearerToken)
			if err != nil {
				log.Err(err).Str("name", name).Msg("unable to resolve restaction")
				encode.InternalError(wri, err)
				return
			}
			log.Debug().Str("name", name).Msg("updating oidc idtoken")
			log.Debug().Str("name", name).Msgf("old idToken - name: %s - preferredUsername: %s - email: %s - groups: %s - avatarURL: %s", idToken.name, idToken.preferredUsername, idToken.email, idToken.groups, idToken.avatarURL)
			log.Debug().Str("name", name).Msgf("values to replace: %s", additionalFieldstoReplace)
			idToken, err = updateConfig(idToken, additionalFieldstoReplace)
			if err != nil {
				log.Err(err).Str("name", name).Msg("unable to parse updated idtoken from restaction")
				encode.InternalError(wri, err)
			}
			log.Debug().Str("name", name).Msgf("new idToken - name: %s - preferredUsername: %s - email: %s - groups: %s - avatarURL: %s", idToken.name, idToken.preferredUsername, idToken.email, idToken.groups, idToken.avatarURL)
		}

		log.Debug().Str("name", name).Msg("validating idtoken")
		nfo, err := r.validate(idToken)
		if err != nil {
			log.Err(err).Str("name", name).
				Str("tokenURL", cfg.TokenURL).
				Msg("user info default user error for oidc")
			encode.Forbidden(wri, err)
			return
		}

		log.Debug().Str("name", name).Msg("generating secret from oidc idtoken")
		dat, err := r.gen.Generate(nfo)
		if err != nil {
			log.Err(err).Msg("kubeconfig creation failure")
			encode.InternalError(wri, err)
			return
		}

		encode.Success(wri, dat, &encode.Extras{
			UserInfo:    nfo,
			JwtDuration: r.jwtDuration,
			JwtSingKey:  r.jwtSignKey,
		})
	}
}

func (r *loginRoute) validate(idToken idToken) (userinfo.Info, error) {
	exts := userinfo.Extensions{}
	exts.Add("name", idToken.name)
	exts.Add("avatarUrl", idToken.avatarURL)
	exts.Add("email", idToken.email)

	uid, _ := shortid.Generate()
	nfo := userinfo.NewDefaultUser(kubeutil.MakeDNS1123Compatible(idToken.preferredUsername), uid, idToken.groups, exts)
	return nfo, nil
}
