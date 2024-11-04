package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/krateoplatformops/authn/internal/helpers/kube/resolvers"
	"github.com/krateoplatformops/authn/internal/helpers/kube/secrets"
	"k8s.io/client-go/rest"
)

type oidcConfig struct {
	DiscoveryURL     string
	AuthorizeURL     string
	TokenURL         string
	UserInfoURL      string
	RedirectURI      string
	ClientID         string
	ClientSecret     string
	AdditionalScopes string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

type idToken struct {
	name              string
	email             string
	preferredUsername string
	groups            []string
	avatarURL         string
}

func getConfig(rc *rest.Config, name string) (*oidcConfig, error) {
	cfg, err := resolvers.OIDCConfigGet(rc, name)
	if err != nil {
		return &oidcConfig{}, fmt.Errorf("unable to resolve OIDC configuration")
	}

	res := &oidcConfig{
		DiscoveryURL:     cfg.Spec.DiscoveryURL,
		AuthorizeURL:     cfg.Spec.AuthorizationURL,
		TokenURL:         cfg.Spec.TokenURL,
		RedirectURI:      cfg.Spec.RedirectURI,
		UserInfoURL:      cfg.Spec.UserInfoURL,
		ClientID:         cfg.Spec.ClientID,
		AdditionalScopes: cfg.Spec.AdditionalScopes,
	}

	if ref := cfg.Spec.ClientSecret; ref != nil {
		sec, err := secrets.Get(context.Background(), rc, ref)
		if err != nil {
			return res, err
		}
		if val, ok := sec.Data[ref.Key]; ok {
			res.ClientSecret = string(val)
		}
	}

	return res, nil
}

func doLogin(code string, cfg *oidcConfig) (idToken, error) {
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", cfg.RedirectURI)
	data.Set("grant_type", "authorization_code")
	//data.Set("grant_type", "password")
	//data.Set("username", username)
	//data.Set("password", password)
	//data.Set("scope", "openid email profile "+cfg.AdditionalScopes)

	resp, err := http.PostForm(cfg.TokenURL, data)
	if err != nil {
		return idToken{}, fmt.Errorf("failed to send request to token endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return idToken{}, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return idToken{}, fmt.Errorf("token endpoint returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	err = json.Unmarshal(body, &token)
	if err != nil {
		return idToken{}, fmt.Errorf("failed to unmarshal token response: %v", err)
	}

	claims, err := decodeJWT(token.IDToken)
	if err != nil {
		return idToken{}, fmt.Errorf("failed to decode JWT token: %v", err)
	}

	callUserInfo := false
	var res idToken
	// Check if the values are in the map, if not, call the userinfo endpoint
	if value, ok := claims["preferred_username"]; ok {
		res.preferredUsername = value.(string)
	} else {
		callUserInfo = true
	}

	if value, ok := claims["name"]; ok {
		res.name = value.(string)
	} else {
		callUserInfo = true
	}

	if value, ok := claims["picture"]; ok {
		res.avatarURL = value.(string)
	} else {
		callUserInfo = true
	}

	if value, ok := claims["email"]; ok {
		res.avatarURL = value.(string)
	} else {
		callUserInfo = true
	}

	if value, ok := claims["groups"]; ok {
		interfaceArray := value.([]interface{})
		stringArray := []string{}
		for _, interfaceValue := range interfaceArray {
			stringArray = append(stringArray, interfaceValue.(string))
		}
		res.groups = stringArray
	} // we do not call userinfo for groups because groups are not part of the standard response for the userinfo endpoint

	if callUserInfo && cfg.UserInfoURL != "" {
		if token.AccessToken != "" {
			request, err := http.NewRequest(http.MethodGet, cfg.UserInfoURL, nil)
			if err != nil {
				return idToken{}, fmt.Errorf("failed to create http request for userinfo endpoint: %v", err)
			}
			request.Header.Set("Authorization", "Bearer "+token.AccessToken)
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				return idToken{}, fmt.Errorf("failed to send userinfo request: %v", err)
			}
			userInfoDataJson, err := io.ReadAll(resp.Body)
			if err != nil {
				return idToken{}, fmt.Errorf("failed to read userinfo response: %v", err)
			}
			var userInfo map[string]interface{}
			err = json.Unmarshal(userInfoDataJson, &userInfo)
			if err != nil {
				return idToken{}, fmt.Errorf("error parsing userinfo payload JSON: %v", err)
			}

			// Replace the missing values that we did not find in the id token
			if _, ok := claims["preferred_username"]; !ok {
				res.preferredUsername = userInfo["preferred_username"].(string)
			}

			if _, ok := claims["name"]; !ok {
				res.name = userInfo["name"].(string)
			}

			if _, ok := claims["picture"]; !ok {
				res.avatarURL = userInfo["picture"].(string)
			}

			if _, ok := claims["email"]; !ok {
				res.avatarURL = userInfo["email"].(string)
			}
		} else {
			return idToken{}, fmt.Errorf("unable to get access_token from response")
		}
	}

	return res, nil

}

func decodeJWT(tokenString string) (map[string]interface{}, error) {
	var claims map[string]interface{}
	// Split the token into header, payload, and signature
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return claims, fmt.Errorf("invalid token: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, fmt.Errorf("error decoding payload: %v", err)
	}

	// Parse the JSON payload
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return claims, fmt.Errorf("error parsing payload JSON: %v", err)
	}

	return claims, nil
}
