package routes

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type Route interface {
	Name() string
	Pattern() string
	Method() string
	Handler() http.HandlerFunc
}

func Serve(all []Route, log zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(wri http.ResponseWriter, req *http.Request) {
		var allow []string
		for _, route := range all {
			if req.URL.Path != route.Pattern() {
				continue
			}

			if req.Method != route.Method() {
				allow = append(allow, route.Method())
				continue
			}

			l := log.With().Logger()
			req = req.WithContext(l.WithContext(req.Context()))
			route.Handler()(wri, req)
			return
		}

		if len(allow) > 0 {
			wri.Header().Set("Allow", strings.Join(allow, ", "))
			http.Error(wri, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.NotFound(wri, req)
	})
}
