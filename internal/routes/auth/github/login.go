package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/authn/internal/helpers/encode"
	kubeconfig "github.com/krateoplatformops/authn/internal/helpers/kube/config"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"github.com/krateoplatformops/authn/internal/helpers/userinfo"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/krateoplatformops/authn/internal/shortid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"k8s.io/client-go/rest"
)

func Login(rc *rest.Config, gen kubeconfig.Generator) routes.Route {
	return &loginRoute{
		rc:  rc,
		gen: gen,
	}
}

const (
	Path = "/github/login"

	authCodeKey = "X-Auth-Code"
)

var _ routes.Route = (*loginRoute)(nil)

type loginRoute struct {
	rc  *rest.Config
	gen kubeconfig.Generator
}

func (r *loginRoute) Name() string {
	return "github.login"
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
			err := fmt.Errorf("GithubConfig 'name' must be specified")
			log.Err(err).Msgf("empty 'name' parameter in query string")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}

		code := req.Header.Get(authCodeKey)
		log.Debug().Str("name", name).Str("code", code).Msg("received authorization code")
		if len(code) == 0 {
			log.Error().Msgf("empty oauth code")
			encode.Error(wri, http.StatusBadRequest, fmt.Errorf("empty oauth code"))
			return
		}

		oc, org, apiUrl, err := getConfig(r.rc, name)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to fetch oauth2 configuration")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}

		// use code to get token.
		tok, err := oc.Exchange(context.Background(), code)
		if err != nil {
			log.Err(err).Msg("unable to auth code for token")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}

		user, err := r.validate(tok, org, apiUrl)
		if err != nil {
			log.Err(err).Msg("unable to fetch user info from github")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}
		log.Info().
			Str("user", user.GetUserName()).
			Strs("groups", user.GetGroups()).
			Msg("user info successfully fetched")

		dat, err := r.gen.Generate(user)
		if err != nil {
			log.Err(err).Msg("kubeconfig creation failure")
			encode.Error(wri, http.StatusInternalServerError, err)
			return
		}

		encode.Success(wri, user, dat)
	}
}

func (r *loginRoute) validate(tok *oauth2.Token, org, apiUrl string) (userinfo.Info, error) {
	cli := newGithubApiClient(tok, org, apiUrl)

	user, err := cli.getUserInfo()
	if err != nil {
		return nil, err
	}

	teams, err := cli.listTeams()
	if err != nil {
		return nil, err
	}

	groups := []string{}
	for _, x := range teams {
		ok, err := cli.isUserMemberOfTeam(user.Login, x)
		if err != nil {
			return nil, err
		}
		if ok {
			groups = append(groups, x.Slug)
		}
	}

	exts := userinfo.Extensions{}
	exts.Add("name", user.Name)
	exts.Add("email", user.Email)
	exts.Add("avatarUrl", user.AvatarURL)
	exts.Add("url", user.URL)

	uid, _ := shortid.Generate()
	info := userinfo.NewDefaultUser(user.Login, uid, groups, exts)

	return info, nil
}
