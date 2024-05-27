package encode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/krateoplatformops/authn/internal/helpers/userinfo"
)

func Attach(w http.ResponseWriter, username string, dat []byte) error {
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s.json", username))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(dat)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(dat)
	return err
}

func Success(w http.ResponseWriter, info userinfo.Info, dat []byte) error {
	out := response{
		Code: http.StatusOK,
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

func Error(w http.ResponseWriter, status int, err error) error {
	out := response{
		Code:  status,
		Error: err.Error(),
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
	Code   int             `json:"code"`
	Error  string          `json:"error,omitempty"`
	User   *user           `json:"user,omitempty"`
	Groups []string        `json:"groups,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

type expireDuration struct {
	time.Duration
}

func (d *expireDuration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

func (d expireDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

type jwtInfo struct {
	AccessToken string         `json:"accessToken"`
	Expires     expireDuration `json:"expiresIn"`
}
