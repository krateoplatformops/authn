package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

func newGithubApiClient(tok *oauth2.Token, org, apiUrl string) *githubMiniClient {
	return &githubMiniClient{
		cli:    oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(tok)),
		org:    org,
		apiUrl: apiUrl,
	}
}

type githubMiniClient struct {
	cli    *http.Client
	org    string
	apiUrl string
}

func (mc *githubMiniClient) getUserInfo() (userInfo, error) {
	uri, err := url.JoinPath(mc.apiUrl, "user")
	if err != nil {
		return userInfo{}, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return userInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := mc.cli.Do(req)
	if err != nil {
		return userInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return userInfo{}, fmt.Errorf(res.Status)
	}

	var val userInfo
	err = json.NewDecoder(res.Body).Decode(&val)
	return val, err
}

func (mc *githubMiniClient) listTeams() ([]teamInfo, error) {
	uri, err := url.JoinPath(mc.apiUrl, fmt.Sprintf("/orgs/%s/teams", mc.org))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := mc.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf(res.Status)
	}

	all := make([]teamInfo, 0)
	err = json.NewDecoder(res.Body).Decode(&all)
	if err != nil {
		return nil, err
	}

	return all, nil
}

func (mc *githubMiniClient) isUserMemberOfTeam(username string, team teamInfo) (bool, error) {
	uri, err := url.JoinPath(mc.apiUrl, "/orgs/%s/teams/%s/memberships/%s",
		mc.org, team.Slug, username)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := mc.cli.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return (res.StatusCode == 200), nil
}

type userInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
}

type teamInfo struct {
	ID          string `json:"node_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
