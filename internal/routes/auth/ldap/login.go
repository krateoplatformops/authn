package ldap

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/authn/internal/helpers/decode"
	"github.com/krateoplatformops/authn/internal/helpers/encode"
	kubeconfig "github.com/krateoplatformops/authn/internal/helpers/kube/config"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
)

func Login(rc *rest.Config, gen kubeconfig.Generator) routes.Route {
	return &loginRoute{
		rc:  rc,
		gen: gen,
	}
}

const (
	Path = "/ldap/login"
)

var (
	_                 routes.Route = (*loginRoute)(nil)
	errNotFound                    = errors.New("no users found")
	errTooManyEntries              = errors.New("too many entries found")
)

type loginRoute struct {
	rc  *rest.Config
	gen kubeconfig.Generator
}

func (r *loginRoute) Name() string {
	return "ldap.login"
}

func (r *loginRoute) Pattern() string {
	return Path
}

func (r *loginRoute) Method() string {
	return http.MethodPost
}

//	curl -X POST "http://localhost:8080/ldap/login?name=forumsys" \
//	  -H 'Content-Type: application/json' \
//	  -d '{"username":"euler","password":"my_password"}'
func (r *loginRoute) Handler() http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		log := zerolog.Ctx(req.Context()).With().
			Str("namespace", os.Getenv(util.NamespaceEnvVar)).
			Logger()

		name := req.URL.Query().Get("name")
		if len(name) == 0 {
			err := fmt.Errorf("LDAPConfig 'name' must be specified")
			log.Err(err).Msgf("empty 'name' parameter in query string")
			encode.Error(wri, http.StatusBadRequest, err)
			return
		}

		var lo loginInfo
		err := decode.JSONBody(wri, req, &lo)
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			encode.Error(wri, http.StatusBadRequest, err)
			return
		}

		cfg, err := getConfig(r.rc, name, lo.Username)
		if err != nil {
			log.Err(err).Str("name", name).Msg("unable to fetch ldap configuration")
			encode.Error(wri, http.StatusExpectationFailed, err)
			return
		}

		nfo, err := doLogin(lo.Username, lo.Password, cfg)
		if err != nil {
			log.Err(err).Str("name", name).
				Str("dialURL", cfg.dialURL).
				Str("user", lo.Username).
				Msg("login with ldap server failed")
			code := http.StatusForbidden
			if errors.Is(err, errNotFound) {
				code = http.StatusNotFound
			} else if errors.Is(err, errTooManyEntries) {
				code = http.StatusMultipleChoices
			}
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

type loginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
