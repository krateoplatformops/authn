package oauth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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

const (
	Path        = "/oauth/login"
	authCodeKey = "X-Auth-Code"
)

var _ routes.Route = (*loginRoute)(nil)

type loginRoute struct {
	rc          *rest.Config
	gen         kubeconfig.Generator
	ctx         context.Context
	jwtDuration time.Duration
	jwtSignKey  string
}

func (r *loginRoute) Name() string {
	return "oauth.login"
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
			err := fmt.Errorf("OAuthConfig 'name' must be specified")
			log.Err(err).Msgf("empty 'name' parameter in query string")
			encode.BadRequest(wri, err)
			return
		}

		code := req.Header.Get(authCodeKey)
		log.Debug().Str("name", name).Str("code", code).Msg("received authorization code")
		if len(code) == 0 {
			log.Error().Msgf("empty oauth code")
			encode.BadRequest(wri, fmt.Errorf("empty oauth code"))
			return
		}

		oc, restactionRef, err := getConfig(r.rc, name)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to fetch oauth2 configuration")
			encode.ExpectationFailed(wri, err)
			return
		}

		// use code to get token.
		tok, err := oc.Exchange(context.Background(), code)
		if err != nil {
			log.Err(err).Msg("unable to auth code for token")
			encode.ExpectationFailed(wri, err)
			return
		}

		ctx := context.WithValue(req.Context(), restaction.RestActionContextKey("username"), r.ctx.Value(restaction.RestActionContextKey("username")).(string))
		ctx = context.WithValue(ctx, restaction.RestActionContextKey("snowplowURL"), r.ctx.Value(restaction.RestActionContextKey("snowplowURL")).(string))
		userinfo := userInfo{}
		log.Debug().Str("name", name).Msg("resolving restaction")
		if restactionRef != nil {
			if tok.TokenType != "bearer" {
				err := fmt.Errorf("oauth2 token is not type bearer: %s", tok.TokenType)
				log.Err(err).Str("name", name).Msgf("error while resolving restaction")
				encode.InternalError(wri, err)
				return
			}
			additionalFieldstoReplace, err := restaction.Resolve(ctx, r.rc, restactionRef, uuid.New().String(), tok.AccessToken)
			if err != nil {
				log.Err(err).Str("name", name).Msg("unable to resolve restaction")
				encode.InternalError(wri, err)
				return
			}
			log.Debug().Str("name", name).Msgf("values to replace: %s", additionalFieldstoReplace)
			userinfo, err = updateConfig(userinfo, additionalFieldstoReplace)
			if err != nil {
				log.Err(err).Str("name", name).Msg("unable to parse updated config from restaction")
				encode.InternalError(wri, err)
			}
			log.Debug().Str("name", name).Msgf("new idToken - name: %s - preferredUsername: %s - email: %s - groups: %s - avatarURL: %s", userinfo.name, userinfo.preferredUsername, userinfo.email, userinfo.groups, userinfo.avatarURL)
		}

		user, err := r.validate(userinfo)
		if err != nil {
			log.Err(err).Msg("unable to fetch user info from provider")
			encode.ExpectationFailed(wri, err)
			return
		}
		log.Info().
			Str("user", user.GetUserName()).
			Strs("groups", user.GetGroups()).
			Msg("user info successfully fetched")

		dat, err := r.gen.Generate(user)
		if err != nil {
			log.Err(err).Msg("kubeconfig creation failure")
			encode.InternalError(wri, err)
			return
		}

		encode.Success(wri, dat, &encode.Extras{
			UserInfo:    user,
			JwtDuration: r.jwtDuration,
			JwtSingKey:  r.jwtSignKey,
		})
	}
}

func (r *loginRoute) validate(user userInfo) (userinfo.Info, error) {
	exts := userinfo.Extensions{}
	exts.Add("name", user.name)
	exts.Add("email", user.email)
	exts.Add("avatarUrl", user.avatarURL)

	uid, _ := shortid.Generate()
	info := userinfo.NewDefaultUser(kubeutil.MakeDNS1123Compatible(user.preferredUsername), uid, user.groups, exts)

	return info, nil
}
