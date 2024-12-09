package encode

import (
	"encoding/json"
	"net/http"

	"github.com/krateoplatformops/authn/internal/helpers/userinfo"
)

func Success(w http.ResponseWriter, info userinfo.Info, dat []byte) error {
	out := response{
		Data: dat,
	}

	if info != nil {
		out.User = &user{
			Username:    info.GetUserName(),
			DisplayName: info.GetExtensions().Get("name"),
			AvatarURL:   info.GetExtensions().Get("avatarUrl"),
		}
		out.Groups = info.GetGroups()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(&out)
}

type user struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatarURL"`
}

type response struct {
	User   *user           `json:"user,omitempty"`
	Groups []string        `json:"groups,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}
